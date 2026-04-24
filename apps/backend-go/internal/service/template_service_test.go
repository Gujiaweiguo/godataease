package service

import (
	"strconv"
	"testing"
	"time"

	"dataease/backend/internal/domain/template"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type testCoreVisualizationTemplate struct {
	ID            int64      `gorm:"column:id;primaryKey;autoIncrement"`
	Name          string     `gorm:"column:name;size:255"`
	Pid           int64      `gorm:"column:pid;index"`
	Level         int        `gorm:"column:level"`
	DvType        string     `gorm:"column:dv_type;size:50"`
	NodeType      string     `gorm:"column:node_type;size:50"`
	CreateBy      string     `gorm:"column:create_by;size:255"`
	CreateTime    *time.Time `gorm:"column:create_time"`
	Snapshot      string     `gorm:"column:snapshot;type:longtext"`
	TemplateType  string     `gorm:"column:template_type;size:50"`
	TemplateStyle string     `gorm:"column:template_style;type:longtext"`
	TemplateData  string     `gorm:"column:template_data;type:longtext"`
	DynamicData   string     `gorm:"column:dynamic_data;type:longtext"`
	AppData       string     `gorm:"column:app_data;type:longtext"`
	UseCount      int        `gorm:"column:use_count;default:0"`
	Version       int        `gorm:"column:version;default:3"`
}

type testVisualizationTemplateCategoryMap struct {
	ID         string `gorm:"column:id;primaryKey"`
	CategoryID string `gorm:"column:category_id"`
	TemplateID string `gorm:"column:template_id"`
}

func (testVisualizationTemplateCategoryMap) TableName() string {
	return "visualization_template_category_map"
}

func (testCoreVisualizationTemplate) TableName() string {
	return "core_visualization_template"
}

func setupTemplateServiceRepoTest(t *testing.T) (*TemplateService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&testCoreVisualizationTemplate{}))
	require.NoError(t, db.AutoMigrate(&testVisualizationTemplateCategoryMap{}))

	repo := repository.NewTemplateRepository(db)
	return NewTemplateService(repo), db
}

func createTemplateFixture(t *testing.T, svc *TemplateService, name string) *template.Template {
	t.Helper()

	created, err := svc.CreateTemplate(&template.TemplateCreateRequest{Name: name, Pid: 0, DvType: "dashboard", NodeType: "leaf"}, "tester")
	require.NoError(t, err)
	return created
}

func TestTemplateService_CreateTemplate(t *testing.T) {
	svc, _ := setupTemplateServiceRepoTest(t)

	created, err := svc.CreateTemplate(&template.TemplateCreateRequest{
		Name:          "Test Template",
		Pid:           0,
		DvType:        "dashboard",
		NodeType:      "folder",
		Snapshot:      "snapshot-data",
		TemplateType:  "self",
		TemplateStyle: "dark",
		TemplateData:  "{}",
	}, "creator1")
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, "Test Template", created.Name)
	assert.Equal(t, "creator1", created.CreateBy)
	assert.Equal(t, 0, created.UseCount)
	assert.Equal(t, 3, created.Version)

	t.Run("create error", func(t *testing.T) {
		svc, db := setupTemplateServiceRepoTest(t)
		require.NoError(t, db.Exec("CREATE TRIGGER deny_template_insert BEFORE INSERT ON core_visualization_template BEGIN SELECT RAISE(FAIL, 'deny insert'); END;").Error)

		created, err := svc.CreateTemplate(&template.TemplateCreateRequest{Name: "Fail Create", Pid: 0, DvType: "dashboard", NodeType: "leaf"}, "creator1")
		require.Error(t, err)
		assert.Nil(t, created)
	})

	t.Run("maps optional payload fields", func(t *testing.T) {
		svc, _ := setupTemplateServiceRepoTest(t)

		created, err := svc.CreateTemplate(&template.TemplateCreateRequest{
			Name:          "Payload Template",
			Pid:           9,
			DvType:        "screen",
			NodeType:      "leaf",
			Snapshot:      "snapshot-payload",
			TemplateType:  "official",
			TemplateStyle: "modern",
			TemplateData:  `{"layout":1}`,
			DynamicData:   `{"dynamic":true}`,
			AppData:       `{"app":"demo"}`,
		}, "creator2")
		require.NoError(t, err)
		require.NotNil(t, created)
		assert.Equal(t, int64(9), created.Pid)
		assert.Equal(t, "screen", created.DvType)
		assert.Equal(t, "official", created.TemplateType)
		assert.Equal(t, "modern", created.TemplateStyle)
		assert.Equal(t, `{"layout":1}`, created.TemplateData)
		assert.Equal(t, `{"dynamic":true}`, created.DynamicData)
		assert.Equal(t, `{"app":"demo"}`, created.AppData)
		require.NotNil(t, created.CreateTime)
	})
}

