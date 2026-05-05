package repository

import (
	"fmt"
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

func (r *VisualizationRepository) CopyChartViews(sourceDvID, newDvID, copyID int64, resourceTable string) error {
	sourceTable := "core_chart_view"
	targetTable := "core_chart_view"
	if resourceTable == "snapshot" {
		sourceTable = "snapshot_core_chart_view"
		targetTable = "snapshot_core_chart_view"
	}

	sql := fmt.Sprintf(`
		INSERT INTO %s (id, title, scene_id, table_id, type, render, result_count, result_mode,
			x_axis, x_axis_ext, y_axis, y_axis_ext, ext_stack, ext_bubble, ext_label, ext_tooltip,
			custom_attr, custom_attr_mobile, custom_style, custom_style_mobile, custom_filter,
			drill_fields, senior, create_by, create_time, update_time, snapshot, style_priority,
			chart_type, is_plugin, data_from, view_fields, refresh_view_enable, refresh_unit, refresh_time,
			linkage_active, jump_active, copy_from, copy_id, flow_map_start_name, flow_map_end_name, ext_color)
		SELECT ccv.id + ?, title, ? AS scene_id, table_id, type, render, result_count, result_mode,
			x_axis, x_axis_ext, y_axis, y_axis_ext, ext_stack, ext_bubble, ext_label, ext_tooltip,
			custom_attr, custom_attr_mobile, custom_style, custom_style_mobile, custom_filter,
			drill_fields, senior, create_by, create_time, update_time, snapshot, style_priority,
			chart_type, is_plugin, data_from, view_fields, refresh_view_enable, refresh_unit, refresh_time,
			linkage_active, jump_active, ccv.id AS copy_from, ? AS copy_id,
			flow_map_start_name, flow_map_end_name, ext_color
		FROM %s ccv WHERE ccv.scene_id = ?`, targetTable, sourceTable)

	return r.db.Exec(sql, copyID, newDvID, copyID, sourceDvID).Error
}

func (r *VisualizationRepository) GetCopiedChartViewMapping(copyID int64) (map[int64]int64, error) {
	type viewMapping struct {
		CopyFrom int64 `gorm:"column:copy_from"`
		ID       int64 `gorm:"column:id"`
	}

	var rows []viewMapping
	if err := r.db.Raw(`SELECT copy_from, id FROM core_chart_view WHERE copy_id = ?`, copyID).Scan(&rows).Error; err != nil {
		return nil, err
	}

	mapping := make(map[int64]int64, len(rows))
	for _, row := range rows {
		mapping[row.CopyFrom] = row.ID
	}
	return mapping, nil
}

func (r *VisualizationRepository) CopyLinkages(copyID int64) error {
	return r.db.Exec(`
		INSERT INTO visualization_linkage (id, dv_id, source_view_id, target_view_id, update_time, update_people, linkage_active, ext1, ext2, copy_from, copy_id)
		SELECT vl.id + ?, pv_source.t_dv_id, pv_source.t_chart_view_id, pv_target.t_chart_view_id, vl.update_time, vl.update_people, vl.linkage_active, vl.ext1, vl.ext2, vl.id, ?
		FROM visualization_linkage vl
		INNER JOIN (SELECT pvs.scene_id AS s_dv_id, pvs.id AS s_chart_view_id, pvt.scene_id AS t_dv_id, pvt.id AS t_chart_view_id
			FROM core_chart_view pvt INNER JOIN core_chart_view pvs ON pvt.copy_from = pvs.id WHERE pvt.copy_id = ?) pv_source
			ON vl.dv_id = pv_source.s_dv_id AND vl.source_view_id = pv_source.s_chart_view_id
		INNER JOIN (SELECT pvs.scene_id AS s_dv_id, pvs.id AS s_chart_view_id, pvt.scene_id AS t_dv_id, pvt.id AS t_chart_view_id
			FROM core_chart_view pvt INNER JOIN core_chart_view pvs ON pvt.copy_from = pvs.id WHERE pvt.copy_id = ?) pv_target
			ON vl.dv_id = pv_target.s_dv_id AND vl.target_view_id = pv_target.s_chart_view_id`, copyID, copyID, copyID, copyID).Error
}

func (r *VisualizationRepository) CopyLinkageFields(copyID int64) error {
	return r.db.Exec(`
		INSERT INTO visualization_linkage_field (id, linkage_id, source_field, target_field, update_time, copy_from, copy_id)
		SELECT vlf.id + ?, pvlf_copy.t_id, vlf.source_field, vlf.target_field, vlf.update_time, vlf.id, ?
		FROM visualization_linkage_field vlf
		INNER JOIN (SELECT id AS t_id, copy_from AS s_id FROM visualization_linkage WHERE copy_id = ?) pvlf_copy
			ON vlf.linkage_id = pvlf_copy.s_id`, copyID, copyID, copyID).Error
}

func (r *VisualizationRepository) CopyLinkJumps(copyID int64) error {
	return r.db.Exec(`
		INSERT INTO visualization_link_jump (id, source_dv_id, source_view_id, link_jump_info, checked, copy_from, copy_id)
		SELECT vlj.id + ?, dv_view_copy.t_dv_id, dv_view_copy.t_chart_view_id, vlj.link_jump_info, vlj.checked, vlj.id, ?
		FROM visualization_link_jump vlj
		INNER JOIN (SELECT pvs.scene_id AS s_dv_id, pvs.id AS s_chart_view_id, pvt.scene_id AS t_dv_id, pvt.id AS t_chart_view_id
			FROM core_chart_view pvt INNER JOIN core_chart_view pvs ON pvt.copy_from = pvs.id WHERE pvt.copy_id = ?) dv_view_copy
			ON vlj.source_dv_id = dv_view_copy.s_dv_id AND vlj.source_view_id = dv_view_copy.s_chart_view_id`, copyID, copyID, copyID).Error
}

func (r *VisualizationRepository) CopyLinkJumpInfos(copyID int64) error {
	return r.db.Exec(`
		INSERT INTO visualization_link_jump_info (id, link_jump_id, link_type, jump_type, target_dv_id, source_field_id, content, checked, attach_params, copy_from, copy_id)
		SELECT vlji.id + ?, plj_copy.t_id, vlji.link_type, vlji.jump_type, vlji.target_dv_id, vlji.source_field_id, vlji.content, vlji.checked, vlji.attach_params, vlji.id, ?
		FROM visualization_link_jump_info vlji
		INNER JOIN (SELECT id AS t_id, copy_from AS s_id FROM visualization_link_jump WHERE copy_id = ?) plj_copy
			ON vlji.link_jump_id = plj_copy.s_id`, copyID, copyID, copyID).Error
}

func (r *VisualizationRepository) CopyLinkJumpTargetInfos(copyID int64) error {
	return r.db.Exec(`
		INSERT INTO visualization_link_jump_target_view_info (target_id, link_jump_info_id, source_field_active_id, target_view_id, target_field_id, copy_from, copy_id)
		SELECT vljtvi.target_id + ?, plji_copy.t_id, vljtvi.source_field_active_id, vljtvi.target_view_id, vljtvi.target_field_id, vljtvi.target_id, ?
		FROM visualization_link_jump_target_view_info vljtvi
		INNER JOIN (SELECT id AS t_id, copy_from AS s_id FROM visualization_link_jump_info WHERE copy_id = ?) plji_copy
			ON vljtvi.link_jump_info_id = plji_copy.s_id`, copyID, copyID, copyID).Error
}
