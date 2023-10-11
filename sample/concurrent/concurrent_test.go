package concurrent

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	mock "go.uber.org/mock/sample/concurrent/mock"
)

func call(ctx context.Context, m Math) (int, error) {
	result := make(chan int)
	go func() {
		result <- m.Sum(1, 2)
		close(result)
	}()
	select {
	case r := <-result:
		return r, nil
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

func waitForMocks(ctx context.Context, ctrl *gomock.Controller) error {
	ticker := time.NewTicker(1 * time.Millisecond)
	defer ticker.Stop()

	timeout := time.After(3 * time.Millisecond)

	for {
		select {
		case <-ticker.C:
			if ctrl.Satisfied() {
				return nil
			}
		case <-timeout:
			return fmt.Errorf("timeout waiting for mocks to be satisfied")
		case <-ctx.Done():
			return fmt.Errorf("context cancelled")
		}
	}
}

// TestConcurrentFails is expected to fail (and is disabled). It
// demonstrates how to use gomock.WithContext to interrupt the test
// from a different goroutine.
func TestConcurrentFails(t *testing.T) {
	t.Skip("Test is expected to fail, remove skip to trying running yourself.")
	ctrl, ctx := gomock.WithContext(context.Background(), t)
	m := mock.NewMockMath(ctrl)
	if _, err := call(ctx, m); err != nil {
		t.Error("call failed:", err)
	}
}

func TestConcurrentWorks(t *testing.T) {
	ctrl, ctx := gomock.WithContext(context.Background(), t)
	m := mock.NewMockMath(ctrl)
	m.EXPECT().Sum(1, 2).Return(3)
	if _, err := call(ctx, m); err != nil {
		t.Error("call failed:", err)
	}
}

func TestCancelWhenMocksSatisfied(t *testing.T) {
	ctrl, ctx := gomock.WithContext(context.Background(), t)
	m := mock.NewMockMath(ctrl)
	m.EXPECT().Sum(1, 2).Return(3).MinTimes(1)

	// This goroutine calls the mock and then waits for the context to be done.
	go func() {
		for {
			m.Sum(1, 2)
			select {
			case <-ctx.Done():
				return
			}
		}
	}()

	// waitForMocks spawns another goroutine which blocks until ctrl.Satisfied() is true.
	if err := waitForMocks(ctx, ctrl); err != nil {
		t.Error("call failed:", err)
	}
	ctrl.Finish()
}
