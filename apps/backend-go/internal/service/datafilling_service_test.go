package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"

	datafillingdomain "dataease/backend/internal/domain/datafilling"
	datasourcedomain "dataease/backend/internal/domain/datasource"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type fakeDataFillingRepo struct {
	records map[int64]*datafillingdomain.DataFillingForm
	nextID  int64
}

type fakeDataFillingDatasourceService struct {
	list []*datasourcedomain.CoreDatasource
	ds   map[int64]*datasourcedomain.CoreDatasource
}

type fakeDDLProvider struct {
	created         []string
	dropped         []string
	insertedRows    []map[string]interface{}
	updatedRows     []map[string]interface{}
	deletedBatches  [][]string
	searchRows      []map[string]interface{}
	countRows       int64
	truncatedTables []string
	columnData      []string
	addedColumns    [][]datafillingdomain.ExtTableField
	droppedColumns  [][]string
	err             error
}

type fakeCommitLogRepo struct {
	created []*datafillingdomain.DfCommitLog
	deleted []int64
	logs    []*datafillingdomain.DfCommitLog
	err     error
}

type fakeDatasourceConnProvider struct {
	db  *gorm.DB
	err error
}

func newFakeDataFillingRepo() *fakeDataFillingRepo {
	return &fakeDataFillingRepo{records: map[int64]*datafillingdomain.DataFillingForm{}, nextID: 1}
}

func (r *fakeDataFillingRepo) Create(ctx context.Context, form *datafillingdomain.DataFillingForm) error {
	_ = ctx
	cloned := *form
	if cloned.ID <= 0 {
		cloned.ID = r.nextID
		r.nextID++
	}
	r.records[cloned.ID] = &cloned
	form.ID = cloned.ID
	return nil
}