func TestTemplateService_GetTemplate(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		svc, _ := setupTemplateServiceRepoTest(t)
		created, err := svc.CreateTemplate(&template.TemplateCreateRequest{Name: "Get Test", Pid: 0, DvType: "dashboard", NodeType: "leaf"}, "tester")
		require.NoError(t, err)

		loaded, loadErr := svc.GetTemplate(created.ID)
		require.NoError(t, loadErr)
		require.NotNil(t, loaded)
		assert.Equal(t, created.ID, loaded.ID)
		assert.Equal(t, "Get Test", loaded.Name)
	})

	t.Run("not found", func(t *testing.T) {
		svc, _ := setupTemplateServiceRepoTest(t)

		loaded, err := svc.GetTemplate(99999)
		require.Error(t, err)
		assert.Nil(t, loaded)
	})
}

func TestTemplateService_ListTemplates(t *testing.T) {
	t.Run("invalid pid falls back to zero", func(t *testing.T) {
		svc, _ := setupTemplateServiceRepoTest(t)
		_, err := svc.CreateTemplate(&template.TemplateCreateRequest{Name: "InvalidPid", Pid: 0, DvType: "dashboard", NodeType: "leaf"}, "tester")
		require.NoError(t, err)

		resp, listErr := svc.ListTemplates(&template.TemplateListRequest{Pid: "bad", DvType: "dashboard"})
		require.NoError(t, listErr)
		require.NotNil(t, resp)
		assert.Equal(t, int64(1), resp.Total)
		assert.Len(t, resp.List, 1)
	})

	t.Run("empty", func(t *testing.T) {
		svc, _ := setupTemplateServiceRepoTest(t)

		resp, err := svc.ListTemplates(&template.TemplateListRequest{Pid: "0", DvType: "dashboard"})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Zero(t, resp.Total)
		assert.Empty(t, resp.List)
	})

	t.Run("list error", func(t *testing.T) {
		svc, db := setupTemplateServiceRepoTest(t)
		require.NoError(t, db.Exec("DROP TABLE core_visualization_template").Error)

		resp, err := svc.ListTemplates(&template.TemplateListRequest{Pid: "0", DvType: "dashboard"})
		require.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("valid pid filter limits list and total", func(t *testing.T) {
		svc, _ := setupTemplateServiceRepoTest(t)
		_, err := svc.CreateTemplate(&template.TemplateCreateRequest{Name: "ParentA", Pid: 1, DvType: "dashboard", NodeType: "leaf"}, "tester")
		require.NoError(t, err)
		_, err = svc.CreateTemplate(&template.TemplateCreateRequest{Name: "ParentB", Pid: 2, DvType: "dashboard", NodeType: "leaf"}, "tester")
		require.NoError(t, err)

		resp, err := svc.ListTemplates(&template.TemplateListRequest{Pid: "1", DvType: "dashboard"})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, int64(1), resp.Total)
		require.Len(t, resp.List, 1)
		assert.Equal(t, int64(1), resp.List[0].Pid)
	})

	t.Run("empty dv type returns all templates for pid", func(t *testing.T) {
		svc, _ := setupTemplateServiceRepoTest(t)
		_, err := svc.CreateTemplate(&template.TemplateCreateRequest{Name: "Dash", Pid: 3, DvType: "dashboard", NodeType: "leaf"}, "tester")
		require.NoError(t, err)
		_, err = svc.CreateTemplate(&template.TemplateCreateRequest{Name: "Screen", Pid: 3, DvType: "screen", NodeType: "leaf"}, "tester")
		require.NoError(t, err)

		resp, err := svc.ListTemplates(&template.TemplateListRequest{Pid: "3", DvType: ""})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, int64(2), resp.Total)
		require.Len(t, resp.List, 2)
	})
}

