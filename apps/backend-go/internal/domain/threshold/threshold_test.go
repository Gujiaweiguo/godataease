package threshold

import (
	"encoding/json"
	"testing"
)

// --- BaseReciDTO ---

func TestBaseReciDTO_JSONRoundTrip(t *testing.T) {
	original := BaseReciDTO{
		ReciFlagList:       []int{1, 2, 3},
		UIDList:            []string{"u1", "u2"},
		RIDList:            []string{"r1"},
		EmailList:          []string{"a@b.com"},
		LarkGroupList:      []string{"lg1"},
		LarksuiteGroupList: []string{"lsg1"},
		WebhookList:        []string{"wh1"},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal BaseReciDTO: %v", err)
	}

	var decoded BaseReciDTO
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal BaseReciDTO: %v", err)
	}

	assertSliceEqual(t, "ReciFlagList", original.ReciFlagList, decoded.ReciFlagList)
	assertStringSliceEqual(t, "UIDList", original.UIDList, decoded.UIDList)
	assertStringSliceEqual(t, "RIDList", original.RIDList, decoded.RIDList)
	assertStringSliceEqual(t, "EmailList", original.EmailList, decoded.EmailList)
	assertStringSliceEqual(t, "LarkGroupList", original.LarkGroupList, decoded.LarkGroupList)
	assertStringSliceEqual(t, "LarksuiteGroupList", original.LarksuiteGroupList, decoded.LarksuiteGroupList)
	assertStringSliceEqual(t, "WebhookList", original.WebhookList, decoded.WebhookList)
}

func TestBaseReciDTO_Empty(t *testing.T) {
	original := BaseReciDTO{}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal empty BaseReciDTO: %v", err)
	}

	// Empty slices should marshal as null or [] depending on Go encoding
	var decoded BaseReciDTO
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal empty BaseReciDTO: %v", err)
	}

	if len(decoded.ReciFlagList) != 0 || len(decoded.UIDList) != 0 {
		t.Errorf("expected empty slices after round-trip, got reciFlags=%v uidList=%v",
			decoded.ReciFlagList, decoded.UIDList)
	}
}

func TestBaseReciDTO_JSONKeys(t *testing.T) {
	dto := BaseReciDTO{EmailList: []string{"x@y.com"}}
	data, err := json.Marshal(dto)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}

	expectedKeys := []string{
		"reciFlagList", "uidList", "ridList",
		"emailList", "larkGroupList", "larksuiteGroupList", "webhookList",
	}
	for _, key := range expectedKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("missing JSON key: %s", key)
		}
	}
}

// --- CreateRequest ---

