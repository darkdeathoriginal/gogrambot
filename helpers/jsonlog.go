package helpers

import (
	"encoding/json"
	"log"
)

func JsonLog(v any) {
	// Pretty print JSON for logging
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Println("Error marshalling JSON:", err)
		return
	}
	log.Println(string(jsonData))
}
