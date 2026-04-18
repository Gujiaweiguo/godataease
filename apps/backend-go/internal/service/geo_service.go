package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dataease/backend/internal/domain/areamap"
	"dataease/backend/internal/domain/geo"
	"dataease/backend/internal/repository"
)

const geoPrefix = "geo_"

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

func (s *GeoService) SaveMapGeo(code, name, pid string, fileContent []byte, fileName string) error {
	busiCode := code
	if strings.HasPrefix(code, geoPrefix) {
		busiCode = code[len(geoPrefix):]
	}

	exists, err := s.repo.CheckAreaExists(busiCode)
	if err != nil {
		return fmt.Errorf("failed to check area: %w", err)
	}
	if exists {
		return fmt.Errorf("area code [%s] already exists in built-in areas", busiCode)
	}

	daoCode := code
	if !strings.HasPrefix(code, geoPrefix) {
		daoCode = geoPrefix + code
	}
	existing, err := s.repo.GetCustomAreaByID(daoCode)
	if err == nil && existing != nil {
		return fmt.Errorf("area code [%s] already exists for [%s]", busiCode, existing.Name)
	}

	area := &areamap.CoreAreaCustom{
		ID:   daoCode,
		Pid:  pid,
		Name: name,
	}
	if err := s.repo.SaveCustomArea(area); err != nil {
		return fmt.Errorf("failed to save area: %w", err)
	}

	geoDir := os.Getenv("GEO_DIR")
	if geoDir == "" {
		homeDir, _ := os.UserHomeDir()
		geoDir = filepath.Join(homeDir, "geo")
	}
	countryCode := busiCode[:3]
	dirPath := filepath.Join(geoDir, countryCode)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create geo directory: %w", err)
	}
	filePath := filepath.Join(dirPath, busiCode+".json")
	if err := os.WriteFile(filePath, fileContent, 0644); err != nil {
		return fmt.Errorf("failed to write geo file: %w", err)
	}

	return nil
}

func (s *GeoService) DeleteGeo(code string) error {
	if !strings.HasPrefix(code, geoPrefix) {
		return fmt.Errorf("built-in geometry, cannot delete")
	}

	area, err := s.repo.GetCustomAreaByID(code)
	if err != nil || area == nil {
		return fmt.Errorf("geometry code does not exist")
	}

	allIDs := []string{code}
	s.collectChildIDs(code, &allIDs)

	if err := s.repo.DeleteCustomAreasBatch(allIDs); err != nil {
		return fmt.Errorf("failed to delete areas: %w", err)
	}

	geoDir := os.Getenv("GEO_DIR")
	if geoDir == "" {
		homeDir, _ := os.UserHomeDir()
		geoDir = filepath.Join(homeDir, "geo")
	}
	for _, id := range allIDs {
		busiCode := id
		if strings.HasPrefix(id, geoPrefix) {
			busiCode = id[len(geoPrefix):]
		}
		if len(busiCode) >= 3 {
			countryCode := busiCode[:3]
			filePath := filepath.Join(geoDir, countryCode, busiCode+".json")
			os.Remove(filePath)
		}
	}

	return nil
}

func (s *GeoService) collectChildIDs(pid string, ids *[]string) {
	children, err := s.repo.GetCustomAreaChildren(pid)
	if err != nil || len(children) == 0 {
		return
	}
	for _, child := range children {
		*ids = append(*ids, child.ID)
		s.collectChildIDs(child.ID, ids)
	}
}
