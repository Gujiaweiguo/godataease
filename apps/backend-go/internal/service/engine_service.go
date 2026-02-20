package service

import (
	"dataease/backend/internal/domain/engine"
	"dataease/backend/internal/repository"

	"gorm.io/gorm"
)

type EngineService struct {
	repo *repository.EngineRepository
}

func NewEngineService(repo *repository.EngineRepository) *EngineService {
	return &EngineService{repo: repo}
}

func (s *EngineService) GetEngine() (*engine.EngineDTO, error) {
	e, err := s.repo.Get()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &engine.EngineDTO{
		ID:            e.ID,
		Name:          e.Name,
		Type:          e.Type,
		Configuration: e.Configuration,
	}, nil
}

func (s *EngineService) Validate(req *engine.ValidateRequest) (*engine.ValidateResponse, error) {
	return &engine.ValidateResponse{
		Status:  "Success",
		Message: "Engine validation not implemented",
	}, nil
}

func (s *EngineService) ValidateByID(id int64) (*engine.ValidateResponse, error) {
	return &engine.ValidateResponse{
		Status:  "Success",
		Message: "Engine validation not implemented",
	}, nil
}

func (s *EngineService) SupportSetKey() (bool, error) {
	return false, nil
}
