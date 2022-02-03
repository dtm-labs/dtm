package busi

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// TransOutUID 1
const TransOutUID = 1

// TransInUID 2
const TransInUID = 2

// Redis 1
const Redis = "redis"

// Mongo 1
const Mongo = "mongo"

func handleGrpcBusiness(in *BusiReq, result1 string, result2 string, busi string) error {
	res := dtmimp.OrString(result1, result2, dtmcli.ResultSuccess)
	logger.Debugf("grpc busi %s %v %s %s result: %s", busi, in, result1, result2, res)
	if res == dtmcli.ResultSuccess {
		return nil
	} else if res == dtmcli.ResultFailure {
		return status.New(codes.Aborted, dtmcli.ResultFailure).Err()
	} else if res == dtmcli.ResultOngoing {
		return status.New(codes.FailedPrecondition, dtmcli.ResultOngoing).Err()
	}
	return status.New(codes.Internal, fmt.Sprintf("unknow result %s", res)).Err()
}

func handleGeneralBusiness(c *gin.Context, result1 string, result2 string, busi string) interface{} {
	info := infoFromContext(c)
	res := dtmimp.OrString(result1, result2, dtmcli.ResultSuccess)
	logger.Debugf("%s %s result: %s", busi, info.String(), res)
	if res == "ERROR" {
		return errors.New("ERROR from user")
	}
	return dtmcli.String2DtmError(res)
}

// old business handler. for compatible usage
func handleGeneralBusinessCompatible(c *gin.Context, result1 string, result2 string, busi string) (interface{}, error) {
	info := infoFromContext(c)
	res := dtmimp.OrString(result1, result2, dtmcli.ResultSuccess)
	logger.Debugf("%s %s result: %s", busi, info.String(), res)
	if res == "ERROR" {
		return nil, errors.New("ERROR from user")
	}
	return map[string]interface{}{"dtm_result": res}, nil
}

func sagaGrpcAdjustBalance(db dtmcli.DB, uid int, amount int64, result string) error {
	if result == dtmcli.ResultFailure {
		return status.New(codes.Aborted, dtmcli.ResultFailure).Err()
	}
	_, err := dtmimp.DBExec(db, "update dtm_busi.user_account set balance = balance + ? where user_id = ?", amount, uid)
	return err
}

// SagaAdjustBalance 1
func SagaAdjustBalance(db dtmcli.DB, uid int, amount int, result string) error {
	if strings.Contains(result, dtmcli.ResultFailure) {
		return dtmcli.ErrFailure
	}
	_, err := dtmimp.DBExec(db, "update dtm_busi.user_account set balance = balance + ? where user_id = ?", amount, uid)
	return err
}

// SagaMongoAdjustBalance 1
func SagaMongoAdjustBalance(ctx context.Context, mc *mongo.Client, uid int, amount int, result string) error {
	if strings.Contains(result, dtmcli.ResultFailure) {
		return dtmcli.ErrFailure
	}
	_, err := mc.Database("dtm_busi").Collection("user_account").UpdateOne(ctx,
		bson.D{{Key: "user_id", Value: uid}},
		bson.D{{Key: "$inc", Value: bson.D{{Key: "balance", Value: amount}}}})
	logger.Debugf("dtm_busi.user_account $inc balance of %d by %d err: %v", uid, amount, err)
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
