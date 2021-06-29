package examples

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm"
	"github.com/yedf/dtm/common"
	"gorm.io/gorm"
)

// 事务参与者的服务地址
const XaBusiPort = 8082
const XaBusiApi = "/api/busi_xa"

var XaBusi = fmt.Sprintf("http://localhost:%d%s", XaBusiPort, XaBusiApi)

var XaClient *dtm.XaClient = nil

type UserAccount struct {
	common.ModelBase
	UserId  int
	Balance string
}

func (u *UserAccount) TableName() string { return "user_account" }

func dbGet() *common.DB {
	return common.DbGet(Config.Mysql)
}

func XaMain() {
	go XaStartSvr()
	time.Sleep(100 * time.Millisecond)
	XaFireRequest()
	time.Sleep(1000 * time.Second)
}

func XaStartSvr() {
	common.InitApp(common.GetProjectDir(), &Config)
	Config.Mysql["database"] = "dtm_busi"
	logrus.Printf("xa examples starting")
	app := common.GetGinApp()
	XaClient = dtm.XaClientNew(DtmServer, Config.Mysql, app, XaBusi+"/xa")
	XaAddRoute(app)
	app.Run(fmt.Sprintf(":%d", XaBusiPort))
}

func XaFireRequest() {
	gid := common.GenGid()
	err := XaClient.XaGlobalTransaction(gid, func() (rerr error) {
		defer common.P2E(&rerr)
		req := GenTransReq(30, false, false)
		resp, err := common.RestyClient.R().SetBody(req).SetQueryParams(map[string]string{
			"gid":     gid,
			"user_id": "1",
		}).Post(XaBusi + "/TransOut")
		common.CheckRestySuccess(resp, err)
		resp, err = common.RestyClient.R().SetBody(req).SetQueryParams(map[string]string{
			"gid":     gid,
			"user_id": "2",
		}).Post(XaBusi + "/TransIn")
		common.CheckRestySuccess(resp, err)
		return nil
	})
	e2p(err)
}

// api
func XaAddRoute(app *gin.Engine) {
	app.POST(XaBusiApi+"/TransIn", common.WrapHandler(xaTransIn))
	app.POST(XaBusiApi+"/TransOut", common.WrapHandler(xaTransOut))
}

func xaTransIn(c *gin.Context) (interface{}, error) {
	err := XaClient.XaLocalTransaction(c.Query("gid"), func(db *common.DB) (rerr error) {
		req := transReqFromContext(c)
		if req.TransInResult != "SUCCESS" {
			return fmt.Errorf("tranIn failed")
		}
		dbr := db.Model(&UserAccount{}).Where("user_id = ?", c.Query("user_id")).
			Update("balance", gorm.Expr("balance + ?", req.Amount))
		return dbr.Error
	})
	e2p(err)
	return M{"result": "SUCCESS"}, nil
}

func xaTransOut(c *gin.Context) (interface{}, error) {
	err := XaClient.XaLocalTransaction(c.Query("gid"), func(db *common.DB) (rerr error) {
		req := transReqFromContext(c)
		if req.TransOutResult != "SUCCESS" {
			return fmt.Errorf("tranOut failed")
		}
		dbr := db.Model(&UserAccount{}).Where("user_id = ?", c.Query("user_id")).
			Update("balance", gorm.Expr("balance - ?", req.Amount))
		return dbr.Error
	})
	e2p(err)
	return M{"result": "SUCCESS"}, nil
}

func ResetXaData() {
	db := dbGet()
	db.Must().Exec("truncate user_account")
	db.Must().Exec("insert into user_account (user_id, balance) values (1, 10000), (2, 10000)")
	type XaRow struct {
		Data string
	}
	xas := []XaRow{}
	db.Must().Raw("xa recover").Scan(&xas)
	for _, xa := range xas {
		db.Must().Exec(fmt.Sprintf("xa rollback '%s'", xa.Data))
	}
}
