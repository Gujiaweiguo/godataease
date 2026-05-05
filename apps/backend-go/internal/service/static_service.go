package service

import (
	"dataease/backend/internal/domain/static"
	"dataease/backend/internal/repository"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const defaultStaticResourceDir = "/opt/dataease2.0/data/static-resource/"

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

// SaveFilesToServe parses static resource JSON and writes base64 files to disk.
func (s *StaticService) SaveFilesToServe(staticResourceJSON string) {
	if strings.TrimSpace(staticResourceJSON) == "" {
		return
	}

	var resources map[string]string
	if err := json.Unmarshal([]byte(staticResourceJSON), &resources); err != nil {
		return
	}

	staticDir := os.Getenv("STATIC_RESOURCE_DIR")
	if staticDir == "" {
		staticDir = defaultStaticResourceDir
	}

	_ = os.MkdirAll(staticDir, 0o755)

	for path, content := range resources {
		if strings.TrimSpace(content) == "" {
			continue
		}

		fileName := path
		if idx := strings.LastIndex(path, "/"); idx >= 0 {
			fileName = path[idx+1:]
		}
		if fileName == "" {
			continue
		}

		filePath := filepath.Join(staticDir, fileName)
		if _, err := os.Stat(filePath); err == nil {
			continue
		}

		data, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			continue
		}
		_ = os.WriteFile(filePath, data, 0o644)
	}
}
