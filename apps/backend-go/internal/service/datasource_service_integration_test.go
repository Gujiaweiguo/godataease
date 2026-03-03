//go:build integration

package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	datasetdomain "dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/repository"
	seatunnelv1 "dataease/backend/proto/seatunnel/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestDatasourceService_Save(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	t.Run("create datasource successfully", func(t *testing.T) {
		cfg := &datasource.ConnectionConfig{
			Host:     "localhost",
			Port:     3306,
			Database: "test_db",
		}
		cfgJSON, _ := json.Marshal(cfg)
		cfgStr := base64.StdEncoding.EncodeToString(cfgJSON)

		req := &datasource.WriteRequest{
			Name:          "Test MySQL",
			Type:          "mysql",
			Configuration: &cfgStr,
		}

		result, err := svc.Save(req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Greater(t, result.ID, int64(0))
		assert.Equal(t, "Test MySQL", result.Name)
		assert.Equal(t, "mysql", result.Type)
	})

	t.Run("create folder successfully", func(t *testing.T) {
		req := &datasource.WriteRequest{
			Name:     "Test Folder",
			Type:     datasource.TypeFolder,
			NodeType: datasource.TypeFolder,
		}

		result, err := svc.Save(req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test Folder", result.Name)
		assert.Equal(t, datasource.TypeFolder, result.Type)
	})

	t.Run("create datasource with PID", func(t *testing.T) {
		// First create a folder
		folder, err := svc.CreateFolder("Parent Folder", 0)
		require.NoError(t, err)

		cfg := &datasource.ConnectionConfig{
			Host:     "localhost",
			Port:     3306,
			Database: "test_db2",
		}
		cfgJSON, _ := json.Marshal(cfg)
		cfgStr := base64.StdEncoding.EncodeToString(cfgJSON)

		req := &datasource.WriteRequest{
			Name:          "Child Datasource",
			PID:           &folder.ID,
			Type:          "mysql",
			Configuration: &cfgStr,
		}

		result, err := svc.Save(req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, folder.ID, *result.PID)
	})

	t.Run("fail with empty name", func(t *testing.T) {
		req := &datasource.WriteRequest{
			Type: "mysql",
		}

		result, err := svc.Save(req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("fail with duplicate name", func(t *testing.T) {
		cfg := &datasource.ConnectionConfig{
			Host:     "localhost",
			Port:     3306,
			Database: "test_db3",
		}
		cfgJSON, _ := json.Marshal(cfg)
		cfgStr := base64.StdEncoding.EncodeToString(cfgJSON)

		req1 := &datasource.WriteRequest{
			Name:          "Duplicate Name",
			Type:          "mysql",
			Configuration: &cfgStr,
		}
		_, err := svc.Save(req1)
		require.NoError(t, err)

		req2 := &datasource.WriteRequest{
			Name:          "Duplicate Name",
			Type:          "mysql",
			Configuration: &cfgStr,
		}
		result, err := svc.Save(req2)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "already exists")
	})
}

func TestDatasourceService_Update(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	cfg := &datasource.ConnectionConfig{
		Host:     "localhost",
		Port:     3306,
		Database: "test_db",
	}
	cfgJSON, _ := json.Marshal(cfg)
	cfgStr := base64.StdEncoding.EncodeToString(cfgJSON)

	created, err := svc.Save(&datasource.WriteRequest{
		Name:          "Original Name",
		Type:          "mysql",
		Configuration: &cfgStr,
	})
	require.NoError(t, err)

	t.Run("update name successfully", func(t *testing.T) {
		req := &datasource.WriteRequest{
			ID:   created.ID,
			Name: "Updated Name",
		}

		result, err := svc.Update(req)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", result.Name)
	})

	t.Run("update configuration", func(t *testing.T) {
		newCfg := &datasource.ConnectionConfig{
			Host:     "newhost",
			Port:     3307,
			Database: "new_db",
		}
		newCfgJSON, _ := json.Marshal(newCfg)
		newCfgStr := base64.StdEncoding.EncodeToString(newCfgJSON)

		req := &datasource.WriteRequest{
			ID:            created.ID,
			Configuration: &newCfgStr,
		}

		result, err := svc.Update(req)
		require.NoError(t, err)
		assert.NotNil(t, result.Configuration)
	})

	t.Run("fail with invalid ID", func(t *testing.T) {
		req := &datasource.WriteRequest{
			ID:   0,
			Name: "Test",
		}

		result, err := svc.Update(req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "id is required")
	})

	t.Run("fail with non-existent ID", func(t *testing.T) {
		req := &datasource.WriteRequest{
			ID:   999999,
			Name: "Test",
		}

		result, err := svc.Update(req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestDatasourceService_GetByID(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	cfg := &datasource.ConnectionConfig{
		Host:     "localhost",
		Port:     3306,
		Database: "test_db",
	}
	cfgJSON, _ := json.Marshal(cfg)
	cfgStr := base64.StdEncoding.EncodeToString(cfgJSON)

	created, err := svc.Save(&datasource.WriteRequest{
		Name:          "Test DS",
		Type:          "mysql",
		Configuration: &cfgStr,
	})
	require.NoError(t, err)

	t.Run("get by valid ID", func(t *testing.T) {
		result, err := svc.GetByID(created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, result.ID)
		assert.Equal(t, "Test DS", result.Name)
	})

	t.Run("fail with non-existent ID", func(t *testing.T) {
		result, err := svc.GetByID(999999)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestDatasourceService_List(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	// Create test data
	for i := 1; i <= 5; i++ {
		cfg := &datasource.ConnectionConfig{
			Host:     "localhost",
			Port:     3306,
			Database: fmt.Sprintf("db_%d", i),
		}
		cfgJSON, _ := json.Marshal(cfg)
		cfgStr := base64.StdEncoding.EncodeToString(cfgJSON)

		_, err := svc.Save(&datasource.WriteRequest{
			Name:          fmt.Sprintf("DS %d", i),
			Type:          "mysql",
			Configuration: &cfgStr,
		})
		require.NoError(t, err)
	}

	t.Run("list all", func(t *testing.T) {
		req := &datasource.ListRequest{}
		result, err := svc.List(req)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, result.Total, int64(5))
		assert.GreaterOrEqual(t, len(result.List), 5)
	})

	t.Run("list with keyword", func(t *testing.T) {
		keyword := "DS 1"
		req := &datasource.ListRequest{Keyword: &keyword}
		result, err := svc.List(req)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, result.Total, int64(1))
	})
}

func TestDatasourceService_Tree(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	// Create folder structure
	folder1, err := svc.CreateFolder("Folder 1", 0)
	require.NoError(t, err)

	folder2, err := svc.CreateFolder("Folder 2", folder1.ID)
	require.NoError(t, err)

	cfg := &datasource.ConnectionConfig{
		Host:     "localhost",
		Port:     3306,
		Database: "test_db",
	}
	cfgJSON, _ := json.Marshal(cfg)
	cfgStr := base64.StdEncoding.EncodeToString(cfgJSON)

	_, err = svc.Save(&datasource.WriteRequest{
		Name:          "DS in Folder 2",
		PID:           &folder2.ID,
		Type:          "mysql",
		Configuration: &cfgStr,
	})
	require.NoError(t, err)

	t.Run("tree returns all nodes", func(t *testing.T) {
		result, err := svc.Tree(&datasource.ListRequest{})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 3)
	})
}

func TestDatasourceService_CreateFolder(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	t.Run("create root folder", func(t *testing.T) {
		result, err := svc.CreateFolder("Root Folder", 0)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Root Folder", result.Name)
		assert.Equal(t, datasource.TypeFolder, result.Type)
		assert.NotNil(t, result.PID)
		assert.Equal(t, int64(0), *result.PID)
	})

	t.Run("create nested folder", func(t *testing.T) {
		parent, err := svc.CreateFolder("Parent", 0)
		require.NoError(t, err)

		child, err := svc.CreateFolder("Child", parent.ID)
		require.NoError(t, err)
		assert.Equal(t, parent.ID, *child.PID)
	})
}

func TestDatasourceService_Rename(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	created, err := svc.CreateFolder("Original Name", 0)
	require.NoError(t, err)

	t.Run("rename successfully", func(t *testing.T) {
		result, err := svc.Rename(created.ID, "New Name")
		require.NoError(t, err)
		assert.Equal(t, "New Name", result.Name)
	})

	t.Run("fail with empty name", func(t *testing.T) {
		result, err := svc.Rename(created.ID, "")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("fail with non-existent ID", func(t *testing.T) {
		result, err := svc.Rename(999999, "Test")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestDatasourceService_Move(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	// Create structure
	folder1, err := svc.CreateFolder("Folder 1", 0)
	require.NoError(t, err)

	folder2, err := svc.CreateFolder("Folder 2", 0)
	require.NoError(t, err)

	ds, err := svc.CreateFolder("Datasource", folder1.ID)
	require.NoError(t, err)

	t.Run("move to another folder", func(t *testing.T) {
		result, err := svc.Move(ds.ID, folder2.ID)
		require.NoError(t, err)
		assert.Equal(t, folder2.ID, *result.PID)
	})

	t.Run("move to root", func(t *testing.T) {
		result, err := svc.Move(ds.ID, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(0), *result.PID)
	})

	t.Run("fail to move to itself", func(t *testing.T) {
		result, err := svc.Move(ds.ID, ds.ID)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "cannot be itself")
	})

	t.Run("fail to move to descendant", func(t *testing.T) {
		// Create parent-child relationship
		parent, err := svc.CreateFolder("Parent", 0)
		require.NoError(t, err)

		child, err := svc.CreateFolder("Child", parent.ID)
		require.NoError(t, err)

		// Try to move parent into child
		result, err := svc.Move(parent.ID, child.ID)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "cannot be child")
	})

	t.Run("fail with invalid ID", func(t *testing.T) {
		result, err := svc.Move(0, folder1.ID)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "id is required")
	})
}

func TestDatasourceService_Delete(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	t.Run("delete single datasource", func(t *testing.T) {
		created, err := svc.CreateFolder("To Delete", 0)
		require.NoError(t, err)

		err = svc.Delete(created.ID)
		require.NoError(t, err)

		// Verify deleted
		result, err := svc.GetByID(created.ID)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("delete recursively", func(t *testing.T) {
		parent, err := svc.CreateFolder("Parent", 0)
		require.NoError(t, err)

		child, err := svc.CreateFolder("Child", parent.ID)
		require.NoError(t, err)

		err = svc.Delete(parent.ID)
		require.NoError(t, err)

		// Both should be deleted
		_, err = svc.GetByID(parent.ID)
		assert.Error(t, err)

		_, err = svc.GetByID(child.ID)
		assert.Error(t, err)
	})

	t.Run("fail with invalid ID", func(t *testing.T) {
		err := svc.Delete(0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "id is required")
	})
}

func TestDatasourceService_CheckRepeat(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	cfg := &datasource.ConnectionConfig{
		Host:     "localhost",
		Port:     3306,
		Database: "test_db",
	}
	cfgJSON, _ := json.Marshal(cfg)
	cfgStr := base64.StdEncoding.EncodeToString(cfgJSON)

	_, err := svc.Save(&datasource.WriteRequest{
		Name:          "Existing DS",
		Type:          "mysql",
		Configuration: &cfgStr,
	})
	require.NoError(t, err)

	t.Run("detect duplicate connection", func(t *testing.T) {
		req := &datasource.WriteRequest{
			Type:          "mysql",
			Configuration: &cfgStr,
		}

		isRepeat, err := svc.CheckRepeat(req)
		require.NoError(t, err)
		assert.True(t, isRepeat)
	})

	t.Run("no duplicate for different connection", func(t *testing.T) {
		newCfg := &datasource.ConnectionConfig{
			Host:     "different-host",
			Port:     3306,
			Database: "different_db",
		}
		newCfgJSON, _ := json.Marshal(newCfg)
		newCfgStr := base64.StdEncoding.EncodeToString(newCfgJSON)

		req := &datasource.WriteRequest{
			Type:          "mysql",
			Configuration: &newCfgStr,
		}

		isRepeat, err := svc.CheckRepeat(req)
		require.NoError(t, err)
		assert.False(t, isRepeat)
	})

	t.Run("skip check for folder type", func(t *testing.T) {
		req := &datasource.WriteRequest{
			Type: datasource.TypeFolder,
		}

		isRepeat, err := svc.CheckRepeat(req)
		require.NoError(t, err)
		assert.False(t, isRepeat)
	})

	t.Run("skip check for nil request", func(t *testing.T) {
		isRepeat, err := svc.CheckRepeat(nil)
		require.NoError(t, err)
		assert.False(t, isRepeat)
	})
}

func TestDatasourceService_PerDelete(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	t.Run("no relations", func(t *testing.T) {
		created, err := svc.CreateFolder("Test", 0)
		require.NoError(t, err)

		hasRelations, err := svc.PerDelete(created.ID)
		require.NoError(t, err)
		assert.False(t, hasRelations)
	})

	t.Run("fail with invalid ID", func(t *testing.T) {
		hasRelations, err := svc.PerDelete(0)
		assert.Error(t, err)
		assert.False(t, hasRelations)
	})
}

func TestDatasourceService_ValidateByID(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	t.Run("validate folder type - skip", func(t *testing.T) {
		created, err := svc.CreateFolder("Folder", 0)
		require.NoError(t, err)

		result, err := svc.ValidateByID(created.ID)
		require.NoError(t, err)
		assert.Equal(t, datasource.StatusSuccess, result.Status)
	})

	t.Run("validate non-existent ID", func(t *testing.T) {
		result, err := svc.ValidateByID(999999)
		require.NoError(t, err)
		assert.Equal(t, datasource.StatusError, result.Status)
	})
}

func TestDatasourceService_LatestTypes(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	t.Run("empty creator", func(t *testing.T) {
		result, err := svc.LatestTypes("")
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("with creator - no datasources", func(t *testing.T) {
		result, err := svc.LatestTypes("nonexistent_creator")
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("with creator - with datasources", func(t *testing.T) {
		// Create some datasources
		for i := 0; i < 3; i++ {
			cfg := &datasource.ConnectionConfig{
				Host:     "localhost",
				Port:     3306,
				Database: fmt.Sprintf("test_db_%d", i),
			}
			cfgJSON, _ := json.Marshal(cfg)
			cfgStr := base64.StdEncoding.EncodeToString(cfgJSON)

			req := &datasource.WriteRequest{
				Name:          fmt.Sprintf("LatestType DS %d", i),
				Type:          fmt.Sprintf("type_%d", i),
				Configuration: &cfgStr,
			}
			_, err := svc.Save(req)
			require.NoError(t, err)
		}

		// Note: LatestTypes requires create_by to be set
		// Since we can't set it through Save, this will return empty
		result, err := svc.LatestTypes("test_creator")
		require.NoError(t, err)
		// Result will be empty because create_by is not set
		assert.Empty(t, result)
	})
}

func TestDatasourceService_ShowFinishPage(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	t.Run("show finish page for new user", func(t *testing.T) {
		userID := time.Now().UnixNano()
		show, err := svc.ShowFinishPage(userID)
		require.NoError(t, err)
		assert.True(t, show)
	})

	t.Run("not show after setting", func(t *testing.T) {
		userID := time.Now().UnixNano()
		err := svc.SetShowFinishPage(userID)
		require.NoError(t, err)

		show, err := svc.ShowFinishPage(userID)
		require.NoError(t, err)
		assert.False(t, show)
	})

	t.Run("skip for invalid user ID", func(t *testing.T) {
		show, err := svc.ShowFinishPage(0)
		require.NoError(t, err)
		assert.False(t, show)

		err = svc.SetShowFinishPage(0)
		require.NoError(t, err)
	})
}

func TestDatasourceService_CompatDatasourceID(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	cfg := &datasource.ConnectionConfig{
		Host:     "localhost",
		Port:     3306,
		Database: "test_db",
	}
	cfgJSON, _ := json.Marshal(cfg)
	cfgStr := base64.StdEncoding.EncodeToString(cfgJSON)

	created, err := svc.Save(&datasource.WriteRequest{
		Name:          "Test DS",
		Type:          "mysql",
		Configuration: &cfgStr,
	})
	require.NoError(t, err)

	t.Run("get by exact ID", func(t *testing.T) {
		result, err := svc.GetByID(created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, result.ID)
	})

	t.Run("get by invalid ID", func(t *testing.T) {
		result, err := svc.GetByID(999999)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestDatasourceService_ListSyncRecord(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	t.Run("fail with invalid datasource ID", func(t *testing.T) {
		result, err := svc.ListSyncRecord(0, 1, 10)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid datasource id")
	})
}

func TestDatasourceService_ListSyncRecord_Success(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	dsID := int64(20260303)
	now := time.Now().UnixMilli()
	err := repo.CreateSyncTaskLog(&datasource.SyncRecord{
		DsID:        dsID,
		TaskID:      1001,
		StartTime:   now,
		CreateTime:  now,
		TaskStatus:  "running",
		TableName:   "orders",
		Name:        "orders",
		TriggerType: "table",
	})
	require.NoError(t, err)

	page, err := svc.ListSyncRecord(dsID, 0, 0)
	require.NoError(t, err)
	require.NotNil(t, page)
	assert.Equal(t, dsID, page.DatasourceID)
	assert.Equal(t, 1, page.Current)
	assert.Equal(t, 10, page.Size)
	assert.GreaterOrEqual(t, page.Total, int64(1))
	assert.NotEmpty(t, page.Records)
}

type mockSeatunnelSyncServiceServer struct {
	seatunnelv1.UnimplementedSyncServiceServer
	fail bool
}

func (m *mockSeatunnelSyncServiceServer) SubmitTask(context.Context, *seatunnelv1.SubmitTaskRequest) (*seatunnelv1.SubmitTaskResponse, error) {
	if m.fail {
		return nil, status.Error(codes.Internal, "submit failed")
	}
	return &seatunnelv1.SubmitTaskResponse{TaskId: "99001"}, nil
}

func startSeatunnelServerForIntegration(t *testing.T, fail bool) (string, func()) {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	seatunnelv1.RegisterSyncServiceServer(grpcServer, &mockSeatunnelSyncServiceServer{fail: fail})
	go func() {
		_ = grpcServer.Serve(lis)
	}()

	cleanup := func() {
		grpcServer.Stop()
		_ = lis.Close()
	}

	return lis.Addr().String(), cleanup
}

func TestDatasourceService_SyncAPITable_WithRepoLog(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	addr, cleanup := startSeatunnelServerForIntegration(t, false)
	defer cleanup()
	svc.SetSeatunnelConfig(addr, 3*time.Second, 0)

	dsID := int64(99001)
	result, err := svc.SyncAPITable(map[string]string{
		"datasourceId": strconv.FormatInt(dsID, 10),
		"source":       "api",
		"target":       "mysql",
		"tableName":    "orders",
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "99001", result["taskId"])

	page, err := svc.ListSyncRecord(dsID, 1, 10)
	require.NoError(t, err)
	require.NotNil(t, page)
	assert.NotEmpty(t, page.Records)
	assert.Equal(t, "running", page.Records[0].TaskStatus)
}

func TestDatasourceService_SyncAPITable_SubmitFailCreatesFailedLog(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	addr, cleanup := startSeatunnelServerForIntegration(t, true)
	defer cleanup()
	svc.SetSeatunnelConfig(addr, 3*time.Second, 0)

	dsID := int64(99002)
	_, err := svc.SyncAPITable(map[string]string{
		"datasourceId": strconv.FormatInt(dsID, 10),
		"source":       "api",
		"target":       "mysql",
		"tableName":    "orders",
	})
	assert.Error(t, err)

	page, listErr := svc.ListSyncRecord(dsID, 1, 10)
	require.NoError(t, listErr)
	require.NotNil(t, page)
	assert.NotEmpty(t, page.Records)
	assert.Equal(t, "failed", page.Records[0].TaskStatus)
}

func TestDatasourceService_CheckAPIDatasource(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	t.Run("fail with empty request", func(t *testing.T) {
		result, err := svc.CheckAPIDatasource(nil)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("fail with empty data", func(t *testing.T) {
		result, err := svc.CheckAPIDatasource(map[string]string{})
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("fail with missing data field", func(t *testing.T) {
		result, err := svc.CheckAPIDatasource(map[string]string{"type": "api"})
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("success with valid data", func(t *testing.T) {
		apiData := map[string]interface{}{
			"url":     "http://example.com/api",
			"method":  "GET",
			"headers": map[string]string{},
		}
		dataJSON, _ := json.Marshal(apiData)
		dataStr := base64.StdEncoding.EncodeToString(dataJSON)

		result, err := svc.CheckAPIDatasource(map[string]string{
			"data": dataStr,
			"type": "apiStructure",
		})
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "table", result["type"])
	})
}

func TestDatasourceService_SetSeatunnelConfig(t *testing.T) {
	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	t.Run("set config successfully", func(t *testing.T) {
		svc.SetSeatunnelConfig("localhost:8080", 30*time.Second, 3)
		assert.Equal(t, "localhost:8080", svc.seatunnelAddress)
		assert.Equal(t, 30*time.Second, svc.seatunnelTimeout)
		assert.Equal(t, 3, svc.seatunnelRetries)
	})

	t.Run("set config with empty address", func(t *testing.T) {
		svc.SetSeatunnelConfig("", 0, -1)
		assert.Equal(t, "", svc.seatunnelAddress)
	})
}

func TestDatasourceService_GetTables(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	t.Run("empty result for invalid datasource ID", func(t *testing.T) {
		result, err := svc.GetTables(&datasource.TableRequest{DatasourceID: 0})
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestDatasourceService_GetTableStatus(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	t.Run("empty result for invalid datasource ID", func(t *testing.T) {
		result, err := svc.GetTableStatus(&datasource.TableRequest{DatasourceID: 0})
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestDatasourceService_TableMetadataAndPreview_Success(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})
	_ = testDB.Exec("DELETE FROM core_dataset_table").Error
	_ = testDB.Exec("DROP TABLE IF EXISTS it_ds_preview_meta").Error

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	err := testDB.Exec("CREATE TABLE it_ds_preview_meta (id BIGINT PRIMARY KEY AUTO_INCREMENT, name VARCHAR(64), amount INT)").Error
	require.NoError(t, err)
	err = testDB.Exec("INSERT INTO it_ds_preview_meta (name, amount) VALUES ('Alice', 100), ('Bob', 90)").Error
	require.NoError(t, err)

	dsID := int64(88001)
	tableType := "table"
	tableName := "it_ds_preview_meta"
	rowName := "preview_meta"
	err = testDB.Create(&datasetdomain.CoreDatasetTable{
		Name:           &rowName,
		DatasourceID:   &dsID,
		DatasetGroupID: 0,
		PhysicalTable:  &tableName,
		Type:           &tableType,
	}).Error
	require.NoError(t, err)

	tables, err := svc.GetTables(&datasource.TableRequest{DatasourceID: dsID})
	require.NoError(t, err)
	require.Len(t, tables, 1)
	assert.Equal(t, tableName, tables[0].TableName)

	statusList, err := svc.GetTableStatus(&datasource.TableRequest{DatasourceID: dsID})
	require.NoError(t, err)
	require.Len(t, statusList, 1)
	assert.Equal(t, datasource.StatusSuccess, statusList[0].Status)
	assert.Equal(t, int64(0), statusList[0].LastUpdate)

	fields, err := svc.GetTableField(&datasource.TableRequest{TableName: tableName})
	require.NoError(t, err)
	assert.NotEmpty(t, fields)

	preview, err := svc.PreviewData(&datasource.TableRequest{TableName: tableName, Limit: 1})
	require.NoError(t, err)
	require.NotNil(t, preview)
	assert.NotEmpty(t, preview.Fields)
	assert.Len(t, preview.Data, 1)
	assert.Equal(t, int64(2), preview.Total)

	schemas, err := svc.GetSchema()
	require.NoError(t, err)
	assert.NotEmpty(t, schemas)

	err = testDB.Exec("DROP TABLE IF EXISTS it_ds_preview_meta").Error
	require.NoError(t, err)
}

func TestDatasourceService_GetTableField(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	t.Run("empty result for empty table name", func(t *testing.T) {
		result, err := svc.GetTableField(&datasource.TableRequest{TableName: ""})
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("empty result for whitespace table name", func(t *testing.T) {
		result, err := svc.GetTableField(&datasource.TableRequest{TableName: "   "})
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestDatasourceService_PreviewData(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	t.Run("empty result for empty table name", func(t *testing.T) {
		result, err := svc.PreviewData(&datasource.TableRequest{TableName: ""})
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.Fields)
		assert.Empty(t, result.Data)
		assert.Equal(t, int64(0), result.Total)
	})
}

func TestParseTaskID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{"empty string", "", 0},
		{"whitespace", "   ", 0},
		{"valid id", "123", 123},
		{"invalid format", "abc", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTaskID(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestShouldSkipRepeatCheck(t *testing.T) {
	tests := []struct {
		dsType   string
		expected bool
	}{
		{"folder", true},
		{"FOLDER", true},
		{"es", true},
		{"ES", true},
		{"api", true},
		{"my-api", true},
		{"excel", true},
		{"Excel", true},
		{"mysql", false},
		{"postgresql", false},
	}

	for _, tt := range tests {
		t.Run(tt.dsType, func(t *testing.T) {
			result := shouldSkipRepeatCheck(tt.dsType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRequiresSchemaMatch(t *testing.T) {
	tests := []struct {
		dsType   string
		expected bool
	}{
		{"sqlserver", true},
		{"db2", true},
		{"oracle", true},
		{"pg", true},
		{"redshift", true},
		{"mysql", false},
		{"Excel", false},
	}

	for _, tt := range tests {
		t.Run(tt.dsType, func(t *testing.T) {
			result := requiresSchemaMatch(tt.dsType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizedPID(t *testing.T) {
	tests := []struct {
		name     string
		pid      *int64
		expected int64
	}{
		{"nil pid", nil, 0},
		{"negative pid", intPtr(-1), 0},
		{"zero pid", intPtr(0), 0},
		{"positive pid", intPtr(123), 123},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizedPID(tt.pid)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDecodeConfig(t *testing.T) {
	t.Run("decode base64 encoded config", func(t *testing.T) {
		cfg := &datasource.ConnectionConfig{
			Host:     "localhost",
			Port:     3306,
			Database: "test_db",
		}
		cfgJSON, _ := json.Marshal(cfg)
		cfgStr := base64.StdEncoding.EncodeToString(cfgJSON)

		result, err := decodeConfig(cfgStr)
		require.NoError(t, err)
		assert.Equal(t, "localhost", result.Host)
		assert.Equal(t, 3306, result.Port)
		assert.Equal(t, "test_db", result.Database)
	})

	t.Run("decode plain JSON config", func(t *testing.T) {
		cfg := &datasource.ConnectionConfig{
			Host:     "localhost",
			Port:     3306,
			Database: "test_db",
		}
		cfgJSON, _ := json.Marshal(cfg)

		result, err := decodeConfig(string(cfgJSON))
		require.NoError(t, err)
		assert.Equal(t, "localhost", result.Host)
	})

	t.Run("fail with invalid JSON", func(t *testing.T) {
		result, err := decodeConfig("invalid json")
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestParseHostPort(t *testing.T) {
	tests := []struct {
		name       string
		cfg        *datasource.ConnectionConfig
		expectHost string
		expectPort int
	}{
		{"host and port set", &datasource.ConnectionConfig{Host: "localhost", Port: 3306}, "localhost", 3306},
		{"JDBC URL", &datasource.ConnectionConfig{JDBCUrl: "jdbc:mysql://myhost:3307/mydb"}, "myhost", 3307},
		{"empty config", &datasource.ConnectionConfig{}, "", 0},
		{"invalid JDBC URL", &datasource.ConnectionConfig{JDBCUrl: "invalid"}, "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, port := parseHostPort(tt.cfg)
			assert.Equal(t, tt.expectHost, host)
			assert.Equal(t, tt.expectPort, port)
		})
	}
}

func TestIsSameDatasourceConnection(t *testing.T) {
	tests := []struct {
		name     string
		dsType   string
		current  *datasource.ConnectionConfig
		compare  *datasource.ConnectionConfig
		expected bool
	}{
		{
			name:     "same connection",
			dsType:   "mysql",
			current:  &datasource.ConnectionConfig{Host: "localhost", Port: 3306, Database: "test"},
			compare:  &datasource.ConnectionConfig{Host: "localhost", Port: 3306, Database: "test"},
			expected: true,
		},
		{
			name:     "different host",
			dsType:   "mysql",
			current:  &datasource.ConnectionConfig{Host: "localhost", Port: 3306, Database: "test"},
			compare:  &datasource.ConnectionConfig{Host: "otherhost", Port: 3306, Database: "test"},
			expected: false,
		},
		{
			name:     "different port",
			dsType:   "mysql",
			current:  &datasource.ConnectionConfig{Host: "localhost", Port: 3306, Database: "test"},
			compare:  &datasource.ConnectionConfig{Host: "localhost", Port: 3307, Database: "test"},
			expected: false,
		},
		{
			name:     "different database",
			dsType:   "mysql",
			current:  &datasource.ConnectionConfig{Host: "localhost", Port: 3306, Database: "test1"},
			compare:  &datasource.ConnectionConfig{Host: "localhost", Port: 3306, Database: "test2"},
			expected: false,
		},
		{
			name:     "nil config",
			dsType:   "mysql",
			current:  nil,
			compare:  &datasource.ConnectionConfig{Host: "localhost", Port: 3306, Database: "test"},
			expected: false,
		},
		{
			name:     "pg with same schema",
			dsType:   "pg",
			current:  &datasource.ConnectionConfig{Host: "localhost", Port: 5432, Database: "test", Schema: "public"},
			compare:  &datasource.ConnectionConfig{Host: "localhost", Port: 5432, Database: "test", Schema: "public"},
			expected: true,
		},
		{
			name:     "pg with different schema",
			dsType:   "pg",
			current:  &datasource.ConnectionConfig{Host: "localhost", Port: 5432, Database: "test", Schema: "public"},
			compare:  &datasource.ConnectionConfig{Host: "localhost", Port: 5432, Database: "test", Schema: "private"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSameDatasourceConnection(tt.dsType, tt.current, tt.compare)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function
func intPtr(i int64) *int64 {
	return &i
}

// Additional test to verify strPtr is available from role_service.go
func TestStrPtrExists(t *testing.T) {
	// This test verifies that strPtr function exists in role_service.go
	// and is accessible from this test file
	result := strPtr("test")
	assert.NotNil(t, result)
	assert.Equal(t, "test", *result)
}

// Test with strconv for unique IDs
func TestDatasourceService_UniqueConstraints(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	uniqueSuffix := strconv.FormatInt(time.Now().UnixNano(), 10)

	cfg := &datasource.ConnectionConfig{
		Host:     "localhost",
		Port:     3306,
		Database: "test_db_" + uniqueSuffix,
	}
	cfgJSON, _ := json.Marshal(cfg)
	cfgStr := base64.StdEncoding.EncodeToString(cfgJSON)

	t.Run("create with unique name", func(t *testing.T) {
		result, err := svc.Save(&datasource.WriteRequest{
			Name:          "Unique DS " + uniqueSuffix,
			Type:          "mysql",
			Configuration: &cfgStr,
		})
		require.NoError(t, err)
		assert.NotNil(t, result)
	})
}

// TestDatasourceService_ResolveConfig tests resolveConfig paths
func TestDatasourceService_ResolveConfig(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	t.Run("resolve with datasource ID", func(t *testing.T) {
		cfg := &datasource.ConnectionConfig{Host: "localhost", Port: 3306, Database: "test"}
		cfgJSON, _ := json.Marshal(cfg)
		cfgStr := base64.StdEncoding.EncodeToString(cfgJSON)

		created, err := svc.Save(&datasource.WriteRequest{
			Name:          "Resolve Test",
			Type:          "mysql",
			Configuration: &cfgStr,
		})
		require.NoError(t, err)

		dsType, cfgRaw, err := svc.resolveConfig(&datasource.ValidateRequest{DatasourceID: &created.ID})
		require.NoError(t, err)
		assert.Equal(t, "mysql", dsType)
		assert.NotEmpty(t, cfgRaw)
	})

	t.Run("resolve with type and configuration", func(t *testing.T) {
		cfg := &datasource.ConnectionConfig{Host: "localhost", Port: 3306, Database: "test"}
		cfgJSON, _ := json.Marshal(cfg)
		cfgStr := base64.StdEncoding.EncodeToString(cfgJSON)
		dsType := "mysql"

		resolvedType, cfgRaw, err := svc.resolveConfig(&datasource.ValidateRequest{
			Type:          &dsType,
			Configuration: &cfgStr,
		})
		require.NoError(t, err)
		assert.Equal(t, "mysql", resolvedType)
		assert.NotEmpty(t, cfgRaw)
	})

	t.Run("fail with non-existent datasource ID", func(t *testing.T) {
		nonExistentID := int64(999999)
		_, _, err := svc.resolveConfig(&datasource.ValidateRequest{DatasourceID: &nonExistentID})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("fail with nil type", func(t *testing.T) {
		_, _, err := svc.resolveConfig(&datasource.ValidateRequest{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "type is required")
	})

	t.Run("fail with empty configuration", func(t *testing.T) {
		dsType := "mysql"
		_, _, err := svc.resolveConfig(&datasource.ValidateRequest{Type: &dsType})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "configuration is required")
	})
}

// TestDatasourceService_IsDescendant tests isDescendant function
func TestDatasourceService_IsDescendant(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	t.Run("not a descendant", func(t *testing.T) {
		parent, err := svc.CreateFolder("Parent", 0)
		require.NoError(t, err)

		other, err := svc.CreateFolder("Other", 0)
		require.NoError(t, err)

		isDesc, err := svc.isDescendant(parent.ID, other.ID)
		require.NoError(t, err)
		assert.False(t, isDesc)
	})

	t.Run("is a direct child", func(t *testing.T) {
		parent, err := svc.CreateFolder("Parent2", 0)
		require.NoError(t, err)

		child, err := svc.CreateFolder("Child2", parent.ID)
		require.NoError(t, err)

		isDesc, err := svc.isDescendant(parent.ID, child.ID)
		require.NoError(t, err)
		assert.True(t, isDesc)
	})

	t.Run("is a grandchild", func(t *testing.T) {
		grandparent, err := svc.CreateFolder("Grandparent", 0)
		require.NoError(t, err)

		parent, err := svc.CreateFolder("Parent3", grandparent.ID)
		require.NoError(t, err)

		child, err := svc.CreateFolder("Grandchild", parent.ID)
		require.NoError(t, err)

		isDesc, err := svc.isDescendant(grandparent.ID, child.ID)
		require.NoError(t, err)
		assert.True(t, isDesc)
	})
}

// TestDatasourceService_CompatDatasourceID_Nearest tests nearest ID resolution
func TestDatasourceService_CompatDatasourceID_Nearest(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	// Create a datasource
	cfg := &datasource.ConnectionConfig{Host: "localhost", Port: 3306, Database: "test"}
	cfgJSON, _ := json.Marshal(cfg)
	cfgStr := base64.StdEncoding.EncodeToString(cfgJSON)

	created, err := svc.Save(&datasource.WriteRequest{
		Name:          "Compat Test",
		Type:          "mysql",
		Configuration: &cfgStr,
	})
	require.NoError(t, err)

	t.Run("compat returns exact ID", func(t *testing.T) {
		fixedID, err := svc.compatDatasourceID(created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, fixedID)
	})

	t.Run("compat returns error for non-existent non-multiple-of-100", func(t *testing.T) {
		_, err := svc.compatDatasourceID(99999) // Not a multiple of 100, doesn't exist
		assert.Error(t, err)
	})

	t.Run("compat with zero ID returns zero", func(t *testing.T) {
		fixedID, err := svc.compatDatasourceID(0)
		require.NoError(t, err)
		assert.Equal(t, int64(0), fixedID)
	})

	t.Run("compat with negative ID returns same", func(t *testing.T) {
		fixedID, err := svc.compatDatasourceID(-1)
		require.NoError(t, err)
		assert.Equal(t, int64(-1), fixedID)
	})

	t.Run("compat resolves legacy multiple-of-100 id to nearest existing", func(t *testing.T) {
		legacyID := ((created.ID / 100) + 1) * 100
		if legacyID == created.ID {
			legacyID += 100
		}

		fixedID, err := svc.compatDatasourceID(legacyID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, fixedID)
	})

	t.Run("compat returns not found when multiple-of-100 has no nearest", func(t *testing.T) {
		_, err := svc.compatDatasourceID(1000000)
		assert.Error(t, err)
	})
}

// TestDatasourceService_CheckRepeat_EdgeCases tests additional CheckRepeat branches
func TestDatasourceService_CheckRepeat_EdgeCases(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	t.Run("empty type falls back to nodeType", func(t *testing.T) {
		cfg := &datasource.ConnectionConfig{Host: "localhost", Port: 3306, Database: "test_fallback"}
		cfgJSON, _ := json.Marshal(cfg)
		cfgStr := base64.StdEncoding.EncodeToString(cfgJSON)

		req := &datasource.WriteRequest{
			Name:          "NodeType Fallback",
			NodeType:      "mysql",
			Configuration: &cfgStr,
		}

		isRepeat, err := svc.CheckRepeat(req)
		require.NoError(t, err)
		assert.False(t, isRepeat)
	})

	t.Run("empty configuration returns false", func(t *testing.T) {
		req := &datasource.WriteRequest{
			Name: "No Config",
			Type: "mysql",
		}

		isRepeat, err := svc.CheckRepeat(req)
		require.NoError(t, err)
		assert.False(t, isRepeat)
	})

	t.Run("whitespace configuration returns false", func(t *testing.T) {
		ws := "   "
		req := &datasource.WriteRequest{
			Name:          "Whitespace Config",
			Type:          "mysql",
			Configuration: &ws,
		}

		isRepeat, err := svc.CheckRepeat(req)
		require.NoError(t, err)
		assert.False(t, isRepeat)
	})

	t.Run("invalid JSON configuration returns false", func(t *testing.T) {
		invalid := "not-base64"
		req := &datasource.WriteRequest{
			Name:          "Invalid Config",
			Type:          "mysql",
			Configuration: &invalid,
		}

		isRepeat, err := svc.CheckRepeat(req)
		require.NoError(t, err)
		assert.False(t, isRepeat)
	})

	t.Run("check with exclude ID", func(t *testing.T) {
		cfg := &datasource.ConnectionConfig{Host: "localhost", Port: 3306, Database: "test_exclude"}
		cfgJSON, _ := json.Marshal(cfg)
		cfgStr := base64.StdEncoding.EncodeToString(cfgJSON)

		// Create first datasource
		created, err := svc.Save(&datasource.WriteRequest{
			Name:          "Exclude Test",
			Type:          "mysql",
			Configuration: &cfgStr,
		})
		require.NoError(t, err)

		// Check repeat with same config but exclude the created ID
		req := &datasource.WriteRequest{
			ID:            created.ID,
			Name:          "Updated Name",
			Type:          "mysql",
			Configuration: &cfgStr,
		}

		isRepeat, err := svc.CheckRepeat(req)
		require.NoError(t, err)
		assert.False(t, isRepeat) // Should not be a repeat because we exclude the same ID
	})
}

// TestDatasourceService_Move_ErrorPaths tests additional Move error paths
func TestDatasourceService_Move_ErrorPaths(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	t.Run("move non-existent datasource", func(t *testing.T) {
		folder, err := svc.CreateFolder("Target Folder", 0)
		require.NoError(t, err)

		result, err := svc.Move(999999, folder.ID)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("move to root (pid=0)", func(t *testing.T) {
		folder, err := svc.CreateFolder("Parent for Root Move", 0)
		require.NoError(t, err)

		child, err := svc.CreateFolder("Child for Root Move", folder.ID)
		require.NoError(t, err)

		// Move child to root
		result, err := svc.Move(child.ID, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(0), *result.PID)
	})
}

// TestDatasourceService_Delete_DeepRecursive tests deep recursive deletion
func TestDatasourceService_Delete_DeepRecursive(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	// Create deeply nested structure: root -> level1 -> level2 -> leaf
	root, err := svc.CreateFolder("Root", 0)
	require.NoError(t, err)

	level1, err := svc.CreateFolder("Level1", root.ID)
	require.NoError(t, err)

	level2, err := svc.CreateFolder("Level2", level1.ID)
	require.NoError(t, err)

	leaf, err := svc.CreateFolder("Leaf", level2.ID)
	require.NoError(t, err)

	// Delete root - should cascade delete all children
	err = svc.Delete(root.ID)
	require.NoError(t, err)

	// Verify all are deleted
	_, err = svc.GetByID(root.ID)
	assert.Error(t, err)
	_, err = svc.GetByID(level1.ID)
	assert.Error(t, err)
	_, err = svc.GetByID(level2.ID)
	assert.Error(t, err)
	_, err = svc.GetByID(leaf.ID)
	assert.Error(t, err)
}

func TestDatasourceService_Move_ToChild(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	// Create parent folder
	parent, err := svc.CreateFolder("Parent", 0)
	require.NoError(t, err)

	// Create child folder
	child, err := svc.CreateFolder("Child", parent.ID)
	require.NoError(t, err)

	// Try to move parent to child - should fail
	_, err = svc.Move(parent.ID, child.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be child")
}

func TestDatasourceService_Move_DeepNested(t *testing.T) {
	cleanupTables(&datasource.CoreDatasource{})

	repo := repository.NewDatasourceRepository(testDB)
	svc := NewDatasourceService(repo)

	// Create nested structure
	root, err := svc.CreateFolder("Root", 0)
	require.NoError(t, err)

	level1, err := svc.CreateFolder("Level1", root.ID)
	require.NoError(t, err)

	level2, err := svc.CreateFolder("Level2", level1.ID)
	require.NoError(t, err)

	// Try to move root to level2 - should fail
	_, err = svc.Move(root.ID, level2.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be child")
}
