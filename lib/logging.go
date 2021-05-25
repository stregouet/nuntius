package lib

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type Level int

const (
	DEBUG Level = iota + 1
	LOG
	WARN
	ERROR
)

func (l Level) ToString() string {
	switch l {
	case DEBUG:
		return "debug"
	case LOG:
		return "log"
	case WARN:
		return "warn"
	case ERROR:
		return "error"
	default:
		return fmt.Sprintf("unknown (%d)", l)
	}
}


func LogParseLevel(lvl string) (Level, error) {
	var l Level
	switch strings.ToLower(lvl) {
	case "debug":
		l = DEBUG
	case "log":
		l = LOG
	case "warn":
		l = WARN
	case "error":
		l = ERROR
	}
	if l == 0 {
		return l, fmt.Errorf("unknown level `%s`", lvl)
	}
	return l, nil
}

type Logger struct {
	lvl Level
	*log.Logger
}

func NewLogger(level, output string) (*Logger, error) {
	lvl, err := LogParseLevel(level)
	if err != nil {
		return nil, err
	}
	if output != "" {
		logOutput, err := os.OpenFile(output, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return nil, err
		}
		flog := log.New(logOutput, "nuntius: ", log.Lshortfile|log.Ltime)
		return &Logger{lvl, flog}, nil
	}
	return &Logger{lvl: 100}, nil
}

func (l *Logger) isEnable(lvl Level) bool {
	return lvl >= l.lvl
}



func (l *Logger) output(lvl Level, msg string) {
	if !l.isEnable(lvl) {
		return
	}
	l.Output(3, "level=" + lvl.ToString() + " " + msg)
}

func (l *Logger) Debugf(msg string, args ...interface{}) {
	l.output(DEBUG, fmt.Sprintf(msg, args...))
}
func (l *Logger) Logf(msg string, args ...interface{}) {
	l.output(LOG, fmt.Sprintf(msg, args...))
}
func (l *Logger) Warnf(msg string, args ...interface{}) {
	l.output(WARN, fmt.Sprintf(msg, args...))
}
func (l *Logger) Errorf(msg string, args ...interface{}) {
	l.output(ERROR, fmt.Sprintf(msg, args...))
}

func (l *Logger) Debug(msg string, args ...interface{}) {
	l.output(DEBUG, msg)
}
func (l *Logger) Log(msg string, args ...interface{}) {
	l.output(LOG, msg)
}
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.output(WARN, msg)
}
func (l *Logger) Error(msg string, args ...interface{}) {
	l.output(ERROR, msg)
}