func TestCreateRequest_JSONRoundTrip(t *testing.T) {
	enable := true
	rateType := 1
	msgType := 2
	repeatSend := true
	showField := false

	req := CreateRequest{
		BaseReciDTO: BaseReciDTO{
			UIDList:   []string{"u1"},
			EmailList: []string{"test@test.com"},
		},
		ID:             100,
		Name:           "test threshold",
		Enable:         &enable,
		RateType:       &rateType,
		RateValue:      "10",
		ResourceID:     200,
		ResourceType:   "panel",
		ChartID:        300,
		ChartType:      "bar",
		ThresholdRules: `[{"field":"value","term":"gt","value":"100"}]`,
		MsgType:        &msgType,
		MsgTitle:       "Alert",
		MsgContent:     "Value exceeded",
		RepeatSend:     &repeatSend,
		ShowFieldValue: &showField,
		ResourceTable:  "chart_view",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal CreateRequest: %v", err)
	}

	var decoded CreateRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal CreateRequest: %v", err)
	}

	if decoded.ID != req.ID {
		t.Errorf("ID: got %d, want %d", decoded.ID, req.ID)
	}
	if decoded.Name != req.Name {
		t.Errorf("Name: got %q, want %q", decoded.Name, req.Name)
	}
	if decoded.Enable == nil || *decoded.Enable != enable {
		t.Errorf("Enable: got %v, want %v", decoded.Enable, enable)
	}
	if decoded.RateType == nil || *decoded.RateType != rateType {
		t.Errorf("RateType: got %v, want %d", decoded.RateType, rateType)
	}
	if decoded.RateValue != req.RateValue {
		t.Errorf("RateValue: got %q, want %q", decoded.RateValue, req.RateValue)
	}
	if decoded.ResourceID != req.ResourceID {
		t.Errorf("ResourceID: got %d, want %d", decoded.ResourceID, req.ResourceID)
	}
	if decoded.ResourceType != req.ResourceType {
		t.Errorf("ResourceType: got %q, want %q", decoded.ResourceType, req.ResourceType)
	}
	if decoded.ChartID != req.ChartID {
		t.Errorf("ChartID: got %d, want %d", decoded.ChartID, req.ChartID)
	}
	if decoded.ChartType != req.ChartType {
		t.Errorf("ChartType: got %q, want %q", decoded.ChartType, req.ChartType)
	}
	if decoded.ThresholdRules != req.ThresholdRules {
		t.Errorf("ThresholdRules: got %q, want %q", decoded.ThresholdRules, req.ThresholdRules)
	}
	if decoded.MsgType == nil || *decoded.MsgType != msgType {
		t.Errorf("MsgType: got %v, want %d", decoded.MsgType, msgType)
	}
	if decoded.MsgTitle != req.MsgTitle {
		t.Errorf("MsgTitle: got %q, want %q", decoded.MsgTitle, req.MsgTitle)
	}
	if decoded.MsgContent != req.MsgContent {
		t.Errorf("MsgContent: got %q, want %q", decoded.MsgContent, req.MsgContent)
	}
	if decoded.RepeatSend == nil || *decoded.RepeatSend != repeatSend {
		t.Errorf("RepeatSend: got %v, want %v", decoded.RepeatSend, repeatSend)
	}
	if decoded.ShowFieldValue == nil || *decoded.ShowFieldValue != showField {
		t.Errorf("ShowFieldValue: got %v, want %v", decoded.ShowFieldValue, showField)
	}
	if decoded.ResourceTable != req.ResourceTable {
		t.Errorf("ResourceTable: got %q, want %q", decoded.ResourceTable, req.ResourceTable)
	}
}

func TestCreateRequest_NilPointerFields(t *testing.T) {
	req := CreateRequest{
		Name: "minimal",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal minimal CreateRequest: %v", err)
	}

	var decoded CreateRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal minimal CreateRequest: %v", err)
	}

	if decoded.Enable != nil {
		t.Errorf("Enable: expected nil, got %v", *decoded.Enable)
	}
	if decoded.RateType != nil {
		t.Errorf("RateType: expected nil, got %v", *decoded.RateType)
	}
	if decoded.MsgType != nil {
		t.Errorf("MsgType: expected nil, got %v", *decoded.MsgType)
	}
	if decoded.RepeatSend != nil {
		t.Errorf("RepeatSend: expected nil, got %v", *decoded.RepeatSend)
	}
	if decoded.ShowFieldValue != nil {
		t.Errorf("ShowFieldValue: expected nil, got %v", *decoded.ShowFieldValue)
	}
}

func TestCreateRequest_EmbedsBaseReciDTO(t *testing.T) {
	req := CreateRequest{
		BaseReciDTO: BaseReciDTO{
			ReciFlagList: []int{1},
			UIDList:      []string{"user-1"},
		},
	}
	if len(req.ReciFlagList) != 1 || req.ReciFlagList[0] != 1 {
		t.Errorf("embedded ReciFlagList not accessible: %v", req.ReciFlagList)
	}
	if len(req.UIDList) != 1 || req.UIDList[0] != "user-1" {
		t.Errorf("embedded UIDList not accessible: %v", req.UIDList)
	}
}

// --- GridRequest ---

func TestGridRequest_JSONRoundTrip(t *testing.T) {
	chartID := int64(42)
	req := GridRequest{
		Keyword:          "alert",
		ResourceTable:    "chart_view",
		ResourceTypeList: []string{"panel", "screen"},
		StatusList:       []int{1, 0},
		EnableList:       []int{1},
		TimeList:         []int64{1000, 2000},
		ChartID:          &chartID,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal GridRequest: %v", err)
	}

	var decoded GridRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal GridRequest: %v", err)
	}

	if decoded.Keyword != req.Keyword {
		t.Errorf("Keyword: got %q, want %q", decoded.Keyword, req.Keyword)
	}
	if decoded.ChartID == nil || *decoded.ChartID != chartID {
		t.Errorf("ChartID: got %v, want %d", decoded.ChartID, chartID)
	}
	assertStringSliceEqual(t, "ResourceTypeList", req.ResourceTypeList, decoded.ResourceTypeList)
}

