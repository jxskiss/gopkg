package errcode

import (
	"encoding/json"
	"fmt"
	"strings"
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
	// It may return nil if no details is attached.
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

func (e *Code) clone() *Code {
	detailsLen := len(e.details)
	return &Code{
		code:    e.code,
		msg:     e.msg,
		details: e.details[:detailsLen:detailsLen],
		reg:     e.reg,
	}
}

// Details returns the error details attached to the Code.
// It may return nil if no details is attached.
func (e *Code) Details() []interface{} { return e.details }

// WithDetails returns a copy of Code with new error details attached.
func (e *Code) WithDetails(details ...interface{}) (code *Code) {
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

// WithMessage returns a copy of Code with the given message.
// If error details are given, the new error details will be attached
// to the returned Code.
func (e *Code) WithMessage(msg string, details ...interface{}) (code *Code) {
	code = e.clone()
	code.msg = msg
	if len(details) > 0 {
		code.details = append(code.details, details...)
	}
	return
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
	if iv, ok := err.(interface {
		IsErrCode() bool
	}); ok {
		return iv.IsErrCode()
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

const encodePrefix = "errcode.Code:"

// Encode encodes a code to a string representation consists of a prefix
// `errcode.Code:` and the code's JSON representation.
//
// A string representation helps to pass error codes across service
// boundary, eg. through RPC metadata or http headers.
func Encode(code *Code) (string, error) {
	buf, err := code.MarshalJSON()
	if err != nil {
		return "", err
	}
	return encodePrefix + string(buf), nil
}

// Decode tries to decode a Code from it's string representation generated
// by Encode. If repr is not a valid code string, it returns (nil, false),
// else it returns the decoded Code and true.
func Decode(repr string) (code *Code, ok bool) {
	if !strings.HasPrefix(repr, encodePrefix) {
		return nil, false
	}
	code = &Code{}
	buf := []byte(repr[len(encodePrefix):])
	err := code.UnmarshalJSON(buf)
	if err != nil {
		return nil, false
	}
	return code, true
}
