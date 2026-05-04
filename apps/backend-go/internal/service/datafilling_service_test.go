package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	datafillingdomain "dataease/backend/internal/domain/datafilling"
	datasourcedomain "dataease/backend/internal/domain/datasource"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
	"gorm.io/driver/sqlite"
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

type fakeTaskRepo struct {
	records map[int64]*datafillingdomain.DataFillingTask
	nextID  int64
}

type fakeSubTaskRepo struct {
	records     map[int64]*datafillingdomain.DataFillingSubTask
	nextID      int64
	decremented []int64
}

type fakeSubInstanceRepo struct {
	records           []*datafillingdomain.DataFillingSubInstance
	statusUpdates     []int64
	statusUpdateValue []int
}

type userTaskPageSubInstanceRepo struct {
	*fakeSubInstanceRepo
	rows  []*datafillingdomain.UserTaskVO
	total int64
	err   error
}

type datasourceOptionRow struct {
	Name      string `gorm:"column:name"`
	SortValue int    `gorm:"column:sort_value"`
}

func (datasourceOptionRow) TableName() string { return "option_table" }

func newFakeDataFillingRepo() *fakeDataFillingRepo {
	return &fakeDataFillingRepo{records: map[int64]*datafillingdomain.DataFillingForm{}, nextID: 1}
}

func newFakeTaskRepo() *fakeTaskRepo {
	return &fakeTaskRepo{records: map[int64]*datafillingdomain.DataFillingTask{}, nextID: 1}
}

func newFakeSubTaskRepo() *fakeSubTaskRepo {
	return &fakeSubTaskRepo{records: map[int64]*datafillingdomain.DataFillingSubTask{}, nextID: 1}
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

func (r *fakeTaskRepo) CreateTask(ctx context.Context, task *datafillingdomain.DataFillingTask) error {
	_ = ctx
	cloned := *task
	if cloned.ID <= 0 {
		cloned.ID = r.nextID
		r.nextID++
	}
	r.records[cloned.ID] = &cloned
	task.ID = cloned.ID
	return nil
}

func (r *fakeTaskRepo) UpdateTask(ctx context.Context, task *datafillingdomain.DataFillingTask) error {
	_ = ctx
	if _, ok := r.records[task.ID]; !ok {
		return gorm.ErrRecordNotFound
	}
	cloned := *task
	r.records[task.ID] = &cloned
	return nil
}

func (r *fakeTaskRepo) GetTaskByID(ctx context.Context, taskID int64) (*datafillingdomain.DataFillingTask, error) {
	_ = ctx
	row, ok := r.records[taskID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cloned := *row
	return &cloned, nil
}

func (r *fakeTaskRepo) ListTasksByFormID(ctx context.Context, formID int64, page, pageSize int) ([]*datafillingdomain.DataFillingTask, int64, error) {
	_ = ctx
	rows := make([]*datafillingdomain.DataFillingTask, 0)
	for _, row := range r.records {
		if row.FormID == formID {
			cloned := *row
			rows = append(rows, &cloned)
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].ID > rows[j].ID })
	total := int64(len(rows))
	start := 0
	if page > 1 {
		start = (page - 1) * pageSize
	}
	if start >= len(rows) {
		return []*datafillingdomain.DataFillingTask{}, total, nil
	}
	end := start + pageSize
	if end > len(rows) {
		end = len(rows)
	}
	return rows[start:end], total, nil
}

func (r *fakeTaskRepo) DeleteTasksByIDs(ctx context.Context, taskIDs []int64) error {
	_ = ctx
	for _, id := range taskIDs {
		delete(r.records, id)
	}
	return nil
}

func (r *fakeTaskRepo) GetStartedTasks(ctx context.Context) ([]*datafillingdomain.DataFillingTask, error) {
	_ = ctx
	rows := make([]*datafillingdomain.DataFillingTask, 0)
	for _, row := range r.records {
		if row.Status == datafillingdomain.TaskStatusStarted {
			cloned := *row
			rows = append(rows, &cloned)
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].ID < rows[j].ID })
	return rows, nil
}

func (r *fakeSubTaskRepo) CreateSubTask(ctx context.Context, subTask *datafillingdomain.DataFillingSubTask) error {
	_ = ctx
	cloned := *subTask
	if cloned.ID <= 0 {
		cloned.ID = r.nextID
		r.nextID++
	}
	r.records[cloned.ID] = &cloned
	subTask.ID = cloned.ID
	return nil
}

func (r *fakeSubTaskRepo) GetSubTaskByID(ctx context.Context, subTaskID int64) (*datafillingdomain.DataFillingSubTask, error) {
	_ = ctx
	row, ok := r.records[subTaskID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cloned := *row
	return &cloned, nil
}

func (r *fakeSubTaskRepo) UpdateSubTaskCounts(ctx context.Context, subTaskID int64, totalCount, unfinishedCount, totalUserCount, unfinishedUserCount int) error {
	_ = ctx
	row, ok := r.records[subTaskID]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	row.TotalCount = totalCount
	row.UnfinishedCount = unfinishedCount
	row.TotalUserCount = totalUserCount
	row.UnfinishedUserCount = unfinishedUserCount
	return nil
}

func (r *fakeSubTaskRepo) ListSubTasksByTaskID(ctx context.Context, taskID int64, page, pageSize int) ([]*datafillingdomain.DataFillingSubTask, int64, error) {
	_ = ctx
	rows := make([]*datafillingdomain.DataFillingSubTask, 0)
	for _, row := range r.records {
		if row.TaskID == taskID {
			cloned := *row
			rows = append(rows, &cloned)
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].ID > rows[j].ID })
	total := int64(len(rows))
	start := 0
	if page > 1 {
		start = (page - 1) * pageSize
	}
	if start >= len(rows) {
		return []*datafillingdomain.DataFillingSubTask{}, total, nil
	}
	end := start + pageSize
	if end > len(rows) {
		end = len(rows)
	}
	return rows[start:end], total, nil
}

func (r *fakeSubTaskRepo) DeleteSubTasksByIDs(ctx context.Context, subTaskIDs []int64) error {
	_ = ctx
	for _, id := range subTaskIDs {
		delete(r.records, id)
	}
	return nil
}

