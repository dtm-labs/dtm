package dtmcli

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
)

// Tcc struct of tcc
type Tcc struct {
	IDGenerator
	Dtm string
	Gid string
}

// TccGlobalFunc type of global tcc call
type TccGlobalFunc func(tcc *Tcc) error

// TccGlobalTransaction begin a tcc global transaction
func TccGlobalTransaction(dtm string, gid string, tccFunc TccGlobalFunc) (rerr error) {
	data := &M{
		"gid":        gid,
		"trans_type": "tcc",
	}
	defer func() {
		var resp *resty.Response
		var err error
		var x interface{}
		if x = recover(); x != nil || rerr != nil {
			resp, err = common.RestyClient.R().SetBody(data).Post(dtm + "/abort")
		} else {
			resp, err = common.RestyClient.R().SetBody(data).Post(dtm + "/submit")
		}
		err2 := CheckDtmResponse(resp, err)
		if err2 != nil {
			logrus.Errorf("submitting or abort global transaction error: %v", err2)
		}
		if x != nil {
			panic(x)
		}
	}()
	tcc := &Tcc{Dtm: dtm, Gid: gid}
	resp, err := common.RestyClient.R().SetBody(data).Post(tcc.Dtm + "/prepare")
	rerr = CheckDtmResponse(resp, err)
	if rerr != nil {
		return
	}
	rerr = tccFunc(tcc)
	return
}

// TccFromReq tcc from request info
func TccFromReq(c *gin.Context) (*Tcc, error) {
	tcc := &Tcc{
		Dtm:         c.Query("dtm"),
		Gid:         c.Query("gid"),
		IDGenerator: IDGenerator{parentID: c.Query("branch_id")},
	}
	if tcc.Dtm == "" || tcc.Gid == "" {
		return nil, fmt.Errorf("bad tcc info. dtm: %s, gid: %s", tcc.Dtm, tcc.Gid)
	}
	return tcc, nil
}

// CallBranch call a tcc branch
func (t *Tcc) CallBranch(body interface{}, tryURL string, confirmURL string, cancelURL string) (*resty.Response, error) {
	branchID := t.NewBranchID()
	resp, err := common.RestyClient.R().
		SetBody(&M{
			"gid":        t.Gid,
			"branch_id":  branchID,
			"trans_type": "tcc",
			"status":     "prepared",
			"data":       string(common.MustMarshal(body)),
			"try":        tryURL,
			"confirm":    confirmURL,
			"cancel":     cancelURL,
		}).
		Post(t.Dtm + "/registerTccBranch")
	if err != nil {
		return resp, err
	}
	if !strings.Contains(resp.String(), "SUCCESS") {
		return nil, fmt.Errorf("registerTccBranch failed: %s", resp.String())
	}
	r, err := common.RestyClient.R().
		SetBody(body).
		SetQueryParams(common.MS{
			"dtm":         t.Dtm,
			"gid":         t.Gid,
			"branch_id":   branchID,
			"trans_type":  "tcc",
			"branch_type": "try",
		}).
		Post(tryURL)
	if err == nil && strings.Contains(r.String(), "FAILURE") {
		return r, fmt.Errorf("branch return failure: %s", r.String())
	}
	return r, err
}
