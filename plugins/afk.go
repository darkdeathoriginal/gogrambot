package plugins

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/darkdeathoriginal/gogrambot/handler"
)

var (
	isAFK     bool
	afkReason string
	afkTime   time.Time
	afkMutex  sync.Mutex
)

func init() {
	handler.NewPlugin("afk").
		Description("Sets your status to AFK").
		Category("Userbot").
		Handle(func(message *telegram.NewMessage) error {
			afkMutex.Lock()
			defer afkMutex.Unlock()

			if isAFK {
				isAFK = false
				afkReason = ""
				message.Reply("Welcome back! You are no longer AFK.")
				return nil
			}

			if message.Args() != "" {
				afkReason = message.Args()
			} else {
				afkReason = "Not specified"
			}
			isAFK = true
			afkTime = time.Now()

			message.Reply(fmt.Sprintf("You are now AFK.\nReason: `%s`", afkReason))
			return nil
		})

	// Add a handler that listens to all incoming messages to auto-reply and auto-disable
	handler.NewPlugin("afk_listener").
		AllowAll(true).
		On("message").
		Handle(func(message *telegram.NewMessage) error {
			afkMutex.Lock()
			defer afkMutex.Unlock()

			// If we are not AFK, do nothing
			if !isAFK {
				return nil
			}

			// If the message is from US (Out = true), it means we did something. Remove AFK.
			// However, ignore the `/afk` command itself so we don't immediately unset it.
			if message.Message != nil && message.Message.Out {
				if !strings.HasPrefix(message.Text(), "/afk") {
					isAFK = false
					afkReason = ""
					message.Reply("Welcome back! I've removed your AFK status.")
				}
				return nil
			}

			// Check if we should reply (they mentioned us or it's a PM)
			shouldReply := false
			if message.ChatID() == message.SenderID() {
				shouldReply = true
			} else if message.Message != nil && message.Message.Mentioned {
				shouldReply = true
			}

			// Do not reply to bots
			if sender, err := message.GetSender(); err == nil && sender != nil && sender.Bot {
				return nil
			}

			if shouldReply {
				duration := time.Since(afkTime).Round(time.Second)
				msg := fmt.Sprintf("I am currently **AFK**.\nReason: `%s`\nSince: `%s`", afkReason, duration.String())
				message.Reply(msg)
			}
			return nil
		})
}
