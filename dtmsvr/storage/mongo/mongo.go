/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package mongo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/dtmsvr/config"
	"github.com/dtm-labs/dtm/dtmsvr/storage"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	// "go.mongodb.org/mongo-driver/x/bsonx"
	"go.mongodb.org/mongo-driver/bson"
)

// TODO: optimize this, it's very strange to use pointer to dtmutil.Config
var conf = &config.Config

// TODO: optimize this, all function should have context as first parameter
var ctx = context.Background()

// Store implements storage.Store, and storage with db
type Store struct {
}

// Ping execs ping cmd to db
func (s *Store) Ping() error {
	err := MongoGet().Ping(ctx, nil)
	return err
}

// PopulateData populates data to db
func (s *Store) PopulateData(skipDrop bool) {
	// skipDrop means keepData
	if !skipDrop {
		mongoc := MongoGet()
		mongoc.Database("dtm").Drop(ctx)
		coll := mongoc.Database("dtm").Collection("trans")
		_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
			{
				Keys: bson.D{
					{Key: "gid", Value: 1},
				},
				Options: options.Index().SetUnique(true),
			},
			{
				Keys: bson.D{
					{Key: "owner", Value: 1},
				},
			},
			{
				Keys: bson.D{
					{Key: "status", Value: 1},
					{Key: "next_cron_time", Value: 1},
				},
				Options: options.Index().SetSparse(true),
			},
			{
				Keys: bson.D{
					{Key: "gid", Value: 1},
					{Key: "branch_trans.branch_id", Value: 1},
					{Key: "branch_trans.op", Value: 1},
				},
				Options: options.Index().SetName("gid_branchId_branchOp"),
			},
		})
		logger.Infof("call mongo PopulateData. result: %v", err)
		dtmimp.PanicIf(err != nil, err)
	}
}

// FindTransGlobalStore finds GlobalTrans data by gid
func (s *Store) FindTransGlobalStore(gid string) *storage.TransGlobalStore {
	var mggtrans MongoGlobalTrans
	r := MongoGet().Database("dtm").Collection("trans").FindOne(ctx, bson.D{
		{Key: "gid", Value: gid},
	})
	err := r.Err()
	if err == mongo.ErrNoDocuments {
		return nil
	}
	dtmimp.E2P(err)
	err = r.Decode(&mggtrans)
	if err != nil {
		dtmimp.E2P(err)
	}
	trans := ConvertMongoTransToTrans(&mggtrans)
	return trans
}

// ScanTransGlobalStores lists GlobalTrans data
func (s *Store) ScanTransGlobalStores(position *string, limit int64) []storage.TransGlobalStore {
	filter := bson.D{{}}
	opts := options.Find().SetLimit(limit).SetSort(bson.D{{"_id", 1}})
	if *position != "" {
		lid, err := primitive.ObjectIDFromHex(*position)
		if err != nil {
			panic(err)
		}
		filter = bson.D{{"_id", bson.D{{"$gt", lid}}}}
	}
	cursor, err := MongoGet().Database("dtm").Collection("trans").Find(ctx, filter, opts)
	if err != nil {
		dtmimp.E2P(err)
	}
	mggtrans := make([]MongoGlobalTrans, limit)
	err = cursor.All(ctx, &mggtrans)
	if err != nil {
		dtmimp.E2P(err)
	}
	len := len(mggtrans)
	if len < int(limit) {
		*position = ""
	} else {
		*position = mggtrans[len-1].ID.Hex()
	}
	err = cursor.Close(ctx)
	if err != nil {
		dtmimp.E2P(err)
	}
	trans := make([]storage.TransGlobalStore, len)
	for i, e := range mggtrans {
		trans[i] = *ConvertMongoTransToTrans(&e)
	}
	return trans
}

// FindBranches finds Branch data by gid
// func (s *Store) FindBranches(gid string) []storage.TransBranchStore {

// }

// UpdateBranches update branches info
// func (s *Store) UpdateBranches(branches []storage.TransBranchStore, updates []string) (int, error) {

// }

// LockGlobalSaveBranches creates branches
func (s *Store) LockGlobalSaveBranches(gid string, status string, branches []storage.TransBranchStore, branchStart int) {

}

// MaySaveNewTrans creates a new trans
// func (s *Store) MaySaveNewTrans(global *storage.TransGlobalStore, branches []storage.TransBranchStore) error {

// }

// ChangeGlobalStatus changes global trans status
func (s *Store) ChangeGlobalStatus(global *storage.TransGlobalStore, newStatus string, updates []string, finished bool) {

}

