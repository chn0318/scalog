// nolint
package cmd

import (
	"time"
	"fmt"

	"github.com/chn0318/scalog/client"
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
	perfCmd.PersistentFlags().IntP("size", "s", 1024, "Size of append message (default is 1024 Bytes)")
	perfCmd.PersistentFlags().DurationP("duration", "d", 30*time.Second, "total run time, e.g. 45s, 2m, 1h (default is 30s)")

	fmt.Printf("In the init function!\n")
	viper.BindPFlag("threads", perfCmd.PersistentFlags().Lookup("threads"))
	viper.BindPFlag("size", perfCmd.PersistentFlags().Lookup("size"))
	viper.BindPFlag("duration", perfCmd.PersistentFlags().Lookup("duration"))
}
