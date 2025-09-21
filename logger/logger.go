// package logger
package logger

import (
	"io"
	"log"
	"os"
	"sync/atomic"
)

type Level int32

const (
	LevelDebug Level = 10
	LevelInfo  Level = 20
	LevelWarn  Level = 30
	LevelError Level = 40
	LevelNone  Level = 99
)

var (
	lg    = log.New(os.Stdout, "[scalog] ", log.Ldate|log.Lmicroseconds)
	level atomic.Int32
)

func init() {
	level.Store(int32(LevelInfo))
}

func SetOutput(w io.Writer) { lg.SetOutput(w) }

func SetLevel(l Level) { level.Store(int32(l)) }

func SetLevelFromString(s string) {
	switch s {
	case "debug":
		SetLevel(LevelDebug)
	case "info", "":
		SetLevel(LevelInfo)
	case "warn", "warning":
		SetLevel(LevelWarn)
	case "error":
		SetLevel(LevelError)
	case "none", "quiet", "silent":
		SetLevel(LevelNone)
	default:
		SetLevel(LevelInfo)
	}
}

func enabled(l Level) bool {
	cur := Level(level.Load())
	return l >= cur && cur != LevelNone
}

func Printf(format string, v ...interface{}) {
	if enabled(LevelInfo) {
		lg.Printf(format, v...)
	}
}

func Debugf(format string, v ...interface{}) {
	if enabled(LevelDebug) {
		lg.Printf("[DEBUG] "+format, v...)
	}
}

func Infof(format string, v ...interface{}) {
	if enabled(LevelInfo) {
		lg.Printf("[INFO] "+format, v...)
	}
}

func Warningf(format string, v ...interface{}) {
	if enabled(LevelWarn) {
		lg.Printf("[WARN] "+format, v...)
	}
}

func Errorf(format string, v ...interface{}) {
	if enabled(LevelError) {
		lg.Printf("[ERROR] "+format, v...)
	}
}

func Fatalf(format string, v ...interface{}) {
	cur := Level(level.Load())
	if cur != LevelNone {
		lg.Fatalf("[FATAL] "+format, v...)
	}
	os.Exit(1)
}

func Panicf(format string, v ...interface{}) {
	if enabled(LevelError) {
		lg.Panicf("[PANIC] "+format, v...)
	}
	panic("scalog panic")
}
