package service

import (
	"context"
	"encoding/json"
	"sort"
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
	created []string
	dropped []string
	err     error
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

func TestDataFillingService_SaveFolderSkipsDDL(t *testing.T) {
	repo := newFakeDataFillingRepo()
	ddl := &fakeDDLProvider{}
	svc := NewDataFillingService(repo, &fakeDataFillingDatasourceService{}, ddl)

	item, err := svc.Save(context.Background(), &datafillingdomain.CreateFormRequest{Name: "root", NodeType: datafillingdomain.NodeTypeFolder}, 9)
	require.NoError(t, err)
	assert.Equal(t, int64(1), item.ID)
	assert.Empty(t, ddl.created)
}

func TestDataFillingService_RenameMoveTree(t *testing.T) {
	repo := newFakeDataFillingRepo()
	repo.records[1] = &datafillingdomain.DataFillingForm{ID: 1, Name: "A", NodeType: datafillingdomain.NodeTypeFolder, PID: 0, Level: 0}
	repo.records[2] = &datafillingdomain.DataFillingForm{ID: 2, Name: "B", NodeType: datafillingdomain.NodeTypeForm, PID: 1, Level: 1}
	svc := NewDataFillingService(repo, &fakeDataFillingDatasourceService{}, &fakeDDLProvider{})

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
	svc := NewDataFillingService(newFakeDataFillingRepo(), &fakeDataFillingDatasourceService{list: []*datasourcedomain.CoreDatasource{{ID: 1, Name: "mysql-a", Type: "mysql", EnableDataFill: &enabled}, {ID: 2, Name: "mysql-b", Type: "mysql", EnableDataFill: &disabled}}}, &fakeDDLProvider{})

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
	svc := NewDataFillingService(repo, &fakeDataFillingDatasourceService{ds: map[int64]*datasourcedomain.CoreDatasource{8: {ID: 8, Type: "mysql", Configuration: &conf}}}, ddl)

	_, err = svc.Save(context.Background(), &datafillingdomain.CreateFormRequest{Name: "form", NodeType: datafillingdomain.NodeTypeForm, TableName: "df_form_1", DatasourceID: 8, Forms: string(formFields), UseExistsTable: true}, 1)
	require.NoError(t, err)
	assert.Empty(t, ddl.created)
}
