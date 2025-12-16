package client

import (
	"math/rand"
	"time"

	"github.com/chn0318/scalog/pkg/view"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type DefaultShardingPolicy struct {
	shardID    int32
	replicaID  int32
	numReplica int32
}

func NewDefaultShardingPolicy(numReplica int32) *DefaultShardingPolicy {
	s := &DefaultShardingPolicy{
		shardID:    -1,
		replicaID:  -1,
		numReplica: numReplica,
	}
	return s
}

func (p *DefaultShardingPolicy) Shard(view *view.View, record string) (int32, int32) {
	if view == nil {
		return -1, -1
	}
	numLiveShards := len(view.LiveShards)
	if numLiveShards < 1 {
		return -1, -1
	}
	rs := rand.Intn(numLiveShards)
	rr := int32(rand.Intn(int(p.numReplica)))
	return view.LiveShards[rs], rr
}
