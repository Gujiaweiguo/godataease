package static

import (
	"testing"
)

func TestStaticResource_TableName(t *testing.T) {
	r := StaticResource{}
	if r.TableName() != "static_resource" {
		t.Errorf("Expected table name 'static_resource', got '%s'", r.TableName())
	}
}

func TestStaticResource_Fields(t *testing.T) {
	r := StaticResource{
		ID:   "static-1",
		Name: "Logo",
		Path: "/static/logo.png",
		Type: "image",
	}

	if r.ID != "static-1" {
		t.Errorf("Expected ID 'static-1', got '%s'", r.ID)
	}
	if r.Name != "Logo" {
		t.Errorf("Expected Name 'Logo', got '%s'", r.Name)
	}
	if r.Path != "/static/logo.png" {
		t.Errorf("Expected Path '/static/logo.png', got '%s'", r.Path)
	}
	if r.Type != "image" {
		t.Errorf("Expected Type 'image', got '%s'", r.Type)
	}
}

func TestStore_TableName(t *testing.T) {
	s := Store{}
	if s.TableName() != "store" {
		t.Errorf("Expected table name 'store', got '%s'", s.TableName())
	}
}

func TestStore_Fields(t *testing.T) {
	s := Store{
		ID:   "store-1",
		Name: "Plugin Store",
		URL:  "https://store.example.com",
	}

	if s.ID != "store-1" {
		t.Errorf("Expected ID 'store-1', got '%s'", s.ID)
	}
	if s.Name != "Plugin Store" {
		t.Errorf("Expected Name 'Plugin Store', got '%s'", s.Name)
	}
	if s.URL != "https://store.example.com" {
		t.Errorf("Expected URL 'https://store.example.com', got '%s'", s.URL)
	}
}

func TestTypeface_TableName(t *testing.T) {
	tf := Typeface{}
	if tf.TableName() != "typeface" {
		t.Errorf("Expected table name 'typeface', got '%s'", tf.TableName())
	}
}

func TestTypeface_Fields(t *testing.T) {
	tf := Typeface{
		ID:   "font-1",
		Name: "Roboto",
		File: "/fonts/Roboto.ttf",
	}

	if tf.ID != "font-1" {
		t.Errorf("Expected ID 'font-1', got '%s'", tf.ID)
	}
	if tf.Name != "Roboto" {
		t.Errorf("Expected Name 'Roboto', got '%s'", tf.Name)
	}
	if tf.File != "/fonts/Roboto.ttf" {
		t.Errorf("Expected File '/fonts/Roboto.ttf', got '%s'", tf.File)
	}
}
