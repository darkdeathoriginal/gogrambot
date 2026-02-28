package plugins

import (
	"fmt"

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/darkdeathoriginal/gogrambot/handler"
)

func init() {
	handler.NewPlugin("purge").
		Description("Deletes messages from the replied message to the current one").
		Category("Userbot").
		Handle(func(message *telegram.NewMessage) error {
			if !message.IsReply() {
				message.Reply("Reply to a message to purge from there.")
				return nil
			}

			replyTo, err := message.GetReplyMessage()
			if err != nil {
				return err
			}

			startID := replyTo.ID
			endID := message.ID

			var idsToDelete []int32
			for i := startID; i <= endID; i++ {
				idsToDelete = append(idsToDelete, i)
			}

			if len(idsToDelete) == 0 {
				return nil
			}

			// Telegram API usually allows deleting 100 messages at once
			chunkSize := 100
			deletedCount := 0

			for i := 0; i < len(idsToDelete); i += chunkSize {
				end := i + chunkSize
				if end > len(idsToDelete) {
					end = len(idsToDelete)
				}

				chunk := idsToDelete[i:end]
				_, err := message.Client.DeleteMessages(message.ChatID(), chunk)
				if err != nil {
					// Some messages might be already deleted or we don't have permission
					// Log err or continue
				} else {
					deletedCount += len(chunk)
				}
			}

			respMsg, _ := message.Client.SendMessage(message.ChatID(), fmt.Sprintf("Purged **%d** messages.", deletedCount))

			// Optional: delete the success message after a few seconds
			// time.Sleep(3 * time.Second)
			// message.Client.DeleteMessages(message.ChatID(), []int32{respMsg.ID})
			_ = respMsg

			return nil
		})

	handler.NewPlugin("del").
		Description("Deletes the replied message").
		Category("Userbot").
		Handle(func(message *telegram.NewMessage) error {
			if !message.IsReply() {
				message.Reply("Reply to a message to delete.")
				return nil
			}

			replyMsg, err := message.GetReplyMessage()
			if err != nil {
				return err
			}

			// Delete both the command and the replied message
			_, err = message.Client.DeleteMessages(message.ChatID(), []int32{replyMsg.ID, message.ID})
			return err
		})
}
