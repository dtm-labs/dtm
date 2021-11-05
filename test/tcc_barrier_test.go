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
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/examples"
)

func TestTccBarrierRollback(t *testing.T) {
	gid := dtmimp.GetFuncName()
	err := dtmcli.TccGlobalTransaction(DtmServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
		_, err := tcc.CallBranch(&examples.TransReq{Amount: 30}, Busi+"/TccBTransOutTry", Busi+"/TccBTransOutConfirm", Busi+"/TccBTransOutCancel")
		assert.Nil(t, err)
		return tcc.CallBranch(&examples.TransReq{Amount: 30, TransInResult: dtmcli.ResultFailure}, Busi+"/TccBTransInTry", Busi+"/TccBTransInConfirm", Busi+"/TccBTransInCancel")
	})
	assert.Error(t, err)
	waitTransProcessed(gid)
	assert.Equal(t, dtmcli.StatusFailed, getTransStatus(gid))
}

func TestTccBarrierNormal(t *testing.T) {
	gid := dtmimp.GetFuncName()
	err := dtmcli.TccGlobalTransaction(DtmServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
		_, err := tcc.CallBranch(&examples.TransReq{Amount: 30}, Busi+"/TccBTransOutTry", Busi+"/TccBTransOutConfirm", Busi+"/TccBTransOutCancel")
		assert.Nil(t, err)
		return tcc.CallBranch(&examples.TransReq{Amount: 30}, Busi+"/TccBTransInTry", Busi+"/TccBTransInConfirm", Busi+"/TccBTransInCancel")
	})
	assert.Nil(t, err)
	waitTransProcessed(gid)
	assert.Equal(t, dtmcli.StatusSucceed, getTransStatus(gid))
}

func TestTccBarrierDisorder(t *testing.T) {
	timeoutChan := make(chan string, 2)
	finishedChan := make(chan string, 2)
	gid := dtmimp.GetFuncName()
	err := dtmcli.TccGlobalTransaction(DtmServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
		body := &examples.TransReq{Amount: 30}
		tryURL := Busi + "/TccBTransOutTry"
		confirmURL := Busi + "/TccBTransOutConfirm"
		cancelURL := Busi + "/TccBSleepCancel"
		// 请参见子事务屏障里的时序图，这里为了模拟该时序图，手动拆解了callbranch
		branchID := tcc.NewSubBranchID()
		sleeped := false
		app.POST(examples.BusiAPI+"/TccBSleepCancel", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
			res, err := examples.TccBarrierTransOutCancel(c)
			if !sleeped {
				sleeped = true
				dtmimp.Logf("sleep before cancel return")
				<-timeoutChan
				finishedChan <- "1"
			}
			return res, err
		}))
		// 注册子事务
		resp, err := dtmimp.RestyClient.R().
			SetBody(map[string]interface{}{
				"gid":                tcc.Gid,
				"branch_id":          branchID,
				"trans_type":         "tcc",
				"status":             dtmcli.StatusPrepared,
				"data":               string(dtmimp.MustMarshal(body)),
				dtmcli.BranchTry:     tryURL,
				dtmcli.BranchConfirm: confirmURL,
				dtmcli.BranchCancel:  cancelURL,
			}).Post(fmt.Sprintf("%s/%s", tcc.Dtm, "registerTccBranch"))
		assert.Nil(t, err)
		assert.Contains(t, resp.String(), dtmcli.ResultSuccess)

		go func() {
			dtmimp.Logf("sleeping to wait for tcc try timeout")
			<-timeoutChan
			r, _ := dtmimp.RestyClient.R().
				SetBody(body).
				SetQueryParams(map[string]string{
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
		dtmimp.Logf("cron to timeout and then call cancel")
		go cronTransOnceForwardNow(300)
		time.Sleep(100 * time.Millisecond)
		dtmimp.Logf("cron to timeout and then call cancelled twice")
		cronTransOnceForwardNow(300)
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

func TestTccBarrierPanic(t *testing.T) {
	bb := &dtmcli.BranchBarrier{TransType: "saga", Gid: "gid1", BranchID: "bid1", BranchType: "action", BarrierID: 1}
	var err error
	func() {
		defer dtmimp.P2E(&err)
		tx, _ := dbGet().ToSQLDB().BeginTx(context.Background(), &sql.TxOptions{})
		bb.Call(tx, func(db dtmcli.DB) error {
			panic(fmt.Errorf("an error"))
		})
	}()
	assert.Error(t, err, fmt.Errorf("an error"))
}
