package service

import (
	"dataease/backend/internal/domain/static"
	"dataease/backend/internal/repository"
)

type StaticService struct {
	repo         *repository.StaticRepository
	storeRepo    *repository.StoreRepository
	typefaceRepo *repository.TypefaceRepository
}

func NewStaticService(repo *repository.StaticRepository, storeRepo *repository.StoreRepository, typefaceRepo *repository.TypefaceRepository) *StaticService {
	return &StaticService{
		repo:         repo,
		storeRepo:    storeRepo,
		typefaceRepo: typefaceRepo,
	}
}

func (s *StaticService) ListResources() ([]*static.StaticResource, error) {
	return s.repo.ListResources()
}

func (s *StaticService) GetResource(id string) (*static.StaticResource, error) {
	return s.repo.GetResourceByID(id)
}

func (s *StaticService) ListStores() ([]*static.Store, error) {
	return s.storeRepo.ListStores()
}

func (s *StaticService) ListTypefaces() ([]*static.Typeface, error) {
	return s.typefaceRepo.ListTypefaces()
}
