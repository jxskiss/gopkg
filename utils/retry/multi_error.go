package retry

import (
	"bytes"
	"fmt"
	"strings"
)

// NewSizedError returns an multiple error which holds at most size errors.
// The SizedError implementation is copied from github.com/jxskiss/errors
// to remove dependency of the package and for better compatibility for
// future go versions.
func NewSizedError(size int) *SizedError {
	return &SizedError{
		errs: make([]error, size),
		size: size,
	}
}

type SizedError struct {
	errs  []error
	size  int
	count int
}

func (E *SizedError) Append(errs ...error) {
	for _, err := range errs {
		if err != nil {
			E.errs[E.count%E.size] = err
			E.count++
		}
	}
}

func (E *SizedError) Error() string {
	if E == nil || E.count == 0 {
		return "<nil>"
	}
	var buf bytes.Buffer
	var first = true
	for _, err := range E.Errors() {
		if first {
			first = false
		} else {
			buf.Write(_singlelineSeparator)
		}
		buf.WriteString(err.Error())
	}
	return buf.String()
}

func (E *SizedError) ErrOrNil() error {
	if E == nil || E.count == 0 {
		return nil
	}
	return E
}

// Errors returns the errors as a slice in reversed order, if the underlying
// errors are more than size, only size errors will be returned, plus an
// additional error indicates the omitted error count.
func (E *SizedError) Errors() (errors []error) {
	if E.count == 0 {
		return nil
	}
	if E.count <= E.size {
		errors = make([]error, 0, E.count)
		for i := E.count - 1; i >= 0; i-- {
			errors = append(errors, E.errs[i])
		}
		return errors
	}
	errors = make([]error, 0, E.count+1)
	for i := E.count%E.size - 1; i >= 0; i-- {
		errors = append(errors, E.errs[i])
	}
	for i := E.size - 1; i >= E.count%E.size; i-- {
		errors = append(errors, E.errs[i])
	}
	errors = append(errors, fmt.Errorf("and %d more errors omitted", E.count-E.size))
	return errors
}

func (E *SizedError) Format(f fmt.State, c rune) {
	if c == 'v' && f.Flag('+') {
		f.Write(formatMultiLine(E.Errors()))
	} else {
		f.Write(formatSingleLine(E.Errors()))
	}
}

var (
	// Separator for single-line error messages.
	_singlelineSeparator = []byte("; ")

	// Prefix for multi-line messages
	_multilinePrefix = []byte("the following errors occurred:")

	// Prefix for the first and following lines of an item in a list of
	// multi-line error messages.
	//
	// For example, if a single item is:
	//
	// 	foo
	// 	bar
	//
	// It will become,
	//
	// 	 -  foo
	// 	    bar
	_multilineSeparator = []byte("\n -  ")
	_multilineIndent    = []byte("    ")
)

func formatSingleLine(errs []error) []byte {
	var buf bytes.Buffer
	var first = true
	for _, err := range errs {
		if first {
			first = false
		} else {
			buf.Write(_singlelineSeparator)
		}
		buf.WriteString(err.Error())
	}
	return buf.Bytes()
}

func formatMultiLine(errs []error) []byte {
	var buf bytes.Buffer
	buf.Write(_multilinePrefix)
	for _, err := range errs {
		buf.Write(_multilineSeparator)
		s := fmt.Sprintf("%+v", err)
		first := true
		for len(s) > 0 {
			if first {
				first = false
			} else {
				buf.Write(_multilineIndent)
			}
			idx := strings.IndexByte(s, '\n')
			if idx < 0 {
				idx = len(s) - 1
			}
			buf.WriteString(s[:idx+1])
			s = s[idx+1:]
		}
	}
	return buf.Bytes()
}
