package healthcheck

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		want *HealthCheck
	}{
		{
			"simple",
			&HealthCheck{
				checkers:    make(map[string]checker),
				backgrounds: make([]backgroundChecker, 0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serveMux := http.NewServeMux()
			if got := New(serveMux, "/healthcheck"); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Only check name of checkers not actual values.
func TestHealthCheck_Register(t *testing.T) {
	checkerFunc := func(_ context.Context) error { return nil }
	type fields struct {
		checkers map[string]checker
	}
	type args struct {
		name    string
		c       Checker
		timeout time.Duration
		opts    []CheckOption
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		checkerNames []string
	}{
		{
			"empty",
			fields{
				map[string]checker{},
			},
			args{
				"checker_1",
				checkerFunc,
				time.Second,
				[]CheckOption{},
			},
			[]string{"checker_1"},
		},
		{
			"with_one",
			fields{
				map[string]checker{
					"checker_1": &mockCheck{},
				},
			},
			args{
				"checker_2",
				checkerFunc,
				time.Second,
				[]CheckOption{},
			},
			[]string{"checker_1", "checker_2"},
		},
		{
			"with_2_options",
			fields{
				map[string]checker{},
			},
			args{
				"checker_1",
				checkerFunc,
				time.Second,
				[]CheckOption{
					func(_ *check) {},
					func(_ *check) {},
				},
			},
			[]string{"checker_1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &HealthCheck{
				checkers: tt.fields.checkers,
			}
			h.Register(tt.args.name, tt.args.c, tt.args.timeout, tt.args.opts...)
			if len(h.checkers) != len(tt.checkerNames) {
				t.Errorf("HealthCheck.Register() len(checkers) = %v, want %v", len(h.checkers), len(tt.checkerNames))
			}
			for _, cn := range tt.checkerNames {
				if _, ok := h.checkers[cn]; !ok {
					t.Errorf("HealthCheck.Register() checkers[%v] not exist", cn)
				}
			}
		})
	}
}

func TestHealthCheck_Run(t *testing.T) {
	type fields struct {
		checkers map[string]checker
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		wantBackgrounds []string
	}{
		{
			"empty",
			fields{},
			args{context.Background()},
			[]string{},
		},
		{
			"2_fronts",
			fields{
				map[string]checker{
					"front_1": &mockCheck{},
					"front_2": &mockCheck{},
				},
			},
			args{context.Background()},
			[]string{},
		},
		{
			"mix",
			fields{
				map[string]checker{
					"front_1":      &mockCheck{},
					"front_2":      &mockCheck{},
					"background_1": &mockCheck{interval: time.Minute},
					"background_2": &mockCheck{interval: time.Hour},
				},
			},
			args{context.Background()},
			[]string{
				"background_1",
				"background_2",
			},
		},
	}
	for _, tt := range tests {
		backgroundMap := make(map[checker]string)
		for _, b := range tt.wantBackgrounds {
			backgroundMap[tt.fields.checkers[b]] = b
		}
		t.Run(tt.name, func(t *testing.T) {
			h := &HealthCheck{
				checkers: tt.fields.checkers,
			}
			h.Run(tt.args.ctx)
			if len(tt.wantBackgrounds) != len(h.backgrounds) {
				t.Errorf("Run() len(backgrounds) = %v, want %v", len(h.backgrounds), len(tt.wantBackgrounds))
			}
			if len(tt.wantBackgrounds) > 0 {
				if h.backgroundCancel == nil {
					t.Error("Run() backgroundCancel expected, got nil")
				}
			}
			for i := range h.backgrounds {
				bc := h.backgrounds[i].checker
				ec := tt.fields.checkers[backgroundMap[bc]]
				if !reflect.DeepEqual(bc, ec) {
					t.Errorf("Run() backgrounds[%v] = %v, want %v", backgroundMap[bc], bc, ec)
				}
			}
		})
	}
}

func TestHealthCheck_Close(t *testing.T) {
	type fields struct {
		backgrounds []backgroundChecker
	}
	tests := []struct {
		name                 string
		fields               fields
		withBackgroundCancel bool
	}{
		{
			"no_background_checker",
			fields{},
			false,
		},
		{
			"2_background_checker",
			fields{
				[]backgroundChecker{
					{
						ticker: time.NewTicker(time.Nanosecond),
					},
					{
						ticker: time.NewTicker(time.Nanosecond),
					},
				},
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				backgroundCancel    func()
				backgroundCancelled bool
			)
			if tt.withBackgroundCancel {
				backgroundCancel = func() {
					backgroundCancelled = true
				}
			}
			h := &HealthCheck{
				backgrounds:      tt.fields.backgrounds,
				backgroundCancel: backgroundCancel,
			}
			time.Sleep(2 * time.Millisecond)
			h.Close()
			if backgroundCancelled != tt.withBackgroundCancel {
				t.Errorf("Close() backgroundCancel got = %v, want %v", backgroundCancelled, tt.withBackgroundCancel)
			}
			for i := range h.backgrounds {
				<-h.backgrounds[i].ticker.C
			}
			time.Sleep(2 * time.Millisecond)
			for i := range h.backgrounds {
				ticker := h.backgrounds[i].ticker
				select {
				case <-ticker.C:
					t.Error("Close() background checkers ticker is not stopped")
				default:
					// OK
				}
			}
		})
	}
}

func TestHealthCheck_check(t *testing.T) {
	testErr := errors.New("HealthCheck.check error")
	type fields struct {
		checkers map[string]checker
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]error
	}{
		{
			"empty",
			fields{
				checkers: map[string]checker{},
			},
			args{context.Background()},
			map[string]error{},
		},
		{
			"2_error_1_healthy",
			fields{
				checkers: map[string]checker{
					"checker_1": &mockCheck{err: testErr},
					"checker_2": &mockCheck{},
					"checker_3": &mockCheck{err: testErr},
				},
			},
			args{context.Background()},
			map[string]error{
				"checker_1": testErr,
				"checker_3": testErr,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &HealthCheck{
				checkers: tt.fields.checkers,
			}
			if got := h.check(tt.args.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("check() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Not checking if select part is actually working as we expected.
func TestHealthCheck_runInBackground(t *testing.T) {
	testErr := errors.New("HealthCheck.runInBackground error")
	type fields struct {
		backgrounds []backgroundChecker
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			"empty",
			fields{},
			args{context.Background()},
		},
		{
			"2_backgrounds",
			fields{[]backgroundChecker{
				{
					&mockCheck{runErr: testErr},
					time.NewTicker(time.Millisecond),
				},
				{
					&mockCheck{runErr: testErr},
					time.NewTicker(time.Millisecond),
				},
			}},
			args{context.Background()},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &HealthCheck{
				backgrounds: tt.fields.backgrounds,
			}
			ctx, cancel := context.WithTimeout(tt.args.ctx, 10*time.Millisecond)
			defer cancel()
			h.runInBackground(ctx)
			for i := range tt.fields.backgrounds {
				if err := tt.fields.backgrounds[i].checker.(*mockCheck).err; err != testErr {
					t.Errorf("runInBackground() checker.err = %v, want %v", err, testErr)
				}
			}
		})
	}
}

type mockCheck struct {
	interval time.Duration
	err      error
	runErr   error
}

func (m *mockCheck) check(_ context.Context) error {
	return m.err
}

func (m *mockCheck) run(_ context.Context) {
	m.err = m.runErr
}

func (m *mockCheck) isInBackground() bool {
	return m.interval != 0
}

func (m *mockCheck) ticker() *time.Ticker {
	return time.NewTicker(m.interval)
}
