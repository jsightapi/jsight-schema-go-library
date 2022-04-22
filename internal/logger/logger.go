package logger

type Logger interface {
	Default(string)
	Notice(string)
	Info(string)
	Warning(string)
	Error(string)
}
