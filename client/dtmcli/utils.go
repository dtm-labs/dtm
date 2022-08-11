package dtmcli

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
	"github.com/go-resty/resty/v2"
)

// MustGenGid generate a new gid
func MustGenGid(server string) string {
	res := map[string]string{}
	resp, err := GetRestyClient().R().SetResult(&res).Get(server + "/newGid")
	if err != nil || res["gid"] == "" {
		panic(fmt.Errorf("newGid error: %v, resp: %s", err, resp))
	}
	return res["gid"]
}

// ErrorMessage2Error return an error fmt.Errorf("%s %w", errMsg, err) but trim out duplicate wrap
// eg. ErrorMessage2Error("an error. FAILURE", ErrFailure) return an error with message: "an error. FAILURE",
// no additional " FAILURE" added
func ErrorMessage2Error(errMsg string, err error) error {
	errMsg = strings.TrimSuffix(errMsg, " "+err.Error())
	return fmt.Errorf("%s %w", errMsg, err)
}

// HTTPResp2DtmError translate a resty response to error
// compatible with version < v1.10
func HTTPResp2DtmError(resp *resty.Response) error {
	code := resp.StatusCode()
	str := resp.String()
	if code == http.StatusTooEarly || strings.Contains(str, ResultOngoing) {
		return ErrorMessage2Error(str, ErrOngoing)
	} else if code == http.StatusConflict || strings.Contains(str, ResultFailure) {
		return ErrorMessage2Error(str, ErrFailure)
	} else if code != http.StatusOK {
		return errors.New(str)
	}
	return nil
}

// Result2HttpJSON return the http code and json result
// if result is error, the return proper code, else return StatusOK
func Result2HttpJSON(result interface{}) (code int, res interface{}) {
	err, _ := result.(error)
	if err == nil {
		code = http.StatusOK
		res = result
	} else {
		res = map[string]string{
			"error": err.Error(),
		}
		if errors.Is(err, ErrFailure) {
			code = http.StatusConflict
		} else if errors.Is(err, ErrOngoing) {
			code = http.StatusTooEarly
		} else if err != nil {
			code = http.StatusInternalServerError
		}
	}
	return
}

func requestBranch(t *dtmimp.TransBase, method string, body interface{}, branchID string, op string, url string) (*resty.Response, error) {
	resp, err := dtmimp.TransRequestBranch(t, method, body, branchID, op, url)
	if err == nil {
		err = HTTPResp2DtmError(resp)
	}
	return resp, err
}
