package logger

import (
	"log"
	"os"
	"runtime"
	"strings"
)

type Level int

const (
	levelError Level = iota
	levelInfo
	levelDebug
)

func parseLevel(s string) Level {
	switch strings.ToLower(s) {
	case "debug":
		return levelDebug
	case "info":
		return levelInfo
	default:
		return levelError
	}
}

type Logger struct {
	l     *log.Logger
	level Level
}

func New(lvl string) *Logger {
	return &Logger{
		l:     log.New(os.Stdout, "", log.LstdFlags),
		level: parseLevel(lvl),
	}
}

func (lg *Logger) Debug(msg string) {
	if lg.level >= levelDebug {
		_, file, line, _ := runtime.Caller(1)
		lg.l.Printf("DEBUG: %s:%d %s\n", short(file), line, msg)
	}
}

func (lg *Logger) Info(msg string) {
	if lg.level >= levelInfo {
		lg.l.Println("INFO:", msg)
	}
}

func (lg *Logger) Error(msg string) {
	lg.l.Println("ERROR:", msg)
}

func short(p string) string {
	if idx := strings.LastIndex(p, "/"); idx != -1 {
		return p[idx+1:]
	}
	return p
}
