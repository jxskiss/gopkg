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

type ThriftType1 struct {
	EntityID *int64  `thrift:"Entity,1"`
	HtMLLink *string `thrift:"HtmlLink,2"`
}

type ThriftType2 struct {
	EntityId *int64  `thrift:"Entity,1"`
	HtmlLink *string `thrift:"HtmlLink,2"`
}

type ThriftType3 struct {
	String_         string                  `frugal:"1,default,string" json:"String"`
	ListSimple      []*ThriftType1          `frugal:"2,default,list<Simple>" json:"ListSimple"`
	Double          float64                 `frugal:"3,default,double" json:"Double"`
	Byte            int8                    `frugal:"14,default,byte" json:"Byte"`
	MapStringSimple map[string]*ThriftType1 `frugal:"15,default,map<string:Simple>" json:"MapStringSimple"`
}

type ThriftType4 struct {
	String4          string                  `frugal:"1,default,string" json:"String"`
	ListSimple4      []*ThriftType2          `frugal:"2,default,list<Simple>" json:"ListSimple"`
	Double4          float64                 `frugal:"3,default,double" json:"Double"`
	Byte4            int8                    `frugal:"14,default,byte" json:"Byte"`
	MapStringSimple4 map[string]*ThriftType2 `frugal:"15,default,map<string:Simple>" json:"MapStringSimple"`
}

func TestIsIdenticalThriftType(t *testing.T) {
	equal, msg := IsIdenticalThriftType(&ThriftType1{}, &ThriftType2{})
	t.Log(msg)
	assert.True(t, equal)

	equal, msg = IsIdenticalThriftType(&ThriftType3{}, &ThriftType4{})
	t.Log(msg)
	assert.True(t, equal)
}