func TestGridRequest_NilChartID(t *testing.T) {
	req := GridRequest{
		Keyword:       "test",
		ResourceTable: "chart_view",
		ChartID:       nil,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// omitempty should exclude chartId when nil
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal raw: %v", err)
	}
	if _, exists := raw["chartId"]; exists {
		t.Error("chartId should be omitted when nil due to omitempty tag")
	}
}

func TestGridRequest_EmptySlices(t *testing.T) {
	req := GridRequest{}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal empty GridRequest: %v", err)
	}
	var decoded GridRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal empty GridRequest: %v", err)
	}
	if len(decoded.ResourceTypeList) != 0 || len(decoded.StatusList) != 0 {
		t.Error("expected empty slices")
	}
}

// --- GridVO ---

func TestGridVO_JSONRoundTrip(t *testing.T) {
	vo := GridVO{
		ID:           1,
		Name:         "test",
		ResourceID:   100,
		ResourceType: "panel",
		ResourceName: "My Panel",
		ChartID:      200,
		ChartType:    "line",
		ChartName:    "My Chart",
		Status:       true,
		Enable:       false,
		Creator:      999,
		CreateName:   "admin",
		CreateTime:   1700000000,
	}

	data, err := json.Marshal(vo)
	if err != nil {
		t.Fatalf("marshal GridVO: %v", err)
	}

	var decoded GridVO
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal GridVO: %v", err)
	}

	if decoded != vo {
		t.Errorf("GridVO round-trip mismatch:\ngot:  %+v\nwant: %+v", decoded, vo)
	}
}

func TestGridVO_ZeroValues(t *testing.T) {
	vo := GridVO{}
	data, err := json.Marshal(vo)
	if err != nil {
		t.Fatalf("marshal zero GridVO: %v", err)
	}
	var decoded GridVO
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal zero GridVO: %v", err)
	}
	if decoded.Status != false || decoded.Enable != false {
		t.Errorf("expected false booleans, got status=%v enable=%v", decoded.Status, decoded.Enable)
	}
}

// --- SwitchRequest ---

func TestSwitchRequest_JSONRoundTrip(t *testing.T) {
	enable := true
	req := SwitchRequest{
		ID:            55,
		Enable:        &enable,
		ResourceTable: "chart_view",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal SwitchRequest: %v", err)
	}

	var decoded SwitchRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal SwitchRequest: %v", err)
	}

	if decoded.ID != req.ID {
		t.Errorf("ID: got %d, want %d", decoded.ID, req.ID)
	}
	if decoded.Enable == nil || *decoded.Enable != enable {
		t.Errorf("Enable: got %v, want %v", decoded.Enable, enable)
	}
	if decoded.ResourceTable != req.ResourceTable {
		t.Errorf("ResourceTable: got %q, want %q", decoded.ResourceTable, req.ResourceTable)
	}
}

func TestSwitchRequest_NilEnable(t *testing.T) {
	req := SwitchRequest{ID: 1, ResourceTable: "chart_view"}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded SwitchRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Enable != nil {
		t.Errorf("Enable should be nil, got %v", *decoded.Enable)
	}
}

// --- BatchReciRequest ---

func TestBatchReciRequest_JSONRoundTrip(t *testing.T) {
	req := BatchReciRequest{
		BaseReciDTO: BaseReciDTO{
			UIDList:   []string{"u1", "u2"},
			EmailList: []string{"a@b.com"},
		},
		IDList: []int64{1, 2, 3},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal BatchReciRequest: %v", err)
	}

	var decoded BatchReciRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal BatchReciRequest: %v", err)
	}

	if len(decoded.IDList) != 3 {
		t.Fatalf("IDList length: got %d, want 3", len(decoded.IDList))
	}
	for i, id := range decoded.IDList {
		if id != req.IDList[i] {
			t.Errorf("IDList[%d]: got %d, want %d", i, id, req.IDList[i])
		}
	}
	assertStringSliceEqual(t, "UIDList", req.UIDList, decoded.UIDList)
}

