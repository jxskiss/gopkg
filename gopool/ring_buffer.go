package gopool

import (
	"runtime"
	"sync/atomic"
	"time"
)

// Code copied from github.com/Workiva/go-datastructures/queue/ring.go

// roundUp takes a uint64 greater than 0 and rounds it up to the next
// power of 2.
func roundUp(v uint64) uint64 {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v |= v >> 32
	v++
	return v
}

type node struct {
	position uint64
	data     interface{}
}

type nodes []*node

// RingBuffer is a MPMC buffer that achieves threadsafety with CAS operations
// only. A put on full or get on empty call will block until an item
// is put or retrieved. This buffer is similar to the buffer described here:
// http://www.1024cores.net/home/lock-free-algorithms/queues/bounded-mpmc-queue
// with some minor additions.
type RingBuffer struct {
	_padding0 [8]uint64
	queue     uint64
	_padding1 [8]uint64
	dequeue   uint64
	_padding2 [8]uint64
	mask      uint64
	_padding3 [8]uint64
	nodes     nodes
}

func (rb *RingBuffer) init(size uint64) {
	size = roundUp(size)
	rb.nodes = make(nodes, size)
	for i := uint64(0); i < size; i++ {
		rb.nodes[i] = &node{position: i}
	}
	rb.mask = size - 1 // so we don't have to do this with every put/get operation
}

// Put adds the provided item to the queue. If the queue is full, this
// call will block indefinitely.
func (rb *RingBuffer) Put(item interface{}) {
	var n *node
	pos := atomic.LoadUint64(&rb.queue)
L:
	for {
		n = rb.nodes[pos&rb.mask]
		seq := atomic.LoadUint64(&n.position)
		switch dif := seq - pos; {
		case dif == 0:
			if atomic.CompareAndSwapUint64(&rb.queue, pos, pos+1) {
				break L
			}
		case dif < 0:
			panic(`RingBuffer in a compromised state during a put operation.`)
		default:
			pos = atomic.LoadUint64(&rb.queue)
		}

		runtime.Gosched() // free up the cpu before the next iteration
	}

	n.data = item
	atomic.StoreUint64(&n.position, pos+1)
}

// Get will return the next item in the queue. This call will block
// if the queue is empty. This call will unblock when an item is added
// to the queue.
func (rb *RingBuffer) Get() (interface{}, error) {
	return rb.Poll(0)
}

// Poll will return the next item in the queue. This call will block
// if the queue is empty. This call will unblock when an item is added
// to the queue, or the timeout is reached. An error will be returned
// if the queue is a timeout occurs. A non-positive timeout will block
// indefinitely.
func (rb *RingBuffer) Poll(timeout time.Duration) (interface{}, error) {
	var (
		n     *node
		pos   = atomic.LoadUint64(&rb.dequeue)
		start time.Time
	)
	if timeout > 0 {
		start = time.Now()
	}
L:
	for {
		n = rb.nodes[pos&rb.mask]
		seq := atomic.LoadUint64(&n.position)
		switch dif := seq - (pos + 1); {
		case dif == 0:
			if atomic.CompareAndSwapUint64(&rb.dequeue, pos, pos+1) {
				break L
			}
		case dif < 0:
			panic(`RingBuffer in a compromised state during a get operation.`)
		default:
			pos = atomic.LoadUint64(&rb.dequeue)
		}

		if timeout > 0 && time.Since(start) >= timeout {
			return nil, ErrTimeout
		}

		runtime.Gosched() // free up the cpu before the next iteration
	}

	data := n.data
	n.data = nil
	atomic.StoreUint64(&n.position, pos+rb.mask+1)
	return data, nil
}

// NewRingBuffer will allocate, initialize, and return a ring buffer
// with the specified size.
func NewRingBuffer(size uint64) *RingBuffer {
	rb := &RingBuffer{}
	rb.init(size)
	return rb
}
