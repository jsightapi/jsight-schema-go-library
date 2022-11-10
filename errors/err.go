package errors

type Err interface {
	error

	Code() ErrorCode
}

type Error interface {
	Filename() string
	Position() uint
	Line() uint
	Column() uint
	Message() string
	ErrCode() int
	IncorrectUserType() string
}
