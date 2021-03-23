// automatically generated by stateify.

package eventfd

import (
	"gvisor.dev/gvisor/pkg/state"
)

func (e *EventOperations) StateTypeName() string {
	return "pkg/sentry/kernel/eventfd.EventOperations"
}

func (e *EventOperations) StateFields() []string {
	return []string{
		"val",
		"semMode",
		"hostfd",
	}
}

func (e *EventOperations) beforeSave() {}

// +checklocksignore
func (e *EventOperations) StateSave(stateSinkObject state.Sink) {
	e.beforeSave()
	if !state.IsZeroValue(&e.wq) {
		state.Failf("wq is %#v, expected zero", &e.wq)
	}
	stateSinkObject.Save(0, &e.val)
	stateSinkObject.Save(1, &e.semMode)
	stateSinkObject.Save(2, &e.hostfd)
}

func (e *EventOperations) afterLoad() {}

// +checklocksignore
func (e *EventOperations) StateLoad(stateSourceObject state.Source) {
	stateSourceObject.Load(0, &e.val)
	stateSourceObject.Load(1, &e.semMode)
	stateSourceObject.Load(2, &e.hostfd)
}

func init() {
	state.Register((*EventOperations)(nil))
}
