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
	Port = Getenv("PORT", "8080")
	AppID = Getenv("API_ID", "")
	AppHash = Getenv("API_HASH", "")
	CommandPrefix = Getenv("COMMAND_PREFIX", ".")
}
func Getenv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
