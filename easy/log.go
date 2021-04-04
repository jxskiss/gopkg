package easy

import (
	"bytes"
	"context"
	"fmt"
	"github.com/json-iterator/go"
	"github.com/jxskiss/gopkg/json"
	"github.com/jxskiss/gopkg/strutil"
	"io"
	"log"
	"os"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"
)

func ConfigLog(cfg LogCfg) {
	_logcfg = cfg
}

var _logcfg LogCfg

type LogCfg struct {
	EnableDebug func() bool
	Logger      func() ErrDebugLogger
	CtxLogger   func(context.Context) ErrDebugLogger
}

func (p LogCfg) getLogger(ctxp *context.Context) ErrDebugLogger {
	if p.CtxLogger != nil && ctxp != nil {
		if lg := p.CtxLogger(*ctxp); lg != nil {
			return lg
		}
	}
	if p.Logger != nil {
		if lg := p.Logger(); lg != nil {
			return lg
		}
	}
	return stdLogger{}
}

type stdLogger struct{}

func (p stdLogger) Debugf(format string, args ...interface{}) { log.Printf("DEBUG: "+format, args...) }
func (p stdLogger) Errorf(format string, args ...interface{}) { log.Printf("ERROR: "+format, args...) }

// ErrLogger is an interface which log an message at ERROR level.
// It's implemented by *logrus.Logger, *logrus.Entry, *zap.SugaredLogger,
// and many other logging packages.
type ErrLogger interface {
	Errorf(format string, args ...interface{})
}

// DebugLogger is an interface which log an message at DEBUG level.
// It's implemented by *logrus.Logger, *logrus.Entry, *zap.SugaredLogger,
// and many other logging packages.
type DebugLogger interface {
	Debugf(format string, args ...interface{})
}

// ErrDebugLogger is an interface which log messages at ERROR and DEBUG level.
// It's implemented by *logrus.Logger, *logrus.Entry, *zap.SugaredLogger,
// and many other logging packages.
type ErrDebugLogger interface {
	ErrLogger
	DebugLogger
}

// PrintFunc is a function to print the given arguments in format to somewhere.
// It implements the interface `ErrDebugLogger`.
type PrintFunc func(format string, args ...interface{})

func (f PrintFunc) Errorf(format string, args ...interface{}) { f(format, args...) }

func (f PrintFunc) Debugf(format string, args ...interface{}) { f(format, args...) }

var logjson = jsoniter.Config{
	// compatible with standard library behavior
	SortMapKeys:            true,
	ValidateJsonRawMessage: true,

	// incompatible with standard library behavior
	EscapeHTML: false,
}.Froze()

// JSON converts given object to a json string, it never returns error.
func JSON(v interface{}) string {
	b, err := logjson.Marshal(v)
	if err != nil {
		return fmt.Sprintf("<error: %v>", err)
	}
	return String_(b)
}

// Logfmt converts given object to a string in logfmt format, it never
// returns error. Note that only struct and map of basic types are
// supported, non-basic types are simply ignored.
func Logfmt(v interface{}) string {
	if IsNil(v) {
		return "null"
	}
	var src Bytes
	switch v := v.(type) {
	case []byte:
		src = v
	case string:
		src = ToBytes_(v)
	}
	if src != nil && utf8.Valid(src) {
		srcstr := src.String_()
		if strings.IndexFunc(srcstr, needsQuoteValueRune) != -1 {
			return JSON(srcstr)
		}
		return srcstr
	}

	// simple values
	val := reflect.Indirect(reflect.ValueOf(v))
	if !val.IsValid() {
		return "null"
	}
	if isBasicType(val.Type()) {
		return fmt.Sprint(val)
	}
	if val.Kind() != reflect.Struct && val.Kind() != reflect.Map {
		return "<error: unsupported logfmt type>"
	}

	keyValues := make([]interface{}, 0)
	if val.Kind() == reflect.Map {
		keys := make([]string, 0, val.Len())
		values := make(map[string]interface{}, val.Len())
		for iter := val.MapRange(); iter.Next(); {
			k, v := iter.Key(), reflect.Indirect(iter.Value())
			if !isBasicType(k.Type()) || !v.IsValid() {
				continue
			}
			v = reflect.ValueOf(v.Interface())
			if !v.IsValid() {
				continue
			}
			kstr := fmt.Sprint(k.Interface())
			if isBasicType(v.Type()) {
				keys = append(keys, kstr)
				values[kstr] = v.Interface()
				continue
			}
			if bv, ok := v.Interface().([]byte); ok {
				if len(bv) > 0 && utf8.Valid(bv) {
					keys = append(keys, kstr)
					values[kstr] = String_(bv)
				}
				continue
			}
			if v.Kind() == reflect.Slice && isBasicType(v.Elem().Type()) {
				keys = append(keys, kstr)
				values[kstr] = JSON(v.Interface())
				continue
			}
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := values[k]
			keyValues = append(keyValues, k, v)
		}
	} else { // reflect.Struct
		typ := val.Type()
		fieldNum := val.NumField()
		for i := 0; i < fieldNum; i++ {
			field := typ.Field(i)
			// ignore unexported fields which we can't take interface
			if len(field.PkgPath) != 0 {
				continue
			}
			fk := strutil.ToSnakeCase(field.Name)
			fv := reflect.Indirect(val.Field(i))
			if !(fv.IsValid() && fv.CanInterface()) {
				continue
			}
			if isBasicType(fv.Type()) {
				keyValues = append(keyValues, fk, fv.Interface())
				continue
			}
			if bv, ok := fv.Interface().([]byte); ok {
				if len(bv) > 0 && utf8.Valid(bv) {
					keyValues = append(keyValues, fk, String_(bv))
				}
				continue
			}
			if fv.Kind() == reflect.Slice && isBasicType(fv.Elem().Type()) {
				keyValues = append(keyValues, fk, JSON(fv.Interface()))
				continue
			}
		}
	}
	if len(keyValues) == 0 {
		return ""
	}

	buf := strings.Builder{}
	needSpace := false
	for i := 0; i < len(keyValues); i += 2 {
		k, v := keyValues[i], keyValues[i+1]
		if needSpace {
			buf.WriteByte(' ')
		}
		buf.WriteString(fmt.Sprint(k))
		buf.WriteString("=")
		vstr, ok := v.(string)
		if !ok {
			vstr = fmt.Sprint(v)
		}
		if strings.IndexFunc(vstr, needsQuoteValueRune) != -1 {
			vstr = JSON(vstr)
		}
		buf.WriteString(vstr)
		needSpace = true
	}
	return buf.String()
}