// --- PreviewRequest ---

func TestPreviewRequest_JSONRoundTrip(t *testing.T) {
	showField := true
	req := PreviewRequest{
		ChartID:        100,
		ThresholdRules: `[{"logic":"and","items":[]}]`,
		MsgContent:     "test content",
		ShowFieldValue: &showField,
		ThresholdLimit: 10,
		ResourceTable:  "chart_view",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal PreviewRequest: %v", err)
	}

	var decoded PreviewRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal PreviewRequest: %v", err)
	}

	if decoded.ChartID != req.ChartID {
		t.Errorf("ChartID: got %d, want %d", decoded.ChartID, req.ChartID)
	}
	if decoded.ThresholdLimit != req.ThresholdLimit {
		t.Errorf("ThresholdLimit: got %d, want %d", decoded.ThresholdLimit, req.ThresholdLimit)
	}
	if decoded.ThresholdRules != req.ThresholdRules {
		t.Errorf("ThresholdRules mismatch")
	}
	if decoded.ShowFieldValue == nil || *decoded.ShowFieldValue != showField {
		t.Errorf("ShowFieldValue: got %v, want %v", decoded.ShowFieldValue, showField)
	}
}

func TestPreviewRequest_NilShowFieldValue(t *testing.T) {
	req := PreviewRequest{ChartID: 1}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded PreviewRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.ShowFieldValue != nil {
		t.Errorf("ShowFieldValue should be nil, got %v", *decoded.ShowFieldValue)
	}
}

// --- PreviewResponse ---

func TestPreviewResponse_JSONRoundTrip(t *testing.T) {
	resp := PreviewResponse{Content: "alert triggered"}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal PreviewResponse: %v", err)
	}
	var decoded PreviewResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal PreviewResponse: %v", err)
	}
	if decoded.Content != resp.Content {
		t.Errorf("Content: got %q, want %q", decoded.Content, resp.Content)
	}
}

// --- InstanceRequest ---

func TestInstanceRequest_JSONRoundTrip(t *testing.T) {
	thresholdID := int64(99)
	req := InstanceRequest{
		Keyword:     "error",
		ThresholdID: &thresholdID,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal InstanceRequest: %v", err)
	}

	var decoded InstanceRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal InstanceRequest: %v", err)
	}

	if decoded.Keyword != req.Keyword {
		t.Errorf("Keyword: got %q, want %q", decoded.Keyword, req.Keyword)
	}
	if decoded.ThresholdID == nil || *decoded.ThresholdID != thresholdID {
		t.Errorf("ThresholdID: got %v, want %d", decoded.ThresholdID, thresholdID)
	}
}

func TestInstanceRequest_NilThresholdID(t *testing.T) {
	req := InstanceRequest{Keyword: "test"}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal raw: %v", err)
	}
	if _, exists := raw["thresholdId"]; exists {
		t.Error("thresholdId should be omitted when nil")
	}
}

// --- InstanceVO ---

func TestInstanceVO_JSONRoundTrip(t *testing.T) {
	vo := InstanceVO{
		ID:       1,
		TaskID:   2,
		Name:     "threshold-1",
		ExecTime: 1700000000,
		Status:   true,
		Content:  "value exceeded",
		Msg:      "alert sent",
	}

	data, err := json.Marshal(vo)
	if err != nil {
		t.Fatalf("marshal InstanceVO: %v", err)
	}

	var decoded InstanceVO
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal InstanceVO: %v", err)
	}

	if decoded != vo {
		t.Errorf("InstanceVO round-trip mismatch:\ngot:  %+v\nwant: %+v", decoded, vo)
	}
}

// --- PageResult ---

