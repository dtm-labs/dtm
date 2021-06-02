package dtmsvr

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/yedf/dtm/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func AddRoute(engine *gin.Engine) {
	engine.POST("/api/dtmsvr/prepare", common.WrapHandler(Prepare))
	engine.POST("/api/dtmsvr/commit", common.WrapHandler(Commit))
	engine.POST("/api/dtmsvr/branch", common.WrapHandler(Branch))
	engine.POST("/api/dtmsvr/rollback", common.WrapHandler(Rollback))
	engine.GET("/api/dtmsvr/query", common.WrapHandler(Query))
}

func Prepare(c *gin.Context) (interface{}, error) {
	m := TransFromContext(c)
	m.Status = "prepared"
	m.SaveNew(dbGet())
	return M{"message": "SUCCESS"}, nil
}

func Commit(c *gin.Context) (interface{}, error) {
	db := dbGet()
	m := TransFromContext(c)
	m.Status = "committed"
	m.SaveNew(db)
	go m.Process(db)
	return M{"message": "SUCCESS"}, nil
}

func Rollback(c *gin.Context) (interface{}, error) {
	db := dbGet()
	m := TransFromContext(c)
	m = TransFromDb(db, m.Gid)
	if m.TransType != "xa" || m.Status != "prepared" {
		return nil, fmt.Errorf("unkown trans data. type: %s status: %s for gid: %s", m.TransType, m.Status, m.Gid)
	}
	// 当前xa trans的状态为prepared，直接处理，则是回滚
	go m.Process(db)
	return M{"message": "SUCCESS"}, nil
}

func Branch(c *gin.Context) (interface{}, error) {
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
	return M{"message": "SUCCESS"}, nil
}

func Query(c *gin.Context) (interface{}, error) {
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
