package service

import (
	"encoding/json"
	"testing"

	"dataease/backend/internal/domain/audit"
	"dataease/backend/internal/domain/visualization"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// RecordExportLog tests
// ---------------------------------------------------------------------------

func TestRecordExportLog_NilRequest(t *testing.T) {
	svc := NewVisualizationService(nil)
	auditSvc, _ := setupAuditServiceRepoTest(t)
	svc.SetAuditService(auditSvc)

	err := svc.RecordExportLog(nil, nil, nil, nil, nil, "pdf")
	assert.NoError(t, err, "nil request should return nil without error")
}

func TestRecordExportLog_NilID(t *testing.T) {
	svc := NewVisualizationService(nil)
	auditSvc, _ := setupAuditServiceRepoTest(t)
	svc.SetAuditService(auditSvc)

	err := svc.RecordExportLog(&visualization.ExportLogRequest{ID: nil, Type: "dashboard"}, nil, nil, nil, nil, "pdf")
	assert.NoError(t, err, "nil ID should return nil without error")
}

func TestRecordExportLog_ZeroID(t *testing.T) {
	svc := NewVisualizationService(nil)
	auditSvc, _ := setupAuditServiceRepoTest(t)
	svc.SetAuditService(auditSvc)

	zeroID := int64(0)
	err := svc.RecordExportLog(&visualization.ExportLogRequest{ID: &zeroID, Type: "dashboard"}, nil, nil, nil, nil, "pdf")
	assert.NoError(t, err, "zero ID should return nil without error")
}

func TestRecordExportLog_NilAuditService(t *testing.T) {
	svc := NewVisualizationService(nil)
	id := int64(42)
	err := svc.RecordExportLog(&visualization.ExportLogRequest{ID: &id, Type: "dashboard"}, nil, nil, nil, nil, "pdf")
	assert.NoError(t, err, "nil auditService should return nil without error")
}

func TestRecordExportLog_LogTypeMapping(t *testing.T) {
	tests := []struct {
		name           string
		logType        string
		expectedAction string
	}{
		{"default empty", "", "导出资源"},
		{"app", "app", "导出应用模板"},
		{"template", "template", "导出样式模板"},
		{"pdf", "pdf", "导出PDF"},
		{"img", "img", "导出图片"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			auditSvc, db := setupAuditServiceRepoTest(t)
			svc := NewVisualizationService(nil)
			svc.SetAuditService(auditSvc)

			id := int64(100)
			userID := int64(1)
			username := "tester"
			ip := "10.0.0.1"
			ua := "test-agent"

			err := svc.RecordExportLog(
				&visualization.ExportLogRequest{ID: &id, Type: "dashboard"},
				&userID, &username, &ip, &ua, tc.logType,
			)
			require.NoError(t, err)

			var logs []audit.AuditLog
			require.NoError(t, db.Order("id ASC").Find(&logs).Error)
			require.Len(t, logs, 1)
			assert.Equal(t, tc.expectedAction, logs[0].ActionName)
		})
	}
}

func TestRecordExportLog_ResourceTypeMapping(t *testing.T) {
	tests := []struct {
		name            string
		reqType         string
		expectedResType string
	}{
		{"dashboard type", "dashboard", "DASHBOARD"},
		{"empty type", "", "DASHBOARD"},
		{"screen type", "screen", "SCREEN"},
		{"dataV type", "dataV", "SCREEN"},
		{"SCREEN case insensitive", "SCREEN", "SCREEN"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			auditSvc, db := setupAuditServiceRepoTest(t)
			svc := NewVisualizationService(nil)
			svc.SetAuditService(auditSvc)

			id := int64(200)
			err := svc.RecordExportLog(
				&visualization.ExportLogRequest{ID: &id, Type: tc.reqType},
				nil, nil, nil, nil, "",
			)
			require.NoError(t, err)

			var logs []audit.AuditLog
			require.NoError(t, db.Order("id ASC").Find(&logs).Error)
			require.Len(t, logs, 1)
			require.NotNil(t, logs[0].ResourceType)
			assert.Equal(t, tc.expectedResType, *logs[0].ResourceType)
		})
	}
}

