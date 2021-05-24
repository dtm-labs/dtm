package examples

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
)

// 事务参与者的服务地址
const XaBusiPort = 8082
const XaBusiApi = "/api/busi_xa"

var XaBusi = fmt.Sprintf("http://localhost:%d%s", XaBusiPort, XaBusiApi)

func XaMain() {
	go XaStartSvr()
	xaFireRequest()
	time.Sleep(1000 * time.Second)
}

func XaStartSvr() {
	logrus.Printf("xa examples starting")
	app := common.GetGinApp()
	AddRoute(app)
	app.Run(":8081")
}

func xaFireRequest() {

}

// api
