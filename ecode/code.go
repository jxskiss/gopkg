package ecode

import (
	"encoding/json"
	"strconv"
)

type Code struct {
	code    int32
	reg     *Registry
	details []interface{}
}

type ErrCode interface {
	Error() string
	Code() int32
	Message() string
	Details() []interface{}
}

func (e Code) String() string {
	return e.Error()
}

func (e Code) Error() string {
	code := strconv.FormatInt(int64(e.code), 10)
	msg := e.Message()
	if msg == "" {
		return code
	}
	return code + ": " + msg
}

func (e Code) Code() int32 {
	return e.code
}

func (e Code) Message() string {
	if e.reg == nil {
		return ""
	}
	return e.reg.getMessage(e.code)
}

func (e Code) Details() []interface{} {
	return e.details
}

func (e Code) WithDetails(details ...interface{}) (code Code) {
	code = Code{
		code: e.code,
		reg:  e.reg,
	}
	if len(details) > 0 {
		code.details = details
	}
	return
}

func (e Code) Is(err error) bool {
	if code, ok := err.(ErrCode); ok {
		return e.Code() == code.Code()
	}
	return false
}

func (e Code) MarshalJSON() ([]byte, error) {
	out := struct {
		Code    int32  `json:"code"`
		Message string `json:"message,omitempty"`
	}{
		Code:    e.Code(),
		Message: e.Message(),
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
	if target == nil {
		return err == nil
	}
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
