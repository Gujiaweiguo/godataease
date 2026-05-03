package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/pkg/logger"

	"go.uber.org/zap"
)

type RowPermissionStore interface {
	PagerByDatasetID(datasetID int64, page, size int) ([]*permission.DataPermRow, int64, error)
	PagerByDatasetIDAndTarget(datasetID int64, targetType string, targetID int64, page, size int) ([]*permission.DataPermRow, int64, error)
	GetByID(id int64) (*permission.DataPermRow, error)
	Create(perm *permission.DataPermRow) error
	Update(perm *permission.DataPermRow) error
	Delete(id int64) error
}

type ColumnPermissionStore interface {
	PagerByDatasetID(datasetID int64, page, size int) ([]*permission.DataPermColumn, int64, error)
	GetByID(id int64) (*permission.DataPermColumn, error)
	Create(perm *permission.DataPermColumn) error
	Update(perm *permission.DataPermColumn) error
	Delete(id int64) error
}

type DatasetFieldProvider interface {
	ListByDQ(datasetGroupID int64, chartID int64) (*chart.ChartFieldListResponse, error)
}

type DataPermissionAdminService struct {
	rowStore    RowPermissionStore
	columnStore ColumnPermissionStore
	fieldSource DatasetFieldProvider
	cache       *permission.PermissionCacheService
}

const (
	maskRuleAll        = "all"
	maskRuleCustom     = "custom"
	maskRuleKeepEnds   = "keep_ends"
	maskRuleKeepMiddle = "keep_middle"
)

type DataPermissionPage struct {
	List    interface{} `json:"list"`
	Total   int64       `json:"total"`
	Current int         `json:"current"`
	Size    int         `json:"size"`
}

type RowPermissionForm struct {
	ID          int64   `json:"id"`
	DatasetID   int64   `json:"datasetId"`
	Name        string  `json:"name"`
	FilterType  string  `json:"filterType"`
	TargetID    int64   `json:"targetId"`
	FilterField string  `json:"filterField"`
	FilterValue string  `json:"filterValue"`
	WhiteList   []int64 `json:"whiteList"`
}

type ColumnPermissionForm struct {
	ID        int64  `json:"id"`
	DatasetID int64  `json:"datasetId"`
	FieldName string `json:"fieldName"`
	FieldType string `json:"fieldType"`
	RuleType  string `json:"ruleType"`
	MaskRule  string `json:"maskRule"`
	MaskStart int    `json:"maskStart"`
	MaskEnd   int    `json:"maskEnd"`
}

type DeletePermissionRequest struct {
	ID int64 `json:"id"`
}

func NewDataPermissionAdminService(
	rowStore RowPermissionStore,
	columnStore ColumnPermissionStore,
	fieldSource DatasetFieldProvider,
	cache *permission.PermissionCacheService,
) *DataPermissionAdminService {
	return &DataPermissionAdminService{
		rowStore:    rowStore,
		columnStore: columnStore,
		fieldSource: fieldSource,
		cache:       cache,
	}
}

func (s *DataPermissionAdminService) RowPermissionPage(datasetID int64, page, size int) (*DataPermissionPage, error) {
	rows, total, err := s.rowStore.PagerByDatasetID(datasetID, page, size)
	if err != nil {
		return nil, err
	}
	return s.buildRowPermissionPage(datasetID, rows, total, page, size)
}

func (s *DataPermissionAdminService) RowPermissionPageByTarget(datasetID int64, targetType string, targetID int64, page, size int) (*DataPermissionPage, error) {
	if !isSupportedRowPermissionTargetType(targetType) {
		return nil, unsupportedRowPermissionTargetTypeError("targetType", targetType)
	}
	if targetID <= 0 {
		return nil, fmt.Errorf("targetId is required")
	}

	rows, total, err := s.rowStore.PagerByDatasetIDAndTarget(datasetID, targetType, targetID, page, size)
	if err != nil {
		return nil, err
	}
	return s.buildRowPermissionPage(datasetID, rows, total, page, size)
}

