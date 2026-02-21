package scheduler

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestNewScheduler(t *testing.T) {
	s := NewScheduler()
	if s == nil {
		t.Fatal("Expected Scheduler instance, got nil")
	}
	if s.cron == nil {
		t.Error("Expected cron to be initialized")
	}
	if s.prefix != "dataease:scheduler:" {
		t.Errorf("Expected prefix 'dataease:scheduler:', got '%s'", s.prefix)
	}
}

func TestScheduler_AddFunc(t *testing.T) {
	s := NewScheduler()
	var executed atomic.Bool

	err := s.AddFunc("* * * * * *", func() {
		executed.Store(true)
	})

	if err != nil {
		t.Errorf("AddFunc failed: %v", err)
	}

	s.Start()
	defer s.Stop()

	time.Sleep(1100 * time.Millisecond)

	if !executed.Load() {
		t.Error("Expected job to be executed")
	}
}

func TestScheduler_AddFunc_InvalidSpec(t *testing.T) {
	s := NewScheduler()

	err := s.AddFunc("invalid-spec", func() {})

	if err == nil {
		t.Error("Expected error for invalid spec")
	}
}

func TestScheduler_StartStop(t *testing.T) {
	s := NewScheduler()

	s.Start()
	s.Stop()
}

func TestScheduler_SetRedis(t *testing.T) {
	s := NewScheduler()
	s.SetRedis(nil)

	if s.redis != nil {
		t.Error("Expected redis to be nil")
	}
}

func TestScheduler_AddDistributedFunc_NoRedis(t *testing.T) {
	s := NewScheduler()
	var executed atomic.Bool

	err := s.AddDistributedFunc("test-job", "* * * * * *", func() {
		executed.Store(true)
	})

	if err != nil {
		t.Errorf("AddDistributedFunc failed: %v", err)
	}

	s.Start()
	defer s.Stop()

	time.Sleep(1100 * time.Millisecond)

	if !executed.Load() {
		t.Error("Expected job to be executed even without Redis")
	}
}

func TestScheduler_Entries(t *testing.T) {
	s := NewScheduler()
	_ = s.AddFunc("* * * * * *", func() {})

	entries := s.Entries()
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
}
