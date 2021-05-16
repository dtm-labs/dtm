package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/service"
)

func main() {
	logrus.Printf("start tc")
	gin.SetMode(gin.ReleaseMode)
	app := gin.Default()
	service.AddRoute(app)
	app.Run()
}
