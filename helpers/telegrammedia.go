package helpers

import (
	"context"
	"fmt"
	"log"

	"github.com/amarnathcjd/gogram/telegram"
)

// SendMediaWithRefresh refetches a queued message once if its file reference
// expired. Fetching and sending are separate paced operations, never nested.
func SendMediaWithRefresh(ctx context.Context, c *telegram.Client, destination any, message *telegram.NewMessage) (*telegram.NewMessage, error) {
	return sendWithFileReferenceRefresh(message, func(m *telegram.NewMessage) (*telegram.NewMessage, error) {
		var sent *telegram.NewMessage
		err := TelegramRequests.Do(ctx, TelegramSendInterval, func() error {
			var err error
			sent, err = c.SendMessage(destination, m)
			return err
		})
		return sent, err
	}, func() (*telegram.NewMessage, error) {
		var messages []telegram.NewMessage
		err := TelegramRequests.Do(ctx, TelegramHistoryInterval, func() error {
			var err error
			messages, err = c.GetMessages(message.Peer, &telegram.SearchOption{IDs: []int32{message.ID}, Context: ctx})
			return err
		})
		if err != nil {
			return nil, err
		}
		for i := range messages {
			if messages[i].ID == message.ID {
				return &messages[i], nil
			}
		}
		return nil, fmt.Errorf("source message %d is no longer available", message.ID)
	})
}

func sendWithFileReferenceRefresh(message *telegram.NewMessage, send func(*telegram.NewMessage) (*telegram.NewMessage, error), refresh func() (*telegram.NewMessage, error)) (*telegram.NewMessage, error) {
	if message == nil || message.Document() == nil {
		return nil, fmt.Errorf("cache source has no document")
	}
	sent, err := send(message)
	if !telegram.MatchError(err, "FILE_REFERENCE_EXPIRED") {
		return sent, err
	}
	log.Printf("[filecache] Refreshing expired file reference for source message %d", message.ID)
	fresh, err := refresh()
	if err != nil {
		return nil, fmt.Errorf("refresh file reference: %w", err)
	}
	if fresh == nil || fresh.ID != message.ID || fresh.Document() == nil || fresh.Document().ID != message.Document().ID {
		return nil, fmt.Errorf("source message %d no longer contains the original document", message.ID)
	}
	return send(fresh)
}
