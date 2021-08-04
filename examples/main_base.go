package examples

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
)

const (
	// BusiAPI busi api prefix
	BusiAPI = "/api/busi"
	// BusiPort busi server port
	BusiPort = 8081
)

// Busi busi service url prefix
var Busi string = fmt.Sprintf("http://localhost:%d%s", BusiPort, BusiAPI)

// BaseAppStartup base app startup
func BaseAppStartup() *gin.Engine {
	dtmcli.Logf("examples starting")
	app := common.GetGinApp()
	BaseAddRoute(app)
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
	res := dtmcli.OrString(result1, result2, "SUCCESS")
	dtmcli.Logf("%s %s result: %s", busi, info.String(), res)
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
		return dtmcli.OrString(MainSwitch.CanSubmitResult.Fetch(), "SUCCESS"), nil
	}))
}
