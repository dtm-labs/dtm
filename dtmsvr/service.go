package dtmsvr

import "github.com/gin-gonic/gin"

func AddRoute(engine *gin.Engine) {
	route := engine.Group("/api/dmtsvr")
	route.POST("/prepare", Prepare)
	route.POST("/commit", Commit)
}

func Prepare(c *gin.Context) {
	data := gin.H{}
	err := c.BindJSON(&data)
	if err == nil {
		return
	}

}

func Commit(c *gin.Context) {

}
