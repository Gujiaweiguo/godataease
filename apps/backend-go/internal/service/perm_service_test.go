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

type failingUpdatePermRepo struct {
	repository.PermRepositoryInterface
}

type failingListPermRepo struct {
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

func (r *failingUpdatePermRepo) Update(p *permission.SysPerm) error {
	return errors.New("update failed")
}

func (r *failingListPermRepo) List() ([]*permission.SysPerm, error) {
	return nil, errors.New("list failed")
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

func TestCreatePerm_DefaultTypeAndCustomStatus(t *testing.T) {
	svc := setupPermService()
	disabled := permission.StatusDisabled

	permID, err := svc.CreatePerm(&permission.PermCreateRequest{PermName: "Default Type", PermKey: "default:type", Status: &disabled})
	if err != nil {
		t.Fatalf("CreatePerm failed: %v", err)
	}

	created, err := svc.GetPermByID(permID)
	if err != nil {
		t.Fatalf("GetPermByID failed: %v", err)
	}
	if created.PermType != permission.PermTypeMenu {
		t.Fatalf("Expected default perm type %s, got %s", permission.PermTypeMenu, created.PermType)
	}
	if created.Status != disabled {
		t.Fatalf("Expected status %d, got %d", disabled, created.Status)
	}
}

func TestUpdatePerm(t *testing.T) {
	t.Run("returns not found when permission missing", func(t *testing.T) {
		svc := setupPermService()

		err := svc.UpdatePerm(&permission.PermUpdateRequest{PermID: 999, PermName: "missing"})
		if err == nil {
			t.Fatal("Expected not found error")
		}
	})

	t.Run("rejects duplicate changed key", func(t *testing.T) {
		svc := setupPermService()
		_, _ = svc.CreatePerm(&permission.PermCreateRequest{PermName: "Perm A", PermKey: "perm:a"})
		permID, _ := svc.CreatePerm(&permission.PermCreateRequest{PermName: "Perm B", PermKey: "perm:b"})

		err := svc.UpdatePerm(&permission.PermUpdateRequest{PermID: permID, PermKey: "perm:a"})
		if err == nil {
			t.Fatal("Expected duplicate key error")
		}
	})

	t.Run("updates changed fields and changed key", func(t *testing.T) {
		svc := setupPermService()
		desc := "old desc"
		permID, _ := svc.CreatePerm(&permission.PermCreateRequest{PermName: "Perm A", PermKey: "perm:a", PermDesc: &desc})
		newDesc := "new desc"
		disabled := permission.StatusDisabled

		err := svc.UpdatePerm(&permission.PermUpdateRequest{PermID: permID, PermName: "Perm A Updated", PermKey: "perm:a:updated", PermType: permission.PermTypeButton, PermDesc: &newDesc, Status: &disabled})
		if err != nil {
			t.Fatalf("UpdatePerm failed: %v", err)
		}

		updated, err := svc.GetPermByID(permID)
		if err != nil {
			t.Fatalf("GetPermByID failed: %v", err)
		}
		if updated.PermName != "Perm A Updated" || updated.PermKey != "perm:a:updated" || updated.PermType != permission.PermTypeButton {
			t.Fatalf("Unexpected updated permission: %+v", updated)
		}
		if updated.PermDesc == nil || *updated.PermDesc != newDesc {
			t.Fatal("Expected updated perm desc")
		}
		if updated.Status != disabled {
			t.Fatalf("Expected status %d, got %d", disabled, updated.Status)
		}
		if updated.UpdateTime == nil {
			t.Fatal("Expected update time to be set")
		}
	})

	t.Run("propagates check and update errors", func(t *testing.T) {
		base := repository.NewMockPermRepository()
		_ = base.Create(&permission.SysPerm{PermName: "Perm A", PermKey: "perm:a", PermType: permission.PermTypeMenu, Status: permission.StatusEnabled, DelFlag: permission.DelFlagNormal})
		svc := NewPermService(&failingCheckPermRepo{PermRepositoryInterface: base})

		err := svc.UpdatePerm(&permission.PermUpdateRequest{PermID: 1, PermKey: "perm:b"})
		if err == nil {
			t.Fatal("Expected check key error")
		}

		base2 := repository.NewMockPermRepository()
		_ = base2.Create(&permission.SysPerm{PermName: "Perm B", PermKey: "perm:b", PermType: permission.PermTypeMenu, Status: permission.StatusEnabled, DelFlag: permission.DelFlagNormal})
		svc = NewPermService(&failingUpdatePermRepo{PermRepositoryInterface: base2})
		err = svc.UpdatePerm(&permission.PermUpdateRequest{PermID: 1, PermName: "broken update"})
		if err == nil {
			t.Fatal("Expected update error")
		}
	})
}

func TestListPerms_NormalizationAndErrors(t *testing.T) {
	t.Run("normalizes invalid page inputs and handles out of range", func(t *testing.T) {
		svc := setupPermService()
		for i := 1; i <= 3; i++ {
			_, _ = svc.CreatePerm(&permission.PermCreateRequest{PermName: fmt.Sprintf("Perm %d", i), PermKey: fmt.Sprintf("perm:%d", i)})
		}

		result, err := svc.ListPerms(&permission.PermQueryRequest{Current: 0, Size: 0})
		if err != nil {
			t.Fatalf("ListPerms failed: %v", err)
		}
		if result.Current != 1 || result.Size != 10 {
			t.Fatalf("Expected normalized pagination 1/10, got %d/%d", result.Current, result.Size)
		}
		if len(result.List.([]*permission.SysPerm)) != 3 {
			t.Fatalf("Expected 3 items, got %d", len(result.List.([]*permission.SysPerm)))
		}

		result, err = svc.ListPerms(&permission.PermQueryRequest{Current: 5, Size: 2})
		if err != nil {
			t.Fatalf("ListPerms failed: %v", err)
		}
		if len(result.List.([]*permission.SysPerm)) != 0 {
			t.Fatalf("Expected empty out-of-range page, got %d", len(result.List.([]*permission.SysPerm)))
		}
	})

	t.Run("propagates repository list error", func(t *testing.T) {
		svc := NewPermService(&failingListPermRepo{PermRepositoryInterface: repository.NewMockPermRepository()})

		_, err := svc.ListPerms(&permission.PermQueryRequest{Current: 1, Size: 10})
		if err == nil {
			t.Fatal("Expected list error")
		}
	})
}