func TestRecordExportLog_ErrorPropagation(t *testing.T) {
	svc := NewVisualizationService(nil)
	auditSvc := buildClosedAuditServiceForTest(t)
	svc.SetAuditService(auditSvc)

	id := int64(999)
	err := svc.RecordExportLog(
		&visualization.ExportLogRequest{ID: &id, Type: "dashboard"},
		nil, nil, nil, nil, "pdf",
	)
	require.Error(t, err, "should propagate auditService error")
}

func TestRecordExportLog_PassesAllFields(t *testing.T) {
	auditSvc, db := setupAuditServiceRepoTest(t)
	svc := NewVisualizationService(nil)
	svc.SetAuditService(auditSvc)

	id := int64(555)
	userID := int64(42)
	username := "admin"
	ip := "192.168.1.1"
	ua := "GoTest/1.0"

	err := svc.RecordExportLog(
		&visualization.ExportLogRequest{ID: &id, Type: "dataV"},
		&userID, &username, &ip, &ua, "img",
	)
	require.NoError(t, err)

	var logs []audit.AuditLog
	require.NoError(t, db.Order("id ASC").Find(&logs).Error)
	require.Len(t, logs, 1)

	assert.Equal(t, audit.ActionTypeDataAccess, logs[0].ActionType)
	assert.Equal(t, audit.OperationExport, logs[0].Operation)
	assert.Equal(t, "导出图片", logs[0].ActionName)
	require.NotNil(t, logs[0].ResourceType)
	assert.Equal(t, "SCREEN", *logs[0].ResourceType)
	require.NotNil(t, logs[0].ResourceID)
	assert.Equal(t, id, *logs[0].ResourceID)
	require.NotNil(t, logs[0].UserID)
	assert.Equal(t, userID, *logs[0].UserID)
	require.NotNil(t, logs[0].Username)
	assert.Equal(t, username, *logs[0].Username)
	require.NotNil(t, logs[0].IPAddress)
	assert.Equal(t, ip, *logs[0].IPAddress)
	require.NotNil(t, logs[0].UserAgent)
	assert.Equal(t, ua, *logs[0].UserAgent)
}

// ---------------------------------------------------------------------------
// AppCanvasNameCheck tests
// ---------------------------------------------------------------------------

func TestAppCanvasNameCheck_NilRequest(t *testing.T) {
	svc, _, _ := setupVisualizationServiceRepoTest(t)
	result, err := svc.AppCanvasNameCheck(nil)
	require.NoError(t, err)
	assert.Equal(t, "success", result)
}

func TestAppCanvasNameCheck_NilDatasetRepo(t *testing.T) {
	svc, _, _ := setupVisualizationServiceRepoTest(t)
	result, err := svc.AppCanvasNameCheck(&visualization.AppCanvasNameCheckRequest{
		DatasetFolderName: "SomeName",
	})
	require.NoError(t, err)
	assert.Equal(t, "success", result, "nil datasetRepo should return success early")
}

func TestAppCanvasNameCheck_BlankFolderName(t *testing.T) {
	svc, db := setupExport2AppCheckTest(t)
	svc.SetDatasetRepository(repository.NewDatasetRepository(db))

	result, err := svc.AppCanvasNameCheck(&visualization.AppCanvasNameCheckRequest{
		DatasetFolderName: "",
	})
	require.NoError(t, err)
	assert.Equal(t, "success", result)

	result, err = svc.AppCanvasNameCheck(&visualization.AppCanvasNameCheckRequest{
		DatasetFolderName: "   ",
	})
	require.NoError(t, err)
	assert.Equal(t, "success", result)
}

