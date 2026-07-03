package helpers

import (
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

// IterMessagesReverse fetches messages from oldest to newest.
func IterMessagesReverse(c *telegram.Client, chatID any, callback func(*telegram.NewMessage) error) error {
	peer, err := c.ResolvePeer(chatID)
	if err != nil {
		return err
	}

	var offsetId int32 = 1 // Start from the oldest possible message ID
	var limit int32 = 100  // API limit for messages.getHistory is up to 100

	for {
		// Using AddOffset = -limit with OffsetID fetches messages forward in time
		history, _ := c.MessagesGetHistory(&telegram.MessagesGetHistoryParams{
			Peer:      peer,
			OffsetID:  offsetId,
			AddOffset: -limit,
			Limit:     limit,
		})

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

		// If the amount of returned messages is less than the limit, we've reached the end
		if len(messages) < int(limit) {
			break
		}

		// Small sleep to prevent hitting FloodWait too aggressively
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}
