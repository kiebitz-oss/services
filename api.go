package services

type Context interface {
	Result(data interface{}) Response
	Error(code int, message string, data interface{}) Response
	InternalError() Response
	InvalidParams(err error) Response
	MethodNotFound() Response
	NotFound() Response
	Acknowledge() Response
	Nil() Response
}

type Response interface {
}

type Request interface {
}
