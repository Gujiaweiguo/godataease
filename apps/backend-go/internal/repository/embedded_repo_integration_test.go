//go:build integration
// +build integration

package repository

import (
	"fmt"
	"testing"

	"dataease/backend/internal/domain/embedded"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestEmbeddedRepository_CRUD(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("core_embedded")

	repo := NewEmbeddedRepository(testDB)
	entity := &embedded.CoreEmbedded{
		Name:         "Embedded CRUD",
		AppId:        "embedded-app-crud",
		AppSecret:    "secret-crud",
		Domain:       "https://crud.example.com",
		SecretLength: 24,
		CreateTime:   1710001001,
		UpdateBy:     "creator",
		UpdateTime:   1710001001,
	}

	require.NoError(t, repo.Create(entity))
	assert.NotZero(t, entity.ID)

	gotByID, err := repo.GetByID(entity.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.AppId, gotByID.AppId)
	assert.Equal(t, entity.Domain, gotByID.Domain)

	gotByAppID, err := repo.GetByAppId(entity.AppId)
	require.NoError(t, err)
	assert.Equal(t, entity.ID, gotByAppID.ID)

	entity.Name = "Embedded CRUD Updated"
	entity.Domain = "https://updated.example.com"
	entity.SecretLength = 32
	entity.UpdateBy = "updater"
	entity.UpdateTime = 1710001002
	require.NoError(t, repo.Update(entity))

	updated, err := repo.GetByID(entity.ID)
	require.NoError(t, err)
	assert.Equal(t, "Embedded CRUD Updated", updated.Name)
	assert.Equal(t, "https://updated.example.com", updated.Domain)
	assert.Equal(t, 32, updated.SecretLength)
	assert.Equal(t, "updater", updated.UpdateBy)

	require.NoError(t, repo.Delete(entity.ID))

	_, err = repo.GetByID(entity.ID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	_, err = repo.GetByAppId(entity.AppId)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	_, err = repo.GetByID(999999)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	_, err = repo.GetByAppId("missing-app-id")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestEmbeddedRepository_QueryDeleteBatchAndDomains(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("core_embedded")

	repo := NewEmbeddedRepository(testDB)
	records := make([]*embedded.CoreEmbedded, 0, 12)
	for i := 1; i <= 12; i++ {
		name := fmt.Sprintf("Portal %02d", i)
		if i <= 3 {
			name = fmt.Sprintf("Sales Portal %02d", i)
		}
		domain := ""
		switch i {
		case 1, 3:
			domain = "https://shared.example.com"
		case 2:
			domain = "https://unique.example.com"
		case 4:
			domain = ""
		}
		record := &embedded.CoreEmbedded{
			Name:         name,
			AppId:        fmt.Sprintf("embedded-app-%02d", i),
			AppSecret:    fmt.Sprintf("secret-%02d", i),
			Domain:       domain,
			SecretLength: 16 + i,
			CreateTime:   1710002000 + int64(i),
			UpdateBy:     "tester",
			UpdateTime:   1710002000 + int64(i),
		}
		require.NoError(t, repo.Create(record))
		records = append(records, record)
	}

	filtered, totalFiltered, err := repo.Query("Sales", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(3), totalFiltered)
	require.Len(t, filtered, 3)
	assert.Equal(t, records[2].ID, filtered[0].ID)
	assert.Equal(t, records[1].ID, filtered[1].ID)
	assert.Equal(t, records[0].ID, filtered[2].ID)

	firstPage, totalAll, err := repo.Query("", 0, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(12), totalAll)
	require.Len(t, firstPage, 10)
	assert.Equal(t, records[11].ID, firstPage[0].ID)
	assert.Equal(t, records[2].ID, firstPage[9].ID)

	secondPage, secondTotal, err := repo.Query("", 2, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(12), secondTotal)
	require.Len(t, secondPage, 2)
	assert.Equal(t, records[9].ID, secondPage[0].ID)
	assert.Equal(t, records[8].ID, secondPage[1].ID)

	domains, err := repo.ListDistinctDomains()
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"https://shared.example.com", "https://unique.example.com"}, domains)

	require.NoError(t, repo.DeleteBatch([]int64{records[0].ID, records[1].ID}))
	err = repo.DeleteBatch([]int64{})
	assert.ErrorContains(t, err, "WHERE conditions required")

	_, err = repo.GetByID(records[0].ID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	_, err = repo.GetByID(records[1].ID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	remaining, err := repo.GetByID(records[2].ID)
	require.NoError(t, err)
	assert.Equal(t, records[2].ID, remaining.ID)
}
