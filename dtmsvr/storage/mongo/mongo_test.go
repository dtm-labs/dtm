package mongo

import (
	"context"
	"testing"

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
	var limit int64 = 2
	var position string = "62c93657cd2317b252c1f740"
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
