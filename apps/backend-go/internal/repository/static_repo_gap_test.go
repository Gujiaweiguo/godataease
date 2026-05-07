package repository

import (
	"testing"

	"dataease/backend/internal/domain/auto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestStaticRepository_Constructors(t *testing.T) {
	_, _, _, db := setupStaticRepoTest(t)

	staticRepo := NewStaticRepository(db)
	storeRepo := NewStoreRepository(db)
	typefaceRepo := NewTypefaceRepository(db)

	require.NotNil(t, staticRepo)
	require.NotNil(t, storeRepo)
	require.NotNil(t, typefaceRepo)
	assert.IsType(t, &gorm.DB{}, staticRepo.db)
	assert.Same(t, db, staticRepo.db)
	assert.Same(t, db, storeRepo.db)
	assert.Same(t, db, typefaceRepo.db)
}

func TestTypefaceRepository_ListFonts_WithDataAndDefaults(t *testing.T) {
	_, _, repo, _ := setupStaticRepoTest(t)

	font1 := &auto.CoreFont{Name: "FontA", FileName: "a.ttf", IsDefault: true, UpdateTime: 100}
	font2 := &auto.CoreFont{Name: "FontB", FileName: "b.ttf", IsDefault: false, UpdateTime: 200}
	require.NoError(t, repo.CreateFont(font1))
	require.NoError(t, repo.CreateFont(font2))

	fonts, err := repo.ListFonts()
	require.NoError(t, err)
	require.Len(t, fonts, 2)
	assert.Equal(t, "FontA", fonts[0].Name)

	defaults, err := repo.ListDefaultFonts()
	require.NoError(t, err)
	require.Len(t, defaults, 1)
	assert.Equal(t, font1.ID, defaults[0].ID)

	require.NoError(t, repo.SetDefaultFont(font1.ID, false))
	defaults, err = repo.ListDefaultFonts()
	require.NoError(t, err)
	assert.Empty(t, defaults)
	assert.NoError(t, repo.ClearDefaultFonts(font1.ID))
}

func TestTypefaceRepository_FindAndDeleteFont_NotFound(t *testing.T) {
	_, _, repo, _ := setupStaticRepoTest(t)

	font, err := repo.FindFontByName("missing-font")
	require.Error(t, err)
	assert.NotNil(t, font)
	assert.Zero(t, font.ID)

	require.NoError(t, repo.DeleteFont(999))
	font, err = repo.GetFontByID(999)
	require.Error(t, err)
	assert.NotNil(t, font)
	assert.Zero(t, font.ID)
}
