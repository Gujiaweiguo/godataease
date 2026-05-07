package service

import (
	"context"
	"dataease/backend/internal/pkg/errno"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/dataset"
	exportdomain "dataease/backend/internal/domain/export"
	"dataease/backend/internal/domain/permission"
	calciteintegration "dataease/backend/internal/integration/calcite"
	"dataease/backend/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DatasetService struct {
	repo                    *repository.DatasetRepository
	datasourceRepo          *repository.DatasourceRepository
	exportRepo              repository.ExportRepositoryInterface
	rowPermissionService    *RowPermissionService
	columnPermissionService *ColumnPermissionService
	resourcePermService     *ResourcePermissionService
	calciteAddress          string
	calciteTimeout          time.Duration
	calciteRetries          int
	calciteClient           *calciteintegration.Client
	calciteMu               sync.Mutex
	userRepo                *repository.UserRepository
	previewExecutorFactory  PreviewExecutorFactory
}

var (
	ErrDatasetDatasourcePermissionDenied       = errors.New("insufficient datasource permissions")
	ErrDatasetViewPermissionDenied             = errors.New("insufficient dataset permissions")
	ErrDatasetFieldDependencyBlocked           = errors.New("dataset field dependency blocked")
	ErrPreviewSQLExternalDatasourceUnsupported = errors.New("external datasource SQL preview is not supported yet; please use synchronized dataset preview")
	ErrPreviewSQLTimeout                       = errors.New("preview query timed out")
	ErrPreviewSQLResultTooLarge                = errors.New("preview result is too large")
	sqlVariableDateTimePatterns                = []string{"DATETIME", "TIMESTAMP", "DATE", "TIME", "YEAR"}
	sqlVariableFloatPatterns                   = []string{"DOUBLE", "FLOAT", "DECIMAL", "NUMERIC", "REAL"}
	sqlVariableIntegerPatterns                 = []string{"INT", "LONG", "SHORT", "BIGINT", "SMALLINT", "TINYINT"}
)

const (
	previewSQLTimeout      = 5 * time.Second
	previewResultMaxCells  = 5000
	previewDefaultRowLimit = 100
)

type sqlVariableDetailRaw struct {
	VariableName string        `json:"variableName"`
	Type         []string      `json:"type"`
	Params       []interface{} `json:"params"`
}

type fieldTreeSpec struct {
	field  *dataset.CoreDatasetTableField
	alias  string
	column string
}

func NewDatasetService(repo *repository.DatasetRepository) *DatasetService {
	return &DatasetService{
		repo:                   repo,
		calciteAddress:         "",
		calciteTimeout:         10 * time.Second,
		calciteRetries:         1,
		previewExecutorFactory: defaultPreviewExecutorFactory,
	}
}

func NewDatasetServiceWithPermission(repo *repository.DatasetRepository, rowPermSvc *RowPermissionService) *DatasetService {
	return &DatasetService{
		repo:                   repo,
		rowPermissionService:   rowPermSvc,
		calciteAddress:         "",
		calciteTimeout:         10 * time.Second,
		calciteRetries:         1,
		previewExecutorFactory: defaultPreviewExecutorFactory,
	}
}

func (s *DatasetService) SetDatasourceRepository(datasourceRepo *repository.DatasourceRepository) {
	s.datasourceRepo = datasourceRepo
}

func (s *DatasetService) SetPreviewExecutorFactory(factory PreviewExecutorFactory) {
	if factory == nil {
		s.previewExecutorFactory = defaultPreviewExecutorFactory
		return
	}
	s.previewExecutorFactory = factory
}

func (s *DatasetService) SetResourcePermissionService(resourcePermSvc *ResourcePermissionService) {
	s.resourcePermService = resourcePermSvc
}

func (s *DatasetService) SetColumnPermissionService(columnPermSvc *ColumnPermissionService) {
	s.columnPermissionService = columnPermSvc
	if columnPermSvc != nil && s.rowPermissionService != nil && s.rowPermissionService.columnPermRepo == nil {
		s.rowPermissionService.columnPermRepo = columnPermSvc.columnPermRepo
	}
}

func (s *DatasetService) SetExportRepository(exportRepo repository.ExportRepositoryInterface) {
	s.exportRepo = exportRepo
}

func (s *DatasetService) SetCalciteConfig(address string, timeout time.Duration, retries int) {
	s.calciteAddress = strings.TrimSpace(address)
	if timeout > 0 {
		s.calciteTimeout = timeout
	}
	if retries >= 0 {
		s.calciteRetries = retries
	}

	s.calciteMu.Lock()
	if s.calciteClient != nil {
		_ = s.calciteClient.Close()
		s.calciteClient = nil
	}
	s.calciteMu.Unlock()
}

func (s *DatasetService) SetUserRepository(userRepo *repository.UserRepository) {
	s.userRepo = userRepo
}

// ResolveUserName resolves a user ID string to a displayable user name.
// Falls back to the raw userID if the user cannot be found.
func (s *DatasetService) ResolveUserName(userID string) string {
	if s.userRepo == nil || userID == "" {
		return userID
	}
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return userID
	}
	u, err := s.userRepo.GetByID(id)
	if err != nil {
		return userID
	}
	if u.NickName != "" {
		return u.NickName
	}
	return u.Username
}

func (s *DatasetService) Tree(req *dataset.TreeRequest) ([]dataset.TreeNode, error) {
	groups, err := s.repo.ListGroups(req.Keyword)
	if err != nil {
		return nil, err
	}

	nodesByID := make(map[int64]*dataset.TreeNode)
	childrenByPID := make(map[int64][]*dataset.TreeNode)

	for _, g := range groups {
		nodeType := dataset.NodeTypeDataset
		if g.NodeType != nil && *g.NodeType != "" {
			nodeType = *g.NodeType
		}
		n := &dataset.TreeNode{
			ID:       g.ID,
			Name:     g.Name,
			NodeType: nodeType,
		}
		nodesByID[g.ID] = n
		pid := int64(0)
		if g.PID != nil {
			pid = *g.PID
		}
		childrenByPID[pid] = append(childrenByPID[pid], n)
	}

	for id, node := range nodesByID {
		children := childrenByPID[id]
		if len(children) == 0 {
			continue
		}
		sort.Slice(children, func(i, j int) bool {
			return children[i].ID < children[j].ID
		})
		node.Children = make([]dataset.TreeNode, 0, len(children))
		for _, c := range children {
			node.Children = append(node.Children, *c)
		}
	}

	rootChildren := childrenByPID[0]
	sort.Slice(rootChildren, func(i, j int) bool {
		return rootChildren[i].ID < rootChildren[j].ID
	})
	roots := make([]dataset.TreeNode, 0, len(rootChildren))
	for _, r := range rootChildren {
		roots = append(roots, *r)
	}
	return roots, nil
}

func (s *DatasetService) GetGroupByID(id int64) (*dataset.CoreDatasetGroup, error) {
	fixedID, err := s.compatDatasetGroupID(id)
	if err != nil {
		return nil, err
	}
	return s.repo.GetGroupByID(fixedID)
}

func (s *DatasetService) compatDatasetGroupID(id int64) (int64, error) {
	if id <= 0 {
		return id, nil
	}
	if _, err := s.repo.GetGroupByID(id); err == nil {
		return id, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return id, err
	}

	if id%100 != 0 {
		return id, gorm.ErrRecordNotFound
	}

	nearestID, err := s.repo.FindNearestGroupIDInWindow(id, 100)
	if err != nil {
		return id, err
	}
	if nearestID == nil {
		return id, gorm.ErrRecordNotFound
	}
	return *nearestID, nil
}

