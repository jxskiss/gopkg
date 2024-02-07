package bbp

import "fmt"

func ExampleGet() {
	buf := Get(0, 50)
	defer Put(buf)

	buf = append(buf, "first line\n"...)
	buf = append(buf, "second line\n"...)

	fmt.Println(string(buf))

	// Output:
	// first line
	// second line
}

func ExampleGrow() {
	buf := []byte("first line\n")
	buf = Grow(buf, 50, true)
	buf = append(buf, "second line\n"...)

	fmt.Println(string(buf))
	Put(buf)

	// Output:
	// first line
	// second line
}

func ExamplePool() {
	var pool Pool
	buf := pool.GetBuffer()
	defer PutBuffer(buf)

	buf.WriteString("first line\n")
	buf.Write([]byte("second line\n"))

	fmt.Println(buf.String())

	// Output:
	// first line
	// second line
}

func ExampleBuffer() {
	var buf Buffer
	defer PutBuffer(&buf)

	buf.WriteString("first line\n")
	buf.Write([]byte("second line\n"))
	buf.buf = append(buf.buf, "third line\n"...)

	fmt.Println(buf.String())

	// Output:
	// first line
	// second line
	// third line
}
