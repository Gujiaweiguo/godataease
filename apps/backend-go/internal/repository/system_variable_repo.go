package repository

import (
	"strings"

	"dataease/backend/internal/domain/system"

	"gorm.io/gorm"
)

type SystemVariableRepository struct {
	db *gorm.DB
}

func NewSystemVariableRepository(db *gorm.DB) *SystemVariableRepository {
	return &SystemVariableRepository{db: db}
}

func (r *SystemVariableRepository) Create(variable *system.SysVariable) error {
	return r.db.Create(variable).Error
}

func (r *SystemVariableRepository) Update(variable *system.SysVariable) error {
	return r.db.Save(variable).Error
}

func (r *SystemVariableRepository) GetByID(id int64) (*system.SysVariable, error) {
	var row system.SysVariable
	if err := r.db.Where("id = ?", id).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *SystemVariableRepository) Query(req *system.SysVariableQueryRequest) ([]system.SysVariable, error) {
	query := r.db.Model(&system.SysVariable{})
	if req != nil {
		if req.ID > 0 {
			query = query.Where("id = ?", req.ID)
		}
		if strings.TrimSpace(req.Type) != "" {
			query = query.Where("type = ?", strings.TrimSpace(req.Type))
		}
		if strings.TrimSpace(req.Name) != "" {
			query = query.Where("name LIKE ?", "%"+strings.TrimSpace(req.Name)+"%")
		}
		if req.Disabled != nil {
			query = query.Where("disabled = ?", *req.Disabled)
		}
	}
	rows := make([]system.SysVariable, 0)
	if err := query.Order("id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *SystemVariableRepository) Delete(id int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("sys_variable_id = ?", id).Delete(&system.SysVariableValue{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&system.SysVariable{}).Error
	})
}

func (r *SystemVariableRepository) CreateValue(value *system.SysVariableValue) error {
	return r.db.Create(value).Error
}

func (r *SystemVariableRepository) UpdateValue(value *system.SysVariableValue) error {
	return r.db.Save(value).Error
}

func (r *SystemVariableRepository) GetValueByID(id int64) (*system.SysVariableValue, error) {
	var row system.SysVariableValue
	if err := r.db.Where("id = ?", id).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *SystemVariableRepository) ListValuesByVariableID(sysVariableID int64) ([]system.SysVariableValue, error) {
	rows := make([]system.SysVariableValue, 0)
	if err := r.db.Where("sys_variable_id = ?", sysVariableID).Order("id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *SystemVariableRepository) PageValues(page int, size int, req *system.SysVariableValueQueryRequest) ([]system.SysVariableValue, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	query := r.db.Model(&system.SysVariableValue{})
	if req != nil {
		if req.ID > 0 {
			query = query.Where("id = ?", req.ID)
		}
		if req.SysVariableID > 0 {
			query = query.Where("sys_variable_id = ?", req.SysVariableID)
		}
		if strings.TrimSpace(req.Value) != "" {
			query = query.Where("value LIKE ?", "%"+strings.TrimSpace(req.Value)+"%")
		}
		if strings.TrimSpace(req.ValueDesc) != "" {
			query = query.Where("value_desc LIKE ?", "%"+strings.TrimSpace(req.ValueDesc)+"%")
		}
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	rows := make([]system.SysVariableValue, 0)
	if err := query.Order("id ASC").Offset((page - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *SystemVariableRepository) DeleteValue(id int64) error {
	return r.db.Where("id = ?", id).Delete(&system.SysVariableValue{}).Error
}

func (r *SystemVariableRepository) BatchDeleteValues(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Where("id IN ?", ids).Delete(&system.SysVariableValue{}).Error
}