func (s *DatasetService) Fields(req *dataset.FieldsRequest) ([]*dataset.CoreDatasetTableField, error) {
	return s.repo.ListFields(req.DatasetGroupID)
}

func (s *DatasetService) FieldsWithPermission(datasetGroupID int64, userID int64) (*chart.ChartFieldListResponse, error) {
	fields, err := s.repo.ListFields(datasetGroupID)
	if err != nil {
		return nil, err
	}

	dimensionList := make([]chart.ChartField, 0, len(fields))
	quotaList := make([]chart.ChartField, 0, len(fields)+1)
	for _, field := range fields {
		converted := convertToChartField(field)
		if strings.EqualFold(converted.GroupType, "d") {
			dimensionList = append(dimensionList, converted)
			continue
		}
		quotaList = append(quotaList, converted)
	}
	quotaList = append(quotaList, countChartField(datasetGroupID))

	resp := &chart.ChartFieldListResponse{
		DimensionList: dimensionList,
		QuotaList:     quotaList,
	}

	if s.rowPermissionService == nil || s.rowPermissionService.columnPermRepo == nil || userID <= 0 {
		return resp, nil
	}
	if s.rowPermissionService.IsAdmin(userID) {
		return resp, nil
	}

	columnSvc := s.getColumnPermissionService()
	disabledColumns, err := columnSvc.GetDisabledColumns(datasetGroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to load disabled columns: %w", err)
	}
	maskRules, err := columnSvc.GetMaskRules(datasetGroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to load mask rules: %w", err)
	}

	chartSvc := &ChartService{}
	return &chart.ChartFieldListResponse{
		DimensionList: chartSvc.filterChartFields(resp.DimensionList, disabledColumns, maskRules),
		QuotaList:     chartSvc.filterChartFields(resp.QuotaList, disabledColumns, maskRules),
	}, nil
}

func (s *DatasetService) CopilotFields(datasetGroupID int64, userID int64) (*chart.ChartFieldListResponse, error) {
	if err := s.ensureDatasetViewable(datasetGroupID, userID); err != nil {
		return nil, err
	}

	fields, err := s.repo.ListFields(datasetGroupID)
	if err != nil {
		return nil, err
	}
	resp := buildCopilotChartFieldList(fields)
	return s.applyDatasetColumnPermissions(datasetGroupID, userID, resp)
}

func (s *DatasetService) ensureDatasetViewable(datasetGroupID int64, userID int64) error {
	if datasetGroupID <= 0 {
		return gorm.ErrRecordNotFound
	}
	if _, err := s.repo.GetGroupByID(datasetGroupID); err != nil {
		return err
	}
	if s.resourcePermService != nil && userID > 0 && !s.resourcePermService.CheckViewPermission(userID, permission.ResourceTypeDataset, datasetGroupID) {
		return ErrDatasetViewPermissionDenied
	}
	return nil
}

func buildCopilotChartFieldList(fields []*dataset.CoreDatasetTableField) *chart.ChartFieldListResponse {
	dimensionList := make([]chart.ChartField, 0, len(fields))
	quotaList := make([]chart.ChartField, 0, len(fields))
	for _, field := range fields {
		if !isCopilotVisibleField(field) {
			continue
		}
		converted := convertToChartField(field)
		if strings.EqualFold(converted.GroupType, "d") {
			dimensionList = append(dimensionList, converted)
			continue
		}
		quotaList = append(quotaList, converted)
	}
	return &chart.ChartFieldListResponse{DimensionList: dimensionList, QuotaList: quotaList}
}

func isCopilotVisibleField(field *dataset.CoreDatasetTableField) bool {
	if field == nil {
		return false
	}
	if field.Checked == nil || !*field.Checked {
		return false
	}
	return field.ExtField == nil || *field.ExtField == 0
}

func (s *DatasetService) applyDatasetColumnPermissions(datasetGroupID int64, userID int64, resp *chart.ChartFieldListResponse) (*chart.ChartFieldListResponse, error) {
	if s.rowPermissionService == nil || s.rowPermissionService.columnPermRepo == nil || userID <= 0 {
		return resp, nil
	}
	if s.rowPermissionService.IsAdmin(userID) {
		return resp, nil
	}

	columnSvc := s.getColumnPermissionService()
	disabledColumns, err := columnSvc.GetDisabledColumns(datasetGroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to load disabled columns: %w", err)
	}
	maskRules, err := columnSvc.GetMaskRules(datasetGroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to load mask rules: %w", err)
	}

	chartSvc := &ChartService{}
	return &chart.ChartFieldListResponse{
		DimensionList: chartSvc.filterChartFields(resp.DimensionList, disabledColumns, maskRules),
		QuotaList:     chartSvc.filterChartFields(resp.QuotaList, disabledColumns, maskRules),
	}, nil
}

func (s *DatasetService) Preview(req *dataset.PreviewRequest) (*dataset.PreviewResponse, error) {
	limit := req.Limit
	if limit < 1 {
		limit = 100
	}

	tableName, err := s.repo.FindPrimaryTableName(req.DatasetGroupID)
	if err != nil {
		return nil, err
	}

	rows, err := s.repo.PreviewRows(tableName, limit)
	if err != nil {
		return nil, err
	}
	total, err := s.repo.CountRows(tableName)
	if err != nil {
		return nil, err
	}

	columns := make([]string, 0)
	if len(rows) > 0 {
		for k := range rows[0] {
			columns = append(columns, k)
		}
		sort.Strings(columns)
	}

	return &dataset.PreviewResponse{
		Columns: columns,
		Rows:    rows,
		Total:   total,
	}, nil
}

func (s *DatasetService) PreviewWithPermission(req *dataset.PreviewRequest, userID int64) (*dataset.PreviewResponse, error) {
	limit := req.Limit
	if limit < 1 {
		limit = 100
	}

	if err := s.ensureDatasourceDependenciesViewable(req.DatasetGroupID, userID); err != nil {
		return nil, err
	}

	tableName, err := s.repo.FindPrimaryTableName(req.DatasetGroupID)
	if err != nil {
		return nil, err
	}

	var selectColumns = "*"
	var whereClause string
	var whereArgs []interface{}

	if s.rowPermissionService != nil {
		selectColumns, err = s.rowPermissionService.BuildSelectColumns(req.DatasetGroupID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to build row permission select columns: %w", err)
		}
		whereResult, err := s.rowPermissionService.BuildWhereClause(req.DatasetGroupID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to build row permission where clause: %w", err)
		}
		if whereResult != nil {
			whereClause = whereResult.Clause
			whereArgs = whereResult.Args
		}
	}

	rows, err := s.repo.PreviewRowsWithFilter(tableName, selectColumns, whereClause, whereArgs, limit)
	if err != nil {
		return nil, err
	}
	rows, err = s.applyColumnPermissionFilter(rows, req.DatasetGroupID)
	if err != nil {
		return nil, err
	}
	total, err := s.repo.CountRows(tableName)
	if err != nil {
		return nil, err
	}

	columns := make([]string, 0)
	if len(rows) > 0 {
		for k := range rows[0] {
			columns = append(columns, k)
		}
		sort.Strings(columns)
	}

	return &dataset.PreviewResponse{
		Columns: columns,
		Rows:    rows,
		Total:   total,
	}, nil
}

