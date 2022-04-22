package logger

import "fmt"

type LogToStdout struct{}

const (
	infoColor    = "\033[1;34m%s\033[0m"
	noticeColor  = "\033[1;36m%s\033[0m"
	warningColor = "\033[1;33m%s\033[0m"
	errorColor   = "\033[1;31m%s\033[0m"
)

func (LogToStdout) Default(s string) { fmt.Println(s) }
func (LogToStdout) Notice(s string)  { fmt.Printf(noticeColor+"\n", s) }
func (LogToStdout) Info(s string)    { fmt.Printf(infoColor+"\n", s) }
func (LogToStdout) Warning(s string) { fmt.Printf(warningColor+"\n", s) }
func (LogToStdout) Error(s string)   { fmt.Printf(errorColor+"\n", s) }
