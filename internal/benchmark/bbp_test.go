package benchmark

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/bytebufferpool"

	"github.com/jxskiss/gopkg/v2/perf/bbp"
)

var str = []string{
	"Lorem ipsum dolor sit amet, consectetur adipiscing elit",
	"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
	`Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris
		nisi ut aliquip ex ea commodo consequat.
		Duis aute irure dolor in reprehenderit in voluptate velit esse cillum
		dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident,
		sunt in culpa qui officia deserunt mollit anim id est laborum`,
	"Sed ut perspiciatis",
	"sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt",
	"Ut enim ad minima veniam, quis nostrum exercitationem ullam corporis suscipit",
	"laboriosam, nisi ut aliquid ex ea commodi consequatur",
	"Quis autem vel eum iure reprehenderit qui in ea voluptate velit esse quam nihil molestiae consequatur",
	"vel illum qui dolorem eum fugiat quo voluptas nulla pariatur",
}

var (
	stdBytesBufferPool = sync.Pool{New: func() any {
		return &bytes.Buffer{}
	}}
	bbpPool bbp.Pool
)

func Test_std_bytes_Buffer(t *testing.T) {
	buf := stdBytesBufferPool.Get().(*bytes.Buffer)
	workWith_std_bytes_Buffer(buf)
	assert.Equal(t, strings.Join(str, ""), buf.String())

	buf.Reset()
	stdBytesBufferPool.Put(buf)
}

func Test_valyala_bytebufferpool_ByteBuffer(t *testing.T) {
	buf := bytebufferpool.Get()
	workWith_valyala_bytebufferpool_ByteBuffer(buf)
	assert.Equal(t, strings.Join(str, ""), buf.String())

	bytebufferpool.Put(buf)
}

func Test_jxskiss_bbp_Buffer(t *testing.T) {
	buf := bbpPool.GetBuffer()
	workWith_jxskiss_bbp_Buffer(buf)
	assert.Equal(t, strings.Join(str, ""), buf.String())

	bbpPool.PutBuffer(buf)
}

func workWith_std_bytes_Buffer(b *bytes.Buffer) {
	for _, s := range str {
		b.WriteString(s)
	}
}

func workWith_valyala_bytebufferpool_ByteBuffer(b *bytebufferpool.ByteBuffer) {
	for _, s := range str {
		b.WriteString(s)
	}
}

func workWith_jxskiss_bbp_Buffer(b *bbp.Buffer) {
	for _, s := range str {
		b.WriteString(s)
	}
}

func test_std_bytes_Buffer() {
	buf := stdBytesBufferPool.Get().(*bytes.Buffer)
	workWith_std_bytes_Buffer(buf)
	buf.Reset()
	stdBytesBufferPool.Put(buf)
}

func test_valyala_bytebufferpool_ByteBuffer() {
	buf := bytebufferpool.Get()
	workWith_valyala_bytebufferpool_ByteBuffer(buf)
	bytebufferpool.Put(buf)
}

func test_jxskiss_bbp_Buffer() {
	buf := bbpPool.GetBuffer()
	workWith_jxskiss_bbp_Buffer(buf)
	bbpPool.PutBuffer(buf)
}

func Benchmark_std_BytesBufferPool(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			test_std_bytes_Buffer()
		}
	})
}

func Benchmark_valyala_ByteBufferPool(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			test_valyala_bytebufferpool_ByteBuffer()
		}
	})
}

func Benchmark_jxskiss_bbpPool_Buffer(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			test_jxskiss_bbp_Buffer()
		}
	})
}
