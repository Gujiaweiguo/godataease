package repository

import (
	"strconv"
	"time"

	"dataease/backend/internal/domain/template"

	"gorm.io/gorm"
)

type coreVisualizationTemplate struct {
	ID            int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name          string     `gorm:"column:name;size:255" json:"name"`
	Pid           int64      `gorm:"column:pid;index" json:"pid"`
	Level         int        `gorm:"column:level" json:"level"`
	DvType        string     `gorm:"column:dv_type;size:50" json:"dvType"`
	NodeType      string     `gorm:"column:node_type;size:50" json:"nodeType"`
	CreateBy      string     `gorm:"column:create_by;size:255" json:"createBy"`
	CreateTime    *time.Time `gorm:"column:create_time" json:"createTime"`
	Snapshot      string     `gorm:"column:snapshot;type:longtext" json:"snapshot"`
	TemplateType  string     `gorm:"column:template_type;size:50" json:"templateType"`
	TemplateStyle string     `gorm:"column:template_style;type:longtext" json:"templateStyle"`
	TemplateData  string     `gorm:"column:template_data;type:longtext" json:"templateData"`
	DynamicData   string     `gorm:"column:dynamic_data;type:longtext" json:"dynamicData"`
	AppData       string     `gorm:"column:app_data;type:longtext" json:"appData"`
	UseCount      int        `gorm:"column:use_count;default:0" json:"useCount"`
	Version       int        `gorm:"column:version;default:3" json:"version"`
}

func (coreVisualizationTemplate) TableName() string {
	return "core_visualization_template"
}

type TemplateRepository struct {
	db *gorm.DB
}

func NewTemplateRepository(db *gorm.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

func (r *TemplateRepository) Create(t *template.Template) error {
	now := time.Now()
	record := coreVisualizationTemplate{
		Name:          t.Name,
		Pid:           t.Pid,
		Level:         t.Level,
		DvType:        t.DvType,
		NodeType:      t.NodeType,
		CreateBy:      t.CreateBy,
		CreateTime:    &now,
		Snapshot:      t.Snapshot,
		TemplateType:  t.TemplateType,
		TemplateStyle: t.TemplateStyle,
		TemplateData:  t.TemplateData,
		DynamicData:   t.DynamicData,
		AppData:       t.AppData,
		UseCount:      t.UseCount,
		Version:       t.Version,
	}
	if err := r.db.Create(&record).Error; err != nil {
		return err
	}
	t.ID = record.ID
	t.CreateTime = record.CreateTime
	return nil
}

func (r *TemplateRepository) GetByID(id int64) (*template.Template, error) {
	var record coreVisualizationTemplate
	if err := r.db.Where("id = ?", id).First(&record).Error; err != nil {
		return nil, err
	}
	return r.toTemplate(record), nil
}

func (r *TemplateRepository) List(pid int64, dvType string) ([]template.Template, error) {
	var records []coreVisualizationTemplate
	query := r.db.Model(&coreVisualizationTemplate{})
	if pid > 0 {
		query = query.Where("pid = ?", pid)
	}
	if dvType != "" {
		query = query.Where("dv_type = ?", dvType)
	}
	if err := query.Order("create_time desc").Find(&records).Error; err != nil {
		return nil, err
	}
	result := make([]template.Template, len(records))
	for i, record := range records {
		result[i] = *r.toTemplate(record)
	}
	return result, nil
}

func (r *TemplateRepository) Update(t *template.Template) error {
	return r.db.Model(&coreVisualizationTemplate{}).Where("id = ?", t.ID).Updates(map[string]interface{}{
		"name":           t.Name,
		"snapshot":       t.Snapshot,
		"template_style": t.TemplateStyle,
		"template_data":  t.TemplateData,
		"dynamic_data":   t.DynamicData,
		"app_data":       t.AppData,
	}).Error
}

func (r *TemplateRepository) Delete(id int64) error {
	return r.db.Delete(&coreVisualizationTemplate{}, id).Error
}

func (r *TemplateRepository) Count(pid int64, dvType string) (int64, error) {
	var count int64
	query := r.db.Model(&coreVisualizationTemplate{})
	if pid > 0 {
		query = query.Where("pid = ?", pid)
	}
	if dvType != "" {
		query = query.Where("dv_type = ?", dvType)
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *TemplateRepository) IncrementUseCount(id int64) error {
	return r.db.Model(&coreVisualizationTemplate{}).Where("id = ?", id).
		UpdateColumn("use_count", gorm.Expr("use_count + 1")).Error
}

func (r *TemplateRepository) CountByName(name string, excludeID *int64) (int64, error) {
	var count int64
	q := r.db.Model(&coreVisualizationTemplate{}).Where("name = ?", name)
	if excludeID != nil && *excludeID > 0 {
		q = q.Where("id <> ?", *excludeID)
	}
	err := q.Count(&count).Error
	return count, err
}

func (r *TemplateRepository) CountByNameInCategories(name string, categories []string) (int64, error) {
	ids, err := r.templateIDsByCategories(categories)
	if err != nil || len(ids) == 0 {
		return 0, err
	}
	var count int64
	err = r.db.Model(&coreVisualizationTemplate{}).Where("name = ? AND id IN ?", name, ids).Count(&count).Error
	return count, err
}

func (r *TemplateRepository) CountBatchNamesInCategories(names []string, categories []string, excludeTemplateIDs []string) (int64, error) {
	ids, err := r.templateIDsByCategories(categories)
	if err != nil || len(ids) == 0 || len(names) == 0 {
		return 0, err
	}
	excludeIDs := make([]int64, 0, len(excludeTemplateIDs))
	for _, raw := range excludeTemplateIDs {
		if parsed, parseErr := strconv.ParseInt(raw, 10, 64); parseErr == nil && parsed > 0 {
			excludeIDs = append(excludeIDs, parsed)
		}
	}
	var count int64
	q := r.db.Model(&coreVisualizationTemplate{}).Where("name IN ? AND id IN ?", names, ids)
	if len(excludeIDs) > 0 {
		q = q.Where("id NOT IN ?", excludeIDs)
	}
	err = q.Count(&count).Error
	return count, err
}

func (r *TemplateRepository) templateIDsByCategories(categories []string) ([]int64, error) {
	if len(categories) == 0 {
		return nil, nil
	}
	var rawIDs []string
	err := r.db.Table("visualization_template_category_map").
		Where("category_id IN ?", categories).
		Distinct("template_id").
		Pluck("template_id", &rawIDs).Error
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(rawIDs))
	for _, raw := range rawIDs {
		parsed, parseErr := strconv.ParseInt(raw, 10, 64)
		if parseErr == nil && parsed > 0 {
			ids = append(ids, parsed)
		}
	}
	return ids, nil
}

func (r *TemplateRepository) toTemplate(record coreVisualizationTemplate) *template.Template {
	return &template.Template{
		ID:            record.ID,
		Name:          record.Name,
		Pid:           record.Pid,
		Level:         record.Level,
		DvType:        record.DvType,
		NodeType:      record.NodeType,
		CreateBy:      record.CreateBy,
		CreateTime:    record.CreateTime,
		Snapshot:      record.Snapshot,
		TemplateType:  record.TemplateType,
		TemplateStyle: record.TemplateStyle,
		TemplateData:  record.TemplateData,
		DynamicData:   record.DynamicData,
		AppData:       record.AppData,
		UseCount:      record.UseCount,
		Version:       record.Version,
	}
}
