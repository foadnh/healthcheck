package healthcheck

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

func Example() {
	ctx := context.Background()
	var (
		serviceIsUnhealthy bool
		// If Checker has changing variable, use mutex to handle race condition.
		mutex sync.RWMutex
	)
	dummyHealthyChecker := func(_ context.Context) error {
		return nil
	}
	dummyUnhealthyChecker := func(_ context.Context) error {
		mutex.RLock()
		defer mutex.RUnlock()
		if serviceIsUnhealthy {
			return errors.New("not_feeling_good")
		}
		return nil
	}

	serveMux := http.NewServeMux()

	h := New(serveMux, "/healthcheck")
	h.Register("dummy_healthy_checker", dummyHealthyChecker, time.Second)
	h.Register("dummy_unhealthy_checker_in_background", dummyUnhealthyChecker, time.Second*2, InBackground(time.Millisecond))
	h.Register("dummy_unhealthy_checker_with_threshold", dummyUnhealthyChecker, time.Second, WithThreshold(2))
	h.Run(ctx)
	defer h.Close()

	time.Sleep(time.Millisecond * 10)

	// check method is not exposed. Don't use it in your codes.
	fmt.Println(h.check(ctx))

	// Lets make unhealthy checkers fail
	mutex.Lock()
	serviceIsUnhealthy = true
	mutex.Unlock()
	time.Sleep(time.Millisecond * 10)

	fmt.Println(h.check(ctx))
	fmt.Println(h.check(ctx))

	// Output:
	// map[]
	// map[dummy_unhealthy_checker_in_background:not_feeling_good]
	// map[dummy_unhealthy_checker_in_background:not_feeling_good dummy_unhealthy_checker_with_threshold:not_feeling_good]
}
