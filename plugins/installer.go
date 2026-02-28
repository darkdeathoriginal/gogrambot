package plugins

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"strings"

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/darkdeathoriginal/gogrambot/handler"
	"github.com/darkdeathoriginal/gogrambot/helpers"
	"github.com/darkdeathoriginal/gogrambot/models"
)

func init() {
	// --- INSTALL COMMAND ---
	handler.NewPlugin("install").
		Description("Downloads and installs a plugin from a URL").
		Category("Owner").
		Handle(func(message *telegram.NewMessage) error {
			args := message.Args()
			if len(args) == 0 {
				message.Reply("Please provide a URL.")
				return nil
			}
			link := args
			log.Println("Installing plugin from URL:", args)
			//convert bytes to string
			// Validate URL
			u, err := url.Parse(link)
			if err != nil || u.Scheme == "" || u.Host == "" {

				message.Reply("Invalid URL.")
				return nil
			}

			// Handle Gist/GitHub Raw links (Ported from your Node.js code)
			switch u.Host {
			case "gist.github.com", "gist.githubusercontent.com":
				if !strings.HasSuffix(u.String(), "/raw") {
					link = u.String() + "/raw"
				}
			case "github.com":
				// Convert github blob links to raw
				link = strings.Replace(link, "github.com", "raw.githubusercontent.com", 1)
				link = strings.Replace(link, "/blob/", "/", 1)
			}
			var existing models.ExternalPlugin
			models.DB.Find(&existing, "url = ?", link)

			if existing.Name != "" {
				message.Reply("Plugin already installed.")
				return nil
			}
			pluginName, err := helpers.InstallExternalPlugin(link)
			if err != nil {
				message.Reply("Failed to install plugin: " + err.Error())
				return nil
			}
			// Save to DB
			newPlugin := models.ExternalPlugin{
				Name: pluginName,
				Url:  link,
			}
			result := models.DB.Create(&newPlugin)
			if result.Error != nil {
				message.Reply("Failed to save plugin to database: " + result.Error.Error())
				return nil
			}
			message.Reply(fmt.Sprintf("Installed **%s**. \n\n__⚠️ System Restart/Rebuild required to apply changes.__", pluginName), &telegram.SendOptions{
				ParseMode: telegram.MarkDown,
			})
			return nil
		})

	// --- LIST PLUGINS COMMAND ---
	handler.NewPlugin("plugin").
		Description("Lists installed plugins").
		Category("Owner").
		On("cmd:plugin").
		Handle(func(message *telegram.NewMessage) error {
			// Specific plugin lookup

			args := message.Args()
			var externalPlugins []models.ExternalPlugin
			result := models.DB.Find(&externalPlugins)
			if result.RowsAffected == 0 {
				message.Reply("No plugins installed.")
				return nil
			}
			if len(args) > 0 {
				query := args
				for _, p := range externalPlugins {
					if p.Name == query {
						message.Reply(fmt.Sprintf("**%s**\nUrl: %s", p.Name, p.Url), &telegram.SendOptions{
							ParseMode: telegram.MarkDown,
						})
						return nil
					}
				}
				message.Reply("Plugin not found.")
				return nil
			}

			// List all
			var msg strings.Builder
			msg.WriteString("**Installed Plugins:**\n\n")
			for _, p := range externalPlugins {
				msg.WriteString(fmt.Sprintf("• `%s`: %s\n", p.Name, p.Url))
			}
			message.Reply(msg.String(), &telegram.SendOptions{
				ParseMode: telegram.MarkDown,
			})
			return nil
		})

	// --- REMOVE COMMAND ---
	handler.NewPlugin("remove").
		Description("Removes a plugin").
		Category("Owner").
		On("cmd:remove").
		Handle(func(message *telegram.NewMessage) error {
			args := message.Args()
			if len(args) == 0 {
				message.Reply("Need plugin name.")
				return nil
			}
			target := args

			var existing models.ExternalPlugin
			result := models.DB.Find(&existing, "name = ?", target)
			if result.RowsAffected == 0 {
				message.Reply("Plugin not found in database.")
				return nil
			}

			// 2. Delete the file
			// We try to guess the filename. In Go, filename doesn't strictly have to match plugin name,
			// but we enforced it in the 'install' command.
			fileName := fmt.Sprintf("./plugins/%s.go", target)

			// Check if file exists
			if _, err := os.Stat(fileName); os.IsNotExist(err) {
				message.Reply("Plugin active, but source file not found at " + fileName)
				return nil
			}

			err := os.Remove(fileName)
			if err != nil {
				message.Reply("Failed to delete file: " + err.Error())
				return nil
			}
			// 3. Remove from DB
			result = models.DB.Delete(&models.ExternalPlugin{}, "name = ?", target)
			if result.Error != nil {
				message.Reply("Failed to remove plugin from database: " + result.Error.Error())
				return nil
			}

			message.Reply(fmt.Sprintf("**%s** removed successfully.\n\n__⚠️ Restart required to fully unload.__", target), &telegram.SendOptions{
				ParseMode: telegram.MarkDown,
			})
			return nil
		})
}
