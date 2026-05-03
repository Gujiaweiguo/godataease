//go:build integration

package repository

import (
	"context"
	"testing"

	datafillingdomain "dataease/backend/internal/domain/datafilling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestDataFillingRepository_CRUDAndTree(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("data_filling_forms")

	repo := NewDataFillingRepository(testDB)
	ctx := context.Background()

	folder := &datafillingdomain.DataFillingForm{Name: "folder", PID: 0, Level: 0, NodeType: datafillingdomain.NodeTypeFolder, CreateBy: 1, CreateTime: 1, UpdateBy: 1, UpdateTime: 1}
	require.NoError(t, repo.Create(ctx, folder))
	form := &datafillingdomain.DataFillingForm{Name: "form", PID: folder.ID, Level: 1, NodeType: datafillingdomain.NodeTypeForm, PhysicalTableName: "df_repo_test", DatasourceID: 1, Forms: "[]", CreateBy: 1, CreateTime: 1, UpdateBy: 1, UpdateTime: 1}
	require.NoError(t, repo.Create(ctx, form))

	got, err := repo.GetByID(ctx, form.ID)
	require.NoError(t, err)
	assert.Equal(t, "form", got.Name)

	got.Name = "form-updated"
	require.NoError(t, repo.Update(ctx, got))

	require.NoError(t, repo.Rename(ctx, form.ID, "form-renamed"))
	require.NoError(t, repo.Move(ctx, form.ID, 0))

	moved, err := repo.GetByID(ctx, form.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), moved.PID)
	assert.Equal(t, 0, moved.Level)

	children, err := repo.GetChildren(ctx, 0)
	require.NoError(t, err)
	assert.Len(t, children, 2)

	tree, err := repo.GetTree(ctx)
	require.NoError(t, err)
	assert.Len(t, tree, 2)
	assert.Equal(t, datafillingdomain.NodeTypeFolder, tree[0].NodeType)

	require.NoError(t, repo.DeleteByID(ctx, folder.ID))
	_, err = repo.GetByID(ctx, folder.ID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