func TestTemplateService_UpdateTemplate(t *testing.T) {
	t.Run("updates only provided fields", func(t *testing.T) {
		svc, _ := setupTemplateServiceRepoTest(t)
		created, err := svc.CreateTemplate(&template.TemplateCreateRequest{Name: "Old Name", Pid: 0, DvType: "dashboard", NodeType: "leaf", Snapshot: "old-snapshot"}, "tester")
		require.NoError(t, err)

		updated, updateErr := svc.UpdateTemplate(&template.TemplateUpdateRequest{ID: created.ID, Name: "New Name", Snapshot: "new-snapshot", TemplateStyle: "light", TemplateData: "{\"updated\":true}"})
		require.NoError(t, updateErr)
		assert.Equal(t, "New Name", updated.Name)
		assert.Equal(t, "new-snapshot", updated.Snapshot)
		assert.Equal(t, "light", updated.TemplateStyle)
		assert.Equal(t, "{\"updated\":true}", updated.TemplateData)
	})

	t.Run("empty fields keep original", func(t *testing.T) {
		svc, _ := setupTemplateServiceRepoTest(t)
		created, err := svc.CreateTemplate(&template.TemplateCreateRequest{Name: "Keep Name", Pid: 0, DvType: "dashboard", NodeType: "leaf", Snapshot: "snap"}, "tester")
		require.NoError(t, err)

		updated, updateErr := svc.UpdateTemplate(&template.TemplateUpdateRequest{ID: created.ID})
		require.NoError(t, updateErr)
		assert.Equal(t, "Keep Name", updated.Name)
		assert.Equal(t, "snap", updated.Snapshot)
	})

	t.Run("not found", func(t *testing.T) {
		svc, _ := setupTemplateServiceRepoTest(t)

		updated, err := svc.UpdateTemplate(&template.TemplateUpdateRequest{ID: 99999, Name: "missing"})
		require.Error(t, err)
		assert.Nil(t, updated)
	})

	t.Run("update error", func(t *testing.T) {
		svc, db := setupTemplateServiceRepoTest(t)
		created := createTemplateFixture(t, svc, "Update Error")
		require.NoError(t, db.Exec("CREATE TRIGGER deny_template_update BEFORE UPDATE ON core_visualization_template BEGIN SELECT RAISE(FAIL, 'deny update'); END;").Error)

		updated, err := svc.UpdateTemplate(&template.TemplateUpdateRequest{ID: created.ID, Name: "new-name"})
		require.Error(t, err)
		assert.Nil(t, updated)
	})

	t.Run("updates dynamic and app data when provided", func(t *testing.T) {
		svc, _ := setupTemplateServiceRepoTest(t)
		created, err := svc.CreateTemplate(&template.TemplateCreateRequest{
			Name:         "Dynamic Template",
			Pid:          0,
			DvType:       "dashboard",
			NodeType:     "leaf",
			DynamicData:  `{"old":true}`,
			AppData:      `{"old":"app"}`,
			TemplateData: `{"same":1}`,
		}, "tester")
		require.NoError(t, err)

		updated, err := svc.UpdateTemplate(&template.TemplateUpdateRequest{ID: created.ID, DynamicData: `{"new":true}`, AppData: `{"new":"app"}`})
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, `{"new":true}`, updated.DynamicData)
		assert.Equal(t, `{"new":"app"}`, updated.AppData)
		assert.Equal(t, `{"same":1}`, updated.TemplateData)
	})
}

func TestTemplateService_NameCheck(t *testing.T) {
	svc, _ := setupTemplateServiceRepoTest(t)
	created := createTemplateFixture(t, svc, "Repeated Name")

	result, err := svc.NameCheck("insert", "Repeated Name", "")
	require.NoError(t, err)
	assert.Equal(t, "existAll", result)

	result, err = svc.NameCheck("update", "Repeated Name", strconv.FormatInt(created.ID, 10))
	require.NoError(t, err)
	assert.Equal(t, "none", result)

	result, err = svc.NameCheck("insert", "Fresh Name", "")
	require.NoError(t, err)
	assert.Equal(t, "none", result)
}

