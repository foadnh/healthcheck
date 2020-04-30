// Provides easy to register and manage health checks for services.

package healthcheck

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"sync"
	"time"
)

type HealthCheck struct {
	mutex       sync.RWMutex
	checkers    map[string]*checker
	backgrounds []backgroundChecker
}

type backgroundChecker struct {
	checker *checker
	ticker  *time.Ticker
}

func New() *HealthCheck {
	h := &HealthCheck{
		checkers:    make(map[string]*checker),
		backgrounds: make([]backgroundChecker, 0),
	}
	http.HandleFunc("/monitor", h.handler)
	return h
}

func (h *HealthCheck) Register(name string, c Checker, timeout time.Duration, opts ...CheckerOption) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.checkers[name] = newChecker(c, timeout, opts...)
}

func (h *HealthCheck) Run(ctx context.Context) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	for _, c := range h.checkers {
		if c.interval != 0 {
			h.backgrounds = append(h.backgrounds, backgroundChecker{c, tickerOfChecker(ctx, c)})
		}
	}
	go h.runInBackground(ctx)
}

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

func tickerOfChecker(_ context.Context, c *checker) *time.Ticker {
	return time.NewTicker(c.interval)
}

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

func (h *HealthCheck) handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	errs := h.check(ctx)
	if len(errs) == 0 {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	result := make(map[string]string)
	for name := range h.checkers {
		err, ok := errs[name]
		if ok {
			result[name] = err.Error()
		} else {
			result[name] = "OK"
		}
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	_ = encoder.Encode(result)
}
