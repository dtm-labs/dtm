package examples

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	// BusiAPI busi api prefix
	BusiAPI = "/api/busi"
	// BusiPort busi server port
	BusiPort = 8081
	// BusiGrpcPort busi server port
	BusiGrpcPort = 58081
)

type setupFunc func(*gin.Engine)

var setupFuncs = map[string]setupFunc{}

// Busi busi service url prefix
var Busi string = fmt.Sprintf("http://localhost:%d%s", BusiPort, BusiAPI)

// BaseAppStartup base app startup
func BaseAppStartup() *gin.Engine {
	dtmcli.Logf("examples starting")
	app := common.GetGinApp()
	BaseAddRoute(app)
	for k, v := range setupFuncs {
		dtmcli.Logf("initing %s", k)
		v(app)
	}
	dtmcli.Logf("Starting busi at: %d", BusiPort)
	go app.Run(fmt.Sprintf(":%d", BusiPort))

	time.Sleep(100 * time.Millisecond)
	return app
}

// AutoEmptyString auto reset to empty when used once
type AutoEmptyString struct {
	value string
}

// SetOnce set a value once
func (s *AutoEmptyString) SetOnce(v string) {
	s.value = v
}

// Fetch fetch the stored value, then reset the value to empty
func (s *AutoEmptyString) Fetch() string {
	v := s.value
	s.value = ""
	return v
}

type mainSwitchType struct {
	TransInResult         AutoEmptyString
	TransOutResult        AutoEmptyString
	TransInConfirmResult  AutoEmptyString
	TransOutConfirmResult AutoEmptyString
	TransInRevertResult   AutoEmptyString
	TransOutRevertResult  AutoEmptyString
	CanSubmitResult       AutoEmptyString
}

// MainSwitch controls busi success or fail
var MainSwitch mainSwitchType

func handleGeneralBusiness(c *gin.Context, result1 string, result2 string, busi string) (interface{}, error) {
	info := infoFromContext(c)
	res := dtmcli.OrString(result1, result2, dtmcli.ResultSuccess)
	dtmcli.Logf("%s %s result: %s", busi, info.String(), res)
	if res == "ERROR" {
		return nil, errors.New("ERROR from user")
	}
	return M{"dtm_result": res}, nil
}

// BaseAddRoute add base route handler
func BaseAddRoute(app *gin.Engine) {
	app.POST(BusiAPI+"/TransIn", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return handleGeneralBusiness(c, MainSwitch.TransInResult.Fetch(), reqFrom(c).TransInResult, "transIn")
	}))
	app.POST(BusiAPI+"/TransOut", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return handleGeneralBusiness(c, MainSwitch.TransOutResult.Fetch(), reqFrom(c).TransOutResult, "TransOut")
	}))
	app.POST(BusiAPI+"/TransInConfirm", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return handleGeneralBusiness(c, MainSwitch.TransInConfirmResult.Fetch(), "", "TransInConfirm")
	}))
	app.POST(BusiAPI+"/TransOutConfirm", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return handleGeneralBusiness(c, MainSwitch.TransOutConfirmResult.Fetch(), "", "TransOutConfirm")
	}))
	app.POST(BusiAPI+"/TransInRevert", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return handleGeneralBusiness(c, MainSwitch.TransInRevertResult.Fetch(), "", "TransInRevert")
	}))
	app.POST(BusiAPI+"/TransOutRevert", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return handleGeneralBusiness(c, MainSwitch.TransOutRevertResult.Fetch(), "", "TransOutRevert")
	}))
	app.GET(BusiAPI+"/CanSubmit", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		dtmcli.Logf("%s CanSubmit", c.Query("gid"))
		return dtmcli.OrString(MainSwitch.CanSubmitResult.Fetch(), dtmcli.ResultSuccess), nil
	}))
	app.POST(BusiAPI+"/TransInXa", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return XaClient.XaLocalTransaction(c.Request.URL.Query(), func(db *sql.DB, xa *dtmcli.Xa) (interface{}, error) {
			if reqFrom(c).TransInResult == dtmcli.ResultFailure {
				return dtmcli.MapFailure, nil
			}
			_, err := dtmcli.DBExec(db, "update dtm_busi.user_account set balance=balance+? where user_id=?", reqFrom(c).Amount, 2)
			return dtmcli.MapSuccess, err
		})
	}))
	app.POST(BusiAPI+"/TransOutXa", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return XaClient.XaLocalTransaction(c.Request.URL.Query(), func(db *sql.DB, xa *dtmcli.Xa) (interface{}, error) {
			if reqFrom(c).TransOutResult == dtmcli.ResultFailure {
				return dtmcli.MapFailure, nil
			}
			_, err := dtmcli.DBExec(db, "update dtm_busi.user_account set balance=balance-? where user_id=?", reqFrom(c).Amount, 1)
			return dtmcli.MapSuccess, err
		})
	}))

	app.POST(BusiAPI+"/TransOutXaGorm", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return XaClient.XaLocalTransaction(c.Request.URL.Query(), func(db *sql.DB, xa *dtmcli.Xa) (interface{}, error) {
			if reqFrom(c).TransOutResult == dtmcli.ResultFailure {
				return dtmcli.MapFailure, nil
			}
			var dia gorm.Dialector = nil
			if dtmcli.GetCurrentDBType() == dtmcli.DBTypeMysql {
				dia = mysql.New(mysql.Config{Conn: db})
			} else if dtmcli.GetCurrentDBType() == dtmcli.DBTypePostgres {
				dia = postgres.New(postgres.Config{Conn: db})
			}
			gdb, err := gorm.Open(dia, &gorm.Config{})
			if err != nil {
				return nil, err
			}
			dbr := gdb.Exec("update dtm_busi.user_account set balance=balance-? where user_id=?", reqFrom(c).Amount, 1)
			return dtmcli.MapSuccess, dbr.Error
		})
	}))

	app.POST(BusiAPI+"/TestPanic", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		if c.Query("panic_error") != "" {
			panic(errors.New("panic_error"))
		} else if c.Query("panic_string") != "" {
			panic("panic_string")
		}
		return "SUCCESS", nil
	}))
}
