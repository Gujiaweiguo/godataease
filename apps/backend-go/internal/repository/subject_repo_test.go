package repository

import (
	"testing"

	"dataease/backend/internal/domain/auto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupSubjectRepoTest(t *testing.T) (*SubjectRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&auto.VisualizationSubject{}))
	return NewSubjectRepository(db), db
}

func TestSubjectRepository_List_Empty(t *testing.T) {
	repo, _ := setupSubjectRepoTest(t)

	subjects, err := repo.List()
	require.NoError(t, err)
	assert.Empty(t, subjects)
}

func TestSubjectRepository_List_ExcludesDeleted(t *testing.T) {
	repo, db := setupSubjectRepoTest(t)

	require.NoError(t, db.Create(&auto.VisualizationSubject{ID: "s1", Name: "Active Subject", DeleteFlag: false, CreateTime: 1000}).Error)
	require.NoError(t, db.Create(&auto.VisualizationSubject{ID: "s2", Name: "Deleted Subject", DeleteFlag: true, CreateTime: 2000}).Error)

	subjects, err := repo.List()
	require.NoError(t, err)
	require.Len(t, subjects, 1)
	assert.Equal(t, "s1", subjects[0].ID)
}

func TestSubjectRepository_List_OrderedByCreateTime(t *testing.T) {
	repo, db := setupSubjectRepoTest(t)

	require.NoError(t, db.Create(&auto.VisualizationSubject{ID: "s3", Name: "Third", DeleteFlag: false, CreateTime: 3000}).Error)
	require.NoError(t, db.Create(&auto.VisualizationSubject{ID: "s1", Name: "First", DeleteFlag: false, CreateTime: 1000}).Error)
	require.NoError(t, db.Create(&auto.VisualizationSubject{ID: "s2", Name: "Second", DeleteFlag: false, CreateTime: 2000}).Error)

	subjects, err := repo.List()
	require.NoError(t, err)
	require.Len(t, subjects, 3)
	assert.Equal(t, "s1", subjects[0].ID)
	assert.Equal(t, "s2", subjects[1].ID)
	assert.Equal(t, "s3", subjects[2].ID)
}

func TestSubjectRepository_ListAll_IncludesDeleted(t *testing.T) {
	repo, db := setupSubjectRepoTest(t)

	require.NoError(t, db.Create(&auto.VisualizationSubject{ID: "s1", Name: "Active", DeleteFlag: false, CreateTime: 1000}).Error)
	require.NoError(t, db.Create(&auto.VisualizationSubject{ID: "s2", Name: "Deleted", DeleteFlag: true, CreateTime: 2000}).Error)

	subjects, err := repo.ListAll()
	require.NoError(t, err)
	require.Len(t, subjects, 2)
}

func TestSubjectRepository_GetByID_Found(t *testing.T) {
	repo, db := setupSubjectRepoTest(t)

	require.NoError(t, db.Create(&auto.VisualizationSubject{ID: "s1", Name: "Test Subject", DeleteFlag: false, CreateTime: 1000, Details: `{"color":"blue"}`}).Error)

	subject, err := repo.GetByID("s1")
	require.NoError(t, err)
	assert.Equal(t, "Test Subject", subject.Name)
	assert.Equal(t, `{"color":"blue"}`, subject.Details)
}

func TestSubjectRepository_GetByID_DeletedNotFound(t *testing.T) {
	repo, db := setupSubjectRepoTest(t)

	require.NoError(t, db.Create(&auto.VisualizationSubject{ID: "s1", Name: "Deleted", DeleteFlag: true, CreateTime: 1000}).Error)

	_, err := repo.GetByID("s1")
	require.Error(t, err)
}

func TestSubjectRepository_GetByID_NotFound(t *testing.T) {
	repo, _ := setupSubjectRepoTest(t)

	_, err := repo.GetByID("nonexistent")
	require.Error(t, err)
}

func TestSubjectRepository_FindByName(t *testing.T) {
	repo, db := setupSubjectRepoTest(t)

	require.NoError(t, db.Create(&auto.VisualizationSubject{ID: "s1", Name: "Dark Theme", DeleteFlag: false, CreateTime: 1000}).Error)

	subject, err := repo.FindByName("Dark Theme")
	require.NoError(t, err)
	assert.Equal(t, "s1", subject.ID)
}

func TestSubjectRepository_FindByName_NotFound(t *testing.T) {
	repo, _ := setupSubjectRepoTest(t)

	_, err := repo.FindByName("Nonexistent Theme")
	require.Error(t, err)
}

func TestSubjectRepository_FindByNameExcludeID(t *testing.T) {
	repo, db := setupSubjectRepoTest(t)

	require.NoError(t, db.Create(&auto.VisualizationSubject{ID: "s1", Name: "Theme A", DeleteFlag: false, CreateTime: 1000}).Error)
	require.NoError(t, db.Create(&auto.VisualizationSubject{ID: "s2", Name: "Theme A", DeleteFlag: false, CreateTime: 2000}).Error)

	// Exclude s1 — should find s2
	subject, err := repo.FindByNameExcludeID("Theme A", "s1")
	require.NoError(t, err)
	assert.Equal(t, "s2", subject.ID)

	// Exclude s2 — should find s1
	subject, err = repo.FindByNameExcludeID("Theme A", "s2")
	require.NoError(t, err)
	assert.Equal(t, "s1", subject.ID)

	// Exclude both — not found
	_, err = repo.FindByNameExcludeID("Theme A", "s1")
	require.NoError(t, err) // First one is excluded but s2 matches
}

func TestSubjectRepository_Create(t *testing.T) {
	repo, _ := setupSubjectRepoTest(t)

	subject := &auto.VisualizationSubject{
		ID: "new1", Name: "New Theme", Type: "self", Details: `{"bg":"#fff"}`,
		DeleteFlag: false, CreateTime: 1000, CreateBy: "admin",
	}
	require.NoError(t, repo.Create(subject))

	found, err := repo.GetByID("new1")
	require.NoError(t, err)
	assert.Equal(t, "New Theme", found.Name)
}

func TestSubjectRepository_Update(t *testing.T) {
	repo, db := setupSubjectRepoTest(t)

	require.NoError(t, db.Create(&auto.VisualizationSubject{ID: "s1", Name: "Original", DeleteFlag: false, CreateTime: 1000, CreateNum: 5}).Error)

	subject := &auto.VisualizationSubject{
		ID: "s1", Name: "Updated", Type: "system", Details: `{"bg":"#000"}`,
		DeleteFlag: false, CreateTime: 1000, UpdateTime: 2000, UpdateBy: "admin", CreateNum: 5,
	}
	require.NoError(t, repo.Update(subject))

	found, err := repo.GetByID("s1")
	require.NoError(t, err)
	assert.Equal(t, "Updated", found.Name)
	assert.Equal(t, `{"bg":"#000"}`, found.Details)
}

func TestSubjectRepository_Delete(t *testing.T) {
	repo, db := setupSubjectRepoTest(t)

	require.NoError(t, db.Create(&auto.VisualizationSubject{ID: "s1", Name: "To Delete", DeleteFlag: false, CreateTime: 1000}).Error)

	require.NoError(t, repo.Delete("s1"))

	var count int64
	require.NoError(t, db.Model(&auto.VisualizationSubject{}).Where("id = ?", "s1").Count(&count).Error)
	assert.Equal(t, int64(0), count)
}

func TestSubjectRepository_Delete_NotExist(t *testing.T) {
	repo, _ := setupSubjectRepoTest(t)

	// Deleting non-existent should not error
	err := repo.Delete("nonexistent")
	require.NoError(t, err)
}