func TestTemplateService_CategoryTemplateNameCheck(t *testing.T) {
	svc, db := setupTemplateServiceRepoTest(t)
	created := createTemplateFixture(t, svc, "Category Template")
	require.NoError(t, db.Create(&testVisualizationTemplateCategoryMap{ID: "map-1", CategoryID: "cat-1", TemplateID: strconv.FormatInt(created.ID, 10)}).Error)

	result, err := svc.CategoryTemplateNameCheck("Category Template", []string{"cat-1"}, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "existAll", result)

	result, err = svc.CategoryTemplateNameCheck("Other Template", []string{"cat-1"}, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "none", result)

	result, err = svc.CategoryTemplateNameCheck("", []string{"cat-1"}, []string{"Category Template"}, nil)
	require.NoError(t, err)
	assert.Equal(t, "existAll", result)

	result, err = svc.CategoryTemplateNameCheck("", []string{"cat-1"}, []string{"Category Template"}, []string{strconv.FormatInt(created.ID, 10)})
	require.NoError(t, err)
	assert.Equal(t, "none", result)
}

func TestTemplateService_ListTemplates_ByCategory(t *testing.T) {
	svc, db := setupTemplateServiceRepoTest(t)
	categoryA := createTemplateFixture(t, svc, "Category A")
	categoryA.NodeType = "folder"
	categoryA.TemplateType = "self"
	require.NoError(t, db.Model(&testCoreVisualizationTemplate{}).Where("id = ?", categoryA.ID).Updates(map[string]interface{}{
		"node_type":     "folder",
		"template_type": "self",
		"level":         0,
	}).Error)
	categoryB := createTemplateFixture(t, svc, "Category B")
	require.NoError(t, db.Model(&testCoreVisualizationTemplate{}).Where("id = ?", categoryB.ID).Updates(map[string]interface{}{
		"node_type":     "folder",
		"template_type": "self",
		"level":         0,
	}).Error)
	templateA := createTemplateFixture(t, svc, "Template A")
	require.NoError(t, db.Model(&testCoreVisualizationTemplate{}).Where("id = ?", templateA.ID).Update("pid", categoryA.ID).Error)
	templateB := createTemplateFixture(t, svc, "Template B")
	require.NoError(t, db.Model(&testCoreVisualizationTemplate{}).Where("id = ?", templateB.ID).Update("pid", categoryB.ID).Error)
	require.NoError(t, db.Create(&testVisualizationTemplateCategoryMap{ID: "map-a", CategoryID: strconv.FormatInt(categoryA.ID, 10), TemplateID: strconv.FormatInt(templateA.ID, 10)}).Error)

	resp, err := svc.ListTemplates(&template.TemplateListRequest{CategoryID: strconv.FormatInt(categoryA.ID, 10), DvType: "dashboard"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int64(1), resp.Total)
	require.Len(t, resp.List, 1)
	assert.Equal(t, templateA.ID, resp.List[0].ID)
}

func TestTemplateService_SaveTemplate(t *testing.T) {
	t.Run("create with categories", func(t *testing.T) {
		svc, db := setupTemplateServiceRepoTest(t)
		category := createTemplateFixture(t, svc, "Save Category")
		require.NoError(t, db.Model(&testCoreVisualizationTemplate{}).Where("id = ?", category.ID).Updates(map[string]interface{}{
			"node_type":     "folder",
			"template_type": "self",
			"level":         0,
		}).Error)

		created, err := svc.SaveTemplate(&template.TemplateSaveRequest{
			Name:          "Saved Template",
			Pid:           category.ID,
			Level:         1,
			DvType:        "dashboard",
			NodeType:      "template",
			TemplateType:  "self",
			TemplateStyle: `{"layout":1}`,
			TemplateData:  `[]`,
			DynamicData:   `{}`,
			AppData:       `{}`,
			Categories:    []string{strconv.FormatInt(category.ID, 10)},
		}, "tester")
		require.NoError(t, err)
		require.NotNil(t, created)

		categoryIDs, err := svc.FindCategoriesByTemplateIDs([]string{strconv.FormatInt(created.ID, 10)})
		require.NoError(t, err)
		assert.Equal(t, []string{strconv.FormatInt(category.ID, 10)}, categoryIDs)
	})

	t.Run("update syncs category and pid", func(t *testing.T) {
		svc, db := setupTemplateServiceRepoTest(t)
		categoryA := createTemplateFixture(t, svc, "Category A")
		categoryB := createTemplateFixture(t, svc, "Category B")
		for _, category := range []*template.Template{categoryA, categoryB} {
			require.NoError(t, db.Model(&testCoreVisualizationTemplate{}).Where("id = ?", category.ID).Updates(map[string]interface{}{
				"node_type":     "folder",
				"template_type": "self",
				"level":         0,
			}).Error)
		}
		created := createTemplateFixture(t, svc, "Template To Update")
		require.NoError(t, db.Create(&testVisualizationTemplateCategoryMap{ID: "map-old", CategoryID: strconv.FormatInt(categoryA.ID, 10), TemplateID: strconv.FormatInt(created.ID, 10)}).Error)

		updated, err := svc.SaveTemplate(&template.TemplateSaveRequest{
			ID:            created.ID,
			Name:          "Template Updated",
			NodeType:      "template",
			TemplateStyle: `{}`,
			TemplateData:  `[]`,
			DynamicData:   `{}`,
			AppData:       `{}`,
			Categories:    []string{strconv.FormatInt(categoryB.ID, 10)},
		}, "tester")
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, categoryB.ID, updated.Pid)

		categoryIDs, err := svc.FindCategoriesByTemplateIDs([]string{strconv.FormatInt(created.ID, 10)})
		require.NoError(t, err)
		assert.Equal(t, []string{strconv.FormatInt(categoryB.ID, 10)}, categoryIDs)
	})
}

