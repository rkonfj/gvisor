// Copyright 2022 The gVisor Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package container

import (
	"encoding/json"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	"gvisor.dev/gvisor/pkg/sentry/seccheck"
	"gvisor.dev/gvisor/pkg/sentry/seccheck/checkers/remote/test"
	pb "gvisor.dev/gvisor/pkg/sentry/seccheck/points/points_go_proto"
	"gvisor.dev/gvisor/pkg/test/testutil"
	"gvisor.dev/gvisor/runsc/boot"
)

func remoteSinkConfig(endpoint string) seccheck.SinkConfig {
	return seccheck.SinkConfig{
		Name: "remote",
		Config: map[string]interface{}{
			"endpoint": endpoint,
		},
	}
}

// Test that setting up a trace session configuration in PodInitConfig creates
// a session before container creation.
func TestTraceStartup(t *testing.T) {
	// Test on all configurations to ensure that point can be sent to an outside
	// process in all cases. Rest of the tests don't require all configs.
	for name, conf := range configs(t, false /* noOverlay */) {
		t.Run(name, func(t *testing.T) {
			server, err := test.NewServer()
			if err != nil {
				t.Fatalf("newServer(): %v", err)
			}
			defer server.Close()

			podInitConfig, err := ioutil.TempFile(testutil.TmpDir(), "config")
			if err != nil {
				t.Fatalf("error creating tmp file: %v", err)
			}
			defer podInitConfig.Close()

			initConfig := boot.InitConfig{
				TraceSession: seccheck.SessionConfig{
					Name: seccheck.DefaultSessionName,
					Points: []seccheck.PointConfig{
						{
							Name:          "container/start",
							ContextFields: []string{"container_id"},
						},
					},
					Sinks: []seccheck.SinkConfig{remoteSinkConfig(server.Path)},
				},
			}
			encoder := json.NewEncoder(podInitConfig)
			if err := encoder.Encode(&initConfig); err != nil {
				t.Fatalf("JSON encode: %v", err)
			}
			conf.PodInitConfig = podInitConfig.Name()

			spec := testutil.NewSpecWithArgs("/bin/true")
			if err := run(spec, conf); err != nil {
				t.Fatalf("Error running container: %v", err)
			}

			// Wait for the point to be received and then check that fields match.
			if err := server.WaitForCount(1); err != nil {
				t.Fatalf("WaitForCount(1): %v", err)
			}
			pt := server.GetPoints()[0]
			if want := pb.MessageType_MESSAGE_CONTAINER_START; pt.MsgType != want {
				t.Errorf("wrong message type, want: %v, got: %v", want, pt.MsgType)
			}
			got := &pb.Start{}
			if err := proto.Unmarshal(pt.Msg, got); err != nil {
				t.Errorf("proto.Unmarshal(Start): %v", err)
			}
			if want := "/bin/true"; len(got.Args) != 1 || want != got.Args[0] {
				t.Errorf("container.Start.Args, want: %q, got: %q", want, got.Args)
			}
			if want, got := got.Id, got.ContextData.ContainerId; want != got {
				t.Errorf("Mismatched container ID, want: %v, got: %v", want, got)
			}
		})
	}
}

