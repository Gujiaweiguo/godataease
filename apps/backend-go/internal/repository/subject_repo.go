package repository

import (
	"dataease/backend/internal/domain/auto"

	"gorm.io/gorm"
)

type SubjectRepository struct {
	db *gorm.DB
}

func NewSubjectRepository(db *gorm.DB) *SubjectRepository {
	return &SubjectRepository{db: db}
}

func (r *SubjectRepository) List() ([]auto.VisualizationSubject, error) {
	var subjects []auto.VisualizationSubject
	err := r.db.Where("delete_flag = ?", false).Order("create_time ASC").Find(&subjects).Error
	return subjects, err
}

func (r *SubjectRepository) ListAll() ([]auto.VisualizationSubject, error) {
	var subjects []auto.VisualizationSubject
	err := r.db.Order("create_time ASC").Find(&subjects).Error
	return subjects, err
}

func (r *SubjectRepository) GetByID(id string) (*auto.VisualizationSubject, error) {
	var subject auto.VisualizationSubject
	err := r.db.Where("id = ? AND delete_flag = ?", id, false).First(&subject).Error
	return &subject, err
}

func (r *SubjectRepository) FindByName(name string) (*auto.VisualizationSubject, error) {
	var subject auto.VisualizationSubject
	err := r.db.Where("name = ?", name).First(&subject).Error
	return &subject, err
}

func (r *SubjectRepository) FindByNameExcludeID(name string, excludeID string) (*auto.VisualizationSubject, error) {
	var subject auto.VisualizationSubject
	err := r.db.Where("name = ? AND id != ?", name, excludeID).First(&subject).Error
	return &subject, err
}

func (r *SubjectRepository) Create(subject *auto.VisualizationSubject) error {
	return r.db.Create(subject).Error
}

func (r *SubjectRepository) Update(subject *auto.VisualizationSubject) error {
	return r.db.Save(subject).Error
}

func (r *SubjectRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&auto.VisualizationSubject{}).Error
}
