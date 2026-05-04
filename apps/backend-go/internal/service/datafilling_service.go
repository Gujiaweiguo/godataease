package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	datafillingdomain "dataease/backend/internal/domain/datafilling"
	datasourcedomain "dataease/backend/internal/domain/datasource"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DataFillingRepo interface {
	Create(ctx context.Context, form *datafillingdomain.DataFillingForm) error
	GetByID(ctx context.Context, id int64) (*datafillingdomain.DataFillingForm, error)
	Update(ctx context.Context, form *datafillingdomain.DataFillingForm) error
	DeleteByID(ctx context.Context, id int64) error
	Rename(ctx context.Context, id int64, name string) error
	Move(ctx context.Context, id int64, pid int64) error
	GetTree(ctx context.Context) ([]*datafillingdomain.DataFillingForm, error)
	GetByPID(ctx context.Context, pid int64) ([]*datafillingdomain.DataFillingForm, error)
	GetChildren(ctx context.Context, pid int64) ([]*datafillingdomain.DataFillingForm, error)
}

type DataFillingDatasourceService interface {
	GetByID(id int64) (*datasourcedomain.CoreDatasource, error)
	Tree(req *datasourcedomain.ListRequest) ([]*datasourcedomain.CoreDatasource, error)
	GetTables(req *datasourcedomain.TableRequest) ([]datasourcedomain.TableInfo, error)
}

type CommitLogRepo interface {
	Create(ctx context.Context, log *datafillingdomain.DfCommitLog) error
	ListByFormID(ctx context.Context, formID int64, page, pageSize int) ([]*datafillingdomain.DfCommitLog, int64, error)
	DeleteByFormID(ctx context.Context, formID int64) error
}

type DataFillingService struct {
	repo                     DataFillingRepo
	datasourceService        DataFillingDatasourceService
	ddlProvider              DDLProvider
	commitLogRepo            CommitLogRepo
	datasourceConnProvider   DatasourceConnectionProvider
}

func NewDataFillingService(repo DataFillingRepo, datasourceService DataFillingDatasourceService, ddlProvider DDLProvider, commitLogRepo CommitLogRepo) *DataFillingService {
	return &DataFillingService{repo: repo, datasourceService: datasourceService, ddlProvider: ddlProvider, commitLogRepo: commitLogRepo}
}

func (s *DataFillingService) SetDatasourceConnectionProvider(provider DatasourceConnectionProvider) {
	s.datasourceConnProvider = provider
}

