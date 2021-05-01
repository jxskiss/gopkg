package intintmap

import (
	"sync"
	"testing"
	"unsafe"
)

func BenchmarkConcurrentStdMapGet_NoLock(b *testing.B) {
	m := make(map[uintptr]uintptr)
	typPtrs := fillMap(func(k, v uintptr) { m[k] = v })

	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for _, ptr := range typPtrs {
				_ = m[ptr]
			}
		}
	})
}

func BenchmarkConcurrentStdMapGet_RWMutex(b *testing.B) {
	var mu sync.RWMutex
	m := make(map[uintptr]uintptr)
	typPtrs := fillMap(func(k, v uintptr) { m[k] = v })

	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for _, ptr := range typPtrs {
				mu.RLock()
				_ = m[ptr]
				mu.RUnlock()
			}
		}
	})
}

func BenchmarkConcurrentSliceIndex(b *testing.B) {
	slice := make([]uintptr, 12)

	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < 12; i++ {
				_ = slice[i]
			}
		}
	})
}

func BenchmarkConcurrentSyncMapGet(b *testing.B) {
	m := sync.Map{}
	typPtrs := fillMap(func(k, v uintptr) { m.Store(k, v) })

	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for _, ptr := range typPtrs {
				got, _ := m.Load(ptr)
				_ = got.(uintptr)
			}
		}
	})
}

func BenchmarkConcurrentCOWMapGet(b *testing.B) {
	m := New(8, 0.6)
	typPtrs := fillMap(func(k, v uintptr) { m.Set(int64(k), int64(v)) })

	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for _, ptr := range typPtrs {
				m.Get(int64(ptr))
			}
		}
	})
}

func fillMap(setfunc func(k, v uintptr)) []uintptr {
	var values = []interface{}{
		TestType1{},
		TestType2{},
		TestType3{},
		TestType4{},
		TestType5{},
		TestType6{},
		TestType7{},
		TestType8{},
		TestType9{},
		TestType10{},
		TestType11{},
		TestType12{},
	}
	ptrs := make([]uintptr, 0, len(values))
	for _, val := range values {
		typPtr := (*(*[2]uintptr)(unsafe.Pointer(&val)))[1]
		setfunc(typPtr, typPtr)
		ptrs = append(ptrs, typPtr)
	}
	return ptrs
}
