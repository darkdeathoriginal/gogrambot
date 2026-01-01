package plugins

import (
	"fmt"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/darkdeathoriginal/gogrambot/handler"
)

func init() {
	handler.NewPlugin("ping").
		Description("Responds with 'Pong!' when you send /ping").
		Category("Utility").
		Handle(func(message *telegram.NewMessage) error {
			startTime := time.Now()
			msg, err := message.Reply("Pinging...")
			endTime := time.Now()
			if err != nil {
				return err
			}
			latency := endTime.Sub(startTime).Milliseconds()
			_, err = msg.Edit(fmt.Sprintf("Latency: %d ms", latency))
			return err
		})
}
