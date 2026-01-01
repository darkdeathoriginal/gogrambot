package plugins

import (
	"os"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/darkdeathoriginal/gogrambot/handler"
)

func init() {
	handler.NewPlugin("restart").
		Description("Restarts the bot and recompiles plugins").
		Category("Owner").
		Handle(func(message *telegram.NewMessage) error {
			message.Reply("Bot is restarting and recompiling... check back in 5 seconds!")

			time.Sleep(1 * time.Second)

			os.Exit(0)
			return nil
		})
}
