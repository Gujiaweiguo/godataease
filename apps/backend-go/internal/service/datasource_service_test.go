package service

import (
	"encoding/base64"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/integration/seatunnel"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupDatasourceServiceRepoTest(t *testing.T) (*DatasourceService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&datasource.CoreDatasource{}, &auto.CoreDatasourceTaskLog{}, &auto.CoreDatasetTable{}, &auto.CoreDsFinishPage{}))

	repo := repository.NewDatasourceRepository(db)
	return NewDatasourceService(repo), db
}

func encodeDatasourceConfig(t *testing.T, cfg *datasource.ConnectionConfig) string {
	t.Helper()

	raw, err := json.Marshal(cfg)
	require.NoError(t, err)
	return base64.StdEncoding.EncodeToString(raw)
}

func TestDatasourceServiceHelpers_DecodeMaybeBase64JSONMap(t *testing.T) {
	raw := base64.StdEncoding.EncodeToString([]byte(`{"a":1}`))
	result, err := decodeMaybeBase64JSONMap(raw)
	require.NoError(t, err)
	assert.Equal(t, float64(1), result["a"])

	result, err = decodeMaybeBase64JSONMap(`{"b":2}`)
	require.NoError(t, err)
	assert.Equal(t, float64(2), result["b"])

	_, err = decodeMaybeBase64JSONMap("")
	assert.Error(t, err)
}

func TestDatasourceServiceHelpers_ParseIDs(t *testing.T) {
	id, err := parseDatasourceID(map[string]string{"datasourceId": "12"})
	require.NoError(t, err)
	assert.Equal(t, int64(12), id)

	_, err = parseDatasourceID(map[string]string{})
	assert.Error(t, err)

	_, err = parseDatasourceID(map[string]string{"id": "bad"})
	assert.Error(t, err)

	assert.Equal(t, int64(99), parseTaskID("99"))
	assert.Equal(t, int64(0), parseTaskID("bad"))
}

func TestDatasourceService_Validate(t *testing.T) {
	svc := NewDatasourceService(nil)

	t.Run("folder datasource skips validation", func(t *testing.T) {
		dsType := datasource.TypeFolder
		cfg := "{}"

		resp, err := svc.Validate(&datasource.ValidateRequest{Type: &dsType, Configuration: &cfg})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, datasource.StatusSuccess, resp.Status)
		assert.Contains(t, resp.Message, "skip validation")
	})

	t.Run("missing host and port returns error status", func(t *testing.T) {
		dsType := "mysql"
		cfgJSON, err := json.Marshal(&datasource.ConnectionConfig{})
		require.NoError(t, err)
		cfgStr := base64.StdEncoding.EncodeToString(cfgJSON)

		resp, err := svc.Validate(&datasource.ValidateRequest{
			Type:          &dsType,
			Configuration: &cfgStr,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, datasource.StatusError, resp.Status)
		assert.Contains(t, resp.Message, "missing host/port")
	})

	t.Run("unreachable host returns connectivity error status", func(t *testing.T) {
		dsType := "mysql"
		cfgJSON, err := json.Marshal(&datasource.ConnectionConfig{Host: "198.51.100.1", Port: 81})
		require.NoError(t, err)
		cfgStr := base64.StdEncoding.EncodeToString(cfgJSON)

		resp, err := svc.Validate(&datasource.ValidateRequest{
			Type:          &dsType,
			Configuration: &cfgStr,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, datasource.StatusError, resp.Status)
		assert.Contains(t, resp.Message, "failed to connect")
	})
}

func TestDatasourceServiceHelpers_PingTCPTimeout(t *testing.T) {
	err := pingTCP("198.51.100.1", 81, time.Millisecond)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect")
}

func TestDatasourceServiceHelpers_NormalizedPID(t *testing.T) {
	assert.Equal(t, int64(0), normalizedPID(nil))

	negative := int64(-2)
	assert.Equal(t, int64(0), normalizedPID(&negative))

	positive := int64(9)
	assert.Equal(t, int64(9), normalizedPID(&positive))
}

func TestDatasourceServiceHelpers_RepeatCheckRules(t *testing.T) {
	assert.True(t, shouldSkipRepeatCheck("folder"))
	assert.True(t, shouldSkipRepeatCheck(" apiExcel "))
	assert.False(t, shouldSkipRepeatCheck("mysql"))

	assert.True(t, requiresSchemaMatch("oracle"))
	assert.True(t, requiresSchemaMatch(" SQLSERVER "))
	assert.False(t, requiresSchemaMatch("mysql"))
}

func TestDatasourceServiceHelpers_IsSameDatasourceConnection(t *testing.T) {
	base := &datasource.ConnectionConfig{Host: "db.local", Port: 3306, Database: "analytics"}
	assert.False(t, isSameDatasourceConnection("mysql", nil, base))
	assert.False(t, isSameDatasourceConnection("mysql", base, nil))

	assert.True(t, isSameDatasourceConnection("mysql", base, &datasource.ConnectionConfig{Host: "DB.LOCAL", Port: 3306, Database: "analytics"}))
	assert.False(t, isSameDatasourceConnection("mysql", base, &datasource.ConnectionConfig{Host: "other", Port: 3306, Database: "analytics"}))
	assert.False(t, isSameDatasourceConnection("mysql", base, &datasource.ConnectionConfig{Host: "db.local", Port: 3307, Database: "analytics"}))
	assert.False(t, isSameDatasourceConnection("mysql", base, &datasource.ConnectionConfig{Host: "db.local", Port: 3306, Database: ""}))

	assert.True(t, isSameDatasourceConnection("oracle", &datasource.ConnectionConfig{Host: "db.local", Port: 1521, Database: "analytics", Schema: "BI"}, &datasource.ConnectionConfig{Host: "db.local", Port: 1521, Database: "analytics", Schema: "bi"}))
	assert.False(t, isSameDatasourceConnection("oracle", &datasource.ConnectionConfig{Host: "db.local", Port: 1521, Database: "analytics", Schema: "BI"}, &datasource.ConnectionConfig{Host: "db.local", Port: 1521, Database: "analytics", Schema: "DW"}))

	jdbcCfg := &datasource.ConnectionConfig{JDBCUrl: "jdbc:mysql://db.local:3306/demo", Database: "demo"}
	assert.True(t, isSameDatasourceConnection("mysql", jdbcCfg, &datasource.ConnectionConfig{Host: "db.local", Port: 3306, Database: "demo"}))
}

func TestDatasourceService_ResolveConfig_ErrorsWithoutRepoLookup(t *testing.T) {
	svc := NewDatasourceService(nil)

	_, _, err := svc.resolveConfig(&datasource.ValidateRequest{})
	require.Error(t, err)
	assert.Equal(t, "datasource type is required", err.Error())

	dsType := "mysql"
	_, _, err = svc.resolveConfig(&datasource.ValidateRequest{Type: &dsType})
	require.Error(t, err)
	assert.Equal(t, "datasource configuration is required", err.Error())
}

func TestDatasourceService_SetSeatunnelConfig_Unit(t *testing.T) {
	svc := NewDatasourceService(nil)
	svc.SetSeatunnelConfig(" 127.0.0.1:1234 ", 5*time.Second, 3)
	svc.seatunnelClient = &seatunnel.Client{}

	svc.SetSeatunnelConfig(" seatunnel:9000 ", 0, -1)

	assert.Equal(t, "seatunnel:9000", svc.seatunnelAddress)
	assert.Equal(t, 5*time.Second, svc.seatunnelTimeout)
	assert.Equal(t, 3, svc.seatunnelRetries)
	assert.Nil(t, svc.seatunnelClient)
}

func TestDatasourceService_CompatDatasourceID_Unit(t *testing.T) {
	svc, db := setupDatasourceServiceRepoTest(t)
	require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 120, Name: "existing", Type: "mysql"}).Error)
	require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 230, Name: "nearest", Type: "mysql"}).Error)

	resolvedID, err := svc.compatDatasourceID(120)
	require.NoError(t, err)
	assert.Equal(t, int64(120), resolvedID)

	resolvedID, err = svc.compatDatasourceID(155)
	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	assert.Equal(t, int64(155), resolvedID)

	resolvedID, err = svc.compatDatasourceID(200)
	require.NoError(t, err)
	assert.Equal(t, int64(230), resolvedID)

	resolvedID, err = svc.compatDatasourceID(400)
	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	assert.Equal(t, int64(400), resolvedID)
}

