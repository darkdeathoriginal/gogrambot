package plugins

import (
	"fmt"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/darkdeathoriginal/gogrambot/handler"
)

func init() {
	handler.NewPlugin("ban").
		Description("Bans the replied user").
		Category("Admin").
		Handle(func(message *telegram.NewMessage) error {
			if !message.IsReply() {
				message.Reply("Reply to a user to ban them.")
				return nil
			}
			replyMsg, err := message.GetReplyMessage()
			if err != nil {
				return err
			}

			_, err = message.Client.ChannelsEditBanned(
				&telegram.InputChannelObj{ChannelID: message.ChatID()},
				&telegram.InputPeerUser{UserID: replyMsg.SenderID()},
				&telegram.ChatBannedRights{ViewMessages: true},
			)
			if err != nil {
				message.Reply("Failed to ban. Am I an admin?")
				return err
			}
			message.Reply(fmt.Sprintf("User `%d` banned.", replyMsg.SenderID()))
			return nil
		})

	handler.NewPlugin("unban").
		Description("Unbans the replied user").
		Category("Admin").
		Handle(func(message *telegram.NewMessage) error {
			if !message.IsReply() {
				message.Reply("Reply to a user to unban them.")
				return nil
			}
			replyMsg, err := message.GetReplyMessage()
			if err != nil {
				return err
			}

			_, err = message.Client.ChannelsEditBanned(
				&telegram.InputChannelObj{ChannelID: message.ChatID()},
				&telegram.InputPeerUser{UserID: replyMsg.SenderID()},
				&telegram.ChatBannedRights{},
			)
			if err != nil {
				message.Reply("Failed to unban. Am I an admin?")
				return err
			}
			message.Reply(fmt.Sprintf("User `%d` unbanned.", replyMsg.SenderID()))
			return nil
		})

	handler.NewPlugin("mute").
		Description("Mutes the replied user").
		Category("Admin").
		Handle(func(message *telegram.NewMessage) error {
			if !message.IsReply() {
				message.Reply("Reply to a user to mute them.")
				return nil
			}
			replyMsg, err := message.GetReplyMessage()
			if err != nil {
				return err
			}

			_, err = message.Client.ChannelsEditBanned(
				&telegram.InputChannelObj{ChannelID: message.ChatID()},
				&telegram.InputPeerUser{UserID: replyMsg.SenderID()},
				&telegram.ChatBannedRights{SendMessages: true},
			)
			if err != nil {
				message.Reply("Failed to mute. Am I an admin?")
				return err
			}
			message.Reply(fmt.Sprintf("User `%d` muted.", replyMsg.SenderID()))
			return nil
		})

	handler.NewPlugin("unmute").
		Description("Unmutes the replied user").
		Category("Admin").
		Handle(func(message *telegram.NewMessage) error {
			if !message.IsReply() {
				message.Reply("Reply to a user to unmute them.")
				return nil
			}
			replyMsg, err := message.GetReplyMessage()
			if err != nil {
				return err
			}

			_, err = message.Client.ChannelsEditBanned(
				&telegram.InputChannelObj{ChannelID: message.ChatID()},
				&telegram.InputPeerUser{UserID: replyMsg.SenderID()},
				&telegram.ChatBannedRights{},
			)
			if err != nil {
				message.Reply("Failed to unmute. Am I an admin?")
				return err
			}
			message.Reply(fmt.Sprintf("User `%d` unmuted.", replyMsg.SenderID()))
			return nil
		})

	handler.NewPlugin("pin").
		Description("Pins the replied message").
		Category("Admin").
		Handle(func(message *telegram.NewMessage) error {
			if !message.IsReply() {
				message.Reply("Reply to a message to pin it.")
				return nil
			}

			replyMsg, err := message.GetReplyMessage()
			if err != nil {
				return err
			}

			loud := true
			if strings.Contains(strings.ToLower(message.Args()), "silent") {
				loud = false
			}

			_, err = message.Client.MessagesUpdatePinnedMessage(&telegram.MessagesUpdatePinnedMessageParams{
				Peer:   &telegram.InputPeerChannel{ChannelID: message.ChatID()},
				ID:     replyMsg.ID,
				Silent: !loud,
			})
			if err != nil {
				message.Reply("Failed to pin. Am I an admin?")
				return err
			}
			message.Reply("Message pinned successfully.")
			return nil
		})
}
