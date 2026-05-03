package repository

import (
	"context"
	"strings"

	"dataease/backend/internal/domain/auto"
	thresholddomain "dataease/backend/internal/domain/threshold"

	"gorm.io/gorm"
)

type ThresholdRepositoryInterface interface {
	Create(ctx context.Context, info *auto.XpackThresholdInfo) error
	Update(ctx context.Context, info *auto.XpackThresholdInfo) error
	GetByID(ctx context.Context, id int64) (*auto.XpackThresholdInfo, error)
	DeleteByIDs(ctx context.Context, ids []int64) error
	DeleteByChartID(ctx context.Context, chartID int64) error
	UpdateEnable(ctx context.Context, id int64, enable bool) error
	UpdateRecipients(ctx context.Context, ids []int64, users, roles, emails, larkGroups, larksuiteGroups, webhooks string) error
	Pager(ctx context.Context, req *thresholddomain.GridRequest, goPage, pageSize int) ([]*thresholddomain.GridVO, int64, error)
	ExistsByChartID(ctx context.Context, chartID int64) (bool, error)
	InstancePager(ctx context.Context, req *thresholddomain.InstanceRequest, goPage, pageSize int) ([]*thresholddomain.InstanceVO, int64, error)
}

var _ ThresholdRepositoryInterface = (*ThresholdRepository)(nil)

type ThresholdRepository struct {
	db *gorm.DB
}

type thresholdInstancePagerRow struct {
	auto.XpackThresholdInstance
	Name string `gorm:"column:name"`
}

func NewThresholdRepository(db *gorm.DB) *ThresholdRepository {
	return &ThresholdRepository{db: db}
}

func (r *ThresholdRepository) Create(ctx context.Context, info *auto.XpackThresholdInfo) error {
	return r.db.WithContext(ctx).Create(info).Error
}

func (r *ThresholdRepository) Update(ctx context.Context, info *auto.XpackThresholdInfo) error {
	return r.db.WithContext(ctx).Save(info).Error
}

func (r *ThresholdRepository) GetByID(ctx context.Context, id int64) (*auto.XpackThresholdInfo, error) {
	var row auto.XpackThresholdInfo
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *ThresholdRepository) DeleteByIDs(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&auto.XpackThresholdInfo{}).Error
}

func (r *ThresholdRepository) DeleteByChartID(ctx context.Context, chartID int64) error {
	return r.db.WithContext(ctx).Where("chart_id = ?", chartID).Delete(&auto.XpackThresholdInfo{}).Error
}

func (r *ThresholdRepository) UpdateEnable(ctx context.Context, id int64, enable bool) error {
	return r.db.WithContext(ctx).Model(&auto.XpackThresholdInfo{}).Where("id = ?", id).Update("enable", enable).Error
}

func (r *ThresholdRepository) UpdateRecipients(ctx context.Context, ids []int64, users, roles, emails, larkGroups, larksuiteGroups, webhooks string) error {
	if len(ids) == 0 {
		return nil
	}
	updates := map[string]any{
		"reci_users":            users,
		"reci_roles":            roles,
		"reci_emails":           emails,
		"reci_lark_groups":      larkGroups,
		"reci_larksuite_groups": larksuiteGroups,
		"reci_webhooks":         webhooks,
	}
	return r.db.WithContext(ctx).Model(&auto.XpackThresholdInfo{}).Where("id IN ?", ids).Updates(updates).Error
}

func (r *ThresholdRepository) Pager(ctx context.Context, req *thresholddomain.GridRequest, goPage, pageSize int) ([]*thresholddomain.GridVO, int64, error) {
	goPage, pageSize = normalizePage(goPage, pageSize)
	query := r.db.WithContext(ctx).Model(&auto.XpackThresholdInfo{})

	if req != nil {
		if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
			query = query.Where("name LIKE ?", "%"+keyword+"%")
		}
		if len(req.ResourceTypeList) > 0 {
			query = query.Where("resource_type IN ?", req.ResourceTypeList)
		}
		if statusValues := intFlagsToBools(req.StatusList); len(statusValues) > 0 {
			query = query.Where("status IN ?", statusValues)
		}
		if enableValues := intFlagsToBools(req.EnableList); len(enableValues) > 0 {
			query = query.Where("enable IN ?", enableValues)
		}
		if req.ChartID != nil {
			query = query.Where("chart_id = ?", *req.ChartID)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := make([]auto.XpackThresholdInfo, 0)
	if err := query.Order("create_time DESC").Offset((goPage - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	result := make([]*thresholddomain.GridVO, 0, len(rows))
	for _, row := range rows {
		result = append(result, &thresholddomain.GridVO{
			ID:           row.ID,
			Name:         row.Name,
			ResourceID:   row.ResourceID,
			ResourceType: row.ResourceType,
			ResourceName: "",
			ChartID:      row.ChartID,
			ChartType:    row.ChartType,
			ChartName:    "",
			Status:       row.Status,
			Enable:       row.Enable,
			Creator:      row.Creator,
			CreateName:   row.CreatorName,
			CreateTime:   row.CreateTime,
		})
	}

	return result, total, nil
}

func (r *ThresholdRepository) ExistsByChartID(ctx context.Context, chartID int64) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&auto.XpackThresholdInfo{}).Where("chart_id = ?", chartID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ThresholdRepository) InstancePager(ctx context.Context, req *thresholddomain.InstanceRequest, goPage, pageSize int) ([]*thresholddomain.InstanceVO, int64, error) {
	goPage, pageSize = normalizePage(goPage, pageSize)
	query := r.db.WithContext(ctx).
		Model(&auto.XpackThresholdInstance{}).
		Select("xpack_threshold_instance.*, xpack_threshold_info.name AS name").
		Joins("LEFT JOIN xpack_threshold_info ON xpack_threshold_info.id = xpack_threshold_instance.task_id")

	if req != nil {
		if req.ThresholdID != nil {
			query = query.Where("xpack_threshold_instance.task_id = ?", *req.ThresholdID)
		}
		if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
			like := "%" + keyword + "%"
			query = query.Where("xpack_threshold_instance.content LIKE ? OR xpack_threshold_instance.msg LIKE ? OR xpack_threshold_info.name LIKE ?", like, like, like)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := make([]thresholdInstancePagerRow, 0)
	if err := query.Order("xpack_threshold_instance.exec_time DESC").Offset((goPage - 1) * pageSize).Limit(pageSize).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	result := make([]*thresholddomain.InstanceVO, 0, len(rows))
	for _, row := range rows {
		result = append(result, &thresholddomain.InstanceVO{
			ID:       row.ID,
			TaskID:   row.TaskID,
			Name:     row.Name,
			ExecTime: row.ExecTime,
			Status:   row.Status,
			Content:  row.Content,
			Msg:      row.Msg,
		})
	}

	return result, total, nil
}

func normalizePage(goPage, pageSize int) (int, int) {
	if goPage < 1 {
		goPage = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	return goPage, pageSize
}

func intFlagsToBools(flags []int) []bool {
	if len(flags) == 0 {
		return nil
	}
	values := make([]bool, 0, 2)
	seen := make(map[bool]struct{}, 2)
	for _, flag := range flags {
		if flag != 0 && flag != 1 {
			continue
		}
		value := flag == 1
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		values = append(values, value)
	}
	return values
}
