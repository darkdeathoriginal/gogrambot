package plugins

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"

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
			args := strings.TrimSpace(message.Args())
			if args == "" {
				message.Reply("Please provide a URL.")
				return nil
			}
			link := args
			pluginName, err := installPlugin(link)
			if err != nil {
				if pluginName != "" {
					removePlugin(pluginName)
				}
				message.Reply("Failed to install plugin: " + err.Error())
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

			args := strings.TrimSpace(message.Args())
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
			args := strings.TrimSpace(message.Args())
			if args == "" {
				message.Reply("Need plugin name.")
				return nil
			}
			target := args

			err := removePlugin(target)
			if err != nil {
				message.Reply("Failed to remove plugin: " + err.Error())
				return nil
			}

			message.Reply(fmt.Sprintf("**%s** removed successfully.\n\n__⚠️ Restart required to fully unload.__", target), &telegram.SendOptions{
				ParseMode: telegram.MarkDown,
			})
			return nil
		})

	handler.NewPlugin("pupdate").
		Description("Updates a plugin").
		Category("Owner").
		Handle(func(message *telegram.NewMessage) error {
			msg, err := message.Reply("Updating plugin...")
			if err != nil {
				return err
			}
			args := strings.TrimSpace(message.Args())
			if args == "" {
				msg.Edit("Need plugin name.")
				return nil
			}
			target := args
			var existing models.ExternalPlugin
			result := models.DB.Find(&existing, "name = ?", target)
			if result.RowsAffected == 0 {
				return fmt.Errorf("plugin not found in database")
			}

			err = removePlugin(target)
			if err != nil {
				msg.Edit("Failed to update plugin: " + err.Error())
				return nil
			}
			// Reinstall
			pluginName, err := installPlugin(existing.Url)
			if err != nil {
				if pluginName != "" {
					removePlugin(pluginName)
				}
				msg.Edit("Failed to update plugin: " + err.Error())
				return nil
			}
			msg.Edit(fmt.Sprintf("**%s** updated successfully.\n\n__⚠️ Restarting to apply changes.__", pluginName), &telegram.SendOptions{
				ParseMode: telegram.MarkDown,
			})
			os.Exit(0)
			return nil
		})
}

func removePlugin(target string) error {
	var existing models.ExternalPlugin
	result := models.DB.Find(&existing, "name = ?", target)
	if result.RowsAffected == 0 {
		return fmt.Errorf("plugin not found in database")
	}

	// 2. Delete the file
	// We try to guess the filename. In Go, filename doesn't strictly have to match plugin name,
	// but we enforced it in the 'install' command.
	fileName := fmt.Sprintf("./plugins/%s.go", target)

	// Check if file exists
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return fmt.Errorf("source file not found at %s", fileName)
	}

	err := os.Remove(fileName)
	if err != nil {
		return err
	}
	result = models.DB.Delete(&models.ExternalPlugin{}, "name = ?", target)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func installPlugin(link string) (string, error) {
	log.Println("Installing plugin from URL:", link)
	//convert bytes to string
	// Validate URL
	u, err := url.Parse(link)
	if err != nil || u.Scheme == "" || u.Host == "" {

		return "", fmt.Errorf("invalid URL")
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
		return "", fmt.Errorf("plugin already installed")
	}
	pluginName, err := helpers.InstallExternalPlugin(link)
	if err != nil {
		return pluginName, err
	}
	// Save to DB
	newPlugin := models.ExternalPlugin{
		Name: pluginName,
		Url:  link,
	}
	result := models.DB.Create(&newPlugin)
	if result.Error != nil {
		return pluginName, result.Error
	}
	// build and check for errors (invoke go directly for portability)
	_, err = exec.Command("go", "build", "-o", "bot", "main.go").Output()
	if err != nil {
		return pluginName, err
	}
	return pluginName, nil
}