func TestTraceLifecycle(t *testing.T) {
	spec, conf := sleepSpecConf(t)
	_, bundleDir, cleanup, err := testutil.SetupContainer(spec, conf)
	if err != nil {
		t.Fatalf("error setting up container: %v", err)
	}
	defer cleanup()

	// Create and start the container.
	args := Args{
		ID:        testutil.RandomContainerID(),
		Spec:      spec,
		BundleDir: bundleDir,
	}
	cont, err := New(conf, args)
	if err != nil {
		t.Fatalf("error creating container: %v", err)
	}
	defer cont.Destroy()
	if err := cont.Start(conf); err != nil {
		t.Fatalf("error starting container: %v", err)
	}

	// Check that no session are created.
	if sessions, err := cont.Sandbox.ListTraceSessions(); err != nil {
		t.Fatalf("ListTraceSessions(): %v", err)
	} else if len(sessions) != 0 {
		t.Fatalf("no session should exist, got: %+v", sessions)
	}

	// Create a new trace session on the fly.
	server, err := test.NewServer()
	if err != nil {
		t.Fatalf("newServer(): %v", err)
	}
	defer server.Close()

	session := seccheck.SessionConfig{
		Name: "Default",
		Points: []seccheck.PointConfig{
			{
				Name:          "sentry/task_exit",
				ContextFields: []string{"container_id"},
			},
		},
		Sinks: []seccheck.SinkConfig{remoteSinkConfig(server.Path)},
	}
	if err := cont.Sandbox.CreateTraceSession(&session, false); err != nil {
		t.Fatalf("CreateTraceSession(): %v", err)
	}

	// Trigger the configured point and want to receive it in the server.
	if ws, err := execute(conf, cont, "/bin/true"); err != nil || ws != 0 {
		t.Fatalf("exec: true, ws: %v, err: %v", ws, err)
	}
	if err := server.WaitForCount(1); err != nil {
		t.Fatalf("WaitForCount(1): %v", err)
	}
	pt := server.GetPoints()[0]
	if want := pb.MessageType_MESSAGE_SENTRY_TASK_EXIT; pt.MsgType != want {
		t.Errorf("wrong message type, want: %v, got: %v", want, pt.MsgType)
	}
	got := &pb.TaskExit{}
	if err := proto.Unmarshal(pt.Msg, got); err != nil {
		t.Errorf("proto.Unmarshal(TaskExit): %v", err)
	}
	if got.ExitStatus != 0 {
		t.Errorf("Wrong TaskExit.ExitStatus, want: 0, got: %+v", got)
	}
	if want, got := cont.ID, got.ContextData.ContainerId; want != got {
		t.Errorf("Wrong TaskExit.ContextData.ContainerId, want: %v, got: %v", want, got)
	}

	// Check that no more points were received and reset the server for the
	// remaining tests.
	if want, got := 1, server.Reset(); want != got {
		t.Errorf("wrong number of points, want: %d, got: %d", want, got)
	}

	// List and check that trace session is reported.
	sessions, err := cont.Sandbox.ListTraceSessions()
	if err != nil {
		t.Fatalf("ListTraceSessions(): %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected a single session, got: %+v", sessions)
	}
	if got := sessions[0].Name; seccheck.DefaultSessionName != got {
		t.Errorf("wrong session, want: %v, got: %v", seccheck.DefaultSessionName, got)
	}

	if err := cont.Sandbox.DeleteTraceSession("Default"); err != nil {
		t.Fatalf("DeleteTraceSession(): %v", err)
	}

	// Check that session was indeed deleted.
	if sessions, err := cont.Sandbox.ListTraceSessions(); err != nil {
		t.Fatalf("ListTraceSessions(): %v", err)
	} else if len(sessions) != 0 {
		t.Fatalf("no session should exist, got: %+v", sessions)
	}

	// Trigger the point again and check that it's not received.
	if ws, err := execute(conf, cont, "/bin/true"); err != nil || ws != 0 {
		t.Fatalf("exec: true, ws: %v, err: %v", ws, err)
	}
	time.Sleep(time.Second) // give some time to receive the point.
	if server.Count() > 0 {
		t.Errorf("point received after session was deleted: %+v", server.GetPoints())
	}
}

