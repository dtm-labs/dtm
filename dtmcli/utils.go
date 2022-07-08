package dtmcli

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/go-resty/resty/v2"
)

// MustGenGid generate a new gid
func MustGenGid(server string) string {
	res := map[string]string{}
	resp, err := dtmimp.RestyClient.R().SetResult(&res).Get(server + "/newGid")
	if err != nil || res["gid"] == "" {
		panic(fmt.Errorf("newGid error: %v, resp: %s", err, resp))
	}
	return res["gid"]
}

// String2DtmError translate string to dtm error
func String2DtmError(str string) error {
	return map[string]error{
		ResultFailure: ErrFailure,
		ResultOngoing: ErrOngoing,
		ResultSuccess: nil,
		"":            nil,
	}[str]
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

// IsRollback returns whether the result is indicating rollback
func IsRollback(resp *resty.Response, err error) bool {
	return err == ErrFailure || dtmimp.RespAsErrorCompatible(resp) == ErrFailure
}

// IsOngoing returns whether the result is indicating ongoing
func IsOngoing(resp *resty.Response, err error) bool {
	return err == ErrOngoing || dtmimp.RespAsErrorCompatible(resp) == ErrOngoing
}
