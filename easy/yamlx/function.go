package yamlx

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	goast "go/ast"
	goparser "go/parser"
	gotoken "go/token"
	"io"
	"reflect"
	"strconv"
	"time"

	"github.com/jxskiss/gopkg/v2/internal/unsafeheader"
	"github.com/jxskiss/gopkg/v2/perf/fastrand"
	"github.com/jxskiss/gopkg/v2/utils/strutil"
)

func (p *parser) addFuncs(funcMap map[string]any) {
	p.funcValMap = make(map[string]reflect.Value, len(builtinFuncs)+len(funcMap))
	for name, fn := range builtinFuncs {
		p.funcValMap[name] = reflect.ValueOf(fn)
	}
	for name, fn := range funcMap {
		p.funcValMap[name] = reflect.ValueOf(fn)
	}
}

func (p *parser) callFunction(str string) (ret any, err error) {
	var funcName string
	var fn reflect.Value
	var callArgs []reflect.Value

	if fn = p.funcValMap[str]; fn.IsValid() {
		funcName = str
	} else {
		var expr *expression
		expr, err = parseExpression(str)
		if err != nil {
			return nil, fmt.Errorf("cannot parse function expression: %w", err)
		}
		fn = p.funcValMap[expr.Func]
		if !fn.IsValid() {
			return nil, fmt.Errorf("function %s is unknown", expr.Func)
		}
		fnTyp := fn.Type()
		if len(expr.Args) != fnTyp.NumIn() {
			return nil, fmt.Errorf("function %s arguments count not match", expr.Func)
		}
		funcName = expr.Func
		for i := 0; i < len(expr.Args); i++ {
			fnArgTyp := fnTyp.In(i)
			if !expr.Args[i].Type().ConvertibleTo(fnArgTyp) {
				return nil, fmt.Errorf("function %s argument type not match: %v", expr.Func, expr.Args[i].Interface())
			}
			callArgs = append(callArgs, expr.Args[i].Convert(fnArgTyp))
		}
	}

	out := fn.Call(callArgs)
	if len(out) > 1 && !out[1].IsNil() {
		return nil, fmt.Errorf("error calling function %s: %w", funcName, out[1].Interface().(error))
	}

	result := out[0]
	return result.Interface(), nil
}

var (
	errNotCallExpression        = errors.New("not a call expression")
	errArgumentTypeNotSupported = errors.New("argument type not supported")
)

type expression struct {
	Func string
	Args []reflect.Value
}

func parseExpression(str string) (*expression, error) {
	expr, err := goparser.ParseExpr(str)
	if err != nil {
		return nil, err
	}
	call, ok := expr.(*goast.CallExpr)
	if !ok {
		return nil, errNotCallExpression
	}
	fnName := call.Fun.(*goast.Ident).String()
	args := make([]reflect.Value, 0, len(call.Args))
	for _, a := range call.Args {
		lit, ok := a.(*goast.BasicLit)
		if !ok {
			return nil, errArgumentTypeNotSupported
		}
		var aVal any
		switch lit.Kind {
		case gotoken.INT:
			aVal, _ = strconv.ParseInt(lit.Value, 10, 64)
		case gotoken.FLOAT:
			aVal, _ = strconv.ParseFloat(lit.Value, 64)
		case gotoken.STRING:
			aVal = lit.Value[1 : len(lit.Value)-1]
		default:
			return nil, errArgumentTypeNotSupported
		}
		args = append(args, reflect.ValueOf(aVal))
	}
	return &expression{
		Func: fnName,
		Args: args,
	}, nil
}

// -------- builtins -------- //

var builtinFuncs = map[string]any{
	"nowUnix":    builtinNowUnix,
	"nowMilli":   builtinNowMilli,
	"nowNano":    builtinNowNano,
	"nowRFC3339": builtinNowRFC3339,
	"nowFormat":  builtinNowFormat,
	"uuid":       builtinUUID,
	"rand":       builtinRand,
	"randN":      builtinRandN,
	"randStr":    builtinRandStr,
}

func builtinNowUnix() int64 {
	return time.Now().Unix()
}

func builtinNowMilli() int64 {
	return time.Now().UnixNano() / 1e6
}

func builtinNowNano() int64 {
	return time.Now().UnixNano()
}

func builtinNowRFC3339() string {
	return time.Now().Format(time.RFC3339)
}

func builtinNowFormat(layout string) string {
	return time.Now().Format(layout)
}

func builtinUUID() string {
	uuid := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, uuid)
	if err != nil {
		panic(err)
	}
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10

	buf := make([]byte, 36)
	hex.Encode(buf[:8], uuid[:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], uuid[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], uuid[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], uuid[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], uuid[10:])
	return unsafeheader.BytesToString(buf)
}

func builtinRand() (x int64) {
	return fastrand.Int63()
}

func builtinRandN(n int64) (x int64) {
	return fastrand.Int63n(n)
}

func builtinRandStr(n int) string {
	return strutil.Random(strutil.AlphaDigits, n)
}
