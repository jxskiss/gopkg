package sqlutil

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/jxskiss/gopkg/v2/unsafe/reflectx"
	"github.com/jxskiss/gopkg/v2/utils/structtag"
	"github.com/jxskiss/gopkg/v2/utils/strutil"
)

// InsertOptions holds options to use with batch inserting operation.
type InsertOptions struct {
	Context   context.Context
	TableName string
	Quote     string
	OmitCols  []string

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

func (p *InsertOptions) quote(name string) string {
	if p.Quote == "" {
		return name
	}
	return p.Quote + name + p.Quote
}

// InsertOpt represents an inserting option to use with batch
// inserting operation.
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

// WithQuote quotes the table name and column names with the given string.
func WithQuote(quote string) InsertOpt {
	return func(opts *InsertOptions) {
		opts.Quote = quote
	}
}

// OmitColumns exclude given columns from the generated query.
func OmitColumns(cols ...string) InsertOpt {
	return func(opts *InsertOptions) {
		opts.OmitCols = cols
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
	Exec(query string, args ...any) (sql.Result, error)
}

// ContextExecutor is an optional interface to support context execution.
// If `BatchInsert` function is called with `WithContext` option, and the
// provided Executor implements this interface, then the method
// `ExecContext` will be called instead of the method `Exec`.
type ContextExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// BatchInsert generates SQL and executes it on the provided Executor.
// The provided param `rows` must be a slice of struct or pointer to struct,
// and the slice must have at least one element, or it returns error.
func BatchInsert(conn Executor, rows any, opts ...InsertOpt) (result sql.Result, err error) {
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
func MakeBatchInsertSQL(rows any, opts ...InsertOpt) (sql string, args []any) {
	options := new(InsertOptions).apply(opts...)
	return makeBatchInsertSQL("MakeBatchInsertSQL", rows, options)
}

func makeBatchInsertSQL(where string, rows any, opts *InsertOptions) (sql string, args []any) {
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

	// table name
	buf.WriteString(opts.quote(opts.TableName))

	// column names
	var omitFieldIndex []int
	buf.WriteString(" (")
	for i, col := range typInfo.colNames {
		if inSlice(opts.OmitCols, col) {
			omitFieldIndex = append(omitFieldIndex, typInfo.fieldIndex[i])
			continue
		}
		buf.WriteString(opts.quote(col))
		if i < len(typInfo.colNames)-1 {
			buf.WriteByte(',')
		}
	}
	buf.WriteByte(')')

	// value placeholders
	placeholders := typInfo.placeholders
	fieldIndex := typInfo.fieldIndex
	if len(omitFieldIndex) > 0 {
		fieldIndex = diffSlice(fieldIndex, omitFieldIndex)
		placeholders = makePlaceholders(len(fieldIndex))
	}
	buf.WriteString(" VALUES ")
	rowsVal := reflect.ValueOf(rows)
	length := rowsVal.Len()
	fieldNum := len(typInfo.fieldIndex)
	args = make([]any, 0, length*fieldNum)
	for i := 0; i < length; i++ {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(placeholders)
		elem := reflect.Indirect(rowsVal.Index(i))
		for _, j := range fieldIndex {
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
	return sql, args
}

var typeCache sync.Map

type typeInfo struct {
	tableName    string
	colNames     []string
	placeholders string
	fieldIndex   []int
}

func parseType(rows any) *typeInfo {
	typ := reflect.TypeOf(rows)
	cachedInfo, ok := typeCache.Load(typ)
	if ok {
		return cachedInfo.(*typeInfo)
	}

	elemTyp := indirectType(indirectType(typ).Elem())
	tableName := strutil.ToSnakeCase(elemTyp.Name())
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
		opts := structtag.ParseOptions(dbTag, ",", "")
		if len(opts) > 0 {
			if opts[0].String() == "-" {
				continue
			}
			col = opts[0].String()
		}

		// be compatible with gorm column name tag
		if col == "" {
			gormTag := field.Tag.Get("gorm")
			opts = structtag.ParseOptions(gormTag, ";", ":")
			if len(opts) > 0 {
				if opts[0].Key() == "-" {
					continue
				}
				colopt, found := opts.Get("column")
				if found && colopt.Value() != "" {
					col = colopt.Value()
				}
			}
		}

		// default
		if col == "" {
			col = strutil.ToSnakeCase(field.Name)
		}

		colNames = append(colNames, col)
		fieldIndex = append(fieldIndex, i)
	}

	placeholders := makePlaceholders(len(fieldIndex))
	info := &typeInfo{
		tableName:    tableName,
		colNames:     colNames,
		placeholders: placeholders,
		fieldIndex:   fieldIndex,
	}
	typeCache.Store(typ, info)
	return info
}

func makePlaceholders(n int) string {
	marks := strings.Repeat("?,", n)
	marks = strings.TrimSuffix(marks, ",")
	return "(" + marks + ")"
}

func indirectType(typ reflect.Type) reflect.Type {
	if typ.Kind() != reflect.Ptr {
		return typ
	}
	return typ.Elem()
}

func inSlice[M ~[]E, E comparable](slice M, elem E) bool {
	for _, x := range slice {
		if x == elem {
			return true
		}
	}
	return false
}

func diffSlice[M ~[]E, E comparable](a, b M) M {
	out := make(M, 0, len(a))
	for _, x := range a {
		if !inSlice(b, x) {
			out = append(out, x)
		}
	}
	return out
}

func assertSliceOfStructAndLength(where string, rows any) {
	sliceTyp := reflect.TypeOf(rows)
	if sliceTyp == nil || sliceTyp.Kind() != reflect.Slice {
		panic(where + ": param is nil or not a slice")
	}
	elemTyp := sliceTyp.Elem()
	if indirectType(elemTyp).Kind() != reflect.Struct {
		panic(where + ": slice element is not struct or pointer to struct")
	}
	if reflectx.SliceLen(rows) == 0 {
		panic(where + ": slice length is zero")
	}
}
