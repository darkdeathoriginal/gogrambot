package plugins

import (
	"strconv"

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/darkdeathoriginal/gogrambot/handler"
)

func init() {
	handler.NewPlugin("id").
		Description("Responds with the user's ID").
		Category("Utility").
		Handle(func(message *telegram.NewMessage) error {
			if message.IsReply() {
				replyMsg, err := message.GetReplyMessage()
				if err != nil {
					return err
				}
				message.Reply("The ID is: " + strconv.FormatInt(replyMsg.SenderID(), 10))
				return nil
			}
			message.Reply("Your ID is: " + strconv.FormatInt(message.ChatID(), 10))
			return nil
		})

}
