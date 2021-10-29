package dtmsvr

import (
	"errors"
	"math"

	"github.com/gin-gonic/gin"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"gorm.io/gorm"
)

func addRoute(engine *gin.Engine) {
	engine.GET("/api/dtmsvr/newGid", common.WrapHandler(newGid))
	engine.POST("/api/dtmsvr/prepare", common.WrapHandler(prepare))
	engine.POST("/api/dtmsvr/submit", common.WrapHandler(submit))
	engine.POST("/api/dtmsvr/abort", common.WrapHandler(abort))
	engine.POST("/api/dtmsvr/registerXaBranch", common.WrapHandler(registerXaBranch))
	engine.POST("/api/dtmsvr/registerTccBranch", common.WrapHandler(registerTccBranch))
	engine.GET("/api/dtmsvr/query", common.WrapHandler(query))
	engine.GET("/api/dtmsvr/all", common.WrapHandler(all))
}

func newGid(c *gin.Context) (interface{}, error) {
	return M{"gid": GenGid(), "dtm_result": dtmcli.ResultSuccess}, nil
}

func prepare(c *gin.Context) (interface{}, error) {
	return svcPrepare(TransFromContext(c))
}

func submit(c *gin.Context) (interface{}, error) {
	return svcSubmit(TransFromContext(c))
}

func abort(c *gin.Context) (interface{}, error) {
	return svcAbort(TransFromContext(c))
}

func registerXaBranch(c *gin.Context) (interface{}, error) {
	branch := TransBranch{}
	err := c.BindJSON(&branch)
	e2p(err)
	return svcRegisterXaBranch(&branch)
}

func registerTccBranch(c *gin.Context) (interface{}, error) {
	data := dtmcli.MS{}
	err := c.BindJSON(&data)
	e2p(err)
	branch := TransBranch{
		Gid:      data["gid"],
		BranchID: data["branch_id"],
		Status:   dtmcli.StatusPrepared,
		Data:     data["data"],
	}
	return svcRegisterTccBranch(&branch, data)
}

func query(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	if gid == "" {
		return nil, errors.New("no gid specified")
	}
	trans := TransGlobal{}
	db := dbGet()
	db.Begin()
	dbr := db.Must().Where("gid", gid).First(&trans)
	if dbr.Error == gorm.ErrRecordNotFound {
		return M{"transaction": nil, "branches": [0]int{}}, nil
	}
	branches := []TransBranch{}
	db.Must().Where("gid", gid).Find(&branches)
	return M{"transaction": trans, "branches": branches}, nil
}

func all(c *gin.Context) (interface{}, error) {
	lastID := c.Query("last_id")
	lid := math.MaxInt64
	if lastID != "" {
		lid = dtmcli.MustAtoi(lastID)
	}
	trans := []TransGlobal{}
	dbGet().Must().Where("id < ?", lid).Order("id desc").Limit(100).Find(&trans)
	return M{"transactions": trans}, nil
}
