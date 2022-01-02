package busi

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/gin-gonic/gin"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const transOutUID = 1
const transInUID = 2

func handleGrpcBusiness(in *BusiReq, result1 string, result2 string, busi string) error {
	res := dtmimp.OrString(result1, result2, dtmcli.ResultSuccess)
	logger.Debugf("grpc busi %s %v %s %s result: %s", busi, in, result1, result2, res)
	if res == dtmcli.ResultSuccess {
		return nil
	} else if res == dtmcli.ResultFailure {
		return status.New(codes.Aborted, dtmcli.ResultFailure).Err()
	} else if res == dtmcli.ResultOngoing {
		return status.New(codes.Aborted, dtmcli.ResultOngoing).Err()
	}
	return status.New(codes.Internal, fmt.Sprintf("unknow result %s", res)).Err()
}

func handleGeneralBusiness(c *gin.Context, result1 string, result2 string, busi string) (interface{}, error) {
	info := infoFromContext(c)
	res := dtmimp.OrString(result1, result2, dtmcli.ResultSuccess)
	logger.Debugf("%s %s result: %s", busi, info.String(), res)
	if res == "ERROR" {
		return nil, errors.New("ERROR from user")
	}
	return map[string]interface{}{"dtm_result": res}, nil
}

func error2Resp(err error) (interface{}, error) {
	if err != nil {
		s := err.Error()
		if strings.Contains(s, dtmcli.ResultFailure) || strings.Contains(s, dtmcli.ResultOngoing) {
			return gin.H{"dtm_result": s}, nil
		}
		return nil, err
	}
	return gin.H{"dtm_result": dtmcli.ResultSuccess}, nil
}

func sagaGrpcAdjustBalance(db dtmcli.DB, uid int, amount int64, result string) error {
	if result == dtmcli.ResultFailure {
		return status.New(codes.Aborted, dtmcli.ResultFailure).Err()
	}
	_, err := dtmimp.DBExec(db, "update dtm_busi.user_account set balance = balance + ? where user_id = ?", amount, uid)
	return err

}

func sagaAdjustBalance(db dtmcli.DB, uid int, amount int, result string) error {
	if strings.Contains(result, dtmcli.ResultFailure) {
		return dtmcli.ErrFailure
	}
	_, err := dtmimp.DBExec(db, "update dtm_busi.user_account set balance = balance + ? where user_id = ?", amount, uid)
	return err
}
func tccAdjustTrading(db dtmcli.DB, uid int, amount int) error {
	affected, err := dtmimp.DBExec(db, `update dtm_busi.user_account set trading_balance=trading_balance+?
		 where user_id=? and trading_balance + ? + balance >= 0`, amount, uid, amount)
	if err == nil && affected == 0 {
		return fmt.Errorf("update error, maybe balance not enough")
	}
	return err
}

func tccAdjustBalance(db dtmcli.DB, uid int, amount int) error {
	affected, err := dtmimp.DBExec(db, `update dtm_busi.user_account set trading_balance=trading_balance-?,
		 balance=balance+? where user_id=?`, amount, amount, uid)
	if err == nil && affected == 0 {
		return fmt.Errorf("update user_account 0 rows")
	}
	return err
}
