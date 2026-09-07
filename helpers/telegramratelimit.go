package helpers

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

// Conservative application defaults, not guaranteed Telegram quotas. All bulk
// operations share one lane so their individual delays cannot add up to a burst.
const (
	TelegramHistoryInterval = time.Second
	TelegramSendInterval    = 10 * time.Second
	TelegramJoinInterval    = 30 * time.Second
)

var TelegramRequests = newTelegramRequestLimiter()

type telegramRequestLimiter struct {
	lane       chan struct{}
	mu         sync.Mutex
	next       time.Time
	floodUntil time.Time
	now        func() time.Time
	sleep      func(context.Context, time.Duration) error
}

func newTelegramRequestLimiter() *telegramRequestLimiter {
	return &telegramRequestLimiter{
		lane: make(chan struct{}, 1),
		now:  time.Now,
		sleep: func(ctx context.Context, delay time.Duration) error {
			timer := time.NewTimer(delay)
			defer timer.Stop()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				return nil
			}
		},
	}
}

// Do serializes bulk operations and spaces attempts (including failed ones).
// request must not call Do recursively. The flood callback uses a separate lock
// so gogram can invoke it while an operation owns the lane.
func (l *telegramRequestLimiter) Do(ctx context.Context, interval time.Duration, request func() error) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case l.lane <- struct{}{}:
	}
	defer func() { <-l.lane }()
	if err := l.wait(ctx, true); err != nil {
		return err
	}
	defer func() {
		l.mu.Lock()
		l.next = l.now().Add(interval)
		l.mu.Unlock()
	}()
	return request()
}

// HandleFlood is the ClientConfig.FloodHandler. In the pinned gogram version,
// true immediately retries the RPC; the callback itself must perform the wait.
func (l *telegramRequestLimiter) HandleFlood(err error) bool {
	seconds := telegram.GetFloodWait(err)
	if seconds <= 0 {
		log.Printf("Telegram flood error without a usable wait; not retrying: %v", err)
		return false
	}
	delay := time.Duration(seconds)*time.Second + time.Second
	l.mu.Lock()
	until := l.now().Add(delay)
	if until.After(l.floodUntil) {
		l.floodUntil = until
	}
	l.mu.Unlock()
	log.Printf("Telegram flood wait: pausing bulk requests for at least %s: %v", delay, err)
	return l.wait(context.Background(), false) == nil
}

func (l *telegramRequestLimiter) wait(ctx context.Context, includePacing bool) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		l.mu.Lock()
		until := l.floodUntil
		if includePacing && l.next.After(until) {
			until = l.next
		}
		delay := until.Sub(l.now())
		l.mu.Unlock()
		if delay <= 0 {
			return nil
		}
		if err := l.sleep(ctx, delay); err != nil {
			return err
		}
		// Another RPC may have extended the account cooldown while we slept.
	}
}