func TestAppCanvasNameCheck_DuplicateName(t *testing.T) {
	svc, db := setupExport2AppCheckTest(t)
	svc.SetDatasetRepository(repository.NewDatasetRepository(db))

	pid := int64(10)
	require.NoError(t, db.Exec(`INSERT INTO core_dataset_group (id, name, pid, node_type, create_by, create_time, update_by, last_update_time) VALUES (1, 'MyFolder', 10, 'folder', 'admin', 1, 'admin', 1)`).Error)

	result, err := svc.AppCanvasNameCheck(&visualization.AppCanvasNameCheckRequest{
		DatasetFolderPid:  &pid,
		DatasetFolderName: "MyFolder",
	})
	require.NoError(t, err)
	assert.Equal(t, "repeat", result)
}

func TestAppCanvasNameCheck_UniqueName(t *testing.T) {
	svc, db := setupExport2AppCheckTest(t)
	svc.SetDatasetRepository(repository.NewDatasetRepository(db))

	pid := int64(10)
	require.NoError(t, db.Exec(`INSERT INTO core_dataset_group (id, name, pid, node_type, create_by, create_time, update_by, last_update_time) VALUES (1, 'Existing', 10, 'folder', 'admin', 1, 'admin', 1)`).Error)

	result, err := svc.AppCanvasNameCheck(&visualization.AppCanvasNameCheckRequest{
		DatasetFolderPid:  &pid,
		DatasetFolderName: "NewName",
	})
	require.NoError(t, err)
	assert.Equal(t, "success", result)
}

func TestAppCanvasNameCheck_NilPID(t *testing.T) {
	svc, db := setupExport2AppCheckTest(t)
	svc.SetDatasetRepository(repository.NewDatasetRepository(db))

	require.NoError(t, db.Exec(`INSERT INTO core_dataset_group (id, name, pid, node_type, create_by, create_time, update_by, last_update_time) VALUES (1, 'RootFolder', 0, 'folder', 'admin', 1, 'admin', 1)`).Error)

	result, err := svc.AppCanvasNameCheck(&visualization.AppCanvasNameCheckRequest{
		DatasetFolderPid:  nil,
		DatasetFolderName: "RootFolder",
	})
	require.NoError(t, err)
	assert.Equal(t, "repeat", result)
}

func TestAppCanvasNameCheck_RepoError(t *testing.T) {
	svc, db := setupExport2AppCheckTest(t)
	svc.SetDatasetRepository(repository.NewDatasetRepository(db))
	closeVisualizationDB(t, db)

	pid := int64(10)
	_, err := svc.AppCanvasNameCheck(&visualization.AppCanvasNameCheckRequest{
		DatasetFolderPid:  &pid,
		DatasetFolderName: "AnyName",
	})
	require.Error(t, err, "closed DB should cause repo error")
}

// ---------------------------------------------------------------------------
// processAppData tests
// ---------------------------------------------------------------------------

func TestProcessAppData_ShortString(t *testing.T) {
	result := processAppData("short", 999)
	assert.Equal(t, "short", result, "string <= 10 chars should pass through unchanged")
}

func TestProcessAppData_EmptyString(t *testing.T) {
	result := processAppData("", 999)
	assert.Equal(t, "", result)
}

func TestProcessAppData_InvalidJSON(t *testing.T) {
	input := "this is not json but it is longer than ten characters"
	result := processAppData(input, 999)
	assert.Equal(t, input, result, "invalid JSON should pass through unchanged")
}

func TestProcessAppData_MissingVisualizationInfo(t *testing.T) {
	input := `{"otherField":[{"id":1}],"somethingElse":true}`
	result := processAppData(input, 999)
	assert.Equal(t, input, result, "missing visualizationInfo key should pass through unchanged")
}