func (s *DatasetService) applyColumnPermissionFilter(rows []map[string]interface{}, datasetGroupID int64) ([]map[string]interface{}, error) {
	if s.rowPermissionService == nil || s.rowPermissionService.columnPermRepo == nil {
		return rows, nil
	}
	columnSvc := s.getColumnPermissionService()
	disabledColumns, err := columnSvc.GetDisabledColumns(datasetGroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to load disabled columns: %w", err)
	}
	maskRules, err := columnSvc.GetMaskRules(datasetGroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to load mask rules: %w", err)
	}
	for i := range rows {
		rows[i] = columnSvc.FilterDisabledColumns(rows[i], disabledColumns)
		rows[i] = columnSvc.MaskRowData(rows[i], maskRules)
	}
	return rows, nil
}

func (s *DatasetService) getColumnPermissionService() *ColumnPermissionService {
	if s.columnPermissionService != nil {
		return s.columnPermissionService
	}
	if s.rowPermissionService == nil || s.rowPermissionService.columnPermRepo == nil {
		return nil
	}
	return NewColumnPermissionService(s.rowPermissionService.columnPermRepo, nil)
}

func (s *DatasetService) ensureDatasourceDependenciesViewable(datasetGroupID, userID int64) error {
	if s.resourcePermService == nil || s.repo == nil || datasetGroupID <= 0 || userID <= 0 {
		return nil
	}

	tables, err := s.repo.ListTablesByDatasetGroupID(datasetGroupID)
	if err != nil {
		return err
	}

	seen := make(map[int64]struct{})
	for _, table := range tables {
		if table == nil || table.DatasourceID == nil || *table.DatasourceID <= 0 {
			continue
		}
		datasourceID := *table.DatasourceID
		if _, ok := seen[datasourceID]; ok {
			continue
		}
		seen[datasourceID] = struct{}{}
		if !s.resourcePermService.CheckViewPermission(userID, permission.ResourceTypeDatasource, datasourceID) {
			return ErrDatasetDatasourcePermissionDenied
		}
	}

	return nil
}

func (s *DatasetService) PreviewSQL(req *dataset.SQLPreviewRequest) (map[string]interface{}, error) {
	return s.PreviewSQLWithUser(req, 0)
}

func (s *DatasetService) PreviewSQLWithUser(req *dataset.SQLPreviewRequest, userID int64) (map[string]interface{}, error) {
	empty := map[string]interface{}{
		"data": dataset.SQLPreviewData{
			Fields: []dataset.SQLPreviewField{},
			Data:   []map[string]interface{}{},
		},
		"sql": "",
	}

	if req == nil {
		return empty, nil
	}

	rawSQL := strings.TrimSpace(req.SQL)
	if rawSQL == "" {
		return empty, nil
	}
	decoded, decodeErr := base64.StdEncoding.DecodeString(rawSQL)
	if decodeErr == nil {
		rawSQL = strings.TrimSpace(string(decoded))
	}
	if rawSQL == "" {
		return empty, nil
	}

	if err := validatePreviewSQL(rawSQL); err != nil {
		return nil, err
	}

	if err := s.validateWithCalciteIfEnabled(rawSQL); err != nil {
		return nil, err
	}

	executor, err := s.resolvePreviewExecutor(req, userID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = executor.Close()
	}()
	ctx, cancel := context.WithTimeout(context.Background(), previewSQLTimeout)
	defer cancel()

	rows, err := executor.PreviewSQL(ctx, rawSQL, previewDefaultRowLimit)
	if err != nil {
		return nil, err
	}
	for i := range rows {
		rows[i] = normalizePreviewRow(rows[i])
	}

	fields := buildPreviewFields(rows)
	return map[string]interface{}{
		"data": dataset.SQLPreviewData{
			Fields: fields,
			Data:   rows,
		},
		"sql": base64.StdEncoding.EncodeToString([]byte(rawSQL)),
	}, nil
}

func isDirectPreviewRequest(req *dataset.SQLPreviewRequest) bool {
	if req == nil {
		return false
	}
	// Stage4 previewSql keeps routing semantics bound to datasource context only.
	// Compatibility metadata such as sqlVariableDetails is accepted but does not
	// change executor selection until variable execution semantics are explicitly
	// specified in a later step.
	return req.DatasourceID > 0
}

func (s *DatasetService) resolvePreviewExecutor(req *dataset.SQLPreviewRequest, userID int64) (PreviewExecutor, error) {
	if !isDirectPreviewRequest(req) {
		return &localPreviewExecutor{repo: s.repo}, nil
	}
	if err := s.ensurePreviewDatasourceViewable(req.DatasourceID, userID); err != nil {
		return nil, err
	}
	if s.datasourceRepo == nil {
		return nil, ErrPreviewSQLExternalDatasourceUnsupported
	}
	ds, err := s.datasourceRepo.GetByID(req.DatasourceID)
	if err != nil {
		return nil, fmt.Errorf(errno.ErrDatasourceNotFound)
	}
	if ds.Configuration == nil || strings.TrimSpace(*ds.Configuration) == "" {
		return nil, errors.New(errDatasourceConfigEmpty)
	}
	cfg, err := decodeConfig(*ds.Configuration)
	if err != nil {
		return nil, fmt.Errorf("invalid datasource configuration: %w", err)
	}
	factory := s.previewExecutorFactory
	if factory == nil {
		factory = defaultPreviewExecutorFactory
	}
	return factory(ds, cfg)
}

func (s *DatasetService) ensurePreviewDatasourceViewable(datasourceID, userID int64) error {
	if datasourceID <= 0 {
		return nil
	}
	if userID <= 0 {
		return ErrDatasetDatasourcePermissionDenied
	}
	if s.resourcePermService == nil {
		return nil
	}
	if !s.resourcePermService.CheckViewPermission(userID, permission.ResourceTypeDatasource, datasourceID) {
		return ErrDatasetDatasourcePermissionDenied
	}
	return nil
}

func (s *DatasetService) GetSQLParams(ids []int64) ([]dataset.SQLVariableDetails, error) {
	if len(ids) == 0 {
		return []dataset.SQLVariableDetails{}, nil
	}

	result := make([]dataset.SQLVariableDetails, 0)
	for _, datasetGroupID := range ids {
		if datasetGroupID <= 0 {
			continue
		}

		tables, err := s.repo.ListTablesByDatasetGroupID(datasetGroupID)
		if err != nil {
			return nil, err
		}
		if len(tables) == 0 {
			continue
		}

		fullName, err := s.datasetFullName(datasetGroupID)
		if err != nil {
			return nil, err
		}

		for _, table := range tables {
			if table == nil || table.SQLVariables == nil || strings.TrimSpace(*table.SQLVariables) == "" {
				continue
			}

			rawList := make([]sqlVariableDetailRaw, 0)
			if err = json.Unmarshal([]byte(*table.SQLVariables), &rawList); err != nil {
				continue
			}

			for _, raw := range rawList {
				name := strings.TrimSpace(raw.VariableName)
				if name == "" {
					continue
				}

				item := dataset.SQLVariableDetails{
					ID:              fmt.Sprintf("%d|DE|%s", table.ID, name),
					VariableName:    name,
					Type:            raw.Type,
					Params:          raw.Params,
					DatasetGroupID:  datasetGroupID,
					DatasetTableID:  table.ID,
					DatasetFullName: fullName,
					DeType:          inferSQLVariableDeType(raw.Type),
				}
				result = append(result, item)
			}
		}
	}

	return result, nil
}