func (s *DataPermissionAdminService) buildRowPermissionPage(datasetID int64, rows []*permission.DataPermRow, total int64, page, size int) (*DataPermissionPage, error) {
	fieldsByID, _, err := s.datasetFieldMaps(datasetID)
	if err != nil {
		return nil, err
	}

	items := make([]RowPermissionForm, 0, len(rows))
	for _, row := range rows {
		item := RowPermissionForm{
			ID:         row.ID,
			DatasetID:  row.DatasetID,
			FilterType: row.AuthTargetType,
			TargetID:   row.AuthTargetID,
			WhiteList:  []int64{},
		}

		fieldID, filterValue, err := decodeSimpleExpression(row.ExpressionTree)
		if err == nil {
			if field, ok := fieldsByID[fieldID]; ok {
				item.FilterField = displayFieldName(field)
			}
			item.FilterValue = filterValue
		}

		item.Name = buildRuleName(item.FilterField, item.FilterValue, row.ID)
		items = append(items, item)
	}

	return &DataPermissionPage{List: items, Total: total, Current: normalizePage(page), Size: normalizeSize(size)}, nil
}

func (s *DataPermissionAdminService) SaveRowPermission(req *RowPermissionForm) error {
	if req.DatasetID <= 0 {
		return fmt.Errorf("datasetId is required")
	}
	if req.TargetID <= 0 {
		return fmt.Errorf("targetId is required")
	}
	if !isSupportedRowPermissionTargetType(req.FilterType) {
		return unsupportedRowPermissionTargetTypeError("filterType", req.FilterType)
	}
	if strings.TrimSpace(req.FilterField) == "" {
		return fmt.Errorf("filterField is required")
	}
	if len(req.WhiteList) > 0 {
		return fmt.Errorf("whiteList is deferred and not supported in permission center")
	}

	_, fieldsByName, err := s.datasetFieldMaps(req.DatasetID)
	if err != nil {
		return err
	}

	fieldID, ok := fieldsByName[strings.TrimSpace(req.FilterField)]
	if !ok {
		return fmt.Errorf("dataset field %s not found", req.FilterField)
	}

	expressionTree, err := encodeSimpleExpression(fieldID.ID, req.FilterValue)
	if err != nil {
		return err
	}

	if req.ID > 0 {
		row, err := s.rowStore.GetByID(req.ID)
		if err != nil {
			return err
		}
		row.DatasetID = req.DatasetID
		row.DatasetGroupID = req.DatasetID
		row.AuthTargetType = req.FilterType
		row.AuthTargetID = req.TargetID
		row.ExpressionTree = expressionTree
		row.Status = 1
		return s.rowStore.Update(row)
	}

	return s.rowStore.Create(&permission.DataPermRow{
		DatasetID:      req.DatasetID,
		DatasetGroupID: req.DatasetID,
		AuthTargetType: req.FilterType,
		AuthTargetID:   req.TargetID,
		ExpressionTree: expressionTree,
		Status:         1,
	})
}

func (s *DataPermissionAdminService) DeleteRowPermission(id int64) error {
	if id <= 0 {
		return fmt.Errorf("id is required")
	}
	return s.rowStore.Delete(id)
}

func isSupportedRowPermissionTargetType(targetType string) bool {
	return targetType == permission.AuthTargetTypeUser || targetType == permission.AuthTargetTypeRole
}

func unsupportedRowPermissionTargetTypeError(fieldName, targetType string) error {
	if strings.EqualFold(targetType, "sysParams") {
		return fmt.Errorf("%s sysParams is deferred and not supported in permission center", fieldName)
	}
	return fmt.Errorf("%s %s is not supported", fieldName, targetType)
}

