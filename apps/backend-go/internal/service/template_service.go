package service

import (
	"strconv"
	"strings"

	"dataease/backend/internal/domain/template"
	"dataease/backend/internal/repository"

	"gorm.io/gorm"
)

type TemplateService struct {
	repo *repository.TemplateRepository
}

const (
	checkResultNone     = "none"
	checkResultExistAll = "existAll"
)

func NewTemplateService(repo *repository.TemplateRepository) *TemplateService {
	return &TemplateService{repo: repo}
}

func (s *TemplateService) CreateTemplate(req *template.TemplateCreateRequest, createBy string) (*template.Template, error) {
	t := &template.Template{
		Name:          req.Name,
		Pid:           req.Pid,
		Level:         req.Level,
		DvType:        req.DvType,
		NodeType:      req.NodeType,
		CreateBy:      createBy,
		Snapshot:      req.Snapshot,
		TemplateType:  req.TemplateType,
		TemplateStyle: req.TemplateStyle,
		TemplateData:  req.TemplateData,
		DynamicData:   req.DynamicData,
		AppData:       req.AppData,
		UseCount:      0,
		Version:       3,
	}

	if err := s.repo.Create(t); err != nil {
		return nil, err
	}

	return t, nil
}

func (s *TemplateService) GetTemplate(id int64) (*template.Template, error) {
	t, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TemplateService) ListTemplates(req *template.TemplateListRequest) (*template.TemplateListResponse, error) {
	if strings.TrimSpace(req.CategoryID) != "" {
		categoryID := strings.TrimSpace(req.CategoryID)
		list, err := s.repo.ListByCategory(categoryID, req.DvType)
		if err != nil {
			return nil, err
		}

		total, err := s.repo.CountByCategory(categoryID, req.DvType)
		if err != nil {
			return nil, err
		}

		return &template.TemplateListResponse{
			List:  list,
			Total: total,
		}, nil
	}

	var pid int64
	if req.Pid != "" {
		p, err := strconv.ParseInt(req.Pid, 10, 64)
		if err == nil {
			pid = p
		}
	}

	list, err := s.repo.List(pid, req.DvType)
	if err != nil {
		return nil, err
	}

	total, err := s.repo.Count(pid, req.DvType)
	if err != nil {
		return nil, err
	}

	return &template.TemplateListResponse{
		List:  list,
		Total: total,
	}, nil
}

func (s *TemplateService) SaveTemplate(req *template.TemplateSaveRequest, createBy string) (*template.Template, error) {
	if req.ID > 0 {
		updated, err := s.UpdateTemplate(&template.TemplateUpdateRequest{
			ID:            req.ID,
			Name:          req.Name,
			Snapshot:      req.Snapshot,
			TemplateStyle: req.TemplateStyle,
			TemplateData:  req.TemplateData,
			DynamicData:   req.DynamicData,
			AppData:       req.AppData,
		})
		if err != nil {
			return nil, err
		}
		if req.NodeType != template.NodeTypeFolder && len(req.Categories) > 0 {
			if err := s.repo.SyncTemplateCategories(req.ID, req.Categories); err != nil {
				return nil, err
			}
			if firstCategoryID, err := strconv.ParseInt(req.Categories[0], 10, 64); err == nil && firstCategoryID > 0 {
				if err := s.repo.UpdateTemplatePid(req.ID, firstCategoryID); err != nil {
					return nil, err
				}
				updated.Pid = firstCategoryID
			}
		}
		return updated, nil
	}

	created, err := s.CreateTemplate(&template.TemplateCreateRequest{
		Name:          req.Name,
		Pid:           req.Pid,
		Level:         req.Level,
		DvType:        req.DvType,
		NodeType:      req.NodeType,
		Snapshot:      req.Snapshot,
		TemplateType:  req.TemplateType,
		TemplateStyle: req.TemplateStyle,
		TemplateData:  req.TemplateData,
		DynamicData:   req.DynamicData,
		AppData:       req.AppData,
	}, createBy)
	if err != nil {
		return nil, err
	}
	if req.NodeType != template.NodeTypeFolder && len(req.Categories) > 0 {
		if err := s.repo.SyncTemplateCategories(created.ID, req.Categories); err != nil {
			return nil, err
		}
	}
	return created, nil
}

