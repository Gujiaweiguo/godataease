package repository

import (
	"strconv"
	"testing"
	"time"

	"dataease/backend/internal/domain/auto"
	templatedomain "dataease/backend/internal/domain/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTemplateRepositoryTest(t *testing.T) (*TemplateRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&coreVisualizationTemplate{}, &auto.VisualizationTemplateCategoryMap{}))

	return NewTemplateRepository(db), db
}

func createTemplateRecord(t *testing.T, db *gorm.DB, record *coreVisualizationTemplate) {
	t.Helper()
	require.NoError(t, db.Create(record).Error)
}

func createTemplateCategoryMap(t *testing.T, db *gorm.DB, categoryID, templateID string) {
	t.Helper()
	require.NoError(t, db.Create(&auto.VisualizationTemplateCategoryMap{ID: categoryID + "-" + templateID, CategoryID: categoryID, TemplateID: templateID}).Error)
}

func TestTemplateRepository_CRUDAndCategoryQueries(t *testing.T) {
	repo, db := setupTemplateRepositoryTest(t)
	now := time.Now()

	rootFolder := &coreVisualizationTemplate{ID: 101, Name: "Folder A", NodeType: templatedomain.NodeTypeFolder, Level: 0, TemplateType: "system", CreateTime: ptrTimeTemplateRepo(now.Add(-5 * time.Hour))}
	otherFolder := &coreVisualizationTemplate{ID: 102, Name: "Folder B", NodeType: templatedomain.NodeTypeFolder, Level: 1, TemplateType: "custom", CreateTime: ptrTimeTemplateRepo(now.Add(-4 * time.Hour))}
	chartA := &coreVisualizationTemplate{ID: 201, Name: "Sales Template", Pid: 101, NodeType: "leaf", DvType: "bar", TemplateType: "system", UseCount: 2, CreateTime: ptrTimeTemplateRepo(now.Add(-3 * time.Hour))}
	chartB := &coreVisualizationTemplate{ID: 202, Name: "Profit Template", Pid: 101, NodeType: "leaf", DvType: "line", TemplateType: "custom", CreateTime: ptrTimeTemplateRepo(now.Add(-2 * time.Hour))}
	chartC := &coreVisualizationTemplate{ID: 203, Name: "Sales Template", Pid: 102, NodeType: "leaf", DvType: "bar", TemplateType: "custom", CreateTime: ptrTimeTemplateRepo(now.Add(-1 * time.Hour))}

	for _, record := range []*coreVisualizationTemplate{rootFolder, otherFolder, chartA, chartB, chartC} {
		createTemplateRecord(t, db, record)
	}

	createTemplateCategoryMap(t, db, "101", "201")
	createTemplateCategoryMap(t, db, "101", "202")
	createTemplateCategoryMap(t, db, "102", "203")

	created := &templatedomain.Template{
		Name:          "Created Template",
		Pid:           101,
		Level:         1,
		DvType:        "pie",
		NodeType:      "leaf",
		CreateBy:      "alice",
		Snapshot:      "snap",
		TemplateType:  "custom",
		TemplateStyle: "style",
		TemplateData:  "data",
		DynamicData:   "dynamic",
		AppData:       "app",
		UseCount:      4,
		Version:       7,
	}
	require.NoError(t, repo.Create(created))
	require.Positive(t, created.ID)
	require.NotNil(t, created.CreateTime)

	found, err := repo.GetByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.Name, found.Name)
	assert.Equal(t, created.Version, found.Version)

	listByPid, err := repo.List(101, "")
	require.NoError(t, err)
	require.Len(t, listByPid, 3)
	assert.Equal(t, created.ID, listByPid[0].ID)
	assert.Equal(t, chartB.ID, listByPid[1].ID)
	assert.Equal(t, chartA.ID, listByPid[2].ID)

	listByDvType, err := repo.List(0, "bar")
	require.NoError(t, err)
	require.Len(t, listByDvType, 2)
	assert.Equal(t, chartC.ID, listByDvType[0].ID)
	assert.Equal(t, chartA.ID, listByDvType[1].ID)

	categories, err := repo.ListCategories(0, "system")
	require.NoError(t, err)
	require.Len(t, categories, 1)
	assert.Equal(t, rootFolder.ID, categories[0].ID)

	byCategory, err := repo.ListByCategory("101", "bar")
	require.NoError(t, err)
	require.Len(t, byCategory, 1)
	assert.Equal(t, chartA.ID, byCategory[0].ID)

	countByCategory, err := repo.CountByCategory("101", "")
	require.NoError(t, err)
	assert.Equal(t, int64(3), countByCategory)

	countByPid, err := repo.Count(101, "")
	require.NoError(t, err)
	assert.Equal(t, int64(3), countByPid)

	countBar, err := repo.Count(0, "bar")
	require.NoError(t, err)
	assert.Equal(t, int64(2), countBar)

	created.Name = "Renamed Template"
	created.Snapshot = "snap-2"
	created.TemplateStyle = "style-2"
	created.TemplateData = "data-2"
	created.DynamicData = "dynamic-2"
	created.AppData = "app-2"
	require.NoError(t, repo.Update(created))

	updated, err := repo.GetByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Renamed Template", updated.Name)
	assert.Equal(t, "app-2", updated.AppData)

	require.NoError(t, repo.IncrementUseCount(chartA.ID))
	chartAAfterIncrement, err := repo.GetByID(chartA.ID)
	require.NoError(t, err)
	assert.Equal(t, 3, chartAAfterIncrement.UseCount)

	nameCount, err := repo.CountByName("Sales Template", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(2), nameCount)

	nameCount, err = repo.CountByName("Sales Template", &chartA.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), nameCount)

	nameInCategories, err := repo.CountByNameInCategories("Sales Template", []string{"101", "102"})
	require.NoError(t, err)
	assert.Equal(t, int64(2), nameInCategories)

	nameInCategories, err = repo.CountByNameInCategories("Missing", []string{"999"})
	require.NoError(t, err)
	assert.Zero(t, nameInCategories)

	batchCount, err := repo.CountBatchNamesInCategories([]string{"Sales Template", "Profit Template"}, []string{"101", "102"}, []string{strconv.FormatInt(chartC.ID, 10)})
	require.NoError(t, err)
	assert.Equal(t, int64(2), batchCount)

	batchCount, err = repo.CountBatchNamesInCategories(nil, []string{"101"}, nil)
	require.NoError(t, err)
	assert.Zero(t, batchCount)

	require.NoError(t, repo.Delete(created.ID))
	_, err = repo.GetByID(created.ID)
	require.Error(t, err)

	createTemplateCategoryMap(t, db, "500", strconv.FormatInt(chartB.ID, 10))
	require.NoError(t, repo.Delete(chartB.ID))
	_, err = repo.GetByID(chartB.ID)
	require.Error(t, err)
	var mappingCount int64
	require.NoError(t, db.Model(&auto.VisualizationTemplateCategoryMap{}).Where("template_id = ?", strconv.FormatInt(chartB.ID, 10)).Count(&mappingCount).Error)
	assert.Zero(t, mappingCount)
}

