package dtmcli

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
)

// M alias
type M = map[string]interface{}

var e2p = common.E2P

// XaGlobalFunc type of xa global function
type XaGlobalFunc func(xa *Xa) error

// XaLocalFunc type of xa local function
type XaLocalFunc func(db *sql.DB, xa *Xa) error

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
		db := common.DbAlone(xa.Conf)
		defer db.Close()
		branchID := req.Gid + "-" + req.BranchID
		if req.Action == "commit" {
			_, err := common.DbExec(db, fmt.Sprintf("xa commit '%s'", branchID))
			e2p(err)
		} else if req.Action == "rollback" {
			_, err := common.DbExec(db, fmt.Sprintf("xa rollback '%s'", branchID))
			e2p(err)
		} else {
			panic(fmt.Errorf("unknown action: %s", req.Action))
		}
		return M{"dtm_result": "SUCCESS"}, nil
	}))
	return xa, nil
}

// XaLocalTransaction start a xa local transaction
func (xc *XaClient) XaLocalTransaction(c *gin.Context, transFunc XaLocalFunc) (rerr error) {
	defer common.P2E(&rerr)
	xa := XaFromReq(c)
	branchID := xa.NewBranchID()
	xaBranch := xa.Gid + "-" + branchID
	db := common.DbAlone(xc.Conf)
	defer func() { db.Close() }()
	_, err := common.DbExec(db, fmt.Sprintf("XA start '%s'", xaBranch))
	e2p(err)
	err = transFunc(db, xa)
	e2p(err)
	resp, err := common.RestyClient.R().
		SetBody(&M{"gid": xa.Gid, "branch_id": branchID, "trans_type": "xa", "status": "prepared", "url": xc.CallbackURL}).
		Post(xc.Server + "/registerXaBranch")
	e2p(err)
	if !strings.Contains(resp.String(), "SUCCESS") {
		e2p(fmt.Errorf("unknown server response: %s", resp.String()))
	}
	_, err = common.DbExec(db, fmt.Sprintf("XA end '%s'", xaBranch))
	e2p(err)
	_, err = common.DbExec(db, fmt.Sprintf("XA prepare '%s'", xaBranch))
	e2p(err)
	return nil
}

// XaGlobalTransaction start a xa global transaction
func (xc *XaClient) XaGlobalTransaction(gid string, transFunc XaGlobalFunc) error {
	xa := Xa{IDGenerator: IDGenerator{}, Gid: gid}
	data := &M{
		"gid":        gid,
		"trans_type": "xa",
	}
	defer func() {
		x := recover()
		if x != nil {
			r, err := common.RestyClient.R().SetBody(data).Post(xc.Server + "/abort")
			if !strings.Contains(r.String(), "SUCCESS") {
				logrus.Errorf("abort xa error: resp: %s err: %v", r.String(), err)
			}
		}
	}()
	resp, err := common.RestyClient.R().SetBody(data).Post(xc.Server + "/prepare")
	rerr := CheckDtmResponse(resp, err)
	if rerr != nil {
		return rerr
	}
	rerr = transFunc(&xa)
	if rerr != nil {
		return rerr
	}
	resp, err = common.RestyClient.R().SetBody(data).Post(xc.Server + "/submit")
	rerr = CheckDtmResponse(resp, err)
	if rerr != nil {
		return rerr
	}
	return nil
}

// CallBranch call a xa branch
func (x *Xa) CallBranch(body interface{}, url string) (*resty.Response, error) {
	branchID := x.NewBranchID()
	resp, err := common.RestyClient.R().
		SetBody(body).
		SetQueryParams(common.MS{
			"gid":         x.Gid,
			"branch_id":   branchID,
			"trans_type":  "xa",
			"branch_type": "action",
		}).
		Post(url)
	if strings.Contains(resp.String(), "FAILURE") {
		return resp, fmt.Errorf("FAILURE result: %s err: %v", resp.String(), err)
	}
	return resp, err
}