func TestDatasourceService_CheckRepeat_Unit(t *testing.T) {
	svc, db := setupDatasourceServiceRepoTest(t)
	oracleCfg := encodeDatasourceConfig(t, &datasource.ConnectionConfig{Host: "db.local", Port: 1521, Database: "analytics", Schema: "BI"})
	malformedCfg := "not-json"
	mysqlCfg := encodeDatasourceConfig(t, &datasource.ConnectionConfig{Host: "mysql.local", Port: 3306, Database: "demo"})
	require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 11, Name: "oracle-a", Type: "oracle", Configuration: &oracleCfg}).Error)
	require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 12, Name: "oracle-bad", Type: "oracle", Configuration: &malformedCfg}).Error)
	require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 21, Name: "mysql-self", Type: "mysql", Configuration: &mysqlCfg}).Error)

	repeated, err := svc.CheckRepeat(&datasource.WriteRequest{Type: "oracle", Configuration: &oracleCfg})
	require.NoError(t, err)
	assert.True(t, repeated)

	repeated, err = svc.CheckRepeat(&datasource.WriteRequest{ID: 21, Type: "mysql", Configuration: &mysqlCfg})
	require.NoError(t, err)
	assert.False(t, repeated)

	folderCfg := "{}"
	repeated, err = svc.CheckRepeat(&datasource.WriteRequest{Type: datasource.TypeFolder, Configuration: &folderCfg})
	require.NoError(t, err)
	assert.False(t, repeated)
}

