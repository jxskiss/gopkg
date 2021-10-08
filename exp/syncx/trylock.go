package syncx

import "context"

// TryLocker represents an object that can be locked and unlocked.
type TryLocker interface {
	TryLock() bool
	LockTimeout(ctx context.Context) error
	Lock()
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

// LockTimeout tries to acquire the lock.
// It waits ctx.Done(), if the context is finished before it successfully
// acquires the lock, it returns the error from ctx.Done().
// A nil error means the lock is acquired successfully.
func (p *tryLock) LockTimeout(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case p.c <- struct{}{}:
		return nil
	}
}

// Lock acquires the lock, it waits the lock if it's already acquired
// by someone else.
func (p *tryLock) Lock() {
	p.c <- struct{}{}
}

// Unlock releases the lock if it is in locked state.
func (p *tryLock) Unlock() {
	select {
	case <-p.c:
	default:
	}
}
