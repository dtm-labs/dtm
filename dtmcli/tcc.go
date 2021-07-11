package dtmcli

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/yedf/dtm/common"
)

type Tcc struct {
	IDGenerator
	Dtm string
	Gid string
}

type TccGlobalFunc func(tcc *Tcc) error

func TccGlobalTransaction(dtm string, tccFunc TccGlobalFunc) (gid string, rerr error) {
	gid = GenGid(dtm)
	data := &M{
		"gid":        gid,
		"trans_type": "tcc",
	}
	defer func() {
		if x := recover(); x != nil {
			_, rerr = common.RestyClient.R().SetBody(data).Post(dtm + "/abort")
		} else {
			_, rerr = common.RestyClient.R().SetBody(data).Post(dtm + "/submit")
		}
	}()
	tcc := &Tcc{Dtm: dtm, Gid: gid}
	_, rerr = common.RestyClient.R().SetBody(data).Post(tcc.Dtm + "/prepare")
	if rerr != nil {
		return
	}
	rerr = tccFunc(tcc)
	return
}

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

func (t *Tcc) CallBranch(body interface{}, tryUrl string, confirmUrl string, cancelUrl string) (*resty.Response, error) {
	branchID := t.NewBranchID()
	resp, err := common.RestyClient.R().
		SetBody(&M{
			"gid":        t.Gid,
			"branch_id":  branchID,
			"trans_type": "tcc",
			"status":     "prepared",
			"data":       string(common.MustMarshal(body)),
			"try":        tryUrl,
			"confirm":    confirmUrl,
			"cancel":     cancelUrl,
		}).
		Post(t.Dtm + "/registerTccBranch")
	if err != nil {
		return resp, err
	}
	return common.RestyClient.R().
		SetBody(body).
		SetQueryParams(common.MS{
			"gid":         t.Gid,
			"branch_id":   branchID,
			"trans_type":  "tcc",
			"branch_type": "try",
		}).
		Post(tryUrl)
}