func (s *TemplateService) UpdateTemplate(req *template.TemplateUpdateRequest) (*template.Template, error) {
	t, err := s.repo.GetByID(req.ID)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		t.Name = req.Name
	}
	if req.Snapshot != "" {
		t.Snapshot = req.Snapshot
	}
	if req.TemplateStyle != "" {
		t.TemplateStyle = req.TemplateStyle
	}
	if req.TemplateData != "" {
		t.TemplateData = req.TemplateData
	}
	if req.DynamicData != "" {
		t.DynamicData = req.DynamicData
	}
	if req.AppData != "" {
		t.AppData = req.AppData
	}

	if err := s.repo.Update(t); err != nil {
		return nil, err
	}

	return t, nil
}

func (s *TemplateService) DeleteTemplate(id int64) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}

	return s.repo.Delete(id)
}

func (s *TemplateService) DeleteWithCategory(id int64, categoryID string) error {
	if strings.TrimSpace(categoryID) == "" || categoryID == "0" {
		return s.DeleteTemplate(id)
	}

	remaining, err := s.repo.UnlinkCategory(id, categoryID)
	if err != nil {
		return err
	}
	if remaining > 0 {
		return nil
	}

	return s.repo.Delete(id)
}

func (s *TemplateService) IncrementUseCount(id int64) error {
	return s.repo.IncrementUseCount(id)
}

func (s *TemplateService) NameCheck(optType string, name string, id string) (string, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return checkResultNone, nil
	}
	var excludeID *int64
	if strings.EqualFold(optType, "update") && strings.TrimSpace(id) != "" {
		if parsed, err := strconv.ParseInt(id, 10, 64); err == nil && parsed > 0 {
			excludeID = &parsed
		}
	}
	count, err := s.repo.CountByName(trimmedName, excludeID)
	if err != nil {
		return "", err
	}
	if count == 0 {
		return checkResultNone, nil
	}
	return checkResultExistAll, nil
}

func (s *TemplateService) CategoryTemplateNameCheck(name string, categories []string, templateNames []string, templateArray []string) (string, error) {
	if len(templateNames) > 0 {
		count, err := s.repo.CountBatchNamesInCategories(templateNames, categories, templateArray)
		if err != nil {
			return "", err
		}
		if count == 0 {
			return checkResultNone, nil
		}
		return checkResultExistAll, nil
	}
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" || len(categories) == 0 {
		return checkResultNone, nil
	}
	count, err := s.repo.CountByNameInCategories(trimmedName, categories)
	if err != nil {
		return "", err
	}
	if count == 0 {
		return checkResultNone, nil
	}
	return checkResultExistAll, nil
}

func (s *TemplateService) ListCategories(level string, templateType string) ([]template.Template, error) {
	parsedLevel := 0
	if strings.TrimSpace(level) != "" {
		if value, err := strconv.Atoi(level); err == nil {
			parsedLevel = value
		}
	}
	return s.repo.ListCategories(parsedLevel, templateType)
}

func (s *TemplateService) FindCategoriesByTemplateIDs(templateIDs []string) ([]string, error) {
	return s.repo.FindCategoryIDsByTemplateIDs(templateIDs)
}

func (s *TemplateService) BatchUpdateCategories(templateIDs []string, categories []string) error {
	if len(templateIDs) == 0 || len(categories) == 0 {
		return nil
	}
	firstCategoryID, err := strconv.ParseInt(categories[0], 10, 64)
	if err != nil || firstCategoryID <= 0 {
		firstCategoryID = 0
	}
	for _, rawID := range templateIDs {
		templateID, parseErr := strconv.ParseInt(rawID, 10, 64)
		if parseErr != nil || templateID <= 0 {
			continue
		}
		if err := s.repo.SyncTemplateCategories(templateID, categories); err != nil {
			return err
		}
		if firstCategoryID > 0 {
			if err := s.repo.UpdateTemplatePid(templateID, firstCategoryID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *TemplateService) DeleteCategory(categoryID string) (bool, error) {
	count, err := s.repo.CountByCategory(categoryID, "")
	if err != nil {
		return false, err
	}
	if count > 0 {
		return false, nil
	}
	if err := s.repo.DeleteCategory(categoryID); err != nil {
		return false, err
	}
	return true, nil
}
