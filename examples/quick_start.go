package examples

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
)

// 事务参与者的服务地址
const qsBusiApi = "/api/busi_start"
const qsBusiPort = 8082

var qsBusi = fmt.Sprintf("http://localhost:%d%s", qsBusiPort, qsBusiApi)

func StartMain() {
	go qsStartSvr()
	qsFireRequest()
	time.Sleep(1000 * time.Second)
}

func qsStartSvr() {
	logrus.Printf("quick start examples starting")
	app := common.GetGinApp()
	qsAddRoute(app)
	app.Run(fmt.Sprintf(":%d", qsBusiPort))
}

func qsFireRequest() {
	req := &gin.H{"amount": 30}
	saga := dtmcli.NewSaga(DtmServer).
		Add(qsBusi+"/TransOut", qsBusi+"/TransOutCompensate", req).
		Add(qsBusi+"/TransIn", qsBusi+"/TransInCompensate", req)
	err := saga.Submit()
	e2p(err)
}

func qsAddRoute(app *gin.Engine) {
	app.POST(qsBusiApi+"/TransIn", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return M{"result": "SUCCESS"}, nil
	}))
	app.POST(qsBusiApi+"/TransInCompensate", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return M{"result": "SUCCESS"}, nil
	}))
	app.POST(qsBusiApi+"/TransOut", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return M{"result": "SUCCESS"}, nil
	}))
	app.POST(qsBusiApi+"/TransOutCompensate", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return M{"result": "SUCCESS"}, nil
	}))
	logrus.Printf("quick qs examples listening at %d", qsBusiPort)
}
