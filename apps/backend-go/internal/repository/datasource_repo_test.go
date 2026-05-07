package repository

import (
	"testing"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/domain/datasource"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupDatasourceRepositoryTest(t *testing.T) (*DatasourceRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&datasource.CoreDatasource{},
		&datasourceTable{},
		&auto.CoreDatasourceTaskLog{},
		&auto.CoreDsFinishPage{},
	))

	return NewDatasourceRepository(db), db
}

func TestDatasourceRepository_GuardsAndConstructor(t *testing.T) {
	repo := NewDatasourceRepository(&gorm.DB{})
	require.NotNil(t, repo)

	var nilRepo *DatasourceRepository
	req := &datasource.ListRequest{Current: 1, Size: 10}
	keyword := "alpha"
	excludeID := int64(1)

	list, total, err := nilRepo.Query(req)
	require.Error(t, err)
	assert.Nil(t, list)
	assert.Zero(t, total)
	assert.Equal(t, "datasource repository is unavailable", err.Error())

	ds, err := nilRepo.GetByID(1)
	require.Error(t, err)
	assert.Nil(t, ds)

	nearest, err := nilRepo.FindNearestIDInWindow(1, 10)
	require.Error(t, err)
	assert.Nil(t, nearest)

	err = nilRepo.Create(&datasource.CoreDatasource{Name: "x", Type: datasource.TypeFolder})
	require.Error(t, err)

	err = nilRepo.Update(&datasource.CoreDatasource{ID: 1, Name: "x", Type: datasource.TypeFolder})
	require.Error(t, err)

	err = nilRepo.SoftDelete(1)
	require.Error(t, err)

	children, err := nilRepo.ListChildren(1)
	require.Error(t, err)
	assert.Nil(t, children)

	count, err := nilRepo.CountByNameAndPID("x", 0, &excludeID)
	require.Error(t, err)
	assert.Zero(t, count)

	all, err := nilRepo.ListAll(&keyword)
	require.Error(t, err)
	assert.Nil(t, all)
}

func TestDatasourceRepository_CRUDAndQueries(t *testing.T) {
	t.Run("supports create update delete and common lookups", func(t *testing.T) {
		repo, db := setupDatasourceRepositoryTest(t)
		createBy := "alice"
		createTime1 := int64(100)
		createTime2 := int64(200)
		createTime3 := int64(300)

		root := &datasource.CoreDatasource{
			Name:       "Alpha Root",
			Type:       datasource.TypeFolder,
			CreateBy:   &createBy,
			CreateTime: &createTime1,
		}
		require.NoError(t, repo.Create(root))
		require.Positive(t, root.ID)

		child := &datasource.CoreDatasource{
			Name:       "Alpha MySQL",
			PID:        &root.ID,
			Type:       "mysql",
			CreateBy:   &createBy,
			CreateTime: &createTime2,
		}
		require.NoError(t, repo.Create(child))

		other := &datasource.CoreDatasource{
			Name:       "Beta Oracle",
			Type:       "oracle",
			CreateBy:   &createBy,
			CreateTime: &createTime3,
		}
		require.NoError(t, repo.Create(other))

		deletedFlag := 1
		deleted := &datasource.CoreDatasource{Name: "Deleted Alpha", Type: "mysql", DelFlag: &deletedFlag}
		require.NoError(t, db.Create(deleted).Error)

		keyword := "Alpha"
		list, total, err := repo.Query(&datasource.ListRequest{Keyword: &keyword, Current: 0, Size: 1000})
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		require.Len(t, list, 2)
		assert.Equal(t, child.ID, list[0].ID)
		assert.Equal(t, root.ID, list[1].ID)

		paged, total, err := repo.Query(&datasource.ListRequest{Current: 2, Size: 1})
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		require.Len(t, paged, 1)

		found, err := repo.GetByID(child.ID)
		require.NoError(t, err)
		assert.Equal(t, "Alpha MySQL", found.Name)

		_, err = repo.GetByID(99999)
		require.Error(t, err)

		nearest, err := repo.FindNearestIDInWindow(child.ID+1, 0)
		require.NoError(t, err)
		require.NotNil(t, nearest)
		assert.Equal(t, other.ID, *nearest)

		nearest, err = repo.FindNearestIDInWindow(99999, 10)
		require.NoError(t, err)
		assert.Nil(t, nearest)

		child.Name = "Alpha MySQL Updated"
		require.NoError(t, repo.Update(child))
		updated, err := repo.GetByID(child.ID)
		require.NoError(t, err)
		assert.Equal(t, "Alpha MySQL Updated", updated.Name)

		children, err := repo.ListChildren(root.ID)
		require.NoError(t, err)
		require.Len(t, children, 1)
		assert.Equal(t, child.ID, children[0].ID)

		nameCount, err := repo.CountByNameAndPID("Alpha MySQL Updated", root.ID, nil)
		require.NoError(t, err)
		assert.Equal(t, int64(1), nameCount)

		nameCount, err = repo.CountByNameAndPID("Alpha MySQL Updated", root.ID, &child.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), nameCount)

		all, err := repo.ListAll(&keyword)
		require.NoError(t, err)
		require.Len(t, all, 2)

		batch, err := repo.ListBatch(&keyword, root.ID, 10)
		require.NoError(t, err)
		require.Len(t, batch, 1)
		assert.Equal(t, child.ID, batch[0].ID)

		batch, err = repo.ListBatch(nil, 0, 2)
		require.NoError(t, err)
		require.Len(t, batch, 2)
		assert.True(t, batch[0].ID < batch[1].ID)

		byType, err := repo.ListByType("mysql", nil)
		require.NoError(t, err)
		require.Len(t, byType, 1)
		assert.Equal(t, child.ID, byType[0].ID)

		byType, err = repo.ListByType("mysql", &child.ID)
		require.NoError(t, err)
		assert.Empty(t, byType)

		require.NoError(t, repo.SoftDelete(child.ID))
		_, err = repo.GetByID(child.ID)
		require.Error(t, err)
	})

	t.Run("returns empty results for unmatched filters", func(t *testing.T) {
		repo, _ := setupDatasourceRepositoryTest(t)
		keyword := "missing"

		list, total, err := repo.Query(&datasource.ListRequest{Keyword: &keyword, Current: 1, Size: 10})
		require.NoError(t, err)
		assert.Empty(t, list)
		assert.Zero(t, total)

		all, err := repo.ListAll(&keyword)
		require.NoError(t, err)
		assert.Empty(t, all)
	})
}

