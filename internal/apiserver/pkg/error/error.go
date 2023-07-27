package service

import "fmt"

type ErrorCode int

const (
	StatusOK        ErrorCode = 0
	Unknown         ErrorCode = -1
	GenericErr      ErrorCode = 7000 // 通用错误
	BadRequestErr   ErrorCode = 7001 // 请求参数错误
	UnauthorizedErr ErrorCode = 7002 // 未授权错误
	AuthServiceErr  ErrorCode = 7003 // auth服务错误
	APIServerErr    ErrorCode = 7004 // apiserver服务错误
	AuditServiceErr ErrorCode = 7005 // audit服务错误
	DMSServiceErr   ErrorCode = 7006 // dms服务错误
)

type CodeError struct {
	code ErrorCode
	err  error
}

func (e *CodeError) Error() string {
	return fmt.Sprintf("%v", e.err)
}

func (e *CodeError) Code() int {
	if e.err == nil {
		return int(StatusOK)
	}
	return int(e.code)
}

func NewCodeError(err error, code ErrorCode) *CodeError {
	return &CodeError{
		code: code,
		err:  err,
	}
}
