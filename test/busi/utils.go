package busi

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	sync "sync"
	"time"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/dtmgrpc"
	"github.com/dtm-labs/dtm/dtmgrpc/dtmgpb"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/go-resty/resty/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func dbGet() *dtmutil.DB {
	return dtmutil.DbGet(BusiConf)
}

func pdbGet() *sql.DB {
	db, err := dtmimp.PooledDB(BusiConf)
	logger.FatalIfError(err)
	return db
}

func txGet() *sql.Tx {
	db := pdbGet()
	tx, err := db.Begin()
	logger.FatalIfError(err)
	return tx
}

func resetXaData() {
	if BusiConf.Driver != "mysql" {
		return
	}

	db := dbGet()
	type XaRow struct {
		Data string
	}
	xas := []XaRow{}
	db.Must().Raw("xa recover").Scan(&xas)
	for _, xa := range xas {
		db.Must().Exec(fmt.Sprintf("xa rollback '%s'", xa.Data))
	}
}

// MustBarrierFromGin 1
func MustBarrierFromGin(c *gin.Context) *dtmcli.BranchBarrier {
	ti, err := dtmcli.BarrierFromQuery(c.Request.URL.Query())
	logger.FatalIfError(err)
	return ti
}

// MustBarrierFromGrpc 1
func MustBarrierFromGrpc(ctx context.Context) *dtmcli.BranchBarrier {
	ti, err := dtmgrpc.BarrierFromGrpc(ctx)
	logger.FatalIfError(err)
	return ti
}

// SetGrpcHeaderForHeadersYes interceptor to set head for HeadersYes
func SetGrpcHeaderForHeadersYes(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	if r, ok := req.(*dtmgpb.DtmRequest); ok && strings.HasSuffix(r.Gid, "HeadersYes") {
		logger.Debugf("writing test_header:test to ctx")
		md := metadata.New(map[string]string{"test_header": "test"})
		ctx = metadata.NewOutgoingContext(ctx, md)
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}

// SetHTTPHeaderForHeadersYes interceptor to set head for HeadersYes
func SetHTTPHeaderForHeadersYes(c *resty.Client, r *resty.Request) error {
	if b, ok := r.Body.(*dtmcli.Saga); ok && strings.HasSuffix(b.Gid, "HeadersYes") {
		logger.Debugf("set test_header for url: %s", r.URL)
		r.SetHeader("test_header", "yes")
	}
	return nil
}

// oldWrapHandler old wrap handler for test use of dtm
func oldWrapHandler(fn func(*gin.Context) (interface{}, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		began := time.Now()
		r, err := func() (r interface{}, rerr error) {
			defer dtmimp.P2E(&rerr)
			return fn(c)
		}()
		var b = []byte{}
		if resp, ok := r.(*resty.Response); ok { // 如果是response，则取出body直接处理
			b = resp.Body()
		} else if err == nil {
			b, err = json.Marshal(r)
		}

		if err != nil {
			logger.Errorf("%2dms 500 %s %s %s %s", time.Since(began).Milliseconds(), err.Error(), c.Request.Method, c.Request.RequestURI, string(b))
			c.JSON(500, map[string]interface{}{"code": 500, "message": err.Error()})
		} else {
			logger.Infof("%2dms 200 %s %s %s", time.Since(began).Milliseconds(), c.Request.Method, c.Request.RequestURI, string(b))
			c.Status(200)
			c.Writer.Header().Add("Content-Type", "application/json")
			_, err = c.Writer.Write(b)
			dtmimp.E2P(err)
		}
	}
}

var (
	rdb  *redis.Client
	once sync.Once
)

// RedisGet 1
func RedisGet() *redis.Client {
	once.Do(func() {
		logger.Debugf("connecting to client redis")
		rdb = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:6379", StoreHost),
			Username: "root",
			Password: "",
		})
	})
	return rdb
}

var (
	mongoOnce sync.Once
	mongoc    *mongo.Client
)

// MongoGet get mongo client
func MongoGet() *mongo.Client {
	mongoOnce.Do(func() {
		uri := fmt.Sprintf("mongodb://%s:27017/?retryWrites=false", StoreHost)
		ctx := context.Background()
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
		dtmimp.E2P(err)
		mongoc = client
	})
	return mongoc
}

// SetRedisBothAccount 1
func SetRedisBothAccount(amountA int, ammountB int) {
	rd := RedisGet()
	_, err := rd.Set(rd.Context(), GetRedisAccountKey(TransOutUID), amountA, 0).Result()
	dtmimp.E2P(err)
	_, err = rd.Set(rd.Context(), GetRedisAccountKey(TransInUID), ammountB, 0).Result()
	dtmimp.E2P(err)
}

// SetMongoBothAccount 1
func SetMongoBothAccount(amountA int, amountB int) {
	mc := MongoGet()
	col := mc.Database("dtm_busi").Collection("user_account")
	_, err := col.InsertOne(context.Background(), bson.D{{Key: "user_id", Value: TransOutUID}, {Key: "balance", Value: amountA}})
	dtmimp.E2P(err)
	_, err = col.InsertOne(context.Background(), bson.D{{Key: "user_id", Value: TransInUID}, {Key: "balance", Value: amountB}})
	dtmimp.E2P(err)

}

// SetupMongoBarrierAndBusi 1
func SetupMongoBarrierAndBusi() {
	mc := MongoGet()
	err := mc.Database("dtm_busi").Drop(context.Background())
	dtmimp.E2P(err)
	err = mc.Database("dtm_barrier").Drop(context.Background())
	dtmimp.E2P(err)
	col := mc.Database("dtm_barrier").Collection("barrier")
	_, err = col.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{
			{Key: "gid", Value: 1},
			{Key: "branch_id", Value: 1},
			{Key: "op", Value: 1},
			{Key: "barrier_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	dtmimp.E2P(err)
	SetMongoBothAccount(10000, 10000)
}
