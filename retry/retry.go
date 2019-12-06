// Package retry implements frequently used retry strategies and options.
package retry

import "time"

// Stop is used to indicate the retry function to stop retry.
type Stop struct {
	Err error
}

func (e Stop) Error() string {
	return e.Err.Error()
}

// Default will call param function f at most 3 times before returning error.
// Between each retry will sleep an exponential time starts at 500ms.
// In case of all retry fails, the total sleep time will be about 750ms - 2250ms.
//
// It is shorthand for Retry(3, 500*time.Millisecond, f).
func Default(f func() error, opts ...Option) Result {
	return Retry(3, 500*time.Millisecond, f, opts...)
}

// Retry retry the target function with exponential sleep time.
// It implements algorithm described in https://upgear.io/blog/simple-golang-retry-function/.
func Retry(attempts int, sleep time.Duration, f func() error, opts ...Option) Result {
	opt := defaultOptions
	opt.Attempts = attempts
	opt.Sleep = sleep
	return retry(opt, f, opts...)
}

// Const retry the target function with constant sleep time.
// It is shorthand for Retry(attempts, sleep, f, C()).
func Const(attempts int, sleep time.Duration, f func() error, opts ...Option) Result {
	opt := defaultOptions
	opt.Attempts = attempts
	opt.Sleep = sleep
	opt.Strategy = constant
	return retry(opt, f, opts...)
}

// Linear retry the target function with linear sleep time.
// It is shorthand for Retry(attempts, sleep, f, L(sleep)).
func Linear(attempts int, sleep time.Duration, f func() error, opts ...Option) Result {
	opt := defaultOptions
	opt.Attempts = attempts
	opt.Sleep = sleep
	opt.Strategy = linear{sleep}.next
	return retry(opt, f, opts...)
}

// Forever retry the target function endlessly if it returns error.
// To stop the the retry loop on error, the target function should return Stop.
//
// The caller should take care of dead loop.
func Forever(sleep, maxSleep time.Duration, f func() error, opts ...Option) Result {
	opt := defaultOptions
	opt.Sleep = sleep
	opt.MaxSleep = maxSleep
	return retry(opt, f, opts...)
}

// retry do the retry job according given options.
func retry(opt options, f func() error, opts ...Option) (r Result) {
	for _, o := range opts {
		opt = o(opt)
	}

	var err error
	r.Attempts++
	if err = f(); err == nil {
		r.Ok = true
		return
	}

	var merr = NewSizedError(opt.MaxErrors)
	var sleep = opt.Sleep
	for {
		// attempts <= 0 means retry forever.
		if opt.Attempts > 0 && r.Attempts >= opt.Attempts {
			break
		}
		if _, ok := err.(Stop); ok {
			break
		}
		merr.Append(err)
		opt.Hook(r.Attempts, err)
		if opt.MaxSleep > 0 && sleep > opt.MaxSleep {
			sleep = opt.MaxSleep
		}
		if opt.Jitter == nil {
			time.Sleep(sleep)
		} else {
			time.Sleep(opt.Jitter(sleep))
		}
		r.Attempts++
		if err = f(); err == nil {
			r.Ok = true
			break
		}
		sleep = opt.Strategy(sleep)
	}
	if err != nil {
		if s, ok := err.(Stop); ok {
			// Return the original error for later checking.
			merr.Append(s.Err)
			opt.Hook(r.Attempts, s.Err)
		} else {
			merr.Append(err)
			opt.Hook(r.Attempts, err)
		}
	}
	r.Error = merr.ErrOrNil()
	return
}

type Result struct {
	Ok       bool
	Attempts int
	Error    error
}
