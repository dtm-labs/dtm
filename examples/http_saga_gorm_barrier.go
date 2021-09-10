package examples

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
)

func init() {
	setupFuncs["SagaGormBarrierSetup"] = func(app *gin.Engine) {
		app.POST(BusiAPI+"/SagaBTransOutGorm", common.WrapHandler(sagaGormBarrierTransOut))
	}
	addSample("saga_gorm_barrier", func() string {
		dtmcli.Logf("a busi transaction begin")
		req := &TransReq{Amount: 30}
		saga := dtmcli.NewSaga(DtmServer, dtmcli.MustGenGid(DtmServer)).
			Add(Busi+"/SagaBTransOutGorm", Busi+"/SagaBTransOutCompensate", req).
			Add(Busi+"/SagaBTransIn", Busi+"/SagaBTransInCompensate", req)
		dtmcli.Logf("busi trans submit")
		err := saga.Submit()
		dtmcli.FatalIfError(err)
		return saga.Gid
	})

}

func sagaGormBarrierTransOut(c *gin.Context) (interface{}, error) {
	req := reqFrom(c)
	barrier := MustBarrierFromGin(c)
	tx := dbGet().DB.Begin()
	return dtmcli.MapSuccess, barrier.Call(tx.Statement.ConnPool.(*sql.Tx), func(db dtmcli.DB) error {
		return tx.Exec("update dtm_busi.user_account set balance = balance + ? where user_id = ?", -req.Amount, 2).Error
	})
}
