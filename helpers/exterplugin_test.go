package helpers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestInstallExternalPlugin(t *testing.T) {
	validServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `package plugins

import "github.com/darkdeathoriginal/gogrambot/handler"

func init() {
	handler.NewPlugin("testplugin").Handle(func(m *telegram.NewMessage) error { return nil })
}`)
	}))
	defer validServer.Close()

	os.Mkdir("./plugins", 0755)
	defer os.RemoveAll("./plugins")

	pluginName, err := InstallExternalPlugin(validServer.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if pluginName != "testplugin" {
		t.Errorf("Expected 'testplugin', got '%s'", pluginName)
	}

	invalidPackageServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `package someotherpackage`)
	}))
	defer invalidPackageServer.Close()

	_, err = InstallExternalPlugin(invalidPackageServer.URL)
	if err == nil {
		t.Errorf("Expected error for invalid package, got nil")
	}

	_, err = InstallExternalPlugin("http://this-is-an-invalid-url-that-does-not-exist.com")
	if err == nil {
		t.Errorf("Expected error for invalid URL, got nil")
	}
}
