package log

// Logger 日志处理接口
type Logger interface {
	// Debug 输出Debug信息
	Debug(info ...interface{})
	// Info 输出Info信息
	Info(info ...interface{})
	// Warning 输出Warning信息
	Warning(info ...interface{})
	// Error 输出Error信息
	Error(info ...interface{})
}

// defaultLogger 默认日志
var defaultLogger Logger = StdLog{}

// Debug 输出Debug信息
func Debug(info ...interface{}) {
	defaultLogger.Debug(info...)
}
// Info 输出Info信息
func Info(info ...interface{}) {
	defaultLogger.Info(info...)
}
// Warning 输出Warning信息
func Warning(info ...interface{}) {
	defaultLogger.Warning(info...)
}
// Error 输出Error信息
func Error(info ...interface{}) {
	defaultLogger.Error(info...)
}

// SetLogger 设置默认日志处理实例
func SetLogger(logger Logger)  {
	defaultLogger = logger
}
