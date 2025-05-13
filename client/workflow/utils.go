package workflow

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/dtm-labs/dtm/client/dtmcli"
	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/client/dtmgrpc/dtmgimp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
)

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
	origin := func(bb *dtmcli.BranchBarrier) *stepResult {
		resp, err := r.old.RoundTrip(req)
		return wf.stepResultFromHTTP(resp, err)
	}
	var sr *stepResult
	if wf.currentOp != dtmimp.OpAction { // in phase 2, do not save, because it is saved outer
		sr = origin(nil)
	} else {
		sr = wf.recordedDo(origin)
	}
	return wf.stepResultToHTTP(sr)
}

func newRoundTripper(old http.RoundTripper, wf *Workflow) http.RoundTripper {
	return &roundTripper{old: old, wf: wf}
}

// HTTPResp2DtmError check for dtm error and return it
func HTTPResp2DtmError(resp *http.Response) ([]byte, error) {
	code := resp.StatusCode
	data, err := io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewBuffer(data))
	if code == http.StatusTooEarly {
		return data, dtmcli.ErrorMessage2Error(string(data), dtmcli.ErrOngoing)
	} else if code == http.StatusConflict {
		return data, dtmcli.ErrorMessage2Error(string(data), dtmcli.ErrFailure)
	} else if err == nil && code != http.StatusOK {
		return data, errors.New(string(data))
	}
	return data, err
}

// GrpcError2DtmError translate grpc error to dtm error
func GrpcError2DtmError(err error) error {
	st, _ := status.FromError(err)
	if st != nil && st.Code() == codes.Aborted {
		return dtmcli.ErrorMessage2Error(st.Message(), dtmcli.ErrFailure)
	} else if st != nil && st.Code() == codes.FailedPrecondition {
		return dtmcli.ErrorMessage2Error(st.Message(), dtmcli.ErrOngoing)
	}
	return err
}

func (wf *Workflow) stepResultFromLocal(data []byte, err error) *stepResult {
	return &stepResult{
		Error:  err,
		Status: wfErrorToStatus(err),
		Data:   data,
	}
}

func (wf *Workflow) stepResultToLocal(sr *stepResult) ([]byte, error) {
	return sr.Data, sr.Error
}

func (wf *Workflow) stepResultFromGrpc(reply interface{}, err error) *stepResult {
	sr := &stepResult{Error: wf.Options.GRPCError2DtmError(err)}
	sr.Status = wfErrorToStatus(sr.Error)
	if sr.Error == nil {
		sr.Data = dtmgimp.MustProtoMarshal(reply.(protoreflect.ProtoMessage))
	} else if sr.Status == dtmcli.StatusFailed {
		sr.Data = []byte(err.Error())
	}
	return sr
}

func (wf *Workflow) stepResultToGrpc(s *stepResult, reply interface{}) error {
	if s.Error == nil && s.Status == dtmcli.StatusSucceed {
		dtmgimp.MustProtoUnmarshal(s.Data, reply.(protoreflect.ProtoMessage))
	}
	return s.Error
}

func (wf *Workflow) stepResultFromHTTP(resp *http.Response, err error) *stepResult {
	sr := &stepResult{Error: err}
	if err == nil {
		sr.Data, sr.Error = wf.Options.HTTPResp2DtmError(resp)
		sr.Status = wfErrorToStatus(sr.Error)
	}
	return sr
}

func (wf *Workflow) stepResultToHTTP(s *stepResult) (*http.Response, error) {
	if s.Error != nil {
		return nil, s.Error
	}
	return newJSONResponse(200, s.Data), nil
}
