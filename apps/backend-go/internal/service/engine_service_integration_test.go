//go:build integration

package service

import (
	"testing"

	"dataease/backend/internal/domain/engine"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngineService_GetEngine(t *testing.T) {
	cleanupTables(&engine.Engine{})

	repo := repository.NewEngineRepository(testDB)
	svc := NewEngineService(repo)

	t.Run("get engine when empty", func(t *testing.T) {
		result, err := svc.GetEngine()
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("get engine with data", func(t *testing.T) {
		cfg := "test configuration"
		testEngine := &engine.Engine{
			Name:          "Test Engine",
			Type:          "doris",
			Configuration: &cfg,
		}
		require.NoError(t, testDB.Create(testEngine).Error)

		result, err := svc.GetEngine()
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "Test Engine", result.Name)
		assert.Equal(t, "doris", result.Type)
		assert.Equal(t, &cfg, result.Configuration)
	})
}

func TestEngineService_Validate(t *testing.T) {
	repo := repository.NewEngineRepository(testDB)
	svc := NewEngineService(repo)

	t.Run("validate always returns success", func(t *testing.T) {
		req := &engine.ValidateRequest{
			Type: strPtr("doris"),
		}

		result, err := svc.Validate(req)
		require.NoError(t, err)
		assert.Equal(t, "Success", result.Status)
		assert.Equal(t, "Engine validation not implemented", result.Message)
	})
}

func TestEngineService_ValidateByID(t *testing.T) {
	repo := repository.NewEngineRepository(testDB)
	svc := NewEngineService(repo)

	t.Run("validate by ID always returns success", func(t *testing.T) {
		result, err := svc.ValidateByID(1)
		require.NoError(t, err)
		assert.Equal(t, "Success", result.Status)
		assert.Equal(t, "Engine validation not implemented", result.Message)
	})
}

func TestEngineService_SupportSetKey(t *testing.T) {
	repo := repository.NewEngineRepository(testDB)
	svc := NewEngineService(repo)

	t.Run("support set key returns false", func(t *testing.T) {
		result, err := svc.SupportSetKey()
		require.NoError(t, err)
		assert.False(t, result)
	})
}
