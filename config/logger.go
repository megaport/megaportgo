package config

type Logger interface {
	Trace(args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})

	Tracef(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})

	Traceln(args ...interface{})
	Debugln(args ...interface{})
	Infoln(args ...interface{})
	Warnln(args ...interface{})
	Errorln(args ...interface{})
}