func TestProcessAppData_NonPositiveID(t *testing.T) {
	input := `{"visualizationInfo":{"id":0,"name":"test"}}`
	result := processAppData(input, 999)
	assert.Equal(t, input, result, "non-positive ID should pass through unchanged")
}

func TestProcessAppData_NegativeID(t *testing.T) {
	input := `{"visualizationInfo":{"id":-5}}`
	result := processAppData(input, 999)
	assert.Equal(t, input, result, "negative ID should pass through unchanged")
}

func TestProcessAppData_ValidReplacement(t *testing.T) {
	input := `{"visualizationInfo":{"id":42,"name":"myScreen"},"otherData":{"ref":42}}`
	newDvID := int64(789)
	result := processAppData(input, newDvID)
	assert.Contains(t, result, "789")
	assert.NotContains(t, result, `"id":42`)
	assert.Contains(t, result, `"name":"myScreen"`)
}

func TestProcessAppData_ValidReplacement_MultipleRefs(t *testing.T) {
	input := `{"visualizationInfo":{"id":100},"component":{"sceneId":100,"otherRef":100}}`
	newDvID := int64(200)
	result := processAppData(input, newDvID)
	assert.NotContains(t, result, "100")
	assert.Contains(t, result, "200")
}

func TestProcessAppData_InvalidVisualizationInfoJSON(t *testing.T) {
	input := `{"visualizationInfo":"not-a-valid-object-but-long-enough"}`
	result := processAppData(input, 999)
	assert.Equal(t, input, result, "unparseable visualizationInfo should return input unchanged")
}

// ---------------------------------------------------------------------------
// normalizeJSONPayload tests
// ---------------------------------------------------------------------------

func TestNormalizeJSONPayload_Empty(t *testing.T) {
	assert.Equal(t, "", normalizeJSONPayload(json.RawMessage("")))
}

func TestNormalizeJSONPayload_Null(t *testing.T) {
	assert.Equal(t, "", normalizeJSONPayload(json.RawMessage("null")))
}

func TestNormalizeJSONPayload_NullWithSpaces(t *testing.T) {
	assert.Equal(t, "", normalizeJSONPayload(json.RawMessage("  null  ")))
}

func TestNormalizeJSONPayload_WhitespaceOnly(t *testing.T) {
	assert.Equal(t, "", normalizeJSONPayload(json.RawMessage("   ")))
}

func TestNormalizeJSONPayload_JSONString(t *testing.T) {
	raw := json.RawMessage(`"hello"`)
	result := normalizeJSONPayload(raw)
	assert.Equal(t, "hello", result)
}

func TestNormalizeJSONPayload_JSONStringComplex(t *testing.T) {
	raw := json.RawMessage(`"eyJrZXkiOiJ2YWx1ZSJ9"`)
	result := normalizeJSONPayload(raw)
	assert.Equal(t, "eyJrZXkiOiJ2YWx1ZSJ9", result)
}

func TestNormalizeJSONPayload_RawJSONObject(t *testing.T) {
	raw := json.RawMessage(`{"bg":"black","scale":100}`)
	result := normalizeJSONPayload(raw)
	assert.Equal(t, `{"bg":"black","scale":100}`, result)
}

func TestNormalizeJSONPayload_RawJSONArray(t *testing.T) {
	raw := json.RawMessage(`[1,2,3]`)
	result := normalizeJSONPayload(raw)
	assert.Equal(t, `[1,2,3]`, result)
}

// ---------------------------------------------------------------------------
// normalizeTemplateFileName tests
// ---------------------------------------------------------------------------

func TestNormalizeTemplateFileName_Empty(t *testing.T) {
	assert.Equal(t, "", normalizeTemplateFileName(json.RawMessage("")))
}

func TestNormalizeTemplateFileName_Null(t *testing.T) {
	assert.Equal(t, "", normalizeTemplateFileName(json.RawMessage("null")))
}

