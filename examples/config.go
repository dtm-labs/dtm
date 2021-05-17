package examples

import "fmt"

const TcServer = "http://localhost:8080/api/dtmsvr"
const BusiPort = 8081
const BusiApi = "/api/busi"

var Busi = fmt.Sprintf("http://localhost:%d%s", BusiPort, BusiApi)
