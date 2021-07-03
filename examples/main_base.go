package examples

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
)

const (
	BusiApi  = "/api/busi"
	BusiPort = 8081
)

var Busi string = fmt.Sprintf("http://localhost:%d%s", BusiPort, BusiApi)

func BaseAppNew() *gin.Engine {
	logrus.Printf("examples starting")
	app := common.GetGinApp()
	return app
}

func BaseAppStart(app *gin.Engine) {
	logrus.Printf("Starting busi at: %d", BusiPort)
	go app.Run(fmt.Sprintf(":%d", BusiPort))
	time.Sleep(100 * time.Millisecond)
}

type AutoEmptyString struct {
	value string
}

func (s *AutoEmptyString) SetOnce(v string) {
	s.value = v
}

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

var MainSwitch mainSwitchType

func handleGeneralBusiness(c *gin.Context, result1 string, result2 string, busi string) (interface{}, error) {
	info := infoFromContext(c)
	res := common.OrString(MainSwitch.TransInResult.Fetch(), result2, "SUCCESS")
	logrus.Printf("%s %s result: %s", info.String(), common.GetFuncName(), res)
	return M{"result": res}, nil

}

func BaseAppSetup(app *gin.Engine) {
	app.POST(BusiApi+"/TransIn", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return handleGeneralBusiness(c, MainSwitch.TransInResult.Fetch(), reqFrom(c).TransInResult, "transIn")
	}))
	app.POST(BusiApi+"/TransOut", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return handleGeneralBusiness(c, MainSwitch.TransOutResult.Fetch(), reqFrom(c).TransOutResult, "transIn")
	}))
	app.POST(BusiApi+"/TransInConfirm", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return handleGeneralBusiness(c, MainSwitch.TransInConfirmResult.Fetch(), "", "transIn")
	}))
	app.POST(BusiApi+"/TransOutConfirm", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return handleGeneralBusiness(c, MainSwitch.TransOutConfirmResult.Fetch(), "", "transIn")
	}))
	app.POST(BusiApi+"/TransInRevert", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return handleGeneralBusiness(c, MainSwitch.TransInRevertResult.Fetch(), "", "transIn")
	}))
	app.POST(BusiApi+"/TransOutRevert", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return handleGeneralBusiness(c, MainSwitch.TransOutRevertResult.Fetch(), "", "transIn")
	}))
	app.GET(BusiApi+"/CanSubmit", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		logrus.Printf("%s CanSubmit", c.Query("gid"))
		return common.OrString(MainSwitch.CanSubmitResult.Fetch(), "SUCCESS"), nil
	}))
}
