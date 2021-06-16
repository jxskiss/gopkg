package syncx

import "context"

// TryLocker represents an object that can be locked and unlocked.
type TryLocker interface {
	TryLock() bool
	Lock(ctx context.Context) error
	Unlock()
}

type tryLock struct {
	c chan struct{}
}

// NewTryLock creates a new TryLocker.
func NewTryLock() TryLocker {
	return &tryLock{c: make(chan struct{}, 1)}
}

// TryLock tries to acquire the lock.
// It returns true if it acquires the lock successfully, otherwise false.
func (p *tryLock) TryLock() bool {
	select {
	case p.c <- struct{}{}:
		return true
	default:
		return false
	}
}

// Lock tries to acquire the lock.
// It waits ctx.Done(), if the context is finished before it successfully
// acquires the lock, it returns the error from ctx.Done().
// A nil error means the lock is acquired successfully, else the error
// may be context.Cancelled or context.DeadlineExceeded.
func (p *tryLock) Lock(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case p.c <- struct{}{}:
		return nil
	}
}

// Unlock releases the lock if it is in locked state.
func (p *tryLock) Unlock() {
	select {
	case <-p.c:
	default:
	}
}
