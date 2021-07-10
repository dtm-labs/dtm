package examples

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
)

// 启动命令：go run app/main.go qs

// 事务参与者的服务地址
const qsBusiApi = "/api/busi_start"
const qsBusiPort = 8082

var qsBusi = fmt.Sprintf("http://localhost:%d%s", qsBusiPort, qsBusiApi)

// 被app/main.go调用，启动服务并运行示例
func QuickStarMain() {
	qsStartSvr()
	qsFireRequest()
}

func qsStartSvr() {
	app := common.GetGinApp()
	qsAddRoute(app)
	logrus.Printf("quick qs examples listening at %d", qsBusiPort)
	go app.Run(fmt.Sprintf(":%d", qsBusiPort))
	time.Sleep(100 * time.Millisecond)
}

func qsFireRequest() {
	req := &gin.H{"amount": 30} // 微服务的载荷
	// DtmServer为DTM服务的地址
	saga := dtmcli.NewSaga(DtmServer).
		// 添加一个TransOut的子事务，正向操作为url: qsBusi+"/TransOut"， 逆向操作为url: qsBusi+"/TransOutCompensate"
		Add(qsBusi+"/TransOut", qsBusi+"/TransOutCompensate", req).
		// 添加一个TransIn的子事务，正向操作为url: qsBusi+"/TransOut"， 逆向操作为url: qsBusi+"/TransInCompensate"
		Add(qsBusi+"/TransIn", qsBusi+"/TransInCompensate", req)
	// 提交saga事务，dtm会完成所有的子事务/回滚所有的子事务
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
}
