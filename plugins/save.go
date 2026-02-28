package plugins

import (
	"github.com/amarnathcjd/gogram/telegram"
	"github.com/darkdeathoriginal/gogrambot/handler"
)

func init() {
	handler.NewPlugin("save").
		Description("Forwards the replied message to Saved Messages").
		Category("Userbot").
		Handle(func(message *telegram.NewMessage) error {
			if !message.IsReply() {
				message.Reply("Reply to a message to save it.")
				return nil
			}

			replyTo, err := message.GetReplyMessage()
			if err != nil {
				return err
			}

			me, err := message.Client.GetMe()
			if err != nil {
				return err
			}

			_, err = message.Client.Forward(me.ID, message.ChatID(), []int32{replyTo.ID})
			if err != nil {
				message.Reply("Failed to save message.")
				return err
			}

			// Optionally delete the trigger command
			_, _ = message.Client.DeleteMessages(message.ChatID(), []int32{message.ID})
			return nil
		})
}
