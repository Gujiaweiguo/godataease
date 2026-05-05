package repository

import (
	"strings"
	"time"

	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/visualization"

	"gorm.io/gorm"
)

type VisualizationRepository struct {
	db *gorm.DB
}

func NewVisualizationRepository(db *gorm.DB) *VisualizationRepository {
	return &VisualizationRepository{db: db}
}

func (r *VisualizationRepository) Create(v *visualization.DataVisualizationInfo) error {
	return r.db.Create(v).Error
}

func (r *VisualizationRepository) Update(v *visualization.DataVisualizationInfo) error {
	return r.db.Save(v).Error
}

func (r *VisualizationRepository) GetByID(id int64) (*visualization.DataVisualizationInfo, error) {
	var item visualization.DataVisualizationInfo
	err := r.db.Model(&visualization.DataVisualizationInfo{}).
		Where("id = ? AND COALESCE(delete_flag, 0) = 0", id).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *VisualizationRepository) DeleteLogic(id int64, deletedBy string) error {
	now := time.Now().UnixMilli()
	return r.db.Model(&visualization.DataVisualizationInfo{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"delete_flag": true,
			"delete_time": now,
			"delete_by":   deletedBy,
			"update_time": now,
			"update_by":   deletedBy,
		}).Error
}

func (r *VisualizationRepository) Query(req *visualization.ListRequest) ([]*visualization.DataVisualizationInfo, int64, error) {
	var list []*visualization.DataVisualizationInfo
	var total int64

	page := req.Current
	if page < 1 {
		page = 1
	}
	size := req.Size
	if size < 1 {
		size = 10
	}
	if size > 100 {
		size = 100
	}
	offset := (page - 1) * size

	q := r.db.Model(&visualization.DataVisualizationInfo{}).
		Where("COALESCE(delete_flag, 0) = 0")
	if req.Keyword != nil && *req.Keyword != "" {
		kw := "%" + *req.Keyword + "%"
		q = q.Where("name LIKE ?", kw)
	}
	if req.Type != nil && *req.Type != "" {
		q = q.Where("type = ?", *req.Type)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("update_time DESC").Offset(offset).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *VisualizationRepository) FindRecent(uid int64, req *visualization.WorkbranchQueryRequest) ([]visualization.VisualizationResourceVO, error) {
	baseSQL := `
		SELECT dvResource.id, dvResource.resource_id, dvResource.name, dvResource.ext_flag,
		       dvResource.type, dvResource.creator,
		       core_opt_recent.uid AS last_editor, core_opt_recent.time AS last_edit_time,
		       CASE WHEN core_store.resource_id IS NULL THEN 0 ELSE 1 END AS favorite,
		       0 AS weight
		FROM (
		    SELECT id, id AS resource_id, name, 0 AS ext_flag, 'dataset' AS type, create_by AS creator
		    FROM core_dataset_group WHERE node_type = 'dataset'
		    UNION ALL
		    SELECT id, id AS resource_id, name, 0 AS ext_flag, 'datasource' AS type, create_by AS creator
		    FROM core_datasource WHERE type != 'folder'
		    UNION ALL
		    SELECT id, id AS resource_id, name, COALESCE(CAST(mobile_layout AS SIGNED), 0) AS ext_flag,
		           CASE WHEN type = 'dataV' THEN 'screen' ELSE 'panel' END AS type, create_by AS creator
		    FROM data_visualization_info WHERE COALESCE(delete_flag, 0) = 0 AND node_type = 'leaf' AND status != 0
		) dvResource
		INNER JOIN core_opt_recent ON dvResource.resource_id = core_opt_recent.resource_id AND core_opt_recent.uid = ?
		LEFT JOIN core_store ON dvResource.id = core_store.resource_id AND core_store.uid = ?`

	query := strings.Builder{}
	query.WriteString(baseSQL)

	args := []interface{}{uid, uid}
	conditions := make([]string, 0, 2)
	if req != nil {
		typeMap := map[string]string{
			"panel":      "panel",
			"screen":     "screen",
			"dataset":    "dataset",
			"datasource": "datasource",
		}
		if resourceType, ok := typeMap[req.Type]; ok {
			conditions = append(conditions, "dvResource.type = ?")
			args = append(args, resourceType)
		}
		if req.Keyword != "" {
			conditions = append(conditions, "LOWER(dvResource.name) LIKE LOWER(CONCAT('%', ?, '%'))")
			args = append(args, req.Keyword)
		}
	}
	if len(conditions) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(conditions, " AND "))
	}
	query.WriteString(" ORDER BY core_opt_recent.time ")
	if req != nil && req.Asc {
		query.WriteString("ASC")
	} else {
		query.WriteString("DESC")
	}

	var results []visualization.VisualizationResourceVO
	err := r.db.Raw(query.String(), args...).Scan(&results).Error
	return results, err
}

