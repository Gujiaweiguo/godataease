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

func NewTemplateService(repo *repository.TemplateRepository) *TemplateService {
	return &TemplateService{repo: repo}
}

func (s *TemplateService) CreateTemplate(req *template.TemplateCreateRequest, createBy string) (*template.Template, error) {
	t := &template.Template{
		Name:          req.Name,
		Pid:           req.Pid,
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

func (s *TemplateService) IncrementUseCount(id int64) error {
	return s.repo.IncrementUseCount(id)
}

func (s *TemplateService) NameCheck(optType string, name string, id string) (string, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return "none", nil
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
		return "none", nil
	}
	return "existAll", nil
}

func (s *TemplateService) CategoryTemplateNameCheck(name string, categories []string, templateNames []string, templateArray []string) (string, error) {
	if len(templateNames) > 0 {
		count, err := s.repo.CountBatchNamesInCategories(templateNames, categories, templateArray)
		if err != nil {
			return "", err
		}
		if count == 0 {
			return "none", nil
		}
		return "existAll", nil
	}
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" || len(categories) == 0 {
		return "none", nil
	}
	count, err := s.repo.CountByNameInCategories(trimmedName, categories)
	if err != nil {
		return "", err
	}
	if count == 0 {
		return "none", nil
	}
	return "existAll", nil
}
