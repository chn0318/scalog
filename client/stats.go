// Made with Gemini
// Note that this assumes that the base scalog directory is nested in the
// first layer of the pringles main project directory
package client

/*
#cgo CPPFLAGS: -I../../code/include
#cgo LDFLAGS: -L../../build -lstats_measure -lspdlog -lfmt -lstdc++
#include "stats_wrapper.h"
#include <stdlib.h>
*/
import "C"
import (
	"unsafe"
)

type Stats struct {
	ptr C.StatsPtr
}

func NewStats(jsonName string, clientIP string, index int64) *Stats {
	cJsonName := C.CString(jsonName)
	cClientIP := C.CString(clientIP)
	defer C.free(unsafe.Pointer(cJsonName))
	defer C.free(unsafe.Pointer(cClientIP))

	return &Stats{
		ptr: C.NewStats(0, false, cJsonName, C.uint64_t(index), cClientIP),
	}
}

func (s *Stats) Close() {
	if s.ptr != nil {
		C.FreeStats(s.ptr)
		s.ptr = nil
	}
}

func (s *Stats) AddOp() {
	C.StatsAddOp(s.ptr)
}

func (s *Stats) AddDuration(start_time float64) {
	C.StatsAddDuration(s.ptr, C.double(start_time))
}

func (s *Stats) ExportResults(elapsedSeconds int64) {
	C.StatsGetThroughput(s.ptr, C.uint64_t(elapsedSeconds))
	C.StatsGetAvgLatency(s.ptr)
	C.StatsExportResultsToJson(s.ptr)
}