func TestTemplateRepository_CategoryMaintenance(t *testing.T) {
	repo, db := setupTemplateRepositoryTest(t)

	category := &coreVisualizationTemplate{ID: 301, Name: "Folder C", NodeType: templatedomain.NodeTypeFolder, Level: 1, CreateTime: ptrTimeTemplateRepo(time.Now().Add(-2 * time.Hour))}
	templateA := &coreVisualizationTemplate{ID: 401, Name: "Template A", Pid: 301, NodeType: "leaf", DvType: "bar", CreateTime: ptrTimeTemplateRepo(time.Now().Add(-time.Hour))}
	templateB := &coreVisualizationTemplate{ID: 402, Name: "Template B", Pid: 301, NodeType: "leaf", DvType: "line", CreateTime: ptrTimeTemplateRepo(time.Now())}
	for _, record := range []*coreVisualizationTemplate{category, templateA, templateB} {
		createTemplateRecord(t, db, record)
	}

	require.NoError(t, repo.SyncTemplateCategories(templateA.ID, []string{"301", "301", "302", ""}))
	require.NoError(t, repo.SyncTemplateCategories(templateB.ID, []string{"302", "303"}))

	categoryIDs, err := repo.FindCategoryIDsByTemplateIDs([]string{strconv.FormatInt(templateA.ID, 10), strconv.FormatInt(templateB.ID, 10)})
	require.NoError(t, err)
	assert.Equal(t, []string{"301", "302", "303"}, categoryIDs)

	emptyCategoryIDs, err := repo.FindCategoryIDsByTemplateIDs(nil)
	require.NoError(t, err)
	assert.Empty(t, emptyCategoryIDs)

	remaining, err := repo.UnlinkCategory(templateB.ID, "302")
	require.NoError(t, err)
	assert.Equal(t, int64(1), remaining)

	linkedToCategory, err := repo.ListByCategory("301", "")
	require.NoError(t, err)
	require.Len(t, linkedToCategory, 2)
	assert.Equal(t, templateB.ID, linkedToCategory[0].ID)
	assert.Equal(t, templateA.ID, linkedToCategory[1].ID)

	require.NoError(t, repo.UpdateTemplatePid(templateA.ID, 999))
	updated, err := repo.GetByID(templateA.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(999), updated.Pid)

	require.NoError(t, repo.DeleteCategory("301"))
	_, err = repo.GetByID(301)
	require.Error(t, err)

	var count int64
	require.NoError(t, db.Model(&auto.VisualizationTemplateCategoryMap{}).Where("category_id = ?", "301").Count(&count).Error)
	assert.Zero(t, count)

	err = repo.DeleteCategory("invalid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid syntax")

	countByMissingCategory, err := repo.CountByCategory("not-a-number", "")
	require.NoError(t, err)
	assert.Zero(t, countByMissingCategory)
	listByMissingCategory, err := repo.ListByCategory("not-a-number", "")
	require.NoError(t, err)
	assert.Empty(t, listByMissingCategory)
	nameCount, err := repo.CountBatchNamesInCategories([]string{"Template A"}, []string{"missing"}, nil)
	require.NoError(t, err)
	assert.Zero(t, nameCount)
}

func ptrTimeTemplateRepo(v time.Time) *time.Time {
	return &v
}
