package examples

import "fmt"

func Startup() {
	InitConfig()
	GrpcStartup()
	BaseAppStartup()
}

func InitConfig() {
	DtmHttpServer = fmt.Sprintf("http://localhost:%d/api/dtmsvr", config.HttpPort)
	DtmGrpcServer = fmt.Sprintf("localhost:%d", config.GrpcPort)
}
