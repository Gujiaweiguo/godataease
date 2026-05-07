package repository

import (
	"testing"
	"time"

	"dataease/backend/internal/domain/visualization"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVisualizationRepository_Gap2EmptyAndDefaultPaths(t *testing.T) {
	repo, db := setupVisualizationRepositoryTest(t)

	now := time.Now().UnixMilli()
	dashboardType := visualization.TypeDashboard
	require.NoError(t, db.Create([]*visualization.DataVisualizationInfo{
		newVisualizationInfo(901, "Alpha", dashboardType, now-10),
		newVisualizationInfo(902, "Beta", dashboardType, now),
	}).Error)

	list, total, err := repo.Query(&visualization.ListRequest{Current: -1, Size: 1000})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, list, 2)
	assert.Equal(t, int64(902), list[0].ID)

	recent, err := repo.FindRecent(404, &visualization.WorkbranchQueryRequest{Type: "unknown", Keyword: "nomatch"})
	require.NoError(t, err)
	assert.Empty(t, recent)

	mapping, err := repo.GetCopiedChartViewMapping(9999)
	require.NoError(t, err)
	assert.Empty(t, mapping)

	require.NoError(t, repo.CopyLinkages(9999))
	require.NoError(t, repo.CopyLinkageFields(9999))
	require.NoError(t, repo.CopyLinkJumps(9999))
	require.NoError(t, repo.CopyLinkJumpInfos(9999))
	require.NoError(t, repo.CopyLinkJumpTargetInfos(9999))

	linkages, err := repo.FindLinkagesByDvID(9999)
	require.NoError(t, err)
	assert.Empty(t, linkages)

	fields, err := repo.FindLinkageFieldsByDvID(9999)
	require.NoError(t, err)
	assert.Empty(t, fields)

	jumps, err := repo.FindLinkJumpsByDvID(9999)
	require.NoError(t, err)
	assert.Empty(t, jumps)

	infos, err := repo.FindLinkJumpInfosByDvID(9999)
	require.NoError(t, err)
	assert.Empty(t, infos)

	targets, err := repo.FindLinkJumpTargetViewInfosByDvID(9999)
	require.NoError(t, err)
	assert.Empty(t, targets)
}
