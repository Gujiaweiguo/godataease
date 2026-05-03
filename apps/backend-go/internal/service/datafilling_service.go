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

type DataFillingService struct {
	repo              DataFillingRepo
	datasourceService DataFillingDatasourceService
	ddlProvider       DDLProvider
}

func NewDataFillingService(repo DataFillingRepo, datasourceService DataFillingDatasourceService, ddlProvider DDLProvider) *DataFillingService {
	return &DataFillingService{repo: repo, datasourceService: datasourceService, ddlProvider: ddlProvider}
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
	if nodeType == "form" && !req.UseExistsTable {
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
	if err := s.repo.Update(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
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
	if normalizeDataFillingNodeType(item.NodeType) == "form" && item.DatasourceID > 0 && strings.TrimSpace(item.PhysicalTableName) != "" && s.ddlProvider != nil {
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
	if normalizeDataFillingNodeType(req.NodeType) == "folder" {
		return nil
	}
	if strings.TrimSpace(req.TableName) == "" || req.DatasourceID <= 0 {
		return gorm.ErrInvalidData
	}
	return nil
}

func normalizeDataFillingNodeType(nodeType string) string {
	if strings.EqualFold(strings.TrimSpace(nodeType), "folder") {
		return "folder"
	}
	return "form"
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