func TestTemplateService_ListCategories(t *testing.T) {
	svc, db := setupTemplateServiceRepoTest(t)
	folderA := createTemplateFixture(t, svc, "Folder A")
	folderB := createTemplateFixture(t, svc, "Folder B")
	leaf := createTemplateFixture(t, svc, "Leaf")
	require.NoError(t, db.Model(&testCoreVisualizationTemplate{}).Where("id = ?", folderA.ID).Updates(map[string]interface{}{
		"node_type":     "folder",
		"template_type": "self",
		"level":         0,
	}).Error)
	require.NoError(t, db.Model(&testCoreVisualizationTemplate{}).Where("id = ?", folderB.ID).Updates(map[string]interface{}{
		"node_type":     "folder",
		"template_type": "system",
		"level":         1,
	}).Error)
	require.NoError(t, db.Model(&testCoreVisualizationTemplate{}).Where("id = ?", leaf.ID).Update("node_type", "template").Error)

	result, err := svc.ListCategories("0", "self")
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, folderA.ID, result[0].ID)
}

func TestTemplateService_FindCategoriesByTemplateIDs(t *testing.T) {
	svc, db := setupTemplateServiceRepoTest(t)
	tpl := createTemplateFixture(t, svc, "Find Category Template")
	require.NoError(t, db.Create(&testVisualizationTemplateCategoryMap{ID: "map-find", CategoryID: "cat-find", TemplateID: strconv.FormatInt(tpl.ID, 10)}).Error)

	result, err := svc.FindCategoriesByTemplateIDs([]string{strconv.FormatInt(tpl.ID, 10)})
	require.NoError(t, err)
	assert.Equal(t, []string{"cat-find"}, result)
}

func TestTemplateService_BatchUpdateCategories(t *testing.T) {
	svc, db := setupTemplateServiceRepoTest(t)
	categoryA := createTemplateFixture(t, svc, "Batch Category A")
	categoryB := createTemplateFixture(t, svc, "Batch Category B")
	for _, category := range []*template.Template{categoryA, categoryB} {
		require.NoError(t, db.Model(&testCoreVisualizationTemplate{}).Where("id = ?", category.ID).Updates(map[string]interface{}{
			"node_type":     "folder",
			"template_type": "self",
			"level":         0,
		}).Error)
	}
	tpl1 := createTemplateFixture(t, svc, "Batch Template 1")
	tpl2 := createTemplateFixture(t, svc, "Batch Template 2")
	require.NoError(t, db.Create(&testVisualizationTemplateCategoryMap{ID: "map-b1", CategoryID: strconv.FormatInt(categoryA.ID, 10), TemplateID: strconv.FormatInt(tpl1.ID, 10)}).Error)

	err := svc.BatchUpdateCategories(
		[]string{strconv.FormatInt(tpl1.ID, 10), strconv.FormatInt(tpl2.ID, 10)},
		[]string{strconv.FormatInt(categoryB.ID, 10)},
	)
	require.NoError(t, err)

	result, err := svc.ListTemplates(&template.TemplateListRequest{CategoryID: strconv.FormatInt(categoryB.ID, 10), DvType: "dashboard"})
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.Total)
}

