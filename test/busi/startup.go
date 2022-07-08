package busi

import (
	"github.com/gin-gonic/gin"
	grpc "google.golang.org/grpc"
)

// Startup startup the busi's grpc and http service
func Startup() (*gin.Engine, *grpc.Server) {
	svr := GrpcStartup()
	app := BaseAppStartup()
	return app, svr
}
