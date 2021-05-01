package intmap

import "testing"

func TestSetSimple(t *testing.T) {
	m := NewSet()
	var i int64
	var ok bool

	// --------------------------------------------------------------------
	// Put() and Get()

	for i = 0; i < 20000; i += 2 {
		m.Add(i)
	}
	for i = 0; i < 20000; i += 2 {
		if ok = m.Has(i); !ok {
			t.Errorf("didn't get expected value")
		}
		if ok = m.Has(i + 1); ok {
			t.Errorf("didn't get expected 'not found' flag")
		}
	}

	if m.Size() != int(20000/2) {
		t.Errorf("size (%d) is not right, should be %d", m.Size(), int(20000/2))
	}

	// --------------------------------------------------------------------
	// Slice()

	m0 := make(map[int64]int64, 1000)
	for i = 0; i < 20000; i += 2 {
		m0[i] = i
	}
	n := len(m0)

	for _, k := range m.Slice() {
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
	// Del()

	for i = 0; i < 10000; i += 2 {
		m.Delete(i)
	}
	for i = 0; i < 10000; i += 2 {
		if ok = m.Has(i); ok {
			t.Errorf("didn't get expected 'not found' flag")
		}
		if ok = m.Has(i + 1); ok {
			t.Errorf("didn't get expected 'not found' flag")
		}
	}
	for i = 10000; i < 20000; i += 2 {
		if ok = m.Has(i); !ok {
			t.Errorf("didn't get expected value")
		}
		if ok = m.Has(i + 1); ok {
			t.Errorf("didn't get expected 'not found' flag")
		}
	}

	for i = 10000; i < 20000; i += 2 {
		m.Delete(i)
	}
	for i = 10000; i < 20000; i += 2 {
		if ok = m.Has(i); ok {
			t.Errorf("didn't get expected 'not found' flag")
		}
	}

	// --------------------------------------------------------------------
	// Add() and Get()

	for i = 0; i < 20000; i += 2 {
		m.Add(i)
	}
	for i = 0; i < 20000; i += 2 {
		if ok = m.Has(i); !ok {
			t.Errorf("didn't get expected value")
		}
		if ok = m.Has(i + 1); ok {
			t.Errorf("didn't get expected 'not found' flag")
		}
	}
}
