package plugins

import (
	"fmt"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/darkdeathoriginal/gogrambot/handler"
)

func init() {
	handler.NewPlugin("ping").
		Description("Responds with bot latency").
		Category("Userbot").
		Handle(func(message *telegram.NewMessage) error {
			startTime := time.Now()
			msg, err := message.Reply("...Pinging...")
			if err != nil {
				return err
			}
			latency := time.Since(startTime).Milliseconds()
			_, err = msg.Edit(fmt.Sprintf("🏓 **Pong!**\nLatency: `%dms`", latency))
			return err
		})
}