func (r *fakeDataFillingRepo) GetByID(ctx context.Context, id int64) (*datafillingdomain.DataFillingForm, error) {
	_ = ctx
	row, ok := r.records[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cloned := *row
	return &cloned, nil
}

func (r *fakeDataFillingRepo) Update(ctx context.Context, form *datafillingdomain.DataFillingForm) error {
	_ = ctx
	if _, ok := r.records[form.ID]; !ok {
		return gorm.ErrRecordNotFound
	}
	cloned := *form
	r.records[form.ID] = &cloned
	return nil
}

func (r *fakeDataFillingRepo) DeleteByID(ctx context.Context, id int64) error {
	_ = ctx
	delete(r.records, id)
	return nil
}

func (r *fakeDataFillingRepo) Rename(ctx context.Context, id int64, name string) error {
	_ = ctx
	row, ok := r.records[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	row.Name = name
	return nil
}

func (r *fakeDataFillingRepo) Move(ctx context.Context, id int64, pid int64) error {
	_ = ctx
	row, ok := r.records[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	level := 0
	if pid > 0 {
		parent, ok := r.records[pid]
		if !ok {
			return gorm.ErrRecordNotFound
		}
		level = parent.Level + 1
	}
	row.PID = pid
	row.Level = level
	return nil
}

func (r *fakeDataFillingRepo) GetTree(ctx context.Context) ([]*datafillingdomain.DataFillingForm, error) {
	_ = ctx
	result := make([]*datafillingdomain.DataFillingForm, 0, len(r.records))
	for _, row := range r.records {
		cloned := *row
		result = append(result, &cloned)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

func (r *fakeDataFillingRepo) GetByPID(ctx context.Context, pid int64) ([]*datafillingdomain.DataFillingForm, error) {
	return r.GetChildren(ctx, pid)
}

func (r *fakeDataFillingRepo) GetChildren(ctx context.Context, pid int64) ([]*datafillingdomain.DataFillingForm, error) {
	_ = ctx
	result := make([]*datafillingdomain.DataFillingForm, 0)
	for _, row := range r.records {
		if row.PID == pid {
			cloned := *row
			result = append(result, &cloned)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

func (f *fakeDataFillingDatasourceService) GetByID(id int64) (*datasourcedomain.CoreDatasource, error) {
	if ds, ok := f.ds[id]; ok {
		return ds, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeDataFillingDatasourceService) Tree(req *datasourcedomain.ListRequest) ([]*datasourcedomain.CoreDatasource, error) {
	_ = req
	return f.list, nil
}

func (f *fakeDataFillingDatasourceService) GetTables(req *datasourcedomain.TableRequest) ([]datasourcedomain.TableInfo, error) {
	_ = req
	return []datasourcedomain.TableInfo{}, nil
}

func (f *fakeDDLProvider) CreateTable(ctx context.Context, db *gorm.DB, tableName string, fields []datafillingdomain.ExtTableField) error {
	_ = ctx
	_ = db
	_ = fields
	f.created = append(f.created, tableName)
	return f.err
}

func (f *fakeDDLProvider) DropTable(ctx context.Context, db *gorm.DB, tableName string) error {
	_ = ctx
	_ = db
	f.dropped = append(f.dropped, tableName)
	return f.err
}

func (f *fakeDDLProvider) InsertRow(ctx context.Context, db *gorm.DB, tableName string, rowData map[string]interface{}) error {
	_ = ctx
	_ = db
	_ = tableName
	if strings.TrimSpace(fmt.Sprint(rowData["id"])) == "" || fmt.Sprint(rowData["id"]) == nilStringValue {
		rowData["id"] = "generated-id"
	}
	f.insertedRows = append(f.insertedRows, copyMap(rowData))
	return f.err
}

func (f *fakeDDLProvider) UpdateRow(ctx context.Context, db *gorm.DB, tableName string, id string, rowData map[string]interface{}) error {
	_ = ctx
	_ = db
	_ = tableName
	cloned := copyMap(rowData)
	cloned["id"] = id
	f.updatedRows = append(f.updatedRows, cloned)
	return f.err
}

func (f *fakeDDLProvider) DeleteRows(ctx context.Context, db *gorm.DB, tableName string, ids []string) error {
	_ = ctx
	_ = db
	_ = tableName
	f.deletedBatches = append(f.deletedBatches, append([]string{}, ids...))
	return f.err
}

func (f *fakeDDLProvider) SearchRows(ctx context.Context, db *gorm.DB, tableName string, whereClause string, args []interface{}, limit, offset int64) ([]map[string]interface{}, error) {
	_ = ctx
	_ = db
	_ = tableName
	_ = whereClause
	_ = args
	_ = limit
	_ = offset
	return f.searchRows, f.err
}

func (f *fakeDDLProvider) CountRows(ctx context.Context, db *gorm.DB, tableName string, whereClause string, args []interface{}) (int64, error) {
	_ = ctx
	_ = db
	_ = tableName
	_ = whereClause
	_ = args
	return f.countRows, f.err
}

func (f *fakeDDLProvider) TruncateTable(ctx context.Context, db *gorm.DB, tableName string) error {
	_ = ctx
	_ = db
	f.truncatedTables = append(f.truncatedTables, tableName)
	return f.err
}

func (f *fakeDDLProvider) ListColumnData(ctx context.Context, db *gorm.DB, tableName string, columnName string) ([]string, error) {
	_ = ctx
	_ = db
	_ = tableName
	_ = columnName
	return f.columnData, f.err
}

func (f *fakeDDLProvider) AddTableColumns(ctx context.Context, db *gorm.DB, tableName string, fields []datafillingdomain.ExtTableField) error {
	_ = ctx
	_ = db
	_ = tableName
	f.addedColumns = append(f.addedColumns, append([]datafillingdomain.ExtTableField{}, fields...))
	return f.err
}

func (f *fakeDDLProvider) DropTableColumns(ctx context.Context, db *gorm.DB, tableName string, columnNames []string) error {
	_ = ctx
	_ = db
	_ = tableName
	f.droppedColumns = append(f.droppedColumns, append([]string{}, columnNames...))
	return f.err
}

func (f *fakeCommitLogRepo) Create(ctx context.Context, log *datafillingdomain.DfCommitLog) error {
	_ = ctx
	f.created = append(f.created, log)
	return f.err
}

func (f *fakeCommitLogRepo) ListByFormID(ctx context.Context, formID int64, page, pageSize int) ([]*datafillingdomain.DfCommitLog, int64, error) {
	_ = ctx
	_ = formID
	_ = page
	_ = pageSize
	return f.logs, int64(len(f.logs)), f.err
}

func (f *fakeCommitLogRepo) DeleteByFormID(ctx context.Context, formID int64) error {
	_ = ctx
	f.deleted = append(f.deleted, formID)
	return f.err
}

func (f *fakeDatasourceConnProvider) GetDatasourceConnection(ctx context.Context, datasourceID int64) (*gorm.DB, error) {
	_ = ctx
	_ = datasourceID
	return f.db, f.err
}

func TestDataFillingService_SaveFolderSkipsDDL(t *testing.T) {
	repo := newFakeDataFillingRepo()
	ddl := &fakeDDLProvider{}
	svc := NewDataFillingService(repo, &fakeDataFillingDatasourceService{}, ddl, nil)

	item, err := svc.Save(context.Background(), &datafillingdomain.CreateFormRequest{Name: "root", NodeType: datafillingdomain.NodeTypeFolder}, 9)
	require.NoError(t, err)
	assert.Equal(t, int64(1), item.ID)
	assert.Empty(t, ddl.created)
}

func TestDataFillingService_RenameMoveTree(t *testing.T) {
	repo := newFakeDataFillingRepo()
	repo.records[1] = &datafillingdomain.DataFillingForm{ID: 1, Name: "A", NodeType: datafillingdomain.NodeTypeFolder, PID: 0, Level: 0}
	repo.records[2] = &datafillingdomain.DataFillingForm{ID: 2, Name: "B", NodeType: datafillingdomain.NodeTypeForm, PID: 1, Level: 1}
	svc := NewDataFillingService(repo, &fakeDataFillingDatasourceService{}, &fakeDDLProvider{}, nil)

	renamed, err := svc.Rename(context.Background(), 2, "B2")
	require.NoError(t, err)
	assert.Equal(t, "B2", renamed.Name)

	moved, err := svc.Move(context.Background(), 2, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(0), moved.PID)

	kw := "b2"
	tree, err := svc.Tree(context.Background(), &datafillingdomain.TreeRequest{Keyword: &kw})
	require.NoError(t, err)
	require.Len(t, tree, 1)
	assert.Equal(t, int64(2), tree[0].ID)
}

func TestDataFillingService_ListDatasourceList(t *testing.T) {
	enabled := true
	disabled := false
	svc := NewDataFillingService(newFakeDataFillingRepo(), &fakeDataFillingDatasourceService{list: []*datasourcedomain.CoreDatasource{{ID: 1, Name: "mysql-a", Type: "mysql", EnableDataFill: &enabled}, {ID: 2, Name: "mysql-b", Type: "mysql", EnableDataFill: &disabled}}}, &fakeDDLProvider{}, nil)

	list, err := svc.ListDatasourceList(context.Background())
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(1), list[0].ID)

	all, err := svc.ListDatasourceListAll(context.Background())
	require.NoError(t, err)
	assert.Len(t, all, 2)
}

func TestDataFillingService_CreatePhysicalTableParsesFields(t *testing.T) {
	repo := newFakeDataFillingRepo()
	ddl := &fakeDDLProvider{}
	formFields, err := json.Marshal([]datafillingdomain.ExtTableField{{Settings: datafillingdomain.ExtTableFieldSetting{Mapping: datafillingdomain.ExtTableFieldMapping{ColumnName: "name", Type: datafillingdomain.BaseTypeNvarchar}}}})
	require.NoError(t, err)

	conf := `{"host":"127.0.0.1","port":3306,"dataBase":"demo","username":"u","password":"p"}`
	svc := NewDataFillingService(repo, &fakeDataFillingDatasourceService{ds: map[int64]*datasourcedomain.CoreDatasource{8: {ID: 8, Type: "mysql", Configuration: &conf}}}, ddl, nil)

	_, err = svc.Save(context.Background(), &datafillingdomain.CreateFormRequest{Name: "form", NodeType: datafillingdomain.NodeTypeForm, TableName: "df_form_1", DatasourceID: 8, Forms: string(formFields), UseExistsTable: true}, 1)
	require.NoError(t, err)
	assert.Empty(t, ddl.created)
}

func TestDataFillingService_SearchSaveDeleteAndLogs(t *testing.T) {
	repo := newFakeDataFillingRepo()
	repo.records[1] = &datafillingdomain.DataFillingForm{ID: 1, Name: "form", NodeType: datafillingdomain.NodeTypeForm, PhysicalTableName: "df_demo", DatasourceID: 8, Forms: "[]"}
	ddl := &fakeDDLProvider{searchRows: []map[string]interface{}{{"id": "1", "name": "alice"}}, countRows: 1, columnData: []string{"alice", "bob"}}
	logs := &fakeCommitLogRepo{logs: []*datafillingdomain.DfCommitLog{{ID: 1, FormID: 1}}}
	svc := NewDataFillingService(repo, &fakeDataFillingDatasourceService{}, ddl, logs)
	svc.datasourceConnProvider = &fakeDatasourceConnProvider{}

	resp, err := svc.SearchTableData(context.Background(), 1, &datafillingdomain.TableDataRequest{CurrentPage: 2, PageSize: 5, SearchParams: []datafillingdomain.SearchParam{{Field: "name", Term: "eq", Value: "alice"}}})
	require.NoError(t, err)
	assert.Equal(t, int64(1), resp.Total)
	assert.Equal(t, int64(2), resp.CurrentPage)

	saved, err := svc.SaveRowData(context.Background(), 1, map[string]interface{}{"name": "alice"}, 9, "tester")
	require.NoError(t, err)
	assert.Equal(t, "generated-id", saved.Data[0]["id"])
	require.Len(t, logs.created, 1)
	assert.Equal(t, 1, logs.created[0].Operate)

	_, err = svc.SaveRowData(context.Background(), 1, map[string]interface{}{"id": "row-1", "name": "updated"}, 9, "tester")
	require.NoError(t, err)
	require.Len(t, ddl.updatedRows, 1)
	assert.Equal(t, "row-1", ddl.updatedRows[0]["id"])

	err = svc.DeleteRowData(context.Background(), 1, "row-1", 9, "tester")
	require.NoError(t, err)
	require.Len(t, ddl.deletedBatches, 1)

	err = svc.BatchDeleteRowData(context.Background(), 1, []string{"a", "b"}, 9, "tester")
	require.NoError(t, err)
	require.Len(t, ddl.deletedBatches, 2)
	assert.Equal(t, []string{"a", "b"}, ddl.deletedBatches[1])

	err = svc.TruncateTableData(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, []string{"df_demo"}, ddl.truncatedTables)

	values, err := svc.ListColumnData(context.Background(), 1, "name")
	require.NoError(t, err)
	assert.Equal(t, []string{"alice", "bob"}, values)

	logRows, total, err := svc.ListCommitLogs(context.Background(), 1, 1, 10)
	require.NoError(t, err)
	assert.Len(t, logRows, 1)
	assert.Equal(t, int64(1), total)

	err = svc.ClearCommitLogs(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, []int64{1}, logs.deleted)
}

func TestDataFillingService_UpdateAltersColumns(t *testing.T) {
	repo := newFakeDataFillingRepo()
	oldFields := `[ {"settings":{"mapping":{"columnName":"name","type":"nvarchar"}}} ]`
	newFields := `[ {"settings":{"mapping":{"columnName":"name","type":"nvarchar"}}}, {"settings":{"mapping":{"columnName":"age","type":"number"}}} ]`
	repo.records[1] = &datafillingdomain.DataFillingForm{ID: 1, Name: "form", PID: 0, Level: 0, NodeType: datafillingdomain.NodeTypeForm, PhysicalTableName: "df_demo", DatasourceID: 8, Forms: oldFields}
	ddl := &fakeDDLProvider{}
	svc := NewDataFillingService(repo, &fakeDataFillingDatasourceService{}, ddl, nil)
	svc.datasourceConnProvider = &fakeDatasourceConnProvider{}

	_, err := svc.Update(context.Background(), &datafillingdomain.UpdateFormRequest{ID: 1, CreateFormRequest: datafillingdomain.CreateFormRequest{Name: "form", NodeType: datafillingdomain.NodeTypeForm, TableName: "df_demo", DatasourceID: 8, Forms: newFields}}, 1)
	require.NoError(t, err)
	require.Len(t, ddl.addedColumns, 1)
	assert.Equal(t, "age", ddl.addedColumns[0][0].Settings.Mapping.ColumnName)
	assert.Empty(t, ddl.droppedColumns)
}

func copyMap(input map[string]interface{}) map[string]interface{} {
	cloned := make(map[string]interface{}, len(input))
	for key, value := range input {
		cloned[key] = value
	}
	return cloned
}
