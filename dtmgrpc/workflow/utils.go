package workflow

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmgrpc/dtmgimp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func statusToCode(status string) int {
	if status == "succeed" {
		return 200
	}
	return 409
}

func wfErrorToStatus(err error) string {
	if err == nil {
		return dtmcli.StatusSucceed
	} else if errors.Is(err, dtmcli.ErrFailure) {
		return dtmcli.StatusFailed
	}
	return ""
}

type stepResult struct {
	Error  error  // if Error != nil || Status == "", result will not be saved
	Status string // succeed | failed | ""
	// if status == succeed, data is the result.
	// if status == failed, data is the error message
	Data []byte
}

type roundTripper struct {
	old http.RoundTripper
	wf  *Workflow
}

func newJSONResponse(status int, result []byte) *http.Response {
	return &http.Response{
		Status:     strconv.Itoa(status),
		StatusCode: status,
		Body:       NewRespBodyFromBytes(result),
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		ContentLength: -1,
	}
}

func (r *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	wf := r.wf
	if wf.currentOp != dtmimp.OpAction { // in phase 2, do not save, because it is saved outer
		return r.old.RoundTrip(req)
	}
	sr := wf.recordedDo(func(bb *dtmcli.BranchBarrier) *stepResult {
		resp, err := r.old.RoundTrip(req)
		return stepResultFromHTTP(resp, err)
	})
	return stepResultToHTTP(sr)
}

func newRoundTripper(old http.RoundTripper, wf *Workflow) http.RoundTripper {
	return &roundTripper{old: old, wf: wf}
}

func stepResultFromLocal(data []byte, err error) *stepResult {
	return &stepResult{
		Error:  err,
		Status: wfErrorToStatus(err),
		Data:   data,
	}
}

func stepResultToLocal(s *stepResult) ([]byte, error) {
	if s.Error != nil {
		return nil, s.Error
	} else if s.Status == dtmcli.StatusFailed {
		return nil, fmt.Errorf("%s. %w", string(s.Data), dtmcli.ErrFailure)
	}
	return s.Data, nil
}

func stepResultFromGrpc(reply interface{}, err error) *stepResult {
	sr := &stepResult{}
	st, ok := status.FromError(err)
	if err == nil {
		sr.Status = dtmcli.StatusSucceed
		sr.Data = dtmgimp.MustProtoMarshal(reply.(protoreflect.ProtoMessage))
	} else if ok && st.Code() == codes.Aborted {
		sr.Status = dtmcli.StatusFailed
		sr.Data = []byte(st.Message())
	} else {
		sr.Error = err
	}
	return sr
}

func stepResultToGrpc(s *stepResult, reply interface{}) error {
	if s.Error != nil {
		return s.Error
	} else if s.Status == dtmcli.StatusSucceed {
		dtmgimp.MustProtoUnmarshal(s.Data, reply.(protoreflect.ProtoMessage))
		return nil
	}
	return status.New(codes.Aborted, string(s.Data)).Err()
}

func stepResultFromHTTP(resp *http.Response, err error) *stepResult {
	sr := &stepResult{Error: err}
	if err == nil {
		sr.Data, sr.Error = ioutil.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusOK {
			sr.Status = dtmcli.StatusSucceed
		} else if resp.StatusCode == http.StatusConflict {
			sr.Status = dtmcli.StatusFailed
		}
	}
	return sr
}

func stepResultToHTTP(s *stepResult) (*http.Response, error) {
	if s.Error != nil {
		return nil, s.Error
	}
	return newJSONResponse(statusToCode(s.Status), s.Data), nil
}
