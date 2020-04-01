// Copyright 2020 The gVisor Authors.
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

package testbench

import (
	"context"
	"flag"
	"net"
	"strconv"
	"syscall"
	"testing"
	"time"

	pb "gvisor.dev/gvisor/test/packetimpact/proto/posix_server_go_proto"

	"golang.org/x/sys/unix"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var (
	posixServerIP   = flag.String("posix_server_ip", "", "ip address to listen to for UDP commands")
	posixServerPort = flag.Int("posix_server_port", 40000, "port to listen to for UDP commands")
	rpcTimeout      = flag.Duration("rpc_timeout", 100*time.Millisecond, "gRPC timeout")
	rpcKeepalive    = flag.Duration("rpc_keepalive", 10*time.Second, "gRPC keepalive")
)

// DUT communicates with the DUT to force it to make POSIX calls.
type DUT struct {
	t           *testing.T
	conn        *grpc.ClientConn
	posixServer PosixClient
}

// NewDUT creates a new connection with the DUT over gRPC.
func NewDUT(t *testing.T) DUT {
	flag.Parse()
	posixServerAddress := *posixServerIP + ":" + strconv.Itoa(*posixServerPort)
	conn, err := grpc.Dial(posixServerAddress, grpc.WithInsecure(), grpc.WithKeepaliveParams(keepalive.ClientParameters{Timeout: *rpcKeepalive}))
	if err != nil {
		t.Fatalf("failed to grpc.Dial(%s): %s", posixServerAddress, err)
	}
	posixServer := NewPosixClient(conn)
	return DUT{
		t:           t,
		conn:        conn,
		posixServer: posixServer,
	}
}

// TearDown closes the underlying connection.
func (dut *DUT) TearDown() {
	dut.conn.Close()
}

// SocketWithErrno calls socket on the DUT and returns the fd and errno.
func (dut *DUT) SocketWithErrno(domain, typ, proto int32) (int32, error) {
	dut.t.Helper()
	req := pb.SocketRequest{
		Domain:   domain,
		Type:     typ,
		Protocol: proto,
	}
	ctx := context.Background()
	resp, err := dut.posixServer.Socket(ctx, &req)
	if err != nil {
		dut.t.Fatalf("failed to call Socket: %s", err)
	}
	return resp.GetFd(), syscall.Errno(resp.GetErrno_())
}

// Socket calls socket on the DUT and returns the file descriptor. If socket
// fails on the DUT, the test ends.
func (dut *DUT) Socket(domain, typ, proto int32) int32 {
	dut.t.Helper()
	fd, err := dut.SocketWithErrno(domain, typ, proto)
	if fd < 0 {
		dut.t.Fatalf("failed to create socket: %s", err)
	}
	return fd
}

func (dut *DUT) sockaddrToProto(sa unix.Sockaddr) *pb.Sockaddr {
	dut.t.Helper()
	switch s := sa.(type) {
	case *unix.SockaddrInet4:
		return &pb.Sockaddr{
			Sockaddr: &pb.Sockaddr_In{
				In: &pb.SockaddrIn{
					Family: unix.AF_INET,
					Port:   uint32(s.Port),
					Addr:   s.Addr[:],
				},
			},
		}
	case *unix.SockaddrInet6:
		return &pb.Sockaddr{
			Sockaddr: &pb.Sockaddr_In6{
				In6: &pb.SockaddrIn6{
					Family:   unix.AF_INET6,
					Port:     uint32(s.Port),
					Flowinfo: 0,
					ScopeId:  s.ZoneId,
					Addr:     s.Addr[:],
				},
			},
		}
	}
	dut.t.Fatalf("can't parse Sockaddr: %+v", sa)
	return nil
}

func (dut *DUT) protoToSockaddr(sa *pb.Sockaddr) unix.Sockaddr {
	dut.t.Helper()
	switch s := sa.Sockaddr.(type) {
	case *pb.Sockaddr_In:
		ret := unix.SockaddrInet4{
			Port: int(s.In.GetPort()),
		}
		copy(ret.Addr[:], s.In.GetAddr())
		return &ret
	case *pb.Sockaddr_In6:
		ret := unix.SockaddrInet6{
			Port:   int(s.In6.GetPort()),
			ZoneId: s.In6.GetScopeId(),
		}
		copy(ret.Addr[:], s.In6.GetAddr())
	}
	dut.t.Fatalf("can't parse Sockaddr: %+v", sa)
	return nil
}

// BindWithErrno calls bind on the DUT.
func (dut *DUT) BindWithErrno(ctx context.Context, fd int32, sa unix.Sockaddr) (int32, error) {
	dut.t.Helper()
	req := pb.BindRequest{
		Sockfd: fd,
		Addr:   dut.sockaddrToProto(sa),
	}
	resp, err := dut.posixServer.Bind(ctx, &req)
	if err != nil {
		dut.t.Fatalf("failed to call Bind: %s", err)
	}
	return resp.GetRet(), syscall.Errno(resp.GetErrno_())
}

// Bind calls bind on the DUT and causes a fatal test failure if it doesn't
// succeed. If more control over the timeout or error handling is
// needed, use BindWithErrno.
func (dut *DUT) Bind(fd int32, sa unix.Sockaddr) {
	dut.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), *rpcTimeout)
	defer cancel()
	ret, err := dut.BindWithErrno(ctx, fd, sa)
	if ret != 0 {
		dut.t.Fatalf("failed to bind socket: %s", err)
	}
}

