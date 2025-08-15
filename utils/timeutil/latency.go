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
// The caller must ensure that the recorder is not used asynchronously
// after Reset, else data-race or panic happens.
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

// GetLatencyMap returns all recorded latencies.
// The returned value marks are the operation names in the order of calling Mark,
// and latency contains each operation's duration.
func (p *LatencyRecorder) GetLatencyMap() (marks []string, latency map[string]time.Duration) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	total := time.Since(p.startTime)
	frames := p.frames
	slots := p.slots

	n := len(frames) + len(slots) + 1
	marks = make([]string, 0, n)
	latency = make(map[string]time.Duration, n)
	for i := range frames {
		marks = append(marks, frames[i].name)
		latency[frames[i].name] = frames[i].latency
	}
	for i := range slots {
		marks = append(marks, slots[i].name)
		latency[slots[i].name] = slots[i].latency
	}
	marks = append(marks, "total")
	latency["total"] = total
	return marks, latency
}

// Format formats the recorded latencies into a string,
// which can be sent to log by a logger.
func (p *LatencyRecorder) Format() string {
	buf := latencyBufPool.GetBuffer()
	p.formatToBuffer(buf)
	out := buf.String()
	latencyBufPool.PutBuffer(buf)
	return out
}

// WriteTo formats and writes the recorded latencies to the given io.Writer.
// It implements the interface io.WriterTo,
func (p *LatencyRecorder) WriteTo(w io.Writer) (n int64, err error) {
	buf := latencyBufPool.GetBuffer()
	p.formatToBuffer(buf)
	x, err := w.Write(buf.Bytes())
	n = int64(x)
	latencyBufPool.PutBuffer(buf)
	return
}

//nolint:errcheck
func (p *LatencyRecorder) formatToBuffer(buf *bbp.Buffer) {
	totalMsec := time.Since(p.startTime).Milliseconds()

	p.mu.RLock()
	defer p.mu.RUnlock()

	frames := p.frames
	slots := p.slots
	tmp := [32]byte{} // enough for formatting milliseconds
	for i := range frames {
		msec := frames[i].latency.Milliseconds()
		buf.WriteString(frames[i].name)
		buf.WriteByte('=')
		buf.Write(strconv.AppendInt(tmp[:0], msec, 10))
		buf.WriteString("ms ")
	}
	for i := range slots {
		msec := slots[i].latency.Milliseconds()
		buf.WriteString(slots[i].name)
		buf.WriteByte('=')
		buf.Write(strconv.AppendInt(tmp[:0], msec, 10))
		buf.WriteString("ms ")
	}
	buf.WriteString("total=")
	buf.Write(strconv.AppendInt(tmp[:0], totalMsec, 10))
	buf.WriteString("ms")
}
