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
}

func Prepare(c *gin.Context) (interface{}, error) {
	db := DbGet()
	m := getSagaModelFromContext(c)
	m.Status = "prepared"
	writeTransLog(m.Gid, "save prepared", m.Status, -1, m.Steps)
	db.Must().Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&m)
	return M{"message": "SUCCESS"}, nil
}

func Commit(c *gin.Context) (interface{}, error) {
	m := getSagaModelFromContext(c)
	saveCommitedSagaModel(m)
	go ProcessCommitedSaga(m.Gid)
	return M{"message": "SUCCESS"}, nil
}

func getSagaModelFromContext(c *gin.Context) *SagaModel {
	data := M{}
	b, err := c.GetRawData()
	common.PanicIfError(err)
	common.MustUnmarshal(b, &data)
	logrus.Printf("creating saga model in prepare")
	data["steps"] = common.MustMarshalString(data["steps"])
	m := SagaModel{}
	common.MustRemarshal(data, &m)
	return &m
}
