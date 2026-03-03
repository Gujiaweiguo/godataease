package service

import (
	"errors"
	"fmt"
	"testing"

	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/repository"
)

type failingDeletePermRepo struct {
	repository.PermRepositoryInterface
}

type failingCheckPermRepo struct {
	repository.PermRepositoryInterface
}

type failingCreatePermRepo struct {
	repository.PermRepositoryInterface
}

func (r *failingDeletePermRepo) Delete(permID int64) error {
	return errors.New("delete failed")
}

func (r *failingCheckPermRepo) CheckKeyExists(permKey string, excludePermID int64) (int64, error) {
	return 0, errors.New("check failed")
}

func (r *failingCreatePermRepo) Create(p *permission.SysPerm) error {
	return errors.New("create failed")
}

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
func TestCheckPermKeyExists(t *testing.T) {
	svc := setupPermService()

	// Create a permission first
	req := &permission.PermCreateRequest{
		PermName: "Test Exists",
		PermKey:  "test:exists",
	}
	_, err := svc.CreatePerm(req)
	if err != nil {
		t.Fatalf("CreatePerm failed: %v", err)
	}

	// Check if key exists
	exists, err := svc.CheckPermKeyExists("test:exists")
	if err != nil {
		t.Fatalf("CheckPermKeyExists failed: %v", err)
	}
	if !exists {
		t.Error("Expected perm key to exist")
	}

	// Check non-existent key
	exists, err = svc.CheckPermKeyExists("non:existent:key")
	if err != nil {
		t.Fatalf("CheckPermKeyExists for non-existent failed: %v", err)
	}
	if exists {
		t.Error("Expected perm key to not exist")
	}
}

func TestDeletePerm_Error(t *testing.T) {
	svc := NewPermService(&failingDeletePermRepo{PermRepositoryInterface: repository.NewMockPermRepository()})

	err := svc.DeletePerm(1)
	if err == nil {
		t.Fatal("Expected error from DeletePerm")
	}
	if err.Error() == "" {
		t.Fatal("Expected non-empty error message")
	}
}

func TestCheckPermKeyExists_Error(t *testing.T) {
	svc := NewPermService(&failingCheckPermRepo{PermRepositoryInterface: repository.NewMockPermRepository()})

	_, err := svc.CheckPermKeyExists("x:y")
	if err == nil {
		t.Fatal("Expected error from CheckPermKeyExists")
	}
}

func TestCreatePerm_CheckKeyError(t *testing.T) {
	svc := NewPermService(&failingCheckPermRepo{PermRepositoryInterface: repository.NewMockPermRepository()})

	_, err := svc.CreatePerm(&permission.PermCreateRequest{PermName: "x", PermKey: "x:y"})
	if err == nil {
		t.Fatal("Expected error from CreatePerm when check key fails")
	}
}

func TestCreatePerm_CreateError(t *testing.T) {
	svc := NewPermService(&failingCreatePermRepo{PermRepositoryInterface: repository.NewMockPermRepository()})

	_, err := svc.CreatePerm(&permission.PermCreateRequest{PermName: "x", PermKey: "x:y"})
	if err == nil {
		t.Fatal("Expected error from CreatePerm when repository create fails")
	}
}
