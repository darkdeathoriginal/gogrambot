package helpers

import (
	"context"
	"fmt"

	"github.com/amarnathcjd/gogram/telegram"
)

// IterMessagesReverse fetches messages from oldest to newest.
func IterMessagesReverse(c *telegram.Client, chatID any, callback func(*telegram.NewMessage) error) error {
	return IterMessagesReverseFrom(c, chatID, 1, callback)
}

// IterMessagesReverseFrom fetches messages from oldest to newest starting after the given message ID.
func IterMessagesReverseFrom(c *telegram.Client, chatID any, startAfterID int32, callback func(*telegram.NewMessage) error) error {
	var peer telegram.InputPeer
	err := TelegramRequests.Do(context.Background(), TelegramHistoryInterval, func() error {
		var err error
		peer, err = c.ResolvePeer(chatID)
		return err
	})
	if err != nil {
		return err
	}

	var offsetId int32 = startAfterID
	if offsetId < 1 {
		offsetId = 1
	}
	var limit int32 = 100 // API limit for messages.getHistory is up to 100

	for {
		// Using AddOffset = -limit with OffsetID fetches messages forward in time
		var history telegram.MessagesMessages
		err := TelegramRequests.Do(context.Background(), TelegramHistoryInterval, func() error {
			var err error
			history, err = c.MessagesGetHistory(&telegram.MessagesGetHistoryParams{
				Peer:      peer,
				OffsetID:  offsetId,
				AddOffset: -limit,
				Limit:     limit,
			})
			return err
		})
		if err != nil {
			return fmt.Errorf("fetch message history after %d: %w", offsetId, err)
		}

		var rawMessages []telegram.Message
		var users []telegram.User
		var chats []telegram.Chat

		// Extract raw data depending on the history object type
		switch r := history.(type) {
		case *telegram.MessagesChannelMessages:
			rawMessages = r.Messages
			users = r.Users
			chats = r.Chats
		case *telegram.MessagesMessagesObj:
			rawMessages = r.Messages
			users = r.Users
			chats = r.Chats
		case *telegram.MessagesMessagesSlice:
			rawMessages = r.Messages
			users = r.Users
			chats = r.Chats
		}

		if len(rawMessages) == 0 {
			break // No more messages found
		}

		// Update gogram's internal cache for Users and Chats
		c.Cache.UpdatePeersToCache(users, chats)

		// Pack the raw MTProto messages into the high-level *NewMessage struct
		messages := telegram.PackMessages(c, rawMessages)

		// The API returns the fetched chunk in newest-to-oldest order.
		// So we must iterate backwards through this specific batch to process them oldest-to-newest.
		previousOffset := offsetId
		for i := len(messages) - 1; i >= 0; i-- {
			msg := messages[i]

			if err := callback(msg); err != nil {
				return err
			}

			// Update the offsetId to the highest ID we've processed
			if msg.ID > offsetId {
				offsetId = msg.ID
			}
		}

		// Stop at the end or if Telegram repeats a page without advancing.
		if len(messages) < int(limit) || offsetId <= previousOffset {
			break
		}
	}

	return nil
}
