package retry

import "time"

var (
	defaultOptions = options{
		Strategy:  exp,
		MaxErrors: 5,

		// default 50% jitter
		Jitter: func(d time.Duration) time.Duration {
			return AddJitter(d, 0.5)
		},
		// default no-op hook
		Hook: func(attempts int, err error) {}, // dummy no-op hook
	}
)

type options struct {
	Attempts  int
	Sleep     time.Duration
	MaxSleep  time.Duration
	Strategy  strategy
	Jitter    strategy
	MaxErrors int
	Hook      func(attempts int, err error)
}

type Option func(options) options

// MaxSleep will restrict the retry sleep time to at most max.
func MaxSleep(max time.Duration) Option {
	return func(opt options) options {
		opt.MaxSleep = max
		return opt
	}
}

// MaxErrors set max errors to hold when retry for many times.
func MaxErrors(max int) Option {
	return func(opt options) options {
		opt.MaxErrors = max
		return opt
	}
}

// Hook let the retry function call the given hook when an error happens.
func Hook(hook func(attempts int, err error)) Option {
	return func(opt options) options {
		opt.Hook = hook
		return opt
	}
}

// NoJitter disables the retry function to add jitter to sleep time between each retry.
func NoJitter() Option {
	return func(opt options) options {
		opt.Jitter = nil
		return opt
	}
}

// J makes the retry function use specified jitter between each retry.
func J(jitter float64) Option {
	return func(opt options) options {
		opt.Jitter = func(d time.Duration) time.Duration {
			return AddJitter(d, jitter)
		}
		return opt
	}
}

// C makes the retry function sleep constant time between each retry.
func C() Option {
	return func(opt options) options {
		opt.Strategy = constant
		return opt
	}
}

// L makes the retry function sleep linear growing time between each retry.
func L(step time.Duration) Option {
	l := linear{step: step}
	return func(opt options) options {
		opt.Strategy = l.next
		return opt
	}
}
