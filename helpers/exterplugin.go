package helpers

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

	"github.com/darkdeathoriginal/gogrambot/models"
)

func InstallExternalPlugin(link string) (string, error) {
	u, err := url.Parse(link)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("Invalid URL.")
	}
	resp, err := http.Get(link + "?timestamp=" + fmt.Sprint(os.Getpid()))
	if err != nil {
		return "", fmt.Errorf("Failed to download plugin: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read plugin data: %v", err)
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
			return "", fmt.Errorf("Invalid plugin. No plugin name found in code!")
		}
	}

	// Security Check: Ensure package is correct
	if !strings.Contains(code, "package plugins") {
		return "", fmt.Errorf("The code must belong to 'package plugins'")
	}

	// Save the file
	// Note: We use the extracted name for the filename
	fileName := fmt.Sprintf("./plugins/%s.go", pluginName)
	err = os.WriteFile(fileName, bodyBytes, 0644)
	if err != nil {
		return "", fmt.Errorf("Failed to save plugin file: %v", err)
	}
	return pluginName, nil
}

func LoadExternalPlugins() {
	// Load plugins from DB
	var externalPlugins []models.ExternalPlugin
	models.DB.Find(&externalPlugins)
	for _, p := range externalPlugins {
		if _, err := os.Stat("./plugins/" + p.Name + ".go"); os.IsNotExist(err) {
			log.Printf("Plugin file for '%s' not found. Downloading.\n", p.Name)
			InstallExternalPlugin(p.Url)
		} else {
			log.Printf("Plugin file for '%s' already exists. Skipping download.\n", p.Name)
		}
	}
}
