package healthcheck

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Types
type (
	// A Checker is a function responsible to check the status of a service and return an error if service is unhealthy.
	Checker func(ctx context.Context) error
	// A CheckOption is a modifier of a check. It can be passed while registering a checker to customize it.
	CheckOption func(c *check)

	// A checkerWithTimeout is a Checker function with timeout handler.
	checkerWithTimeout func(ctx context.Context) error
)

// Pre defined errors
var (
	// New Checkers have errNeverChecked error. It is useful for background checkers.
	errNeverChecked = errors.New("this checker never checked")
	// A errTimeout returns when a Checker reach the timeout.
	errTimeout = errors.New("timeout")
)

// A check holds data related to Checker and its results and other params.
type check struct {
	checker      checkerWithTimeout
	timeout      time.Duration
	interval     time.Duration
	threshold    uint
	err          error
	errorsInARow uint
	mutex        sync.RWMutex
}

// check checks the healthiness of a service.
// If the check is not a background check, it runs the Checker.
// It checks the threshold of the errors and return err value if threshold passes. err value can be nil.
func (c *check) check(ctx context.Context) error {
	if c.interval == 0 {
		c.run(ctx)
	}
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if c.errorsInARow < c.threshold {
		return nil
	}
	return c.err
}

// run executes a Checker.
func (c *check) run(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.err = c.checker(ctx)
	if c.err != nil {
		c.errorsInARow++
	} else {
		c.errorsInARow = 0
	}
}

// isInBackground shows if a check should be running in the background.
func (c *check) isInBackground() bool {
	return c.interval != 0
}

// ticker creates a ticker for a check.
func (c *check) ticker() *time.Ticker {
	return time.NewTicker(c.interval)
}

// newCheck creates a new instance of check.
// 	c		The Checker function
// 	timeout	The timeout of a check when executing
//	ops		Check Options e.g. InBackground
func newCheck(c Checker, timeout time.Duration, opts ...CheckOption) *check {
	s := check{
		checker: newCheckerWithTimeout(c, timeout),
		timeout: timeout,
		err:     errNeverChecked,
	}
	for i := range opts {
		opts[i](&s)
	}
	return &s
}

// newCheckerWithTimeout creates a Checker with a timeout
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
			return errTimeout
		}
	}
}

// InBackground forces a check to run in the background.
// Returns a CheckOption that can be passed during the Checker registration.
func InBackground(interval time.Duration) CheckOption {
	return func(c *check) {
		c.interval = interval
	}
}

// WithThreshold adds a threshold of errors in the row to show unhealthy state.
// Returns a CheckOption that can be passed during the Checker registration.
func WithThreshold(threshold uint) CheckOption {
	return func(c *check) {
		c.threshold = threshold
		c.errorsInARow = threshold
	}
}
