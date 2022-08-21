package bbp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testObject struct {
	A string
	B int64
	C *testObject
	D []byte
}

func TestObjectPool(t *testing.T) {
	var pool = NewObjectPool[testObject]()

	obj := pool.Get()
	assert.NotNil(t, obj)
	assert.Zero(t, *obj)

	obj.A = "hello there"
	obj.B = 1234
	obj.D = []byte("D")
	obj.C = pool.Get()
	obj.C.A = "inner object"
	obj.C.B = 4567
	obj.C.D = []byte("inner D")

	pool.Put(obj.C)
	pool.Put(obj)
	obj = pool.Get()
	assert.NotNil(t, obj)
	assert.Zero(t, *obj)
}