func TestPageResult_JSONRoundTrip(t *testing.T) {
	pr := PageResult{
		List:    []string{"a", "b"},
		Total:   100,
		Current: 1,
		Size:    10,
	}

	data, err := json.Marshal(pr)
	if err != nil {
		t.Fatalf("marshal PageResult: %v", err)
	}

	var decoded PageResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal PageResult: %v", err)
	}

	if decoded.Total != pr.Total {
		t.Errorf("Total: got %d, want %d", decoded.Total, pr.Total)
	}
	if decoded.Current != pr.Current {
		t.Errorf("Current: got %d, want %d", decoded.Current, pr.Current)
	}
	if decoded.Size != pr.Size {
		t.Errorf("Size: got %d, want %d", decoded.Size, pr.Size)
	}
}

func TestPageResult_ListAsAny(t *testing.T) {
	// Verify List field (any type) can hold different data shapes
	tests := []struct {
		name string
		list any
	}{
		{"nil list", nil},
		{"string slice", []string{"x", "y"}},
		{"int slice", []int{1, 2, 3}},
		{"empty slice", []string{}},
		{"single object", map[string]any{"key": "value"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := PageResult{List: tt.list, Total: 1, Current: 1, Size: 10}
			data, err := json.Marshal(pr)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var decoded PageResult
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if decoded.Total != 1 {
				t.Errorf("Total mismatch: got %d", decoded.Total)
			}
		})
	}
}

// --- JSON deserialization from external input ---

func TestCreateRequest_DeserializeFromJSON(t *testing.T) {
	input := `{
		"id": 1,
		"name": "CPU Alert",
		"enable": true,
		"rateType": 0,
		"rateValue": "5",
		"resourceId": 10,
		"resourceType": "panel",
		"chartId": 20,
		"chartType": "gauge",
		"thresholdRules": "[{\"field\":\"cpu\",\"term\":\"gt\",\"value\":\"90\"}]",
		"msgType": 0,
		"msgTitle": "CPU High",
		"msgContent": "CPU usage above 90%",
		"repeatSend": false,
		"showFieldValue": true,
		"resourceTable": "chart_view",
		"reciFlagList": [1,2],
		"uidList": ["user1"],
		"emailList": ["admin@test.com"]
	}`

	var req CreateRequest
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		t.Fatalf("unmarshal from JSON string: %v", err)
	}

	if req.ID != 1 {
		t.Errorf("ID: got %d, want 1", req.ID)
	}
	if req.Name != "CPU Alert" {
		t.Errorf("Name: got %q, want 'CPU Alert'", req.Name)
	}
	if req.Enable == nil || !*req.Enable {
		t.Error("Enable should be true")
	}
	if req.RateValue != "5" {
		t.Errorf("RateValue: got %q, want '5'", req.RateValue)
	}
	if len(req.ReciFlagList) != 2 {
		t.Errorf("ReciFlagList length: got %d, want 2", len(req.ReciFlagList))
	}
	if len(req.EmailList) != 1 || req.EmailList[0] != "admin@test.com" {
		t.Errorf("EmailList: got %v", req.EmailList)
	}
}

func TestGridRequest_DeserializeFromJSON(t *testing.T) {
	input := `{
		"keyword": "test",
		"resourceTable": "chart_view",
		"resourceTypeList": ["panel"],
		"statusList": [1],
		"enableList": [0,1],
		"timeList": [100,200]
	}`

	var req GridRequest
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if req.Keyword != "test" {
		t.Errorf("Keyword: got %q", req.Keyword)
	}
	if req.ChartID != nil {
		t.Errorf("ChartID should be nil when omitted, got %d", *req.ChartID)
	}
	assertStringSliceEqual(t, "ResourceTypeList", []string{"panel"}, req.ResourceTypeList)
}

// --- Helpers ---

func assertSliceEqual(t *testing.T, name string, want, got []int) {
	t.Helper()
	if len(want) != len(got) {
		t.Errorf("%s length: got %d, want %d", name, len(got), len(want))
		return
	}
	for i := range want {
		if want[i] != got[i] {
			t.Errorf("%s[%d]: got %d, want %d", name, i, got[i], want[i])
		}
	}
}

func assertStringSliceEqual(t *testing.T, name string, want, got []string) {
	t.Helper()
	if len(want) != len(got) {
		t.Errorf("%s length: got %d, want %d", name, len(got), len(want))
		return
	}
	for i := range want {
		if want[i] != got[i] {
			t.Errorf("%s[%d]: got %q, want %q", name, i, got[i], want[i])
		}
	}
}