func TestNormalizeTemplateFileName_Whitespace(t *testing.T) {
	assert.Equal(t, "", normalizeTemplateFileName(json.RawMessage("   ")))
}

func TestNormalizeTemplateFileName_PlainString(t *testing.T) {
	raw := json.RawMessage(`"My Template"`)
	result := normalizeTemplateFileName(raw)
	assert.Equal(t, "My Template", result)
}

func TestNormalizeTemplateFileName_WrappedObject(t *testing.T) {
	raw := json.RawMessage(`{"name":"Wrapped Name"}`)
	result := normalizeTemplateFileName(raw)
	assert.Equal(t, "Wrapped Name", result, "should extract name from wrapped object")
}

func TestNormalizeTemplateFileName_RawStringNotJSON(t *testing.T) {
	raw := json.RawMessage(`just a plain name`)
	result := normalizeTemplateFileName(raw)
	assert.Equal(t, "just a plain name", result)
}

// ---------------------------------------------------------------------------
// resolveMappedID tests
// ---------------------------------------------------------------------------

func TestResolveMappedID_Found(t *testing.T) {
	idMap := map[int64]int64{10: 100, 20: 200}
	info := map[string]interface{}{"groupId": float64(10)}
	result := resolveMappedID(info, "groupId", idMap, 999)
	assert.Equal(t, int64(100), result)
}

func TestResolveMappedID_NotFound(t *testing.T) {
	idMap := map[int64]int64{10: 100}
	info := map[string]interface{}{"groupId": float64(99)}
	result := resolveMappedID(info, "groupId", idMap, 42)
	assert.Equal(t, int64(42), result, "should return fallback when key not in map")
}

func TestResolveMappedID_KeyMissing(t *testing.T) {
	idMap := map[int64]int64{10: 100}
	info := map[string]interface{}{"otherKey": float64(10)}
	result := resolveMappedID(info, "groupId", idMap, 7)
	assert.Equal(t, int64(7), result, "should return fallback when key is missing from info")
}

func TestResolveMappedID_NilValue(t *testing.T) {
	idMap := map[int64]int64{0: 100}
	info := map[string]interface{}{"groupId": nil}
	result := resolveMappedID(info, "groupId", idMap, 5)
	assert.Equal(t, int64(100), result, "nil value extracts as 0, which maps to 100 in idMap")
}

// ---------------------------------------------------------------------------
// resolveRemappedID tests
// ---------------------------------------------------------------------------

func TestResolveRemappedID_Found(t *testing.T) {
	idMap := map[int64]int64{50: 500}
	info := map[string]interface{}{"datasourceId": float64(50)}
	result := resolveRemappedID(info, "datasourceId", idMap)
	assert.Equal(t, int64(500), result)
}

func TestResolveRemappedID_NotFound(t *testing.T) {
	idMap := map[int64]int64{50: 500}
	info := map[string]interface{}{"datasourceId": float64(99)}
	result := resolveRemappedID(info, "datasourceId", idMap)
	assert.Equal(t, int64(99), result, "should return original ID when not in map")
}

func TestResolveRemappedID_KeyMissing(t *testing.T) {
	idMap := map[int64]int64{50: 500}
	info := map[string]interface{}{}
	result := resolveRemappedID(info, "datasourceId", idMap)
	assert.Equal(t, int64(0), result, "should return 0 when key missing and no oldID")
}

// ---------------------------------------------------------------------------
// parseJSONFieldFromAppData tests
// ---------------------------------------------------------------------------

func TestParseJSONFieldFromAppData_KeyPresent(t *testing.T) {
	appData := map[string]json.RawMessage{
		"items": json.RawMessage(`[{"id":1},{"id":2}]`),
	}
	result, err := parseJSONFieldFromAppData[[]map[string]interface{}](appData, "items")
	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, float64(1), result[0]["id"])
}

