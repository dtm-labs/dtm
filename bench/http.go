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
const total = 200000

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
	_, err := dtmcli.DBExec(db, "drop table if exists dtm_busi.user_account_log")
	dtmcli.FatalIfError(err)
	_, err = dtmcli.DBExec(db, `create table if not exists dtm_busi.user_account_log (
	id      INT(11) AUTO_INCREMENT PRIMARY KEY,
	user_id INT(11) NOT NULL,
	delta DECIMAL(11, 2) not null,
	gid varchar(45) not null,
	branch_id varchar(45) not null,
	branch_type varchar(45) not null,
	reason varchar(45),
	create_time datetime not null default now(),
	update_time datetime not null default now(),
	key(user_id),
	key(create_time)
)
`)
	dtmcli.FatalIfError(err)
	tables := []string{"dtm_busi.user_account", "dtm_busi.user_account_log", "dtm.trans_global", "dtm.trans_branch", "dtm_barrier.barrier"}
	for _, t := range tables {
		_, err = dtmcli.DBExec(db, fmt.Sprintf("truncate %s", t))
		dtmcli.FatalIfError(err)
	}
	s := "insert ignore into dtm_busi.user_account(user_id, balance) values "
	ss := []string{}
	for i := 1; i <= total; i++ {
		ss = append(ss, fmt.Sprintf("(%d, 1000000)", i))
	}
	_, err = db.Exec(s + strings.Join(ss, ","))
	dtmcli.FatalIfError(err)
	dtmcli.Logf("%d users inserted. used: %dms", total, time.Since(began).Milliseconds())
}

var uidCounter int32 = 0
var mode string = ""
var sqls int = 1

// StartSvr 1
func StartSvr() {
	app := common.GetGinApp()
	benchAddRoute(app)
	dtmcli.Logf("bench listening at %d", benchPort)
	go app.Run(fmt.Sprintf(":%d", benchPort))
	reloadData()
	time.Sleep(1100 * time.Millisecond) // sleep 1 second for async branch status update to finish
}

func qsAdjustBalance(uid int, amount int, c *gin.Context) (interface{}, error) {
	if strings.Contains(mode, "empty") {
		return dtmcli.MapSuccess, nil
	}
	tb := dtmcli.TransBaseFromQuery(c.Request.URL.Query())
	f := func(tx dtmcli.DB) error {
		for i := 0; i < sqls; i++ {
			_, err := dtmcli.DBExec(tx, "insert into dtm_busi.user_account_log(user_id, delta, gid, branch_id, branch_type, reason)  values(?,?,?,?,?,?)",
				uid, amount, tb.Gid, c.Query("branch_id"), tb.TransType, fmt.Sprintf("inserted by dtm transaction %s %s", tb.Gid, c.Query("branch_id")))
			dtmcli.FatalIfError(err)
			_, err = dtmcli.DBExec(tx, "update dtm_busi.user_account set balance = balance + ?, update_time = now() where user_id = ?", amount, uid)
			dtmcli.FatalIfError(err)
		}
		return nil
	}
	if strings.Contains(mode, "barrier") {
		barrier, err := dtmcli.BarrierFromQuery(c.Request.URL.Query())
		dtmcli.FatalIfError(err)
		barrier.Call(txGet(), f)
	} else {
		tx := txGet()
		f(tx)
		err := tx.Commit()
		dtmcli.FatalIfError(err)
	}

	return dtmcli.MapSuccess, nil
}

func benchAddRoute(app *gin.Engine) {
	app.POST(benchAPI+"/TransIn", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return qsAdjustBalance(dtmcli.MustAtoi(c.Query("uid")), 1, c)
	}))
	app.POST(benchAPI+"/TransInCompensate", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return qsAdjustBalance(dtmcli.MustAtoi(c.Query("uid")), -1, c)
	}))
	app.POST(benchAPI+"/TransOut", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return qsAdjustBalance(dtmcli.MustAtoi(c.Query("uid")), -1, c)
	}))
	app.POST(benchAPI+"/TransOutCompensate", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return qsAdjustBalance(dtmcli.MustAtoi(c.Query("uid")), 30, c)
	}))
	app.Any(benchAPI+"/reloadData", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		reloadData()
		mode = c.Query("m")
		s := c.Query("sqls")
		if s != "" {
			sqls = dtmcli.MustAtoi(s)
		}
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
			dtmcli.E2P(err)
		} else {
			_, err := dtmcli.RestyClient.R().SetBody(gin.H{}).SetQueryParam("uid", suid2).Post(benchBusi + "/TransOut")
			dtmcli.E2P(err)
			_, err = dtmcli.RestyClient.R().SetBody(gin.H{}).SetQueryParam("uid", suid).Post(benchBusi + "/TransIn")
			dtmcli.E2P(err)
		}
		return nil, nil
	}))
}
