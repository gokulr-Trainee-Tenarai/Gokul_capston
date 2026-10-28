package resilience

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type CircuitState string

const (
	StateClosed   CircuitState = "CLOSED"
	StateOpen     CircuitState = "OPEN"
	StateHalfOpen CircuitState = "HALF_OPEN"
)

var ErrCircuitOpen = errors.New("circuit breaker is OPEN: downstream/product service unavailable")

type CircuitBreaker struct {
	mu           sync.Mutex
	state        CircuitState
	failureCount int
	maxFailures  int
	timeout      time.Duration
	openedAt     time.Time
}

func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:       StateClosed,
		maxFailures: maxFailures,
		timeout:     timeout,
	}
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()
	if cb.state == StateOpen {
		if time.Since(cb.openedAt) > cb.timeout {
			cb.state = StateHalfOpen
		} else {
			cb.mu.Unlock()
			return ErrCircuitOpen
		}
	}
	cb.mu.Unlock()

	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failureCount++
		if cb.failureCount >= cb.maxFailures || cb.state == StateHalfOpen {
			cb.state = StateOpen
			cb.openedAt = time.Now()
		}
		return err
	}

	cb.failureCount = 0
	cb.state = StateClosed
	return nil
}

func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// ExecuteWithRetry retries up to maxAttempts with exponential backoff and context cancellation awareness
func ExecuteWithRetry(ctx context.Context, maxAttempts int, initialBackoff time.Duration, fn func() error) error {
	backoff := initialBackoff
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// 1. Context Cancellation check before execution
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation aborted due to context cancellation: %w", ctx.Err())
		default:
		}

		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err

		// 2. Interruptible exponential backoff listening to ctx.Done()
		if attempt < maxAttempts {
			select {
			case <-ctx.Done():
				return fmt.Errorf("context cancelled during retry backoff: %w", ctx.Err())
			case <-time.After(backoff):
				backoff *= 2
			}
		}
	}
	return fmt.Errorf("failed after %d attempts: %w", maxAttempts, lastErr)
}
