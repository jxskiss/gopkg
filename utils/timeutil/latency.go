package timeutil

import (
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/jxskiss/gopkg/v2/perf/bbp"
)

var latencyBufPool = bbp.NewPool(bbp.Recorder{
	DefaultSize: 128,
})

type latencyFrame struct {
	name    string
	latency time.Duration
}

// LatencyRecorder helps to record operation latencies.
type LatencyRecorder struct {
	startTime time.Time

	mu       sync.RWMutex
	frames   []latencyFrame
	fMaxTime time.Time
	slots    []latencyFrame
}

// NewLatencyRecorder creates a new LatencyRecorder.
func NewLatencyRecorder() *LatencyRecorder {
	buf := make([]latencyFrame, 16)
	recorder := &LatencyRecorder{
		startTime: time.Now(),
		frames:    buf[0:0:10],
		slots:     buf[10:10:16],
	}
	recorder.fMaxTime = recorder.startTime
	return recorder
}

// Reset resets the recorder to initial state, which is ready to be reused.
func (p *LatencyRecorder) Reset() {
	p.startTime = time.Now()
	p.frames = p.frames[:0]
	p.fMaxTime = p.startTime
	p.slots = p.slots[:0]
}

// Mark records latency of an operation, from the previous operation,
// or from the startTime if this is the first calling.
// It is designed to record each operation latency one by one
// in synchronous operations.
func (p *LatencyRecorder) Mark(name string) {
	now := time.Now()
	p.mu.Lock()
	latency := now.Sub(p.fMaxTime)
	p.frames = append(p.frames, latencyFrame{name, latency})
	p.fMaxTime = now
	p.mu.Unlock()
}

// MarkFromStartTime records latency of an operation, from the startTime,
// it does not affect recording of calling Mark.
func (p *LatencyRecorder) MarkFromStartTime(name string) {
	latency := time.Since(p.startTime)
	p.mu.Lock()
	p.slots = append(p.slots, latencyFrame{name, latency})
	p.mu.Unlock()
}

// MarkWithStartTime records latency of an operation, from the given time,
// it does not affect recording of calling Mark.
func (p *LatencyRecorder) MarkWithStartTime(name string, start time.Time) {
	latency := time.Since(start)
	p.mu.Lock()
	p.slots = append(p.slots, latencyFrame{name, latency})
	p.mu.Unlock()
}

// GetLatencyMap returns a map of all recorded latencies.
func (p *LatencyRecorder) GetLatencyMap() map[string]time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()

	total := time.Since(p.startTime)
	n := len(p.frames) + len(p.slots) + 1
	out := make(map[string]time.Duration, n)
	for i := range p.frames {
		out[p.frames[i].name] = p.frames[i].latency
	}
	for i := range p.slots {
		out[p.slots[i].name] = p.slots[i].latency
	}
	out["total"] = total
	return out
}

// Format formats the recorded latencies into a string,
// which can be sent to log by a logger.
func (p *LatencyRecorder) Format() string {
	buf := p.formatToBuffer()
	out := buf.String()
	latencyBufPool.PutBuffer(buf)
	return out
}

// WriteTo formats and writes the recorded latencies to the given io.Writer.
// It implements the interface io.WriterTo,
func (p *LatencyRecorder) WriteTo(w io.Writer) (n int64, err error) {
	buf := p.formatToBuffer()
	x, err := w.Write(buf.Bytes())
	n = int64(x)
	latencyBufPool.PutBuffer(buf)
	return
}

//nolint:errcheck
func (p *LatencyRecorder) formatToBuffer() *bbp.Buffer {
	var tmp [8]byte
	buf := latencyBufPool.GetBuffer()
	totalMsec := time.Since(p.startTime).Milliseconds()

	p.mu.RLock()
	defer p.mu.RUnlock()

	for i := range p.frames {
		msec := p.frames[i].latency.Milliseconds()
		buf.WriteString(p.frames[i].name)
		buf.WriteByte('=')
		buf.Write(strconv.AppendInt(tmp[:0], msec, 10))
		buf.WriteByte(' ')
	}
	for i := range p.slots {
		msec := p.slots[i].latency.Milliseconds()
		buf.WriteString(p.slots[i].name)
		buf.WriteByte('=')
		buf.Write(strconv.AppendInt(tmp[:0], msec, 10))
		buf.WriteByte(' ')
	}
	buf.WriteString("total=")
	buf.Write(strconv.AppendInt(tmp[:0], totalMsec, 10))
	return buf
}
