package service

import (
	"fmt"
	"testing"

	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/repository"
)

func setupPermService() *PermService {
	mockRepo := repository.NewMockPermRepository()
	return &PermService{permRepo: mockRepo}
}

func TestCreatePerm_Success(t *testing.T) {
	svc := setupPermService()

	desc := "test permission"
	req := &permission.PermCreateRequest{
		PermName: "Test Perm",
		PermKey:  "test:perm",
		PermType: permission.PermTypeMenu,
		PermDesc: &desc,
	}

	permID, err := svc.CreatePerm(req)
	if err != nil {
		t.Fatalf("CreatePerm failed: %v", err)
	}
	if permID != 1 {
		t.Errorf("Expected permID 1, got %d", permID)
	}
}

func TestCreatePerm_DuplicateKey(t *testing.T) {
	svc := setupPermService()

	req := &permission.PermCreateRequest{
		PermName: "Test Perm",
		PermKey:  "test:perm",
	}

	_, err := svc.CreatePerm(req)
	if err != nil {
		t.Fatalf("First CreatePerm failed: %v", err)
	}

	_, err = svc.CreatePerm(req)
	if err == nil {
		t.Error("Expected error for duplicate key, got nil")
	}
}

func TestListPerms(t *testing.T) {
	svc := setupPermService()

	for i := 1; i <= 15; i++ {
		key := fmt.Sprintf("test:perm:%d", i)
		req := &permission.PermCreateRequest{
			PermName: fmt.Sprintf("Test Perm %d", i),
			PermKey:  key,
		}
		_, _ = svc.CreatePerm(req)
	}

	req := &permission.PermQueryRequest{Current: 1, Size: 10}
	result, err := svc.ListPerms(req)
	if err != nil {
		t.Fatalf("ListPerms failed: %v", err)
	}

	if len(result.List.([]*permission.SysPerm)) != 10 {
		t.Errorf("Expected 10 items, got %d", len(result.List.([]*permission.SysPerm)))
	}
	if result.Total != 15 {
		t.Errorf("Expected total 15, got %d", result.Total)
	}
}

func TestDeletePerm(t *testing.T) {
	svc := setupPermService()

	req := &permission.PermCreateRequest{
		PermName: "Test Perm",
		PermKey:  "test:perm",
	}
	permID, _ := svc.CreatePerm(req)

	err := svc.DeletePerm(permID)
	if err != nil {
		t.Fatalf("DeletePerm failed: %v", err)
	}

	_, err = svc.GetPermByID(permID)
	if err == nil {
		t.Error("Expected error after delete, got nil")
	}
}
