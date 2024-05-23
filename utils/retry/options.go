package retry

import "time"

var (
	defaultOptions = options{
		Strategy:  exp,
		MaxErrors: 5,

		// default 50% jitter
		Jitter: func(d time.Duration) time.Duration {
			return addJitter(d, 0.5)
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
	Breaker   *breaker
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
			return addJitter(d, jitter)
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

// Breaker uses sliding window algorithm to protect system from overload
// with default overload ratio 0.1 (10%).
//
// To prevent overload, Google SRE has some recommendations:
//
// First, we implement a per-request retry budget of up to three attempts.
// If a request has already failed three times, we let the failure bubble
// up to the caller. The rationale is that if a request has already landed
// on overloaded tasks three times, it's relatively unlikely that attempting
// it again will help because the whole datacenter is likely overloaded.
//
// Secondly, we implement a per-client retry budget. Each client keeps track
// of the ratio of requests that correspond to retries. A request will only
// be retried as long as this ratio is below 10%. The rationale is that if
// only a small subset of tasks are overloaded, there will be relatively
// little need to retry.
//
// A third approach has clients include a counter of how many times the
// request has already been tried in the request metadata. For instance,
// the counter starts at 0 in the first attempt and is incremented on every
// retry until it reaches 2, at which point the per-request budget causes
// it to stop being retried. Backends keep histograms of these values in
// recent history. When a backend needs to reject a request, it consults
// these histograms to determine the likelihood that other backend tasks
// are also overloaded. If these histograms reveal a significant amount of
// retries (indicating that other backend tasks are likely also overloaded),
// they return an "overloaded; don't retry" error response instead of the
// standard "task overloaded" error that triggers retries.
//
// Reference: https://sre.google/sre-book/handling-overload/
func Breaker(name string) Option {
	return BreakerWithOverloadRatio(name, 0.1)
}

// BreakerWithOverloadRatio is similar to Breaker, excepts that it
// accepts an additional param `overloadRatio` to specify the overload
// ratio to control the retry behavior, it's value should be greater
// than zero, else the default value 0.1 will be used.
//
// NOTE: generally, the default overload ratio 0.1 or even smaller value
// should be used, a big overload ratio will not really protect the
// backend system.
//
// Reference: https://sre.google/sre-book/handling-overload/
func BreakerWithOverloadRatio(name string, overloadRatio float64) Option {
	if overloadRatio <= 0 {
		overloadRatio = 0.1
	}
	br := getBreaker(name, overloadRatio)
	return func(opt options) options {
		opt.Breaker = br
		return opt
	}
}
