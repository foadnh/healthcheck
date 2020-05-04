package healthcheck

import (
	"context"
	"encoding/json"
	"net/http"
)

// handler will handle health check requests.
// Return 200 if all checkers pass, otherwise 503.
// If no parameter set, handler will only return the status code and no body.
// If detail query parameter set, it will show the detail of each checker and
// their errors, or OK status. The body is in JSON fromat.
func (h *HealthCheck) handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	errs := h.check(ctx)
	if len(errs) == 0 {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	_, ok := r.URL.Query()["detail"]
	if ok {
		h.handlerDetail(ctx, w, errs)
	}
}

// handlerDetail writes json version of details of checkers to the response.
func (h *HealthCheck) handlerDetail(_ context.Context, w http.ResponseWriter, errs map[string]error) {
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
