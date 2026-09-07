package helpers

import (
	"context"
	"errors"
	"testing"
	"time"
)

func fakeTelegramLimiter() (*telegramRequestLimiter, *time.Time) {
	l := newTelegramRequestLimiter()
	now := time.Unix(1000, 0)
	l.now = func() time.Time { return now }
	l.sleep = func(ctx context.Context, delay time.Duration) error {
		now = now.Add(delay)
		return ctx.Err()
	}
	return l, &now
}

func TestTelegramBulkOperationsSharePacingIncludingFailures(t *testing.T) {
	l, now := fakeTelegramLimiter()
	start := *now
	failure := errors.New("request failed")
	for i, tc := range []struct {
		interval time.Duration
		at       time.Duration
		err      error
	}{
		{TelegramSendInterval, 0, nil},
		{TelegramJoinInterval, 10 * time.Second, failure},
		{TelegramHistoryInterval, 40 * time.Second, nil},
		{TelegramSendInterval, 41 * time.Second, nil},
	} {
		err := l.Do(context.Background(), tc.interval, func() error {
			if got := now.Sub(start); got != tc.at {
				t.Errorf("operation %d started at %s, want %s", i, got, tc.at)
			}
			return tc.err
		})
		if !errors.Is(err, tc.err) {
			t.Fatalf("operation %d: error = %v, want %v", i, err, tc.err)
		}
	}
}

func TestTelegramFloodHandlerWaitsBeforeRetryInsideOperation(t *testing.T) {
	for _, message := range []string{"FLOOD_WAIT_12", "FLOOD_PREMIUM_WAIT_12"} {
		t.Run(message, func(t *testing.T) {
			l, now := fakeTelegramLimiter()
			start := *now
			err := l.Do(context.Background(), TelegramSendInterval, func() error {
				if !l.HandleFlood(errors.New(message)) {
					t.Fatal("valid flood wait was not retried")
				}
				if got := now.Sub(start); got != 13*time.Second {
					t.Fatalf("retry after %s, want 13s", got)
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if err := l.Do(context.Background(), TelegramHistoryInterval, func() error {
				if got := now.Sub(start); got != 23*time.Second {
					t.Fatalf("next worker started after %s, want 23s", got)
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestTelegramFloodHandlerRejectsUnknownWait(t *testing.T) {
	l, _ := fakeTelegramLimiter()
	l.sleep = func(context.Context, time.Duration) error {
		t.Fatal("unexpected sleep")
		return nil
	}
	for _, err := range []error{nil, errors.New("PEER_FLOOD"), errors.New("FLOOD_WAIT_0")} {
		if l.HandleFlood(err) {
			t.Fatalf("unexpected retry for %v", err)
		}
	}
}

func TestTelegramCooldownExtensionAndShorterWait(t *testing.T) {
	l, now := fakeTelegramLimiter()
	start := *now
	l.floodUntil = start.Add(20 * time.Second)
	sleeps := 0
	l.sleep = func(_ context.Context, delay time.Duration) error {
		sleeps++
		*now = now.Add(delay)
		if sleeps == 1 {
			// Model another RPC extending the cooldown during this sleep.
			l.mu.Lock()
			l.floodUntil = now.Add(5 * time.Second)
			l.mu.Unlock()
		}
		return nil
	}
	if !l.HandleFlood(errors.New("FLOOD_WAIT_2")) {
		t.Fatal("expected retry")
	}
	if got := now.Sub(start); got != 25*time.Second || sleeps != 2 {
		t.Fatalf("wait = %s in %d sleeps, want 25s in 2", got, sleeps)
	}
}

func TestTelegramWaitingWorkerCanCancel(t *testing.T) {
	l := newTelegramRequestLimiter()
	started := make(chan struct{})
	release := make(chan struct{})
	finished := make(chan error, 1)
	go func() {
		finished <- l.Do(context.Background(), time.Hour, func() error {
			close(started)
			<-release
			return nil
		})
	}()
	<-started
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	err := l.Do(ctx, 0, func() error {
		t.Error("second worker ran while first owned the lane")
		return nil
	})
	close(release)
	if firstErr := <-finished; firstErr != nil {
		t.Fatal(firstErr)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("queued request error = %v, want deadline exceeded", err)
	}
	// Cancellation during pacing must also prevent the next request.
	ctx, cancelPacing := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancelPacing()
	err = l.Do(ctx, 0, func() error {
		t.Error("request ran before pacing delay expired")
		return nil
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("pacing error = %v, want deadline exceeded", err)
	}
}
