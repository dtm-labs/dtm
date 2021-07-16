package dtmsvr

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/yedf/dtm/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func addRoute(engine *gin.Engine) {
	engine.POST("/api/dtmsvr/prepare", common.WrapHandler(prepare))
	engine.POST("/api/dtmsvr/submit", common.WrapHandler(submit))
	engine.POST("/api/dtmsvr/registerXaBranch", common.WrapHandler(registerXaBranch))
	engine.POST("/api/dtmsvr/registerTccBranch", common.WrapHandler(registerTccBranch))
	engine.POST("/api/dtmsvr/abort", common.WrapHandler(abort))
	engine.GET("/api/dtmsvr/query", common.WrapHandler(query))
	engine.GET("/api/dtmsvr/newGid", common.WrapHandler(newGid))
}

func newGid(c *gin.Context) (interface{}, error) {
	return M{"gid": GenGid()}, nil
}

func prepare(c *gin.Context) (interface{}, error) {
	m := TransFromContext(c)
	m.Status = "prepared"
	m.saveNew(dbGet())
	return M{"message": "SUCCESS", "gid": m.Gid}, nil
}

func submit(c *gin.Context) (interface{}, error) {
	db := dbGet()
	m := TransFromContext(c)
	m.Status = "submitted"
	m.saveNew(db)
	go m.Process(db)
	return M{"message": "SUCCESS", "gid": m.Gid}, nil
}

func abort(c *gin.Context) (interface{}, error) {
	db := dbGet()
	m := TransFromContext(c)
	m = TransFromDb(db, m.Gid)
	if m.TransType != "xa" && m.TransType != "tcc" || m.Status != "prepared" {
		return nil, fmt.Errorf("unexpected trans data. type: %s status: %s for gid: %s", m.TransType, m.Status, m.Gid)
	}
	go m.Process(db)
	return M{"message": "SUCCESS"}, nil
}

func registerXaBranch(c *gin.Context) (interface{}, error) {
	branch := TransBranch{}
	err := c.BindJSON(&branch)
	e2p(err)
	branches := []TransBranch{branch, branch}
	db := dbGet()
	branches[0].BranchType = "rollback"
	branches[1].BranchType = "commit"
	db.Must().Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(branches)
	e2p(err)
	global := TransGlobal{Gid: branch.Gid}
	global.touch(db, config.TransCronInterval)
	return M{"message": "SUCCESS"}, nil
}

func registerTccBranch(c *gin.Context) (interface{}, error) {
	data := common.MS{}
	err := c.BindJSON(&data)
	e2p(err)
	branch := TransBranch{
		Gid:      data["gid"],
		BranchID: data["branch_id"],
		Status:   data["status"],
		Data:     data["data"],
	}

	branches := []TransBranch{branch, branch, branch}
	for i, b := range []string{"cancel", "confirm", "try"} {
		branches[i].BranchType = b
		branches[i].URL = data[b]
	}

	dbGet().Must().Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(branches)
	e2p(err)
	global := TransGlobal{Gid: branch.Gid}
	global.touch(dbGet(), config.TransCronInterval)
	return M{"message": "SUCCESS"}, nil
}

func query(c *gin.Context) (interface{}, error) {
	gid := c.Query("gid")
	if gid == "" {
		return nil, errors.New("no gid specified")
	}
	trans := TransGlobal{}
	db := dbGet()
	dbr := db.Must().Where("gid", gid).First(&trans)
	if dbr.Error == gorm.ErrRecordNotFound {
		return M{"transaction": nil, "branches": [0]int{}}, nil
	}
	branches := []TransBranch{}
	db.Must().Where("gid", gid).Find(&branches)
	return M{"transaction": trans, "branches": branches}, nil
}
