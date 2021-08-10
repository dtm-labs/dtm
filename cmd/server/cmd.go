package server

import (
	"github.com/spf13/cobra"
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
		main()
	},
}