func TestDatasourceRepository_ListLatestTypesByCreator(t *testing.T) {
	repo, _ := setupDatasourceRepositoryTest(t)
	createBy := "alice"
	createTime1 := int64(100)
	createTime2 := int64(200)

	require.NoError(t, repo.Create(&datasource.CoreDatasource{Name: "Folder", Type: datasource.TypeFolder, CreateBy: &createBy, CreateTime: &createTime1}))
	require.NoError(t, repo.Create(&datasource.CoreDatasource{Name: "Mysql", Type: "mysql", CreateBy: &createBy, CreateTime: &createTime2}))

	assert.Panics(t, func() {
		_, _ = repo.ListLatestTypesByCreator(createBy, 50)
	})
}

func TestDatasourceRepository_TableMetadataAndRows(t *testing.T) {
	t.Run("lists datasource tables and counts relations", func(t *testing.T) {
		repo, db := setupDatasourceRepositoryTest(t)
		typeVal := "db"
		require.NoError(t, db.Create(&datasourceTable{ID: 11, Name: "Orders", PhysicalName: "orders", DatasourceID: 7, Type: &typeVal}).Error)
		require.NoError(t, db.Create(&datasourceTable{ID: 12, Name: "Users", PhysicalName: "users", DatasourceID: 7}).Error)
		require.NoError(t, db.Create(&datasourceTable{ID: 13, Name: "Other", PhysicalName: "other", DatasourceID: 8}).Error)

		tables, err := repo.ListTables(7)
		require.NoError(t, err)
		require.Len(t, tables, 2)
		assert.Equal(t, int64(12), tables[0].ID)
		assert.Equal(t, "", tables[0].Type)
		assert.Equal(t, "db", tables[1].Type)

		count, err := repo.CountDatasourceRelations(7)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)

		count, err = repo.CountDatasourceRelations(999)
		require.NoError(t, err)
		assert.Zero(t, count)
	})

	t.Run("reads schemas from attached information_schema", func(t *testing.T) {
		repo, db := setupDatasourceRepositoryTest(t)
		require.NoError(t, db.Exec("ATTACH DATABASE ':memory:' AS information_schema").Error)
		require.NoError(t, db.Exec("CREATE TABLE information_schema.schemata (schema_name TEXT)").Error)
		require.NoError(t, db.Exec("INSERT INTO information_schema.schemata (schema_name) VALUES ('analytics'), (''), ('warehouse')").Error)

		schemas, err := repo.ListSchemas()
		require.NoError(t, err)
		assert.Equal(t, []string{"analytics", "warehouse"}, schemas)
	})

	t.Run("returns schema query error when metadata table is absent", func(t *testing.T) {
		repo, _ := setupDatasourceRepositoryTest(t)

		schemas, err := repo.ListSchemas()
		require.Error(t, err)
		assert.Nil(t, schemas)
	})

	t.Run("validates table names and previews sqlite rows", func(t *testing.T) {
		repo, db := setupDatasourceRepositoryTest(t)
		require.NoError(t, db.Exec("CREATE TABLE preview_rows (id INTEGER PRIMARY KEY, city TEXT, amount REAL)").Error)
		require.NoError(t, db.Exec("INSERT INTO preview_rows (id, city, amount) VALUES (1, 'Shanghai', 10.5), (2, 'Beijing', 20.0), (3, 'Shenzhen', 30.0)").Error)

		rows, err := repo.PreviewRows("preview_rows", 0)
		require.NoError(t, err)
		require.Len(t, rows, 3)

		rows, err = repo.PreviewRows("preview_rows", 2)
		require.NoError(t, err)
		require.Len(t, rows, 2)

		count, err := repo.CountRows("preview_rows")
		require.NoError(t, err)
		assert.Equal(t, int64(3), count)

		fields, err := repo.ListTableFields("bad-name;")
		require.Error(t, err)
		assert.Nil(t, fields)
		assert.Equal(t, "invalid table name", err.Error())

		rows, err = repo.PreviewRows("bad-name;", 10)
		require.Error(t, err)
		assert.Nil(t, rows)

		count, err = repo.CountRows("bad-name;")
		require.Error(t, err)
		assert.Zero(t, count)
	})
}

