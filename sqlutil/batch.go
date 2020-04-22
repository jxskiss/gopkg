package sqlutil

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jxskiss/gopkg/strutil"
	"reflect"
	"strings"
	"sync"
	"unsafe"
)

type InsertOptions struct {
	Context   context.Context
	TableName string

	Ignore         bool
	OnDuplicateKey string
	OnConflict     string
}

func (p *InsertOptions) apply(opts ...InsertOpt) *InsertOptions {
	for _, f := range opts {
		f(p)
	}
	return p
}

type InsertOpt func(*InsertOptions)

// WithContext makes the query executed with `ExecContext` if available.
func WithContext(ctx context.Context) InsertOpt {
	return func(opts *InsertOptions) {
		opts.Context = ctx
	}
}

// WithTable makes the generated query to use provided table name.
func WithTable(tableName string) InsertOpt {
	return func(opts *InsertOptions) {
		opts.TableName = tableName
	}
}

// WithIgnore adds the mysql "IGNORE" adverb to the the generated query.
func WithIgnore() InsertOpt {
	return func(opts *InsertOptions) {
		opts.Ignore = true
	}
}

// OnDuplicateKey appends the mysql "ON DUPLICATE KEY" clause to the generated query.
func OnDuplicateKey(clause string) InsertOpt {
	return func(opts *InsertOptions) {
		opts.OnDuplicateKey = clause
	}
}

// OnConflict appends the postgresql "ON CONFLICT" clause to the generated query.
func OnConflict(clause string) InsertOpt {
	return func(opts *InsertOptions) {
		opts.OnConflict = clause
	}
}

// Executor is the minimal interface for batch inserting requires.
// The interface is implemented by *sql.DB, *sql.Tx, *sqlx.DB, *sqlx.Tx.
type Executor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// ContextExecutor is an optional interface to support context execution.
// If `BatchInsert` function is called with `WithContext` option, and the
// provided Executor implements this interface, then the method
// `ExecContext` will be called instead of the method `Exec`.
type ContextExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// BatchInsert generates SQL and executes it on the provided Executor.
// The provided param `rows` must be a slice of struct or pointer to struct,
// and the slice must have at least one element, or it returns error.
func BatchInsert(conn Executor, rows interface{}, opts ...InsertOpt) (result sql.Result, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	options := new(InsertOptions).apply(opts...)
	query, args := makeBatchInsertSQL("BatchInsert", rows, options)
	if options.Context != nil {
		if ctxConn, ok := conn.(ContextExecutor); ok {
			result, err = ctxConn.ExecContext(options.Context, query, args...)
		} else {
			result, err = conn.Exec(query, args...)
		}
	} else {
		result, err = conn.Exec(query, args...)
	}
	return
}

// MakeBatchInsertSQL generates SQL and returns the arguments to execute on database connection.
// The provided param `rows` must be a slice of struct or pointer to struct,
// and the slice must have at least one element, or it panics.
//
// The returned query uses `?` as parameter placeholder, if you are using this function
// with database which don't use `?` as placeholder, you may check the `Rebind` function
// from package `github.com/jmoiron/sqlx` to replace placeholders.
func MakeBatchInsertSQL(rows interface{}, opts ...InsertOpt) (sql string, args []interface{}) {
	options := new(InsertOptions).apply(opts...)
	return makeBatchInsertSQL("MakeBatchInsertSQL", rows, options)
}

