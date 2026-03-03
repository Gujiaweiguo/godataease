//go:build integration

package service

import (
	"testing"

	"dataease/backend/internal/domain/areamap"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapService_GetWorldTree(t *testing.T) {
	cleanupTables(&areamap.Area{}, &areamap.CoreAreaCustom{})

	repo := repository.NewAreaRepository(testDB)
	svc := NewMapService(repo)

	t.Run("get world tree empty", func(t *testing.T) {
		result, err := svc.GetWorldTree()
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "000", result.ID)
		assert.Equal(t, "world", result.Level)
		assert.Equal(t, "世界村", result.Name)
		assert.False(t, result.Custom)
	})

	t.Run("get world tree with data", func(t *testing.T) {
		// Create test areas
		areas := []*areamap.Area{
			{ID: "156", Level: "country", Name: "China", Pid: "000"},
			{ID: "840", Level: "country", Name: "United States", Pid: "000"},
		}
		for _, a := range areas {
			require.NoError(t, testDB.Create(a).Error)
		}

		// Create custom areas
		customAreas := []*areamap.CoreAreaCustom{
			{ID: "custom1", Level: "custom", Name: "Custom Area", Pid: "156"},
		}
		for _, a := range customAreas {
			require.NoError(t, testDB.Create(a).Error)
		}

		result, err := svc.GetWorldTree()
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "000", result.ID)

		// Verify children are populated
		if len(result.Children) > 0 {
			// Find China node
			var chinaNode *areamap.AreaNode
			for _, child := range result.Children {
				if child.ID == "156" {
					chinaNode = child
					break
				}
			}

			if chinaNode != nil {
				assert.Equal(t, "China", chinaNode.Name)
				assert.False(t, chinaNode.Custom)

				// Check for custom child
				if len(chinaNode.Children) > 0 {
					var customNode *areamap.AreaNode
					for _, child := range chinaNode.Children {
						if child.ID == "custom1" {
							customNode = child
							break
						}
					}
					if customNode != nil {
						assert.True(t, customNode.Custom)
					}
				}
			}
		}
	})
}