func TestDatasourceService_ResolveConfig_WithRepo(t *testing.T) {
	svc, db := setupDatasourceServiceRepoTest(t)
	configured := encodeDatasourceConfig(t, &datasource.ConnectionConfig{Host: "db.local", Port: 3306, Database: "demo"})
	require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 31, Name: "configured", Type: "mysql", Configuration: &configured}).Error)
	require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 32, Name: "empty-config", Type: "mysql"}).Error)

	dsID := int64(31)
	dsType, cfg, err := svc.resolveConfig(&datasource.ValidateRequest{DatasourceID: &dsID})
	require.NoError(t, err)
	assert.Equal(t, "mysql", dsType)
	assert.Equal(t, configured, cfg)

	dsID = 32
	_, _, err = svc.resolveConfig(&datasource.ValidateRequest{DatasourceID: &dsID})
	require.Error(t, err)
	assert.Equal(t, "datasource configuration is empty", err.Error())

	dsID = 999
	_, _, err = svc.resolveConfig(&datasource.ValidateRequest{DatasourceID: &dsID})
	require.Error(t, err)
	assert.Equal(t, "datasource not found", err.Error())
}

func TestDatasourceService_ListSyncRecord_Unit(t *testing.T) {
	t.Run("validates datasource id and repo availability", func(t *testing.T) {
		svc := NewDatasourceService(nil)

		page, err := svc.ListSyncRecord(0, 1, 10)
		require.Error(t, err)
		assert.Nil(t, page)
		assert.Equal(t, "invalid datasource id", err.Error())

		svc.repo = nil
		page, err = svc.ListSyncRecord(7, 1, 10)
		require.Error(t, err)
		assert.Nil(t, page)
		assert.Equal(t, "datasource repository is unavailable", err.Error())
	})

	t.Run("normalizes page size and returns records", func(t *testing.T) {
		svc, db := setupDatasourceServiceRepoTest(t)
		require.NoError(t, db.Create(&auto.CoreDatasourceTaskLog{DsID: 7, TaskID: 1001, StartTime: 10, TaskStatus: "running", PhysicalTableName: "table_a", CreateTime: 10, TriggerType: "datasource"}).Error)
		require.NoError(t, db.Create(&auto.CoreDatasourceTaskLog{DsID: 7, TaskID: 1002, StartTime: 20, TaskStatus: "success", PhysicalTableName: "table_b", CreateTime: 20, TriggerType: "table"}).Error)

		page, err := svc.ListSyncRecord(7, 0, 0)
		require.NoError(t, err)
		require.NotNil(t, page)
		assert.Equal(t, int64(2), page.Total)
		assert.Equal(t, 1, page.Current)
		assert.Equal(t, 10, page.Size)
		assert.Equal(t, int64(7), page.DatasourceID)
		require.Len(t, page.Records, 2)
		assert.Equal(t, int64(1002), page.Records[0].TaskID)
		assert.Equal(t, int64(1001), page.Records[1].TaskID)
	})
}

func TestDatasourceService_ListAndTree(t *testing.T) {
	svc, db := setupDatasourceServiceRepoTest(t)
	require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 1, Name: "Alpha", Type: "mysql"}).Error)
	require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 2, Name: "Beta", Type: datasource.TypeFolder}).Error)

	resp, err := svc.List(&datasource.ListRequest{Current: 0, Size: 0})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int64(2), resp.Total)
	assert.Equal(t, 1, resp.Current)
	assert.Equal(t, 10, resp.Size)
	require.Len(t, resp.List, 2)

	keyword := "Alpha"
	tree, err := svc.Tree(&datasource.ListRequest{Keyword: &keyword})
	require.NoError(t, err)
	require.Len(t, tree, 1)
	assert.Equal(t, "Alpha", tree[0].Name)
}

func TestDatasourceService_GetByIDAndValidateByID(t *testing.T) {
	svc, db := setupDatasourceServiceRepoTest(t)
	folderCfg := "{}"
	require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 130, Name: "Folder DS", Type: datasource.TypeFolder, Configuration: &folderCfg}).Error)

	item, err := svc.GetByID(100)
	require.NoError(t, err)
	require.NotNil(t, item)
	assert.Equal(t, int64(130), item.ID)

	resp, err := svc.ValidateByID(100)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, datasource.StatusSuccess, resp.Status)
	assert.Contains(t, resp.Message, "skip validation")
}

