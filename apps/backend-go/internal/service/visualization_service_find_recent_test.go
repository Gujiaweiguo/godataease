package service

import (
	"testing"

	"dataease/backend/internal/domain/visualization"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupVisualizationRecentTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&visualization.DataVisualizationInfo{}))

	statements := []string{
		`CREATE TABLE core_dataset_group (id INTEGER PRIMARY KEY, name TEXT, node_type TEXT, create_by TEXT)`,
		`CREATE TABLE core_datasource (id INTEGER PRIMARY KEY, name TEXT, type TEXT, create_by TEXT)`,
		`CREATE TABLE core_opt_recent (uid INTEGER, resource_id INTEGER, time INTEGER)`,
		`CREATE TABLE core_store (uid INTEGER, resource_id INTEGER)`,
	}
	for _, statement := range statements {
		require.NoError(t, db.Exec(statement).Error)
	}

	t.Cleanup(func() {
		sqlDB, dbErr := db.DB()
		require.NoError(t, dbErr)
		require.NoError(t, sqlDB.Close())
	})

	return db
}

func seedVisualizationRecentFixtures(t *testing.T, db *gorm.DB, uid int64) {
	t.Helper()

	panelType := "dashboard"
	screenType := "dataV"
	leafType := "leaf"
	status := 1
	mobile := true
	creator := "creator"

	require.NoError(t, db.Create(&visualization.DataVisualizationInfo{ID: 1001, Name: "Sales Panel", Type: &panelType, NodeType: &leafType, Status: &status, MobileLayout: &mobile, CreateBy: &creator}).Error)
	require.NoError(t, db.Create(&visualization.DataVisualizationInfo{ID: 1002, Name: "Ops Screen", Type: &screenType, NodeType: &leafType, Status: &status, CreateBy: &creator}).Error)
	require.NoError(t, db.Exec(`INSERT INTO core_dataset_group (id, name, node_type, create_by) VALUES (2001, 'Orders Dataset', 'dataset', 'dataset_creator')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO core_datasource (id, name, type, create_by) VALUES (3001, 'Main Datasource', 'mysql', 'datasource_creator')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO core_opt_recent (uid, resource_id, time) VALUES (?, 1001, 4001), (?, 1002, 4002), (?, 2001, 4003), (?, 3001, 4004)`, uid, uid, uid, uid).Error)
	require.NoError(t, db.Exec(`INSERT INTO core_store (uid, resource_id) VALUES (?, 1001)`, uid).Error)
}

func TestVisualizationService_FindRecent(t *testing.T) {
	t.Parallel()

	t.Run("returns recent items across supported resource types", func(t *testing.T) {
		t.Parallel()

		db := setupVisualizationRecentTestDB(t)
		seedVisualizationRecentFixtures(t, db, 7)
		svc := NewVisualizationService(repository.NewVisualizationRepository(db))

		tests := []struct {
			name string
			typ  string
			want string
		}{
			{name: "panel", typ: "panel", want: "Sales Panel"},
			{name: "screen", typ: "screen", want: "Ops Screen"},
			{name: "dataset", typ: "dataset", want: "Orders Dataset"},
			{name: "datasource", typ: "datasource", want: "Main Datasource"},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				items, err := svc.FindRecent(&visualization.WorkbranchQueryRequest{Type: tt.typ}, 7)
				require.NoError(t, err)
				require.Len(t, items, 1)
				assert.Equal(t, tt.typ, items[0].Type)
				assert.Equal(t, tt.want, items[0].Name)
			})
		}
	})

	t.Run("returns empty when no recent item matches", func(t *testing.T) {
		t.Parallel()

		db := setupVisualizationRecentTestDB(t)
		seedVisualizationRecentFixtures(t, db, 7)
		svc := NewVisualizationService(repository.NewVisualizationRepository(db))

		items, err := svc.FindRecent(&visualization.WorkbranchQueryRequest{Type: "panel"}, 99)
		require.NoError(t, err)
		assert.Empty(t, items)
	})

	t.Run("propagates repository errors", func(t *testing.T) {
		t.Parallel()

		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		t.Cleanup(func() {
			sqlDB, dbErr := db.DB()
			require.NoError(t, dbErr)
			require.NoError(t, sqlDB.Close())
		})

		svc := NewVisualizationService(repository.NewVisualizationRepository(db))

		items, svcErr := svc.FindRecent(&visualization.WorkbranchQueryRequest{Type: "panel"}, 7)
		require.Error(t, svcErr)
		assert.Nil(t, items)
	})
}
