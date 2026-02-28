package plugins

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/darkdeathoriginal/gogrambot/handler"
)

func init() {
	handler.NewPlugin("menu").
		Description("Displays available commands categorized").
		Usage("menu [command name]").
		Category("General").
		Handle(menuHandler)

	// Alias 'help' to 'menu'
	handler.NewPlugin("help").
		Description("Alias for menu").
		Category("General").
		On("cmd:help").
		Handle(menuHandler)

	handler.NewPlugin("list").
		Description("Lists installed plugins").
		Category("Owner").
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
}

func menuHandler(m *telegram.NewMessage) error {
	args := m.Args()

	// 1. Specific Command Details (e.g., .menu ping)
	if len(args) > 0 {
		cmdName := strings.ToLower(args)
		for _, p := range handler.Plugins {
			if strings.EqualFold(p.Name, cmdName) {
				m.Reply(fmt.Sprintf(
					"╭─── [ **%s** ]\n"+
						"│\n"+
						"│ 📝 **Desc:** %s\n"+
						"│ 📂 **Cat:** %s\n"+
						"│ ⌨️ **Usage:** %s\n"+
						"│\n"+
						"╰───────────────",
					strings.ToUpper(p.Name),
					p.Description,
					p.Category,
					p.Usage,
				))
				return nil
			}
		}
		m.Reply("❌ Command not found.")
		return nil
	}

	// 2. Full Menu Generation

	// Step A: Group plugins by Category
	catMap := make(map[string][]handler.Plugin)
	var categories []string

	for _, p := range handler.Plugins {
		cat := p.Category
		if cat == "" {
			cat = "Uncategorized"
		}

		// Initialize slice if key doesn't exist
		if _, ok := catMap[cat]; !ok {
			categories = append(categories, cat)
		}
		catMap[cat] = append(catMap[cat], p)
	}

	// Step B: Sort Categories Alphabetically
	sort.Strings(categories)

	// Step C: Build the ASCII Message
	var sb strings.Builder

	// -- Header --
	sb.WriteString(getAsciiHeader())

	// -- Body --
	for _, cat := range categories {
		plugins := catMap[cat]

		// Sort plugins inside category
		sort.Slice(plugins, func(i, j int) bool {
			return plugins[i].Name < plugins[j].Name
		})

		// Category Header
		sb.WriteString(fmt.Sprintf("\n╭─── 「 %s 」\n", strings.ToUpper(cat)))

		// Plugin Rows
		for _, p := range plugins {
			sb.WriteString(fmt.Sprintf("│ ◦ %s\n", p.Name))
		}

		// Category Footer
		sb.WriteString("╰───────────────\n")
	}

	// -- Footer --
	sb.WriteString(fmt.Sprintf("\n__Server Time: %s__", time.Now().Format("15:04")))

	m.Reply(sb.String())
	return nil
}

// Simple ASCII Header
func getAsciiHeader() string {
	return `
╭───────────────╮
  🤖 𝗕𝗢𝗧 𝗠𝗘𝗡𝗨
╰───────────────╯`
}
