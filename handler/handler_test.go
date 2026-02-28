package handler

import (
	"testing"

	"github.com/amarnathcjd/gogram/telegram"
)

func TestPluginBuilder(t *testing.T) {
	builder := NewPlugin("test_plugin").
		Description("A test plugin").
		Category("test").
		Usage(".test").
		On("message").
		AllowAll(true)

	p := builder.p

	if p.Name != "test_plugin" {
		t.Errorf("Expected name 'test_plugin', got '%s'", p.Name)
	}
	if p.Description != "A test plugin" {
		t.Errorf("Expected description 'A test plugin', got '%s'", p.Description)
	}
	if p.Category != "test" {
		t.Errorf("Expected category 'test', got '%s'", p.Category)
	}
	if p.Usage != ".test" {
		t.Errorf("Expected usage '.test', got '%s'", p.Usage)
	}
	if p.On != "message" {
		t.Errorf("Expected on event 'message', got '%s'", p.On)
	}
	if p.AllowAll != true {
		t.Errorf("Expected AllowAll true, got %v", p.AllowAll)
	}
}

func TestRegisterPlugin(t *testing.T) {
	initialCount := len(Plugins)

	NewPlugin("unique_test_plugin").Handle(func(m *telegram.NewMessage) error {
		return nil
	})

	if len(Plugins) != initialCount+1 {
		t.Errorf("Expected %d plugins, got %d", initialCount+1, len(Plugins))
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic on duplicate registration")
		}
	}()

	NewPlugin("unique_test_plugin").Handle(func(m *telegram.NewMessage) error {
		return nil
	})
}
