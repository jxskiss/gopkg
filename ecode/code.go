package ecode

import (
	"encoding/json"
	"fmt"
)

type Code struct {
	code    int32
	msg     string
	details []interface{}
	reg     *Registry
}

type ErrCode interface {
	Error() string
	Code() int32
	Message() string
	Details() []interface{}
}

func (e *Code) String() string { return e.Error() }

func (e *Code) Error() string {
	code := e.Code()
	msg := e.Message()
	if msg == "" {
		msg = "<no message>"
	}
	return fmt.Sprintf("[%d] %s", code, msg)
}

func (e *Code) Code() int32 { return e.code }

func (e *Code) Message() string {
	if e.msg != "" {
		return e.msg
	}
	return e.reg.getMessage(e.code)
}

func (e *Code) Details() []interface{} { return e.details }

func (e *Code) WithDetails(details ...interface{}) (code *Code) {
	return &Code{
		code:    e.code,
		msg:     e.msg,
		details: details,
		reg:     e.reg,
	}
}

func (e *Code) WithMessage(msg string, details ...interface{}) (code *Code) {
	return &Code{
		code:    e.code,
		msg:     msg,
		details: details,
		reg:     e.reg,
	}
}

func (e *Code) MarshalJSON() ([]byte, error) {
	out := struct {
		Code    int32         `json:"code"`
		Message string        `json:"message,omitempty"`
		Details []interface{} `json:"details,omitempty"`
	}{
		Code:    e.Code(),
		Message: e.Message(),
		Details: e.details,
	}
	return json.Marshal(out)
}

func unwrapErrCode(err error) ErrCode {
	if err == nil {
		return nil
	}

	type causer interface {
		Cause() error
	}
	type wrapper interface {
		Unwrap() error
	}

	if wrappedErr, ok := err.(causer); ok {
		err = wrappedErr.Cause()
	} else if wrappedErr, ok := err.(wrapper); ok {
		err = wrappedErr.Unwrap()
	}
	if code, ok := err.(ErrCode); ok && code != nil {
		return code
	}
	return nil
}

func Is(err error, target ErrCode) bool {
	errCode := unwrapErrCode(err)
	if errCode != nil && errCode.Code() == target.Code() {
		return true
	}
	return false
}

func IsErrCode(err error) bool {
	if err == nil {
		return false
	}
	errCode := unwrapErrCode(err)
	return errCode != nil
}
