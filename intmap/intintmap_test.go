package intmap

import (
	"math/rand"
	"testing"
)

func TestMapSimple(t *testing.T) {
	m := New(10, 0.99)
	var i int64
	var v int64
	var ok bool

	// --------------------------------------------------------------------
	// Put() and Get()

	for i = 0; i < 20000; i += 2 {
		m.Set(i, i)
	}
	for i = 0; i < 20000; i += 2 {
		if v, ok = m.Get(i); !ok || v != i {
			t.Errorf("didn't get expected value")
		}
		if _, ok = m.Get(i + 1); ok {
			t.Errorf("didn't get expected 'not found' flag")
		}
	}

	if m.Size() != int(20000/2) {
		t.Errorf("size (%d) is not right, should be %d", m.Size(), int(20000/2))
	}

	// --------------------------------------------------------------------
	// Keys()

	m0 := make(map[int64]int64, 1000)
	for i = 0; i < 20000; i += 2 {
		m0[i] = i
	}
	n := len(m0)

	for _, k := range m.Keys() {
		m0[k] = -k
	}
	if n != len(m0) {
		t.Errorf("get unexpected more keys")
	}

	for k, v := range m0 {
		if k != -v {
			t.Errorf("didn't get expected changed value")
		}
	}

	// --------------------------------------------------------------------
	// Items()

	m0 = make(map[int64]int64, 1000)
	for i = 0; i < 20000; i += 2 {
		m0[i] = i
	}
	n = len(m0)

	for _, kv := range m.Items() {
		m0[kv.K] = -kv.V
		if kv.K != kv.V {
			t.Errorf("didn't get expected key-value pair")
		}
	}
	if n != len(m0) {
		t.Errorf("get unexpected more keys")
	}

	for k, v := range m0 {
		if k != -v {
			t.Errorf("didn't get expected changed value")
		}
	}

	// --------------------------------------------------------------------
	// Del()

	for i = 0; i < 10000; i += 2 {
		m.Delete(i)
	}
	for i = 0; i < 10000; i += 2 {
		if _, ok = m.Get(i); ok {
			t.Errorf("didn't get expected 'not found' flag")
		}
		if _, ok = m.Get(i + 1); ok {
			t.Errorf("didn't get expected 'not found' flag")
		}
	}
	for i = 10000; i < 20000; i += 2 {
		if v, ok = m.Get(i); !ok || v != i {
			t.Errorf("didn't get expected value")
		}
		if _, ok = m.Get(i + 1); ok {
			t.Errorf("didn't get expected 'not found' flag")
		}
	}

	for i = 10000; i < 20000; i += 2 {
		m.Delete(i)
	}
	for i = 10000; i < 20000; i += 2 {
		if _, ok = m.Get(i); ok {
			t.Errorf("didn't get expected 'not found' flag")
		}
	}

	// --------------------------------------------------------------------
	// Put() and Get()

	for i = 0; i < 20000; i += 2 {
		m.Set(i, i*2)
	}
	for i = 0; i < 20000; i += 2 {
		if v, ok = m.Get(i); !ok || v != i*2 {
			t.Errorf("didn't get expected value")
		}
		if _, ok = m.Get(i + 1); ok {
			t.Errorf("didn't get expected 'not found' flag")
		}
	}

}

func TestDeleteFreeKey(t *testing.T) {
	m := New(10, 0.6)

	m.Delete(0)
	if m.Size() != 0 {
		t.Errorf("size (%d) is not right, should be %d", m.Size(), 0)
	}

	m.Set(0, 1)
	if v, ok := m.Get(0); !ok || v != 1 {
		t.Errorf("didn't get exprected value")
	}

	m.Delete(0)
	if m.Size() != 0 {
		t.Errorf("size (%d) is not right, should be %d", m.Size(), 0)
	}
}

