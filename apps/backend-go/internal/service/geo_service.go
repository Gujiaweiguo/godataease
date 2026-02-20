package service

import (
	"dataease/backend/internal/domain/geo"
	"dataease/backend/internal/repository"
)

type GeoService struct {
	repo *repository.GeoRepository
}

func NewGeoService(repo *repository.GeoRepository) *GeoService {
	return &GeoService{repo: repo}
}

func (s *GeoService) ListAreas() ([]*geo.GeometryArea, error) {
	return s.repo.ListAreas()
}

func (s *GeoService) GetArea(id string) (*geo.GeometryArea, error) {
	return s.repo.GetAreaByID(id)
}