func TestDatasourceService_TableMetadataHelpers(t *testing.T) {
	svc, db := setupDatasourceServiceRepoTest(t)
	tableType := "db"
	require.NoError(t, db.Create(&auto.CoreDatasetTable{ID: 10, Name: "Orders", PhysicalTableName: "orders", DatasourceID: 7, DatasetGroupID: 1, Type: tableType}).Error)
	require.NoError(t, db.Create(&auto.CoreDatasetTable{ID: 11, Name: "Users", PhysicalTableName: "users", DatasourceID: 7, DatasetGroupID: 1, Type: tableType}).Error)
	require.NoError(t, db.Create(&auto.CoreDatasetTable{ID: 12, Name: "Payments", PhysicalTableName: "payments", DatasourceID: 7, DatasetGroupID: 1, Type: tableType}).Error)
	require.NoError(t, db.Create(&auto.CoreDatasetTable{ID: 13, Name: "Shipments", PhysicalTableName: "shipments", DatasourceID: 7, DatasetGroupID: 1, Type: tableType}).Error)
	require.NoError(t, db.Create(&auto.CoreDatasourceTaskLog{DsID: 7, TaskID: 501, StartTime: 11, EndTime: 12, TaskStatus: "failed", PhysicalTableName: "orders", CreateTime: 10, TriggerType: "table"}).Error)
	require.NoError(t, db.Create(&auto.CoreDatasourceTaskLog{DsID: 7, TaskID: 502, StartTime: 21, EndTime: 0, TaskStatus: "running", PhysicalTableName: "orders", CreateTime: 20, TriggerType: "table"}).Error)
	require.NoError(t, db.Create(&auto.CoreDatasourceTaskLog{DsID: 7, TaskID: 503, StartTime: 0, EndTime: 0, TaskStatus: "queued", PhysicalTableName: "payments", CreateTime: 31, TriggerType: "table"}).Error)
	require.NoError(t, db.Create(&auto.CoreDatasourceTaskLog{DsID: 7, TaskID: 504, StartTime: 41, EndTime: 45, TaskStatus: "cancelled", PhysicalTableName: "shipments", CreateTime: 40, TriggerType: "table"}).Error)

	list, err := svc.GetTables(&datasource.TableRequest{DatasourceID: 0})
	require.NoError(t, err)
	assert.Empty(t, list)

	list, err = svc.GetTables(&datasource.TableRequest{DatasourceID: 7})
	require.NoError(t, err)
	require.Len(t, list, 4)
	tableByName := make(map[string]datasource.TableInfo, len(list))
	for _, item := range list {
		tableByName[item.TableName] = item
	}
	assert.Equal(t, "Orders", tableByName["orders"].Name)
	assert.Equal(t, "db", tableByName["orders"].Type)
	assert.Equal(t, "Users", tableByName["users"].Name)
	assert.Equal(t, "db", tableByName["users"].Type)
	assert.Equal(t, "Payments", tableByName["payments"].Name)
	assert.Equal(t, "Shipments", tableByName["shipments"].Name)

	statusList, err := svc.GetTableStatus(&datasource.TableRequest{DatasourceID: 7})
	require.NoError(t, err)
	require.Len(t, statusList, 4)
	statusByTable := make(map[string]datasource.TableInfo, len(statusList))
	for _, item := range statusList {
		statusByTable[item.TableName] = item
	}
	assert.Equal(t, datasource.TableStatusUnderExecution, statusByTable["orders"].Status)
	assert.Equal(t, int64(21), statusByTable["orders"].LastUpdate)
	assert.Equal(t, datasource.TableStatusWarning, statusByTable["users"].Status)
	assert.Zero(t, statusByTable["users"].LastUpdate)
	assert.Equal(t, datasource.TableStatusPending, statusByTable["payments"].Status)
	assert.Equal(t, int64(31), statusByTable["payments"].LastUpdate)
	assert.Equal(t, datasource.TableStatusCancelled, statusByTable["shipments"].Status)
	assert.Equal(t, int64(45), statusByTable["shipments"].LastUpdate)
}

func TestDatasourceService_GetSchemaAndPreviewGuards(t *testing.T) {
	svc, _ := setupDatasourceServiceRepoTest(t)

	_, err := svc.GetSchema()
	require.Error(t, err)

	fields, err := svc.GetTableField(&datasource.TableRequest{TableName: "   "})
	require.NoError(t, err)
	assert.Empty(t, fields)

	_, err = svc.GetTableField(&datasource.TableRequest{TableName: "bad-name;drop"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table name")

	preview, err := svc.PreviewData(&datasource.TableRequest{TableName: "   "})
	require.NoError(t, err)
	require.NotNil(t, preview)
	assert.Empty(t, preview.Fields)
	assert.Empty(t, preview.Data)
	assert.Zero(t, preview.Total)

	_, err = svc.PreviewData(&datasource.TableRequest{TableName: "bad-name;drop", Limit: 5})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table name")
}

func TestDatasourceService_PreviewAndExcelWrappers(t *testing.T) {
	t.Run("preview data with valid sqlite table name propagates field introspection error", func(t *testing.T) {
		svc, db := setupDatasourceServiceRepoTest(t)
		require.NoError(t, db.Exec("CREATE TABLE preview_rows (name TEXT, amount INTEGER)").Error)
		require.NoError(t, db.Exec("INSERT INTO preview_rows (name, amount) VALUES ('alice', 10), ('bob', 20)").Error)

		preview, err := svc.PreviewData(&datasource.TableRequest{TableName: "preview_rows", Limit: 1})
		require.Error(t, err)
		assert.Nil(t, preview)
	})

	t.Run("upload file delegates to excel service", func(t *testing.T) {
		svc, _ := setupDatasourceServiceRepoTest(t)
		content := []byte("name,amount\nAlice,100\nBob,200\n")
		file := &failingMultipartFile{data: content}

		result, err := svc.UploadFile(file, &multipart.FileHeader{Filename: "upload.csv", Size: int64(len(content))}, 1, 0)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "upload", result.ExcelLabel)
		require.Len(t, result.Sheets, 1)
		assert.Equal(t, "upload.csv", result.Sheets[0].FileName)
	})

	t.Run("load remote file delegates to excel service", func(t *testing.T) {
		svc, _ := setupDatasourceServiceRepoTest(t)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("name,amount\nAlice,100\nBob,200\n"))
		}))
		defer server.Close()

		result, err := svc.LoadRemoteFile(server.URL+"/remote.csv", "", "", 1)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsSheet)
		require.Len(t, result.Sheets, 1)
		assert.Equal(t, "remote", result.Sheets[0].ExcelLabel)
	})
}

