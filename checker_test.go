package healthcheck

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestInBackground(t *testing.T) {
	type args struct {
		interval time.Duration
	}
	tests := []struct {
		name string
		args args
		c    check
	}{
		{
			"in_background",
			args{
				time.Minute,
			},
			check{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := InBackground(tt.args.interval)
			opt(&tt.c)
			if tt.c.interval != tt.args.interval {
				t.Errorf("InBackground().interval = %v, want %v", tt.c.interval, tt.args.interval)
			}
		})
	}
}

func TestWithThreshold(t *testing.T) {
	type args struct {
		threshold uint
	}
	tests := []struct {
		name string
		args args
		c    check
	}{
		{
			"in_background",
			args{
				5,
			},
			check{},
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := WithThreshold(tt.args.threshold)
			opt(&tt.c)
			if tt.c.threshold != tt.args.threshold {
				t.Errorf("InBackground().threshold = %v, want %v", tt.c.threshold, tt.args.threshold)
			}
		})
	}
}

func Test_check_check(t *testing.T) {
	testErr := errors.New("check.check error")
	checkerCreator := func(err error) checkerWithTimeout {
		return func(_ context.Context) error {
			return err
		}
	}

	type fields struct {
		checker     checkerWithTimeout
		timeout     time.Duration
		interval    time.Duration
		threshold   uint
		err         error
		errorsOnRow uint
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   error
	}{
		{
			"simple",
			fields{
				checker: checkerCreator(testErr),
				err:     nil,
			},
			args{context.Background()},
			testErr,
		},
		{
			"with_threshold_not_passed",
			fields{
				checker:     checkerCreator(testErr),
				err:         testErr,
				threshold:   3,
				errorsOnRow: 1,
			},
			args{context.Background()},
			nil,
		},
		{
			"with_threshold_passed",
			fields{
				checker:     checkerCreator(testErr),
				err:         testErr,
				threshold:   2,
				errorsOnRow: 1,
			},
			args{context.Background()},
			testErr,
		},
		{
			"in_background",
			fields{
				checker:  checkerCreator(nil),
				err:      testErr,
				interval: time.Second,
			},
			args{context.Background()},
			testErr,
		},
		{
			"in_background_with_threshold_not_passed",
			fields{
				checker:   checkerCreator(testErr),
				err:       testErr,
				threshold: 1,
				interval:  time.Second,
			},
			args{context.Background()},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &check{
				checker:      tt.fields.checker,
				timeout:      tt.fields.timeout,
				interval:     tt.fields.interval,
				threshold:    tt.fields.threshold,
				err:          tt.fields.err,
				errorsInARow: tt.fields.errorsOnRow,
			}
			if got := c.check(tt.args.ctx); got != tt.want {
				t.Errorf("check() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_check_run(t *testing.T) {
	testErr := errors.New("check.run error")
	checkerCreator := func(err error) checkerWithTimeout {
		return func(_ context.Context) error {
			return err
		}
	}
	type fields struct {
		checker      checkerWithTimeout
		timeout      time.Duration
		interval     time.Duration
		threshold    uint
		err          error
		errorsInARow uint
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name             string
		fields           fields
		args             args
		wantErr          error
		wantErrorsInARow uint
	}{
		{
			"with_error",
			fields{
				checker:      checkerCreator(testErr),
				err:          nil,
				errorsInARow: 0,
			},
			args{context.Background()},
			testErr,
			1,
		},
		{
			"with_error",
			fields{
				checker:      checkerCreator(nil),
				err:          testErr,
				errorsInARow: 2,
			},
			args{context.Background()},
			nil,
			0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &check{
				checker:      tt.fields.checker,
				timeout:      tt.fields.timeout,
				interval:     tt.fields.interval,
				threshold:    tt.fields.threshold,
				err:          tt.fields.err,
				errorsInARow: tt.fields.errorsInARow,
			}
			c.run(tt.args.ctx)
			if c.err != tt.wantErr {
				t.Errorf("check.run() err = %v, want %v", c.err, tt.wantErr)
			}
			if c.errorsInARow != tt.wantErrorsInARow {
				t.Errorf("check.run() errorsInARow = %v, want %v", c.errorsInARow, tt.wantErrorsInARow)
			}
		})
	}
}

func Test_check_isInBackground(t *testing.T) {
	type fields struct {
		interval time.Duration
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			"in_background",
			fields{
				time.Minute,
			},
			true,
		},
		{
			"not_in_background",
			fields{},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &check{
				interval: tt.fields.interval,
			}
			if got := c.isInBackground(); got != tt.want {
				t.Errorf("isInBackground() = %v, want %v", got, tt.want)
			}
		})
	}
}

// We can't compare actual value in here. Only check if function works.
func Test_check_ticker(t *testing.T) {
	type fields struct {
		interval time.Duration
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			"simple",
			fields{
				interval: time.Minute,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &check{
				interval: tt.fields.interval,
			}
			if got := c.ticker(); got == nil {
				t.Errorf("ticker() want *time.Ticker, got %v", got)
			}
		})
	}
}

func Test_newChecker(t *testing.T) {
	testErr := errors.New("newCheck error")
	type args struct {
		c       Checker
		timeout time.Duration
		opts    []CheckOption
	}
	tests := []struct {
		name string
		args args
		want *check
	}{
		{
			"simple",
			args{
				func(_ context.Context) error { return testErr },
				time.Minute,
				[]CheckOption{},
			},
			&check{
				timeout: time.Minute,
				err:     neverCheckedErr,
			},
		},
		{
			"in_background",
			args{
				func(_ context.Context) error { return testErr },
				time.Minute,
				[]CheckOption{InBackground(time.Hour)},
			},
			&check{
				timeout:  time.Minute,
				err:      neverCheckedErr,
				interval: time.Hour,
			},
		},
		{
			"with_threshold",
			args{
				func(_ context.Context) error { return testErr },
				time.Minute,
				[]CheckOption{WithThreshold(5)},
			},
			&check{
				timeout:      time.Minute,
				err:          neverCheckedErr,
				threshold:    5,
				errorsInARow: 5,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newCheck(tt.args.c, tt.args.timeout, tt.args.opts...)
			if err := got.checker(context.Background()); err != testErr {
				t.Errorf("newCheck().check() = %v, want %v", err, testErr)
			}
			got.checker = nil
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newCheck() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newCheckerWithTimeout(t *testing.T) {
	checkerCreator := func(sleep time.Duration) Checker {
		return func(_ context.Context) error {
			time.Sleep(sleep)
			return nil
		}
	}
	type args struct {
		c       Checker
		timeout time.Duration
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			"timeout_not_passed",
			args{
				checkerCreator(0),
				time.Second,
			},
			nil,
		},
		{
			"timeout_passed",
			args{
				checkerCreator(time.Second),
				0,
			},
			timeoutErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newCheckerWithTimeout(tt.args.c, tt.args.timeout)
			if err := c(context.Background()); err != tt.want {
				t.Errorf("newCheckerWithTimeout()().error = %v, want %v", err, tt.want)
			}
		})
	}
}
