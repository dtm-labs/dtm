package dtmsvr

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
	"gorm.io/gorm/clause"
)

func AddRoute(engine *gin.Engine) {
	engine.POST("/api/dtmsvr/prepare", common.WrapHandler(Prepare))
	engine.POST("/api/dtmsvr/commit", common.WrapHandler(Commit))
	engine.POST("/api/dtmsvr/branch", common.WrapHandler(Branch))
}

func Prepare(c *gin.Context) (interface{}, error) {
	db := dbGet()
	m := getTransFromContext(c)
	m.Status = "prepared"
	writeTransLog(m.Gid, "save prepared", m.Status, "", m.Data)
	db.Must().Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&m)
	return M{"message": "SUCCESS"}, nil
}

func Commit(c *gin.Context) (interface{}, error) {
	m := getTransFromContext(c)
	saveCommitted(m)
	go ProcessCommitted(m)
	return M{"message": "SUCCESS"}, nil
}

func Branch(c *gin.Context) (interface{}, error) {
	branch := TransBranchModel{}
	err := c.BindJSON(&branch)
	common.PanicIfError(err)
	db := dbGet()
	db.Must().Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&branch)
	return M{"message": "SUCCESS"}, nil
}

func getTransFromContext(c *gin.Context) *TransGlobalModel {
	data := M{}
	b, err := c.GetRawData()
	common.PanicIfError(err)
	common.MustUnmarshal(b, &data)
	logrus.Printf("creating trans model in prepare")
	if data["trans_type"].(string) == "saga" {
		data["data"] = common.MustMarshalString(data["steps"])
	}
	m := TransGlobalModel{}
	common.MustRemarshal(data, &m)
	return &m
}
