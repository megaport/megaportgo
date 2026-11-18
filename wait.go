package megaport

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const (
	defaultWaitTimeout  = 5 * time.Minute
	defaultPollInterval = 30 * time.Second
)

// waitUntilReady calls ready every 30 seconds until it reports true or fails, the timeout passes or ctx ends.
// A zero timeout defaults to 5 minutes. what names the wait in the errors, such as "Port <uid> to provision".
func waitUntilReady(ctx context.Context, timeout time.Duration, what string, ready func() (bool, error)) error {
	if timeout == 0 {
		timeout = defaultWaitTimeout
	}

	ticker := time.NewTicker(defaultPollInterval)
	defer ticker.Stop()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			// the last tick can land together with the timer, so check once more before giving up.
			// a failed check here still counts as time expired.
			if done, err := ready(); err == nil && done {
				return nil
			}
			return fmt.Errorf("time expired waiting for %s", what)
		case <-ctx.Done():
			return &waitContextError{what: what, err: ctx.Err()}
		case <-ticker.C:
			if done, err := ready(); err != nil || done {
				return err
			}
		}
	}
}

// waitContextError is the error of a wait that ended because its context did.
// It does not unwrap to the context's error: most callers have never exposed it, so errors.Is must keep reporting false.
type waitContextError struct {
	what string
	err  error
}

func (e *waitContextError) Error() string {
	return "context expired waiting for " + e.what
}

// withContextCause adds the context's error to the message of a wait that ended because its context did.
// Any other error is returned as it is.
func withContextCause(err error) error {
	var ctxErr *waitContextError
	if errors.As(err, &ctxErr) {
		return fmt.Errorf("%s: %w", ctxErr.Error(), ctxErr.err)
	}
	return err
}
