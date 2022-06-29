package busi

import (
	"github.com/gin-gonic/gin"
)

// Startup startup the busi's grpc and http service
func Startup() *gin.Engine {
	svr := GrpcStartup()
	app := BaseAppStartup()
	WorkflowStarup(svr)
	go GrpcServe(svr)
	return app
}
