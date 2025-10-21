package cmd

import (
	"fmt"
	"os"
	"os/user"
	"time"

	"github.com/spf13/cobra"
)

var (
	cfgFile     string
	sshUser     string
	sshKey      string
	concurrency int
	timeout     time.Duration
	yes         bool
	dryRun      bool
	image       string
)

var RootCmd = &cobra.Command{
	Use:   "scalogctl",
	Short: "Launch/stop Scalog cluster via SSH + Docker",
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", ".scalog.yaml", "Path to .scalog.yaml (used locally AND on all remote nodes)")
	cur, _ := user.Current()
	defaultUser := ""
	if cur != nil {
		defaultUser = cur.Username
	}
	RootCmd.PersistentFlags().StringVar(&sshUser, "user", defaultUser, "SSH username")
	RootCmd.PersistentFlags().StringVar(&sshKey, "identity", "", "SSH private key path (e.g. ~/.ssh/id_rsa)")
	RootCmd.PersistentFlags().IntVar(&concurrency, "concurrency", 8, "Max concurrent SSH/SCP sessions")
	RootCmd.PersistentFlags().DurationVar(&timeout, "timeout", 60*time.Second, "Per-host SSH/SCP timeout")
	RootCmd.PersistentFlags().BoolVar(&yes, "yes", false, "Do not prompt for confirmation")
	RootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Only print commands")
	RootCmd.PersistentFlags().StringVar(&image, "image", "chn0318/scalog:v1.0", "Scalog docker image")
}
