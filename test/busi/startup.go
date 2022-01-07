package busi

import (
	"fmt"

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
}