func TestDatasourceService_Save_Unit(t *testing.T) {
	t.Run("rejects empty name", func(t *testing.T) {
		svc, _ := setupDatasourceServiceRepoTest(t)

		item, err := svc.Save(&datasource.WriteRequest{Name: "   "})
		require.Error(t, err)
		assert.Nil(t, item)
		assert.Equal(t, "datasource name is required", err.Error())
	})

	t.Run("rejects duplicate name under same pid", func(t *testing.T) {
		svc, db := setupDatasourceServiceRepoTest(t)
		pid := int64(0)
		require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 201, PID: &pid, Name: "Dup", Type: "mysql"}).Error)

		item, err := svc.Save(&datasource.WriteRequest{Name: "Dup", PID: &pid, Type: "mysql"})
		require.Error(t, err)
		assert.Nil(t, item)
		assert.Equal(t, "datasource name already exists", err.Error())
	})

	t.Run("creates folder with default config", func(t *testing.T) {
		svc, _ := setupDatasourceServiceRepoTest(t)
		pid := int64(-9)

		item, err := svc.Save(&datasource.WriteRequest{Name: "Folder A", PID: &pid, Type: datasource.TypeFolder})
		require.NoError(t, err)
		require.NotNil(t, item)
		assert.Equal(t, "Folder A", item.Name)
		assert.Equal(t, datasource.TypeFolder, item.Type)
		require.NotNil(t, item.PID)
		assert.Equal(t, int64(0), *item.PID)
		require.NotNil(t, item.Configuration)
		assert.Equal(t, "{}", *item.Configuration)
	})
}

func TestDatasourceService_Update_Unit(t *testing.T) {
	t.Run("requires datasource id", func(t *testing.T) {
		svc, _ := setupDatasourceServiceRepoTest(t)

		item, err := svc.Update(&datasource.WriteRequest{Name: "No ID"})
		require.Error(t, err)
		assert.Nil(t, item)
		assert.Equal(t, "datasource id is required", err.Error())
	})

	t.Run("returns not found for missing datasource", func(t *testing.T) {
		svc, _ := setupDatasourceServiceRepoTest(t)

		item, err := svc.Update(&datasource.WriteRequest{ID: 999, Name: "Missing"})
		require.Error(t, err)
		assert.Nil(t, item)
		assert.Equal(t, "datasource not found", err.Error())
	})

	t.Run("rejects duplicate name and updates existing fields", func(t *testing.T) {
		svc, db := setupDatasourceServiceRepoTest(t)
		rootPID := int64(0)
		oldDesc := "old desc"
		require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 210, PID: &rootPID, Name: "Source A", Type: "mysql", Description: &oldDesc}).Error)
		require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 211, PID: &rootPID, Name: "Source B", Type: "mysql"}).Error)

		item, err := svc.Update(&datasource.WriteRequest{ID: 210, Name: "Source B"})
		require.Error(t, err)
		assert.Nil(t, item)
		assert.Equal(t, "datasource name already exists", err.Error())

		newDesc := "new desc"
		enableFill := true
		item, err = svc.Update(&datasource.WriteRequest{ID: 210, Name: "Source A Updated", Description: &newDesc, EnableDataFill: &enableFill})
		require.NoError(t, err)
		require.NotNil(t, item)
		assert.Equal(t, "Source A Updated", item.Name)
		require.NotNil(t, item.Description)
		assert.Equal(t, newDesc, *item.Description)
		require.NotNil(t, item.EnableDataFill)
		assert.True(t, *item.EnableDataFill)
	})
}

func TestDatasourceService_Rename_Unit(t *testing.T) {
	t.Run("validates missing datasource and name", func(t *testing.T) {
		svc, _ := setupDatasourceServiceRepoTest(t)

		item, err := svc.Rename(999, "Renamed")
		require.Error(t, err)
		assert.Nil(t, item)
		assert.Equal(t, "datasource not found", err.Error())

		pid := int64(0)
		svc, db := setupDatasourceServiceRepoTest(t)
		require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 220, PID: &pid, Name: "Original", Type: "mysql"}).Error)
		item, err = svc.Rename(220, "   ")
		require.Error(t, err)
		assert.Nil(t, item)
		assert.Equal(t, "datasource name is required", err.Error())
	})

	t.Run("rejects duplicate and updates name", func(t *testing.T) {
		svc, db := setupDatasourceServiceRepoTest(t)
		pid := int64(0)
		require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 221, PID: &pid, Name: "First", Type: "mysql"}).Error)
		require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 222, PID: &pid, Name: "Second", Type: "mysql"}).Error)

		item, err := svc.Rename(221, "Second")
		require.Error(t, err)
		assert.Nil(t, item)
		assert.Equal(t, "datasource name already exists", err.Error())

		item, err = svc.Rename(221, "Renamed")
		require.NoError(t, err)
		require.NotNil(t, item)
		assert.Equal(t, "Renamed", item.Name)
	})
}

