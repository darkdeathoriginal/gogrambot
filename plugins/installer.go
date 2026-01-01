package plugins

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/darkdeathoriginal/gogrambot/handler"
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

			// Add timestamp to bypass cache
			resp, err := http.Get(link + "?timestamp=" + fmt.Sprint(os.Getpid()))
			if err != nil {
				message.Reply("Failed to fetch URL: " + err.Error())
				return nil
			}
			defer resp.Body.Close()

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				message.Reply("Failed to read data.")
				return nil
			}
			code := string(bodyBytes)

			// Extract Plugin Name using Regex
			// Matches: handler.NewPlugin("name")
			re := regexp.MustCompile(`NewPlugin\(\s*["']([^"']+)["']\s*\)`)
			matches := re.FindStringSubmatch(code)

			var pluginName string
			if len(matches) > 1 {
				pluginName = matches[1]
			} else {
				// Fallback: Try to use filename from URL
				base := filepath.Base(u.Path)
				pluginName = strings.TrimSuffix(base, filepath.Ext(base))
				if pluginName == "" || pluginName == "." {
					message.Reply("__Invalid plugin. No plugin name found in code!__")
					return nil
				}
			}

			// Security Check: Ensure package is correct
			if !strings.Contains(code, "package plugins") {
				message.Reply("The code must belong to 'package plugins'")
				return nil
			}

			// Save the file
			// Note: We use the extracted name for the filename
			fileName := fmt.Sprintf("./plugins/%s.go", pluginName)
			err = os.WriteFile(fileName, bodyBytes, 0644)
			if err != nil {
				message.Reply("Failed to save file: " + err.Error())
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
			if len(args) > 0 {
				query := args
				for _, p := range handler.Plugins {
					if p.Name == query {
						message.Reply(fmt.Sprintf("**%s**\nDesc: %s\nUsage: %s", p.Name, p.Description, p.Usage), &telegram.SendOptions{
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
			for _, p := range handler.Plugins {
				msg.WriteString(fmt.Sprintf("• `%s`: %s\n", p.Name, p.Description))
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

			// 1. Check if plugin exists in memory
			var found bool
			for _, p := range handler.Plugins {
				if p.Name == target {
					found = true
					break
				}
			}

			if !found {
				message.Reply("Plugin not active in memory.")
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

			message.Reply(fmt.Sprintf("**%s** removed successfully.\n\n__⚠️ Restart required to fully unload.__", target), &telegram.SendOptions{
				ParseMode: telegram.MarkDown,
			})
			return nil
		})
}
