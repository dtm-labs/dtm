package server

import (
	"github.com/spf13/cobra"

	"github.com/yedf/dtm/dtmsvr"
)

var (
	port  int
	isDev bool
)

func init() {
	Cmd.Flags().IntVar(&port, "port", 8080, "listen port")
	Cmd.Flags().BoolVar(&isDev, "import", false, "populate db test data")
}

var Cmd = &cobra.Command{
	Use:   "server",
	Short: "run dtm server",
	Long:  `run dtm server`,
	Run: func(cmd *cobra.Command, args []string) {
		if isDev {
			dtmsvr.PopulateDB(true)
		}
		dtmsvr.StartSvr(port)          // 启动dtmsvr的api服务
		go dtmsvr.CronExpiredTrans(-1) // 启动dtmsvr的定时过期查询
	},
}