func TestDatasourceService_Move_Unit(t *testing.T) {
	t.Run("validates id and self destination", func(t *testing.T) {
		svc, _ := setupDatasourceServiceRepoTest(t)

		item, err := svc.Move(0, 1)
		require.Error(t, err)
		assert.Nil(t, item)
		assert.Equal(t, "datasource id is required", err.Error())

		item, err = svc.Move(5, 5)
		require.Error(t, err)
		assert.Nil(t, item)
		assert.Equal(t, "destination folder cannot be itself", err.Error())
	})

	t.Run("rejects missing datasource descendant destination and duplicate name", func(t *testing.T) {
		svc, db := setupDatasourceServiceRepoTest(t)
		rootPID := int64(0)
		parentID := int64(230)
		childPID := int64(231)
		require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 230, PID: &rootPID, Name: "Parent", Type: datasource.TypeFolder}).Error)
		require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 231, PID: &parentID, Name: "Child", Type: datasource.TypeFolder}).Error)
		require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 232, PID: &childPID, Name: "Leaf", Type: "mysql"}).Error)
		require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 233, PID: &rootPID, Name: "Leaf", Type: datasource.TypeFolder}).Error)

		item, err := svc.Move(999, 0)
		require.Error(t, err)
		assert.Nil(t, item)
		assert.Equal(t, "datasource not found", err.Error())

		item, err = svc.Move(230, 232)
		require.Error(t, err)
		assert.Nil(t, item)
		assert.Equal(t, "destination folder cannot be child of current datasource", err.Error())

		item, err = svc.Move(232, 0)
		require.Error(t, err)
		assert.Nil(t, item)
		assert.Equal(t, "datasource name already exists", err.Error())

		item, err = svc.Move(232, 231)
		require.NoError(t, err)
		require.NotNil(t, item)
		require.NotNil(t, item.PID)
		assert.Equal(t, int64(231), *item.PID)
	})
}

func TestDatasourceService_DeleteAndPerDelete(t *testing.T) {
	t.Run("validates id and relation checks", func(t *testing.T) {
		svc, db := setupDatasourceServiceRepoTest(t)

		err := svc.Delete(0)
		require.Error(t, err)
		assert.Equal(t, "datasource id is required", err.Error())

		hasRelation, err := svc.PerDelete(0)
		require.Error(t, err)
		assert.False(t, hasRelation)
		assert.Equal(t, "datasource id is required", err.Error())

		require.NoError(t, db.Create(&auto.CoreDatasetTable{ID: 501, Name: "orders", PhysicalTableName: "orders", DatasourceID: 240, DatasetGroupID: 1, Type: "db"}).Error)
		hasRelation, err = svc.PerDelete(240)
		require.NoError(t, err)
		assert.True(t, hasRelation)
	})

	t.Run("soft deletes descendants recursively", func(t *testing.T) {
		svc, db := setupDatasourceServiceRepoTest(t)
		rootPID := int64(0)
		parentID := int64(241)
		require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 241, PID: &rootPID, Name: "Parent", Type: datasource.TypeFolder}).Error)
		require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 242, PID: &parentID, Name: "Child", Type: datasource.TypeFolder}).Error)
		require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 243, PID: &parentID, Name: "Leaf", Type: "mysql"}).Error)

		err := svc.Delete(241)
		require.NoError(t, err)

		for _, id := range []int64{241, 242, 243} {
			var item datasource.CoreDatasource
			err = db.Unscoped().Where("id = ?", id).First(&item).Error
			require.NoError(t, err)
			require.NotNil(t, item.DelFlag)
			assert.Equal(t, 1, *item.DelFlag)
		}
	})
}

func TestDatasourceService_LightweightWrappersAndGuards(t *testing.T) {
	t.Run("set resource permission service stores pointer", func(t *testing.T) {
		svc := NewDatasourceService(nil)
		permSvc := &ResourcePermissionService{}

		svc.SetResourcePermissionService(permSvc)

		assert.Same(t, permSvc, svc.resourcePermService)
	})

	t.Run("create folder delegates to save semantics", func(t *testing.T) {
		svc, _ := setupDatasourceServiceRepoTest(t)

		item, err := svc.CreateFolder("Folder Wrapper", 6)
		require.NoError(t, err)
		require.NotNil(t, item)
		assert.Equal(t, datasource.TypeFolder, item.Type)
		require.NotNil(t, item.PID)
		assert.Equal(t, int64(6), *item.PID)
		require.NotNil(t, item.Configuration)
		assert.Equal(t, "{}", *item.Configuration)
	})

	t.Run("inherited permission helper short-circuits without service or parent", func(t *testing.T) {
		svc := NewDatasourceService(nil)

		err := svc.applyInheritedPermissionsOnCreate(1, "demo", 0)
		require.NoError(t, err)

		svc.resourcePermService = &ResourcePermissionService{}
		err = svc.applyInheritedPermissionsOnCreate(1, "demo", 0)
		require.NoError(t, err)
	})

	t.Run("backfill guards missing dependencies", func(t *testing.T) {
		svc, _ := setupDatasourceServiceRepoTest(t)

		report, err := svc.BackfillGovernedResources()
		require.Error(t, err)
		assert.Nil(t, report)
		assert.Equal(t, "resource permission service not initialized", err.Error())

		svc = NewDatasourceService(nil)
		svc.repo = nil
		report, err = svc.BackfillGovernedResourcesWithOptions(nil)
		require.Error(t, err)
		assert.Nil(t, report)
		assert.Equal(t, "datasource repository not initialized", err.Error())
	})

	t.Run("latest types and finish page wrappers handle lightweight cases", func(t *testing.T) {
		svc, db := setupDatasourceServiceRepoTest(t)

		types, err := svc.LatestTypes("")
		require.NoError(t, err)
		assert.Empty(t, types)

		show, err := svc.ShowFinishPage(0)
		require.NoError(t, err)
		assert.False(t, show)

		show, err = svc.ShowFinishPage(9)
		require.NoError(t, err)
		assert.True(t, show)

		require.NoError(t, db.Create(&auto.CoreDsFinishPage{ID: 9}).Error)
		show, err = svc.ShowFinishPage(9)
		require.NoError(t, err)
		assert.False(t, show)

		err = svc.SetShowFinishPage(0)
		require.NoError(t, err)
	})
}

