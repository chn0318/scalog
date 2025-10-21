package config

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

// Config holds both well-known keys and dynamic node IP entries.
// We keep a raw KV so we can flexibly parse keys like order-0-ip, data-1-0-ip, etc.
type Config struct {
	Raw map[string]any

	// Common ports/settings (read if present; still optional because container reads from mounted file)
	OrderPort              int
	RaftPort               int
	OrderReplicationFactor int
	OrderBatchingInterval  string

	DataPort              int
	DataReplicationFactor int
	DataBatchingInterval  string

	DiscPort int
	DiscIP   string

	// Mount path of the local YAML into container (host side). Defaults to the provided cfg path.
	cfgPath string
}

// OrderIPs indexed by oid
type OrderIPs map[int]string

// DataIPs indexed by sid -> rid
type DataIPs map[int]map[int]string

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	all := v.AllSettings()
	c := &Config{
		Raw:     all,
		cfgPath: path,
	}
	// Best-effort typed reads (optional)
	c.OrderPort = v.GetInt("order-port")
	c.RaftPort = v.GetInt("raft-port")
	c.OrderReplicationFactor = v.GetInt("order-replication-factor")
	c.OrderBatchingInterval = v.GetString("order-batching-interval")
	c.DataPort = v.GetInt("data-port")
	c.DataReplicationFactor = v.GetInt("data-replication-factor")
	c.DataBatchingInterval = v.GetString("data-batching-interval")
	c.DiscPort = v.GetInt("disc-port")
	c.DiscIP = v.GetString("disc-ip")
	return c, nil
}

// ConfigMountPath returns the absolute path we will mount into container (host side path).
func (c *Config) ConfigMountPath() string {
	// normalize to absolute path for docker -v
	abs, err := filepath.Abs(c.cfgPath)
	if err != nil {
		return c.cfgPath
	}
	return abs
}

var (
	reOrder = regexp.MustCompile(`^order-(\d+)-ip$`)
	reData  = regexp.MustCompile(`^data-(\d+)-(\d+)-ip$`)
)

// ExtractOrderIPs scans raw kvs and returns map[oid]ip
func (c *Config) ExtractOrderIPs() (OrderIPs, error) {
	out := make(OrderIPs)
	for k, v := range c.Raw {
		m := reOrder.FindStringSubmatch(strings.ToLower(k))
		if m == nil {
			continue
		}
		oid, _ := strconv.Atoi(m[1])
		ip := fmt.Sprintf("%v", v)
		if ip == "" {
			return nil, fmt.Errorf("empty ip for key %q", k)
		}
		out[oid] = ip
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no order-*-ip entries found")
	}
	return out, nil
}

// ExtractDataIPs scans raw kvs and returns map[sid]map[rid]ip
func (c *Config) ExtractDataIPs() (DataIPs, error) {
	out := make(DataIPs)
	for k, v := range c.Raw {
		m := reData.FindStringSubmatch(strings.ToLower(k))
		if m == nil {
			continue
		}
		sid, _ := strconv.Atoi(m[1])
		rid, _ := strconv.Atoi(m[2])
		if _, ok := out[sid]; !ok {
			out[sid] = make(map[int]string)
		}
		ip := fmt.Sprintf("%v", v)
		if ip == "" {
			return nil, fmt.Errorf("empty ip for key %q", k)
		}
		out[sid][rid] = ip
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no data-*-*-ip entries found")
	}
	return out, nil
}
