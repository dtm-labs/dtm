package main

import (
	_ "net/http/pprof" // 注册 pprof 接口

	"github.com/spf13/cobra"
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/yedf/dtm/cmd/example"
	"github.com/yedf/dtm/cmd/server"
)

func main() {
	nopLog := func(string, ...interface{}) {}
	maxprocs.Set(maxprocs.Logger(nopLog))
	root := cobra.Command{Use: "dtm"}
	root.AddCommand(
		server.Cmd,
		example.Cmd,
	)
	root.Execute()
}
