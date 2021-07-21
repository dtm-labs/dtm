package dtmsvr

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/examples"
)

func TestTccBarrier(t *testing.T) {
	tccBarrierDisorder(t)
	tccBarrierNormal(t)
	tccBarrierRollback(t)

}

func tccBarrierRollback(t *testing.T) {
	gid, err := dtmcli.TccGlobalTransaction(DtmServer, func(tcc *dtmcli.Tcc) (rerr error) {
		res1, rerr := tcc.CallBranch(&examples.TransReq{Amount: 30}, Busi+"/TccBTransOutTry", Busi+"/TccBTransOutConfirm", Busi+"/TccBTransOutCancel")
		e2p(rerr)
		if res1.StatusCode() != 200 {
			return fmt.Errorf("bad status code: %d", res1.StatusCode())
		}
		res2, rerr := tcc.CallBranch(&examples.TransReq{Amount: 30, TransInResult: "FAILURE"}, Busi+"/TccBTransInTry", Busi+"/TccBTransInConfirm", Busi+"/TccBTransInCancel")
		e2p(rerr)
		if res2.StatusCode() != 200 {
			return fmt.Errorf("bad status code: %d", res2.StatusCode())
		}
		if strings.Contains(res2.String(), "FAILURE") {
			return fmt.Errorf("branch trans in fail")
		}
		logrus.Printf("tcc returns: %s, %s", res1.String(), res2.String())
		return
	})
	assert.Equal(t, err, fmt.Errorf("branch trans in fail"))
	WaitTransProcessed(gid)
	assert.Equal(t, "failed", getTransStatus(gid))
}

func tccBarrierNormal(t *testing.T) {
	_, err := dtmcli.TccGlobalTransaction(DtmServer, func(tcc *dtmcli.Tcc) (rerr error) {
		res1, rerr := tcc.CallBranch(&examples.TransReq{Amount: 30}, Busi+"/TccBTransOutTry", Busi+"/TccBTransOutConfirm", Busi+"/TccBTransOutCancel")
		e2p(rerr)
		if res1.StatusCode() != 200 {
			return fmt.Errorf("bad status code: %d", res1.StatusCode())
		}
		res2, rerr := tcc.CallBranch(&examples.TransReq{Amount: 30}, Busi+"/TccBTransInTry", Busi+"/TccBTransInConfirm", Busi+"/TccBTransInCancel")
		e2p(rerr)
		if res2.StatusCode() != 200 {
			return fmt.Errorf("bad status code: %d", res2.StatusCode())
		}
		logrus.Printf("tcc returns: %s, %s", res1.String(), res2.String())
		return
	})
	e2p(err)
}

func tccBarrierDisorder(t *testing.T) {
	timeoutChan := make(chan string, 2)
	finishedChan := make(chan string, 2)
	gid, err := dtmcli.TccGlobalTransaction(DtmServer, func(tcc *dtmcli.Tcc) (rerr error) {
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
				logrus.Printf("sleep before cancel return")
				<-timeoutChan
				finishedChan <- "1"
			}
			return res, err
		}))
		// 注册子事务
		r, err := common.RestyClient.R().
			SetBody(&M{
				"gid":        tcc.Gid,
				"branch_id":  branchID,
				"trans_type": "tcc",
				"status":     "prepared",
				"data":       string(common.MustMarshal(body)),
				"try":        tryURL,
				"confirm":    confirmURL,
				"cancel":     cancelURL,
			}).
			Post(tcc.Dtm + "/registerTccBranch")
		e2p(err)
		assert.True(t, strings.Contains(r.String(), "SUCCESS"))
		go func() {
			logrus.Printf("sleeping to wait for tcc try timeout")
			<-timeoutChan
			r, _ = common.RestyClient.R().
				SetBody(body).
				SetQueryParams(common.MS{
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
		logrus.Printf("cron to timeout and then call cancel")
		go CronTransOnce(60 * time.Second)
		time.Sleep(100 * time.Millisecond)
		logrus.Printf("cron to timeout and then call cancelled twice")
		CronTransOnce(60 * time.Second)
		timeoutChan <- "wake"
		timeoutChan <- "wake"
		<-finishedChan
		<-finishedChan
		time.Sleep(100 * time.Millisecond)
		return fmt.Errorf("a cancelled tcc")
	})
	assert.Error(t, err, fmt.Errorf("a cancelled tcc"))
	assert.Equal(t, []string{"succeed", "prepared", "prepared"}, getBranchesStatus(gid))
	assert.Equal(t, "failed", getTransStatus(gid))
}
