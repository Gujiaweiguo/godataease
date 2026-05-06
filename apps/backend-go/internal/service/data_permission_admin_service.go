package service

import (
	"context"
	"dataease/backend/internal/pkg/errno"
	"encoding/json"
	"fmt"
	"strings"

	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type datasetOrgScopeValidator interface {
	DatasetBelongsToOrg(datasetID, orgID int64) (bool, error)
}

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
	rowStore         RowPermissionStore
	columnStore      ColumnPermissionStore
	fieldSource      DatasetFieldProvider
	cache            *permission.PermissionCacheService
	deferredRegistry *permission.DeferredDimensionRegistry
	auditor          permissionMutationAuditor
	datasetOrgScope  datasetOrgScopeValidator
	adminChecker     AdminChecker
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
	opts ...DataPermissionAdminServiceOption,
) *DataPermissionAdminService {
	svc := &DataPermissionAdminService{
		rowStore:         rowStore,
		columnStore:      columnStore,
		fieldSource:      fieldSource,
		cache:            cache,
		deferredRegistry: permission.NewDeferredDimensionRegistry(),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(svc)
		}
	}
	if svc.auditor == nil {
		if dbProvider, ok := rowStore.(interface{ DB() *gorm.DB }); ok {
			svc.auditor = newPermAuditHelperFromDB(dbProvider.DB())
		} else if dbProvider, ok := columnStore.(interface{ DB() *gorm.DB }); ok {
			svc.auditor = newPermAuditHelperFromDB(dbProvider.DB())
		}
	}
	return svc
}

type DataPermissionAdminServiceOption func(*DataPermissionAdminService)

func WithDataPermissionAuditor(auditor permissionMutationAuditor) DataPermissionAdminServiceOption {
	return func(s *DataPermissionAdminService) {
		s.auditor = auditor
	}
}

func WithDatasetOrgScopeValidator(validator datasetOrgScopeValidator) DataPermissionAdminServiceOption {
	return func(s *DataPermissionAdminService) {
		s.datasetOrgScope = validator
	}
}

func WithDataPermissionAdminChecker(adminChecker AdminChecker) DataPermissionAdminServiceOption {
	return func(s *DataPermissionAdminService) {
		s.adminChecker = adminChecker
	}
}

func (s *DataPermissionAdminService) RowPermissionPage(datasetID int64, page, size int, scopes ...PermissionMutationScope) (*DataPermissionPage, error) {
	scope := resolvePermissionScope(scopes)
	if err := s.requireDatasetScope(datasetID, scope); err != nil {
		return nil, err
	}
	rows, total, err := s.rowStore.PagerByDatasetID(datasetID, page, size)
	if err != nil {
		return nil, err
	}
	return s.buildRowPermissionPage(datasetID, rows, total, page, size)
}

