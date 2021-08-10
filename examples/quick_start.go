package examples

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
)

// 启动命令：go run app/main.go qs

// 事务参与者的服务地址
const qsBusiAPI = "/api/busi_start"

var (
	qsBusiPort = 8082
	qsBusi     string
	srv        *http.Server
)

// QsStartSvr 1
func QsStartSvr(port int) {
	if port == 0 {
		port = qsBusiPort
	}
	qsBusi = fmt.Sprintf("http://localhost:%d%s", port, qsBusiAPI)
	app := common.GetGinApp()
	qsAddRoute(app)
	srv = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: app,
	}
	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			dtmcli.Logf("quick qs examples listening at: %d,err:%s", port, err)
		}
	}()
}

func StopExampleSvr() {
	dtmcli.Logf("shutdown examples server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		dtmcli.LogFatalf("examples server shutdown:", err)
	}
	dtmcli.Logf("examples server exiting")
}

// QsFireRequest 1
func QsFireRequest() string {
	req := &gin.H{"amount": 30} // 微服务的载荷
	// DtmServer为DTM服务的地址
	saga := dtmcli.NewSaga(DtmServer, dtmcli.MustGenGid(DtmServer)).
		// 添加一个TransOut的子事务，正向操作为url: qsBusi+"/TransOut"， 逆向操作为url: qsBusi+"/TransOutCompensate"
		Add(qsBusi+"/TransOut", qsBusi+"/TransOutCompensate", req).
		// 添加一个TransIn的子事务，正向操作为url: qsBusi+"/TransOut"， 逆向操作为url: qsBusi+"/TransInCompensate"
		Add(qsBusi+"/TransIn", qsBusi+"/TransInCompensate", req)
	// 提交saga事务，dtm会完成所有的子事务/回滚所有的子事务
	err := saga.Submit()
	e2p(err)
	return saga.Gid
}

func qsAdjustBalance(uid int, amount int) (interface{}, error) {
	_, err := dtmcli.SdbExec(sdbGet(), "update dtm_busi.user_account set balance = balance + ? where user_id = ?", amount, uid)
	return dtmcli.ResultSuccess, err
}

func qsAddRoute(app *gin.Engine) {
	app.POST(qsBusiAPI+"/TransIn", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return qsAdjustBalance(2, 30)
	}))
	app.POST(qsBusiAPI+"/TransInCompensate", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return qsAdjustBalance(2, -30)
	}))
	app.POST(qsBusiAPI+"/TransOut", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return qsAdjustBalance(1, -30)
	}))
	app.POST(qsBusiAPI+"/TransOutCompensate", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return qsAdjustBalance(1, 30)
	}))
}