func needsQuoteValueRune(r rune) bool {
	return r <= ' ' || r == '=' || r == '"' || r == utf8.RuneError
}

// Pretty converts given object to a pretty formatted json string.
// If the input is an json string, it will be formatted using json.Indent
// with four space characters as indent.
func Pretty(v interface{}) string {
	var src Bytes
	switch v := v.(type) {
	case []byte:
		src = v
	case string:
		src = ToBytes_(v)
	}
	if src != nil {
		if json.Valid(src) {
			buf := bytes.NewBuffer(nil)
			_ = json.Indent(buf, src, "", "    ")
			return String_(buf.Bytes())
		}
		if utf8.Valid(src) {
			return src.String_()
		}
		return "<pretty: non-printable bytes>"
	}
	buf, err := logjson.MarshalIndent(v, "", "    ")
	if err != nil {
		return fmt.Sprintf("<error: %v>", err)
	}
	return String_(buf)
}

// Caller returns function name, filename, and the line number of the caller.
// The argument skip is the number of stack frames to ascend, with 0
// identifying the caller of Caller.
func Caller(skip int) (name, file string, line int) {
	pc, file, line, _ := runtime.Caller(skip + 1)
	name = runtime.FuncForPC(pc).Name()
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '/' {
			name = name[i+1:]
			break
		}
	}
	pathSepCnt := 0
	for i := len(file) - 1; i >= 0; i-- {
		if file[i] == '/' {
			pathSepCnt++
			if pathSepCnt == 2 {
				file = file[i+1:]
				break
			}
		}
	}
	return
}

var (
	stdoutMu sync.Mutex
	stdlogMu sync.Mutex
)

// CopyStdout replaces os.Stdout with a file created by `os.Pipe()`, and
// copies the content written to os.Stdout.
// This is not safe and most likely problematic, it's mainly to help intercepting
// output in testing.
func CopyStdout(f func()) (Bytes, error) {
	stdoutMu.Lock()
	defer stdoutMu.Unlock()
	old := os.Stdout
	defer func() { os.Stdout = old }()

	r, w, err := os.Pipe()
	// just to make sure the error didn't happen
	// in case of unfortunate, we should still do the specified work
	if err != nil {
		f()
		return nil, err
	}

	// copy the output in a separate goroutine, so printing can't block indefinitely
	outCh := make(chan []byte)
	go func() {
		var buf bytes.Buffer
		multi := io.MultiWriter(&buf, old)
		io.Copy(multi, r)
		outCh <- buf.Bytes()
	}()

	// do the work, write the stdout to pipe
	os.Stdout = w
	f()
	w.Close()

	out := <-outCh
	return out, nil
}

// CopyStdLog replaces the out Writer of the default logger of `log` package,
// and copies the content written to it.
// This is unsafe and most likely problematic, it's mainly to help intercepting
// log output in testing.
//
// Also NOTE if the out Writer of the default logger has already been replaced
// with another writer, we won't know anything about that writer and will
// restore the out Writer to os.Stderr before it returns.
// It will be a real mess.
func CopyStdLog(f func()) Bytes {
	stdlogMu.Lock()
	defer stdlogMu.Unlock()
	defer log.SetOutput(os.Stderr)

	var buf bytes.Buffer
	multi := io.MultiWriter(&buf, os.Stderr)
	log.SetOutput(multi)
	f()
	return buf.Bytes()
}

func formatArgs(stringer stringer, args []interface{}) []interface{} {
	retArgs := make([]interface{}, 0, len(args))
	for _, v := range args {
		x := v
		if v != nil {
			typ := reflect.TypeOf(v)
			for typ.Kind() == reflect.Ptr && isBasicType(typ.Elem()) {
				typ = typ.Elem()
				v = reflect.ValueOf(v).Elem().Interface()
			}
			if isBasicType(typ) {
				x = v
			} else if bv, ok := v.([]byte); ok && utf8.Valid(bv) {
				x = string(bv)
			} else {
				x = stringer(v)
			}
		}
		retArgs = append(retArgs, x)
	}
	return retArgs
}

func isBasicType(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Bool, reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return true
	}
	return false
}