// GetSockNameWithErrno calls getsockname on the DUT.
func (dut *DUT) GetSockNameWithErrno(ctx context.Context, sockfd int32) (int32, unix.Sockaddr, error) {
	dut.t.Helper()
	req := pb.GetSockNameRequest{
		Sockfd: sockfd,
	}
	resp, err := dut.posixServer.GetSockName(ctx, &req)
	if err != nil {
		dut.t.Fatalf("failed to call Bind: %s", err)
	}
	return resp.GetRet(), dut.protoToSockaddr(resp.GetAddr()), syscall.Errno(resp.GetErrno_())
}

// GetSockName calls getsockname on the DUT and causes a fatal test failure if
// it doesn't succeed. If more control over the timeout or error handling is
// needed, use GetSockNameWithErrno.
func (dut *DUT) GetSockName(sockfd int32) unix.Sockaddr {
	dut.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), *rpcTimeout)
	defer cancel()
	ret, sa, err := dut.GetSockNameWithErrno(ctx, sockfd)
	if ret != 0 {
		dut.t.Fatalf("failed to getsockname: %s", err)
	}
	return sa
}

// ListenWithErrno calls listen on the DUT.
func (dut *DUT) ListenWithErrno(ctx context.Context, sockfd, backlog int32) (int32, error) {
	dut.t.Helper()
	req := pb.ListenRequest{
		Sockfd:  sockfd,
		Backlog: backlog,
	}
	resp, err := dut.posixServer.Listen(ctx, &req)
	if err != nil {
		dut.t.Fatalf("failed to call Listen: %s", err)
	}
	return resp.GetRet(), syscall.Errno(resp.GetErrno_())
}

// Listen calls listen on the DUT and causes a fatal test failure if it doesn't
// succeed. If more control over the timeout or error handling is needed, use
// ListenWithErrno.
func (dut *DUT) Listen(sockfd, backlog int32) {
	dut.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), *rpcTimeout)
	defer cancel()
	ret, err := dut.ListenWithErrno(ctx, sockfd, backlog)
	if ret != 0 {
		dut.t.Fatalf("failed to listen: %s", err)
	}
}

// AcceptWithErrno calls accept on the DUT.
func (dut *DUT) AcceptWithErrno(ctx context.Context, sockfd int32) (int32, unix.Sockaddr, error) {
	dut.t.Helper()
	req := pb.AcceptRequest{
		Sockfd: sockfd,
	}
	resp, err := dut.posixServer.Accept(ctx, &req)
	if err != nil {
		dut.t.Fatalf("failed to call Accept: %s", err)
	}
	return resp.GetFd(), dut.protoToSockaddr(resp.GetAddr()), syscall.Errno(resp.GetErrno_())
}

// Accept calls accept on the DUT and causes a fatal test failure if it doesn't
// succeed. If more control over the timeout or error handling is needed, use
// AcceptWithErrno.
func (dut *DUT) Accept(sockfd int32) (int32, unix.Sockaddr) {
	dut.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), *rpcTimeout)
	defer cancel()
	fd, sa, err := dut.AcceptWithErrno(ctx, sockfd)
	if fd < 0 {
		dut.t.Fatalf("failed to accept: %s", err)
	}
	return fd, sa
}

// SetSockOptWithErrno calls setsockopt on the DUT. Because endianess and the
// width of values might differ between the testbench and DUT architectures,
// prefer to use a more specific SetSockOptXxxWithErrno function.
func (dut *DUT) SetSockOptWithErrno(ctx context.Context, sockfd, level, optname int32, optval []byte) (int32, error) {
	dut.t.Helper()
	req := pb.SetSockOptRequest{
		Sockfd:  sockfd,
		Level:   level,
		Optname: optname,
		Optval:  optval,
	}
	resp, err := dut.posixServer.SetSockOpt(ctx, &req)
	if err != nil {
		dut.t.Fatalf("failed to call SetSockOpt: %s", err)
	}
	return resp.GetRet(), syscall.Errno(resp.GetErrno_())
}

// SetSockOpt calls setsockopt on the DUT and causes a fatal test failure if it
// doesn't succeed. If more control over the timeout or error handling is
// needed, use SetSockOptWithErrno. Because endianess and the width of values
// might differ between the testbench and DUT architectures, prefer to use a
// more specific SetSockOptXxx function.
func (dut *DUT) SetSockOpt(sockfd, level, optname int32, optval []byte) {
	dut.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), *rpcTimeout)
	defer cancel()
	ret, err := dut.SetSockOptWithErrno(ctx, sockfd, level, optname, optval)
	if ret != 0 {
		dut.t.Fatalf("failed to SetSockOpt: %s", err)
	}
}

