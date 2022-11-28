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

// Is reports whether an error is ErrCode and the code is same.
//
// This method allows Code to be tested using errors.Is.
func (e *Code) Is(target error) bool {
	if errCode, ok := target.(ErrCode); ok {
		return e.code == errCode.Code()
	}
	return false
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

	type causer interface {
		Cause() error
	}
	type wrapper interface {
		Unwrap() error
	}

	// We make sure that a poor implementation that causes a cycle
	// does not run forever.
	const unwrapLimit = 100

	for i := 0; err != nil && i < unwrapLimit; i++ {
		if code, _ := err.(ErrCode); code != nil {
			return code
		}
		switch e := err.(type) {
		case causer:
			err = e.Cause()
		case wrapper:
			err = e.Unwrap()
		default:
			return nil
		}
	}
	return nil
}