func (r *fakeSubTaskRepo) ListSubTaskIDsByTaskIDs(ctx context.Context, taskIDs []int64) ([]int64, error) {
	_ = ctx
	lookup := make(map[int64]struct{}, len(taskIDs))
	for _, id := range taskIDs {
		lookup[id] = struct{}{}
	}
	result := make([]int64, 0)
	for _, row := range r.records {
		if _, ok := lookup[row.TaskID]; ok {
			result = append(result, row.ID)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result, nil
}

func (r *fakeSubTaskRepo) DecrementSubTaskUnfinishedCount(ctx context.Context, subTaskID int64) error {
	_ = ctx
	row, ok := r.records[subTaskID]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	if row.UnfinishedCount > 0 {
		row.UnfinishedCount--
	}
	if row.UnfinishedUserCount > 0 {
		row.UnfinishedUserCount--
	}
	r.decremented = append(r.decremented, subTaskID)
	return nil
}

func (r *fakeSubInstanceRepo) BatchCreateSubInstances(ctx context.Context, instances []*datafillingdomain.DataFillingSubInstance) error {
	_ = ctx
	for _, instance := range instances {
		cloned := *instance
		cloned.ID = int64(len(r.records) + 1)
		r.records = append(r.records, &cloned)
		instance.ID = cloned.ID
	}
	return nil
}

func (r *fakeSubInstanceRepo) DeleteSubInstancesByPID(ctx context.Context, pid int64) error {
	_ = ctx
	return r.DeleteSubInstancesByPIDs(context.Background(), []int64{pid})
}

func (r *fakeSubInstanceRepo) DeleteSubInstancesByPIDs(ctx context.Context, pids []int64) error {
	_ = ctx
	lookup := make(map[int64]struct{}, len(pids))
	for _, id := range pids {
		lookup[id] = struct{}{}
	}
	filtered := r.records[:0]
	for _, row := range r.records {
		if _, ok := lookup[row.PID]; !ok {
			filtered = append(filtered, row)
		}
	}
	r.records = filtered
	return nil
}

func (r *fakeSubInstanceRepo) DeleteSubInstancesByTaskIDs(ctx context.Context, taskIDs []int64) error {
	_ = ctx
	lookup := make(map[int64]struct{}, len(taskIDs))
	for _, id := range taskIDs {
		lookup[id] = struct{}{}
	}
	filtered := r.records[:0]
	for _, row := range r.records {
		if _, ok := lookup[row.TaskID]; !ok {
			filtered = append(filtered, row)
		}
	}
	r.records = filtered
	return nil
}

func (r *fakeSubInstanceRepo) ListSubInstancesByPID(ctx context.Context, pid int64, statusFilter *int) ([]*datafillingdomain.DataFillingSubInstance, error) {
	_ = ctx
	rows := make([]*datafillingdomain.DataFillingSubInstance, 0)
	for _, row := range r.records {
		if row.PID != pid {
			continue
		}
		if statusFilter != nil && row.Status != *statusFilter {
			continue
		}
		cloned := *row
		rows = append(rows, &cloned)
	}
	return rows, nil
}

func (r *fakeSubInstanceRepo) ListSubInstancesByUID(ctx context.Context, uid int64, page, pageSize int, req *datafillingdomain.UserTaskPageRequest) ([]*datafillingdomain.UserTaskVO, int64, error) {
	_ = ctx
	_ = page
	_ = pageSize
	result := make([]*datafillingdomain.UserTaskVO, 0)
	for _, row := range r.records {
		if row.UID != uid {
			continue
		}
		if req != nil && req.Type != nil && row.Status != *req.Type {
			continue
		}
		finishCount := int64(0)
		if row.Status == datafillingdomain.SubInstanceStatusFinished {
			finishCount = 1
		}
		var finishTime *int64
		if row.FinishTime > 0 {
			value := row.FinishTime
			finishTime = &value
		}
		result = append(result, &datafillingdomain.UserTaskVO{ID: row.ID, TaskID: row.TaskID, FormID: row.FormID, Status: row.Status, FinishTime: finishTime, TotalCount: 1, FinishCount: finishCount})
	}
	return result, int64(len(result)), nil
}

func (r *fakeSubInstanceRepo) CountOpenSubInstancesByUID(ctx context.Context, uid int64) (int64, error) {
	_ = ctx
	var total int64
	for _, row := range r.records {
		if row.UID == uid && row.Status == datafillingdomain.SubInstanceStatusOpen {
			total++
		}
	}
	return total, nil
}

func (r *fakeSubInstanceRepo) GetSubInstanceByID(ctx context.Context, id int64) (*datafillingdomain.DataFillingSubInstance, error) {
	_ = ctx
	for _, row := range r.records {
		if row.ID == id {
			cloned := *row
			return &cloned, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *fakeSubInstanceRepo) GetSubInstanceByPIDAndUID(ctx context.Context, pid, uid int64) ([]*datafillingdomain.DataFillingSubInstance, error) {
	_ = ctx
	rows := make([]*datafillingdomain.DataFillingSubInstance, 0)
	for _, row := range r.records {
		if row.PID == pid && row.UID == uid {
			cloned := *row
			rows = append(rows, &cloned)
		}
	}
	return rows, nil
}

func (r *fakeSubInstanceRepo) UpdateSubInstanceStatus(ctx context.Context, id int64, status int, finishTime int64) error {
	_ = ctx
	for _, row := range r.records {
		if row.ID == id {
			row.Status = status
			row.FinishTime = finishTime
			r.statusUpdates = append(r.statusUpdates, id)
			r.statusUpdateValue = append(r.statusUpdateValue, status)
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (r *userTaskPageSubInstanceRepo) ListSubInstancesByUID(ctx context.Context, uid int64, page, pageSize int, req *datafillingdomain.UserTaskPageRequest) ([]*datafillingdomain.UserTaskVO, int64, error) {
	_ = ctx
	_ = uid
	_ = page
	_ = pageSize
	_ = req
	return r.rows, r.total, r.err
}

func TestDataFillingService_SaveFolderSkipsDDL(t *testing.T) {
	repo := newFakeDataFillingRepo()
	ddl := &fakeDDLProvider{}
	svc := NewDataFillingService(repo, &fakeDataFillingDatasourceService{}, ddl, nil, nil, nil, nil, nil)

	item, err := svc.Save(context.Background(), &datafillingdomain.CreateFormRequest{Name: "root", NodeType: datafillingdomain.NodeTypeFolder}, 9)
	require.NoError(t, err)
	assert.Equal(t, int64(1), item.ID)
	assert.Empty(t, ddl.created)
}

func TestDataFillingService_Get(t *testing.T) {
	repo := newFakeDataFillingRepo()
	repo.records[1] = &datafillingdomain.DataFillingForm{ID: 1, Name: "form"}
	svc := NewDataFillingService(repo, &fakeDataFillingDatasourceService{}, &fakeDDLProvider{}, nil, nil, nil, nil, nil)

	item, err := svc.Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), item.ID)

	_, err = svc.Get(context.Background(), 99)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestDataFillingService_RenameMoveTree(t *testing.T) {
	repo := newFakeDataFillingRepo()
	repo.records[1] = &datafillingdomain.DataFillingForm{ID: 1, Name: "A", NodeType: datafillingdomain.NodeTypeFolder, PID: 0, Level: 0}
	repo.records[2] = &datafillingdomain.DataFillingForm{ID: 2, Name: "B", NodeType: datafillingdomain.NodeTypeForm, PID: 1, Level: 1}
	svc := NewDataFillingService(repo, &fakeDataFillingDatasourceService{}, &fakeDDLProvider{}, nil, nil, nil, nil, nil)

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

func TestDataFillingService_Delete_CascadesAndDropsPhysicalTable(t *testing.T) {
	repo := newFakeDataFillingRepo()
	repo.records[1] = &datafillingdomain.DataFillingForm{ID: 1, Name: "folder", NodeType: datafillingdomain.NodeTypeFolder}
	repo.records[2] = &datafillingdomain.DataFillingForm{ID: 2, Name: "form", PID: 1, NodeType: datafillingdomain.NodeTypeForm, PhysicalTableName: "df_demo", DatasourceID: 8}
	ddl := &fakeDDLProvider{}
	svc := NewDataFillingService(repo, &fakeDataFillingDatasourceService{}, ddl, nil, nil, nil, nil, nil)
	svc.datasourceConnProvider = &fakeDatasourceConnProvider{}

	err := svc.Delete(context.Background(), 1)
	require.NoError(t, err)
	assert.Empty(t, repo.records)
	assert.Equal(t, []string{"df_demo"}, ddl.dropped)
}

func TestDataFillingService_ListDatasourceList(t *testing.T) {
	enabled := true
	disabled := false
	svc := NewDataFillingService(newFakeDataFillingRepo(), &fakeDataFillingDatasourceService{list: []*datasourcedomain.CoreDatasource{{ID: 1, Name: "mysql-a", Type: "mysql", EnableDataFill: &enabled}, {ID: 2, Name: "mysql-b", Type: "mysql", EnableDataFill: &disabled}}}, &fakeDDLProvider{}, nil, nil, nil, nil, nil)

	list, err := svc.ListDatasourceList(context.Background())
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(1), list[0].ID)

	all, err := svc.ListDatasourceListAll(context.Background())
	require.NoError(t, err)
	assert.Len(t, all, 2)
}

func TestDataFillingService_GetBuiltInTables(t *testing.T) {
	svc := NewDataFillingService(newFakeDataFillingRepo(), &fakeDataFillingDatasourceService{}, &fakeDDLProvider{}, nil, nil, nil, nil, nil)

	tables, err := svc.GetBuiltInTables(context.Background())
	require.NoError(t, err)
	assert.Empty(t, tables)
}

func TestDataFillingService_CreatePhysicalTableParsesFields(t *testing.T) {
	repo := newFakeDataFillingRepo()
	ddl := &fakeDDLProvider{}
	formFields, err := json.Marshal([]datafillingdomain.ExtTableField{{Settings: datafillingdomain.ExtTableFieldSetting{Mapping: datafillingdomain.ExtTableFieldMapping{ColumnName: "name", Type: datafillingdomain.BaseTypeNvarchar}}}})
	require.NoError(t, err)

	conf := `{"host":"127.0.0.1","port":3306,"dataBase":"demo","username":"u","password":"p"}`
	svc := NewDataFillingService(repo, &fakeDataFillingDatasourceService{ds: map[int64]*datasourcedomain.CoreDatasource{8: {ID: 8, Type: "mysql", Configuration: &conf}}}, ddl, nil, nil, nil, nil, nil)

	_, err = svc.Save(context.Background(), &datafillingdomain.CreateFormRequest{Name: "form", NodeType: datafillingdomain.NodeTypeForm, TableName: "df_form_1", DatasourceID: 8, Forms: string(formFields), UseExistsTable: true}, 1)
	require.NoError(t, err)
	assert.Empty(t, ddl.created)
}

func TestDataFillingService_SaveFormCreatesPhysicalTable(t *testing.T) {
	repo := newFakeDataFillingRepo()
	ddl := &fakeDDLProvider{}
	svc := NewDataFillingService(repo, &fakeDataFillingDatasourceService{}, ddl, nil, nil, nil, nil, nil)
	svc.SetDatasourceConnectionProvider(&fakeDatasourceConnProvider{})

	form, err := svc.Save(context.Background(), &datafillingdomain.CreateFormRequest{Name: "form", NodeType: datafillingdomain.NodeTypeForm, TableName: "df_new", DatasourceID: 8, Forms: `[{"settings":{"mapping":{"columnName":"name","type":"nvarchar"}}}]`}, 7)
	require.NoError(t, err)
	assert.Equal(t, int64(1), form.ID)
	assert.Equal(t, []string{"df_new"}, ddl.created)
}

func TestDataFillingService_SearchSaveDeleteAndLogs(t *testing.T) {
	repo := newFakeDataFillingRepo()
	repo.records[1] = &datafillingdomain.DataFillingForm{ID: 1, Name: "form", NodeType: datafillingdomain.NodeTypeForm, PhysicalTableName: "df_demo", DatasourceID: 8, Forms: "[]"}
	ddl := &fakeDDLProvider{searchRows: []map[string]interface{}{{"id": "1", "name": "alice"}}, countRows: 1, columnData: []string{"alice", "bob"}}
	logs := &fakeCommitLogRepo{logs: []*datafillingdomain.DfCommitLog{{ID: 1, FormID: 1}}}
	svc := NewDataFillingService(repo, &fakeDataFillingDatasourceService{}, ddl, logs, nil, nil, nil, nil)
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
	svc := NewDataFillingService(repo, &fakeDataFillingDatasourceService{}, ddl, nil, nil, nil, nil, nil)
	svc.datasourceConnProvider = &fakeDatasourceConnProvider{}

	_, err := svc.Update(context.Background(), &datafillingdomain.UpdateFormRequest{ID: 1, CreateFormRequest: datafillingdomain.CreateFormRequest{Name: "form", NodeType: datafillingdomain.NodeTypeForm, TableName: "df_demo", DatasourceID: 8, Forms: newFields}}, 1)
	require.NoError(t, err)
	require.Len(t, ddl.addedColumns, 1)
	assert.Equal(t, "age", ddl.addedColumns[0][0].Settings.Mapping.ColumnName)
	assert.Empty(t, ddl.droppedColumns)
}

func TestDataFillingService_TaskLifecycle(t *testing.T) {
	formRepo := newFakeDataFillingRepo()
	formRepo.records[1] = &datafillingdomain.DataFillingForm{ID: 1, Name: "form", NodeType: datafillingdomain.NodeTypeForm}
	taskRepo := newFakeTaskRepo()
	subTaskRepo := newFakeSubTaskRepo()
	subInstanceRepo := &fakeSubInstanceRepo{}
	scheduler := NewDataFillingScheduler(taskRepo, subTaskRepo, subInstanceRepo, formRepo)
	svc := NewDataFillingService(formRepo, &fakeDataFillingDatasourceService{}, &fakeDDLProvider{}, nil, taskRepo, subTaskRepo, subInstanceRepo, scheduler)

	taskID, err := svc.SaveTask(context.Background(), &datafillingdomain.TaskSaveRequest{FormID: 1, Name: "task-a", UIDList: []int64{9, 10}, RateType: 1, RateVal: "09:30:00"}, 7)
	require.NoError(t, err)
	assert.Equal(t, int64(1), taskID)

	info, err := svc.GetTaskInfo(context.Background(), taskID)
	require.NoError(t, err)
	assert.Equal(t, []int64{9, 10}, info.UIDList)

	err = svc.StartTask(context.Background(), 1, taskID)
	require.NoError(t, err)
	started, err := taskRepo.GetTaskByID(context.Background(), taskID)
	require.NoError(t, err)
	assert.Equal(t, datafillingdomain.TaskStatusStarted, started.Status)
	assert.NotZero(t, started.NextExecTime)

	err = svc.ExecuteNowTask(context.Background(), taskID)
	require.NoError(t, err)
	assert.Len(t, subTaskRepo.records, 1)
	assert.Len(t, subInstanceRepo.records, 2)

	page, err := svc.TaskPageList(context.Background(), 1, 1, 10)
	require.NoError(t, err)
	require.Len(t, page.Records, 1)
	assert.Equal(t, int64(1), page.Total)

	subPage, err := svc.SubTaskPageList(context.Background(), taskID, 1, 10)
	require.NoError(t, err)
	require.Len(t, subPage.Records, 1)

	users, err := svc.SubTaskUsersList(context.Background(), subPage.Records[0].ID, "unfinished")
	require.NoError(t, err)
	assert.Len(t, users, 2)

	err = svc.StopTask(context.Background(), 1, taskID)
	require.NoError(t, err)
	stopped, err := taskRepo.GetTaskByID(context.Background(), taskID)
	require.NoError(t, err)
	assert.Equal(t, datafillingdomain.TaskStatusStopped, stopped.Status)

	err = svc.DeleteTasks(context.Background(), 1, []int64{taskID})
	require.NoError(t, err)
	assert.Empty(t, taskRepo.records)
	assert.Empty(t, subTaskRepo.records)
	assert.Empty(t, subInstanceRepo.records)
}

func TestDataFillingService_UpdateExistingTask_ReregistersStartedTask(t *testing.T) {
	formRepo := newFakeDataFillingRepo()
	formRepo.records[1] = &datafillingdomain.DataFillingForm{ID: 1, Name: "form"}
	taskRepo := newFakeTaskRepo()
	taskRepo.records[1] = &datafillingdomain.DataFillingTask{ID: 1, FormID: 1, Name: "old", UIDList: `[1]`, RIDList: `[2]`, ReciFlagList: `[1]`, RateType: 1, RateVal: "08:00:00", Status: datafillingdomain.TaskStatusStarted}
	scheduler := NewDataFillingScheduler(taskRepo, newFakeSubTaskRepo(), &fakeSubInstanceRepo{}, formRepo)
	require.NoError(t, scheduler.RegisterTask(context.Background(), 1))
	svc := NewDataFillingService(formRepo, &fakeDataFillingDatasourceService{}, &fakeDDLProvider{}, nil, taskRepo, newFakeSubTaskRepo(), &fakeSubInstanceRepo{}, scheduler)
	taskID := int64(1)
	req := &datafillingdomain.TaskSaveRequest{ID: &taskID, FormID: 1, Name: "updated", FillType: 2, FitType: 3, FitColumn: " city ", RateType: 2, RateVal: "7 09:30:15", OneTimeType: 1, StartTime: 10, EndTime: 20, PublishRangeTime: 30, PublishRangeTimeType: 4, FormExtSetting: " ext ", FormFilterSetting: " filter "}

	id, err := svc.updateExistingTask(context.Background(), req, 9, `[1,2]`, `[9,10]`, `[11]`)
	require.NoError(t, err)
	assert.Equal(t, int64(1), id)
	updated, err := taskRepo.GetTaskByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "updated", updated.Name)
	assert.Equal(t, `[1,2]`, updated.ReciFlagList)
	assert.Equal(t, `[9,10]`, updated.UIDList)
	assert.Equal(t, `[11]`, updated.RIDList)
	assert.Equal(t, 2, updated.FillType)
	assert.Equal(t, 3, updated.FitType)
	assert.Equal(t, "city", updated.FitColumn)
	assert.Equal(t, 2, updated.RateType)
	assert.Equal(t, "7 09:30:15", updated.RateVal)
	assert.Equal(t, "ext", updated.FormExtSetting)
	assert.Equal(t, "filter", updated.FormFilterSetting)
	assert.Equal(t, int64(9), updated.UpdateBy)
	assert.NotZero(t, updated.UpdateTime)
	assert.NotZero(t, updated.NextExecTime)
	assert.Len(t, scheduler.entries, 1)
	_, ok := scheduler.entries[1]
	assert.True(t, ok)
}

func TestDataFillingService_DeleteSubTasks(t *testing.T) {
	formRepo := newFakeDataFillingRepo()
	formRepo.records[1] = &datafillingdomain.DataFillingForm{ID: 1, Name: "form"}
	subTaskRepo := newFakeSubTaskRepo()
	subTaskRepo.records[10] = &datafillingdomain.DataFillingSubTask{ID: 10, TaskID: 1}
	subTaskRepo.records[11] = &datafillingdomain.DataFillingSubTask{ID: 11, TaskID: 1}
	subInstanceRepo := &fakeSubInstanceRepo{records: []*datafillingdomain.DataFillingSubInstance{{ID: 100, PID: 10}, {ID: 101, PID: 11}, {ID: 102, PID: 99}}}
	svc := NewDataFillingService(formRepo, &fakeDataFillingDatasourceService{}, &fakeDDLProvider{}, nil, nil, subTaskRepo, subInstanceRepo, nil)

	err := svc.DeleteSubTasks(context.Background(), 1, []int64{10, 11})
	require.NoError(t, err)
	assert.Empty(t, subTaskRepo.records)
	require.Len(t, subInstanceRepo.records, 1)
	assert.Equal(t, int64(102), subInstanceRepo.records[0].ID)
}

func TestDataFillingService_UserTaskPageList(t *testing.T) {
	repo := &userTaskPageSubInstanceRepo{fakeSubInstanceRepo: &fakeSubInstanceRepo{}, rows: []*datafillingdomain.UserTaskVO{{ID: 1, TaskID: 2, FormID: 3, EndTime: time.Now().Add(-time.Minute).UnixMilli()}, {ID: 2, TaskID: 2, FormID: 3, EndTime: time.Now().Add(time.Minute).UnixMilli()}}, total: 2}
	svc := NewDataFillingService(newFakeDataFillingRepo(), &fakeDataFillingDatasourceService{}, &fakeDDLProvider{}, nil, nil, nil, repo, nil)
	typeFilter := datafillingdomain.SubInstanceStatusOpen

	rows, total, err := svc.UserTaskPageList(context.Background(), 9, 1, 10, &datafillingdomain.UserTaskPageRequest{Type: &typeFilter})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, rows, 2)
	assert.True(t, rows[0].Expired)
	assert.False(t, rows[1].Expired)
}

func TestDataFillingService_UserTaskTodoCount(t *testing.T) {
	repo := &fakeSubInstanceRepo{records: []*datafillingdomain.DataFillingSubInstance{{ID: 1, UID: 9, Status: datafillingdomain.SubInstanceStatusOpen}, {ID: 2, UID: 9, Status: datafillingdomain.SubInstanceStatusFinished}, {ID: 3, UID: 10, Status: datafillingdomain.SubInstanceStatusOpen}}}
	svc := NewDataFillingService(newFakeDataFillingRepo(), &fakeDataFillingDatasourceService{}, &fakeDDLProvider{}, nil, nil, nil, repo, nil)

	count, err := svc.UserTaskTodoCount(context.Background(), 9)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestDataFillingService_GetUserTaskData(t *testing.T) {
	formRepo := newFakeDataFillingRepo()
	formRepo.records[1] = &datafillingdomain.DataFillingForm{ID: 1, Name: "form", Forms: `[{"settings":{"name":"姓名","mapping":{"columnName":"name","type":"nvarchar"}}}]`}
	taskRepo := newFakeTaskRepo()
	taskRepo.records[1] = &datafillingdomain.DataFillingTask{ID: 1, FormID: 1, FormExtSetting: "ext", FillType: 2}
	subTaskRepo := newFakeSubTaskRepo()
	subTaskRepo.records[10] = &datafillingdomain.DataFillingSubTask{ID: 10, TaskID: 1}
	finishTime := time.Now().UnixMilli()
	subInstanceRepo := &fakeSubInstanceRepo{records: []*datafillingdomain.DataFillingSubInstance{{ID: 100, TaskID: 1, PID: 10, UID: 9, FormID: 1, DataID: "row-1", Status: datafillingdomain.SubInstanceStatusFinished, FinishTime: finishTime}, {ID: 101, TaskID: 1, PID: 10, UID: 9, FormID: 1, Status: datafillingdomain.SubInstanceStatusOpen}}}
	svc := NewDataFillingService(formRepo, &fakeDataFillingDatasourceService{}, &fakeDDLProvider{}, nil, taskRepo, subTaskRepo, subInstanceRepo, nil)

	data, err := svc.GetUserTaskData(context.Background(), 9, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), data.FormID)
	assert.Equal(t, "form", data.FormTitle)
	assert.Equal(t, []string{"row-1"}, data.DataIDs)
	assert.Equal(t, "ext", data.FormExtSetting)
	assert.Equal(t, 2, data.FillType)
	require.Len(t, data.SubInstances, 2)
	assert.NotNil(t, data.SubInstances[0].FinishTime)

	_, err = svc.GetUserTaskData(context.Background(), 99, 10)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	_, err = svc.GetUserTaskData(context.Background(), 9, 99)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestDataFillingService_SaveUserTaskData_TransitionsOpenInstance(t *testing.T) {
	formRepo := newFakeDataFillingRepo()
	formRepo.records[1] = &datafillingdomain.DataFillingForm{ID: 1, Name: "form", NodeType: datafillingdomain.NodeTypeForm, PhysicalTableName: "df_demo", DatasourceID: 8, Forms: "[]"}
	taskRepo := newFakeTaskRepo()
	taskRepo.records[1] = &datafillingdomain.DataFillingTask{ID: 1, FormID: 1, FillType: 0}
	subTaskRepo := newFakeSubTaskRepo()
	subTaskRepo.records[10] = &datafillingdomain.DataFillingSubTask{ID: 10, TaskID: 1, UnfinishedCount: 1, UnfinishedUserCount: 1}
	subInstanceRepo := &fakeSubInstanceRepo{records: []*datafillingdomain.DataFillingSubInstance{{ID: 100, TaskID: 1, PID: 10, UID: 9, FormID: 1, Status: datafillingdomain.SubInstanceStatusOpen}}}
	ddl := &fakeDDLProvider{}
	svc := NewDataFillingService(formRepo, &fakeDataFillingDatasourceService{}, ddl, &fakeCommitLogRepo{}, taskRepo, subTaskRepo, subInstanceRepo, nil)
	svc.datasourceConnProvider = &fakeDatasourceConnProvider{}

	err := svc.SaveUserTaskData(context.Background(), 9, 10, []map[string]interface{}{{"name": "alice"}})
	require.NoError(t, err)
	assert.Len(t, ddl.insertedRows, 1)
	assert.Equal(t, []int64{100}, subInstanceRepo.statusUpdates)
	assert.Equal(t, datafillingdomain.SubInstanceStatusFinished, subInstanceRepo.records[0].Status)
	assert.Equal(t, []int64{10}, subTaskRepo.decremented)
}

func TestDataFillingService_SaveUserTaskData_ReSaveFinishedInstance(t *testing.T) {
	formRepo := newFakeDataFillingRepo()
	formRepo.records[1] = &datafillingdomain.DataFillingForm{ID: 1, Name: "form", NodeType: datafillingdomain.NodeTypeForm, PhysicalTableName: "df_demo", DatasourceID: 8, Forms: "[]"}
	taskRepo := newFakeTaskRepo()
	taskRepo.records[1] = &datafillingdomain.DataFillingTask{ID: 1, FormID: 1}
	subTaskRepo := newFakeSubTaskRepo()
	subTaskRepo.records[10] = &datafillingdomain.DataFillingSubTask{ID: 10, TaskID: 1, UnfinishedCount: 0, UnfinishedUserCount: 0}
	subInstanceRepo := &fakeSubInstanceRepo{records: []*datafillingdomain.DataFillingSubInstance{{ID: 100, TaskID: 1, PID: 10, UID: 9, FormID: 1, Status: datafillingdomain.SubInstanceStatusFinished, FinishTime: 123}}}
	ddl := &fakeDDLProvider{}
	svc := NewDataFillingService(formRepo, &fakeDataFillingDatasourceService{}, ddl, &fakeCommitLogRepo{}, taskRepo, subTaskRepo, subInstanceRepo, nil)
	svc.datasourceConnProvider = &fakeDatasourceConnProvider{}

	err := svc.SaveUserTaskData(context.Background(), 9, 10, []map[string]interface{}{{"id": "row-1", "name": "updated"}})
	require.NoError(t, err)
	assert.Len(t, ddl.updatedRows, 1)
	assert.Empty(t, subInstanceRepo.statusUpdates)
	assert.Empty(t, subTaskRepo.decremented)
}

func TestDataFillingService_SaveUserTaskData_RejectsWrongUID(t *testing.T) {
	svc := NewDataFillingService(newFakeDataFillingRepo(), &fakeDataFillingDatasourceService{}, &fakeDDLProvider{}, &fakeCommitLogRepo{}, newFakeTaskRepo(), newFakeSubTaskRepo(), &fakeSubInstanceRepo{}, nil)
	err := svc.SaveUserTaskData(context.Background(), 99, 10, []map[string]interface{}{{"name": "alice"}})
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestDataFillingService_AppendUserTaskData_AppendsAndTransitions(t *testing.T) {
	formRepo := newFakeDataFillingRepo()
	formRepo.records[1] = &datafillingdomain.DataFillingForm{ID: 1, Name: "form", NodeType: datafillingdomain.NodeTypeForm, PhysicalTableName: "df_demo", DatasourceID: 8, Forms: "[]"}
	taskRepo := newFakeTaskRepo()
	taskRepo.records[1] = &datafillingdomain.DataFillingTask{ID: 1, FormID: 1}
	subTaskRepo := newFakeSubTaskRepo()
	subTaskRepo.records[10] = &datafillingdomain.DataFillingSubTask{ID: 10, TaskID: 1, UnfinishedCount: 1, UnfinishedUserCount: 1}
	subInstanceRepo := &fakeSubInstanceRepo{records: []*datafillingdomain.DataFillingSubInstance{{ID: 100, TaskID: 1, PID: 10, UID: 9, FormID: 1, Status: datafillingdomain.SubInstanceStatusOpen}}}
	ddl := &fakeDDLProvider{}
	svc := NewDataFillingService(formRepo, &fakeDataFillingDatasourceService{}, ddl, &fakeCommitLogRepo{}, taskRepo, subTaskRepo, subInstanceRepo, nil)
	svc.datasourceConnProvider = &fakeDatasourceConnProvider{}

	err := svc.AppendUserTaskData(context.Background(), 9, 10, []map[string]interface{}{{"name": "alice"}})
	require.NoError(t, err)
	assert.Len(t, ddl.insertedRows, 1)
	assert.Equal(t, []int64{100}, subInstanceRepo.statusUpdates)
	assert.Equal(t, []int64{10}, subTaskRepo.decremented)
}

func TestDataFillingService_DeleteUserTaskData(t *testing.T) {
	formRepo := newFakeDataFillingRepo()
	formRepo.records[1] = &datafillingdomain.DataFillingForm{ID: 1, Name: "form", NodeType: datafillingdomain.NodeTypeForm, PhysicalTableName: "df_demo", DatasourceID: 8, Forms: "[]"}
	taskRepo := newFakeTaskRepo()
	taskRepo.records[1] = &datafillingdomain.DataFillingTask{ID: 1, FormID: 1}
	subTaskRepo := newFakeSubTaskRepo()
	subTaskRepo.records[10] = &datafillingdomain.DataFillingSubTask{ID: 10, TaskID: 1}
	subInstanceRepo := &fakeSubInstanceRepo{records: []*datafillingdomain.DataFillingSubInstance{{ID: 100, TaskID: 1, PID: 10, UID: 9, FormID: 1, Status: datafillingdomain.SubInstanceStatusFinished}}}
	ddl := &fakeDDLProvider{}
	svc := NewDataFillingService(formRepo, &fakeDataFillingDatasourceService{}, ddl, &fakeCommitLogRepo{}, taskRepo, subTaskRepo, subInstanceRepo, nil)
	svc.datasourceConnProvider = &fakeDatasourceConnProvider{}

	err := svc.DeleteUserTaskData(context.Background(), 9, 10, []string{"row-1", "row-2"})
	require.NoError(t, err)
	assert.Equal(t, [][]string{{"row-1"}, {"row-2"}}, ddl.deletedBatches)
	assert.Empty(t, subInstanceRepo.statusUpdates)

	err = svc.DeleteUserTaskData(context.Background(), 99, 10, []string{"row-1"})
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestDataFillingService_ExcelTemplateDownload(t *testing.T) {
	repo := newFakeDataFillingRepo()
	repo.records[1] = &datafillingdomain.DataFillingForm{
		ID:    1,
		Forms: `[{"id":"f1","settings":{"name":"姓名","mapping":{"columnName":"name","type":"nvarchar"}}},{"id":"f2","settings":{"name":"年龄","mapping":{"columnName":"age","type":"number"}}}]`,
	}
	svc := NewDataFillingService(repo, &fakeDataFillingDatasourceService{}, &fakeDDLProvider{}, nil, nil, nil, nil, nil)

	buf := bytes.NewBuffer(nil)
	err := svc.ExcelTemplateDownload(context.Background(), 1, buf)
	require.NoError(t, err)

	book, err := excelize.OpenReader(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	defer book.Close()
	rows, err := book.GetRows(book.GetSheetName(book.GetActiveSheetIndex()))
	require.NoError(t, err)
	require.NotEmpty(t, rows)
	assert.Equal(t, []string{"姓名", "年龄"}, rows[0])

	err = svc.ExcelTemplateDownload(context.Background(), 999, bytes.NewBuffer(nil))
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestDataFillingService_ExcelUpload(t *testing.T) {
	repo := newFakeDataFillingRepo()
	repo.records[1] = &datafillingdomain.DataFillingForm{
		ID:    1,
		Forms: `[{"id":"f1","settings":{"name":"姓名","mapping":{"columnName":"name","type":"nvarchar"}}},{"id":"f2","settings":{"name":"年龄","mapping":{"columnName":"age","type":"number"}}}]`,
	}
	svc := NewDataFillingService(repo, &fakeDataFillingDatasourceService{}, &fakeDDLProvider{}, nil, nil, nil, nil, nil)

	header := newExcelUploadFileHeader(t, "upload.xlsx", [][]string{{"姓名", "年龄"}, {"alice", "18"}, {"bob", "20"}})
	result, err := svc.ExcelUpload(context.Background(), 1, header)
	require.NoError(t, err)
	require.Len(t, result.DataList, 2)
	assert.Equal(t, "alice", result.DataList[0].Data["name"])
	assert.Equal(t, "18", result.DataList[0].Data["age"])
	assert.True(t, result.DataList[0].Insert)
	assert.NotEmpty(t, result.ID)

	emptyHeader := newExcelUploadFileHeader(t, "empty.xlsx", [][]string{{"姓名", "年龄"}})
	emptyResult, err := svc.ExcelUpload(context.Background(), 1, emptyHeader)
	require.NoError(t, err)
	assert.Empty(t, emptyResult.DataList)

	large := *header
	large.Size = maxDataFillingExcelUploadSize + 1
	_, err = svc.ExcelUpload(context.Background(), 1, &large)
	require.EqualError(t, err, "file size exceeds 10MB limit")
}

func TestDataFillingService_ConfirmUpload(t *testing.T) {
	formRepo := newFakeDataFillingRepo()
	formRepo.records[1] = &datafillingdomain.DataFillingForm{ID: 1, Name: "form", NodeType: datafillingdomain.NodeTypeForm, PhysicalTableName: "df_demo", DatasourceID: 8, Forms: "[]"}
	ddl := &fakeDDLProvider{}
	logs := &fakeCommitLogRepo{}
	svc := NewDataFillingService(formRepo, &fakeDataFillingDatasourceService{}, ddl, logs, nil, nil, nil, nil)
	svc.datasourceConnProvider = &fakeDatasourceConnProvider{}

	uploadID := svc.excelUploadSession.Store(&datafillingdomain.DfExcelData{DataList: []datafillingdomain.RowDataDatum{{Data: map[string]interface{}{"name": "alice"}, Insert: true}, {ID: "row-1", Data: map[string]interface{}{"id": "row-1", "name": "bob"}, Insert: false}}})
	err := svc.ConfirmUpload(context.Background(), 1, uploadID, 9, "tester")
	require.NoError(t, err)
	assert.Len(t, ddl.insertedRows, 1)
	assert.Len(t, ddl.updatedRows, 1)
	assert.Len(t, logs.created, 2)
	_, ok := svc.excelUploadSession.Load(uploadID)
	assert.False(t, ok)

	err = svc.ConfirmUpload(context.Background(), 1, "missing", 9, "tester")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestDataFillingService_UserTaskConfirmUpload(t *testing.T) {
	formRepo := newFakeDataFillingRepo()
	formRepo.records[1] = &datafillingdomain.DataFillingForm{ID: 1, Name: "form", NodeType: datafillingdomain.NodeTypeForm, PhysicalTableName: "df_demo", DatasourceID: 8, Forms: "[]"}
	taskRepo := newFakeTaskRepo()
	taskRepo.records[1] = &datafillingdomain.DataFillingTask{ID: 1, FormID: 1}
	subTaskRepo := newFakeSubTaskRepo()
	subTaskRepo.records[10] = &datafillingdomain.DataFillingSubTask{ID: 10, TaskID: 1, UnfinishedCount: 1, UnfinishedUserCount: 1}
	subInstanceRepo := &fakeSubInstanceRepo{records: []*datafillingdomain.DataFillingSubInstance{{ID: 100, TaskID: 1, PID: 10, UID: 9, FormID: 1, Status: datafillingdomain.SubInstanceStatusOpen}}}
	ddl := &fakeDDLProvider{}
	svc := NewDataFillingService(formRepo, &fakeDataFillingDatasourceService{}, ddl, &fakeCommitLogRepo{}, taskRepo, subTaskRepo, subInstanceRepo, nil)
	svc.datasourceConnProvider = &fakeDatasourceConnProvider{}
	uploadID := svc.excelUploadSession.Store(&datafillingdomain.DfExcelData{DataList: []datafillingdomain.RowDataDatum{{Data: map[string]interface{}{"name": "alice"}, Insert: true}}})

	err := svc.UserTaskConfirmUpload(context.Background(), 9, 10, 1, uploadID)
	require.NoError(t, err)
	assert.Len(t, ddl.insertedRows, 1)
	assert.Equal(t, []int64{100}, subInstanceRepo.statusUpdates)
	assert.Equal(t, []int64{10}, subTaskRepo.decremented)
	_, ok := svc.excelUploadSession.Load(uploadID)
	assert.False(t, ok)
}

func TestDataFillingService_ListDatasourceOptions(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&datasourceOptionRow{}))
	require.NoError(t, db.Create([]datasourceOptionRow{{Name: "beta", SortValue: 2}, {Name: "", SortValue: 1}, {Name: "alpha", SortValue: 0}}).Error)
	svc := NewDataFillingService(newFakeDataFillingRepo(), &fakeDataFillingDatasourceService{}, &fakeDDLProvider{}, nil, nil, nil, nil, nil)
	svc.datasourceConnProvider = &fakeDatasourceConnProvider{db: db}

	options, err := svc.ListDatasourceOptions(context.Background(), 8, &datafillingdomain.DatasourceOptionsRequest{OptionTable: "option_table", OptionColumn: "name", OptionOrder: "sort_value"})
	require.NoError(t, err)
	require.Len(t, options, 2)
	assert.Equal(t, "alpha", options[0].Value)
	assert.Equal(t, "beta", options[1].Value)
}

func TestDataFillingService_ExportFormData(t *testing.T) {
	formRepo := newFakeDataFillingRepo()
	formRepo.records[1] = &datafillingdomain.DataFillingForm{ID: 1, Name: "form", PhysicalTableName: "df_demo", DatasourceID: 8, Forms: `[{"settings":{"name":"姓名","mapping":{"columnName":"name","type":"nvarchar"}}},{"settings":{"name":"年龄","mapping":{"columnName":"age","type":"number"}}}]`}
	ddl := &fakeDDLProvider{countRows: 2, searchRows: []map[string]interface{}{{"name": "alice", "age": 18}, {"name": "bob", "age": "20"}}}
	svc := NewDataFillingService(formRepo, &fakeDataFillingDatasourceService{}, ddl, nil, nil, nil, nil, nil)
	svc.datasourceConnProvider = &fakeDatasourceConnProvider{}
	buf := bytes.NewBuffer(nil)

	err := svc.ExportFormData(context.Background(), 1, buf)
	require.NoError(t, err)
	book, err := excelize.OpenReader(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	defer book.Close()
	rows, err := book.GetRows(book.GetSheetName(book.GetActiveSheetIndex()))
	require.NoError(t, err)
	require.Len(t, rows, 3)
	assert.Equal(t, []string{"姓名", "年龄"}, rows[0])
	assert.Equal(t, []string{"alice", "18"}, rows[1])
	assert.Equal(t, []string{"bob", "20"}, rows[2])
}

func TestDataFillingService_ExtraDetails(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE user_extra (name TEXT, city TEXT, dept TEXT)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO user_extra(name, city, dept) VALUES (?, ?, ?), (?, ?, ?)`, "alice", "shanghai", "sales", "bob", "beijing", "finance").Error)
	svc := NewDataFillingService(newFakeDataFillingRepo(), &fakeDataFillingDatasourceService{}, &fakeDDLProvider{}, nil, nil, nil, nil, nil)
	svc.SetDatasourceConnectionProvider(&fakeDatasourceConnProvider{db: db})

	details, err := svc.ExtraDetails(context.Background(), &datafillingdomain.ExtraDetailsRequest{OptionDatasource: "8", OptionTable: "user_extra", OptionColumn: "name", ExtraColumns: []string{"city", "dept"}, Value: "alice"})
	require.NoError(t, err)
	require.Len(t, details, 2)
	assert.Equal(t, &datafillingdomain.ExtraDetails{Name: "city", Value: "shanghai"}, details[0])
	assert.Equal(t, &datafillingdomain.ExtraDetails{Name: "dept", Value: "sales"}, details[1])
}

func TestDataFillingService_ExtraDetails_ValidateIdentifiers(t *testing.T) {
	svc := NewDataFillingService(newFakeDataFillingRepo(), &fakeDataFillingDatasourceService{}, &fakeDDLProvider{}, nil, nil, nil, nil, nil)
	_, err := svc.ExtraDetails(context.Background(), &datafillingdomain.ExtraDetailsRequest{OptionDatasource: "1", OptionTable: "bad-table", OptionColumn: "name", ExtraColumns: []string{"city"}, Value: "alice"})
	require.ErrorIs(t, err, gorm.ErrInvalidData)
}

func TestValidateExtraDetailsRequest(t *testing.T) {
	datasourceID, tableName, optionColumn, extraColumns, value, err := validateExtraDetailsRequest(&datafillingdomain.ExtraDetailsRequest{OptionDatasource: " 8 ", OptionTable: "user_table", OptionColumn: "user_name", ExtraColumns: []string{" city ", "age"}, Value: " alice "})
	require.NoError(t, err)
	assert.Equal(t, int64(8), datasourceID)
	assert.Equal(t, "user_table", tableName)
	assert.Equal(t, "user_name", optionColumn)
	assert.Equal(t, []string{"city", "age"}, extraColumns)
	assert.Equal(t, "alice", value)

	_, _, _, _, _, err = validateExtraDetailsRequest(&datafillingdomain.ExtraDetailsRequest{OptionDatasource: "1", OptionTable: "users;drop", OptionColumn: "name", ExtraColumns: []string{"city"}})
	require.ErrorIs(t, err, gorm.ErrInvalidData)
}

func TestBuildTaskCronSpec(t *testing.T) {
	tests := []struct {
		name string
		task *datafillingdomain.DataFillingTask
		want string
	}{
		{name: "daily", task: &datafillingdomain.DataFillingTask{RateType: 1, RateVal: "09:30:15"}, want: "15 30 9 * * *"},
		{name: "weekly", task: &datafillingdomain.DataFillingTask{RateType: 2, RateVal: "7 09:30:15"}, want: "15 30 9 * * 0"},
		{name: "monthly", task: &datafillingdomain.DataFillingTask{RateType: 3, RateVal: "15 09:30:15"}, want: "15 30 9 15 * *"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := buildTaskCronSpec(tt.task)
			require.NoError(t, err)
			assert.Equal(t, tt.want, spec)
		})
	}
}

func TestParseJSONInt64List(t *testing.T) {
	values, err := parseJSONInt64List(`[1,2,3]`)
	require.NoError(t, err)
	assert.Equal(t, []int64{1, 2, 3}, values)

	values, err = parseJSONInt64List("")
	require.NoError(t, err)
	assert.Empty(t, values)

	_, err = parseJSONInt64List(`not-json`)
	require.EqualError(t, err, "parse int64 list failed")
}

func TestMapSubTaskUserStatusFilter(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *int
		wantErr error
	}{
		{name: "all empty", input: "", want: nil},
		{name: "all literal", input: "all", want: nil},
		{name: "unfinished", input: "unfinished", want: intPtr(datafillingdomain.SubInstanceStatusOpen)},
		{name: "open alias", input: "open", want: intPtr(datafillingdomain.SubInstanceStatusOpen)},
		{name: "todo alias", input: "todo", want: intPtr(datafillingdomain.SubInstanceStatusOpen)},
		{name: "finished", input: "finished", want: intPtr(datafillingdomain.SubInstanceStatusFinished)},
		{name: "done alias", input: "done", want: intPtr(datafillingdomain.SubInstanceStatusFinished)},
		{name: "invalid", input: "other", wantErr: gorm.ErrInvalidData},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mapSubTaskUserStatusFilter(tt.input)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			if tt.want == nil {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.Equal(t, *tt.want, *got)
		})
	}
}

func TestNormalizeWeekday(t *testing.T) {
	weekday, err := normalizeWeekday(0)
	require.NoError(t, err)
	assert.Equal(t, 0, weekday)

	weekday, err = normalizeWeekday(6)
	require.NoError(t, err)
	assert.Equal(t, 6, weekday)

	weekday, err = normalizeWeekday(7)
	require.NoError(t, err)
	assert.Equal(t, 0, weekday)

	_, err = normalizeWeekday(8)
	require.EqualError(t, err, "invalid weekday")
}

func TestComputeSubTaskStatus(t *testing.T) {
	now := time.Now()
	assert.Equal(t, datafillingdomain.SubTaskStatusActive, computeSubTaskStatus(&datafillingdomain.DataFillingTask{EndTime: now.Add(time.Minute).UnixMilli()}, now))
	assert.Equal(t, datafillingdomain.SubTaskStatusExpired, computeSubTaskStatus(&datafillingdomain.DataFillingTask{EndTime: now.Add(-time.Minute).UnixMilli()}, now))
}

func TestDataFillingService_GetDatasourceConnection(t *testing.T) {
	t.Run("returns provider connection", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		svc := NewDataFillingService(newFakeDataFillingRepo(), &fakeDataFillingDatasourceService{}, &fakeDDLProvider{}, nil, nil, nil, nil, nil)
		svc.SetDatasourceConnectionProvider(&fakeDatasourceConnProvider{db: db})

		got, err := svc.GetDatasourceConnection(context.Background(), 1)
		require.NoError(t, err)
		assert.Same(t, db, got)
	})

	t.Run("validates datasource metadata", func(t *testing.T) {
		conf := `{"host":"127.0.0.1","port":3306,"dataBase":"demo","username":"u","password":"p"}`
		svc := NewDataFillingService(newFakeDataFillingRepo(), &fakeDataFillingDatasourceService{ds: map[int64]*datasourcedomain.CoreDatasource{1: {ID: 1, Type: "mysql", Configuration: &conf}, 2: {ID: 2, Type: "pgsql"}, 3: {ID: 3, Type: "mysql"}, 4: {ID: 4, Type: "mysql", Configuration: localStringPtr("bad-json")}}}, &fakeDDLProvider{}, nil, nil, nil, nil, nil)

		_, err := svc.GetDatasourceConnection(context.Background(), 1)
		require.Error(t, err)

		_, err = svc.GetDatasourceConnection(context.Background(), 2)
		require.EqualError(t, err, "unsupported datasource type: pgsql")

		_, err = svc.GetDatasourceConnection(context.Background(), 3)
		require.EqualError(t, err, "datasource configuration is empty")

		_, err = svc.GetDatasourceConnection(context.Background(), 4)
		require.Error(t, err)
	})
}

func TestDataFillingService_ResolveLevelAndRecalculateChildrenLevels(t *testing.T) {
	repo := newFakeDataFillingRepo()
	repo.records[1] = &datafillingdomain.DataFillingForm{ID: 1, Level: 0}
	repo.records[2] = &datafillingdomain.DataFillingForm{ID: 2, PID: 1, Level: 1}
	repo.records[3] = &datafillingdomain.DataFillingForm{ID: 3, PID: 2, Level: 2}
	svc := NewDataFillingService(repo, &fakeDataFillingDatasourceService{}, &fakeDDLProvider{}, nil, nil, nil, nil, nil)

	level, err := svc.resolveLevel(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 1, level)

	err = svc.recalculateChildrenLevels(context.Background(), 1, 3)
	require.NoError(t, err)
	assert.Equal(t, 4, repo.records[2].Level)
	assert.Equal(t, 5, repo.records[3].Level)
}

func TestDataFillingScheduler_StartStopAndLoadStartedTasks(t *testing.T) {
	formRepo := newFakeDataFillingRepo()
	formRepo.records[1] = &datafillingdomain.DataFillingForm{ID: 1, Name: "form"}
	taskRepo := newFakeTaskRepo()
	taskRepo.records[1] = &datafillingdomain.DataFillingTask{ID: 1, FormID: 1, UIDList: `[9,10]`, RIDList: `[]`, RateType: 1, RateVal: "09:30:00", Status: datafillingdomain.TaskStatusStarted, NextExecTime: time.Now().Add(-time.Minute).UnixMilli()}
	subTaskRepo := newFakeSubTaskRepo()
	subInstanceRepo := &fakeSubInstanceRepo{}
	scheduler := NewDataFillingScheduler(taskRepo, subTaskRepo, subInstanceRepo, formRepo)

	scheduler.Start(context.Background())
	defer scheduler.Stop()

	assert.Len(t, scheduler.entries, 1)
	assert.Len(t, subTaskRepo.records, 1)
	assert.Len(t, subInstanceRepo.records, 2)
	updated, err := taskRepo.GetTaskByID(context.Background(), 1)
	require.NoError(t, err)
	assert.NotZero(t, updated.LastExecTime)
	assert.NotZero(t, updated.NextExecTime)
}

func TestDataFillingScheduler_Stop_NoPanicOnNil(t *testing.T) {
	var scheduler *DataFillingScheduler
	assert.NotPanics(t, func() { scheduler.Stop() })
}

func TestDataFillingService_GetTemplateByUserTaskItem(t *testing.T) {
	formRepo := newFakeDataFillingRepo()
	formRepo.records[1] = &datafillingdomain.DataFillingForm{ID: 1, Forms: `[{"settings":{"name":"姓名","mapping":{"columnName":"name","type":"nvarchar"}}}]`}
	taskRepo := newFakeTaskRepo()
	taskRepo.records[1] = &datafillingdomain.DataFillingTask{ID: 1, FormID: 1}
	subInstanceRepo := &fakeSubInstanceRepo{records: []*datafillingdomain.DataFillingSubInstance{{ID: 100, TaskID: 1, FormID: 1}}}
	svc := NewDataFillingService(formRepo, &fakeDataFillingDatasourceService{}, &fakeDDLProvider{}, nil, taskRepo, newFakeSubTaskRepo(), subInstanceRepo, nil)

	template, err := svc.GetTemplateByUserTaskItem(context.Background(), 100)
	require.NoError(t, err)
	assert.Contains(t, template, "姓名")

	_, err = svc.GetTemplateByUserTaskItem(context.Background(), 999)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func intPtr(v int) *int { return &v }

func localStringPtr(v string) *string { return &v }

func newExcelUploadFileHeader(t *testing.T, filename string, rows [][]string) *multipart.FileHeader {
	t.Helper()
	file := excelize.NewFile()
	defer file.Close()
	sheet := file.GetSheetName(file.GetActiveSheetIndex())
	for rowIndex, row := range rows {
		for colIndex, value := range row {
			cell, err := excelize.CoordinatesToCellName(colIndex+1, rowIndex+1)
			require.NoError(t, err)
			require.NoError(t, file.SetCellValue(sheet, cell, value))
		}
	}
	buf, err := file.WriteToBuffer()
	require.NoError(t, err)
	body := bytes.NewBuffer(nil)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	require.NoError(t, err)
	_, err = part.Write(buf.Bytes())
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	require.NoError(t, req.ParseMultipartForm(int64(body.Len())+1024))
	_, header, err := req.FormFile("file")
	require.NoError(t, err)
	return header
}

func TestDataFillingScheduler_ComputeNextExecTime(t *testing.T) {
	scheduler := NewDataFillingScheduler(newFakeTaskRepo(), newFakeSubTaskRepo(), &fakeSubInstanceRepo{}, newFakeDataFillingRepo())

	cronNext, err := scheduler.computeNextExecTime(&datafillingdomain.DataFillingTask{RateType: 0, RateVal: "0 0 10 * * *"})
	require.NoError(t, err)
	assert.NotZero(t, cronNext)

	dailyNext, err := computeTaskNextExecTime(&datafillingdomain.DataFillingTask{RateType: 1, RateVal: "23:30:00"}, time.Date(2026, 5, 4, 10, 0, 0, 0, time.Local))
	require.NoError(t, err)
	assert.Equal(t, 23, dailyNext.Hour())

	weeklyNext, err := computeTaskNextExecTime(&datafillingdomain.DataFillingTask{RateType: 2, RateVal: "1 08:00:00"}, time.Date(2026, 5, 4, 10, 0, 0, 0, time.Local))
	require.NoError(t, err)
	assert.Equal(t, time.Monday, weeklyNext.Weekday())

	monthlyNext, err := computeTaskNextExecTime(&datafillingdomain.DataFillingTask{RateType: 3, RateVal: "15 08:00:00"}, time.Date(2026, 5, 4, 10, 0, 0, 0, time.Local))
	require.NoError(t, err)
	assert.Equal(t, 15, monthlyNext.Day())
}
