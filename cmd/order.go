// nolint
package cmd

import (
	"github.com/scalog/scalog/order"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// orderCmd represents the order command
var orderCmd = &cobra.Command{
	Use:   "order",
	Short: "The ordering layer",
	Long:  `The ordering layer`,
	Run: func(cmd *cobra.Command, args []string) {
		order.Start()
	},
}

func init() {
	RootCmd.AddCommand(orderCmd)
	orderCmd.PersistentFlags().IntP("oid", "i", 0, "Replica index")
	orderCmd.PersistentFlags().BoolP("monitor", "m", false, "monitor the channel")
	viper.BindPFlag("oid", orderCmd.PersistentFlags().Lookup("oid"))
	viper.BindPFlag("monitor", orderCmd.PersistentFlags().Lookup("monitor"))
}