func (s *DataPermissionAdminService) RowPermissionPageByTarget(datasetID int64, targetType string, targetID int64, page, size int, scopes ...PermissionMutationScope) (*DataPermissionPage, error) {
	scope := resolvePermissionScope(scopes)
	if err := s.requireDatasetScope(datasetID, scope); err != nil {
		return nil, err
	}
	if !isSupportedRowPermissionTargetType(targetType) {
		return nil, s.unsupportedRowPermissionTargetTypeError("targetType", targetType)
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

func (s *DataPermissionAdminService) SaveRowPermission(req *RowPermissionForm, scopes ...PermissionMutationScope) error {
	scope := resolvePermissionScope(scopes)
	if err := s.requireDatasetScope(req.DatasetID, scope); err != nil {
		return err
	}
	if req.DatasetID <= 0 {
		return fmt.Errorf("datasetId is required")
	}
	if req.TargetID <= 0 {
		return fmt.Errorf("targetId is required")
	}
	if !isSupportedRowPermissionTargetType(req.FilterType) {
		return s.unsupportedRowPermissionTargetTypeError("filterType", req.FilterType)
	}
	if strings.TrimSpace(req.FilterField) == "" {
		return fmt.Errorf("filterField is required")
	}
	if len(req.WhiteList) > 0 {
		return s.deferredRegistry.GetRejectionError("whiteList")
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
		if err := s.rowStore.Update(row); err != nil {
			return err
		}
		return s.recordPermissionAudit("SAVE_ROW_PERM", scope, req.DatasetID, req.TargetID, map[string]interface{}{"filterType": req.FilterType, "filterField": req.FilterField, "mode": "update"})
	}

	if err := s.rowStore.Create(&permission.DataPermRow{
		DatasetID:      req.DatasetID,
		DatasetGroupID: req.DatasetID,
		AuthTargetType: req.FilterType,
		AuthTargetID:   req.TargetID,
		ExpressionTree: expressionTree,
		Status:         1,
	}); err != nil {
		return err
	}
	return s.recordPermissionAudit("SAVE_ROW_PERM", scope, req.DatasetID, req.TargetID, map[string]interface{}{"filterType": req.FilterType, "filterField": req.FilterField, "mode": "create"})
}

func (s *DataPermissionAdminService) DeleteRowPermission(id int64) error {
	if id <= 0 {
		return fmt.Errorf(errno.ErrIDRequired)
	}
	return s.rowStore.Delete(id)
}

func isSupportedRowPermissionTargetType(targetType string) bool {
	return targetType == permission.AuthTargetTypeUser || targetType == permission.AuthTargetTypeRole
}

func (s *DataPermissionAdminService) unsupportedRowPermissionTargetTypeError(fieldName, targetType string) error {
	if s.deferredRegistry.IsDeferred(targetType) {
		return s.deferredRegistry.GetRejectionError(targetType)
	}
	return fmt.Errorf("%s %s is not supported", fieldName, targetType)
}

func (s *DataPermissionAdminService) ColumnPermissionPage(datasetID int64, page, size int, scopes ...PermissionMutationScope) (*DataPermissionPage, error) {
	scope := resolvePermissionScope(scopes)
	if err := s.requireDatasetScope(datasetID, scope); err != nil {
		return nil, err
	}
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

func (s *DataPermissionAdminService) SaveColumnPermission(req *ColumnPermissionForm, scopes ...PermissionMutationScope) error {
	scope := resolvePermissionScope(scopes)
	if err := s.requireDatasetScope(req.DatasetID, scope); err != nil {
		return err
	}
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
		return s.recordPermissionAudit("SAVE_COLUMN_PERM", scope, req.DatasetID, 0, map[string]interface{}{"fieldName": req.FieldName, "ruleType": req.RuleType, "mode": "update"})
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
	return s.recordPermissionAudit("SAVE_COLUMN_PERM", scope, req.DatasetID, 0, map[string]interface{}{"fieldName": req.FieldName, "ruleType": req.RuleType, "mode": "create"})
}

func (s *DataPermissionAdminService) requireDatasetScope(datasetID int64, scope PermissionMutationScope) error {
	if datasetID <= 0 || scope.OrgID <= 0 || s.isAdminScope(scope) {
		return nil
	}
	if err := requireOrgScope(scope); err != nil {
		return err
	}
	if s.datasetOrgScope == nil {
		return requireDatasetOrgValidator()
	}
	allowed, err := s.datasetOrgScope.DatasetBelongsToOrg(datasetID, scope.OrgID)
	if err != nil {
		return err
	}
	if !allowed {
		return fmt.Errorf("dataset does not belong to current organization")
	}
	return nil
}

func (s *DataPermissionAdminService) isAdminScope(scope PermissionMutationScope) bool {
	return scope.ActorID > 0 && s.adminChecker != nil && s.adminChecker.IsAdmin(scope.ActorID)
}

func (s *DataPermissionAdminService) recordPermissionAudit(operation string, scope PermissionMutationScope, datasetID, targetID int64, details map[string]interface{}) error {
	if s.auditor == nil {
		return nil
	}
	return s.auditor.RecordPermissionMutationAudit(operation, scope, "dataset", targetID, 0, datasetID, details)
}

func (s *DataPermissionAdminService) DeleteColumnPermission(id int64) error {
	if id <= 0 {
		return fmt.Errorf(errno.ErrIDRequired)
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
