package test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/examples"
)

func TestBarrierTcc(t *testing.T) {
	tccBarrierDisorder(t)
	tccBarrierNormal(t)
	tccBarrierRollback(t)

}

func tccBarrierRollback(t *testing.T) {
	gid := "tccBarrierRollback"
	err := dtmcli.TccGlobalTransaction(DtmServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
		_, err := tcc.CallBranch(&examples.TransReq{Amount: 30}, Busi+"/TccBTransOutTry", Busi+"/TccBTransOutConfirm", Busi+"/TccBTransOutCancel")
		assert.Nil(t, err)
		return tcc.CallBranch(&examples.TransReq{Amount: 30, TransInResult: "FAILURE"}, Busi+"/TccBTransInTry", Busi+"/TccBTransInConfirm", Busi+"/TccBTransInCancel")
	})
	assert.Error(t, err)
	WaitTransProcessed(gid)
	assert.Equal(t, "failed", getTransStatus(gid))
}

func tccBarrierNormal(t *testing.T) {
	gid := "tccBarrierNormal"
	err := dtmcli.TccGlobalTransaction(DtmServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
		_, err := tcc.CallBranch(&examples.TransReq{Amount: 30}, Busi+"/TccBTransOutTry", Busi+"/TccBTransOutConfirm", Busi+"/TccBTransOutCancel")
		assert.Nil(t, err)
		return tcc.CallBranch(&examples.TransReq{Amount: 30}, Busi+"/TccBTransInTry", Busi+"/TccBTransInConfirm", Busi+"/TccBTransInCancel")
	})
	assert.Nil(t, err)
	WaitTransProcessed(gid)
	assert.Equal(t, "succeed", getTransStatus(gid))
}

func tccBarrierDisorder(t *testing.T) {
	timeoutChan := make(chan string, 2)
	finishedChan := make(chan string, 2)
	gid := "tccBarrierDisorder"
	err := dtmcli.TccGlobalTransaction(DtmServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
		body := &examples.TransReq{Amount: 30}
		tryURL := Busi + "/TccBTransOutTry"
		confirmURL := Busi + "/TccBTransOutConfirm"
		cancelURL := Busi + "/TccBSleepCancel"
		// 请参见子事务屏障里的时序图，这里为了模拟该时序图，手动拆解了callbranch
		branchID := tcc.NewBranchID()
		sleeped := false
		app.POST(examples.BusiAPI+"/TccBSleepCancel", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
			res, err := examples.TccBarrierTransOutCancel(c)
			if !sleeped {
				sleeped = true
				dtmcli.Logf("sleep before cancel return")
				<-timeoutChan
				finishedChan <- "1"
			}
			return res, err
		}))
		// 注册子事务
		err := tcc.CallDtm(M{
			"gid":        tcc.Gid,
			"branch_id":  branchID,
			"trans_type": "tcc",
			"status":     "prepared",
			"data":       string(dtmcli.MustMarshal(body)),
			"try":        tryURL,
			"confirm":    confirmURL,
			"cancel":     cancelURL,
		}, "registerTccBranch")
		assert.Nil(t, err)
		go func() {
			dtmcli.Logf("sleeping to wait for tcc try timeout")
			<-timeoutChan
			r, _ := dtmcli.RestyClient.R().
				SetBody(body).
				SetQueryParams(dtmcli.MS{
					"dtm":         tcc.Dtm,
					"gid":         tcc.Gid,
					"branch_id":   branchID,
					"trans_type":  "tcc",
					"branch_type": "try",
				}).
				Post(tryURL)
			assert.True(t, strings.Contains(r.String(), "FAILURE"))
			finishedChan <- "1"
		}()
		dtmcli.Logf("cron to timeout and then call cancel")
		go CronTransOnce(60 * time.Second)
		time.Sleep(100 * time.Millisecond)
		dtmcli.Logf("cron to timeout and then call cancelled twice")
		CronTransOnce(60 * time.Second)
		timeoutChan <- "wake"
		timeoutChan <- "wake"
		<-finishedChan
		<-finishedChan
		time.Sleep(100 * time.Millisecond)
		return nil, fmt.Errorf("a cancelled tcc")
	})
	assert.Error(t, err, fmt.Errorf("a cancelled tcc"))
	assert.Equal(t, []string{"succeed", "prepared", "prepared"}, getBranchesStatus(gid))
	assert.Equal(t, "failed", getTransStatus(gid))
}
