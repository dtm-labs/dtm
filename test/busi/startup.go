package busi

import (
	"context"
	"fmt"

	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/gin-gonic/gin"
)

// Startup startup the busi's grpc and http service
func Startup() *gin.Engine {
	GrpcStartup()
	return BaseAppStartup()
}

// PopulateDB populate example mysql data
func PopulateDB(skipDrop bool) {
	resetXaData()
	file := fmt.Sprintf("%s/busi.%s.sql", dtmutil.GetSQLDir(), BusiConf.Driver)
	dtmutil.RunSQLScript(BusiConf, file, skipDrop)
	file = fmt.Sprintf("%s/dtmcli.barrier.%s.sql", dtmutil.GetSQLDir(), BusiConf.Driver)
	dtmutil.RunSQLScript(BusiConf, file, skipDrop)
	_, err := RedisGet().FlushAll(context.Background()).Result() // redis barrier need clear
	dtmimp.E2P(err)
	SetRedisBothAccount(10000, 10000)
	SetupMongoBarrierAndBusi()
}