func TestParseJSONFieldFromAppData_KeyMissing(t *testing.T) {
	appData := map[string]json.RawMessage{
		"other": json.RawMessage(`"val"`),
	}
	result, err := parseJSONFieldFromAppData[[]map[string]interface{}](appData, "items")
	require.NoError(t, err)
	assert.Nil(t, result, "missing key should return zero value")
}

func TestParseJSONFieldFromAppData_InvalidJSON(t *testing.T) {
	appData := map[string]json.RawMessage{
		"items": json.RawMessage(`not valid json`),
	}
	_, err := parseJSONFieldFromAppData[[]map[string]interface{}](appData, "items")
	require.Error(t, err, "invalid JSON should return error")
}

func TestParseJSONFieldFromAppData_StringType(t *testing.T) {
	appData := map[string]json.RawMessage{
		"name": json.RawMessage(`"hello"`),
	}
	result, err := parseJSONFieldFromAppData[string](appData, "name")
	require.NoError(t, err)
	assert.Equal(t, "hello", result)
}

// ---------------------------------------------------------------------------
// extractInt64Value tests (lightweight helper)
// ---------------------------------------------------------------------------

func TestExtractInt64Value_Types(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int64
		ok       bool
	}{
		{"float64", float64(42), 42, true},
		{"int64", int64(100), 100, true},
		{"int", int(7), 7, true},
		{"json.Number", json.Number("999"), 999, true},
		{"string numeric", "42", 42, true},
		{"string non-numeric", "abc", 0, false},
		{"nil", nil, 0, false},
		{"bool", true, 0, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			val, ok := extractInt64Value(tc.input)
			assert.Equal(t, tc.ok, ok)
			assert.Equal(t, tc.expected, val)
		})
	}
}

// ---------------------------------------------------------------------------
// extractIntValueOrDefault tests
// ---------------------------------------------------------------------------

func TestExtractIntValueOrDefault(t *testing.T) {
	assert.Equal(t, 5, extractIntValueOrDefault(float64(5), 0))
	assert.Equal(t, 0, extractIntValueOrDefault("bad", 0))
	assert.Equal(t, 99, extractIntValueOrDefault(nil, 99))
}

// ---------------------------------------------------------------------------
// stringifyExportIDs tests
// ---------------------------------------------------------------------------

func TestStringifyExportIDs_IntTypes(t *testing.T) {
	rows := []map[string]interface{}{
		{"id": int64(42), "name": "test", "parentId": int(10)},
	}
	result := stringifyExportIDs(rows)
	assert.Equal(t, "42", result[0]["id"])
	assert.Equal(t, "test", result[0]["name"])
	assert.Equal(t, "10", result[0]["parentId"])
}

func TestStringifyExportIDs_UintTypes(t *testing.T) {
	rows := []map[string]interface{}{
		{"id": uint(42), "some_id": uint32(10), "big_id": uint64(99)},
	}
	result := stringifyExportIDs(rows)
	assert.Equal(t, "42", result[0]["id"])
	assert.Equal(t, "10", result[0]["some_id"])
	assert.Equal(t, "99", result[0]["big_id"])
}

func TestStringifyExportIDs_Int32(t *testing.T) {
	rows := []map[string]interface{}{
		{"dataId": int32(77)},
	}
	result := stringifyExportIDs(rows)
	assert.Equal(t, "77", result[0]["dataId"])
}

// ---------------------------------------------------------------------------
// Helper for closed DB test reuse (audit)
// ---------------------------------------------------------------------------
func buildClosedAuditServiceForTest(t *testing.T) *AuditService {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&audit.AuditLog{}, &audit.LoginFailure{}, &audit.AuditLogDetail{}))
	sqlDB, dbErr := db.DB()
	require.NoError(t, dbErr)
	require.NoError(t, sqlDB.Close())
	return NewAuditService(
		repository.NewAuditLogRepository(db),
		repository.NewLoginFailureRepository(db),
		repository.NewAuditLogDetailRepository(db),
	)
}
