package dtmsvr

import (
	"github.com/gin-gonic/gin"
	"github.com/yedf/dtm/common"
)

func AddRoute(engine *gin.Engine) {
	engine.POST("/api/dtmsvr/prepare", Prepare)
	engine.POST("/api/dtmsvr/commit", Commit)
}

func Prepare(c *gin.Context) {
	data := gin.H{}
	err := c.BindJSON(&data)
	if err != nil {
		return
	}
	rabbit := RabbitmqGet()
	err = rabbit.SendAndConfirm(RabbitmqConstPrepared, data)
	common.PanicIfError(err)
	c.JSON(200, gin.H{"message": "SUCCESS"})
}

func Commit(c *gin.Context) {
	data := gin.H{}
	err := c.BindJSON(&data)
	if err != nil {
		return
	}
	rabbit := RabbitmqGet()
	err = rabbit.SendAndConfirm(RabbitmqConstCommited, data)
	common.PanicIfError(err)
	c.JSON(200, gin.H{"message": "SUCCESS"})
}