func (s *DatasetService) GetFieldEnum(req *dataset.MultFieldValuesRequest) ([]string, error) {
	if req == nil || len(req.FieldIDs) == 0 {
		return []string{}, nil
	}

	limit := 1000
	if req.ResultMode == 1 {
		limit = 5000
	}

	uniqField := make(map[int64]struct{}, len(req.FieldIDs))
	result := make([]string, 0)
	seen := make(map[string]struct{})
	for _, fieldID := range req.FieldIDs {
		if fieldID <= 0 {
			continue
		}
		if _, ok := uniqField[fieldID]; ok {
			continue
		}
		uniqField[fieldID] = struct{}{}

		field, tableName, columnName, err := s.resolveEnumFieldTarget(fieldID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return nil, err
		}

		filters, err := s.buildEnumFilterClauses(req.Filter, tableName)
		if err != nil {
			return nil, err
		}
		values, err := s.repo.QueryDistinctValues(tableName, columnName, filters, limit)
		if err != nil {
			return nil, err
		}

		for _, value := range values {
			normalized := normalizeEnumValue(value, field.DeType)
			if normalized == "" {
				continue
			}
			if _, ok := seen[normalized]; ok {
				continue
			}
			seen[normalized] = struct{}{}
			result = append(result, normalized)
		}
	}
	return result, nil
}

func (s *DatasetService) GetFieldTree(req *dataset.MultFieldValuesRequest) ([]dataset.BaseTreeNode, error) {
	if req == nil || len(req.FieldIDs) == 0 {
		return []dataset.BaseTreeNode{}, nil
	}

	limit := 1000
	if req.ResultMode == 1 {
		limit = 5000
	}

	var tableName string
	specs := make([]fieldTreeSpec, 0, len(req.FieldIDs))
	for _, fieldID := range req.FieldIDs {
		if fieldID <= 0 {
			continue
		}
		field, resolvedTableName, columnName, err := s.resolveEnumFieldTarget(fieldID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return nil, err
		}
		if tableName == "" {
			tableName = resolvedTableName
		}
		if resolvedTableName != tableName {
			continue
		}
		specs = append(specs, fieldTreeSpec{
			field:  field,
			alias:  enumAlias(fieldID),
			column: columnName,
		})
	}
	if tableName == "" || len(specs) == 0 {
		return []dataset.BaseTreeNode{}, nil
	}

	filters, err := s.buildEnumFilterClauses(req.Filter, tableName)
	if err != nil {
		return nil, err
	}

	columns := make([]dataset.EnumObjectColumn, 0, len(specs))
	for _, spec := range specs {
		columns = append(columns, dataset.EnumObjectColumn{Column: spec.column, Alias: spec.alias})
	}

	rows, err := s.repo.QueryFieldTreeValues(tableName, columns, filters, limit)
	if err != nil {
		return nil, err
	}

	return buildFieldTreeNodes(rows, specs), nil
}

func (s *DatasetService) GetFieldEnumObj(req *dataset.EnumValueRequest) ([]map[string]interface{}, error) {
	if req == nil || req.QueryID <= 0 {
		return []map[string]interface{}{}, nil
	}

	queryField, tableName, queryColumn, err := s.resolveEnumFieldTarget(req.QueryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []map[string]interface{}{}, nil
		}
		return nil, err
	}

	displayID, displayField, displayColumn, err := s.resolveEnumDisplayField(req, queryField, queryColumn, tableName)
	if err != nil {
		return nil, err
	}

	columns := []dataset.EnumObjectColumn{{Column: queryColumn, Alias: enumAlias(req.QueryID)}}
	if displayID != req.QueryID {
		columns = append(columns, dataset.EnumObjectColumn{Column: displayColumn, Alias: enumAlias(displayID)})
	}

	filters, err := s.buildEnumFilterClauses(req.Filter, tableName)
	if err != nil {
		return nil, err
	}

	limit := 1000
	if req.ResultMode == 1 {
		limit = 5000
	}

	searchColumn := queryColumn
	if displayID != req.QueryID {
		searchColumn = displayColumn
	}

	sortColumn, err := s.resolveEnumSortColumn(req, tableName, searchColumn)
	if err != nil {
		return nil, err
	}

	rows, err := s.repo.QueryDistinctObjectValues(
		tableName,
		columns,
		filters,
		searchColumn,
		req.SearchText,
		sortColumn,
		req.Sort,
		limit,
	)
	if err != nil {
		return nil, err
	}

	return buildEnumResultRows(rows, columns, displayID, queryField, displayField), nil
}

func (s *DatasetService) GetFieldEnumDs(fieldID int64) ([]string, error) {
	if fieldID <= 0 {
		return []string{}, nil
	}
	return s.GetFieldEnum(&dataset.MultFieldValuesRequest{FieldIDs: []int64{fieldID}, ResultMode: 0})
}

func (s *DatasetService) PerDelete(id int64) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf(errno.ErrDatasetIDRequired)
	}
	count, err := s.repo.CountChartRelations(id)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *DatasetService) DeleteField(id int64) error {
	if id <= 0 {
		return fmt.Errorf("field id is required")
	}
	field, err := s.repo.GetFieldByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("dataset field not found")
		}
		return err
	}
	if field.ChartID == nil || *field.ChartID <= 0 {
		return fmt.Errorf("dataset field is not chart-scoped")
	}
	deps, err := s.collectFieldDeleteDependencies(field)
	if err != nil {
		return err
	}
	if len(deps) > 0 {
		return fmt.Errorf("%w: %s", ErrDatasetFieldDependencyBlocked, strings.Join(deps, ", "))
	}

	deleted, err := s.repo.DeleteFieldByIDAndChartID(id, *field.ChartID)
	if err != nil {
		return err
	}
	if !deleted {
		return fmt.Errorf("dataset field not found")
	}
	return nil
}

func (s *DatasetService) DeleteFieldByChart(chartID int64) error {
	if chartID <= 0 {
		return fmt.Errorf(errno.ErrChartIDRequired)
	}
	fields, err := s.repo.ListFieldsByChartID(chartID)
	if err != nil {
		return err
	}
	for _, field := range fields {
		deps, depErr := s.collectFieldDeleteDependencies(field)
		if depErr != nil {
			return depErr
		}
		if len(deps) > 0 {
			return fmt.Errorf("%w: %s", ErrDatasetFieldDependencyBlocked, strings.Join(deps, ", "))
		}
	}
	_, err = s.repo.DeleteFieldsByChartID(chartID)
	return err
}

func (s *DatasetService) collectFieldDeleteDependencies(field *dataset.CoreDatasetTableField) ([]string, error) {
	if field == nil {
		return nil, nil
	}
	deps := make([]string, 0)
	var err error

	deps, err = s.appendFieldLevelDependencies(deps, field)
	if err != nil {
		return nil, err
	}
	deps, err = s.appendDatasetScopedDependencies(deps, field)
	if err != nil {
		return nil, err
	}
	deps, err = s.appendVisualizationDependencies(deps, field)
	if err != nil {
		return nil, err
	}

	return deps, nil
}

func (s *DatasetService) appendFieldLevelDependencies(deps []string, field *dataset.CoreDatasetTableField) ([]string, error) {
	derivedCount, err := s.repo.CountDerivedFieldReferences(field.ID)
	if err != nil {
		return nil, err
	}
	if derivedCount > 0 {
		deps = append(deps, "derived fields")
	}
	return deps, nil
}

func (s *DatasetService) appendDatasetScopedDependencies(deps []string, field *dataset.CoreDatasetTableField) ([]string, error) {
	chartViews, err := s.repo.ListChartViewsByDatasetGroupID(field.DatasetGroupID)
	if err != nil {
		return nil, err
	}
	if fieldReferencedInChartViews(field.ID, chartViews) {
		deps = append(deps, "chart views")
	}

	rowPerms, err := s.repo.ListRowPermissionsByDatasetGroupID(field.DatasetGroupID)
	if err != nil {
		return nil, err
	}
	if fieldReferencedInRowPermissions(field.ID, rowPerms) {
		deps = append(deps, "row permissions")
	}

	columnPerms, err := s.repo.ListColumnPermissionsByDatasetGroupID(field.DatasetGroupID)
	if err != nil {
		return nil, err
	}
	if fieldReferencedInColumnPermissions(field, columnPerms) {
		deps = append(deps, "column permissions")
	}

	return deps, nil
}

