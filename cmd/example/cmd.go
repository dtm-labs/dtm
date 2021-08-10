package example

import (
	"github.com/spf13/cobra"
)

var (
	port        int
	examplePort int
	isDev       bool
	tutorial    string
)

func init() {
	Cmd.Flags().IntVar(&port, "port", 8080, "listen dtm server port")
	Cmd.Flags().IntVar(&examplePort, "eport", 8081, "listen dtm example port")
	Cmd.Flags().BoolVar(&isDev, "import", true, "populate db test data")
	Cmd.Flags().StringVar(&tutorial, "tutorial", "quick_start", "tutorial value should be in [quick_start,qs,xa,saga,tcc,msg,all,saga_barrier,tcc_barrier]")
}

var Cmd = &cobra.Command{
	Use:   "example",
	Short: "run dtm example server",
	Long:  `run dtm example server`,
	Run: func(cmd *cobra.Command, args []string) {
		main()
	},
}
