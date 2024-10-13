package address

import (
	"fmt"
)

type LocalDataAddr struct {
	numReplica int32
	basePort   uint16
}

var ipMap = map[uint16]string{
	0: "10.10.1.5",
	1: "10.10.1.6",
	2: "10.10.1.7",
	3: "10.10.1.8",
}

func NewLocalDataAddr(numReplica int32, basePort uint16) *LocalDataAddr {
	return &LocalDataAddr{numReplica, basePort}
}

func (s *LocalDataAddr) UpdateBasePort(basePort uint16) {
	s.basePort = basePort
}

func (s *LocalDataAddr) Get(sid, rid int32) string {
	port := s.basePort + uint16(sid*s.numReplica+rid)

	return fmt.Sprintf("%v:%v", ipMap[uint16(sid*s.numReplica+rid)], port)
}
