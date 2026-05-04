//go:build integration

package repository

import (
	"context"
	"testing"

	datafillingdomain "dataease/backend/internal/domain/datafilling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommitLogRepository_CRUDAndPagination(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("df_commit_log")

	repo := NewCommitLogRepository(testDB)
	ctx := context.Background()
	require.NoError(t, repo.Create(ctx, &datafillingdomain.DfCommitLog{FormID: 9, DataID: "a", Operate: 1, CommitBy: 1, Committer: "u1", CommitTime: 1, Count: 1}))
	require.NoError(t, repo.Create(ctx, &datafillingdomain.DfCommitLog{FormID: 9, DataID: "b", Operate: 2, CommitBy: 1, Committer: "u1", CommitTime: 2, Count: 1}))
	require.NoError(t, repo.Create(ctx, &datafillingdomain.DfCommitLog{FormID: 10, DataID: "c", Operate: 0, CommitBy: 1, Committer: "u1", CommitTime: 3, Count: 1}))

	rows, total, err := repo.ListByFormID(ctx, 9, 1, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, rows, 1)
	assert.Equal(t, "b", rows[0].DataID)

	require.NoError(t, repo.DeleteByFormID(ctx, 9))
	rows, total, err = repo.ListByFormID(ctx, 9, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, rows)
	rows, total, err = repo.ListByFormID(ctx, 10, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, rows, 1)
}
