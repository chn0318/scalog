package address

import (
	"fmt"
)

type LocalOrderAddr struct {
	port uint16
}

func NewLocalOrderAddr(port uint16) *LocalOrderAddr {
	return &LocalOrderAddr{port}
}

func (s *LocalOrderAddr) UpdateAddr(port uint16) {
	s.port = port
}

func (s *LocalOrderAddr) Get() string {
	return fmt.Sprintf("10.10.1.2:%v", s.port)
}
