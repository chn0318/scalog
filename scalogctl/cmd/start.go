package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/scalog/scalog/scalogctl/pkg/config"
	"github.com/scalog/scalog/scalogctl/pkg/plan"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a Scalog cluster based on .scalog.yaml",
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

		fmt.Printf("Targets: order=%d, data=%d, discovery=%d\n",
			len(lp.OrderTargets), len(lp.DataTargets), len(lp.DiscoveryTargets))
		fmt.Printf("Unified config path: %s\n", absCfg)

		if !yes {
			if !askForYes("SCP config -> start all nodes (order/data/discovery)? [y/N]: ") {
				fmt.Println("Aborted.")
				return nil
			}
		}

		hosts := lp.AllHosts()
		if err := scpConfigToHosts(absCfg, hosts, absCfg); err != nil {
			return err
		}

		if err := runTargets(lp, dryRun); err != nil {
			return err
		}

		fmt.Println("All tasks done.")
		return nil
	},
}

func init() { RootCmd.AddCommand(startCmd) }

func askForYes(prompt string) bool {
	fmt.Print(prompt)
	in := bufio.NewScanner(os.Stdin)
	if in.Scan() {
		t := strings.TrimSpace(strings.ToLower(in.Text()))
		return t == "y" || t == "yes"
	}
	return false
}

type target struct {
	Host string
	Cmd  string
	Role string
}

func runTargets(lp *plan.LaunchPlan, dry bool) error {
	type target struct{ Host, Cmd, Role string }

	all := make([]target, 0,
		len(lp.OrderTargets)+len(lp.DataTargets)+len(lp.DiscoveryTargets))

	for _, t := range lp.OrderTargets {
		all = append(all, target{Host: t.Host, Cmd: t.Command, Role: "order"})
	}
	for _, t := range lp.DataTargets {
		all = append(all, target{Host: t.Host, Cmd: t.Command, Role: "data"})
	}
	for _, t := range lp.DiscoveryTargets {
		all = append(all, target{Host: t.Host, Cmd: t.Command, Role: "discovery"})
	}

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	errCh := make(chan error, len(all))

	for _, t := range all {
		t := t
		wg.Add(1)
		go func() {
			defer wg.Done()
			if dry {
				fmt.Printf("[dry-run] %s@%s: %s\n", sshUser, t.Host, t.Cmd)
				return
			}
			sem <- struct{}{}
			defer func() { <-sem }()
			if err := sshRun(t.Host, t.Cmd); err != nil {
				errCh <- fmt.Errorf("[%s %s] %w", t.Role, t.Host, err)
			} else {
				fmt.Printf("[ok] %s %s\n", t.Role, t.Host)
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for e := range errCh {
		return e
	}
	return nil
}

func sshRun(host, remoteCmd string) error {
	args := []string{"-o", "StrictHostKeyChecking=no"}
	if sshKey != "" {
		args = append(args, "-i", sshKey)
	}
	dst := host
	if sshUser != "" {
		dst = sshUser + "@" + host
	}
	args = append(args, dst, remoteCmd)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	c := exec.CommandContext(ctx, "ssh", args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func scpConfigToHosts(localCfg string, hosts []string, remoteCfg string) error {
	abs, _ := filepath.Abs(localCfg)
	fmt.Printf("SCP %s -> %d hosts (%s)\n", abs, len(hosts), remoteCfg)

	if dryRun {
		for _, h := range hosts {
			dst := h + ":" + remoteCfg
			if sshUser != "" {
				dst = sshUser + "@" + dst
			}
			fmt.Printf("[dry-run] scp %s -> %s\n", abs, dst)
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
			if err := scpOne(abs, h, remoteCfg); err != nil {
				errCh <- fmt.Errorf("[scp %s] %w", h, err)
			} else {
				fmt.Printf("[scp ok] %s\n", h)
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for e := range errCh {
		return e
	}
	return nil
}

func scpOne(local, host, remote string) error {
	args := []string{"-o", "StrictHostKeyChecking=no"}
	if sshKey != "" {
		args = append(args, "-i", sshKey)
	}
	src := local
	dst := host + ":" + remote
	if sshUser != "" {
		dst = sshUser + "@" + dst
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var c *exec.Cmd
	if sshKey != "" {
		c = exec.CommandContext(ctx, "scp", "-o", "StrictHostKeyChecking=no", "-i", sshKey, src, dst)
	} else {
		c = exec.CommandContext(ctx, "scp", "-o", "StrictHostKeyChecking=no", src, dst)
	}
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
