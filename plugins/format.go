package plugins

import (
	"fmt"

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/darkdeathoriginal/gogrambot/handler"
)

func init() {
	formats := map[string]string{
		"bold":   "**",
		"italic": "__",
		"strike": "~~",
		"mono":   "`",
	}

	for cmdName, affix := range formats {
		cmdAffix := affix // capture for closure

		handler.NewPlugin(cmdName).
			Description(fmt.Sprintf("Formats text as %s", cmdName)).
			Category("Userbot").
			Handle(func(message *telegram.NewMessage) error {
				args := message.Args()

				if args == "" && !message.IsReply() {
					message.Reply(fmt.Sprintf("Provide text or reply to a message to format it as %s.", cmdName))
					return nil
				}

				textToFormat := args
				if message.IsReply() && args == "" {
					replyMsg, err := message.GetReplyMessage()
					if err == nil && replyMsg != nil && replyMsg.Message != nil {
						textToFormat = replyMsg.Message.Message
					}
				}

				if textToFormat == "" {
					return nil
				}

				formattedText := fmt.Sprintf("%s%s%s", cmdAffix, textToFormat, cmdAffix)

				isEdit := false
				if !message.IsReply() && args != "" {
					// Try to edit if it's our own message
					_, err := message.Edit(formattedText)
					if err == nil {
						isEdit = true
					}
				}

				if !isEdit {
					message.Reply(formattedText)
					if args != "" { // If we sent the text, we can delete the command msg
						message.Client.DeleteMessages(message.ChatID(), []int32{message.ID})
					}
				}

				return nil
			})
	}
}
