package healthcheck

import (
	"context"
	"errors"
	"time"
)

type (
	Checker       func(ctx context.Context) error
	CheckerOption func(c *checker)

	checkerWithTimeout func(ctx context.Context) error
)

var neverCheckedErr = errors.New("this checker never checked")
var timoutErr = errors.New("timeout")

type checker struct {
	checker     checkerWithTimeout
	timeout     time.Duration
	interval    time.Duration
	capacity    uint
	err         error
	errorsOnRow uint
}

func (c *checker) check(ctx context.Context) error {
	if c.interval == 0 {
		c.run(ctx)
	}
	if c.errorsOnRow < c.capacity {
		return nil
	}
	return c.err
}

func (c *checker) run(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	c.err = c.checker(ctx)
	if c.err != nil {
		c.errorsOnRow++
	} else {
		c.errorsOnRow = 0
	}
}

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

func newCheckerWithTimeout(c Checker, timeout time.Duration) checkerWithTimeout {
	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		errChan := make(chan error)
		go func() {
			errChan <- c(ctx)
		}()
		select {
		case err := <-errChan:
			return err
		case <-ctx.Done():
			return timoutErr
		}
	}
}

func InBackground(interval time.Duration) CheckerOption {
	return func(c *checker) {
		c.interval = interval
	}
}

func WithCapacity(capacity uint) CheckerOption {
	return func(c *checker) {
		c.capacity = capacity
		c.errorsOnRow = capacity
	}
}
