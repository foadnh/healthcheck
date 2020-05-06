// Provides easy to register and manage health checks for services.

package healthcheck

import (
	"context"
	"net/http"
	"reflect"
	"sync"
	"time"
)

// A checker is all that HealthCheck needs to know about the check.
type checker interface {
	check(ctx context.Context) error
	run(ctx context.Context)
	isInBackground() bool
	ticker() *time.Ticker
}

// A HealthCheck holds all details of checkers.
type HealthCheck struct {
	mutex            sync.RWMutex
	checkers         map[string]checker
	backgrounds      []backgroundChecker
	backgroundCancel context.CancelFunc
}

// A backgroundChecker holds a background check and its ticker.
type backgroundChecker struct {
	checker checker
	ticker  *time.Ticker
}

// New creates a new HealthCheck.
func New() *HealthCheck {
	h := &HealthCheck{
		checkers:    make(map[string]checker),
		backgrounds: make([]backgroundChecker, 0),
	}
	http.HandleFunc("/monitor", h.handler)
	return h
}

// Register will register a Checker for a HealthCheck.
// Params:
// 	name	Name of the check. Will be used in the detailed output.
// 	c 		The check function.
// 	timeout	Timeout of the check execution.
// 	opts	Checker options e.g. run in background.
func (h *HealthCheck) Register(name string, c Checker, timeout time.Duration, opts ...CheckOption) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.checkers[name] = newCheck(c, timeout, opts...)
}

// Run executes a goroutine that runs background checkers.
func (h *HealthCheck) Run(ctx context.Context) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	for _, c := range h.checkers {
		if c.isInBackground() {
			h.backgrounds = append(h.backgrounds, backgroundChecker{c, c.ticker()})
		}
	}
	go h.runInBackground(ctx)
}

// Close stops running of the background checkers and release resources.
func (h *HealthCheck) Close() {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	for i := range h.backgrounds {
		h.backgrounds[i].ticker.Stop()
	}
	if h.backgroundCancel != nil {
		h.backgroundCancel()
	}
}

// Check will check health of all checkers.
func (h *HealthCheck) check(ctx context.Context) map[string]error {
	var err error
	errs := make(map[string]error)
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	for name, checker := range h.checkers {
		err = checker.check(ctx)
		if err != nil {
			errs[name] = err
		}
	}
	return errs
}

// runInBackground listens to background checkers tickers and run the checkers checkers.
func (h *HealthCheck) runInBackground(ctx context.Context) {
	if len(h.backgrounds) == 0 {
		return
	}
	ctx, h.backgroundCancel = context.WithCancel(ctx)
	selects := make([]reflect.SelectCase, len(h.backgrounds)+1)
	for i := range h.backgrounds {
		h.backgrounds[i].checker.run(ctx)
		selects[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(h.backgrounds[i].ticker.C)}
	}
	selects[len(h.backgrounds)] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ctx.Done())}
	for {
		chosen, _, ok := reflect.Select(selects)
		if !ok {
			// Context canceled
			return
		}
		h.backgrounds[chosen].checker.run(ctx)
	}
}
