package dtmcli

import (
	"context"
	"strings"

	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoCall sub-trans barrier for mongo. see http://dtm.pub/practice/barrier
// experimental
func (bb *BranchBarrier) MongoCall(mc *mongo.Client, busiCall func(mongo.SessionContext) error) (rerr error) {
	bid := bb.newBarrierID()
	return mc.UseSession(context.Background(), func(sc mongo.SessionContext) (rerr error) {
		rerr = sc.StartTransaction()
		if rerr != nil {
			return nil
		}
		defer dtmimp.DeferDo(&rerr, func() error {
			return sc.CommitTransaction(sc)
		}, func() error {
			return sc.AbortTransaction(sc)
		})
		originOp := map[string]string{
			BranchCancel:     BranchTry,
			BranchCompensate: BranchAction,
		}[bb.Op]

		originAffected, oerr := mongoInsertBarrier(sc, mc, bb.TransType, bb.Gid, bb.BranchID, originOp, bid, bb.Op)
		currentAffected, rerr := mongoInsertBarrier(sc, mc, bb.TransType, bb.Gid, bb.BranchID, bb.Op, bid, bb.Op)
		logger.Debugf("originAffected: %d currentAffected: %d", originAffected, currentAffected)

		if rerr == nil && bb.Op == opMsg && currentAffected == 0 { // for msg's DoAndSubmit, repeated insert should be rejected.
			return ErrDuplicated
		}

		if rerr == nil {
			rerr = oerr
		}
		if (bb.Op == BranchCancel || bb.Op == BranchCompensate) && originAffected > 0 || // null compensate
			currentAffected == 0 { // repeated request or dangled request
			return
		}
		if rerr == nil {
			rerr = busiCall(sc)
		}
		return
	})
}

// MongoQueryPrepared query prepared for redis
// experimental
func (bb *BranchBarrier) MongoQueryPrepared(mc *mongo.Client) error {
	_, err := mongoInsertBarrier(context.Background(), mc, bb.TransType, bb.Gid, "00", "msg", "01", "rollback")
	var result bson.M
	if err == nil {
		fs := strings.Split(dtmimp.BarrierTableName, ".")
		barrier := mc.Database(fs[0]).Collection(fs[1])
		err = barrier.FindOne(context.Background(), bson.D{
			{Key: "gid", Value: bb.Gid},
			{Key: "branch_id", Value: "00"},
			{Key: "op", Value: "msg"},
			{Key: "barrier_id", Value: "01"},
		}).Decode(&result)
	}
	var reason string
	if err == nil {
		reason, _ = result["reason"].(string)
	}
	if err == nil && reason == "rollback" {
		return ErrFailure
	}
	return err
}

func mongoInsertBarrier(sc context.Context, mc *mongo.Client, transType string, gid string, branchID string, op string, barrierID string, reason string) (int64, error) {
	if op == "" {
		return 0, nil
	}
	fs := strings.Split(dtmimp.BarrierTableName, ".")
	barrier := mc.Database(fs[0]).Collection(fs[1])
	r := barrier.FindOne(sc, bson.D{
		{Key: "gid", Value: gid},
		{Key: "branch_id", Value: branchID},
		{Key: "op", Value: op},
		{Key: "barrier_id", Value: barrierID},
	})
	err := r.Err()
	if err == mongo.ErrNoDocuments {
		_, err = barrier.InsertOne(sc,
			bson.D{
				{Key: "trans_type", Value: transType},
				{Key: "gid", Value: gid},
				{Key: "branch_id", Value: branchID},
				{Key: "op", Value: op},
				{Key: "barrier_id", Value: barrierID},
				{Key: "reason", Value: reason},
			})
		return 1, err
	}
	return 0, err
}
