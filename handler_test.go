package healthcheck

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestHealthCheck_handler(t *testing.T) {
	type fields struct {
		checkers map[string]checker
	}
	type args struct {
		r *http.Request
	}
	type want struct {
		code int
		body bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			"empty",
			fields{map[string]checker{}},
			args{httptest.NewRequest(http.MethodGet, "/metrics", nil)},
			want{
				http.StatusOK,
				false,
			},
		},
		{
			"success",
			fields{map[string]checker{
				"checker_1": &mockCheck{},
				"checker_2": &mockCheck{},
			}},
			args{httptest.NewRequest(http.MethodGet, "/metrics", nil)},
			want{
				http.StatusOK,
				false,
			},
		},
		{
			"fail",
			fields{map[string]checker{
				"checker_1": &mockCheck{},
				"checker_2": &mockCheck{err: errors.New("checker_2 failed")},
			}},
			args{httptest.NewRequest(http.MethodGet, "/metrics", nil)},
			want{
				http.StatusServiceUnavailable,
				false,
			},
		},
		{
			"detail",
			fields{map[string]checker{
				"checker_1": &mockCheck{},
				"checker_2": &mockCheck{err: errors.New("checker_2 failed")},
			}},
			args{httptest.NewRequest(http.MethodGet, "/metrics?detail", nil)},
			want{
				http.StatusServiceUnavailable,
				true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			h := &HealthCheck{
				checkers: tt.fields.checkers,
			}
			h.handler(w, tt.args.r)
			if w.Code != tt.want.code {
				t.Errorf("handler() code = %v, want %v", w.Code, tt.want.code)
			}
			if (w.Body.Len() != 0) != tt.want.body {
				t.Errorf("handler() body = %v, want body %v", w.Body, tt.want.body)
			}
		})
	}
}

func TestHealthCheck_handlerDetail(t *testing.T) {
	type fields struct {
		checkers map[string]checker
	}
	type args struct {
		ctx  context.Context
		errs map[string]error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]string
	}{
		{
			"empty",
			fields{map[string]checker{}},
			args{
				context.Background(),
				map[string]error{},
			},
			map[string]string{},
		},
		{
			"no_error",
			fields{map[string]checker{
				"checker_1": &mockCheck{},
				"checker_2": &mockCheck{},
			}},
			args{
				context.Background(),
				map[string]error{},
			},
			map[string]string{
				"checker_1": "OK",
				"checker_2": "OK",
			},
		},
		{
			"2_errors_2_success",
			fields{map[string]checker{
				"checker_1": &mockCheck{},
				"checker_2": &mockCheck{},
				"checker_3": &mockCheck{},
				"checker_4": &mockCheck{},
			}},
			args{
				context.Background(),
				map[string]error{
					"checker_2": errors.New("checker_2 failed"),
					"checker_4": errors.New("checker_4 failed"),
				},
			},
			map[string]string{
				"checker_1": "OK",
				"checker_2": "checker_2 failed",
				"checker_3": "OK",
				"checker_4": "checker_4 failed",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			h := &HealthCheck{
				checkers: tt.fields.checkers,
			}
			h.handlerDetail(tt.args.ctx, w, tt.args.errs)
			got := make(map[string]string)
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Errorf("handlerDetail() response is not JSON %v", w.Body.String())
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handlerDetail() body = %v, want %v", got, tt.want)
			}
		})
	}
}
