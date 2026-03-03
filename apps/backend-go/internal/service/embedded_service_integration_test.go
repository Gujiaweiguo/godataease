//go:build integration

package service

import (
	"testing"

	"dataease/backend/internal/domain/embedded"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmbeddedService_Create(t *testing.T) {
	cleanupTables(&embedded.CoreEmbedded{})

	repo := repository.NewEmbeddedRepository(testDB)
	svc := NewEmbeddedService(repo)

	t.Run("create embedded app with defaults", func(t *testing.T) {
		req := &embedded.EmbeddedCreator{
			Name:   "Test App",
			Domain: "http://localhost:8080",
		}

		id, err := svc.Create(req, "tester")
		require.NoError(t, err)
		assert.Greater(t, id, int64(0))
	})

	t.Run("create embedded app with custom secret length", func(t *testing.T) {
		secretLength := 32
		req := &embedded.EmbeddedCreator{
			Name:         "Custom App",
			Domain:       "http://example.com",
			SecretLength: &secretLength,
		}

		id, err := svc.Create(req, "tester")
		require.NoError(t, err)
		assert.Greater(t, id, int64(0))
	})
}

func TestEmbeddedService_Edit(t *testing.T) {
	cleanupTables(&embedded.CoreEmbedded{})

	repo := repository.NewEmbeddedRepository(testDB)
	svc := NewEmbeddedService(repo)

	// Create test app
	req := &embedded.EmbeddedCreator{
		Name:   "Original App",
		Domain: "http://original.com",
	}
	id, err := svc.Create(req, "tester")
	require.NoError(t, err)

	t.Run("edit name", func(t *testing.T) {
		editReq := &embedded.EmbeddedEditor{
			ID:   id,
			Name: "Updated App",
		}
		err := svc.Edit(editReq, "updater")
		require.NoError(t, err)

		// Verify update
		result, err := repo.GetByID(id)
		require.NoError(t, err)
		assert.Equal(t, "Updated App", result.Name)
	})

	t.Run("edit domain", func(t *testing.T) {
		newDomain := "http://updated.com"
		editReq := &embedded.EmbeddedEditor{
			ID:     id,
			Domain: &newDomain,
		}
		err := svc.Edit(editReq, "updater")
		require.NoError(t, err)
	})

	t.Run("edit non-existent app", func(t *testing.T) {
		editReq := &embedded.EmbeddedEditor{
			ID:   999999,
			Name: "Should Fail",
		}
		err := svc.Edit(editReq, "updater")
		assert.Error(t, err)
	})
}

func TestEmbeddedService_Delete(t *testing.T) {
	cleanupTables(&embedded.CoreEmbedded{})

	repo := repository.NewEmbeddedRepository(testDB)
	svc := NewEmbeddedService(repo)

	t.Run("delete existing app", func(t *testing.T) {
		req := &embedded.EmbeddedCreator{
			Name:   "To Delete",
			Domain: "http://delete.com",
		}
		id, err := svc.Create(req, "tester")
		require.NoError(t, err)

		err = svc.Delete(id)
		require.NoError(t, err)

		// Verify deleted
		_, err = repo.GetByID(id)
		assert.Error(t, err)
	})

	t.Run("delete non-existent app", func(t *testing.T) {
		// GORM Delete does not return error for non-existent records
		err := svc.Delete(999999)
		assert.NoError(t, err)
	})
}

func TestEmbeddedService_BatchDelete(t *testing.T) {
	cleanupTables(&embedded.CoreEmbedded{})

	repo := repository.NewEmbeddedRepository(testDB)
	svc := NewEmbeddedService(repo)

	t.Run("batch delete apps", func(t *testing.T) {
		// Create multiple apps
		var ids []int64
		for i := 0; i < 3; i++ {
			req := &embedded.EmbeddedCreator{
				Name:   "Batch App",
				Domain: "http://batch.com",
			}
			id, err := svc.Create(req, "tester")
			require.NoError(t, err)
			ids = append(ids, id)
		}

		err := svc.BatchDelete(ids)
		require.NoError(t, err)
	})

	t.Run("batch delete empty list", func(t *testing.T) {
		err := svc.BatchDelete([]int64{})
		assert.Error(t, err)
	})
}

func TestEmbeddedService_ResetSecret(t *testing.T) {
	cleanupTables(&embedded.CoreEmbedded{})

	repo := repository.NewEmbeddedRepository(testDB)
	svc := NewEmbeddedService(repo)

	// Create test app
	req := &embedded.EmbeddedCreator{
		Name:   "Reset Test",
		Domain: "http://reset.com",
	}
	id, err := svc.Create(req, "tester")
	require.NoError(t, err)

	t.Run("reset secret auto generate", func(t *testing.T) {
		resetReq := &embedded.EmbeddedResetRequest{
			ID: id,
		}
		err := svc.ResetSecret(resetReq, "updater")
		require.NoError(t, err)
	})

	t.Run("reset secret with custom value", func(t *testing.T) {
		customSecret := "my-custom-secret-12345"
		resetReq := &embedded.EmbeddedResetRequest{
			ID:        id,
			AppSecret: &customSecret,
		}
		err := svc.ResetSecret(resetReq, "updater")
		require.NoError(t, err)

		// Verify secret
		result, err := repo.GetByID(id)
		require.NoError(t, err)
		assert.Equal(t, customSecret, result.AppSecret)
	})

	t.Run("reset non-existent app", func(t *testing.T) {
		resetReq := &embedded.EmbeddedResetRequest{
			ID: 999999,
		}
		err := svc.ResetSecret(resetReq, "updater")
		assert.Error(t, err)
	})
}

func TestEmbeddedService_QueryGrid(t *testing.T) {
	cleanupTables(&embedded.CoreEmbedded{})

	repo := repository.NewEmbeddedRepository(testDB)
	svc := NewEmbeddedService(repo)

	// Create multiple apps
	for i := 0; i < 5; i++ {
		req := &embedded.EmbeddedCreator{
			Name:   "Query App",
			Domain: "http://query.com",
		}
		_, err := svc.Create(req, "tester")
		require.NoError(t, err)
	}

	t.Run("query all", func(t *testing.T) {
		result, err := svc.QueryGrid("", 1, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, result.Total, int64(5))
	})

	t.Run("query with keyword", func(t *testing.T) {
		result, err := svc.QueryGrid("Query", 1, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, result.Total, int64(1))
	})

	t.Run("query pagination", func(t *testing.T) {
		result, err := svc.QueryGrid("", 2, 2)
		require.NoError(t, err)
		assert.Equal(t, 2, result.Current)
		assert.Equal(t, 2, result.Size)
	})
}

func TestEmbeddedService_GetDomainList(t *testing.T) {
	cleanupTables(&embedded.CoreEmbedded{})

	repo := repository.NewEmbeddedRepository(testDB)
	svc := NewEmbeddedService(repo)

	t.Run("get domain list empty", func(t *testing.T) {
		result, err := svc.GetDomainList()
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("get domain list with data", func(t *testing.T) {
		req := &embedded.EmbeddedCreator{
			Name:   "Domain App",
			Domain: "http://domain1.com,http://domain2.com",
		}
		_, err := svc.Create(req, "tester")
		require.NoError(t, err)

		result, err := svc.GetDomainList()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 1)
	})
}

func TestEmbeddedService_GetTokenArgs(t *testing.T) {
	repo := repository.NewEmbeddedRepository(testDB)
	svc := NewEmbeddedService(repo)

	t.Run("get token args", func(t *testing.T) {
		result := svc.GetTokenArgs(123, 456)
		assert.Equal(t, int64(123), result.UserId)
		assert.Equal(t, int64(456), result.OrgId)
	})
}

func TestEmbeddedService_GetLimitCount(t *testing.T) {
	repo := repository.NewEmbeddedRepository(testDB)
	svc := NewEmbeddedService(repo)

	t.Run("get limit count", func(t *testing.T) {
		result := svc.GetLimitCount()
		assert.Equal(t, 5, result)
	})
}

func TestEmbeddedService_InitIframe(t *testing.T) {
	cleanupTables(&embedded.CoreEmbedded{})

	repo := repository.NewEmbeddedRepository(testDB)
	svc := NewEmbeddedService(repo)

	// Create test app
	req := &embedded.EmbeddedCreator{
		Name:   "Iframe Test",
		Domain: "http://localhost:8080,http://example.com",
	}
	_, err := svc.Create(req, "tester")
	require.NoError(t, err)

	// Get the created app to find AppId
	apps, _, err := repo.Query("", 1, 10)
	require.NoError(t, err)
	require.Greater(t, len(apps), 0)

	appId := apps[0].AppId

	t.Run("init iframe with empty token", func(t *testing.T) {
		result, err := svc.InitIframe("", "http://localhost:8080")
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("init iframe with invalid token", func(t *testing.T) {
		result, err := svc.InitIframe("invalid-token", "http://localhost:8080")
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("init iframe with valid token but wrong origin", func(t *testing.T) {
		// Create a mock token with appId claim
		token := "header." + `{"appId":"` + appId + `"}` + ".signature"
		result, err := svc.InitIframe(token, "http://unauthorized.com")
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestEmbeddedSplitToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected int
	}{
		{"empty token", "", 0},
		{"no dots", "nodots", 0},
		{"one dot", "one.two", 1},
		{"valid JWT format", "header.payload.signature", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitToken(tt.token)
			assert.GreaterOrEqual(t, len(result), tt.expected)
		})
	}
}

func TestEmbeddedParseClaim(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectKey string
		expectVal string
	}{
		{"empty", "", "", ""},
		{"with colon", `"appId":"test123"`, "appId", "test123"},
		{"with equals", "userId=456", "userId", "456"},
		{"no separator", "noseparator", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, val := parseClaim(tt.input)
			assert.Equal(t, tt.expectKey, key)
			assert.Equal(t, tt.expectVal, val)
		})
	}
}

func TestEmbeddedTrimQuotes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", ""},
		{"double quotes", `"test"`, "test"},
		{"single quotes", `'test'`, "test"},
		{"no quotes", "test", "test"},
		{"mixed quotes", `"'test'"`, "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := trimQuotes(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEmbeddedDecodeBase64(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", ""},
		{"no special chars", "abc123", "abc123"},
		{"url safe minus", "a-b_c", "a+b/c"},
		{"url safe underscore", "test_value", "test/value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decodeBase64(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEmbeddedSplitBy(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		sep      rune
		expected []string
	}{
		{"empty", "", ',', nil},
		{"single item", "test", ',', []string{"test"}},
		{"comma separated", "a,b,c", ',', []string{"a", "b", "c"}},
		{"dot separated", "a.b.c", '.', []string{"a", "b", "c"}},
		{"consecutive separators", "a,,b", ',', []string{"a", "b"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitBy(tt.input, tt.sep)
			assert.Equal(t, tt.expected, result)
		})
	}
}