func TestDatasourceService_SyncTaskHelpers(t *testing.T) {
	t.Run("ensure seatunnel client uses cache and validates config", func(t *testing.T) {
		svc := NewDatasourceService(nil)

		client, err := svc.ensureSeatunnelClient()
		require.Error(t, err)
		assert.Nil(t, client)
		assert.Equal(t, "seatunnel grpc address is not configured", err.Error())

		cached := &seatunnel.Client{}
		svc.seatunnelClient = cached
		client, err = svc.ensureSeatunnelClient()
		require.NoError(t, err)
		assert.Same(t, cached, client)
	})

	t.Run("submit sync task validates request and logs failed submission", func(t *testing.T) {
		svc, db := setupDatasourceServiceRepoTest(t)

		result, err := svc.submitSyncTask(nil, "datasource")
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "request is required", err.Error())

		svc.seatunnelClient = &seatunnel.Client{}
		result, err = svc.submitSyncTask(map[string]string{"datasourceId": "8", "tableName": "orders"}, "datasource")
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "seatunnel submit task failed")

		var rows []auto.CoreDatasourceTaskLog
		require.NoError(t, db.Order("id ASC").Find(&rows).Error)
		require.Len(t, rows, 1)
		assert.Equal(t, int64(8), rows[0].DsID)
		assert.Equal(t, int64(0), rows[0].TaskID)
		assert.Equal(t, seatunnel.StatusFailed, rows[0].TaskStatus)
		assert.Equal(t, "orders", rows[0].PhysicalTableName)
		assert.Contains(t, rows[0].Info, "seatunnel submit task failed")
		assert.Equal(t, "datasource", rows[0].TriggerType)
	})

	t.Run("sync wrappers and cancel task expose helper errors", func(t *testing.T) {
		svc := NewDatasourceService(nil)

		result, err := svc.SyncAPITable(nil)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "request is required", err.Error())

		result, err = svc.SyncAPIDs(map[string]string{})
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "request is required", err.Error())

		err = svc.CancelSyncTask("task-1")
		require.Error(t, err)
		assert.Equal(t, "seatunnel grpc address is not configured", err.Error())

		svc.seatunnelClient = &seatunnel.Client{}
		err = svc.CancelSyncTask("")
		require.Error(t, err)
		assert.Equal(t, "task id is required", err.Error())
	})
}

func TestDatasourceService_CheckAPIDatasource_Unit(t *testing.T) {
	svc := NewDatasourceService(nil)

	_, err := svc.CheckAPIDatasource(nil)
	require.Error(t, err)
	assert.Equal(t, "request is required", err.Error())

	_, err = svc.CheckAPIDatasource(map[string]string{"type": "apiStructure"})
	require.Error(t, err)
	assert.Equal(t, "data is required", err.Error())

	payload := base64.StdEncoding.EncodeToString([]byte(`{"url":"https://example.com"}`))
	result, err := svc.CheckAPIDatasource(map[string]string{"data": payload, "type": "apiStructure"})
	require.NoError(t, err)
	assert.Equal(t, "table", result["type"])
	assert.Equal(t, true, result["showApiStructure"])
	assert.Equal(t, "api_table", result["name"])
	assert.Equal(t, "https://example.com", result["url"])

	result, err = svc.CheckAPIDatasource(map[string]string{"data": `{"name":"kept-name"}`})
	require.NoError(t, err)
	assert.Equal(t, "kept-name", result["name"])
}

