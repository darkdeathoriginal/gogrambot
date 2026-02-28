package plugins

import (
	"fmt"

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/darkdeathoriginal/gogrambot/handler"
)

func init() {
	handler.NewPlugin("whois").
		Description("Fetches detailed user information").
		Category("Userbot").
		Handle(func(message *telegram.NewMessage) error {
			var userID int64

			args := message.Args()
			if message.IsReply() {
				replyMsg, err := message.GetReplyMessage()
				if err == nil {
					userID = replyMsg.SenderID()
				}
			} else if args != "" {
				// We don't have a resolve username method directly exposed in the same easy way
				// but we can try to resolve it if it's a number
				fmt.Sscanf(args, "%d", &userID)
			} else {
				userID = message.SenderID()
			}

			if userID == 0 {
				message.Reply("Please reply to a user or provide a User ID to use /whois.")
				return nil
			}

			// We use GetUser to get more info
			user, err := message.Client.UsersGetUsers([]telegram.InputUser{
				&telegram.InputUserObj{
					UserID: userID,
				},
			})

			if err != nil || len(user) == 0 {
				message.Reply("Could not fetch user information. Make sure the ID is correct.")
				return err
			}

			u, ok := user[0].(*telegram.UserObj)
			if !ok {
				message.Reply("Could not assert user information.")
				return nil
			}

			var info string
			info += "👤 **User Info:**\n"
			info += fmt.Sprintf("├ **ID:** `%d`\n", u.ID)

			// Additional Info if available might be in u.UserObj fields if needed but gogram user struct is missing some fields from tl
			// We skip the Missing fields for brevity
			info += fmt.Sprintf("├ **Is Bot:** %t\n", u.Bot)
			info += fmt.Sprintf("├ **Is Premium:** %t\n", u.Premium)
			info += fmt.Sprintf("├ **Is Verified:** %t\n", u.Verified)
			info += fmt.Sprintf("╰ **Is Restricted:** %t\n", u.Restricted)

			message.Reply(info, &telegram.SendOptions{
				ParseMode: telegram.MarkDown,
			})

			return nil
		})
}