func TestDatasourceRepository_FinishPageAndSyncLogs(t *testing.T) {
	t.Run("checks finish page record existence and sqlite insert behavior", func(t *testing.T) {
		repo, db := setupDatasourceRepositoryTest(t)

		exists, err := repo.ExistsFinishPageRecord(9)
		require.NoError(t, err)
		assert.False(t, exists)

		require.NoError(t, db.Create(&auto.CoreDsFinishPage{ID: 9}).Error)
		exists, err = repo.ExistsFinishPageRecord(9)
		require.NoError(t, err)
		assert.True(t, exists)

		err = repo.CreateFinishPageRecord(10)
		require.Error(t, err)
	})

	t.Run("creates and lists sync task logs", func(t *testing.T) {
		repo, db := setupDatasourceRepositoryTest(t)

		err := repo.CreateSyncTaskLog(nil)
		require.Error(t, err)
		assert.Equal(t, "sync record is required", err.Error())

		require.NoError(t, db.Create(&auto.CoreDatasourceTaskLog{
			DsID:              1,
			TaskID:            11,
			StartTime:         100,
			EndTime:           110,
			TaskStatus:        datasource.StatusSuccess,
			PhysicalTableName: "seed_table",
			Info:              "seed",
			CreateTime:        90,
			TriggerType:       "manual",
		}).Error)

		record := &datasource.SyncRecord{
			DsID:        1,
			TaskID:      12,
			StartTime:   200,
			EndTime:     210,
			TaskStatus:  datasource.StatusError,
			TableName:   "new_table",
			Info:        "failed",
			CreateTime:  190,
			TriggerType: "schedule",
		}
		require.NoError(t, repo.CreateSyncTaskLog(record))
		require.Positive(t, record.ID)

		logs, total, err := repo.ListSyncTaskLogs(1, 0, 500)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		require.Len(t, logs, 2)
		assert.Equal(t, int64(12), logs[0].TaskID)
		assert.Equal(t, "new_table", logs[0].Name)
		assert.Equal(t, int64(11), logs[1].TaskID)

		logs, total, err = repo.ListSyncTaskLogs(1, 2, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		require.Len(t, logs, 1)
		assert.Equal(t, int64(11), logs[0].TaskID)

		logs, total, err = repo.ListSyncTaskLogs(999, 1, 10)
		require.NoError(t, err)
		assert.Empty(t, logs)
		assert.Zero(t, total)
	})
}

func TestDatasourceRepository_InferDeType(t *testing.T) {
	assert.Equal(t, 2, inferDeType("decimal(10,2)"))
	assert.Equal(t, 1, inferDeType("datetime"))
	assert.Equal(t, 0, inferDeType("varchar(255)"))
}
