package retrier

import (
	"context"
	"errors"
	"testing"
	"time"
)

var i int

func genWork(returns []error) func() error {
	i = 0
	return func() error {
		i++
		if i > len(returns) {
			return nil
		}
		return returns[i-1]
	}
}

func genWorkWithCtx() func(ctx context.Context) error {
	i = 0
	return func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return errFoo
		default:
			i++
		}
		return nil
	}
}

func TestRetrier(t *testing.T) {
	r := New([]time.Duration{0, 10 * time.Millisecond}, NewWhiltelistClassifier([]error{errFoo}))

	err := r.Run(genWork([]error{errFoo, errFoo}))
	if err != nil {
		t.Error(err)
	}
	if i != 3 {
		t.Error("run wrong number of times")
	}

	err = r.Run(genWork([]error{errFoo, errBar}))
	if err != errBar {
		t.Error(err)
	}
	if i != 2 {
		t.Error("run wrong number of times")
	}

	err = r.Run(genWork([]error{errBar, errBaz}))
	if err != errBar {
		t.Error(err)
	}
	if i != 1 {
		t.Error("run wrong number of times")
	}
}

func TestRetrierCtx(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	r := New([]time.Duration{0, 10 * time.Millisecond}, NewWhiltelistClassifier(nil))

	err := r.RunCtx(ctx, genWorkWithCtx())
	if err != nil {
		t.Error(err)
	}
	if i != 1 {
		t.Error("run wrong number of times")
	}

	cancel()

	err = r.RunCtx(ctx, genWorkWithCtx())
	if err != errFoo {
		t.Error("context must be cancelled")
	}
	if i != 0 {
		t.Error("run wrong number of times")
	}
}

func TestRetrierNone(t *testing.T) {
	r := New(nil, nil)

	i = 0
	err := r.Run(func() error {
		i++
		return errFoo
	})
	if err != errFoo {
		t.Error(err)
	}
	if i != 1 {
		t.Error("run wrong number of times")
	}

	i = 0
	err = r.Run(func() error {
		i++
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	if i != 1 {
		t.Error("run wrong number of times")
	}
}

func TestRetrierJitter(t *testing.T) {
	r := new([]time.Duration{0, 10 * time.Millisecond, 4 * time.Hour}, nil)

	if r.calcSleep(0) != 0 {
		t.Error("Incorrect sleep calculated")
	}
	if r.calcSleep(1) != 10*time.Millisecond {
		t.Error("Incorrect sleep calculated")
	}
	if r.calcSleep(2) != 4*time.Hour {
		t.Error("Incorrect sleep calculated")
	}

	r.SetJitter(0.25)
	for i := 0; i < 20; i++ {
		if r.calcSleep(0) != 0 {
			t.Error("Incorrect sleep calculated")
		}

		slp := r.calcSleep(1)
		if slp < 7500*time.Microsecond || slp > 12500*time.Microsecond {
			t.Error("Incorrect sleep calculated")
		}

		slp = r.calcSleep(2)
		if slp < 3*time.Hour || slp > 5*time.Hour {
			t.Error("Incorrect sleep calculated")
		}
	}

	r.SetJitter(-1)
	if r.jitter != 0.25 {
		t.Error("Invalid jitter value accepted")
	}

	r.SetJitter(2)
	if r.jitter != 0.25 {
		t.Error("Invalid jitter value accepted")
	}
}

func TestRetrierThreadSafety(t *testing.T) {
	r := New([]time.Duration{0}, nil)
	for i := 0; i < 2; i++ {
		go func() {
			r.Run(func() error {
				return errors.New("error")
			})
		}()
	}
}

func ExampleRetrier() {
	r := New(ConstantBackoff(3, 100*time.Millisecond), nil)

	err := r.Run(func() error {
		// do some work
		return nil
	})

	if err != nil {
		// handle the case where the work failed three times
	}
}