func (s *DataFillingService) Save(ctx context.Context, req *datafillingdomain.CreateFormRequest, userID int64) (*datafillingdomain.DataFillingForm, error) {
	if err := validateDataFillingCreateRequest(req); err != nil {
		return nil, err
	}
	nodeType := normalizeDataFillingNodeType(req.NodeType)
	level, err := s.resolveLevel(ctx, req.PID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UnixMilli()
	form := &datafillingdomain.DataFillingForm{
		Name:              strings.TrimSpace(req.Name),
		PID:               req.PID,
		Level:             level,
		NodeType:          nodeType,
		PhysicalTableName: strings.TrimSpace(req.TableName),
		DatasourceID:      req.DatasourceID,
		Forms:             strings.TrimSpace(req.Forms),
		CreateIndex:       req.CreateIndex,
		TableIndexes:      strings.TrimSpace(req.TableIndexes),
		CreateBy:          userID,
		CreateTime:        now,
		UpdateBy:          userID,
		UpdateTime:        now,
		UseExistsTable:    req.UseExistsTable,
	}
	if err := s.repo.Create(ctx, form); err != nil {
		return nil, err
	}
	if nodeType == datafillingdomain.NodeTypeForm && !req.UseExistsTable {
		if err := s.createPhysicalTable(ctx, form); err != nil {
			_ = s.repo.DeleteByID(ctx, form.ID)
			return nil, err
		}
	}
	return form, nil
}

func (s *DataFillingService) Get(ctx context.Context, id int64) (*datafillingdomain.DataFillingForm, error) {
	if id <= 0 {
		return nil, gorm.ErrInvalidData
	}
	return s.repo.GetByID(ctx, id)
}

func (s *DataFillingService) Update(ctx context.Context, req *datafillingdomain.UpdateFormRequest, userID int64) (*datafillingdomain.DataFillingForm, error) {
	if req == nil || req.ID <= 0 {
		return nil, gorm.ErrInvalidData
	}
	if err := validateDataFillingCreateRequest(&req.CreateFormRequest); err != nil {
		return nil, err
	}
	current, err := s.repo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	originalFields := strings.TrimSpace(current.Forms)
	level, err := s.resolveLevel(ctx, req.PID)
	if err != nil {
		return nil, err
	}
	current.Name = strings.TrimSpace(req.Name)
	current.PID = req.PID
	current.Level = level
	current.NodeType = normalizeDataFillingNodeType(req.NodeType)
	current.PhysicalTableName = strings.TrimSpace(req.TableName)
	current.DatasourceID = req.DatasourceID
	current.Forms = strings.TrimSpace(req.Forms)
	current.CreateIndex = req.CreateIndex
	current.TableIndexes = strings.TrimSpace(req.TableIndexes)
	current.UpdateBy = userID
	current.UpdateTime = time.Now().UnixMilli()
	if normalizeDataFillingNodeType(current.NodeType) == datafillingdomain.NodeTypeForm && originalFields != current.Forms {
		if err := s.alterPhysicalTableForFieldChanges(ctx, current, originalFields, current.Forms); err != nil {
			return nil, err
		}
	}
	if err := s.repo.Update(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *DataFillingService) SearchTableData(ctx context.Context, formID int64, req *datafillingdomain.TableDataRequest) (*datafillingdomain.TableDataResponse, error) {
	form, db, err := s.loadFormAndDatasource(ctx, formID)
	if err != nil {
		return nil, err
	}
	currentPage := int64(1)
	pageSize := int64(20)
	searchParams := []datafillingdomain.SearchParam{}
	if req != nil {
		if req.CurrentPage > 0 {
			currentPage = req.CurrentPage
		}
		if req.PageSize > 0 {
			pageSize = req.PageSize
		}
		searchParams = req.SearchParams
	}
	whereClause, args, err := buildWhereClause(searchParams)
	if err != nil {
		return nil, err
	}
	offset := (currentPage - 1) * pageSize
	rows, err := s.ddlProvider.SearchRows(ctx, db, form.PhysicalTableName, whereClause, args, pageSize, offset)
	if err != nil {
		return nil, err
	}
	total, err := s.ddlProvider.CountRows(ctx, db, form.PhysicalTableName, whereClause, args)
	if err != nil {
		return nil, err
	}
	return &datafillingdomain.TableDataResponse{Data: rows, Fields: form.Forms, Total: total, CurrentPage: currentPage, PageSize: pageSize, Key: "id"}, nil
}

func (s *DataFillingService) SaveRowData(ctx context.Context, formID int64, rowData map[string]interface{}, userID int64, userName string) (*datafillingdomain.TableDataResponse, error) {
	if len(rowData) == 0 {
		return nil, gorm.ErrInvalidData
	}
	form, db, err := s.loadFormAndDatasource(ctx, formID)
	if err != nil {
		return nil, err
	}
	rowID := strings.TrimSpace(fmt.Sprint(rowData["id"]))
	operate := 1
	if rowID == "" || rowID == "<nil>" {
		delete(rowData, "id")
		if err := s.ddlProvider.InsertRow(ctx, db, form.PhysicalTableName, rowData); err != nil {
			return nil, err
		}
		rowID = strings.TrimSpace(fmt.Sprint(rowData["id"]))
	} else {
		operate = 2
		if err := s.ddlProvider.UpdateRow(ctx, db, form.PhysicalTableName, rowID, rowData); err != nil {
			return nil, err
		}
	}
	if err := s.writeCommitLog(ctx, formID, rowID, operate, userID, userName, 1); err != nil {
		return nil, err
	}
	return &datafillingdomain.TableDataResponse{Data: []map[string]interface{}{rowData}, Fields: form.Forms, Total: 1, CurrentPage: 1, PageSize: 1, Key: "id"}, nil
}

func (s *DataFillingService) DeleteRowData(ctx context.Context, formID int64, rowID string, userID int64, userName string) error {
	if strings.TrimSpace(rowID) == "" {
		return gorm.ErrInvalidData
	}
	form, db, err := s.loadFormAndDatasource(ctx, formID)
	if err != nil {
		return err
	}
	if err := s.ddlProvider.DeleteRows(ctx, db, form.PhysicalTableName, []string{rowID}); err != nil {
		return err
	}
	return s.writeCommitLog(ctx, formID, rowID, 0, userID, userName, 1)
}

func (s *DataFillingService) BatchDeleteRowData(ctx context.Context, formID int64, ids []string, userID int64, userName string) error {
	cleaned := cleanRowIDs(ids)
	if len(cleaned) == 0 {
		return gorm.ErrInvalidData
	}
	form, db, err := s.loadFormAndDatasource(ctx, formID)
	if err != nil {
		return err
	}
	for start := 0; start < len(cleaned); start += 500 {
		end := start + 500
		if end > len(cleaned) {
			end = len(cleaned)
		}
		if err := s.ddlProvider.DeleteRows(ctx, db, form.PhysicalTableName, cleaned[start:end]); err != nil {
			return err
		}
	}
	return s.writeCommitLog(ctx, formID, "", 0, userID, userName, len(cleaned))
}

func (s *DataFillingService) TruncateTableData(ctx context.Context, formID int64) error {
	form, db, err := s.loadFormAndDatasource(ctx, formID)
	if err != nil {
		return err
	}
	return s.ddlProvider.TruncateTable(ctx, db, form.PhysicalTableName)
}

func (s *DataFillingService) ListColumnData(ctx context.Context, formID int64, columnName string) ([]string, error) {
	if !isValidDDLIdentifier(columnName) {
		return nil, gorm.ErrInvalidData
	}
	form, db, err := s.loadFormAndDatasource(ctx, formID)
	if err != nil {
		return nil, err
	}
	return s.ddlProvider.ListColumnData(ctx, db, form.PhysicalTableName, columnName)
}

func (s *DataFillingService) ListCommitLogs(ctx context.Context, formID int64, page, pageSize int) ([]*datafillingdomain.DfCommitLog, int64, error) {
	if s.commitLogRepo == nil {
		return nil, 0, fmt.Errorf("commit log repository not configured")
	}
	if formID <= 0 || page <= 0 || pageSize <= 0 {
		return nil, 0, gorm.ErrInvalidData
	}
	return s.commitLogRepo.ListByFormID(ctx, formID, page, pageSize)
}

func (s *DataFillingService) ClearCommitLogs(ctx context.Context, formID int64) error {
	if s.commitLogRepo == nil {
		return fmt.Errorf("commit log repository not configured")
	}
	if formID <= 0 {
		return gorm.ErrInvalidData
	}
	return s.commitLogRepo.DeleteByFormID(ctx, formID)
}

func (s *DataFillingService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return gorm.ErrInvalidData
	}
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	children, err := s.repo.GetChildren(ctx, item.ID)
	if err != nil {
		return err
	}
	for _, child := range children {
		if err := s.Delete(ctx, child.ID); err != nil {
			return err
		}
	}
	if normalizeDataFillingNodeType(item.NodeType) == datafillingdomain.NodeTypeForm && item.DatasourceID > 0 && strings.TrimSpace(item.PhysicalTableName) != "" && s.ddlProvider != nil {
		db, err := s.GetDatasourceConnection(ctx, item.DatasourceID)
		if err != nil {
			return err
		}
		if err := s.ddlProvider.DropTable(ctx, db, item.PhysicalTableName); err != nil {
			return err
		}
	}
	return s.repo.DeleteByID(ctx, id)
}

func (s *DataFillingService) Rename(ctx context.Context, id int64, name string) (*datafillingdomain.DataFillingForm, error) {
	if id <= 0 || strings.TrimSpace(name) == "" {
		return nil, gorm.ErrInvalidData
	}
	if err := s.repo.Rename(ctx, id, strings.TrimSpace(name)); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *DataFillingService) Move(ctx context.Context, id int64, pid int64) (*datafillingdomain.DataFillingForm, error) {
	if id <= 0 {
		return nil, gorm.ErrInvalidData
	}
	if id == pid {
		return nil, fmt.Errorf("cannot move node to itself")
	}
	if err := s.repo.Move(ctx, id, pid); err != nil {
		return nil, err
	}
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.recalculateChildrenLevels(ctx, item.ID, item.Level); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *DataFillingService) Tree(ctx context.Context, req *datafillingdomain.TreeRequest) (datafillingdomain.TreeResponse, error) {
	rows, err := s.repo.GetTree(ctx)
	if err != nil {
		return nil, err
	}
	root := buildDataFillingTree(rows)
	if req == nil || req.Keyword == nil || strings.TrimSpace(*req.Keyword) == "" {
		return root, nil
	}
	keyword := strings.ToLower(strings.TrimSpace(*req.Keyword))
	return filterDataFillingTree(root, keyword), nil
}

func (s *DataFillingService) ListDatasourceList(ctx context.Context) ([]datafillingdomain.DatasourceSummary, error) {
	_ = ctx
	return s.listDatasourceSummaries(true)
}

func (s *DataFillingService) ListDatasourceListAll(ctx context.Context) ([]datafillingdomain.DatasourceSummary, error) {
	_ = ctx
	return s.listDatasourceSummaries(false)
}

func (s *DataFillingService) GetBuiltInTables(ctx context.Context) ([]datafillingdomain.BuiltInTable, error) {
	_ = ctx
	return []datafillingdomain.BuiltInTable{}, nil
}

func (s *DataFillingService) GetDatasourceConnection(ctx context.Context, datasourceID int64) (*gorm.DB, error) {
	if s.datasourceConnProvider != nil {
		return s.datasourceConnProvider.GetDatasourceConnection(ctx, datasourceID)
	}
	_ = ctx
	if s.datasourceService == nil {
		return nil, fmt.Errorf("datasource service not configured")
	}
	ds, err := s.datasourceService.GetByID(datasourceID)
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(strings.TrimSpace(ds.Type), "mysql") {
		return nil, fmt.Errorf("unsupported datasource type: %s", ds.Type)
	}
	if ds.Configuration == nil || strings.TrimSpace(*ds.Configuration) == "" {
		return nil, fmt.Errorf("datasource configuration is empty")
	}
	cfg, err := decodeConfig(*ds.Configuration)
	if err != nil {
		return nil, err
	}
	mysqlCfg, err := buildMySQLPreviewConfig(cfg)
	if err != nil {
		return nil, err
	}
	return gorm.Open(mysql.Open(mysqlCfg.FormatDSN()), &gorm.Config{})
}

func (s *DataFillingService) createPhysicalTable(ctx context.Context, form *datafillingdomain.DataFillingForm) error {
	if s.ddlProvider == nil {
		return fmt.Errorf("ddl provider not configured")
	}
	fields := make([]datafillingdomain.ExtTableField, 0)
	if strings.TrimSpace(form.Forms) != "" {
		if err := json.Unmarshal([]byte(form.Forms), &fields); err != nil {
			return fmt.Errorf("parse form fields: %w", err)
		}
	}
	db, err := s.GetDatasourceConnection(ctx, form.DatasourceID)
	if err != nil {
		return err
	}
	return s.ddlProvider.CreateTable(ctx, db, form.PhysicalTableName, fields)
}

func (s *DataFillingService) loadFormAndDatasource(ctx context.Context, formID int64) (*datafillingdomain.DataFillingForm, *gorm.DB, error) {
	if formID <= 0 {
		return nil, nil, gorm.ErrInvalidData
	}
	if s.ddlProvider == nil {
		return nil, nil, fmt.Errorf("ddl provider not configured")
	}
	form, err := s.repo.GetByID(ctx, formID)
	if err != nil {
		return nil, nil, err
	}
	db, err := s.GetDatasourceConnection(ctx, form.DatasourceID)
	if err != nil {
		return nil, nil, err
	}
	return form, db, nil
}

func (s *DataFillingService) alterPhysicalTableForFieldChanges(ctx context.Context, form *datafillingdomain.DataFillingForm, oldForms, newForms string) error {
	if s.ddlProvider == nil || form == nil || strings.TrimSpace(form.PhysicalTableName) == "" || form.DatasourceID <= 0 {
		return nil
	}
	oldFields, err := parseExtTableFields(oldForms)
	if err != nil {
		return err
	}
	newFields, err := parseExtTableFields(newForms)
	if err != nil {
		return err
	}
	toAdd, toDrop := diffExtTableFields(oldFields, newFields)
	if len(toAdd) == 0 && len(toDrop) == 0 {
		return nil
	}
	db, err := s.GetDatasourceConnection(ctx, form.DatasourceID)
	if err != nil {
		return err
	}
	if len(toAdd) > 0 {
		if err := s.ddlProvider.AddTableColumns(ctx, db, form.PhysicalTableName, toAdd); err != nil {
			return err
		}
	}
	if len(toDrop) > 0 {
		if err := s.ddlProvider.DropTableColumns(ctx, db, form.PhysicalTableName, toDrop); err != nil {
			return err
		}
	}
	return nil
}

func (s *DataFillingService) writeCommitLog(ctx context.Context, formID int64, dataID string, operate int, userID int64, userName string, count int) error {
	if s.commitLogRepo == nil {
		return nil
	}
	return s.commitLogRepo.Create(ctx, &datafillingdomain.DfCommitLog{FormID: formID, DataID: dataID, Operate: operate, CommitBy: userID, Committer: userName, CommitTime: time.Now().UnixMilli(), Count: count})
}

func parseExtTableFields(raw string) ([]datafillingdomain.ExtTableField, error) {
	if strings.TrimSpace(raw) == "" {
		return []datafillingdomain.ExtTableField{}, nil
	}
	fields := make([]datafillingdomain.ExtTableField, 0)
	if err := json.Unmarshal([]byte(raw), &fields); err != nil {
		return nil, fmt.Errorf("parse form fields: %w", err)
	}
	return fields, nil
}

func diffExtTableFields(oldFields, newFields []datafillingdomain.ExtTableField) ([]datafillingdomain.ExtTableField, []string) {
	oldMap := make(map[string]datafillingdomain.ExtTableField)
	newMap := make(map[string]datafillingdomain.ExtTableField)
	for _, field := range oldFields {
		column := strings.TrimSpace(field.Settings.Mapping.ColumnName)
		if column != "" && !field.Removed {
			oldMap[column] = field
		}
	}
	for _, field := range newFields {
		column := strings.TrimSpace(field.Settings.Mapping.ColumnName)
		if column != "" && !field.Removed {
			newMap[column] = field
		}
	}
	toAdd := make([]datafillingdomain.ExtTableField, 0)
	toDrop := make([]string, 0)
	for column, field := range newMap {
		if _, ok := oldMap[column]; !ok {
			toAdd = append(toAdd, field)
		}
	}
	for column := range oldMap {
		if _, ok := newMap[column]; !ok {
			toDrop = append(toDrop, column)
		}
	}
	sort.Slice(toAdd, func(i, j int) bool { return toAdd[i].Settings.Mapping.ColumnName < toAdd[j].Settings.Mapping.ColumnName })
	sort.Strings(toDrop)
	return toAdd, toDrop
}

func cleanRowIDs(ids []string) []string {
	cleaned := make([]string, 0, len(ids))
	for _, id := range ids {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}
	return cleaned
}

func (s *DataFillingService) resolveLevel(ctx context.Context, pid int64) (int, error) {
	if pid <= 0 {
		return 0, nil
	}
	parent, err := s.repo.GetByID(ctx, pid)
	if err != nil {
		return 0, err
	}
	return parent.Level + 1, nil
}

func (s *DataFillingService) recalculateChildrenLevels(ctx context.Context, pid int64, parentLevel int) error {
	children, err := s.repo.GetChildren(ctx, pid)
	if err != nil {
		return err
	}
	for _, child := range children {
		child.Level = parentLevel + 1
		if err := s.repo.Update(ctx, child); err != nil {
			return err
		}
		if err := s.recalculateChildrenLevels(ctx, child.ID, child.Level); err != nil {
			return err
		}
	}
	return nil
}

func (s *DataFillingService) listDatasourceSummaries(enabledOnly bool) ([]datafillingdomain.DatasourceSummary, error) {
	if s.datasourceService == nil {
		return []datafillingdomain.DatasourceSummary{}, nil
	}
	list, err := s.datasourceService.Tree(&datasourcedomain.ListRequest{})
	if err != nil {
		return nil, err
	}
	result := make([]datafillingdomain.DatasourceSummary, 0, len(list))
	for _, item := range list {
		if item == nil {
			continue
		}
		enabled := item.EnableDataFill != nil && *item.EnableDataFill
		if enabledOnly && !enabled {
			continue
		}
		pid := int64(0)
		if item.PID != nil {
			pid = *item.PID
		}
		status := ""
		if item.Status != nil {
			status = *item.Status
		}
		result = append(result, datafillingdomain.DatasourceSummary{ID: item.ID, PID: pid, Name: item.Name, Type: item.Type, Status: status, EnableDataFill: enabled})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

func validateDataFillingCreateRequest(req *datafillingdomain.CreateFormRequest) error {
	if req == nil || strings.TrimSpace(req.Name) == "" {
		return gorm.ErrInvalidData
	}
	if normalizeDataFillingNodeType(req.NodeType) == datafillingdomain.NodeTypeFolder {
		return nil
	}
	if strings.TrimSpace(req.TableName) == "" || req.DatasourceID <= 0 {
		return gorm.ErrInvalidData
	}
	return nil
}

func normalizeDataFillingNodeType(nodeType string) string {
	if strings.EqualFold(strings.TrimSpace(nodeType), datafillingdomain.NodeTypeFolder) {
		return datafillingdomain.NodeTypeFolder
	}
	return datafillingdomain.NodeTypeForm
}

func buildDataFillingTree(rows []*datafillingdomain.DataFillingForm) []*datafillingdomain.TreeNode {
	nodeMap := make(map[int64]*datafillingdomain.TreeNode, len(rows))
	result := make([]*datafillingdomain.TreeNode, 0)
	for _, row := range rows {
		if row == nil {
			continue
		}
		nodeMap[row.ID] = &datafillingdomain.TreeNode{ID: row.ID, Name: row.Name, PID: row.PID, NodeType: normalizeDataFillingNodeType(row.NodeType)}
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		node := nodeMap[row.ID]
		if row.PID <= 0 {
			result = append(result, node)
			continue
		}
		if parent, ok := nodeMap[row.PID]; ok {
			parent.Children = append(parent.Children, node)
		} else {
			result = append(result, node)
		}
	}
	return result
}

func filterDataFillingTree(nodes []*datafillingdomain.TreeNode, keyword string) []*datafillingdomain.TreeNode {
	filtered := make([]*datafillingdomain.TreeNode, 0)
	for _, node := range nodes {
		children := filterDataFillingTree(node.Children, keyword)
		if strings.Contains(strings.ToLower(node.Name), keyword) || len(children) > 0 {
			copied := *node
			copied.Children = children
			filtered = append(filtered, &copied)
		}
	}
	return filtered
}