func TestDatasourceService_RepeatCheckHelpers(t *testing.T) {
	t.Run("resolveRepeatCheckInput guards", func(t *testing.T) {
		dsType, cfg := resolveRepeatCheckInput(nil)
		assert.Equal(t, "", dsType)
		assert.Nil(t, cfg)

		dsType, cfg = resolveRepeatCheckInput(&datasource.WriteRequest{})
		assert.Equal(t, "", dsType)
		assert.Nil(t, cfg)

		dsType, cfg = resolveRepeatCheckInput(&datasource.WriteRequest{Type: "api"})
		assert.Equal(t, "", dsType)
		assert.Nil(t, cfg)

		blank := ""
		dsType, cfg = resolveRepeatCheckInput(&datasource.WriteRequest{Type: "mysql", Configuration: &blank})
		assert.Equal(t, "", dsType)
		assert.Nil(t, cfg)

		cfgStr := "not-json"
		dsType, cfg = resolveRepeatCheckInput(&datasource.WriteRequest{Type: "mysql", Configuration: &cfgStr})
		assert.Equal(t, "", dsType)
		assert.Nil(t, cfg)

		valid := encodeDatasourceConfig(t, &datasource.ConnectionConfig{Host: "localhost"})
		dsType, cfg = resolveRepeatCheckInput(&datasource.WriteRequest{Type: "mysql", Configuration: &valid})
		assert.Equal(t, "mysql", dsType)
		assert.NotNil(t, cfg)
	})

	t.Run("candidateRepeatCheckConfig guards", func(t *testing.T) {
		assert.Nil(t, candidateRepeatCheckConfig(nil))
		assert.Nil(t, candidateRepeatCheckConfig(&datasource.CoreDatasource{Type: "api"}))

		blank := ""
		assert.Nil(t, candidateRepeatCheckConfig(&datasource.CoreDatasource{Type: "mysql", Configuration: &blank}))

		assert.Nil(t, candidateRepeatCheckConfig(&datasource.CoreDatasource{Type: "mysql", Configuration: strPtrDatasource("bad-json")}))

		valid := encodeDatasourceConfig(t, &datasource.ConnectionConfig{Host: "localhost"})
		cfg := candidateRepeatCheckConfig(&datasource.CoreDatasource{Type: "mysql", Configuration: &valid})
		assert.NotNil(t, cfg)
	})
}

func TestDatasourceService_LatestTypesAndFinishPage(t *testing.T) {
	svc, _ := setupDatasourceServiceRepoTest(t)

	types, err := svc.LatestTypes("")
	require.NoError(t, err)
	assert.Empty(t, types)

	show, err := svc.ShowFinishPage(0)
	require.NoError(t, err)
	assert.False(t, show)

	show, err = svc.ShowFinishPage(42)
	require.NoError(t, err)
	assert.True(t, show)
}

func TestDatasourceService_BackfillGuardBranches(t *testing.T) {
	svc, _ := setupDatasourceServiceRepoTest(t)

	_, err := svc.BackfillGovernedResourcesWithOptions(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resource permission service not initialized")

	_, err = svc.BackfillGovernedResources()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resource permission service not initialized")
}

func TestDatasourceService_DeleteRecursiveBranches(t *testing.T) {
	svc, db := setupDatasourceServiceRepoTest(t)

	require.NoError(t, db.Create(&datasource.CoreDatasource{
		Name: "parent", Type: "mysql", Status: strPtrDatasource("Success"),
		Configuration: strPtrDatasource(encodeDatasourceConfig(t, &datasource.ConnectionConfig{Host: "h"})),
	}).Error)
	var parent datasource.CoreDatasource
	require.NoError(t, db.Where("name = ?", "parent").First(&parent).Error)

	require.NoError(t, db.Create(&datasource.CoreDatasource{
		Name: "child", Type: "mysql", Status: strPtrDatasource("Success"), PID: int64PtrDatasource(parent.ID),
		Configuration: strPtrDatasource(encodeDatasourceConfig(t, &datasource.ConnectionConfig{Host: "h"})),
	}).Error)

	_, err := svc.PerDelete(0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "datasource id is required")

	hasRelations, err := svc.PerDelete(parent.ID)
	require.NoError(t, err)
	assert.False(t, hasRelations)

	require.NoError(t, svc.Delete(parent.ID))
	var remaining int64
	require.NoError(t, db.Model(&datasource.CoreDatasource{}).Where("del_flag = 0 OR del_flag IS NULL").Count(&remaining).Error)
	assert.Equal(t, int64(0), remaining)
}

func TestDatasourceService_MoveAndValidateBranches(t *testing.T) {
	svc, db := setupDatasourceServiceRepoTest(t)

	require.NoError(t, db.Create(&datasource.CoreDatasource{
		Name: "root", Type: "mysql", Status: strPtrDatasource("Success"),
		Configuration: strPtrDatasource(encodeDatasourceConfig(t, &datasource.ConnectionConfig{Host: "h"})),
	}).Error)
	var root datasource.CoreDatasource
	require.NoError(t, db.Where("name = ?", "root").First(&root).Error)

	require.NoError(t, db.Create(&datasource.CoreDatasource{
		Name: "leaf", Type: "mysql", Status: strPtrDatasource("Success"), PID: int64PtrDatasource(root.ID),
		Configuration: strPtrDatasource(encodeDatasourceConfig(t, &datasource.ConnectionConfig{Host: "h"})),
	}).Error)
	var leaf datasource.CoreDatasource
	require.NoError(t, db.Where("name = ?", "leaf").First(&leaf).Error)

	_, err := svc.Move(0, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "datasource id is required")

	moved, err := svc.Move(root.ID, 0)
	require.NoError(t, err)
	assert.NotNil(t, moved)

	_, err = svc.Move(root.ID, root.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "destination folder cannot be itself")

	resp, err := svc.Validate(&datasource.ValidateRequest{})
	require.NoError(t, err)
	assert.Equal(t, "Error", resp.Status)
	assert.Contains(t, resp.Message, "datasource type is required")
}

func strPtrDatasource(v string) *string {
	return &v
}

func int64PtrDatasource(v int64) *int64 {
	return &v
}
