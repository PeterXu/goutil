package goutil

import (
	"fmt"
	"log"
)

type LoggingLevel uint32

const (
	LoggingLevelUnknown LoggingLevel = iota
	LoggingLevelVerbose
	LoggingLevelInfo
	LoggingLevelWarn
	LoggingLevelError
	LoggingLevelMaxValue
)

type Logging struct {
	TAG   string
	Level LoggingLevel
}

func (l *Logging) SetLevel(level LoggingLevel) {
	if level > LoggingLevelUnknown && level < LoggingLevelMaxValue {
		l.Level = level
	}
}

func (l *Logging) IsFiltered(level LoggingLevel) bool {
	if l.Level == LoggingLevelUnknown {
		l.Level = LoggingLevelInfo
	}
	return (level < l.Level)
}

func (l Logging) Println(v ...interface{}) {
	if !l.IsFiltered(LoggingLevelInfo) {
		log.Printf("[INFO][%s] %s\n", l.TAG, fmt.Sprint(v...))
	}
}

func (l Logging) Printf(format string, v ...interface{}) {
	if !l.IsFiltered(LoggingLevelInfo) {
		log.Printf("[INFO][%s] %s", l.TAG, fmt.Sprintf(format, v...))
	}
}

func (l Logging) Warnln(v ...interface{}) {
	if !l.IsFiltered(LoggingLevelWarn) {
		log.Printf("[WARN][%s] %s\n", l.TAG, fmt.Sprint(v...))
	}
}

func (l Logging) Warnf(format string, v ...interface{}) {
	if !l.IsFiltered(LoggingLevelWarn) {
		log.Printf("[WARN][%s] %s", l.TAG, fmt.Sprintf(format, v...))
	}
}

func (l Logging) Errorln(v ...interface{}) {
	if !l.IsFiltered(LoggingLevelError) {
		log.Printf("[ERRO][%s] %s\n", l.TAG, fmt.Sprint(v...))
	}
}

func (l Logging) Errorf(format string, v ...interface{}) {
	if !l.IsFiltered(LoggingLevelError) {
		log.Printf("[ERRO][%s] %s\n", l.TAG, fmt.Sprintf(format, v...))
	}
}
