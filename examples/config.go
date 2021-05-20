package examples

import "fmt"

// 指定dtm服务地址
const DtmServer = "http://localhost:8080/api/dtmsvr"

// 事务参与制的服务地址
const BusiPort = 8081
const BusiApi = "/api/busi"

var Busi = fmt.Sprintf("http://localhost:%d%s", BusiPort, BusiApi)
