package healthcheck

import (
	"context"
	"errors"
	"time"
)

// Types
type (
	// A Checker is a function responsible to check the status of a service and return an error if checker is unhealthy.
	Checker func(ctx context.Context) error
	// A CheckerOption is a modifier of a checker. It can be passed while registering a checker to customize it.
	CheckerOption func(c *checker)

	// A checkerWithTimeout is a Checker function with timeout handl
	checkerWithTimeout func(ctx context.Context) error
)

// Pre defined errors
var (
	// New Checkers have neverCheckedErr error. It is useful for background checkers.
	neverCheckedErr = errors.New("this checker never checked")
	// A timeoutErr returns when a checker reach the timeout.
	timeoutErr = errors.New("timeout")
)

// A checker holds data related to Checker and its results and other params.
type checker struct {
	checker      checkerWithTimeout
	timeout      time.Duration
	interval     time.Duration
	threshold    uint
	err          error
	errorsInARow uint
}

// Check checks the health of a Checker.
// If the Checker is not a background checker, it runs the checker.
// It checks the threshold of the checker and return err value if threshold passes. err value can be nil.
func (c *checker) check(ctx context.Context) error {
	if c.interval == 0 {
		c.run(ctx)
	}
	if c.errorsInARow < c.threshold {
		return nil
	}
	return c.err
}

// run executes a Checker.
func (c *checker) run(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	c.err = c.checker(ctx)
	if c.err != nil {
		c.errorsInARow++
	} else {
		c.errorsInARow = 0
	}
}

// newChecker creates a new instance of checker.
// 	c		The Checker function
// 	timeout	The timeout of a checker when executing
//	ops		Checker Options e.g. InBackground
func newChecker(c Checker, timeout time.Duration, opts ...CheckerOption) *checker {
	s := checker{
		checker: newCheckerWithTimeout(c, timeout),
		timeout: timeout,
		err:     neverCheckedErr,
	}
	for i := range opts {
		opts[i](&s)
	}
	return &s
}

// newCheckerWithTimeout creates a checker with a timeout
func newCheckerWithTimeout(c Checker, timeout time.Duration) checkerWithTimeout {
	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		errChan := make(chan error, 1)
		go func() {
			errChan <- c(ctx)
		}()
		select {
		case err := <-errChan:
			return err
		case <-ctx.Done():
			return timeoutErr
		}
	}
}

// InBackground force a checker to run in the background.
// Returns a CheckerOption that can be passed during the Checker registration.
func InBackground(interval time.Duration) CheckerOption {
	return func(c *checker) {
		c.interval = interval
	}
}

// WithThreshold adds a threshold of errors in the row to actually checker shows unhealthy state.
// Returns a CheckerOption that can be passed during the Checker registration.
func WithThreshold(threshold uint) CheckerOption {
	return func(c *checker) {
		c.threshold = threshold
		c.errorsInARow = threshold
	}
}
