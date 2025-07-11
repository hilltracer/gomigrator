package logger

import (
	"log"
	"os"
	"strings"
)

type level int

const (
	errorLvl level = iota
	infoLvl
	debugLvl
)

func parseLevel(s string) level {
	switch strings.ToLower(s) {
	case "debug":
		return debugLvl
	case "info":
		return infoLvl
	default:
		return errorLvl
	}
}

type Logger struct {
	l     *log.Logger
	level level
}

func New(lvl string) *Logger {
	return &Logger{
		l:     log.New(os.Stdout, "", log.LstdFlags),
		level: parseLevel(lvl),
	}
}

func (l *Logger) Info(msg string) {
	if l.level >= infoLvl {
		l.l.Println("INFO:", msg)
	}
}
func (l *Logger) Error(msg string) { l.l.Println("ERROR:", msg) }