func TestTraceForceCreate(t *testing.T) {
	spec, conf := sleepSpecConf(t)
	_, bundleDir, cleanup, err := testutil.SetupContainer(spec, conf)
	if err != nil {
		t.Fatalf("error setting up container: %v", err)
	}
	defer cleanup()

	// Create and start the container.
	args := Args{
		ID:        testutil.RandomContainerID(),
		Spec:      spec,
		BundleDir: bundleDir,
	}
	cont, err := New(conf, args)
	if err != nil {
		t.Fatalf("error creating container: %v", err)
	}
	defer cont.Destroy()
	if err := cont.Start(conf); err != nil {
		t.Fatalf("error starting container: %v", err)
	}

	// Create a new trace session that will be overwritten.
	server, err := test.NewServer()
	if err != nil {
		t.Fatalf("newServer(): %v", err)
	}
	defer server.Close()

	session := seccheck.SessionConfig{
		Name: "Default",
		Points: []seccheck.PointConfig{
			{Name: "sentry/exit_notify_parent"},
		},
		Sinks: []seccheck.SinkConfig{remoteSinkConfig(server.Path)},
	}
	if err := cont.Sandbox.CreateTraceSession(&session, false); err != nil {
		t.Fatalf("CreateTraceSession(): %v", err)
	}

	// Trigger the configured point to check that trace session is enabled.
	if ws, err := execute(conf, cont, "/bin/true"); err != nil || ws != 0 {
		t.Fatalf("exec: true, ws: %v, err: %v", ws, err)
	}
	if err := server.WaitForCount(1); err != nil {
		t.Fatalf("WaitForCount(1): %v", err)
	}
	pt := server.GetPoints()[0]
	if want := pb.MessageType_MESSAGE_SENTRY_EXIT_NOTIFY_PARENT; pt.MsgType != want {
		t.Errorf("wrong message type, want: %v, got: %v", want, pt.MsgType)
	}
	server.Reset()

	// Check that creating the same session fails.
	if err := cont.Sandbox.CreateTraceSession(&session, false); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Errorf("CreateTraceSession() again failed with wrong error: %v", err)
	}

	// Re-create the session with a different point using force=true and check
	// that it overwrote the other trace session.
	session = seccheck.SessionConfig{
		Name: "Default",
		Points: []seccheck.PointConfig{
			{Name: "sentry/task_exit"},
		},
		Sinks: []seccheck.SinkConfig{remoteSinkConfig(server.Path)},
	}
	if err := cont.Sandbox.CreateTraceSession(&session, true); err != nil {
		t.Fatalf("CreateTraceSession(force): %v", err)
	}

	if ws, err := execute(conf, cont, "/bin/true"); err != nil || ws != 0 {
		t.Fatalf("exec: true, ws: %v, err: %v", ws, err)
	}
	if err := server.WaitForCount(1); err != nil {
		t.Fatalf("WaitForCount(1): %v", err)
	}
	pt = server.GetPoints()[0]
	if want := pb.MessageType_MESSAGE_SENTRY_TASK_EXIT; pt.MsgType != want {
		t.Errorf("wrong message type, want: %v, got: %v", want, pt.MsgType)
	}
}

func TestProcfsDump(t *testing.T) {
	spec, conf := sleepSpecConf(t)
	testEnv := "GVISOR_IS_GREAT=true"
	spec.Process.Env = append(spec.Process.Env, testEnv)
	spec.Process.Cwd = "/"
	_, bundleDir, cleanup, err := testutil.SetupContainer(spec, conf)
	if err != nil {
		t.Fatalf("error setting up container: %v", err)
	}
	defer cleanup()

	// Create and start the container.
	args := Args{
		ID:        testutil.RandomContainerID(),
		Spec:      spec,
		BundleDir: bundleDir,
	}
	cont, err := New(conf, args)
	if err != nil {
		t.Fatalf("error creating container: %v", err)
	}
	defer cont.Destroy()
	if err := cont.Start(conf); err != nil {
		t.Fatalf("error starting container: %v", err)
	}

	procfsDump, err := cont.Sandbox.ProcfsDump()
	if err != nil {
		t.Fatalf("ProcfsDump() failed: %v", err)
	}

	// Sleep should be the only process running in the container.
	if len(procfsDump) != 1 {
		t.Fatalf("got incorrect number of proc results: %+v", procfsDump)
	}

	// Sleep should be PID 1.
	if procfsDump[0].PID != 1 {
		t.Errorf("expected sleep process to be pid 1, got %d", procfsDump[0].PID)
	}

	// Check that bin/sleep is part of the executable path.
	if wantExeSubStr := "bin/sleep"; !strings.HasSuffix(procfsDump[0].Exe, wantExeSubStr) {
		t.Errorf("expected %q to be part of execuable path %q", wantExeSubStr, procfsDump[0].Exe)
	}

	if len(procfsDump[0].Args) != 2 {
		t.Errorf("expected 2 args, but got %+v", procfsDump[0].Args)
	} else {
		if procfsDump[0].Args[0] != "sleep" || procfsDump[0].Args[1] != "1000" {
			t.Errorf("expected args %q but got %+v", "sleep 1000", procfsDump[0].Args)
		}
	}

	testEnvFound := false
	for _, env := range procfsDump[0].Env {
		if env == testEnv {
			testEnvFound = true
		}
	}
	if !testEnvFound {
		t.Errorf("expected to find %q env but did not find it, got env %+v", testEnv, procfsDump[0].Env)
	}

	if spec.Process.Cwd != procfsDump[0].CWD {
		t.Errorf("expected CWD %q, got %q", spec.Process.Cwd, procfsDump[0].CWD)
	}
}
