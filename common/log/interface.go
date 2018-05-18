package log

type stclogging interface {
	Trace(a ...interface{})
	Tracef(format string, a ...interface{})
	Debug(a ...interface{})
	Debugf(format string, a ...interface{})
	Info(a ...interface{})
	Infof(format string, a ...interface{})
	Warn(a ...interface{})
	Warnf(format string,a ...interface{})
	Error(a ...interface{})
	Errorf(format string,a ...interface{})
}
