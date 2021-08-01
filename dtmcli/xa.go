package dtmcli

import (
	"database/sql"
	"fmt"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/yedf/dtm/common"
)

// M alias
type M = map[string]interface{}

var e2p = common.E2P

// XaGlobalFunc type of xa global function
type XaGlobalFunc func(xa *Xa) (interface{}, error)

// XaLocalFunc type of xa local function
type XaLocalFunc func(db *sql.DB, xa *Xa) (interface{}, error)

// XaClient xa client
type XaClient struct {
	Server      string
	Conf        map[string]string
	CallbackURL string
}

// Xa xa transaction
type Xa struct {
	IDGenerator
	Gid string
}

// XaFromReq construct xa info from request
func XaFromReq(c *gin.Context) *Xa {
	return &Xa{
		Gid:         c.Query("gid"),
		IDGenerator: IDGenerator{parentID: c.Query("branch_id")},
	}
}

// NewXaClient construct a xa client
func NewXaClient(server string, mysqlConf map[string]string, app *gin.Engine, callbackURL string) (*XaClient, error) {
	xa := &XaClient{
		Server:      server,
		Conf:        mysqlConf,
		CallbackURL: callbackURL,
	}
	u, err := url.Parse(callbackURL)
	if err != nil {
		return nil, err
	}
	app.POST(u.Path, common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		type CallbackReq struct {
			Gid      string `json:"gid"`
			BranchID string `json:"branch_id"`
			Action   string `json:"action"`
		}
		req := CallbackReq{}
		b, err := c.GetRawData()
		if err != nil {
			return nil, err
		}
		common.MustUnmarshal(b, &req)
		db := common.SdbAlone(xa.Conf)
		defer db.Close()
		branchID := req.Gid + "-" + req.BranchID
		if req.Action == "commit" {
			_, err = common.SdbExec(db, fmt.Sprintf("xa commit '%s'", branchID))
		} else if req.Action == "rollback" {
			_, err = common.SdbExec(db, fmt.Sprintf("xa rollback '%s'", branchID))
		} else {
			panic(fmt.Errorf("unknown action: %s", req.Action))
		}
		return M{"dtm_result": "SUCCESS"}, err
	}))
	return xa, nil
}

// XaLocalTransaction start a xa local transaction
func (xc *XaClient) XaLocalTransaction(c *gin.Context, xaFunc XaLocalFunc) (ret interface{}, rerr error) {
	xa := XaFromReq(c)
	branchID := xa.NewBranchID()
	xaBranch := xa.Gid + "-" + branchID
	db := common.SdbAlone(xc.Conf)
	defer func() { db.Close() }()
	defer func() {
		var x interface{}
		_, err := common.SdbExec(db, fmt.Sprintf("XA end '%s'", xaBranch))
		if err != nil {
			common.RedLogf("sql db exec error: %v", err)
		}
		if x = recover(); x != nil || IsFailure(ret, rerr) {
		} else {
			_, err = common.SdbExec(db, fmt.Sprintf("XA prepare '%s'", xaBranch))
		}
		if err != nil {
			common.RedLogf("sql db exec error: %v", err)
		}
		if x != nil {
			panic(x)
		}
	}()
	_, rerr = common.SdbExec(db, fmt.Sprintf("XA start '%s'", xaBranch))
	if rerr != nil {
		return
	}
	ret, rerr = xaFunc(db, xa)
	if IsFailure(ret, rerr) {
		return
	}
	ret, rerr = common.RestyClient.R().
		SetBody(&M{"gid": xa.Gid, "branch_id": branchID, "trans_type": "xa", "status": "prepared", "url": xc.CallbackURL}).
		Post(xc.Server + "/registerXaBranch")
	return
}

// XaGlobalTransaction start a xa global transaction
func (xc *XaClient) XaGlobalTransaction(gid string, xaFunc XaGlobalFunc) (ret interface{}, rerr error) {
	xa := Xa{IDGenerator: IDGenerator{}, Gid: gid}
	data := &M{
		"gid":        gid,
		"trans_type": "xa",
	}
	resp, err := common.RestyClient.R().SetBody(data).Post(xc.Server + "/prepare")
	if IsFailure(resp, err) {
		return resp, err
	}
	// 小概率情况下，prepare成功了，但是由于网络状况导致上面Failure，那么不执行下面defer的内容，等待超时后再回滚标记事务失败，也没有问题
	defer func() {
		var x interface{}
		if x = recover(); x != nil || IsFailure(ret, rerr) {
			resp, err = common.RestyClient.R().SetBody(data).Post(xc.Server + "/abort")
		} else {
			resp, err = common.RestyClient.R().SetBody(data).Post(xc.Server + "/submit")
		}
		if IsFailure(resp, err) {
			common.RedLogf("submitting or abort global transaction error: %v resp: %s", err, resp.String())
		}
		if x != nil {
			panic(x)
		}
	}()
	ret, rerr = xaFunc(&xa)
	return
}

// CallBranch call a xa branch
func (x *Xa) CallBranch(body interface{}, url string) (*resty.Response, error) {
	branchID := x.NewBranchID()
	return common.RestyClient.R().
		SetBody(body).
		SetQueryParams(common.MS{
			"gid":         x.Gid,
			"branch_id":   branchID,
			"trans_type":  "xa",
			"branch_type": "action",
		}).
		Post(url)
}