func (s *DatasetService) appendVisualizationDependencies(deps []string, field *dataset.CoreDatasetTableField) ([]string, error) {
	deps, err := s.appendCountedDependency(deps, field.ID, "visualization linkage", s.repo.CountVisualizationLinkageFieldReferences)
	if err != nil {
		return nil, err
	}
	deps, err = s.appendCountedDependency(deps, field.ID, "visualization jumps", s.repo.CountVisualizationLinkJumpReferences)
	if err != nil {
		return nil, err
	}
	deps, err = s.appendCountedDependency(deps, field.ID, "outer parameter bindings", s.repo.CountVisualizationOuterParamReferences)
	if err != nil {
		return nil, err
	}
	return deps, nil
}

func (s *DatasetService) appendCountedDependency(deps []string, fieldID int64, label string, counter func(int64) (int64, error)) ([]string, error) {
	count, err := counter(fieldID)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		deps = append(deps, label)
	}
	return deps, nil
}

func fieldReferencedInChartViews(fieldID int64, views []auto.CoreChartView) bool {
	patterns := []string{
		fmt.Sprintf(`"id":%d,`, fieldID),
		fmt.Sprintf(`"id":%d}`, fieldID),
		fmt.Sprintf(`"id": %d,`, fieldID),
		fmt.Sprintf(`"id": %d}`, fieldID),
	}
	for _, view := range views {
		payloads := []string{
			view.XAxis,
			view.XAxisExt,
			view.YAxis,
			view.YAxisExt,
			view.ExtStack,
			view.ExtBubble,
			view.ExtLabel,
			view.ExtTooltip,
			view.CustomFilter,
			view.DrillFields,
			view.ViewFields,
			view.FlowMapStartName,
			view.FlowMapEndName,
			view.ExtColor,
			view.Senior,
		}
		for _, payload := range payloads {
			for _, pattern := range patterns {
				if strings.Contains(payload, pattern) {
					return true
				}
			}
		}
	}
	return false
}

func fieldReferencedInRowPermissions(fieldID int64, rows []permission.DataPermRow) bool {
	patterns := []string{
		fmt.Sprintf(`"fieldId":%d`, fieldID),
		fmt.Sprintf(`"fieldId": %d`, fieldID),
	}
	for _, row := range rows {
		for _, pattern := range patterns {
			if strings.Contains(row.ExpressionTree, pattern) {
				return true
			}
		}
	}
	return false
}

func fieldReferencedInColumnPermissions(field *dataset.CoreDatasetTableField, rows []permission.DataPermColumn) bool {
	if field == nil {
		return false
	}
	names := make(map[string]struct{})
	for _, candidate := range []*string{field.Name, field.OriginName, field.DataeaseName, field.FieldShortName} {
		if candidate == nil {
			continue
		}
		name := strings.TrimSpace(*candidate)
		if name != "" {
			names[name] = struct{}{}
		}
	}
	for _, row := range rows {
		if _, ok := names[strings.TrimSpace(row.FieldName)]; ok {
			return true
		}
	}
	return false
}

func (s *DatasetService) ExportDataset(req *dataset.ExportDatasetRequest, userID int64) (*dataset.ExportDatasetResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("export request is required")
	}
	if req.ID <= 0 {
		return nil, fmt.Errorf(errno.ErrDatasetIDRequired)
	}
	if s.repo == nil {
		return nil, fmt.Errorf("dataset repository not initialized")
	}
	if s.exportRepo == nil {
		return nil, fmt.Errorf("export repository not initialized")
	}

	group, err := s.GetGroupByID(req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf(errno.ErrDatasetNotFound)
		}
		return nil, err
	}
	resolvedDatasetID := group.ID

	viewName := strings.TrimSpace(req.ViewName)
	if viewName == "" {
		viewName = strings.TrimSpace(group.Name)
	}
	if viewName == "" {
		viewName = fmt.Sprintf("dataset_%d", req.ID)
	}

	taskID := strings.ReplaceAll(uuid.NewString(), "-", "")
	task := &exportdomain.ExportTask{
		ID:             taskID,
		UserID:         userID,
		FileName:       GenerateExcelFilename(viewName),
		FileSize:       0,
		FileSizeUnit:   "B",
		ExportFrom:     resolvedDatasetID,
		ExportStatus:   exportStatusPending,
		ExportFromType: permission.ResourceTypeDataset,
		ExportTime:     time.Now().UnixMilli(),
		ExportProgress: "0",
		ExportFromName: viewName,
	}
	if err = s.exportRepo.Create(task); err != nil {
		return nil, err
	}

	return &dataset.ExportDatasetResponse{
		TaskID:         taskID,
		Status:         exportStatusPending,
		ExportFrom:     resolvedDatasetID,
		ExportFromType: permission.ResourceTypeDataset,
		ExportFromName: viewName,
	}, nil
}

func (s *DatasetService) resolveEnumFieldTarget(fieldID int64) (*dataset.CoreDatasetTableField, string, string, error) {
	field, err := s.repo.GetFieldByID(fieldID)
	if err != nil {
		return nil, "", "", err
	}

	columnName := ""
	if field.OriginName != nil {
		columnName = strings.TrimSpace(*field.OriginName)
	}
	if columnName == "" && field.DataeaseName != nil {
		columnName = strings.TrimSpace(*field.DataeaseName)
	}
	if columnName == "" && field.Name != nil {
		columnName = strings.TrimSpace(*field.Name)
	}
	if columnName == "" {
		return nil, "", "", fmt.Errorf("dataset field origin name is required")
	}

	tableName := ""
	if field.DatasetTableID != nil && *field.DatasetTableID > 0 {
		table, tableErr := s.repo.GetTableByID(*field.DatasetTableID)
		if tableErr == nil && table.PhysicalTable != nil {
			tableName = strings.TrimSpace(*table.PhysicalTable)
		}
	}
	if tableName == "" {
		tableName, err = s.repo.FindPrimaryTableName(field.DatasetGroupID)
		if err != nil {
			return nil, "", "", err
		}
	}

	return field, tableName, columnName, nil
}

func (s *DatasetService) resolveEnumDisplayField(req *dataset.EnumValueRequest, queryField *dataset.CoreDatasetTableField, queryColumn, tableName string) (int64, *dataset.CoreDatasetTableField, string, error) {
	displayID := req.DisplayID
	if displayID <= 0 {
		return req.QueryID, queryField, queryColumn, nil
	}

	displayField, displayTableName, displayColumn, err := s.resolveEnumFieldTarget(displayID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return req.QueryID, queryField, queryColumn, nil
		}
		return 0, nil, "", err
	}
	if displayTableName != tableName {
		return req.QueryID, queryField, queryColumn, nil
	}

	return displayID, displayField, displayColumn, nil
}

func (s *DatasetService) resolveEnumSortColumn(req *dataset.EnumValueRequest, tableName, fallbackColumn string) (string, error) {
	if req.SortID <= 0 {
		return fallbackColumn, nil
	}

	_, sortTableName, sortColumn, err := s.resolveEnumFieldTarget(req.SortID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fallbackColumn, nil
		}
		return "", err
	}
	if sortTableName != tableName {
		return fallbackColumn, nil
	}

	return sortColumn, nil
}

