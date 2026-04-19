package database

import (
	"testing"
)

func TestPtrString(t *testing.T) {
	s := "hello"
	p := ptrString(s)
	if p == nil {
		t.Fatal("ptrString returned nil")
	}
	if *p != s {
		t.Fatalf("expected %q, got %q", s, *p)
	}
}

func TestPtrInt(t *testing.T) {
	i := 42
	p := ptrInt(i)
	if p == nil {
		t.Fatal("ptrInt returned nil")
	}
	if *p != i {
		t.Fatalf("expected %d, got %d", i, *p)
	}
}

func TestPtrStringEmpty(t *testing.T) {
	p := ptrString("")
	if p == nil {
		t.Fatal("ptrString returned nil for empty string")
	}
	if *p != "" {
		t.Fatalf("expected empty string, got %q", *p)
	}
}

func TestPtrIntZero(t *testing.T) {
	p := ptrInt(0)
	if p == nil {
		t.Fatal("ptrInt returned nil for zero")
	}
	if *p != 0 {
		t.Fatalf("expected 0, got %d", *p)
	}
}
