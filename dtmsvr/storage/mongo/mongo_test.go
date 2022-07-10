package mongo

import (
	"context"
	"testing"
	"time"

	// "errors"
	"fmt"
	// "sync"

	// "github.com/dtm-labs/dtm/dtmsvr/config"
	"github.com/dtm-labs/dtm/dtmsvr/storage"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMongoGet(t *testing.T) {
	r := mongoGetForTest()
	if r == nil {
		t.Errorf("connect mongo failed")
	}
}
func TestMongoInsert(t *testing.T) {
	mongc := mongoGetForTest()
	co := mongc.Database("test").Collection("books")
	doc := bson.D{{"title", "Invisible Cities"}, {"author", "Italo Calvino"}, {"year_published", 1974}}
	result, err := co.InsertOne(context.TODO(), doc)
	if err != nil {
		t.Errorf("insert mongo failed")
	}
	fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)

}
func TestMongoFindAll(t *testing.T) {
	mongc := mongoGetForTest()
	coll := mongc.Database("test").Collection("books")

	filter := bson.D{{}}
	cursor, err := coll.Find(context.TODO(), filter)
	if err != nil {
		t.Errorf("mongo find failed")
	}
	var results []bson.D
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	for _, result := range results {
		fmt.Println(result)
	}
}

func TestMongoFindAllTrans(t *testing.T) {
	mongoc := mongoGetForTest()
	coll := mongoc.Database("dtm").Collection("trans")
	var res []MongoGlobalTrans
	// coll.Find(ctx, bson.D{{}})
	cursor, err := coll.Find(ctx, bson.D{{}})
	// cursor.Decode(&res)
	if err != nil {
		t.Errorf("mongo find failed")
	}
	if err := cursor.All(ctx, &res); err != nil {
		t.Errorf("mongo cursor failed")
		fmt.Printf("err: %v\n", err)
	}
	for _, result := range res {
		fmt.Println(result)
	}
}

func TestPopulate(t *testing.T) {
	mongoc := mongoGetForTest()
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
		{
			Keys: bson.D{
				{Key: "ext_data", Value: "text"},
			},
			Options: options.Index().SetSparse(true),
		},
	})
	if err != nil {
		t.Errorf("populate data failed")
	}
	doc := bson.D{
		{"gid", "1234"},
		{"owner", "test_user"},
		{"status", "test_status"},
		{"next_cron_time", "2006-01-02T15:04:05.999Z"},
		{"branch_trans", []bson.D{
			{
				{"branch_id", "branchId123"},
				{"op", "test_op"},
			},
			{
				{"branch_id", "branchId456"},
				{"op", "test_op"},
			},
		}},
	}
	coll.InsertOne(ctx, doc)
	filter := bson.M{
		"gid":                    "1234",
		"branch_trans.branch_id": "branchId123",
		"branch_trans.op":        "test_op",
	}
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		t.Errorf("mongo find failed")
	}
	var results []bson.D
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	for _, result := range results {
		fmt.Println(result)
	}
}
func TestFindTransGlobalStore(t *testing.T) {
	var mggtrans MongoGlobalTrans
	r := mongoGetForTest().Database("dtm").Collection("trans").FindOne(ctx, bson.D{
		{Key: "gid", Value: "1234"},
	})
	err := r.Err()
	if err == mongo.ErrNoDocuments || err != nil {
		panic(err)
	}
	err = r.Decode(&mggtrans)
	if err != nil {
		panic(err)
	}
	trans := ConvertMongoTransToTrans(&mggtrans)
	fmt.Println(trans)
	fmt.Println(trans.NextCronTime == mggtrans.NextCronTime)
}
func TestScan(t *testing.T) {
	var limit int64 = 3
	var position string = ""
	filter := bson.D{{}}
	opts := options.Find().SetLimit(limit).SetSort(bson.D{{"_id", 1}})
	if position != "" {
		lid, err := primitive.ObjectIDFromHex(position)
		if err != nil {
			panic(err)
		}
		filter = bson.D{{"_id", bson.D{{"$gt", lid}}}}
	}
	cursor, err := mongoGetForTest().Database("dtm").Collection("trans").Find(ctx, filter, opts)
	if err != nil {
		panic(err)
	}
	mggtrans := make([]MongoGlobalTrans, limit)
	err = cursor.All(ctx, &mggtrans)
	if err != nil {
		panic(err)
	}
	len := len(mggtrans)
	if len < int(limit) {
		position = ""
	} else {
		position = mggtrans[len-1].ID.Hex()
	}
	err = cursor.Close(ctx)
	if err != nil {
		panic(err)
	}
	trans := make([]storage.TransGlobalStore, len)
	for i, e := range mggtrans {
		trans[i] = *ConvertMongoTransToTrans(&e)
	}
	fmt.Println(trans)
}

func TestFindBranch(t *testing.T) {
	gid := "lainlkno10923"
	filter := bson.D{{"gid", gid}}
	r := mongoGetForTest().Database(database).Collection(collection).FindOne(ctx, filter)
	err := r.Err()
	if err == mongo.ErrNoDocuments {
		fmt.Println("no document found")
		return
	}
	mggtrans := MongoGlobalTrans{}
	err = r.Decode(&mggtrans)
	if err != nil {
		panic(err)
	}
	trans := ConvertMongoTransToBranch(&mggtrans)
	fmt.Println(trans)
}
func TestInsertOne(t *testing.T) {
	coll := mongoGetForTest().Database(database).Collection(collection)
	data := bson.D{
		{"gid", "lainlkno10923"},
		{"create_time", time.Now()},
		{"update_time", time.Now()},
		{"trans_type", "saga"},
		{"status", "prepared"},
		{"query_prepared", "/api/query_prepared"},
		{"protocol", "http"},
		{"finish_time", time.Now()},
		{"rollback_time", time.Now()},
		{"options", "TimeoutToFail:1000"},
		{"custom_data", "customData"},
		{"next_cron_interval", 100},
		{"next_cron_time", time.Now()},
		{"owner", "owner1"},
		{"ext_data", "extData"},
		{"branch_trans", []bson.D{
			{
				{"url", "/api/branch1"},
				{"branch_id", "branch1"},
				{"op", "insert"},
				{"branch_status", "status1"},
				{"branch_finish_time", time.Now()},
				{"branch_rollback_time", time.Now()},
				{"branch_bin_data", []byte{1, 2, 3, 4}},
				{"branch_create_time", time.Now()},
				{"branch_update_time", time.Now()},
			},
			{
				{"url", "/api/branch2"},
				{"branch_id", "branch2"},
				{"op", "delete"},
				{"branch_status", "status2"},
				{"branch_finish_time", time.Now()},
				{"branch_rollback_time", time.Now()},
				{"branch_bin_data", []byte{5, 6, 7, 8}},
				{"branch_create_time", time.Now()},
				{"branch_update_time", time.Now()},
			},
		}},
	}
	coll.InsertOne(ctx, data)
}
func mongoGetForTest() *mongo.Client {
	uri := fmt.Sprintf("mongodb://%s:27017/?retryWrites=false&directConnection=true", "localhost")
	ctx := context.Background()
	// logger.Infof("connecting to mongo: %s", uri)
	fmt.Println("connecting to mongo")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil
	}
	// dtmimp.E2P(err)
	// logger.Infof("connected to mongo: %s", uri)
	fmt.Println("connected to mongo")
	return client
}