func TestMap(t *testing.T) {
	m := New(10, 0.6)
	var ok bool
	var v int64

	step := int64(61)

	var i int64
	m.Set(0, 12345)
	for i = 1; i < 100000000; i += step {
		m.Set(i, i+7)
		m.Set(-i, i-7)

		if v, ok = m.Get(i); !ok || v != i+7 {
			t.Errorf("expected %d as value for key %d, got %d", i+7, i, v)
		}
		if v, ok = m.Get(-i); !ok || v != i-7 {
			t.Errorf("expected %d as value for key %d, got %d", i-7, -i, v)
		}
	}
	for i = 1; i < 100000000; i += step {
		if v, ok = m.Get(i); !ok || v != i+7 {
			t.Errorf("expected %d as value for key %d, got %d", i+7, i, v)
		}
		if v, ok = m.Get(-i); !ok || v != i-7 {
			t.Errorf("expected %d as value for key %d, got %d", i-7, -i, v)
		}

		for j := i + 1; j < i+step; j++ {
			if v, ok = m.Get(j); ok {
				t.Errorf("expected 'not found' flag for %d, found %d", j, v)
			}
		}
	}

	if v, ok = m.Get(0); !ok || v != 12345 {
		t.Errorf("expected 12345 for key 0")
	}
}

const MAX = 999999999
const STEP = 9534

func fillIntIntMap(m *Map) {
	var j int64
	for j = 0; j < MAX; j += STEP {
		m.Set(j, -j)
		for k := j; k < j+16; k++ {
			m.Set(k, -k)
		}

	}
}

func fillStdMap(m map[int64]int64) {
	var j int64
	for j = 0; j < MAX; j += STEP {
		m[j] = -j
		for k := j; k < j+16; k++ {
			m[k] = -k
		}
	}
}

func BenchmarkIntIntMapFill(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := New(2048, 0.60)
		fillIntIntMap(m)
	}
}

func BenchmarkStdMapFill(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := make(map[int64]int64, 2048)
		fillStdMap(m)
	}
}

func BenchmarkIntIntMapGet10PercentHitRate(b *testing.B) {
	var j, k, v, sum int64
	var ok bool
	m := New(2048, 0.60)
	fillIntIntMap(m)
	for i := 0; i < b.N; i++ {
		sum = int64(0)
		for j = 0; j < MAX; j += STEP {
			for k = j; k < 10; k++ {
				if v, ok = m.Get(k); ok {
					sum += v
				}
			}
		}
		//log.Println("int int sum:", sum)
	}
}

func BenchmarkStdMapGet10PercentHitRate(b *testing.B) {
	var j, k, v, sum int64
	var ok bool
	m := make(map[int64]int64, 2048)
	fillStdMap(m)
	for i := 0; i < b.N; i++ {
		sum = int64(0)
		for j = 0; j < MAX; j += STEP {
			for k = j; k < 10; k++ {
				if v, ok = m[k]; ok {
					sum += v
				}
			}
		}
		//log.Println("map sum:", sum)
	}
}

func BenchmarkIntIntMapGet100PercentHitRate(b *testing.B) {
	var j, v, sum int64
	var ok bool
	m := New(2048, 0.60)
	fillIntIntMap(m)
	for i := 0; i < b.N; i++ {
		sum = int64(0)
		for j = 0; j < MAX; j += STEP {
			if v, ok = m.Get(j); ok {
				sum += v
			}
		}
		//log.Println("int int sum:", sum)
	}
}

func BenchmarkStdMapGet100PercentHitRate(b *testing.B) {
	var j, v, sum int64
	var ok bool
	m := make(map[int64]int64, 2048)
	fillStdMap(m)
	for i := 0; i < b.N; i++ {
		sum = int64(0)
		for j = 0; j < MAX; j += STEP {
			if v, ok = m[j]; ok {
				sum += v
			}
		}
		//log.Println("map sum:", sum)
	}
}

var randIntegers = func() []int64 {
	out := make([]int64, 1024)
	for i := range out {
		out[i] = rand.Int63()
	}
	return out
}()

func BenchmarkIntIntMapGet_Size_1024_FillFactor_60(b *testing.B) {
	m := New(1024, 0.6)
	for _, x := range randIntegers {
		m.Set(x, -x)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k := randIntegers[i&1023]
		_, _ = m.Get(k)
	}
}

func BenchmarkStdMapGet_Size_1024(b *testing.B) {
	m := make(map[int64]int64, 1024)
	for _, x := range randIntegers {
		m[x] = -x
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k := randIntegers[i&1023]
		_ = m[k]
	}
}
