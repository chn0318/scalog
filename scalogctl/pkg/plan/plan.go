package plan

import (
	"fmt"
	"sort"

	"github.com/scalog/scalog/scalogctl/pkg/config"
)

type RemoteCmd struct {
	Host    string
	Command string
}

type LaunchPlan struct {
	OrderTargets     []RemoteCmd
	DataTargets      []RemoteCmd
	DiscoveryTargets []RemoteCmd
}

func (lp *LaunchPlan) AllHosts() []string {
	m := map[string]struct{}{}
	for _, t := range lp.OrderTargets {
		m[t.Host] = struct{}{}
	}
	for _, t := range lp.DataTargets {
		m[t.Host] = struct{}{}
	}
	for _, t := range lp.DiscoveryTargets {
		m[t.Host] = struct{}{}
	}
	out := make([]string, 0, len(m))
	for h := range m {
		out = append(out, h)
	}
	sort.Strings(out)
	return out
}

func BuildLaunchPlan(c *config.Config, image, cfgPath string) (*LaunchPlan, error) {
	orderIPs, err := c.ExtractOrderIPs()
	if err != nil {
		return nil, err
	}
	dataIPs, err := c.ExtractDataIPs()
	if err != nil {
		return nil, err
	}

	// --- order ---
	orderIdx := make([]int, 0, len(orderIPs))
	for oid := range orderIPs {
		orderIdx = append(orderIdx, oid)
	}
	sort.Ints(orderIdx)

	var orderTargets []RemoteCmd
	for _, oid := range orderIdx {
		host := orderIPs[oid]
		cmd := fmt.Sprintf(
			`bash -lc 'docker run -d --rm --network=host -v %s:/root/.scalog.yaml %s order --oid=%d'`,
			cfgPath, image, oid,
		)
		orderTargets = append(orderTargets, RemoteCmd{Host: host, Command: cmd})
	}

	// --- data ---
	sids := make([]int, 0, len(dataIPs))
	for sid := range dataIPs {
		sids = append(sids, sid)
	}
	sort.Ints(sids)

	var dataTargets []RemoteCmd
	for _, sid := range sids {
		rids := make([]int, 0, len(dataIPs[sid]))
		for rid := range dataIPs[sid] {
			rids = append(rids, rid)
		}
		sort.Ints(rids)
		for _, rid := range rids {
			host := dataIPs[sid][rid]
			cmd := fmt.Sprintf(
				`bash -lc 'docker run -d --rm --network=host -v %s:/root/.scalog.yaml %s data --sid=%d --rid=%d'`,
				cfgPath, image, sid, rid,
			)
			dataTargets = append(dataTargets, RemoteCmd{Host: host, Command: cmd})
		}
	}

	// --- discovery ---
	var discTargets []RemoteCmd
	if c.DiscIP != "" {
		host := c.DiscIP
		cmd := fmt.Sprintf(
			`bash -lc 'docker run -d --rm --network=host -v %s:/root/.scalog.yaml %s discovery'`,
			cfgPath, image,
		)
		discTargets = append(discTargets, RemoteCmd{Host: host, Command: cmd})
	}

	return &LaunchPlan{
		OrderTargets:     orderTargets,
		DataTargets:      dataTargets,
		DiscoveryTargets: discTargets,
	}, nil
}
