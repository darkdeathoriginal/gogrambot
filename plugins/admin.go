package plugins

import (
	"fmt"
	"log"
	"os"
	"os/exec"
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

	handler.NewPlugin("update").
		Description("Updates the bot to the latest version").
		Category("Admin").
		Handle(updateCommand)
}

func updateCommand(message *telegram.NewMessage) error {

	message.Reply("🛠 Update started. Pulling latest changes and building...")

	go func() {
		// Get latest changes without modifying the working tree
		fetchCmd := exec.Command("git", "fetch", "origin")
		if err := fetchCmd.Run(); err != nil {
			log.Printf("git fetch failed: %v", err)
			return
		}

		// Show commits that will be pulled
		logCmd := exec.Command("git", "log", "--oneline", "HEAD..origin/HEAD")
		output, err := logCmd.Output()
		if err != nil {
			log.Printf("git log failed: %v", err)
			return
		}

		changes := string(output)
		if changes == "" {
			changes = "No new commits."
		}

		_, _ = message.Client.SendMessage(
			message.Chat.ID,
			fmt.Sprintf("📥 **Incoming commits:**\n```\n%s\n```", changes),
			&telegram.SendOptions{ParseMode: "markdown"},
		)

		// Pull and build
		buildCmd := exec.Command("sh", "-lc", "git pull && go build -v")
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr

		if err := buildCmd.Run(); err != nil {
			log.Printf("build command failed: %v", err)
			_, _ = message.Client.SendMessage(
				message.Chat.ID,
				"❌ Build command failed. Check server logs for details.",
				&telegram.SendOptions{ParseMode: "markdown"},
			)
			return
		}
		//kill the current process to allow the new build to take over
		log.Println("Build successful. Restarting bot...")
		_, _ = message.Client.SendMessage(message.Chat.ID, "✅ Update complete. Restarting bot...", &telegram.SendOptions{ParseMode: "markdown"})
		os.Exit(0)
	}()

	return nil
}
