package storage

import (
	"fmt"
)

type Partition struct {
	path           string
	nextLSN        int64
	segments       []*Segment
	activeSegment  *Segment
	activebaseLSN  int64
	segLen         int32
	prevSSN        int32
	recordSegments []*Segment
}

func NewPartition(path string, segLen int32) (*Partition, error) {
	var err error
	p := &Partition{path: path, nextLSN: 0, activebaseLSN: 0}
	p.segments = make([]*Segment, 0)
	p.recordSegments = make([]*Segment, 0)
	p.activeSegment, err = NewSegment(path, p.activebaseLSN)
	p.recordSegments = append(p.recordSegments, p.activeSegment)
	p.prevSSN = 0
	if err != nil {
		return nil, err
	}
	p.segLen = segLen
	return p, nil
}

func (p *Partition) Write(record string) (int64, error) {
	lsn := p.nextLSN
	p.nextLSN++
	//create new segment before write
	//fix bug: Assign GSN error, no data in ssn = 0
	if p.prevSSN >= p.segLen-1 {
		err := p.CreateSegment()
		if err != nil {
			return 0, err
		}
	}
	ssn, err := p.activeSegment.Write(record)
	p.prevSSN = ssn
	return lsn, err
}

func (p *Partition) Read(gsn int64) (string, error) {
	return p.ReadGSN(gsn)
}

func (p *Partition) ReadGSN(gsn int64) (string, error) {
	if gsn >= p.activeSegment.baseGSN {
		return p.activeSegment.ReadGSN(gsn)
	}
	f := func(s *Segment) int64 {
		return s.baseGSN
	}
	s := binarySearch(p.segments, f, gsn)
	if s == nil {
		return "", fmt.Errorf("Segment with gsn %v not exist", gsn)
	}
	return s.ReadGSN(gsn)
}

func (p *Partition) ReadLSN(lsn int64) (string, error) {
	if lsn >= p.activeSegment.baseLSN {
		return p.activeSegment.ReadLSN(lsn)
	}
	f := func(s *Segment) int64 {
		return s.baseLSN
	}
	s := binarySearch(p.segments, f, lsn)
	return s.ReadLSN(lsn)
}

func binarySearch(segs []*Segment, get func(*Segment) int64, target int64) *Segment {
	// Currently the function is scanning the list.
	// TODO: implement binary search.
	if len(segs) == 0 {
		return nil
	}
	for i, s := range segs {
		if get(s) > target {
			return segs[i-1]
		}
	}
	return segs[len(segs)-1]
}

func (p *Partition) CreateSegment() error {
	var err error
	p.activebaseLSN += int64(p.segLen)
	p.segments = append(p.segments, p.activeSegment)
	p.activeSegment, err = NewSegment(p.path, p.activebaseLSN)
	p.recordSegments = append(p.recordSegments, p.activeSegment)
	return err
}

func (p *Partition) Assign(lsn int64, length int32, gsn int64) error {
	startSegmet := lsn / int64(p.segLen)
	startSSN := lsn - startSegmet*int64(p.segLen)
	for {
		assignLength := min(p.segLen-int32(startSSN), length)
		err := p.recordSegments[startSegmet].Assign(int32(startSSN), assignLength, gsn)
		if err != nil {
			return err
		}
		startSegmet++
		startSSN = 0
		gsn += int64(assignLength)
		length -= assignLength
		if length <= 0 {
			return nil
		}
	}
	// return p.recordSegments[lsn/int64(p.segLen)].Assign(int32(lsn-((lsn/int64(p.segLen))*int64(p.segLen))), length, gsn)
}
