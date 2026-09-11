package helpers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/amarnathcjd/gogram"
)

// Conservative application defaults, not guaranteed Telegram quotas. All bulk
// operations share one lane so their individual delays cannot add up to a burst.
const (
	TelegramHistoryInterval = time.Second
	TelegramSendInterval    = 10 * time.Second
	TelegramJoinInterval    = 30 * time.Second
	telegramFloodRetries    = 3
)

var TelegramRequests = newTelegramRequestLimiter()

type telegramRequestLimiter struct {
	lane  chan struct{}
	next  time.Time // protected by lane; never held during a flood wait
	now   func() time.Time
	sleep func(context.Context, time.Duration) error
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

// Do paces individual attempts. Only the rejected operation waits on a flood;
// unrelated workers can continue, and cancellation can interrupt that wait.
// request must not call Do recursively.
func (l *telegramRequestLimiter) Do(ctx context.Context, interval time.Duration, request func() error) error {
	return l.retryFlood(ctx, "bulk request", func() error {
		return l.attempt(ctx, interval, request)
	})
}

func (l *telegramRequestLimiter) attempt(ctx context.Context, interval time.Duration, request func() error) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case l.lane <- struct{}{}:
	}
	defer func() { <-l.lane }()
	if err := ctx.Err(); err != nil {
		return err
	}
	if delay := l.next.Sub(l.now()); delay > 0 {
		if err := l.sleep(ctx, delay); err != nil {
			return err
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	defer func() {
		l.next = l.now().Add(interval)
	}()
	return request()
}

// Never sleep or recursively retry inside gogram's global RPC callback. The
// caller must receive the error so startup and workers retain control.
func (l *telegramRequestLimiter) HandleFlood(error) bool {
	return false
}

// RetryTelegramFlood is for startup requests that must not wait for bulk pacing.
func RetryTelegramFlood(ctx context.Context, operation string, request func() error) error {
	return TelegramRequests.retryFlood(ctx, operation, request)
}

func (l *telegramRequestLimiter) retryFlood(ctx context.Context, operation string, request func() error) error {
	for retries := 0; ; retries++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := request()
		delay, flood := telegramFloodDelay(err)
		if !flood {
			return err
		}
		if retries >= telegramFloodRetries {
			return fmt.Errorf("%s: Telegram still reports a flood wait after %d retries: %w", operation, retries, err)
		}
		log.Printf("%s: Telegram rejected this operation; retry %d/%d at %s (wait %s): %v",
			operation, retries+1, telegramFloodRetries, l.now().Add(delay).Format(time.RFC3339), delay, err)
		if err := l.sleep(ctx, delay); err != nil {
			return err
		}
		log.Printf("%s: flood wait finished; retrying now", operation)
	}
}

var telegramFloodPattern = regexp.MustCompile(`(?:^|[\s\[])FLOOD_(?:PREMIUM_)?WAIT_([0-9]+)(?:$|[\s\]])`)

func telegramFloodDelay(err error) (time.Duration, bool) {
	if err == nil {
		return 0, false
	}
	var seconds int64
	var rpc *gogram.ErrResponseCode
	if errors.As(err, &rpc) {
		if rpc.Code != 420 || (rpc.Message != "FLOOD_WAIT_X" && rpc.Message != "FLOOD_PREMIUM_WAIT_X") {
			return 0, false
		}
		value, ok := rpc.AdditionalInfo.(int)
		if !ok {
			return 0, false
		}
		seconds = int64(value)
	} else {
		// Support wrapped textual RPC errors, but never interpret a generic
		// "Please wait" message or the local pacing delay as a Telegram limit.
		match := telegramFloodPattern.FindStringSubmatch(err.Error())
		if match == nil {
			return 0, false
		}
		var parseErr error
		seconds, parseErr = strconv.ParseInt(match[1], 10, 64)
		if parseErr != nil {
			return 0, false
		}
	}
	if seconds < 0 || seconds > int64((time.Duration(1<<63-1)-time.Second)/time.Second) {
		return 0, false
	}
	return time.Duration(seconds)*time.Second + time.Second, true
}
