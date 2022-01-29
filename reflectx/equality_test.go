package reflectx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testtype1 struct {
	A string
	B int64
}

type testtype2 struct {
	A string
	B int64
}

type testtype3 struct {
	A string
	B int8
}

type recurtype1 struct {
	A    string
	b    int32
	self *recurtype1
}

type recurtype2 struct {
	A    string
	b    int32
	self *recurtype2
}

type recurtype3 struct {
	A *testtype1
	B *recurtype3
}

type recurtype4 struct {
	A *testtype2
	B *recurtype3
}

func TestIsIdenticalType(t *testing.T) {
	equal, msg := IsIdenticalType(&testtype1{}, &testtype2{})
	t.Log(msg)
	assert.True(t, equal)

	equal, msg = IsIdenticalType(&testtype1{}, &testtype3{})
	t.Log(msg)
	assert.False(t, equal)

	equal, msg = IsIdenticalType(&testtype2{}, &testtype3{})
	t.Log(msg)
	assert.False(t, equal)
}

func TestIsIdenticalType_Recursive(t *testing.T) {
	equal, msg := IsIdenticalType(&recurtype1{}, &recurtype2{})
	t.Log(msg)
	assert.True(t, equal)

	equal, msg = IsIdenticalType(&recurtype3{}, &recurtype4{})
	t.Log(msg)
	assert.True(t, equal)
}

type thriftType1 struct {
	EntityID *int64  `thrift:"Entity,1"`
	HtMLLink *string `thrift:"HtmlLink,2"`
}

type thriftType2 struct {
	EntityId *int64  `thrift:"Entity,1"`
	HtmlLink *string `thrift:"HtmlLink,2"`
}

func TestIsIdenticalThriftType(t *testing.T) {
	equal, msg := IsIdenticalThriftType(&thriftType1{}, &thriftType2{})
	t.Log(msg)
	assert.True(t, equal)
}