// TouchCronTime updates cronTime
func (s *Store) TouchCronTime(global *storage.TransGlobalStore, nextCronInterval int64, nextCronTime *time.Time) {

}

// LockOneGlobalTrans finds GlobalTrans
// func (s *Store) LockOneGlobalTrans(expireIn time.Duration) *storage.TransGlobalStore {

// }

// ResetCronTime reset nextCronTime
// unfinished transactions need to be retried as soon as possible after business downtime is recovered
// func (s *Store) ResetCronTime(after time.Duration, limit int64) (succeedCount int64, hasRemaining bool, err error) {

// }

var (
	mongoOnce sync.Once
	mongoc    *mongo.Client
)

// MongoGet get mongo client
func MongoGet() *mongo.Client {
	mongoOnce.Do(func() {
		uri := fmt.Sprintf("mongodb://%s:%d/?retryWrites=false&directConnection=true", conf.Store.Host, conf.Store.Port)
		ctx := context.Background()
		logger.Infof("connecting to mongo: %s", uri)
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
		dtmimp.E2P(err)
		logger.Infof("connected to mongo: %s", uri)
		mongoc = client
	})
	return mongoc
}

// manually convert mongo query result to storage.TransGlobalStore
// func MongoConvertGlobalTrans(mongoc *mongo.Client, cursor *mongo.Cursor) *storage.TransGlobalStore {

// }

type MongoGlobalTrans struct {
	ID         primitive.ObjectID `bson:"_id"`
	CreateTime *time.Time         `bson:"create_time"`
	UpdateTime *time.Time         `bson:"update_time"`
	Gid        string             `bson:"gid,omitempty"`
	TransType  string             `bson:"trans_type,omitempty"`
	// Steps            []map[string]string `bson:"steps,omitempty" gorm:"-"`
	Payloads []string `bson:"payloads,omitempty" gorm:"-"`
	// BinPayloads      [][]byte            `bson:"-" gorm:"-"`
	Status           string     `bson:"status,omitempty"`
	QueryPrepared    string     `bson:"query_prepared,omitempty"`
	Protocol         string     `bson:"protocol,omitempty"`
	FinishTime       *time.Time `bson:"finish_time,omitempty"`
	RollbackTime     *time.Time `bson:"rollback_time,omitempty"`
	Options          string     `bson:"options,omitempty"`
	CustomData       string     `bson:"custom_data,omitempty"`
	NextCronInterval int64      `bson:"next_cron_interval,omitempty"`
	NextCronTime     *time.Time `bson:"next_cron_time,omitempty"`
	Owner            string     `bson:"owner,omitempty"`
	// Ext              TransGlobalExt      `bson:"-" gorm:"-"`
	ExtData string `bson:"ext_data,omitempty"` // storage of ext. a db field to store many values. like Options
	// dtmcli.TransOptions
	BranchTrans []MongoBranchTrans `bson:"branch_trans,omitempty"`
}

type MongoBranchTrans struct {
	// Gid          string `json:"gid,omitempty"`
	URL string `bson:"url,omitempty"`
	// BinData      []byte
	BranchID     string     `bson:"branch_id,omitempty"`
	Op           string     `bson:"op,omitempty"`
	Status       string     `bson:"branch_status,omitempty"`
	FinishTime   *time.Time `bson:"branch_finish_time,omitempty"`
	RollbackTime *time.Time `bson:"branch_rollback_time,omitempty"`

	CreateTime *time.Time `bson:"branch_create_time"`
	UpdateTime *time.Time `bson:"branch_update_time"`
}

func ConvertMongoTransToTrans(mgt *MongoGlobalTrans) *storage.TransGlobalStore {
	trans := &storage.TransGlobalStore{}
	// pointer copy
	// trans.ID = uint64(mgt.ID[0])
	trans.CreateTime = mgt.CreateTime
	trans.UpdateTime = mgt.UpdateTime
	trans.Gid = mgt.Gid
	trans.TransType = mgt.TransType
	// trans.Steps = mgt.S
	trans.Payloads = mgt.Payloads
	// trans.BinPayloads = mgt.b
	trans.Status = mgt.Status
	trans.QueryPrepared = mgt.QueryPrepared
	trans.Protocol = mgt.Protocol
	trans.FinishTime = mgt.FinishTime
	trans.RollbackTime = mgt.RollbackTime
	trans.Options = mgt.Options
	trans.CustomData = mgt.CustomData
	trans.NextCronInterval = mgt.NextCronInterval
	trans.NextCronTime = mgt.NextCronTime
	trans.Owner = mgt.Owner
	trans.ExtData = mgt.ExtData
	return trans
}
