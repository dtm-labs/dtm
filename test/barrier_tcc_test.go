package test

import (
	"context"
	"database/sql"
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
	barrierPanic(t)
}

func tccBarrierRollback(t *testing.T) {
	gid := "tccBarrierRollback"
	err := dtmcli.TccGlobalTransaction(DtmServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
		_, err := tcc.CallBranch(&examples.TransReq{Amount: 30}, Busi+"/TccBTransOutTry", Busi+"/TccBTransOutConfirm", Busi+"/TccBTransOutCancel")
		assert.Nil(t, err)
		return tcc.CallBranch(&examples.TransReq{Amount: 30, TransInResult: dtmcli.ResultFailure}, Busi+"/TccBTransInTry", Busi+"/TccBTransInConfirm", Busi+"/TccBTransInCancel")
	})
	assert.Error(t, err)
	WaitTransProcessed(gid)
	assert.Equal(t, dtmcli.StatusFailed, getTransStatus(gid))
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
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(gid))
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
		resp, err := dtmcli.RestyClient.R().
			SetResult(&dtmcli.TransResult{}).SetBody(M{
			"gid":                tcc.Gid,
			"branch_id":          branchID,
			"trans_type":         "tcc",
			"status":             dtmcli.StatusPrepared,
			"data":               string(dtmcli.MustMarshal(body)),
			dtmcli.BranchTry:     tryURL,
			dtmcli.BranchConfirm: confirmURL,
			dtmcli.BranchCancel:  cancelURL,
		}).Post(fmt.Sprintf("%s/%s", tcc.Dtm, "registerTccBranch"))
		assert.Nil(t, err)
		tr := resp.Result().(*dtmcli.TransResult)
		assert.Equal(t, dtmcli.ResultSuccess, tr.DtmResult)

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
					"branch_type": dtmcli.BranchTry,
				}).
				Post(tryURL)
			assert.True(t, strings.Contains(r.String(), dtmcli.ResultSuccess)) // 这个是悬挂操作，为了简单起见，依旧让他返回成功
			finishedChan <- "1"
		}()
		dtmcli.Logf("cron to timeout and then call cancel")
		go CronTransOnce()
		time.Sleep(100 * time.Millisecond)
		dtmcli.Logf("cron to timeout and then call cancelled twice")
		CronTransOnce()
		timeoutChan <- "wake"
		timeoutChan <- "wake"
		<-finishedChan
		<-finishedChan
		time.Sleep(100 * time.Millisecond)
		return nil, fmt.Errorf("a cancelled tcc")
	})
	assert.Error(t, err, fmt.Errorf("a cancelled tcc"))
	assert.Equal(t, []string{dtmcli.StatusSucceed, dtmcli.StatusPrepared, dtmcli.StatusPrepared}, getBranchesStatus(gid))
	assert.Equal(t, dtmcli.StatusFailed, getTransStatus(gid))
}

func barrierPanic(t *testing.T) {
	bb := &dtmcli.BranchBarrier{TransType: "saga", Gid: "gid1", BranchID: "bid1", BranchType: "action", BarrierID: 1}
	var err error
	func() {
		defer dtmcli.P2E(&err)
		tx, _ := dbGet().ToSQLDB().BeginTx(context.Background(), &sql.TxOptions{})
		bb.Call(tx, func(db dtmcli.DB) error {
			panic(fmt.Errorf("an error"))
		})
	}()
	assert.Error(t, err, fmt.Errorf("an error"))
}
