package plugins

import (
	"fmt"

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/darkdeathoriginal/gogrambot/handler"
)

func init() {
	handler.NewPlugin("id").
		Description("Gets current Chat ID or replied User ID").
		Category("Userbot").
		Handle(func(message *telegram.NewMessage) error {
			if !message.IsReply() {
				// Return the current Chat ID
				message.Reply(fmt.Sprintf("Chat ID: `%d`", message.ChatID()))
				return nil
			}

			replyMsg, err := message.GetReplyMessage()
			if err != nil {
				return err
			}

			// Return the replied user's ID
			message.Reply(fmt.Sprintf("Chat ID: `%d`\nUser ID: `%d`", message.ChatID(), replyMsg.SenderID()))
			return nil
		})
}
