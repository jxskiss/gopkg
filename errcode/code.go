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

	// Details returns the error details attached to the Code.
	// It returns nil if no details are attached.
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

// Details returns the error details attached to the Code.
// It returns nil if no details are attached.
func (e *Code) Details() []interface{} { return e.details }

func (e *Code) clone() *Code {
	detailsLen := len(e.details)
	return &Code{
		code:    e.code,
		msg:     e.msg,
		details: e.details[:detailsLen:detailsLen],
		reg:     e.reg,
	}
}

// WithMessage returns a copy of Code with the given message.
func (e *Code) WithMessage(msg string) (code *Code) {
	code = e.clone()
	code.msg = msg
	return
}

// AddDetails returns a copy of Code with new error details attached
// to the returned Code.
func (e *Code) AddDetails(details ...interface{}) (code *Code) {
	code = e.clone()
	code.details = append(code.details, details...)
	return
}

// RemoveDetails returns a copy of Code without the error details.
// If the Code does not have error details, it returns the Code
// directly instead of a copy.
// When returning an error code to end-users, you may want to remove
// the error details which generally should not be exposed to them.
func (e *Code) RemoveDetails() (code *Code) {
	if len(e.details) == 0 {
		return e
	}
	return &Code{code: e.code, msg: e.msg, reg: e.reg}
}

type jsonCode struct {
	Code    int32         `json:"code"`
	Message string        `json:"message,omitempty"`
	Details []interface{} `json:"details,omitempty"`
}

// MarshalJSON implements json.Marshaler.
func (e *Code) MarshalJSON() ([]byte, error) {
	out := &jsonCode{
		Code:    e.Code(),
		Message: e.Message(),
		Details: e.details,
	}
	return json.Marshal(out)
}

// UnmarshalJSON implements json.Unmarshaler.
func (e *Code) UnmarshalJSON(data []byte) error {
	tmp := &jsonCode{}
	err := json.Unmarshal(data, tmp)
	if err != nil {
		return err
	}
	e.code = tmp.Code
	e.msg = tmp.Message
	e.details = tmp.Details
	return nil
}

// Is reports whether any error in err's chain matches the target ErrCode.
func Is(err error, target ErrCode) bool {
	errCode := unwrapErrCode(err)
	return errCode != nil && errCode.Code() == target.Code()
}

// IsErrCode reports whether any error in err's chain is an ErrCode.
func IsErrCode(err error) bool {
	if errCode := unwrapErrCode(err); errCode != nil {
		return true
	}
	return false
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

	for _err := err; _err != nil; {
		if code, ok := _err.(ErrCode); ok && code != nil {
			return code
		}
		wrapped, ok := _err.(causer)
		if !ok {
			break
		}
		_err = wrapped.Cause()
	}
	for _err := err; err != nil; {
		if code, ok := _err.(ErrCode); ok && code != nil {
			return code
		}
		wrapped, ok := _err.(wrapper)
		if !ok {
			break
		}
		_err = wrapped.Unwrap()
	}
	return nil
}
