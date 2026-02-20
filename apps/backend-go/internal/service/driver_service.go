package service

import (
	"dataease/backend/internal/domain/driver"
	"dataease/backend/internal/repository"
)

type DriverService struct {
	repo *repository.DriverRepository
}

func NewDriverService(repo *repository.DriverRepository) *DriverService {
	return &DriverService{repo: repo}
}

func (s *DriverService) List() ([]driver.DriverDTO, error) {
	list, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	result := make([]driver.DriverDTO, 0, len(list))
	for _, d := range list {
		result = append(result, driver.DriverDTO{
			ID:       d.ID,
			Name:     d.Name,
			Type:     d.Type,
			TypeDesc: d.TypeDesc,
			Desc:     d.Desc,
		})
	}
	return result, nil
}

func (s *DriverService) ListByType(dsType string) ([]driver.DriverDTO, error) {
	list, err := s.repo.ListByType(dsType)
	if err != nil {
		return nil, err
	}
	result := make([]driver.DriverDTO, 0, len(list))
	for _, d := range list {
		result = append(result, driver.DriverDTO{
			ID:       d.ID,
			Name:     d.Name,
			Type:     d.Type,
			TypeDesc: d.TypeDesc,
			Desc:     d.Desc,
		})
	}
	return result, nil
}

func (s *DriverService) GetByID(id int64) (*driver.DriverDTO, error) {
	d, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return &driver.DriverDTO{
		ID:       d.ID,
		Name:     d.Name,
		Type:     d.Type,
		TypeDesc: d.TypeDesc,
		Desc:     d.Desc,
	}, nil
}

func (s *DriverService) ListDriverJars(driverID int64) ([]driver.DriverJarDTO, error) {
	list, err := s.repo.ListDriverJars(driverID)
	if err != nil {
		return nil, err
	}
	result := make([]driver.DriverJarDTO, 0, len(list))
	for _, j := range list {
		result = append(result, driver.DriverJarDTO{
			ID:       j.ID,
			DriverID: j.DriverID,
			FileName: j.FileName,
			FilePath: j.FilePath,
			Version:  j.Version,
		})
	}
	return result, nil
}