func (r *VisualizationRepository) ListAllByTypes(types []string) ([]*visualization.DataVisualizationInfo, error) {
	var list []*visualization.DataVisualizationInfo

	q := r.db.Model(&visualization.DataVisualizationInfo{}).
		Where("COALESCE(delete_flag, 0) = 0")
	if len(types) > 0 {
		q = q.Where("type IN ?", types)
	}

	err := q.Order("COALESCE(pid, 0) ASC").Order("COALESCE(sort, 0) ASC").Order("update_time DESC").Find(&list).Error
	return list, err
}

func (r *VisualizationRepository) ListAllByTypesBatch(types []string, afterID int64, limit int, orgID *int64) ([]*visualization.DataVisualizationInfo, error) {
	var list []*visualization.DataVisualizationInfo

	q := r.db.Model(&visualization.DataVisualizationInfo{}).
		Where("COALESCE(delete_flag, 0) = 0")
	if len(types) > 0 {
		q = q.Where("type IN ?", types)
	}
	if afterID > 0 {
		q = q.Where("id > ?", afterID)
	}
	if orgID != nil && *orgID > 0 {
		q = q.Where("org_id = ?", *orgID)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}

	err := q.Order("id ASC").Find(&list).Error
	return list, err
}

func (r *VisualizationRepository) CountByNameAndPID(name string, pid *int64, excludeID *int64) (int64, error) {
	var count int64
	normalizedPID := int64(0)
	if pid != nil {
		normalizedPID = *pid
	}

	query := r.db.Model(&visualization.DataVisualizationInfo{}).
		Where("name = ? AND COALESCE(pid, 0) = ? AND COALESCE(delete_flag, 0) = 0", name, normalizedPID)
	if excludeID != nil && *excludeID > 0 {
		query = query.Where("id != ?", *excludeID)
	}

	err := query.Count(&count).Error
	return count, err
}

func (r *VisualizationRepository) GetChartViewsBySceneID(sceneID int64) ([]chart.CoreChartView, error) {
	var results []chart.CoreChartView
	err := r.db.Where("scene_id = ?", sceneID).Find(&results).Error
	return results, err
}

func (r *VisualizationRepository) FindChartViewsByIDs(viewIDs []int64) ([]map[string]interface{}, error) {
	if len(viewIDs) == 0 {
		return []map[string]interface{}{}, nil
	}
	var results []map[string]interface{}
	err := r.db.Table("core_chart_view").Where("id IN ?", viewIDs).Find(&results).Error
	return results, err
}

func (r *VisualizationRepository) FindDatasetGroupsByIDs(dsIDs []int64) ([]map[string]interface{}, error) {
	if len(dsIDs) == 0 {
		return []map[string]interface{}{}, nil
	}
	var results []map[string]interface{}
	err := r.db.Table("core_dataset_group").Where("id IN ?", dsIDs).Find(&results).Error
	return results, err
}

