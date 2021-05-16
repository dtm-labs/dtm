package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gitlab.ivydad.com/dongfu.ye/go1/service"
)

func main() {
	logrus.Printf("start server")
	gin.SetMode(gin.ReleaseMode)
	app := gin.Default()
	service.AddRoute(app)
	app.Run()
}
