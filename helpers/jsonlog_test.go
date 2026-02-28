package helpers

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestJsonLog(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(log.Writer())

	data := map[string]string{"key": "value"}
	JsonLog(data)

	output := buf.String()
	if !strings.Contains(output, `"key": "value"`) {
		t.Errorf("Expected log output to contain '\"key\": \"value\"', got '%s'", output)
	}

	buf.Reset()
	invalidData := make(chan int)
	JsonLog(invalidData)
	output = buf.String()
	if !strings.Contains(output, "Error marshalling JSON:") {
		t.Errorf("Expected log output to contain 'Error marshalling JSON:', got '%s'", output)
	}
}
