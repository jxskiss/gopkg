package mselect

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestManySelect_NormalCase(t *testing.T) {
	msel := New()

	type testData struct {
		a [200]byte
		b string
		c int64
	}

	var mu sync.Mutex
	var result struct {
		got1 string
		got2 int64
		got3 testData
		got4 *testData
		got5 any
		got6 fmt.Stringer
		got7 any
	}

	ch1 := make(chan string)
	ch2 := make(chan int64)
	ch3 := make(chan testData)
	ch4 := make(chan *testData)
	ch5 := make(chan any)
	ch6 := make(chan fmt.Stringer)
	ch7 := make(chan *struct{})

	msel.Add(NewTask(ch1,
		func(v string, ok bool) {
			mu.Lock()
			defer mu.Unlock()
			result.got1 = v
		}, nil))
	msel.Add(NewTask(ch2,
		func(v int64, ok bool) {
			assert.True(t, ok)
			mu.Lock()
			defer mu.Unlock()
			result.got2 = v
		}, nil))
	msel.Add(NewTask(ch3, nil, func(v testData, ok bool) {
		assert.True(t, ok)
		mu.Lock()
		defer mu.Unlock()
		result.got3 = v
	}))
	msel.Add(NewTask(ch4, nil, func(v *testData, ok bool) {
		assert.True(t, ok)
		mu.Lock()
		defer mu.Unlock()
		result.got4 = v
	}))
	msel.Add(NewTask(ch5, nil, func(v any, ok bool) {
		assert.True(t, ok)
		mu.Lock()
		defer mu.Unlock()
		result.got5 = v
	}))
	msel.Add(NewTask(ch6, func(v fmt.Stringer, ok bool) {
		assert.True(t, ok)
		mu.Lock()
		defer mu.Unlock()
		result.got6 = v
	}, nil))
	msel.Add(NewTask(ch7, nil, func(v *struct{}, ok bool) {
		assert.False(t, ok)
		mu.Lock()
		defer mu.Unlock()
		result.got7 = v
	}))

	assert.Equal(t, 7, msel.Count())

	time.Sleep(100 * time.Millisecond)

	ch1 <- "ch1 value"
	ch2 <- int64(23456)
	ch3 <- testData{
		a: [200]byte{},
		b: "ch3 value b",
		c: 34567,
	}
	ch4 <- &testData{
		a: [200]byte{},
		b: "ch4 value b",
		c: 45678,
	}
	ch5 <- nil
	ch6 <- stringFunc(func() string { return "stringFunc" })
	close(ch7)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	copyResult := result
	mu.Unlock()

	assert.Equal(t, "ch1 value", copyResult.got1)
	assert.Equal(t, int64(23456), copyResult.got2)
	assert.Equal(t, testData{
		a: [200]byte{},
		b: "ch3 value b",
		c: 34567,
	}, copyResult.got3)
	assert.Equal(t, testData{
		a: [200]byte{},
		b: "ch4 value b",
		c: 45678,
	}, *copyResult.got4)
	assert.Equal(t, nil, copyResult.got5)
	assert.Equal(t, "stringFunc", copyResult.got6.String())
	assert.True(t, copyResult.got7 != nil)
	assert.True(t, copyResult.got7.(*struct{}) == nil)

	assert.Equal(t, 6, msel.Count())
}

type stringFunc func() string

func (f stringFunc) String() string {
	return f()
}

func TestManySelect_ManyChannels(t *testing.T) {
	N := 5000

	mu := sync.Mutex{}
	result := make([][]int, N)

	makeTask := func(i int) *Task {
		ch := make(chan int)
		task := NewTask(ch, func(v int, ok bool) {
			mu.Lock()
			defer mu.Unlock()
			result[i] = append(result[i], v)
		}, nil)

		go func() {
			time.Sleep(10 * time.Millisecond)
			ch <- i
			time.Sleep(10 * time.Millisecond)
			ch <- i + 1
		}()

		return task
	}

	msel := New()
	for i := 0; i < N; i++ {
		msel.Add(makeTask(i))
	}
	assert.Equal(t, N, msel.Count())

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	copyResult := make([][]int, N)
	copy(copyResult, result)
	mu.Unlock()

	for i := range copyResult {
		assert.Len(t, copyResult[i], 2)
		assert.Equal(t, i, copyResult[i][0])
		assert.Equal(t, i+1, copyResult[i][1])
	}
}

func TestManySelect_Stop(t *testing.T) {
	N := 5000

	mu := sync.Mutex{}
	result := make([][]int, N)

	makeTask := func(i int) *Task {
		ch := make(chan int)
		task := NewTask(ch, func(v int, ok bool) {
			mu.Lock()
			defer mu.Unlock()
			result[i] = append(result[i], v)
		}, nil)

		go func() {
			time.Sleep(10 * time.Millisecond)
			ch <- i
			time.Sleep(10 * time.Millisecond)
			ch <- i + 1
		}()

		return task
	}

	msel := New()
	for i := 0; i < N; i++ {
		msel.Add(makeTask(i))
	}
	assert.Equal(t, N, msel.Count())

	time.Sleep(11 * time.Millisecond)
	msel.Stop()

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 0, msel.Count())
}
