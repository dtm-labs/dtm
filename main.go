package main

import (
	_ "net/http/pprof" // 注册 pprof 接口

	"github.com/spf13/cobra"
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/yedf/dtm/cmd/example"
	"github.com/yedf/dtm/cmd/server"
	"github.com/yedf/dtm/cmd/version"
)

var (
	a string
	v string
	c string
	d string
)

func init() {
	version.BinAppName = a
	version.BinBuildCommit = c
	version.BinBuildVersion = v
	version.BinBuildDate = d
}

func main() {
	nopLog := func(string, ...interface{}) {}
	maxprocs.Set(maxprocs.Logger(nopLog))
	root := cobra.Command{Use: "dtm"}
	root.AddCommand(
		server.Cmd,
		example.Cmd,
		version.Cmd,
	)
	root.Execute()
}
