package config

import (
	"fmt"
	"log"
	"strings"
)

type Level int8

const (
	TraceLevel Level = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	Off
)

func (l Level) String() string {
	switch l {
	case TraceLevel:
		return "trace"
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case Off:
		return "off"
	default:
		return "unknown"
	}
}

func StringToLogLevel(level string) Level {
	switch level {
	case "TRACE":
		return TraceLevel
	case "DEBUG":
		return DebugLevel
	case "INFO":
		return InfoLevel
	case "WARN":
		return WarnLevel
	case "ERROR":
		return ErrorLevel
	default:
		return Off
	}
}

type DefaultLogger struct {
	level Level
}

func NewDefaultLogger() *DefaultLogger {
	d := DefaultLogger{level: DebugLevel}
	return &d
}

func (d *DefaultLogger) SetLevel(l Level) {
	d.level = l
}

func (d *DefaultLogger) log(level Level, args ...interface{}) {
	if level >= d.level {
		msg := fmt.Sprint(args...)
		log.Print(fmt.Sprintf("%s: %s", strings.ToUpper(level.String()), msg))
	}
}

// Emit the message and args at DEBUG level
func (d *DefaultLogger) Debug(args ...interface{}) {
	d.log(DebugLevel, args...)
}

// Emit the message and args at TRACE level
func (d *DefaultLogger) Trace(args ...interface{}) {
	d.log(TraceLevel, args...)
}

// Emit the message and args at INFO level
func (d *DefaultLogger) Info(args ...interface{}) {
	d.log(InfoLevel, args...)
}

// Emit the message and args at WARN level
func (d *DefaultLogger) Warn(args ...interface{}) {
	d.log(WarnLevel, args...)
}

// Emit the message and args at ERROR level
func (d *DefaultLogger) Error(args ...interface{}) {
	d.log(ErrorLevel, args...)
}

func (d *DefaultLogger) logf(level Level, format string, args ...interface{}) {
	if level >= d.level {
		msg := fmt.Sprintf(format, args...)
		log.Print(fmt.Sprintf("%s: %s", strings.ToUpper(level.String()), msg))
	}
}

// Emit the message and args at DEBUG level
func (d *DefaultLogger) Debugf(format string, args ...interface{}) {
	d.logf(DebugLevel, format, args...)
}

// Emit the message and args at TRACE level
func (d *DefaultLogger) Tracef(format string, args ...interface{}) {
	d.logf(TraceLevel, format, args...)
}

// Emit the message and args at INFO level
func (d *DefaultLogger) Infof(format string, args ...interface{}) {
	d.logf(InfoLevel, format, args...)
}

// Emit the message and args at WARN level
func (d *DefaultLogger) Warnf(format string, args ...interface{}) {
	d.logf(WarnLevel, format, args...)
}

// Emit the message and args at ERROR level
func (d *DefaultLogger) Errorf(format string, args ...interface{}) {
	d.logf(ErrorLevel, format, args...)
}

func (d *DefaultLogger) logln(level Level, args ...interface{}) {
	if level >= d.level {
		msg := fmt.Sprintln(args...)
		prefix := fmt.Sprintf("%s:", strings.ToUpper(level.String()))

		log.Println(prefix, msg[:len(msg)-1])
	}
}

// Emit the message and args at DEBUG level
func (d *DefaultLogger) Debugln(args ...interface{}) {
	d.logln(DebugLevel, args...)
}

// Emit the message and args at TRACE level
func (d *DefaultLogger) Traceln(args ...interface{}) {
	d.logln(TraceLevel, args...)
}

// Emit the message and args at INFO level
func (d *DefaultLogger) Infoln(args ...interface{}) {
	d.logln(InfoLevel, args...)
}

// Emit the message and args at WARN level
func (d *DefaultLogger) Warnln(args ...interface{}) {
	d.logln(WarnLevel, args...)
}

// Emit the message and args at ERROR level
func (d *DefaultLogger) Errorln(args ...interface{}) {
	d.logln(ErrorLevel, args...)
}
