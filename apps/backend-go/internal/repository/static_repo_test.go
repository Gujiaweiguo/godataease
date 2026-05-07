package repository

import (
	"testing"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/domain/static"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupStaticRepoTest(t *testing.T) (*StaticRepository, *StoreRepository, *TypefaceRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&static.StaticResource{}, &static.Store{}, &static.Typeface{}, &auto.CoreFont{}))
	return NewStaticRepository(db), NewStoreRepository(db), NewTypefaceRepository(db), db
}

// --- StaticRepository ---

func TestStaticRepository_ListResources_Empty(t *testing.T) {
	repo, _, _, _ := setupStaticRepoTest(t)

	resources, err := repo.ListResources()
	require.NoError(t, err)
	assert.Empty(t, resources)
}

func TestStaticRepository_ListResources_WithData(t *testing.T) {
	repo, _, _, db := setupStaticRepoTest(t)

	require.NoError(t, db.Create(&static.StaticResource{ID: "r1", Name: "Icon Set", Path: "/icons", Type: "icon"}).Error)
	require.NoError(t, db.Create(&static.StaticResource{ID: "r2", Name: "Image Pack", Path: "/images", Type: "image"}).Error)

	resources, err := repo.ListResources()
	require.NoError(t, err)
	require.Len(t, resources, 2)
}

func TestStaticRepository_GetResourceByID_Found(t *testing.T) {
	repo, _, _, db := setupStaticRepoTest(t)

	require.NoError(t, db.Create(&static.StaticResource{ID: "r1", Name: "Icon Set", Path: "/icons", Type: "icon"}).Error)

	resource, err := repo.GetResourceByID("r1")
	require.NoError(t, err)
	assert.Equal(t, "Icon Set", resource.Name)
	assert.Equal(t, "/icons", resource.Path)
}

func TestStaticRepository_GetResourceByID_NotFound(t *testing.T) {
	repo, _, _, _ := setupStaticRepoTest(t)

	_, err := repo.GetResourceByID("nonexistent")
	require.Error(t, err)
}

// --- StoreRepository ---

func TestStoreRepository_ListStores_Empty(t *testing.T) {
	_, repo, _, _ := setupStaticRepoTest(t)

	stores, err := repo.ListStores()
	require.NoError(t, err)
	assert.Empty(t, stores)
}

func TestStoreRepository_ListStores_WithData(t *testing.T) {
	_, repo, _, db := setupStaticRepoTest(t)

	require.NoError(t, db.Create(&static.Store{ID: "s1", Name: "Store A", URL: "http://example.com/a"}).Error)
	require.NoError(t, db.Create(&static.Store{ID: "s2", Name: "Store B", URL: "http://example.com/b"}).Error)

	stores, err := repo.ListStores()
	require.NoError(t, err)
	require.Len(t, stores, 2)
}

// --- TypefaceRepository ---

func TestTypefaceRepository_ListTypefaces_Empty(t *testing.T) {
	_, _, repo, _ := setupStaticRepoTest(t)

	typefaces, err := repo.ListTypefaces()
	require.NoError(t, err)
	assert.Empty(t, typefaces)
}

func TestTypefaceRepository_ListTypefaces_WithData(t *testing.T) {
	_, _, repo, db := setupStaticRepoTest(t)

	require.NoError(t, db.Create(&static.Typeface{ID: "tf1", Name: "Arial", File: "arial.ttf"}).Error)

	typefaces, err := repo.ListTypefaces()
	require.NoError(t, err)
	require.Len(t, typefaces, 1)
	assert.Equal(t, "Arial", typefaces[0].Name)
}

func TestTypefaceRepository_ListFonts_Empty(t *testing.T) {
	_, _, repo, _ := setupStaticRepoTest(t)

	fonts, err := repo.ListFonts()
	require.NoError(t, err)
	assert.Empty(t, fonts)
}

func TestTypefaceRepository_FontCRUD(t *testing.T) {
	_, _, repo, _ := setupStaticRepoTest(t)

	// Create
	font := &auto.CoreFont{
		Name: "TestFont", FileName: "test.ttf", FileTransName: "test_trans", IsDefault: false, UpdateTime: 1000,
	}
	require.NoError(t, repo.CreateFont(font))
	require.Positive(t, font.ID)

	// Get by ID
	found, err := repo.GetFontByID(font.ID)
	require.NoError(t, err)
	assert.Equal(t, "TestFont", found.Name)

	// Find by name
	found, err = repo.FindFontByName("TestFont")
	require.NoError(t, err)
	assert.Equal(t, font.ID, found.ID)

	// Find by name — not found
	_, err = repo.FindFontByName("NonExistent")
	require.Error(t, err)

	// Update
	found.Name = "UpdatedFont"
	require.NoError(t, repo.UpdateFont(found))
	updated, err := repo.GetFontByID(font.ID)
	require.NoError(t, err)
	assert.Equal(t, "UpdatedFont", updated.Name)

	// Delete
	require.NoError(t, repo.DeleteFont(font.ID))
	_, err = repo.GetFontByID(font.ID)
	require.Error(t, err)
}

func TestTypefaceRepository_SetDefaultAndClear(t *testing.T) {
	_, _, repo, _ := setupStaticRepoTest(t)

	font1 := &auto.CoreFont{Name: "Font1", FileName: "f1.ttf", IsDefault: false, UpdateTime: 1000}
	font2 := &auto.CoreFont{Name: "Font2", FileName: "f2.ttf", IsDefault: false, UpdateTime: 2000}
	require.NoError(t, repo.CreateFont(font1))
	require.NoError(t, repo.CreateFont(font2))

	// Set font1 as default
	require.NoError(t, repo.SetDefaultFont(font1.ID, true))

	defaults, err := repo.ListDefaultFonts()
	require.NoError(t, err)
	require.Len(t, defaults, 1)
	assert.Equal(t, font1.ID, defaults[0].ID)

	// Clear defaults excluding font1
	require.NoError(t, repo.ClearDefaultFonts(font1.ID))

	// Set font2 as default
	require.NoError(t, repo.SetDefaultFont(font2.ID, true))

	defaults, err = repo.ListDefaultFonts()
	require.NoError(t, err)
	require.Len(t, defaults, 2)
}

func TestTypefaceRepository_GetFontByID_NotFound(t *testing.T) {
	_, _, repo, _ := setupStaticRepoTest(t)

	_, err := repo.GetFontByID(999)
	require.Error(t, err)
}
