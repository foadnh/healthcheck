// Package checkers provides ready to use healthcheck.Checker functions.
// These are also simple Checker examples.
package checkers

import (
	"context"
	"fmt"
	"runtime"
)

//A Pinger interface is used to call Database Checker.
// A *sql.DB implements Pinger interface.
type Pinger interface {
	PingContext(ctx context.Context) error
}

// Goroutines checks if the number of running goroutines exceeds the threshold.
// It is also a proxy to check if we release resources that we occupy.
// It returns healthcheck.Checker type.
func Goroutines(threshold int) func(context.Context) error {
	return func(_ context.Context) error {
		if runtime.NumGoroutine() > threshold {
			return fmt.Errorf("number of gorutines exceeded %v", threshold)
		}
		return nil
	}
}

// Database checks if DB is up by pinging it.
// 	database	should be a *sql.DB.
// It returns healthcheck.Checker type.
func Database(database Pinger) func(context.Context) error {
	return func(ctx context.Context) error {
		return database.PingContext(ctx)
	}
}
