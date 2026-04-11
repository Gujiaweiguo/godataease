package repository

import (
	"testing"

	"dataease/backend/internal/domain/dataset"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupDatasetRepositoryTest(t *testing.T) (*DatasetRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&dataset.CoreDatasetGroup{}, &dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{}))
	require.NoError(t, db.Exec("CREATE TABLE preview_rows (id INTEGER PRIMARY KEY, category TEXT, city TEXT)").Error)
	require.NoError(t, db.Exec("INSERT INTO preview_rows (id, category, city) VALUES (1, 'A', 'Shanghai'), (2, 'B', 'Beijing'), (3, NULL, 'Shenzhen')").Error)
	require.NoError(t, db.Exec("CREATE TABLE core_chart_view (id INTEGER PRIMARY KEY, table_id INTEGER)").Error)

	return NewDatasetRepository(db), db
}

func intPtrDatasetRepo(v int) *int       { return &v }
func strPtrDatasetRepo(v string) *string { return &v }

func TestDatasetRepository_ListGroupsAndBatch(t *testing.T) {
	t.Run("guards unavailable repository", func(t *testing.T) {
		var repo *DatasetRepository

		groups, err := repo.ListGroups(nil)
		require.Error(t, err)
		assert.Nil(t, groups)
		assert.Equal(t, "dataset repository is unavailable", err.Error())

		groups, err = repo.ListGroupsBatch(nil, 0, 10)
		require.Error(t, err)
		assert.Nil(t, groups)
		assert.Equal(t, "dataset repository is unavailable", err.Error())
	})

	t.Run("filters keyword after id and limit", func(t *testing.T) {
		repo, db := setupDatasetRepositoryTest(t)
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 1, Name: "Alpha Folder", Level: intPtrDatasetRepo(1)}).Error)
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 2, Name: "Beta Folder", Level: intPtrDatasetRepo(2)}).Error)
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 3, Name: "Alpha Dataset", Level: intPtrDatasetRepo(3)}).Error)

		keyword := "Alpha"
		groups, err := repo.ListGroups(&keyword)
		require.NoError(t, err)
		require.Len(t, groups, 2)
		assert.Equal(t, int64(1), groups[0].ID)
		assert.Equal(t, int64(3), groups[1].ID)

		groups, err = repo.ListGroupsBatch(&keyword, 1, 1)
		require.NoError(t, err)
		require.Len(t, groups, 1)
		assert.Equal(t, int64(3), groups[0].ID)
	})
}

func TestDatasetRepository_PreviewSQL(t *testing.T) {
	repo, _ := setupDatasetRepositoryTest(t)

	rows, err := repo.PreviewSQL("SELECT id FROM preview_rows ORDER BY id ASC", 0)
	require.NoError(t, err)
	require.Len(t, rows, 3)

	rows, err = repo.PreviewSQL("SELECT id FROM preview_rows ORDER BY id ASC", 1)
	require.NoError(t, err)
	require.Len(t, rows, 1)

	rows, err = repo.PreviewSQL("SELECT id FROM preview_rows ORDER BY id ASC", 999)
	require.NoError(t, err)
	require.Len(t, rows, 3)
}

func TestDatasetRepository_FindPrimaryTableName(t *testing.T) {
	t.Run("returns first valid table name", func(t *testing.T) {
		repo, db := setupDatasetRepositoryTest(t)
		require.NoError(t, db.Create(&dataset.CoreDatasetTable{ID: 10, DatasetGroupID: 7, PhysicalTable: strPtrDatasetRepo("preview_rows")}).Error)

		tableName, err := repo.FindPrimaryTableName(7)
		require.NoError(t, err)
		assert.Equal(t, "preview_rows", tableName)
	})

	t.Run("rejects empty and invalid physical table name", func(t *testing.T) {
		repo, db := setupDatasetRepositoryTest(t)
		require.NoError(t, db.Create(&dataset.CoreDatasetTable{ID: 11, DatasetGroupID: 8}).Error)

		tableName, err := repo.FindPrimaryTableName(8)
		require.Error(t, err)
		assert.Empty(t, tableName)
		assert.Equal(t, "dataset table_name is empty", err.Error())

		require.NoError(t, db.Create(&dataset.CoreDatasetTable{ID: 12, DatasetGroupID: 9, PhysicalTable: strPtrDatasetRepo("bad-name;")}).Error)
		tableName, err = repo.FindPrimaryTableName(9)
		require.Error(t, err)
		assert.Empty(t, tableName)
		assert.Equal(t, "invalid dataset table name", err.Error())
	})
}

