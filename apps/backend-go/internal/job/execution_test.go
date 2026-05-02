package scheduler

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestScheduler_RunDefinitionClassifiesSuccessAndReleasesLock(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	s := NewScheduler()
	s.SetRedis(client)
	var called atomic.Bool

	def := Definition{
		Metadata: Metadata{Key: "sample", Spec: "* * * * * *", Description: "sample", Enabled: true, Distributed: true},
		Run: func(context.Context) error {
			called.Store(true)
			return nil
		},
	}

	got := s.runDefinition(context.Background(), def, func(Result) {})
	if !called.Load() {
		t.Fatal("expected job to run")
	}
	if got.Outcome != OutcomeSuccess {
		t.Fatalf("expected success outcome, got %s", got.Outcome)
	}
	if got.Err != nil {
		t.Fatalf("expected nil error, got %v", got.Err)
	}
	if mr.Exists("dataease:scheduler:sample:lock") {
		t.Fatal("expected lock key to be released")
	}
}

func TestScheduler_RunDefinitionClassifiesFailedAndReleasesLock(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	s := NewScheduler()
	s.SetRedis(client)
	def := Definition{
		Metadata: Metadata{Key: "sample", Spec: "* * * * * *", Description: "sample", Enabled: true, Distributed: true},
		Run: func(context.Context) error {
			return errors.New("boom")
		},
	}

	got := s.runDefinition(context.Background(), def, func(Result) {})
	if got.Outcome != OutcomeFailed {
		t.Fatalf("expected failed outcome, got %s", got.Outcome)
	}
	if got.Err == nil || got.Err.Error() != "boom" {
		t.Fatalf("expected boom error, got %v", got.Err)
	}
	if mr.Exists("dataease:scheduler:sample:lock") {
		t.Fatal("expected lock key to be released after failure")
	}
}

func TestScheduler_RunDefinitionClassifiesSkippedOnLockContention(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	s := NewScheduler()
	s.SetRedis(client)
	if err := client.Set(context.Background(), "dataease:scheduler:sample:lock", "1", defaultDistributedLockTTL).Err(); err != nil {
		t.Fatalf("seed lock key: %v", err)
	}

	var called atomic.Bool
	def := Definition{
		Metadata: Metadata{Key: "sample", Spec: "* * * * * *", Description: "sample", Enabled: true, Distributed: true},
		Run: func(context.Context) error {
			called.Store(true)
			return nil
		},
	}

	got := s.runDefinition(context.Background(), def, func(Result) {})
	if got.Outcome != OutcomeSkipped {
		t.Fatalf("expected skipped outcome, got %s", got.Outcome)
	}
	if called.Load() {
		t.Fatal("expected job body not to run when lock is held")
	}
}

func TestScheduler_RunDefinitionClassifiesFailedOnLockError(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	s := NewScheduler()
	s.SetRedis(client)
	mr.Close()
	t.Cleanup(func() { _ = client.Close() })

	def := Definition{
		Metadata: Metadata{Key: "sample", Spec: "* * * * * *", Description: "sample", Enabled: true, Distributed: true},
		Run:      func(context.Context) error { return nil },
	}

	got := s.runDefinition(context.Background(), def, func(Result) {})
	if got.Outcome != OutcomeFailed {
		t.Fatalf("expected failed outcome, got %s", got.Outcome)
	}
	if got.Err == nil {
		t.Fatal("expected lock acquisition error")
	}
}

func TestScheduler_RunDefinitionWithoutDistributedLock(t *testing.T) {
	s := NewScheduler()
	var called atomic.Bool
	def := Definition{
		Metadata: Metadata{Key: "local", Spec: "* * * * * *", Description: "local", Enabled: true, Distributed: false},
		Run: func(context.Context) error {
			called.Store(true)
			return nil
		},
	}

	got := s.runDefinition(context.Background(), def, func(Result) {})
	if !called.Load() {
		t.Fatal("expected non-distributed job to run")
	}
	if got.Outcome != OutcomeSuccess {
		t.Fatalf("expected success outcome, got %s", got.Outcome)
	}
}

func TestScheduler_RunDefinitionClassifiesPanicAsFailed(t *testing.T) {
	s := NewScheduler()
	got := s.runDefinition(context.Background(), Definition{
		Metadata: Metadata{Key: "panic", Spec: "* * * * * *", Description: "panic", Enabled: true},
		Run: func(context.Context) error {
			panic("boom")
		},
	}, func(Result) {})

	if got.Outcome != OutcomeFailed {
		t.Fatalf("expected failed outcome, got %s", got.Outcome)
	}
	if got.Err == nil || got.Err.Error() != "panic executing scheduled job: boom" {
		t.Fatalf("expected panic error, got %v", got.Err)
	}
}

func TestReleaseDistributedLockHonorsOwner(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	s := NewScheduler()
	s.SetRedis(client)

	if err := client.Set(context.Background(), "lock:key", "owner-b", defaultDistributedLockTTL).Err(); err != nil {
		t.Fatalf("seed lock key: %v", err)
	}
	if err := releaseDistributedLock(context.Background(), s, "lock:key", "owner-a"); err != nil {
		t.Fatalf("release lock: %v", err)
	}
	if !mr.Exists("lock:key") {
		t.Fatal("expected lock key to remain when owner mismatches")
	}
}
