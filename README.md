# GoGramBot

GoGramBot is a powerful and extensible Telegram Userbot entirely written in Go using the [`gogram`](https://github.com/amarnathcjd/gogram) library. It features a Web UI for effortless authentication (QR Code and 2-Step Verification support) and an intuitive plugin architecture.

## ✨ Features

- **Web-based Authentication:** Easily log in to your Telegram account by scanning a QR Code or entering your 2FA password via the built-in Web dashboard.
- **Dynamic Plugin Ecosystem:** Plugins are loaded seamlessly via a builder API, supporting custom filters and event handlers.
- **Built-in Plugins:** Comes with robust plugins including Admin utilities (`ban`, `mute`, `pin`), Userbot utilities (`afk`, `ping`, `whois`, `kang`, `purge`), system status checkers, and more.
- **Auto-Recompilation:** The included `start.sh` automatically compiles and handles restarts for continuous execution.
- **Docker & Panel Ready:** Full support for containerized deployments with a `Dockerfile`, making it perfect for Pterodactyl panels or generic VPS hosting.
- **Database Supported:** Integrates `gorm` (SQLite/PostgreSQL) for structured data retention.

---

## 🚀 Setup & Installation

### Prerequisites

- [Go 1.25.0+](https://golang.org/dl/) installed.
- Your Telegram `API_ID` and `API_HASH` from [my.telegram.org](https://my.telegram.org/).

### Configuration

Create a `.env` file in the root directory and add your credentials:

```dotenv
PORT=8080
API_ID=your_api_id
API_HASH=your_api_hash
COMMAND_PREFIX=.
```

### Running Locally

To run the bot locally, simply execute:

```bash
go run main.go
```

Then, open your browser and navigate to `http://localhost:8080`.
The web interface will guide you through scanning the QR code with your Telegram app to finalize the login process.

---

## 🐳 Deployment (Docker)

If you're using Docker, you can build and run the image directly:

```bash
docker build -t gogrambot .
docker run -p 8080:8080 --env-file .env gogrambot
```

> **Note:** The `Dockerfile` relies on `start.sh` as the entry point, resolving dependencies, compiling the Go binary (`bot`), and handling restarts natively within the container.

---

## 🧩 Plugin Development

Writing a new plugin is incredibly straightforward thanks to the Plugin Builder API.

Create a new go file in the `plugins/` folder (e.g., `hello.go`):

```go
package plugins

import (
	"github.com/amarnathcjd/gogram/telegram"
	"github.com/darkdeathoriginal/gogrambot/handler"
)

func init() {
	handler.NewPlugin("hello").
		Description("Says Hello World").
		Category("General").
		Handle(func(message *telegram.NewMessage) error {
			message.Reply("Hello World!")
			return nil
		})
}
```

The system will automatically register and route commands matching `.hello` to this handler.

---

## 📄 Available Plugins Outline

- **Admin module**: `.ban`, `.unban`, `.mute`, `.unmute`, `.pin`
- **Userbot basics**: `.ping`, `.afk`, `.purge`, `.kang`, `.whois`, `.save`
- **Misc/Dev**: `.id`, formatting tools, torrent management, cache tools.

---

## 📜 License

This project relies on the [gogram library](https://github.com/amarnathcjd/gogram). Check the source repositories for detailed license information.