func (s *DatasetService) buildEnumFilterClauses(filters []dataset.EnumFilter, targetTableName string) ([]dataset.EnumFilterClause, error) {
	clauses := make([]dataset.EnumFilterClause, 0)
	for _, filter := range filters {
		if strings.TrimSpace(filter.Operator) != "" && !strings.EqualFold(strings.TrimSpace(filter.Operator), "in") {
			continue
		}
		ids := parseFilterFieldIDs(filter.FieldID)
		if len(ids) == 0 {
			continue
		}
		values := extractFilterValues(filter.Value)
		if len(values) == 0 {
			continue
		}

		for _, id := range ids {
			_, tableName, columnName, err := s.resolveEnumFieldTarget(id)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return nil, err
			}
			if tableName != targetTableName {
				continue
			}
			clauses = append(clauses, dataset.EnumFilterClause{Column: columnName, Values: values})
		}
	}
	return clauses, nil
}

func parseFilterFieldIDs(fieldID string) []int64 {
	text := strings.TrimSpace(fieldID)
	if text == "" {
		return []int64{}
	}
	parts := strings.Split(text, ",")
	result := make([]int64, 0, len(parts))
	seen := make(map[int64]struct{}, len(parts))
	for _, part := range parts {
		id, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err != nil || id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func extractFilterValues(values []interface{}) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		text := strings.TrimSpace(fmt.Sprintf("%v", value))
		if text == "" || text == "<nil>" {
			continue
		}
		if _, ok := seen[text]; ok {
			continue
		}
		seen[text] = struct{}{}
		result = append(result, text)
	}
	return result
}

func normalizeEnumValue(value string, deType *int) string {
	text := strings.TrimSpace(value)
	if text == "" || strings.EqualFold(text, "<nil>") {
		return ""
	}
	if deType != nil && *deType == 3 && strings.Contains(strings.ToUpper(text), "E") {
		if f, _, err := big.ParseFloat(text, 10, 128, big.ToNearestEven); err == nil {
			return strings.TrimRight(strings.TrimRight(f.Text('f', 8), "0"), ".")
		}
	}
	return text
}

func enumAlias(fieldID int64) string {
	return fmt.Sprintf("f_%d", fieldID)
}

func enumFieldIDFromAlias(alias string) int64 {
	trimmed := strings.TrimSpace(alias)
	if !strings.HasPrefix(trimmed, "f_") {
		return 0
	}
	id, err := strconv.ParseInt(strings.TrimPrefix(trimmed, "f_"), 10, 64)
	if err != nil {
		return 0
	}
	return id
}

func buildEnumResultRows(rows []map[string]interface{}, columns []dataset.EnumObjectColumn, displayID int64, queryField, displayField *dataset.CoreDatasetTableField) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))

	for _, row := range rows {
		item, ok := buildEnumResultRow(row, columns, displayID, queryField, displayField)
		if !ok {
			continue
		}
		keyBytes, err := json.Marshal(item)
		if err != nil {
			continue
		}
		key := string(keyBytes)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, item)
	}

	return result
}

func buildEnumResultRow(row map[string]interface{}, columns []dataset.EnumObjectColumn, displayID int64, queryField, displayField *dataset.CoreDatasetTableField) (map[string]interface{}, bool) {
	item := make(map[string]interface{}, len(columns))
	for _, column := range columns {
		fieldID, normalized, ok := buildEnumResultValue(row, column, displayID, queryField, displayField)
		if !ok {
			return nil, false
		}
		item[strconv.FormatInt(fieldID, 10)] = normalized
	}
	return item, true
}

func buildEnumResultValue(row map[string]interface{}, column dataset.EnumObjectColumn, displayID int64, queryField, displayField *dataset.CoreDatasetTableField) (int64, string, bool) {
	rawValue, exists := row[column.Alias]
	if !exists {
		return 0, "", false
	}

	fieldID := enumFieldIDFromAlias(column.Alias)
	deType := queryField.DeType
	if fieldID == displayID {
		deType = displayField.DeType
	}

	normalized := normalizeEnumValue(fmt.Sprintf("%v", normalizePreviewValue(rawValue)), deType)
	if normalized == "" {
		return 0, "", false
	}

	return fieldID, normalized, true
}

func buildFieldTreeNodes(rows []map[string]interface{}, specs []fieldTreeSpec) []dataset.BaseTreeNode {
	if len(rows) == 0 || len(specs) == 0 {
		return []dataset.BaseTreeNode{}
	}

	type mutableTreeNode struct {
		node     dataset.BaseTreeNode
		children []*mutableTreeNode
	}

	rootNodes := make([]*mutableTreeNode, 0)
	index := make(map[string]*mutableTreeNode)

	for _, row := range rows {
		parentPath := ""
		var parent *mutableTreeNode
		for _, spec := range specs {
			rawValue, exists := row[spec.alias]
			if !exists {
				break
			}
			text := normalizeEnumValue(fmt.Sprintf("%v", normalizePreviewValue(rawValue)), spec.field.DeType)
			if text == "" {
				break
			}

			currentPath := spec.alias + ":" + text
			if parentPath != "" {
				currentPath = parentPath + "->" + currentPath
			}

			node, ok := index[currentPath]
			if !ok {
				node = &mutableTreeNode{node: dataset.BaseTreeNode{ID: currentPath, Pid: parentPath, Text: text, NodeType: ""}}
				index[currentPath] = node
				if parent == nil {
					rootNodes = append(rootNodes, node)
				} else {
					parent.children = append(parent.children, node)
				}
			}

			parent = node
			parentPath = currentPath
		}
	}

	var freeze func(nodes []*mutableTreeNode) []dataset.BaseTreeNode
	freeze = func(nodes []*mutableTreeNode) []dataset.BaseTreeNode {
		result := make([]dataset.BaseTreeNode, 0, len(nodes))
		for _, node := range nodes {
			item := node.node
			if len(node.children) > 0 {
				item.Children = freeze(node.children)
			}
			result = append(result, item)
		}
		return result
	}

	return freeze(rootNodes)
}

func (s *DatasetService) Save(req *dataset.WriteRequest) (*dataset.CoreDatasetGroup, error) {
	if req == nil {
		return nil, fmt.Errorf("dataset request is required")
	}
	if req.ID <= 0 {
		return s.Create(req)
	}

	existing, err := s.repo.GetGroupByID(req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf(errno.ErrDatasetNotFound)
		}
		return nil, err
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = existing.Name
	}
	pid := normalizedDatasetPID(req.PID)
	if req.PID == nil && existing.PID != nil {
		pid = *existing.PID
	}

	count, err := s.repo.CountGroupByNameAndPID(name, pid, &req.ID)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf(errno.ErrDatasetNameExists)
	}

	existing.Name = name
	existing.PID = &pid
	nodeType := normalizedDatasetNodeType(req.NodeType)
	if nodeType == "" {
		if existing.NodeType != nil {
			nodeType = strings.TrimSpace(*existing.NodeType)
		}
	}
	if nodeType != "" {
		existing.NodeType = &nodeType
	}
	if req.Type != nil {
		existing.Type = req.Type
	}

	if err = s.repo.UpdateGroup(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *DatasetService) Create(req *dataset.WriteRequest) (*dataset.CoreDatasetGroup, error) {
	if req == nil {
		return nil, fmt.Errorf("dataset request is required")
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf(errno.ErrDatasetNameRequired)
	}

	pid := normalizedDatasetPID(req.PID)
	level := 0
	if pid > 0 {
		parent, err := s.repo.GetGroupByID(pid)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("destination folder not found")
			}
			return nil, err
		}
		if parent.Level != nil {
			level = *parent.Level + 1
		} else {
			level = 1
		}
	}

	count, err := s.repo.CountGroupByNameAndPID(name, pid, nil)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf(errno.ErrDatasetNameExists)
	}

	nodeType := normalizedDatasetNodeType(req.NodeType)
	if nodeType == "" {
		nodeType = dataset.NodeTypeFolder
	}
	delFlag := 0
	group := &dataset.CoreDatasetGroup{
		Name:     name,
		PID:      &pid,
		Level:    &level,
		NodeType: &nodeType,
		Type:     req.Type,
		DelFlag:  &delFlag,
	}

	if err = s.repo.CreateGroup(group); err != nil {
		return nil, err
	}
	if err = s.applyInheritedPermissionsOnCreate(group.ID, group.Name, pid); err != nil {
		_ = s.repo.SoftDeleteGroup(group.ID)
		return nil, err
	}

	return group, nil
}

