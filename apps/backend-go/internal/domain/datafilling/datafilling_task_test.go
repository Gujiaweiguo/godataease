package datafilling

import "testing"

func TestTaskStatusConstants(t *testing.T) {
	if TaskStatusStopped != 0 || TaskStatusStarted != 1 {
		t.Fatalf("unexpected task status constants")
	}
	if SubTaskStatusExpired != 0 || SubTaskStatusActive != 1 {
		t.Fatalf("unexpected sub task status constants")
	}
	if SubInstanceStatusOpen != 0 || SubInstanceStatusFinished != 1 {
		t.Fatalf("unexpected sub instance status constants")
	}
}
