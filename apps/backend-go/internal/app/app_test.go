package app

import (
	"testing"
)

func TestApplication_Fields(t *testing.T) {
	app := &Application{
		Name:    "test-app",
		Version: "1.0.0",
		Config:  nil,
	}

	if app.Name != "test-app" {
		t.Errorf("Expected Name 'test-app', got '%s'", app.Name)
	}
	if app.Version != "1.0.0" {
		t.Errorf("Expected Version '1.0.0', got '%s'", app.Version)
	}
}

func TestApplication_NilConfig(t *testing.T) {
	app := &Application{
		Name:    "test",
		Version: "1.0",
		Config:  nil,
	}

	if app.Config != nil {
		t.Error("Expected Config to be nil")
	}
}