func TestDatasetRepository_PreviewRowsWithFilter(t *testing.T) {
	repo, _ := setupDatasetRepositoryTest(t)

	rows, err := repo.PreviewRowsWithFilter("preview_rows", "", "", nil, 0)
	require.NoError(t, err)
	require.Len(t, rows, 3)

	rows, err = repo.PreviewRowsWithFilter("preview_rows", "city", "category = ?", []any{"A"}, 10)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "Shanghai", rows[0]["city"])

	rows, err = repo.PreviewRowsWithFilter("preview_rows", "city", "", nil, 1)
	require.NoError(t, err)
	require.Len(t, rows, 1)
}

func TestDatasetRepository_QueryDistinctValues(t *testing.T) {
	repo, _ := setupDatasetRepositoryTest(t)

	values, err := repo.QueryDistinctValues("preview_rows", "city", nil, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{"Beijing", "Shanghai", "Shenzhen"}, values)

	values, err = repo.QueryDistinctValues("preview_rows", "city", []dataset.EnumFilterClause{{Column: "category", Values: []string{"A", "B"}}}, 1)
	require.NoError(t, err)
	assert.Len(t, values, 1)

	values, err = repo.QueryDistinctValues("preview_rows", "city", []dataset.EnumFilterClause{{Column: " ", Values: []string{"A"}}, {Column: "category", Values: []string{"B"}}}, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{"Beijing"}, values)
}

func TestDatasetRepository_GroupCRUDAndLookups(t *testing.T) {
	repo, db := setupDatasetRepositoryTest(t)
	rootLevel := 1
	rootType := "folder"
	childLevel := 2
	childType := "dataset"

	group := &dataset.CoreDatasetGroup{Name: "Root Group", Level: &rootLevel, NodeType: &rootType}
	require.NoError(t, repo.CreateGroup(group))
	require.Positive(t, group.ID)

	found, err := repo.GetGroupByID(group.ID)
	require.NoError(t, err)
	assert.Equal(t, "Root Group", found.Name)

	group.Name = "Updated Root Group"
	require.NoError(t, repo.UpdateGroup(group))
	found, err = repo.GetGroupByID(group.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Root Group", found.Name)

	child := &dataset.CoreDatasetGroup{Name: "Child Group", PID: &group.ID, Level: &childLevel, NodeType: &childType}
	require.NoError(t, repo.CreateGroup(child))
	deletedFlag := 1
	deleted := &dataset.CoreDatasetGroup{Name: "Deleted Child", PID: &group.ID, DelFlag: &deletedFlag}
	require.NoError(t, db.Create(deleted).Error)

	count, err := repo.CountGroupByNameAndPID("Updated Root Group", 0, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	count, err = repo.CountGroupByNameAndPID("Updated Root Group", 0, &group.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	children, err := repo.ListGroupChildren(group.ID)
	require.NoError(t, err)
	require.Len(t, children, 1)
	assert.Equal(t, child.ID, children[0].ID)

	require.NoError(t, repo.SoftDeleteGroup(child.ID))
	_, err = repo.GetGroupByID(child.ID)
	require.Error(t, err)
}

func TestDatasetRepository_TableAndFieldLookups(t *testing.T) {
	repo, db := setupDatasetRepositoryTest(t)
	tableName := "preview_rows"
	name := "Preview Table"
	table := &dataset.CoreDatasetTable{ID: 41, Name: &name, DatasetGroupID: 9, PhysicalTable: &tableName}
	require.NoError(t, db.Create(table).Error)
	otherName := "Other Table"
	otherTable := &dataset.CoreDatasetTable{ID: 42, Name: &otherName, DatasetGroupID: 10, PhysicalTable: strPtrDatasetRepo("other_preview")}
	require.NoError(t, db.Create(otherTable).Error)
	origin := "city"
	alias := "City"
	fieldType := "varchar"
	field := &dataset.CoreDatasetTableField{ID: 51, DatasetGroupID: 9, OriginName: &origin, Name: &alias, Type: &fieldType}
	require.NoError(t, db.Create(field).Error)

	tables, err := repo.ListTablesByDatasetGroupID(9)
	require.NoError(t, err)
	require.Len(t, tables, 1)
	assert.Equal(t, int64(41), tables[0].ID)

	foundTable, err := repo.GetTableByID(41)
	require.NoError(t, err)
	require.NotNil(t, foundTable.PhysicalTable)
	assert.Equal(t, "preview_rows", *foundTable.PhysicalTable)

	fields, err := repo.ListFields(9)
	require.NoError(t, err)
	require.Len(t, fields, 1)
	assert.Equal(t, int64(51), fields[0].ID)

	foundField, err := repo.GetFieldByID(51)
	require.NoError(t, err)
	require.NotNil(t, foundField.Name)
	assert.Equal(t, "City", *foundField.Name)
}

func TestDatasetRepository_DeleteFieldOperations(t *testing.T) {
	repo, db := setupDatasetRepositoryTest(t)
	chartID := int64(900)
	otherChartID := int64(901)
	nameA := "field_a"
	originA := "field_a"
	nameB := "field_b"
	originB := "field_b"
	require.NoError(t, db.Create(&dataset.CoreDatasetTableField{ID: 801, DatasetGroupID: 77, ChartID: &chartID, Name: &nameA, OriginName: &originA}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetTableField{ID: 802, DatasetGroupID: 77, ChartID: &chartID, Name: &nameB, OriginName: &originB}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetTableField{ID: 803, DatasetGroupID: 77, ChartID: &otherChartID, Name: &nameB, OriginName: &originB}).Error)

	deleted, err := repo.DeleteFieldByIDAndChartID(801, chartID)
	require.NoError(t, err)
	assert.True(t, deleted)
	_, err = repo.GetFieldByID(801)
	require.Error(t, err)

	deleted, err = repo.DeleteFieldByIDAndChartID(802, otherChartID)
	require.NoError(t, err)
	assert.False(t, deleted)

	rows, err := repo.DeleteFieldsByChartID(chartID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), rows)

	_, err = repo.GetFieldByID(802)
	require.Error(t, err)
	found, err := repo.GetFieldByID(803)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, int64(803), found.ID)
}

func TestDatasetRepository_FindNearestGroupIDInWindow(t *testing.T) {
	repo, db := setupDatasetRepositoryTest(t)
	level := 1
	nodeType := "dataset"
	require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 101, Name: "G101", Level: &level, NodeType: &nodeType}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 240, Name: "G240", Level: &level, NodeType: &nodeType}).Error)

	id, err := repo.FindNearestGroupIDInWindow(200, 100)
	require.NoError(t, err)
	require.NotNil(t, id)
	assert.Equal(t, int64(240), *id)

	id, err = repo.FindNearestGroupIDInWindow(10000, 50)
	require.NoError(t, err)
	assert.Nil(t, id)
}