func (s *DatasetService) applyInheritedPermissionsOnCreate(resourceID int64, resourceName string, pid int64) error {
	if s.resourcePermService == nil || pid <= 0 {
		return nil
	}
	return s.resourcePermService.InheritParentResourcePermissions(pid, resourceID, resourceName, permission.ResourceTypeDataset)
}

func (s *DatasetService) BackfillGovernedResources() (*DatasetGovernanceBackfillReport, error) {
	return s.BackfillGovernedResourcesWithOptions(nil)
}

func (s *DatasetService) BackfillGovernedResourcesWithOptions(options *GovernanceBackfillOptions) (*DatasetGovernanceBackfillReport, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("dataset repository not initialized")
	}
	if s.resourcePermService == nil {
		return nil, fmt.Errorf(errno.ErrResourcePermNotInitialized)
	}
	return runGovernanceBackfillWithOptions(options, permission.ResourceTypeDataset, func(normalized GovernanceBackfillOptions) ([]*dataset.CoreDatasetGroup, error) {
		return s.repo.ListGroupsBatch(nil, normalized.AfterID, normalized.Limit)
	}, func(item *dataset.CoreDatasetGroup) governanceBackfillItem {
		return governanceBackfillItem{resourceID: item.ID, parentID: item.PID, resourceName: item.Name}
	}, func(parentID, resourceID int64, resourceName string) (bool, error) {
		return s.resourcePermService.TryInheritParentResourcePermissions(parentID, resourceID, resourceName, permission.ResourceTypeDataset)
	})
}

func (s *DatasetService) Rename(id int64, name string) (*dataset.CoreDatasetGroup, error) {
	if id <= 0 {
		return nil, fmt.Errorf(errno.ErrDatasetIDRequired)
	}

	existing, err := s.repo.GetGroupByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf(errno.ErrDatasetNotFound)
		}
		return nil, err
	}

	newName := strings.TrimSpace(name)
	if newName == "" {
		return nil, fmt.Errorf(errno.ErrDatasetNameRequired)
	}

	pid := int64(0)
	if existing.PID != nil {
		pid = *existing.PID
	}
	count, err := s.repo.CountGroupByNameAndPID(newName, pid, &id)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf(errno.ErrDatasetNameExists)
	}

	existing.Name = newName
	if err = s.repo.UpdateGroup(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *DatasetService) Move(id int64, pid int64) (*dataset.CoreDatasetGroup, error) {
	if id <= 0 {
		return nil, fmt.Errorf(errno.ErrDatasetIDRequired)
	}
	if id == pid {
		return nil, fmt.Errorf("destination folder cannot be itself")
	}

	existing, err := s.repo.GetGroupByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf(errno.ErrDatasetNotFound)
		}
		return nil, err
	}

	if pid > 0 {
		if _, err = s.repo.GetGroupByID(pid); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("destination folder not found")
			}
			return nil, err
		}
		isDescendant, checkErr := s.isDescendant(id, pid)
		if checkErr != nil {
			return nil, checkErr
		}
		if isDescendant {
			return nil, fmt.Errorf("destination folder cannot be child of current dataset")
		}
	}

	count, err := s.repo.CountGroupByNameAndPID(existing.Name, pid, &id)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf(errno.ErrDatasetNameExists)
	}

	existing.PID = &pid
	if err = s.repo.UpdateGroup(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *DatasetService) Delete(id int64) error {
	if id <= 0 {
		return fmt.Errorf(errno.ErrDatasetIDRequired)
	}
	return s.deleteRecursive(id)
}

func (s *DatasetService) deleteRecursive(id int64) error {
	children, err := s.repo.ListGroupChildren(id)
	if err != nil {
		return err
	}
	for _, child := range children {
		if err = s.deleteRecursive(child.ID); err != nil {
			return err
		}
	}
	return s.repo.SoftDeleteGroup(id)
}

func (s *DatasetService) isDescendant(rootID int64, targetID int64) (bool, error) {
	children, err := s.repo.ListGroupChildren(rootID)
	if err != nil {
		return false, err
	}
	for _, child := range children {
		if child.ID == targetID {
			return true, nil
		}
		descendant, innerErr := s.isDescendant(child.ID, targetID)
		if innerErr != nil {
			return false, innerErr
		}
		if descendant {
			return true, nil
		}
	}
	return false, nil
}

func normalizedDatasetPID(pid *int64) int64 {
	if pid == nil {
		return 0
	}
	if *pid < 0 {
		return 0
	}
	return *pid
}

func normalizedDatasetNodeType(nodeType string) string {
	nt := strings.TrimSpace(nodeType)
	if nt == "" {
		return ""
	}
	switch nt {
	case dataset.NodeTypeFolder, dataset.NodeTypeDataset:
		return nt
	default:
		return dataset.NodeTypeDataset
	}
}

func (s *DatasetService) datasetFullName(datasetGroupID int64) (string, error) {
	if datasetGroupID <= 0 {
		return "", nil
	}

	names := make([]string, 0)
	visited := make(map[int64]struct{})
	currentID := datasetGroupID
	for currentID > 0 {
		if _, ok := visited[currentID]; ok {
			break
		}
		visited[currentID] = struct{}{}

		group, err := s.repo.GetGroupByID(currentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break
			}
			return "", err
		}
		n := strings.TrimSpace(group.Name)
		if n != "" {
			names = append(names, n)
		}
		if group.PID == nil || *group.PID <= 0 || *group.PID == currentID {
			break
		}
		currentID = *group.PID
	}

	for i, j := 0, len(names)-1; i < j; i, j = i+1, j-1 {
		names[i], names[j] = names[j], names[i]
	}
	return strings.Join(names, "/"), nil
}

func validatePreviewSQL(rawSQL string) error {
	text := strings.TrimSpace(strings.TrimSuffix(rawSQL, ";"))
	lower := strings.ToLower(text)
	if text == "" {
		return fmt.Errorf(errno.ErrSQLRequired)
	}
	if !strings.HasPrefix(lower, "select") && !strings.HasPrefix(lower, "with") {
		return fmt.Errorf("only select query is supported")
	}
	if strings.Contains(text, ";") {
		return fmt.Errorf("only single select statement is supported")
	}
	blocked := []string{" insert ", " update ", " delete ", " drop ", " alter ", " truncate ", " create ", " union "}
	padded := " " + lower + " "
	for _, token := range blocked {
		if strings.Contains(padded, token) {
			return fmt.Errorf("unsupported sql statement")
		}
	}
	return nil
}

