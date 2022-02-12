package bbp

import (
	"fmt"
	"testing"
)

func TestExampleGet(t *testing.T) {
	buf := Get(0, 50)
	buf.WriteString("first line\n")
	buf.Write([]byte("second line\n"))

	fmt.Printf("buffer.B = %q\n", buf.Bytes())

	// It is safe to release byte buffer now, since it is
	// no longer used.
	Put(buf)
}

func TestExampleGrow(t *testing.T) {
	buf := []byte("first line\n")
	buf = Grow(buf, 50)
	buf = append(buf, "second line\n"...)

	fmt.Printf("buffer.B = %q\n", buf)

	// It is safe to release byte buffer now, since it is
	// no longer used.
	PutSlice(buf)
}

func TestExamplePool(t *testing.T) {
	var pool Pool
	buf := pool.Get()
	buf.WriteString("first line\n")
	buf.Write([]byte("second line\n"))

	fmt.Printf("buffer.B = %q\n", buf.Bytes())

	// It is safe to release byte buffer now, since it is
	// no longer used.
	Put(buf)
}

func TestExampleBuffer(t *testing.T) {
	var buf Buffer
	buf.WriteString("first line\n")
	buf.Write([]byte("second line\n"))
	buf.buf = append(buf.buf, "third line\n"...)

	fmt.Printf("buffer.B = %q\n", buf.buf)

	// It is safe to release byte buffer now, since it is
	// no longer used.
	Put(&buf)
}