func (r *VisualizationRepository) FindDatasetTablesByGroupIDs(dsIDs []int64) ([]map[string]interface{}, error) {
	if len(dsIDs) == 0 {
		return []map[string]interface{}{}, nil
	}
	var results []map[string]interface{}
	err := r.db.Table("core_dataset_table").Where("dataset_group_id IN ?", dsIDs).Find(&results).Error
	return results, err
}

func (r *VisualizationRepository) FindDatasetTableFieldsByGroupIDs(dsIDs []int64) ([]map[string]interface{}, error) {
	if len(dsIDs) == 0 {
		return []map[string]interface{}{}, nil
	}
	var results []map[string]interface{}
	err := r.db.Table("core_dataset_table_field").Where("dataset_group_id IN ?", dsIDs).Find(&results).Error
	return results, err
}

func (r *VisualizationRepository) FindDatasourcesByGroupIDs(dsIDs []int64) ([]map[string]interface{}, error) {
	if len(dsIDs) == 0 {
		return []map[string]interface{}{}, nil
	}
	var results []map[string]interface{}
	err := r.db.Raw(`
		SELECT DISTINCT core_datasource.* FROM core_datasource
		INNER JOIN core_dataset_table ON core_dataset_table.datasource_id = core_datasource.id
		WHERE core_dataset_table.dataset_group_id IN ?`, dsIDs).Scan(&results).Error
	return results, err
}

func (r *VisualizationRepository) FindDatasourceTasksByGroupIDs(dsIDs []int64) ([]map[string]interface{}, error) {
	if len(dsIDs) == 0 {
		return []map[string]interface{}{}, nil
	}
	var results []map[string]interface{}
	err := r.db.Raw(`
		SELECT core_datasource_task.* FROM core_datasource_task
		INNER JOIN core_datasource ON core_datasource_task.ds_id = core_datasource.id
		INNER JOIN core_dataset_table ON core_dataset_table.datasource_id = core_datasource.id
		WHERE core_dataset_table.dataset_group_id IN ?`, dsIDs).Scan(&results).Error
	return results, err
}

func (r *VisualizationRepository) FindLinkagesByDvID(dvID int64) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := r.db.Table("visualization_linkage").Where("dv_id = ?", dvID).Find(&results).Error
	return results, err
}

func (r *VisualizationRepository) FindLinkageFieldsByDvID(dvID int64) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := r.db.Raw(`
		SELECT visualization_linkage_field.* FROM visualization_linkage_field
		INNER JOIN visualization_linkage ON visualization_linkage.id = visualization_linkage_field.linkage_id
		WHERE visualization_linkage.dv_id = ?`, dvID).Scan(&results).Error
	return results, err
}

func (r *VisualizationRepository) FindLinkJumpsByDvID(dvID int64) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := r.db.Table("visualization_link_jump").Where("source_dv_id = ?", dvID).Find(&results).Error
	return results, err
}

func (r *VisualizationRepository) FindLinkJumpInfosByDvID(dvID int64) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := r.db.Raw(`
		SELECT visualization_link_jump_info.* FROM visualization_link_jump_info
		INNER JOIN visualization_link_jump ON visualization_link_jump.id = visualization_link_jump_info.link_jump_id
		WHERE visualization_link_jump.source_dv_id = ?`, dvID).Scan(&results).Error
	return results, err
}

func (r *VisualizationRepository) FindLinkJumpTargetViewInfosByDvID(dvID int64) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := r.db.Raw(`
		SELECT visualization_link_jump_target_view_info.* FROM visualization_link_jump_target_view_info
		INNER JOIN visualization_link_jump_info ON visualization_link_jump_target_view_info.link_jump_info_id = visualization_link_jump_info.id
		INNER JOIN visualization_link_jump ON visualization_link_jump.id = visualization_link_jump_info.link_jump_id
		WHERE visualization_link_jump.source_dv_id = ?`, dvID).Scan(&results).Error
	return results, err
}
