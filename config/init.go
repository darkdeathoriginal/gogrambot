package config

import (
	"os"

	"github.com/joho/godotenv"
)

var (
	Port          string
	AppID         string
	AppHash       string
	CommandPrefix string
)

func init() {
	godotenv.Load()
	Port = getenv("PORT", "8080")
	AppID = getenv("API_ID", "")
	AppHash = getenv("API_HASH", "")
	CommandPrefix = getenv("COMMAND_PREFIX", ".")
}
func getenv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
