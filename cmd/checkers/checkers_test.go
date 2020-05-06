package checkers

import (
	"context"
	"errors"
	"testing"
)

func TestDatabase(t *testing.T) {
	type args struct {
		database Pinger
	}
	tests := []struct {
		name      string
		args      args
		wantError bool
	}{
		{
			"success",
			args{mockPing{}},
			false,
		},
		{
			"fail",
			args{mockPing{errors.New("ping failed")}},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			c := Database(tt.args.database)
			if got := c(ctx); (got != nil) != tt.wantError {
				t.Errorf("Database() = %v, want error %v", got, tt.wantError)
			}
		})
	}
}

func TestGoroutines(t *testing.T) {
	const maxInt = int(^uint(0) >> 1)
	type args struct {
		threshold int
	}
	tests := []struct {
		name      string
		args      args
		wantError bool
	}{
		{
			"success",
			args{maxInt},
			false,
		},
		{
			"fail",
			args{0},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			c := Goroutines(tt.args.threshold)
			if got := c(ctx); (got != nil) != tt.wantError {
				t.Errorf("Goroutines() = %v, want error %v", got, tt.wantError)
			}
		})
	}
}

type mockPing struct {
	err error
}

func (m mockPing) PingContext(_ context.Context) error {
	return m.err
}
