package cmd

import (
	"fmt"
	"strings"
	"sync"

	"github.com/chn0318/scalog/scalogctl/pkg/config"
	"github.com/chn0318/scalog/scalogctl/pkg/plan"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop all Scalog containers on every node (order/data/discovery)",
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		absCfg := conf.ConfigMountPath()
		lp, err := plan.BuildLaunchPlan(conf, image, absCfg)
		if err != nil {
			return fmt.Errorf("build plan: %w", err)
		}

		hosts := lp.AllHosts()
		fmt.Printf("Stopping containers (image=%s) on %d hosts\n", image, len(hosts))

		if dryRun {
			for _, h := range hosts {
				fmt.Printf("[dry-run] ssh %s@%s: %s\n", sshUser, h, remoteStopCmd(image))
			}
			return nil
		}

		sem := make(chan struct{}, concurrency)
		var wg sync.WaitGroup
		errCh := make(chan error, len(hosts))

		for _, h := range hosts {
			h := h
			wg.Add(1)
			go func() {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				if err := sshRun(h, remoteStopCmd(image)); err != nil {
					errCh <- fmt.Errorf("[stop %s] %w", h, err)
				} else {
					fmt.Printf("[stop ok] %s\n", h)
				}
			}()
		}
		wg.Wait()
		close(errCh)
		for e := range errCh {
			return e
		}
		return nil
	},
}

func init() { RootCmd.AddCommand(stopCmd) }

func remoteStopCmd(img string) string {
	return `bash -lc '` + strings.Join([]string{
		`ids=$(docker ps --filter "ancestor=` + img + `" -q)`,
		`if [ -n "$ids" ]; then docker stop $ids; fi`,
	}, " && ") + `'`
}
