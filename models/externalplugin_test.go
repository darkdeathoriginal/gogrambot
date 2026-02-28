package models

import (
	"testing"
)

func TestExternalPluginTableName(t *testing.T) {
	plugin := ExternalPlugin{}
	if plugin.TableName() != "external_plugin" {
		t.Errorf("Expected 'external_plugin', got '%s'", plugin.TableName())
	}
}