func makeBatchInsertSQL(where string, rows interface{}, opts *InsertOptions) (sql string, args []interface{}) {
	assertSliceOfStructAndLength(where, rows)

	typInfo := parseType(rows)
	if len(opts.TableName) == 0 {
		opts.TableName = typInfo.tableName
	}

	var buf strings.Builder

	// mysql: insert ignore
	if opts.Ignore {
		buf.WriteString("INSERT IGNORE INTO ")
	} else {
		buf.WriteString("INSERT INTO ")
	}
	buf.WriteString(opts.TableName)
	buf.WriteByte(' ')
	buf.WriteString(typInfo.colNames)
	buf.WriteString(" VALUES ")

	rowsVal := reflect.ValueOf(rows)
	length := rowsVal.Len()
	fieldNum := len(typInfo.fieldIndex)
	args = make([]interface{}, 0, length*fieldNum)
	for i := 0; i < length; i++ {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(typInfo.placeholders)
		elem := reflect.Indirect(rowsVal.Index(i))
		for _, j := range typInfo.fieldIndex {
			args = append(args, elem.Field(j).Interface())
		}
	}

	// mysql: on duplicate key clause
	if len(opts.OnDuplicateKey) > 0 {
		buf.WriteString(" ON DUPLICATE KEY ")
		buf.WriteString(opts.OnDuplicateKey)
	}

	// postgresql: on conflict clause
	if len(opts.OnConflict) > 0 {
		buf.WriteString(" ON CONFLICT ")
		buf.WriteString(opts.OnConflict)
	}

	sql = buf.String()
	return
}

var typeCache sync.Map

type typeInfo struct {
	tableName    string
	colNames     string
	placeholders string
	fieldIndex   []int
}

func parseType(rows interface{}) *typeInfo {
	typ := reflect.TypeOf(rows)
	cachedInfo, ok := typeCache.Load(typ)
	if ok {
		return cachedInfo.(*typeInfo)
	}

	elemTyp := indirectType(indirectType(typ).Elem())
	tableName := strutil.ToSnake(elemTyp.Name())
	fieldNum := elemTyp.NumField()
	colNames := make([]string, 0, fieldNum)
	fieldIndex := make([]int, 0)
	for i := 0; i < fieldNum; i++ {
		field := elemTyp.Field(i)
		col := ""

		// ignore unexported fields
		if len(field.PkgPath) != 0 {
			continue
		}

		// be compatible with sqlx column name tag
		dbTag := field.Tag.Get("db")
		if dbTag == "-" {
			continue
		}
		if parts := strings.Split(dbTag, ","); len(parts) > 0 {
			if x := strings.TrimSpace(parts[0]); x != "" {
				col = x
			}
		}

		// be compatible with gorm column name tag
		if col == "" {
			gormTag := field.Tag.Get("gorm")
			if gormTag == "-" {
				continue
			}
			tags := strings.Split(gormTag, ";")
			for _, value := range tags {
				kv := strings.Split(value, ":")
				if len(kv) >= 2 && kv[0] == "column" {
					if x := strings.TrimSpace(kv[1]); x != "" {
						col = x
					}
					break
				}
			}
		}

		// default
		if col == "" {
			col = strutil.ToSnake(field.Name)
		}

		colNames = append(colNames, col)
		fieldIndex = append(fieldIndex, i)
	}

	placeholders := strings.Repeat("?,", len(fieldIndex))
	placeholders = strings.TrimSuffix(placeholders, ",")
	info := &typeInfo{
		tableName:    tableName,
		colNames:     "(" + strings.Join(colNames, ",") + ")",
		placeholders: "(" + placeholders + ")",
		fieldIndex:   fieldIndex,
	}
	typeCache.Store(typ, info)
	return info
}

func indirectType(typ reflect.Type) reflect.Type {
	if typ.Kind() != reflect.Ptr {
		return typ
	}
	return typ.Elem()
}

func assertSliceOfStructAndLength(where string, rows interface{}) {
	sliceTyp := reflect.TypeOf(rows)
	if sliceTyp == nil || sliceTyp.Kind() != reflect.Slice {
		panic(where + ": param is nil or not a slice")
	}
	elemTyp := sliceTyp.Elem()
	elemIsPtr := elemTyp.Kind() == reflect.Ptr
	if !(elemTyp.Kind() == reflect.Struct ||
		(elemIsPtr && elemTyp.Elem().Kind() == reflect.Struct)) {
		panic(where + ": slice element is not struct of pointer to struct")
	}
	eface := *(*[2]unsafe.Pointer)(unsafe.Pointer(&rows))
	sh := (*reflect.SliceHeader)(eface[1])
	if sh.Len == 0 {
		panic(where + ": slice length is zero")
	}
}