func TestDatasetRepository_PreviewRowsAndCountRows_Success(t *testing.T) {
	repo, _ := setupDatasetRepositoryTest(t)

	rows, err := repo.PreviewRows("preview_rows", 2)
	require.NoError(t, err)
	require.Len(t, rows, 2)

	count, err := repo.CountRows("preview_rows")
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestDatasetRepository_CountChartRelations(t *testing.T) {
	repo, db := setupDatasetRepositoryTest(t)
	groupName := "Group A"
	tableName := "preview_rows"
	require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 61, Name: groupName}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetTable{ID: 71, DatasetGroupID: 61, PhysicalTable: &tableName}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetTable{ID: 72, DatasetGroupID: 61, PhysicalTable: &tableName}).Error)
	require.NoError(t, db.Exec("INSERT INTO core_chart_view (id, table_id) VALUES (1, 71), (2, 71), (3, 72)").Error)

	count, err := repo.CountChartRelations(61)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	count, err = repo.CountChartRelations(999)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestDatasetRepository_QueryDistinctObjectValues(t *testing.T) {
	repo, _ := setupDatasetRepositoryTest(t)

	rows, err := repo.QueryDistinctObjectValues("preview_rows", nil, nil, "", "", "", "", 0)
	require.NoError(t, err)
	assert.Empty(t, rows)

	rows, err = repo.QueryDistinctObjectValues(
		"preview_rows",
		[]dataset.EnumObjectColumn{{Column: "city", Alias: "label"}, {Column: "category", Alias: "value"}},
		[]dataset.EnumFilterClause{{Column: "category", Values: []string{"A", "B"}}, {Column: " ", Values: []string{"skip"}}},
		"city",
		"h",
		"city",
		"DESC",
		1,
	)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "Shanghai", rows[0]["label"])
	assert.Equal(t, "A", rows[0]["value"])

	rows, err = repo.QueryDistinctObjectValues(
		"preview_rows",
		[]dataset.EnumObjectColumn{{Column: "city", Alias: "label"}},
		nil,
		"",
		"",
		"city",
		"bad",
		0,
	)
	require.NoError(t, err)
	require.Len(t, rows, 3)
	assert.Equal(t, "Beijing", rows[0]["label"])
}

func TestQuoteIdentifier(t *testing.T) {
	quoted, err := quoteIdentifier(" city ")
	require.NoError(t, err)
	assert.Equal(t, "`city`", quoted)

	quoted, err = quoteIdentifier("a`b")
	require.NoError(t, err)
	assert.Equal(t, "`a``b`", quoted)

	_, err = quoteIdentifier("   ")
	require.Error(t, err)
	assert.Equal(t, "invalid identifier", err.Error())
}

func TestPreviewRows_InvalidTableName(t *testing.T) {
	repo := &DatasetRepository{}
	_, err := repo.PreviewRows("core_dataset_table;drop table x", 10)
	if err == nil {
		t.Fatal("expected invalid table name error")
	}
}

func TestCountRows_InvalidTableName(t *testing.T) {
	repo := &DatasetRepository{}
	_, err := repo.CountRows("x` or 1=1 --")
	if err == nil {
		t.Fatal("expected invalid table name error")
	}
}