func (s *DatasetService) validateWithCalciteIfEnabled(rawSQL string) error {
	if strings.TrimSpace(s.calciteAddress) == "" {
		return nil
	}

	client, err := s.ensureCalciteClient()
	if err != nil {
		return fmt.Errorf("calcite client unavailable: %w", err)
	}

	valid, err := client.ValidateSQL(context.TODO(), rawSQL)
	if err != nil {
		return fmt.Errorf("calcite validate sql failed: %w", err)
	}
	if !valid {
		return fmt.Errorf("sql validation failed")
	}

	return nil
}

func (s *DatasetService) ensureCalciteClient() (*calciteintegration.Client, error) {
	s.calciteMu.Lock()
	defer s.calciteMu.Unlock()

	if s.calciteClient != nil {
		return s.calciteClient, nil
	}

	timeout := s.calciteTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	retries := s.calciteRetries
	if retries < 0 {
		retries = 0
	}

	client, err := calciteintegration.NewClient(&calciteintegration.Config{Address: s.calciteAddress, Timeout: timeout, MaxRetries: retries})
	if err != nil {
		return nil, err
	}
	s.calciteClient = client
	return s.calciteClient, nil
}

func buildPreviewFields(rows []map[string]interface{}) []dataset.SQLPreviewField {
	if len(rows) == 0 {
		return []dataset.SQLPreviewField{}
	}

	keys := make([]string, 0, len(rows[0]))
	for k := range rows[0] {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fields := make([]dataset.SQLPreviewField, 0, len(keys))
	for _, k := range keys {
		fields = append(fields, dataset.SQLPreviewField{
			OriginName: k,
			DeType:     inferPreviewDeType(rows[0][k]),
		})
	}
	return fields
}

func normalizePreviewRow(row map[string]interface{}) map[string]interface{} {
	if row == nil {
		return map[string]interface{}{}
	}
	result := make(map[string]interface{}, len(row))
	for key, val := range row {
		result[key] = normalizePreviewValue(val)
	}
	return result
}

func normalizePreviewValue(v interface{}) interface{} {
	switch value := v.(type) {
	case []byte:
		return string(value)
	case time.Time:
		return value.Format("2006-01-02 15:04:05")
	default:
		return value
	}
}

func inferPreviewDeType(v interface{}) int {
	switch value := v.(type) {
	case bool:
		return 4
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return 2
	case float32, float64:
		return 3
	case time.Time:
		return 1
	case string:
		text := strings.TrimSpace(value)
		if text == "" {
			return 0
		}
		if isDateTimeText(text) {
			return 1
		}
		if _, err := strconv.ParseInt(text, 10, 64); err == nil {
			return 2
		}
		if _, err := strconv.ParseFloat(text, 64); err == nil {
			return 3
		}
		return 0
	default:
		return 0
	}
}

func inferSQLVariableDeType(typeList []string) int {
	if len(typeList) == 0 {
		return 0
	}
	typeText := strings.ToUpper(strings.TrimSpace(typeList[0]))
	if typeText == "" {
		return 0
	}
	if containsAnyText(typeText, sqlVariableDateTimePatterns) {
		return 1
	}
	if containsAnyText(typeText, sqlVariableFloatPatterns) {
		return 3
	}
	if containsAnyText(typeText, sqlVariableIntegerPatterns) {
		return 2
	}
	if strings.Contains(typeText, "BOOL") {
		return 4
	}
	return 0
}

func containsAnyText(text string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}

// SaveField creates or updates a dataset field. If field.ID == 0, creates; otherwise updates.
func (s *DatasetService) SaveField(field *dataset.CoreDatasetTableField) (*dataset.CoreDatasetTableField, error) {
	if field.Name == nil || *field.Name == "" {
		return nil, fmt.Errorf("field name is required")
	}
	if field.DatasetGroupID == 0 {
		return nil, fmt.Errorf("datasetGroupId is required")
	}
	if field.Type == nil || *field.Type == "" {
		return nil, fmt.Errorf("field type is required")
	}

	if field.ID == 0 {
		if err := s.repo.CreateDatasetField(field); err != nil {
			return nil, fmt.Errorf("failed to create field: %w", err)
		}
		return field, nil
	}

	if err := s.repo.UpdateDatasetField(field); err != nil {
		return nil, fmt.Errorf("failed to update field: %w", err)
	}
	return field, nil
}

// FunctionCategory represents a category of SQL functions for the calculated field editor.
type FunctionCategory struct {
	Name      string        `json:"name"`
	Functions []FunctionDef `json:"functions"`
}

// FunctionDef represents a single SQL function definition.
type FunctionDef struct {
	Name string `json:"name"`
	Hint string `json:"hint"`
}

// GetFieldFunctions returns static SQL function categories for the calculated field editor.
func (s *DatasetService) GetFieldFunctions() []FunctionCategory {
	return []FunctionCategory{
		{
			Name: "聚合函数",
			Functions: []FunctionDef{
				{Name: "SUM", Hint: "SUM(field)"},
				{Name: "AVG", Hint: "AVG(field)"},
				{Name: "MAX", Hint: "MAX(field)"},
				{Name: "MIN", Hint: "MIN(field)"},
				{Name: "COUNT", Hint: "COUNT(field)"},
			},
		},
		{
			Name: "日期函数",
			Functions: []FunctionDef{
				{Name: "DATE_FORMAT", Hint: "DATE_FORMAT(date, format)"},
				{Name: "YEAR", Hint: "YEAR(date)"},
				{Name: "MONTH", Hint: "MONTH(date)"},
				{Name: "DAY", Hint: "DAY(date)"},
				{Name: "HOUR", Hint: "HOUR(date)"},
				{Name: "NOW", Hint: "NOW()"},
			},
		},
		{
			Name: "字符串函数",
			Functions: []FunctionDef{
				{Name: "CONCAT", Hint: "CONCAT(str1, str2, ...)"},
				{Name: "SUBSTRING", Hint: "SUBSTRING(str, pos, len)"},
				{Name: "LENGTH", Hint: "LENGTH(str)"},
				{Name: "LOWER", Hint: "LOWER(str)"},
				{Name: "UPPER", Hint: "UPPER(str)"},
				{Name: "TRIM", Hint: "TRIM(str)"},
				{Name: "REPLACE", Hint: "REPLACE(str, from, to)"},
			},
		},
		{
			Name: "数学函数",
			Functions: []FunctionDef{
				{Name: "ABS", Hint: "ABS(x)"},
				{Name: "ROUND", Hint: "ROUND(x, d)"},
				{Name: "CEIL", Hint: "CEIL(x)"},
				{Name: "FLOOR", Hint: "FLOOR(x)"},
				{Name: "MOD", Hint: "MOD(x, y)"},
			},
		},
		{
			Name: "条件函数",
			Functions: []FunctionDef{
				{Name: "IF", Hint: "IF(cond, true_val, false_val)"},
				{Name: "CASE", Hint: "CASE WHEN ... THEN ... END"},
				{Name: "COALESCE", Hint: "COALESCE(val1, val2, ...)"},
				{Name: "IFNULL", Hint: "IFNULL(expr, alt)"},
			},
		},
	}
}

// ListFieldsByDsIds returns dataset fields filtered by datasource IDs.
func (s *DatasetService) ListFieldsByDsIds(dsIds []int64) ([]dataset.CoreDatasetTableField, error) {
	return s.repo.ListFieldsByDsIds(dsIds)
}

func isDateTimeText(text string) bool {
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
		"2006/01/02 15:04:05",
		"2006/01/02",
	}
	for _, layout := range layouts {
		if _, err := time.Parse(layout, text); err == nil {
			return true
		}
	}
	return false
}