// SetSockOptTimevalWithErrno calls setsockopt with the timeval converted to
// bytes.
func (dut *DUT) SetSockOptTimevalWithErrno(ctx context.Context, sockfd, level, optname int32, tv *unix.Timeval) (int32, error) {
	dut.t.Helper()
	timeval := pb.Timeval{
		Seconds:      int64(tv.Sec),
		Microseconds: int64(tv.Usec),
	}
	req := pb.SetSockOptTimevalRequest{
		Sockfd:  sockfd,
		Level:   level,
		Optname: optname,
		Timeval: &timeval,
	}
	resp, err := dut.posixServer.SetSockOptTimeval(ctx, &req)
	if err != nil {
		dut.t.Fatalf("failed to call SetSockOptTimeval: %s", err)
	}
	return resp.GetRet(), syscall.Errno(resp.GetErrno_())
}

// SetSockOptTimeval calls setsockopt on the DUT and causes a fatal test failure
// if it doesn't succeed. If more control over the timeout or error handling is
// needed, use SetSockOptTimevalWithErrno.
func (dut *DUT) SetSockOptTimeval(sockfd, level, optname int32, tv *unix.Timeval) {
	dut.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), *rpcTimeout)
	defer cancel()
	ret, err := dut.SetSockOptTimevalWithErrno(ctx, sockfd, level, optname, tv)
	if ret != 0 {
		dut.t.Fatalf("failed to SetSockOptTimeval: %s", err)
	}
}

// RecvWithErrno calls recv on the DUT.
func (dut *DUT) RecvWithErrno(ctx context.Context, sockfd, len, flags int32) (int32, []byte, error) {
	dut.t.Helper()
	req := pb.RecvRequest{
		Sockfd: sockfd,
		Len:    len,
		Flags:  flags,
	}
	resp, err := dut.posixServer.Recv(ctx, &req)
	if err != nil {
		dut.t.Fatalf("failed to call Recv: %s", err)
	}
	return resp.GetRet(), resp.GetBuf(), syscall.Errno(resp.GetErrno_())
}

// Recv calls recv on the DUT and causes a fatal test failure if it doesn't
// succeed. If more control over the timeout or error handling is needed, use
// RecvWithErrno.
func (dut *DUT) Recv(sockfd, len, flags int32) []byte {
	dut.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), *rpcTimeout)
	defer cancel()
	ret, buf, err := dut.RecvWithErrno(ctx, sockfd, len, flags)
	if ret == -1 {
		dut.t.Fatalf("failed to recv: %s", err)
	}
	return buf
}

// CloseWithErrno calls close on the DUT.
func (dut *DUT) CloseWithErrno(ctx context.Context, fd int32) (int32, error) {
	dut.t.Helper()
	req := pb.CloseRequest{
		Fd: fd,
	}
	resp, err := dut.posixServer.Close(ctx, &req)
	if err != nil {
		dut.t.Fatalf("failed to call Close: %s", err)
	}
	return resp.GetRet(), syscall.Errno(resp.GetErrno_())
}

// Close calls close on the DUT and causes a fatal test failure if it doesn't
// succeed. If more control over the timeout or error handling is needed, use
// CloseWithErrno.
func (dut *DUT) Close(fd int32) {
	dut.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), *rpcTimeout)
	defer cancel()
	ret, err := dut.CloseWithErrno(ctx, fd)
	if ret != 0 {
		dut.t.Fatalf("failed to close: %s", err)
	}
}

// CreateBoundSocket makes a new socket on the DUT, with type typ and protocol
// proto, and bound to the IP address addr. Returns the new file descriptor and
// the port that was selected on the DUT.
func (dut *DUT) CreateBoundSocket(typ, proto int32, addr net.IP) (int32, uint16) {
	dut.t.Helper()
	var fd int32
	if addr.To4() != nil {
		fd = dut.Socket(unix.AF_INET, typ, proto)
		sa := unix.SockaddrInet4{}
		copy(sa.Addr[:], addr.To4())
		dut.Bind(fd, &sa)
	} else if addr.To16() != nil {
		fd = dut.Socket(unix.AF_INET6, typ, proto)
		sa := unix.SockaddrInet6{}
		copy(sa.Addr[:], addr.To16())
		dut.Bind(fd, &sa)
	} else {
		dut.t.Fatal("unknown ip addr type for remoteIP")
	}
	sa := dut.GetSockName(fd)
	var port int
	switch s := sa.(type) {
	case *unix.SockaddrInet4:
		port = s.Port
	case *unix.SockaddrInet6:
		port = s.Port
	default:
		dut.t.Fatalf("unknown sockaddr type from getsockname: %t", sa)
	}
	return fd, uint16(port)
}

// CreateListener makes a new TCP connection. If it fails, the test ends.
func (dut *DUT) CreateListener(typ, proto, backlog int32) (int32, uint16) {
	fd, remotePort := dut.CreateBoundSocket(typ, proto, net.ParseIP(*remoteIPv4))
	dut.Listen(fd, backlog)
	return fd, remotePort
}
