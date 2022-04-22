package logger

type LogToNull struct{}

func (LogToNull) Default(string) {}
func (LogToNull) Notice(string)  {}
func (LogToNull) Info(string)    {}
func (LogToNull) Warning(string) {}
func (LogToNull) Error(string)   {}