func (s *DataPermissionAdminService) ColumnPermissionPage(datasetID int64, page, size int) (*DataPermissionPage, error) {
	columns, total, err := s.columnStore.PagerByDatasetID(datasetID, page, size)
	if err != nil {
		return nil, err
	}

	_, fieldsByName, err := s.datasetFieldMaps(datasetID)
	if err != nil {
		return nil, err
	}

	items := make([]ColumnPermissionForm, 0, len(columns))
	for _, column := range columns {
		item := ColumnPermissionForm{
			ID:        column.ID,
			DatasetID: column.DatasetID,
			FieldName: column.FieldName,
			RuleType:  column.PermType,
		}
		if field, ok := fieldsByName[column.FieldName]; ok {
			item.FieldType = field.Type
		}
		applyMaskRuleToForm(&item, column.MaskRule)
		items = append(items, item)
	}

	return &DataPermissionPage{List: items, Total: total, Current: normalizePage(page), Size: normalizeSize(size)}, nil
}

func (s *DataPermissionAdminService) SaveColumnPermission(req *ColumnPermissionForm) error {
	if req.DatasetID <= 0 {
		return fmt.Errorf("datasetId is required")
	}
	if strings.TrimSpace(req.FieldName) == "" {
		return fmt.Errorf("fieldName is required")
	}
	if req.RuleType != permission.PermTypeDisable && req.RuleType != permission.PermTypeMask {
		return fmt.Errorf("ruleType %s is not supported", req.RuleType)
	}

	_, fieldsByName, err := s.datasetFieldMaps(req.DatasetID)
	if err != nil {
		return err
	}
	if _, ok := fieldsByName[strings.TrimSpace(req.FieldName)]; !ok {
		return fmt.Errorf("dataset field %s not found", req.FieldName)
	}

	maskRule, err := encodeMaskRule(req)
	if err != nil {
		return err
	}

	if req.ID > 0 {
		column, err := s.columnStore.GetByID(req.ID)
		if err != nil {
			return err
		}
		column.DatasetID = req.DatasetID
		column.DatasetGroupID = req.DatasetID
		column.FieldName = req.FieldName
		column.PermType = req.RuleType
		column.MaskRule = maskRule
		column.Status = 1
		if err := s.columnStore.Update(column); err != nil {
			return err
		}
		s.invalidateColumnPermissionCache(req.DatasetID)
		return nil
	}

	if err := s.columnStore.Create(&permission.DataPermColumn{
		DatasetID:      req.DatasetID,
		DatasetGroupID: req.DatasetID,
		FieldName:      req.FieldName,
		PermType:       req.RuleType,
		MaskRule:       maskRule,
		Status:         1,
	}); err != nil {
		return err
	}
	s.invalidateColumnPermissionCache(req.DatasetID)
	return nil
}

func (s *DataPermissionAdminService) DeleteColumnPermission(id int64) error {
	if id <= 0 {
		return fmt.Errorf("id is required")
	}
	column, err := s.columnStore.GetByID(id)
	if err != nil {
		return err
	}
	if column == nil {
		return fmt.Errorf("column permission %d not found", id)
	}
	if err := s.columnStore.Delete(id); err != nil {
		return err
	}
	s.invalidateColumnPermissionCache(column.DatasetID)
	return nil
}

func (s *DataPermissionAdminService) invalidateColumnPermissionCache(datasetID int64) {
	if s.cache == nil || datasetID <= 0 {
		return
	}
	if err := s.cache.InvalidateColumnPermissions(context.Background(), datasetID); err != nil {
		logger.Warn("Failed to invalidate column permission cache", zap.Int64("datasetId", datasetID), zap.Error(err))
	}
}

