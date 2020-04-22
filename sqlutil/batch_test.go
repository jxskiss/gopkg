package sqlutil

import (
	"context"
	"database/sql"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

type TestObject struct {
	ID           *int64
	COLUMN_1     int64
	Column2      *string
	privateField int
	Column_3_ABC string
	Column_4     int32      `gorm:"column:bling4"`
	Column_5     *int32     `db:"bling5"`
	Column_6     time.Time  `gorm:"-"`
	Column_7     *time.Time ``
	Column_8     []byte     `db:"-"`
}

func TestMakeBatchInsertSQL(t *testing.T) {
	var rows1 []TestObject
	for i := 0; i < 50; i++ {
		rows1 = append(rows1, TestObject{})
	}
	got1, args1 := MakeBatchInsertSQL(rows1)
	_, _ = MakeBatchInsertSQL(rows1)
	want1 := "INSERT INTO test_object (id,column_1,column2,column_3_abc,bling4,bling5,column_7) VALUES "
	varCount := 50 * 7

	assert.Contains(t, got1, want1)
	assert.Len(t, args1, varCount)
	assert.Equal(t, varCount, countPlaceholder(got1))

	var rows2 []*TestObject
	for i := 0; i < 50; i++ {
		rows2 = append(rows2, &TestObject{})
	}
	got2, args2 := MakeBatchInsertSQL(rows2, WithContext(context.Background()), WithTable("dummy_table"), WithIgnore())
	_, _ = MakeBatchInsertSQL(rows2)
	want2 := "INSERT IGNORE INTO dummy_table (id,column_1,column2,column_3_abc,bling4,bling5,column_7) VALUES "

	assert.Contains(t, got2, want2)
	assert.Len(t, args2, varCount)
	assert.Equal(t, varCount, countPlaceholder(got2))
}

func TestMakeBatchInsertSQL_Panic(t *testing.T) {
	for _, test := range []interface{}{
		nil,
		TestObject{},
		&TestObject{},
		[]int64{1, 2, 3},
		[]TestObject{},
		[]*TestObject{},
	} {
		assert.Panics(t, func() { _, _ = MakeBatchInsertSQL(test) })
		_, err := BatchInsert(dummyExecutor{}, test)
		t.Log(err)
		assert.True(t, strings.HasPrefix(err.Error(), "BatchInsert: "))
	}
}

func countPlaceholder(query string) int {
	return len(strings.Split(query, "?")) - 1
}

type dummyExecutor struct{}

func (p dummyExecutor) Exec(query string, args ...interface{}) (sql.Result, error) {
	return dummyResult{}, nil
}

func (p dummyExecutor) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return dummyResult{}, nil
}

type dummyResult struct{}

func (d dummyResult) LastInsertId() (int64, error) { return 0, nil }

func (d dummyResult) RowsAffected() (int64, error) { return 0, nil }