package errcode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegistry(t *testing.T) {
	reg := New()
	got1 := reg.Register(-1024, "oops")
	got2 := reg.Register(0, "success")
	got3 := reg.Register(404, "not found")

	assert.Equal(t, "[-1024] oops", got1.Error())
	assert.Equal(t, "[0] success", got2.Error())
	assert.Equal(t, "[404] not found", got3.Error())

	list := reg.Dump()
	assert.Equal(t, int32(-1024), list[0].Code())
	assert.Equal(t, int32(0), list[1].Code())
	assert.Equal(t, int32(404), list[2].Code())

	reg.UpdateMessages(map[int32]string{
		400: "bad request",
		401: "unauthorized",
		403: "forbidden",
	})
	assert.Equal(t, "bad request", reg.getMessage(400))
	assert.Equal(t, "unauthorized", reg.getMessage(401))
	assert.Equal(t, "forbidden", reg.getMessage(403))

	reg.UpdateMessages(map[int32]string{
		-1: "error",
		0:  "success",
	})
	assert.Equal(t, "oops", got1.Message())
	assert.Equal(t, "success", got2.Message())
	assert.Equal(t, "not found", got3.Message())
}

func TestRegistryWithReserved(t *testing.T) {
	reg := New(WithReserved(func(code int32) bool { return code <= 99 }))

	assert.Panics(t, func() { reg.Register(-1, "") })
	assert.Panics(t, func() { reg.Register(99, "") })
	assert.NotPanics(t, func() { reg.Register(100, "") })
	assert.NotPanics(t, func() { reg.RegisterReserved(-1, "") })
	assert.NotPanics(t, func() { reg.RegisterReserved(99, "") })
}