func (s *DataPermissionAdminService) datasetFieldMaps(datasetID int64) (map[int64]chart.ChartField, map[string]chart.ChartField, error) {
	fields, err := s.fieldSource.ListByDQ(datasetID, 0)
	if err != nil {
		return nil, nil, err
	}

	byID := make(map[int64]chart.ChartField)
	byName := make(map[string]chart.ChartField)
	all := append(fields.DimensionList, fields.QuotaList...)
	for _, field := range all {
		byID[field.ID] = field
		for _, key := range []string{field.Name, field.OriginName, field.DataeaseName, field.FieldShortName} {
			trimmed := strings.TrimSpace(key)
			if trimmed != "" {
				byName[trimmed] = field
			}
		}
	}

	return byID, byName, nil
}

func encodeSimpleExpression(fieldID int64, filterValue string) (string, error) {
	obj := permission.DatasetRowPermissionsTreeObj{
		Logic: "OR",
		Items: []permission.DatasetRowPermissionsTreeItem{{
			Type:       permission.NodeTypeItem,
			FieldID:    fieldID,
			FilterType: "logic",
			Term:       permission.OperatorEq,
			Value:      filterValue,
		}},
	}
	bytes, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func decodeSimpleExpression(expressionTree string) (int64, string, error) {
	if strings.TrimSpace(expressionTree) == "" {
		return 0, "", nil
	}

	var obj permission.DatasetRowPermissionsTreeObj
	if err := json.Unmarshal([]byte(expressionTree), &obj); err != nil {
		return 0, "", err
	}
	if len(obj.Items) == 0 {
		return 0, "", nil
	}
	first := obj.Items[0]
	return first.FieldID, first.Value, nil
}

func encodeMaskRule(req *ColumnPermissionForm) (string, error) {
	if req.RuleType != permission.PermTypeMask {
		return "", nil
	}

	rule := permission.DesensitizationRule{}
	switch req.MaskRule {
	case "keep_ends":
		rule.BuiltInRule = permission.BuiltInRuleKeepFirstAndLastThree
	case maskRuleKeepMiddle:
		rule.BuiltInRule = permission.BuiltInRuleKeepMiddleThree
	case maskRuleCustom:
		rule.BuiltInRule = permission.BuiltInRuleCustom
		rule.CustomBuiltInRule = permission.CustomRuleRetainBeforeMAndAfterN
		rule.M = req.MaskStart
		rule.N = req.MaskEnd
	default:
		rule.BuiltInRule = permission.BuiltInRuleCompleteDesensitization
	}

	bytes, err := json.Marshal(rule)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func applyMaskRuleToForm(item *ColumnPermissionForm, raw string) {
	if strings.TrimSpace(raw) == "" {
		item.MaskRule = maskRuleAll
		return
	}

	var rule permission.DesensitizationRule
	if err := json.Unmarshal([]byte(raw), &rule); err != nil {
		item.MaskRule = "all"
		return
	}

	switch rule.BuiltInRule {
	case permission.BuiltInRuleKeepFirstAndLastThree:
		item.MaskRule = maskRuleKeepEnds
	case permission.BuiltInRuleKeepMiddleThree:
		item.MaskRule = maskRuleKeepMiddle
	case permission.BuiltInRuleCustom:
		item.MaskRule = maskRuleCustom
		item.MaskStart = rule.M
		item.MaskEnd = rule.N
	default:
		item.MaskRule = maskRuleAll
	}
}

func displayFieldName(field chart.ChartField) string {
	for _, value := range []string{field.OriginName, field.Name, field.DataeaseName, field.FieldShortName} {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return fmt.Sprintf("field_%d", field.ID)
}

func buildRuleName(fieldName, filterValue string, id int64) string {
	if fieldName != "" && filterValue != "" {
		return fmt.Sprintf("%s = %s", fieldName, filterValue)
	}
	if fieldName != "" {
		return fieldName
	}
	return fmt.Sprintf("规则-%d", id)
}

func normalizePage(page int) int {
	if page < 1 {
		return 1
	}
	return page
}

func normalizeSize(size int) int {
	if size < 1 {
		return 10
	}
	return size
}