func TestTemplateService_DeleteCategory(t *testing.T) {
	t.Run("refuses non-empty category", func(t *testing.T) {
		svc, db := setupTemplateServiceRepoTest(t)
		category := createTemplateFixture(t, svc, "Delete Category")
		require.NoError(t, db.Model(&testCoreVisualizationTemplate{}).Where("id = ?", category.ID).Updates(map[string]interface{}{
			"node_type":     "folder",
			"template_type": "self",
			"level":         0,
		}).Error)
		tpl := createTemplateFixture(t, svc, "Delete Category Template")
		require.NoError(t, db.Model(&testCoreVisualizationTemplate{}).Where("id = ?", tpl.ID).Update("pid", category.ID).Error)
		require.NoError(t, db.Create(&testVisualizationTemplateCategoryMap{ID: "map-delete", CategoryID: strconv.FormatInt(category.ID, 10), TemplateID: strconv.FormatInt(tpl.ID, 10)}).Error)

		deleted, err := svc.DeleteCategory(strconv.FormatInt(category.ID, 10))
		require.NoError(t, err)
		assert.False(t, deleted)
	})

	t.Run("deletes empty category", func(t *testing.T) {
		svc, db := setupTemplateServiceRepoTest(t)
		category := createTemplateFixture(t, svc, "Empty Category")
		require.NoError(t, db.Model(&testCoreVisualizationTemplate{}).Where("id = ?", category.ID).Updates(map[string]interface{}{
			"node_type":     "folder",
			"template_type": "self",
			"level":         0,
		}).Error)

		deleted, err := svc.DeleteCategory(strconv.FormatInt(category.ID, 10))
		require.NoError(t, err)
		assert.True(t, deleted)

		_, loadErr := svc.GetTemplate(category.ID)
		require.Error(t, loadErr)
	})
}

func TestTemplateService_DeleteTemplate(t *testing.T) {
	t.Run("not found returns nil", func(t *testing.T) {
		svc, _ := setupTemplateServiceRepoTest(t)
		require.NoError(t, svc.DeleteTemplate(99999))
	})

	t.Run("deletes existing", func(t *testing.T) {
		svc, _ := setupTemplateServiceRepoTest(t)
		created, err := svc.CreateTemplate(&template.TemplateCreateRequest{Name: "To Delete", Pid: 0, DvType: "dashboard", NodeType: "leaf"}, "tester")
		require.NoError(t, err)

		require.NoError(t, svc.DeleteTemplate(created.ID))
		loaded, loadErr := svc.GetTemplate(created.ID)
		require.Error(t, loadErr)
		assert.Nil(t, loaded)
	})

	t.Run("delete error", func(t *testing.T) {
		svc, db := setupTemplateServiceRepoTest(t)
		created := createTemplateFixture(t, svc, "Delete Error")
		require.NoError(t, db.Exec("CREATE TRIGGER deny_template_delete BEFORE DELETE ON core_visualization_template BEGIN SELECT RAISE(FAIL, 'deny delete'); END;").Error)

		err := svc.DeleteTemplate(created.ID)
		require.Error(t, err)
	})

	t.Run("get by id repo error propagates", func(t *testing.T) {
		svc, db := setupTemplateServiceRepoTest(t)
		sqlDB, err := db.DB()
		require.NoError(t, err)
		require.NoError(t, sqlDB.Close())

		err = svc.DeleteTemplate(1)
		require.Error(t, err)
	})
}

func TestTemplateService_IncrementUseCount(t *testing.T) {
	svc, _ := setupTemplateServiceRepoTest(t)
	created, err := svc.CreateTemplate(&template.TemplateCreateRequest{Name: "Counter", Pid: 0, DvType: "dashboard", NodeType: "leaf"}, "tester")
	require.NoError(t, err)

	require.NoError(t, svc.IncrementUseCount(created.ID))
	loaded, loadErr := svc.GetTemplate(created.ID)
	require.NoError(t, loadErr)
	require.NotNil(t, loaded)
	assert.Equal(t, 1, loaded.UseCount)

	t.Run("repo error", func(t *testing.T) {
		svc, db := setupTemplateServiceRepoTest(t)
		created := createTemplateFixture(t, svc, "Counter Error")
		require.NoError(t, db.Exec("CREATE TRIGGER deny_template_increment BEFORE UPDATE ON core_visualization_template BEGIN SELECT RAISE(FAIL, 'deny increment'); END;").Error)

		err := svc.IncrementUseCount(created.ID)
		require.Error(t, err)
	})
}
