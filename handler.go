package healthcheck

import (
	"context"
	"encoding/json"
	"net/http"
)

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
