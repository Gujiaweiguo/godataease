package repository

import (
	"dataease/backend/internal/domain/org"
	"gorm.io/gorm"
)

type OrgRepository struct {
	db *gorm.DB
}

func NewOrgRepository(db *gorm.DB) *OrgRepository {
	return &OrgRepository{db: db}
}

func (r *OrgRepository) Create(o *org.SysOrg) error {
	return r.db.Create(o).Error
}

func (r *OrgRepository) Update(o *org.SysOrg) error {
	return r.db.Save(o).Error
}

func (r *OrgRepository) Delete(orgID int64) error {
	return r.db.Model(&org.SysOrg{}).
		Where("org_id = ?", orgID).
		Update("del_flag", org.DelFlagDeleted).Error
}

func (r *OrgRepository) GetByID(orgID int64) (*org.SysOrg, error) {
	var o org.SysOrg
	err := r.db.Where("org_id = ? AND del_flag = ?", orgID, org.DelFlagNormal).
		First(&o).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OrgRepository) GetByName(orgName string) (*org.SysOrg, error) {
	var o org.SysOrg
	err := r.db.Where("org_name = ? AND del_flag = ?", orgName, org.DelFlagNormal).
		First(&o).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OrgRepository) List() ([]*org.SysOrg, error) {
	var orgs []*org.SysOrg
	err := r.db.Where("del_flag = ?", org.DelFlagNormal).
		Order("level ASC, create_time ASC").
		Find(&orgs).Error
	return orgs, err
}

func (r *OrgRepository) ListByParentID(parentID int64) ([]*org.SysOrg, error) {
	var orgs []*org.SysOrg
	err := r.db.Where("parent_id = ? AND del_flag = ?", parentID, org.DelFlagNormal).
		Order("create_time ASC").
		Find(&orgs).Error
	return orgs, err
}

func (r *OrgRepository) CheckNameExists(orgName string, excludeOrgID int64) (int64, error) {
	var count int64
	query := r.db.Model(&org.SysOrg{}).
		Where("org_name = ? AND del_flag = ?", orgName, org.DelFlagNormal)
	if excludeOrgID > 0 {
		query = query.Where("org_id != ?", excludeOrgID)
	}
	err := query.Count(&count).Error
	return count, err
}

func (r *OrgRepository) CountChildren(orgID int64) (int64, error) {
	var count int64
	err := r.db.Model(&org.SysOrg{}).
		Where("parent_id = ? AND del_flag = ?", orgID, org.DelFlagNormal).
		Count(&count).Error
	return count, err
}

func (r *OrgRepository) GetByIDs(ids []int64) ([]*org.SysOrg, error) {
	var orgs []*org.SysOrg
	err := r.db.Where("org_id IN ? AND del_flag = ?", ids, org.DelFlagNormal).
		Find(&orgs).Error
	return orgs, err
}
