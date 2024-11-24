// nolint
package cmd

import (
	"github.com/scalog/scalog/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var perfCmd = &cobra.Command{
	Use:   "perf",
	Short: "The perf test for scalog",
	Long:  "The bandwidth and latency benchmark for scalog system",
	Run: func(cmd *cobra.Command, args []string) {
		client.Perf()
	},
}

func init() {
	RootCmd.AddCommand(perfCmd)
	perfCmd.PersistentFlags().IntP("threads", "t", 1, "Number of concurrent threads (default is 1)")
	viper.BindPFlag("threads", perfCmd.PersistentFlags().Lookup("threads"))
}
