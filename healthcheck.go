// Provides easy to register and manage health checks for services.

package healthcheck

import (
	"context"
	"net/http"
	"reflect"
	"sync"
	"time"
)

// A HealthCheck holds all details of checkers.
type HealthCheck struct {
	mutex            sync.RWMutex
	checkers         map[string]*checker
	backgrounds      []backgroundChecker
	backgroundCancel context.CancelFunc
}

// A backgroundChecker holds a background checker and its ticker.
type backgroundChecker struct {
	checker *checker
	ticker  *time.Ticker
}

// New creates a new HealthCheck.
func New() *HealthCheck {
	h := &HealthCheck{
		checkers:    make(map[string]*checker),
		backgrounds: make([]backgroundChecker, 0),
	}
	http.HandleFunc("/monitor", h.handler)
	return h
}

// Register will register a Checker for a HealthCheck.
// Params:
// 	name	Name of the checker. Will be used in the detailed output.
// 	c 		The checker function.
// 	timeout	Timeout of the checker execution.
// 	opts	Checker options e.g. run in background.
func (h *HealthCheck) Register(name string, c Checker, timeout time.Duration, opts ...CheckerOption) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.checkers[name] = newChecker(c, timeout, opts...)
}

// Run executes a goroutine that runs background checkers.
func (h *HealthCheck) Run(ctx context.Context) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	for _, c := range h.checkers {
		if c.interval != 0 {
			h.backgrounds = append(h.backgrounds, backgroundChecker{c, tickerOfChecker(c)})
		}
	}
	ctx, h.backgroundCancel = context.WithCancel(ctx)
	go h.runInBackground(ctx)
}

// Close stops running of the background checkers and release resources.
func (h *HealthCheck) Close() {
	h.mutex.RLock()
	for i := range h.backgrounds {
		h.backgrounds[i].ticker.Stop()
	}
	h.backgroundCancel()
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

// tickerOfChecker creates a ticker for a checker.
func tickerOfChecker(c *checker) *time.Ticker {
	return time.NewTicker(c.interval)
}

// runInBackground listens to background checkers tickers and run the checkers checkers.
func (h *HealthCheck) runInBackground(ctx context.Context) {
	if len(h.backgrounds) == 0 {
		return
	}
	selects := make([]reflect.SelectCase, len(h.backgrounds)+1)
	for i := range h.backgrounds {
		h.backgrounds[i].checker.run(ctx)
		selects[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(h.backgrounds[i].ticker.C)}
	}
	selects[len(h.backgrounds)] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ctx.Done())}
	for {
		chosen, _, ok := reflect.Select(selects)
		if !ok {
			return
		}
		// To run in background if we have many slow goroutines
		h.backgrounds[chosen].checker.run(ctx)
	}
}
