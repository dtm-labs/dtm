package workflow

import (
	"fmt"
	"net/url"

	"github.com/dtm-labs/dtm/dtmcli/logger"
)

type workflowFactory struct {
	protocol     string
	httpDtm      string
	httpCallback string
	grpcDtm      string
	grpcCallback string
	handlers     map[string]*wfItem
}

var defaultFac = workflowFactory{
	handlers: map[string]*wfItem{},
}

func (w *workflowFactory) execute(name string, gid string, data []byte) error {
	handler := w.handlers[name]
	if handler == nil {
		return fmt.Errorf("workflow '%s' not registered. please register at startup", name)
	}
	wf := w.newWorkflow(name, gid, data)
	for _, fn := range handler.custom {
		fn(wf)
	}
	return wf.process(handler.fn, data)
}

func (w *workflowFactory) executeByQS(qs url.Values, body []byte) error {
	name := qs.Get("op")
	gid := qs.Get("gid")
	return w.execute(name, gid, body)
}

func (w *workflowFactory) register(name string, handler WfFunc, custom ...func(wf *Workflow)) error {
	e := w.handlers[name]
	if e != nil {
		return fmt.Errorf("a handler already exists for %s", name)
	}
	logger.Debugf("workflow '%s' registered.", name)
	w.handlers[name] = &wfItem{
		fn:     handler,
		custom: custom,
	}
	return nil
}
