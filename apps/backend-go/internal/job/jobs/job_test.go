package jobs

import (
	"context"
	"testing"

	"dataease/backend/internal/app"
)

func TestPackageExists(t *testing.T) {
}

func TestNewRegistry_UsesSampleJobToggle(t *testing.T) {
	registry, err := NewRegistry(app.SchedulerConfig{SampleJobEnabled: true})
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}

	jobs := registry.Jobs()
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if !jobs[0].Metadata.Enabled {
		t.Fatal("expected sample job to be enabled")
	}
	if jobs[0].Run == nil {
		t.Fatal("expected sample job run function")
	}
	if err := jobs[0].Run(context.Background()); err != nil {
		t.Fatalf("run sample job: %v", err)
	}
}

func TestNewRegistry_DisablesSampleJobWhenToggleOff(t *testing.T) {
	registry, err := NewRegistry(app.SchedulerConfig{SampleJobEnabled: false})
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}

	jobs := registry.Jobs()
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Metadata.Enabled {
		t.Fatal("expected sample job to be disabled")
	}
	if enabled := registry.EnabledJobs(); len(enabled) != 0 {
		t.Fatalf("expected no enabled jobs, got %d", len(enabled))
	}
}
