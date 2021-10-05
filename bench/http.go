package bench

import (
	"database/sql"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/examples"
)

// 启动命令：go run app/main.go qs

// 事务参与者的服务地址
const benchAPI = "/api/busi_bench"
const benchPort = 8083
const total = 1000000

var benchBusi = fmt.Sprintf("http://localhost:%d%s", benchPort, benchAPI)

func sdbGet() *sql.DB {
	db, err := dtmcli.PooledDB(common.DtmConfig.DB)
	dtmcli.FatalIfError(err)
	return db
}

func txGet() *sql.Tx {
	db := sdbGet()
	tx, err := db.Begin()
	dtmcli.FatalIfError(err)
	return tx
}

func reloadData() {
	began := time.Now()
	db := sdbGet()
	tables := []string{"dtm_busi.user_account", "dtm.trans_global", "dtm.trans_branch"}
	for _, t := range tables {
		dtmcli.DBExec(db, fmt.Sprintf("truncate %s", t))
	}
	s := "insert ignore into dtm_busi.user_account(user_id, balance) values "
	ss := []string{}
	for i := 1; i <= total; i++ {
		ss = append(ss, fmt.Sprintf("(%d, 1000000)", i))
	}
	db.Exec(s + strings.Join(ss, ","))
	dtmcli.Logf("%d users inserted. used: %dms", total, time.Since(began).Milliseconds())
}

var uidCounter int32 = 0
var mode string = ""

// StartSvr 1
func StartSvr() {
	app := common.GetGinApp()
	benchAddRoute(app)
	dtmcli.Logf("bench listening at %d", benchPort)
	reloadData()
	go app.Run(fmt.Sprintf(":%d", benchPort))
	time.Sleep(100 * time.Millisecond)
}

func qsAdjustBalance(uid int, amount int) (interface{}, error) {
	if strings.Contains(mode, "empty") {
		return dtmcli.MapSuccess, nil
	} else {
		tx := txGet()
		for i := 0; i < 5; i++ {
			_, err := dtmcli.DBExec(tx, "update dtm_busi.user_account set balance = balance + ? where user_id = ?", amount, uid)
			dtmcli.FatalIfError(err)
		}
		err := tx.Commit()
		dtmcli.FatalIfError(err)
	}

	return dtmcli.MapSuccess, nil
}

func benchAddRoute(app *gin.Engine) {
	app.POST(benchAPI+"/TransIn", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return qsAdjustBalance(dtmcli.MustAtoi(c.Query("uid")), 1)
	}))
	app.POST(benchAPI+"/TransInCompensate", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return qsAdjustBalance(dtmcli.MustAtoi(c.Query("uid")), -1)
	}))
	app.POST(benchAPI+"/TransOut", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return qsAdjustBalance(dtmcli.MustAtoi(c.Query("uid")), -1)
	}))
	app.POST(benchAPI+"/TransOutCompensate", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return qsAdjustBalance(dtmcli.MustAtoi(c.Query("uid")), 30)
	}))
	app.Any(benchAPI+"/reloadData", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		reloadData()
		mode = c.Query("m")
		return nil, nil
	}))
	app.Any(benchAPI+"/bench", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		uid := (atomic.AddInt32(&uidCounter, 1)-1)%total + 1
		suid := fmt.Sprintf("%d", uid)
		suid2 := fmt.Sprintf("%d", total+1-uid)
		req := gin.H{}
		params := fmt.Sprintf("?uid=%s", suid)
		params2 := fmt.Sprintf("?uid=%s", suid2)
		dtmcli.Logf("mode: %s contains dtm: %t", mode, strings.Contains(mode, "dtm"))
		if strings.Contains(mode, "dtm") {
			saga := dtmcli.NewSaga(examples.DtmServer, fmt.Sprintf("bench-%d", uid)).
				Add(benchBusi+"/TransOut"+params, benchBusi+"/TransOutCompensate"+params, req).
				Add(benchBusi+"/TransIn"+params2, benchBusi+"/TransInCompensate"+params2, req)
			saga.WaitResult = true
			err := saga.Submit()
			dtmcli.FatalIfError(err)
		} else {
			_, err := dtmcli.RestyClient.R().SetBody(gin.H{}).SetQueryParam("uid", suid2).Post(benchBusi + "/TransOut")
			dtmcli.FatalIfError(err)
			_, err = dtmcli.RestyClient.R().SetBody(gin.H{}).SetQueryParam("uid", suid).Post(benchBusi + "/TransIn")
			dtmcli.FatalIfError(err)
		}
		return nil, nil
	}))
}
