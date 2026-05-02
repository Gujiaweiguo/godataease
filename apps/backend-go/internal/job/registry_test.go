package scheduler

import (
	"context"
	"testing"
)

func TestRegistry_RegisterValidatesDefinitions(t *testing.T) {
	registry := NewRegistry()
	validRun := func(context.Context) error { return nil }

	tests := []struct {
		name string
		def  Definition
		err  string
	}{
		{name: "missing key", def: Definition{Metadata: Metadata{Spec: "* * * * * *", Description: "desc"}, Run: validRun}, err: "job key is required"},
		{name: "missing spec", def: Definition{Metadata: Metadata{Key: "key", Description: "desc"}, Run: validRun}, err: "job spec is required"},
		{name: "missing description", def: Definition{Metadata: Metadata{Key: "key", Spec: "* * * * * *"}, Run: validRun}, err: "job description is required"},
		{name: "missing run func", def: Definition{Metadata: Metadata{Key: "key", Spec: "* * * * * *", Description: "desc"}}, err: "job run function is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.Register(tt.def)
			if err == nil || err.Error() != tt.err {
				t.Fatalf("expected error %q, got %v", tt.err, err)
			}
		})
	}
}

func TestRegistry_EnabledJobsFiltersDisabledEntries(t *testing.T) {
	registry := NewRegistry()
	run := func(context.Context) error { return nil }

	if err := registry.Register(Definition{Metadata: Metadata{Key: "enabled", Spec: "* * * * * *", Description: "enabled", Enabled: true}, Run: run}); err != nil {
		t.Fatalf("register enabled: %v", err)
	}
	if err := registry.Register(Definition{Metadata: Metadata{Key: "disabled", Spec: "* * * * * *", Description: "disabled", Enabled: false}, Run: run}); err != nil {
		t.Fatalf("register disabled: %v", err)
	}

	enabled := registry.EnabledJobs()
	if len(enabled) != 1 {
		t.Fatalf("expected 1 enabled job, got %d", len(enabled))
	}
	if enabled[0].Metadata.Key != "enabled" {
		t.Fatalf("expected enabled job key, got %s", enabled[0].Metadata.Key)
	}
}

func TestRegistry_RegisterRejectsDuplicateKeys(t *testing.T) {
	registry := NewRegistry()
	run := func(context.Context) error { return nil }
	def := Definition{Metadata: Metadata{Key: "dup", Spec: "* * * * * *", Description: "dup", Enabled: true}, Run: run}

	if err := registry.Register(def); err != nil {
		t.Fatalf("first register: %v", err)
	}
	if err := registry.Register(def); err == nil || err.Error() != "job key dup already registered" {
		t.Fatalf("expected duplicate key error, got %v", err)
	}
}

func TestScheduler_AddRegistryLoadsOnlyEnabledJobs(t *testing.T) {
	s := NewScheduler()
	registry := NewRegistry()
	run := func(context.Context) error { return nil }

	if err := registry.Register(Definition{Metadata: Metadata{Key: "enabled", Spec: "* * * * * *", Description: "enabled", Enabled: true}, Run: run}); err != nil {
		t.Fatalf("register enabled: %v", err)
	}
	if err := registry.Register(Definition{Metadata: Metadata{Key: "disabled", Spec: "* * * * * *", Description: "disabled", Enabled: false}, Run: run}); err != nil {
		t.Fatalf("register disabled: %v", err)
	}

	if err := s.AddRegistry(registry, func(Result) {}); err != nil {
		t.Fatalf("add registry: %v", err)
	}

	if got := len(s.Entries()); got != 1 {
		t.Fatalf("expected 1 registered entry, got %d", got)
	}
}
