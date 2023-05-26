package stats

// logger is an implementation to logs information inside this package
type logger interface {
	Info(format string, args ...interface{})
	Error(format string, args ...interface{})
}
