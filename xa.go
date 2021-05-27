package dtm

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yedf/dtm/common"
)

type M = map[string]interface{}

var e2p = common.E2P

type XaGlobalFunc func() error

type XaLocalFunc func(db *common.MyDb) error

type XaClient struct {
	Server      string
	Conf        map[string]string
	CallbackUrl string
}

func XaClientNew(server string, mysqlConf map[string]string, app *gin.Engine, callbackUrl string) *XaClient {
	xa := &XaClient{
		Server:      server,
		Conf:        mysqlConf,
		CallbackUrl: callbackUrl,
	}
	u, err := url.Parse(callbackUrl)
	e2p(err)
	app.POST(u.Path, common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		type CallbackReq struct {
			Gid    string `json:"gid"`
			Branch string `json:"branch"`
			Action string `json:"action"`
		}
		req := CallbackReq{}
		b, err := c.GetRawData()
		e2p(err)
		common.MustUnmarshal(b, &req)
		tx, my := common.DbAlone(xa.Conf)
		defer func() {
			my.Close()
		}()
		if req.Action == "commit" {
			tx.Must().Exec(fmt.Sprintf("xa commit '%s'", req.Branch))
		} else if req.Action == "rollback" {
			tx.Must().Exec(fmt.Sprintf("xa rollback '%s'", req.Branch))
		} else {
			panic(fmt.Errorf("unknown action: %s", req.Action))
		}
		return M{"result": "SUCCESS"}, nil
	}))
	return xa
}
func (xa *XaClient) XaLocalTransaction(gid string, transFunc XaLocalFunc) (rerr error) {
	defer common.P2E(&rerr)
	branch := common.GenGid()
	tx, my := common.DbAlone(xa.Conf)
	defer func() { my.Close() }()
	tx.Must().Exec(fmt.Sprintf("XA start '%s'", branch))
	err := transFunc(tx)
	e2p(err)
	resp, err := common.RestyClient.R().
		SetBody(&M{"gid": gid, "branch": branch, "trans_type": "xa", "status": "prepared", "url": xa.CallbackUrl}).
		Post(xa.Server + "/branch")
	e2p(err)
	if !strings.Contains(resp.String(), "SUCCESS") {
		e2p(fmt.Errorf("unknown server response: %s", resp.String()))
	}
	tx.Must().Exec(fmt.Sprintf("XA end '%s'", branch))
	tx.Must().Exec(fmt.Sprintf("XA prepare '%s'", branch))
	return nil
}

func (xa *XaClient) XaGlobalTransaction(gid string, transFunc XaGlobalFunc) (rerr error) {
	data := &M{
		"gid":        gid,
		"trans_type": "xa",
	}
	defer func() {
		x := recover()
		if x != nil {
			_, _ = common.RestyClient.R().SetBody(data).Post(xa.Server + "/rollback")
			rerr = x.(error)
		}
	}()
	resp, err := common.RestyClient.R().SetBody(data).Post(xa.Server + "/prepare")
	e2p(err)
	if !strings.Contains(resp.String(), "SUCCESS") {
		panic(fmt.Errorf("unexpected result: %s", resp.String()))
	}
	err = transFunc()
	e2p(err)
	resp, err = common.RestyClient.R().SetBody(data).Post(xa.Server + "/commit")
	e2p(err)
	if !strings.Contains(resp.String(), "SUCCESS") {
		panic(fmt.Errorf("unexpected result: %s", resp.String()))
	}
	return nil
}

func getDb(conf map[string]string) *common.MyDb {
	return common.DbGet(conf)
}
