package errcode

import (
	"encoding/json"
	"fmt"
)

// ErrCode is the interface implemented by an error code.
type ErrCode interface {

	// Error returns the error message, it implements the error interface.
	Error() string

	// Code returns the integer error code.
	Code() int32

	// Message returns the registered message for the error code.
	// If message is not available, it returns an empty string "".
	Message() string

	// Details returns the additional detail information attached to
	// the error code.
	// It may return nil if no additional information is attached.
	Details() []interface{}
}

// Code represents an error code. It can be created by calling
// Registry.Register or Registry.RegisterReserved.
// Code implements the ErrCode interface.
type Code struct {
	code    int32
	msg     string
	details []interface{}
	reg     *Registry
}

func (e *Code) String() string { return e.Error() }

// Error returns the error message, it implements the error interface.
// If message is not registered for the error code, it uses
// "(no message)" as a default message.
func (e *Code) Error() string {
	code := e.Code()
	msg := e.Message()
	if msg == "" {
		msg = "(no message)"
	}
	return fmt.Sprintf("[%d] %s", code, msg)
}

// Code returns the integer error code.
func (e *Code) Code() int32 { return e.code }

// Message returns the error message associated with the error code.
// If message is not available, it returns an empty string "".
func (e *Code) Message() string {
	if e.msg != "" {
		return e.msg
	}
	return e.reg.getMessage(e.code)
}

// Details returns the additional detail information attached to
// the error code.
// It may return nil if no additional information is attached.
func (e *Code) Details() []interface{} { return e.details }

// WithDetails returns a copy of Code with new additional information attached.
func (e *Code) WithDetails(details ...interface{}) (code *Code) {
	details = append(e.details, details...)
	return &Code{
		code:    e.code,
		msg:     e.msg,
		details: details,
		reg:     e.reg,
	}
}

// WithMessage returns a copy of Code with the given message.
// If details is given, the new additional information will be attached
// to the returned Code.
func (e *Code) WithMessage(msg string, details ...interface{}) (code *Code) {
	details = append(e.details, details...)
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

// Is reports whether any error in err's chain matches the target ErrCode.
func Is(err error, target ErrCode) bool {
	errCode := unwrapErrCode(err)
	return errCode != nil && errCode.Code() == target.Code()
}

// IsErrCode reports whether any error in err's chain is an ErrCode.
func IsErrCode(err error) bool {
	errCode := unwrapErrCode(err)
	return errCode != nil
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
