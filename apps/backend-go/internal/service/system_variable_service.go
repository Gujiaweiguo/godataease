package service

import (
	"errors"
	"strings"

	"dataease/backend/internal/domain/system"
	"dataease/backend/internal/repository"

	"gorm.io/gorm"
)

var errSystemVariableRepoNotReady = errors.New("system variable repository not initialized")

type SystemVariableService struct {
	repo *repository.SystemVariableRepository
}

func NewSystemVariableService(repo *repository.SystemVariableRepository) *SystemVariableService {
	return &SystemVariableService{repo: repo}
}

func (s *SystemVariableService) Create(req *system.SysVariable) (*system.SysVariable, error) {
	if s.repo == nil {
		return nil, errSystemVariableRepoNotReady
	}
	if err := validateVariable(req); err != nil {
		return nil, err
	}
	item := *req
	item.ID = 0
	if err := s.repo.Create(&item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *SystemVariableService) Edit(req *system.SysVariable) (*system.SysVariable, error) {
	if s.repo == nil {
		return nil, errSystemVariableRepoNotReady
	}
	if req == nil || req.ID <= 0 {
		return nil, gorm.ErrInvalidData
	}
	if err := validateVariable(req); err != nil {
		return nil, err
	}
	current, err := s.repo.GetByID(req.ID)
	if err != nil {
		return nil, err
	}
	current.Type = req.Type
	current.Name = req.Name
	current.Min = req.Min
	current.Max = req.Max
	current.StartTime = req.StartTime
	current.EndTime = req.EndTime
	current.Root = req.Root
	current.Disabled = req.Disabled
	if err = s.repo.Update(current); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *SystemVariableService) Detail(id int64) (*system.SysVariable, error) {
	if s.repo == nil {
		return nil, errSystemVariableRepoNotReady
	}
	return s.repo.GetByID(id)
}

func (s *SystemVariableService) Delete(id int64) error {
	if s.repo == nil {
		return errSystemVariableRepoNotReady
	}
	return s.repo.Delete(id)
}

func (s *SystemVariableService) Query(req *system.SysVariableQueryRequest) ([]system.SysVariable, error) {
	if s.repo == nil {
		return nil, errSystemVariableRepoNotReady
	}
	return s.repo.Query(req)
}

func (s *SystemVariableService) CreateValue(req *system.SysVariableValue) (*system.SysVariableValue, error) {
	if s.repo == nil {
		return nil, errSystemVariableRepoNotReady
	}
	if err := validateVariableValue(req); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetByID(req.SysVariableID); err != nil {
		return nil, err
	}
	item := *req
	item.ID = 0
	if err := s.repo.CreateValue(&item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *SystemVariableService) EditValue(req *system.SysVariableValue) (*system.SysVariableValue, error) {
	if s.repo == nil {
		return nil, errSystemVariableRepoNotReady
	}
	if req == nil || req.ID <= 0 {
		return nil, gorm.ErrInvalidData
	}
	if err := validateVariableValue(req); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetByID(req.SysVariableID); err != nil {
		return nil, err
	}
	current, err := s.repo.GetValueByID(req.ID)
	if err != nil {
		return nil, err
	}
	current.SysVariableID = req.SysVariableID
	current.Value = req.Value
	current.ValueDesc = req.ValueDesc
	current.Begin = req.Begin
	current.End = req.End
	if err = s.repo.UpdateValue(current); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *SystemVariableService) DeleteValue(id int64) error {
	if s.repo == nil {
		return errSystemVariableRepoNotReady
	}
	return s.repo.DeleteValue(id)
}

func (s *SystemVariableService) BatchDeleteValues(ids []int64) error {
	if s.repo == nil {
		return errSystemVariableRepoNotReady
	}
	return s.repo.BatchDeleteValues(ids)
}

func (s *SystemVariableService) SelectedValues(id int64) ([]system.SysVariableValue, error) {
	if s.repo == nil {
		return nil, errSystemVariableRepoNotReady
	}
	return s.repo.ListValuesByVariableID(id)
}

func (s *SystemVariableService) SelectedValuePage(page int, size int, req *system.SysVariableValueQueryRequest) (*system.SysVariableValuePage, error) {
	if s.repo == nil {
		return nil, errSystemVariableRepoNotReady
	}
	rows, total, err := s.repo.PageValues(page, size, req)
	if err != nil {
		return nil, err
	}
	return &system.SysVariableValuePage{Records: rows, Total: total, Current: page, Size: size}, nil
}

func validateVariable(req *system.SysVariable) error {
	if req == nil {
		return gorm.ErrInvalidData
	}
	if strings.TrimSpace(req.Type) == "" || strings.TrimSpace(req.Name) == "" {
		return gorm.ErrInvalidData
	}
	if req.Min > 0 && req.Max > 0 && req.Min > req.Max {
		return gorm.ErrInvalidData
	}
	return nil
}

func validateVariableValue(req *system.SysVariableValue) error {
	if req == nil || req.SysVariableID <= 0 || strings.TrimSpace(req.Value) == "" {
		return gorm.ErrInvalidData
	}
	return nil
}
