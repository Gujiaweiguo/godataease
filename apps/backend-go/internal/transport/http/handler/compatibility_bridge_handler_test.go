package handler

import (
	"encoding/json"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type bridgeResp struct {
	Code string        `json:"code"`
	Data []interface{} `json:"data"`
}

type bridgeAnyResp struct {
	Code string                 `json:"code"`
	Data map[string]interface{} `json:"data"`
}

type bridgeCodeResp struct {
	Code string `json:"code"`
}

type bridgeFieldListResp struct {
	Code string `json:"code"`
	Data struct {
		DimensionList []map[string]interface{} `json:"dimensionList"`
		QuotaList     []map[string]interface{} `json:"quotaList"`
	} `json:"data"`
}

type fakeBridgeChartRepo struct {
	charts        map[int64]*chart.CoreChartView
	dsFields      map[int64][]*dataset.CoreDatasetTableField
	chartFields   map[int64][]*dataset.CoreDatasetTableField
	nextFieldID   int64
	fieldRegistry map[int64]*dataset.CoreDatasetTableField
}

func (r *fakeBridgeChartRepo) GetByID(id int64) (*chart.CoreChartView, error) {
	if r.charts == nil {
		return nil, nil
	}
	v := r.charts[id]
	if v == nil {
		return nil, nil
	}
	clone := *v
	return &clone, nil
}

func (r *fakeBridgeChartRepo) Update(view *chart.CoreChartView) error {
	if r.charts == nil {
		r.charts = make(map[int64]*chart.CoreChartView)
	}
	clone := *view
	r.charts[view.ID] = &clone
	return nil
}

func (r *fakeBridgeChartRepo) QueryRows(chartID int64, limit int) ([]map[string]interface{}, int64, error) {
	return []map[string]interface{}{}, 0, nil
}

func (r *fakeBridgeChartRepo) ListDatasetFieldsByGroup(datasetGroupID int64) ([]*dataset.CoreDatasetTableField, error) {
	if r.dsFields == nil {
		return []*dataset.CoreDatasetTableField{}, nil
	}
	list := r.dsFields[datasetGroupID]
	result := make([]*dataset.CoreDatasetTableField, 0, len(list))
	for _, f := range list {
		cloned := *f
		result = append(result, &cloned)
	}
	return result, nil
}

func (r *fakeBridgeChartRepo) ListDatasetFieldsByChart(chartID int64) ([]*dataset.CoreDatasetTableField, error) {
	if r.chartFields == nil {
		return []*dataset.CoreDatasetTableField{}, nil
	}
	list := r.chartFields[chartID]
	result := make([]*dataset.CoreDatasetTableField, 0, len(list))
	for _, f := range list {
		cloned := *f
		result = append(result, &cloned)
	}
	return result, nil
}

func (r *fakeBridgeChartRepo) GetDatasetFieldByID(id int64) (*dataset.CoreDatasetTableField, error) {
	if r.fieldRegistry == nil {
		return nil, nil
	}
	f := r.fieldRegistry[id]
	if f == nil {
		return nil, nil
	}
	clone := *f
	return &clone, nil
}

func (r *fakeBridgeChartRepo) CountDatasetFieldName(datasetGroupID int64, name string) (int64, error) {
	if r.fieldRegistry == nil {
		return 0, nil
	}
	var count int64
	for _, f := range r.fieldRegistry {
		if f == nil || f.Name == nil {
			continue
		}
		if f.DatasetGroupID == datasetGroupID && strings.EqualFold(*f.Name, name) {
			count++
		}
	}
	return count, nil
}

func (r *fakeBridgeChartRepo) CreateDatasetField(field *dataset.CoreDatasetTableField) error {
	if r.nextFieldID <= 0 {
		r.nextFieldID = 1000
	}
	if field.ID <= 0 {
		field.ID = r.nextFieldID
		r.nextFieldID++
	}
	if r.fieldRegistry == nil {
		r.fieldRegistry = make(map[int64]*dataset.CoreDatasetTableField)
	}
	if r.chartFields == nil {
		r.chartFields = make(map[int64][]*dataset.CoreDatasetTableField)
	}
	clone := *field
	r.fieldRegistry[field.ID] = &clone
	if clone.ChartID != nil {
		r.chartFields[*clone.ChartID] = append(r.chartFields[*clone.ChartID], &clone)
	}
	return nil
}

func (r *fakeBridgeChartRepo) UpdateDatasetFieldNames(id int64, dataeaseName string, fieldShortName string) error {
	if r.fieldRegistry != nil {
		if f, ok := r.fieldRegistry[id]; ok && f != nil {
			f.DataeaseName = &dataeaseName
			f.FieldShortName = &fieldShortName
		}
	}
	if r.chartFields != nil {
		for _, items := range r.chartFields {
			for _, item := range items {
				if item == nil || item.ID != id {
					continue
				}
				item.DataeaseName = &dataeaseName
				item.FieldShortName = &fieldShortName
			}
		}
	}
	return nil
}

func (r *fakeBridgeChartRepo) DeleteDatasetField(id int64) error {
	if r.fieldRegistry != nil {
		delete(r.fieldRegistry, id)
	}
	if r.chartFields != nil {
		for chartID, fields := range r.chartFields {
			filtered := make([]*dataset.CoreDatasetTableField, 0, len(fields))
			for _, item := range fields {
				if item == nil || item.ID == id {
					continue
				}
				filtered = append(filtered, item)
			}
			r.chartFields[chartID] = filtered
		}
	}
	return nil
}

func (r *fakeBridgeChartRepo) DeleteDatasetFieldsByChart(chartID int64) error {
	if r.chartFields != nil {
		for _, item := range r.chartFields[chartID] {
			if item == nil {
				continue
			}
			if r.fieldRegistry != nil {
				delete(r.fieldRegistry, item.ID)
			}
		}
		delete(r.chartFields, chartID)
	}
	return nil
}

func TestChartDataGetFieldDataInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, nil, &ChartHandler{})

	req := httptest.NewRequest("POST", "/chartData/getFieldData/not-number/xAxis", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	resp := bridgeResp{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "500000" {
		t.Fatalf("expected code 500000, got %s", resp.Code)
	}
}

func TestChartDataGetFieldDataFallbackEmptyWhenDatasetHandlerNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, nil, &ChartHandler{})

	req := httptest.NewRequest("POST", "/chartData/getFieldData/100/xAxis", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	resp := bridgeResp{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "000000" {
		t.Fatalf("expected code 000000, got %s", resp.Code)
	}
	if len(resp.Data) != 0 {
		t.Fatalf("expected empty data, got %#v", resp.Data)
	}
}

func TestChartDataGetDrillFieldDataFallbackEmptyWhenDatasetHandlerNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, nil, &ChartHandler{})

	req := httptest.NewRequest("POST", "/chartData/getDrillFieldData/100", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	resp := bridgeResp{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "000000" {
		t.Fatalf("expected code 000000, got %s", resp.Code)
	}
	if len(resp.Data) != 0 {
		t.Fatalf("expected empty data, got %#v", resp.Data)
	}
}

func TestChartSaveRouteUpdatesCoreFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &fakeBridgeChartRepo{charts: map[int64]*chart.CoreChartView{}}
	title := "old"
	repo.charts[101] = &chart.CoreChartView{ID: 101, Title: &title}
	chartHandler := NewChartHandler(service.NewChartService(repo))

	r := gin.New()
	RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, nil, chartHandler)

	reqBody := `{"id":101,"title":"new-title","resultMode":"all"}`
	req := httptest.NewRequest("POST", "/chart/save", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	resp := bridgeAnyResp{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "000000" {
		t.Fatalf("expected code 000000, got %s", resp.Code)
	}
	if resp.Data["title"] != "new-title" {
		t.Fatalf("expected title updated, got %#v", resp.Data["title"])
	}
}

func TestChartListByDQRouteReturnsDimensionAndQuota(t *testing.T) {
	gin.SetMode(gin.TestMode)
	checked := true
	groupD := "d"
	typeD := "VARCHAR"
	nameD := "region"
	originD := "region"
	dataeaseD := "region"
	deTypeD := 0

	repo := &fakeBridgeChartRepo{
		charts:   map[int64]*chart.CoreChartView{},
		dsFields: map[int64][]*dataset.CoreDatasetTableField{},
	}
	repo.dsFields[11] = []*dataset.CoreDatasetTableField{{
		ID:             1,
		DatasetGroupID: 11,
		Name:           &nameD,
		OriginName:     &originD,
		DataeaseName:   &dataeaseD,
		GroupType:      &groupD,
		Type:           &typeD,
		DeType:         &deTypeD,
		Checked:        &checked,
	}}
	chartHandler := NewChartHandler(service.NewChartService(repo))

	r := gin.New()
	RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, nil, chartHandler)

	req := httptest.NewRequest("POST", "/chart/listByDQ/11/9", strings.NewReader(`{"type":"bar"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	resp := bridgeFieldListResp{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != "000000" {
		t.Fatalf("expected code 000000, got %s", resp.Code)
	}
	if len(resp.Data.DimensionList) != 1 {
		t.Fatalf("expected 1 dimension field, got %d", len(resp.Data.DimensionList))
	}
	if len(resp.Data.QuotaList) == 0 {
		t.Fatal("expected quota list contains pseudo count field")
	}
}

func TestApiAliasChartSaveAndListByDQ(t *testing.T) {
	gin.SetMode(gin.TestMode)
	checked := true
	groupD := "d"
	typeD := "VARCHAR"
	nameD := "city"
	originD := "city"
	dataeaseD := "city"
	deTypeD := 0
	title := "origin"

	repo := &fakeBridgeChartRepo{
		charts: map[int64]*chart.CoreChartView{201: {ID: 201, Title: &title}},
		dsFields: map[int64][]*dataset.CoreDatasetTableField{21: {{
			ID:             2,
			DatasetGroupID: 21,
			Name:           &nameD,
			OriginName:     &originD,
			DataeaseName:   &dataeaseD,
			GroupType:      &groupD,
			Type:           &typeD,
			DeType:         &deTypeD,
			Checked:        &checked,
		}}},
	}
	chartHandler := NewChartHandler(service.NewChartService(repo))

	r := gin.New()
	api := r.Group("/api")
	RegisterCompatibilityBridgeRoutes(api, nil, nil, nil, nil, chartHandler)

	saveReq := httptest.NewRequest("POST", "/api/chart/save", strings.NewReader(`{"id":201,"title":"alias-title"}`))
	saveReq.Header.Set("Content-Type", "application/json")
	saveW := httptest.NewRecorder()
	r.ServeHTTP(saveW, saveReq)
	if saveW.Code != 200 {
		t.Fatalf("expected status 200, got %d", saveW.Code)
	}
	saveResp := bridgeAnyResp{}
	if err := json.Unmarshal(saveW.Body.Bytes(), &saveResp); err != nil {
		t.Fatalf("unmarshal save response failed: %v", err)
	}
	if saveResp.Code != "000000" {
		t.Fatalf("expected save code 000000, got %s", saveResp.Code)
	}
	if saveResp.Data["title"] != "alias-title" {
		t.Fatalf("expected alias save title updated, got %#v", saveResp.Data["title"])
	}

	listReq := httptest.NewRequest("POST", "/api/chart/listByDQ/21/201", strings.NewReader(`{"type":"bar"}`))
	listReq.Header.Set("Content-Type", "application/json")
	listW := httptest.NewRecorder()
	r.ServeHTTP(listW, listReq)
	if listW.Code != 200 {
		t.Fatalf("expected status 200, got %d", listW.Code)
	}
	listResp := bridgeFieldListResp{}
	if err := json.Unmarshal(listW.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("unmarshal list response failed: %v", err)
	}
	if listResp.Code != "000000" {
		t.Fatalf("expected list code 000000, got %s", listResp.Code)
	}
	if len(listResp.Data.DimensionList) != 1 {
		t.Fatalf("expected alias list dimension size 1, got %d", len(listResp.Data.DimensionList))
	}
}

func TestChartCopyAndDeleteFieldRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	name := "sales"
	origin := "sales"
	dataeaseName := "sales"
	groupType := "q"
	typeName := "DECIMAL"
	deType := 3
	checked := true

	repo := &fakeBridgeChartRepo{
		charts:        map[int64]*chart.CoreChartView{},
		chartFields:   map[int64][]*dataset.CoreDatasetTableField{},
		fieldRegistry: map[int64]*dataset.CoreDatasetTableField{},
		nextFieldID:   2000,
	}
	repo.fieldRegistry[10] = &dataset.CoreDatasetTableField{
		ID:             10,
		DatasetGroupID: 30,
		Name:           &name,
		OriginName:     &origin,
		DataeaseName:   &dataeaseName,
		GroupType:      &groupType,
		Type:           &typeName,
		DeType:         &deType,
		Checked:        &checked,
	}
	chartHandler := NewChartHandler(service.NewChartService(repo))

	r := gin.New()
	RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, nil, chartHandler)

	copyReq := httptest.NewRequest("POST", "/chart/copyField/10/300", strings.NewReader("{}"))
	copyReq.Header.Set("Content-Type", "application/json")
	copyW := httptest.NewRecorder()
	r.ServeHTTP(copyW, copyReq)
	if copyW.Code != 200 {
		t.Fatalf("expected status 200, got %d", copyW.Code)
	}
	copyResp := bridgeCodeResp{}
	if err := json.Unmarshal(copyW.Body.Bytes(), &copyResp); err != nil {
		t.Fatalf("unmarshal copy response failed: %v", err)
	}
	if copyResp.Code != "000000" {
		t.Fatalf("expected copy code 000000, got %s", copyResp.Code)
	}
	if len(repo.chartFields[300]) != 1 {
		t.Fatalf("expected 1 copied field for chart 300, got %d", len(repo.chartFields[300]))
	}
	copiedID := repo.chartFields[300][0].ID

	delReq := httptest.NewRequest("POST", "/chart/deleteField/"+strconv.FormatInt(copiedID, 10), strings.NewReader("{}"))
	delReq.Header.Set("Content-Type", "application/json")
	delW := httptest.NewRecorder()
	r.ServeHTTP(delW, delReq)
	if delW.Code != 200 {
		t.Fatalf("expected status 200, got %d", delW.Code)
	}
	delResp := bridgeCodeResp{}
	if err := json.Unmarshal(delW.Body.Bytes(), &delResp); err != nil {
		t.Fatalf("unmarshal delete response failed: %v", err)
	}
	if delResp.Code != "000000" {
		t.Fatalf("expected delete code 000000, got %s", delResp.Code)
	}
	if len(repo.chartFields[300]) != 0 {
		t.Fatalf("expected copied field deleted, remaining %d", len(repo.chartFields[300]))
	}
}

func TestApiAliasChartCopyAndDeleteFieldByChart(t *testing.T) {
	gin.SetMode(gin.TestMode)
	name := "profit"
	origin := "profit"
	dataeaseName := "profit"
	groupType := "q"
	typeName := "DECIMAL"
	deType := 3
	checked := true

	repo := &fakeBridgeChartRepo{
		charts:        map[int64]*chart.CoreChartView{},
		chartFields:   map[int64][]*dataset.CoreDatasetTableField{},
		fieldRegistry: map[int64]*dataset.CoreDatasetTableField{},
		nextFieldID:   3000,
	}
	repo.fieldRegistry[11] = &dataset.CoreDatasetTableField{
		ID:             11,
		DatasetGroupID: 31,
		Name:           &name,
		OriginName:     &origin,
		DataeaseName:   &dataeaseName,
		GroupType:      &groupType,
		Type:           &typeName,
		DeType:         &deType,
		Checked:        &checked,
	}
	chartHandler := NewChartHandler(service.NewChartService(repo))

	r := gin.New()
	api := r.Group("/api")
	RegisterCompatibilityBridgeRoutes(api, nil, nil, nil, nil, chartHandler)

	copyReq := httptest.NewRequest("POST", "/api/chart/copyField/11/400", strings.NewReader("{}"))
	copyReq.Header.Set("Content-Type", "application/json")
	copyW := httptest.NewRecorder()
	r.ServeHTTP(copyW, copyReq)
	if copyW.Code != 200 {
		t.Fatalf("expected status 200, got %d", copyW.Code)
	}
	copyResp := bridgeCodeResp{}
	if err := json.Unmarshal(copyW.Body.Bytes(), &copyResp); err != nil {
		t.Fatalf("unmarshal copy response failed: %v", err)
	}
	if copyResp.Code != "000000" {
		t.Fatalf("expected copy code 000000, got %s", copyResp.Code)
	}
	if len(repo.chartFields[400]) != 1 {
		t.Fatalf("expected alias copied field count 1, got %d", len(repo.chartFields[400]))
	}

	delByChartReq := httptest.NewRequest("POST", "/api/chart/deleteFieldByChart/400", strings.NewReader("{}"))
	delByChartReq.Header.Set("Content-Type", "application/json")
	delByChartW := httptest.NewRecorder()
	r.ServeHTTP(delByChartW, delByChartReq)
	if delByChartW.Code != 200 {
		t.Fatalf("expected status 200, got %d", delByChartW.Code)
	}
	delByChartResp := bridgeCodeResp{}
	if err := json.Unmarshal(delByChartW.Body.Bytes(), &delByChartResp); err != nil {
		t.Fatalf("unmarshal delete-by-chart response failed: %v", err)
	}
	if delByChartResp.Code != "000000" {
		t.Fatalf("expected delete-by-chart code 000000, got %s", delByChartResp.Code)
	}
	if _, ok := repo.chartFields[400]; ok {
		t.Fatal("expected chart fields deleted by chart id")
	}
}

func TestOldPathDatasourceList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("POST", "/datasource/list", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Logf("Route /datasource/list registered, status: %d", w.Code)
	}
}

func TestOldPathDatasetTree(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("POST", "/datasetTree/tree", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Logf("Route /datasetTree/tree registered, status: %d", w.Code)
	}
}

func TestApiAliasDatasourceList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api")
	RegisterCompatibilityBridgeRoutes(api, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("POST", "/api/datasource/list", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Logf("Route /api/datasource/list registered, status: %d", w.Code)
	}
}

func TestApiAliasDatasetTree(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api")
	RegisterCompatibilityBridgeRoutes(api, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("POST", "/api/datasetTree/tree", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Logf("Route /api/datasetTree/tree registered, status: %d", w.Code)
	}
}

type bridgePaginationResp struct {
	Code    string                   `json:"code"`
	Data    []map[string]interface{} `json:"data"`
	Total   int64                    `json:"total"`
	Current int                      `json:"current"`
	Size    int                      `json:"size"`
}

func TestPaginationResponseFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("POST", "/chart/listByDQ/1/1", strings.NewReader(`{"type":"bar"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code == 200 {
		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err == nil {
			if code, ok := resp["code"]; ok {
				if _, isString := code.(string); !isString {
					t.Errorf("expected code field to be string, got %T", code)
				}
			}
		}
	}
}

func TestErrorResponseFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &fakeBridgeChartRepo{charts: map[int64]*chart.CoreChartView{}}
	chartHandler := NewChartHandler(service.NewChartService(repo))

	r := gin.New()
	RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, nil, chartHandler)

	req := httptest.NewRequest("POST", "/chartData/getFieldData/invalid/xAxis", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	resp := bridgeCodeResp{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}

	if resp.Code != "500000" {
		t.Fatalf("expected error code '500000', got '%s'", resp.Code)
	}
}

func TestOldPathChartSaveWithNilHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterCompatibilityBridgeRoutes(r, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("POST", "/chart/save", strings.NewReader(`{"id":1}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Logf("Route /chart/save registered, status: %d", w.Code)
	}
}

func TestApiAliasChartDataGetData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api")
	RegisterCompatibilityBridgeRoutes(api, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("POST", "/api/chartData/getData", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Logf("Route /api/chartData/getData registered, status: %d", w.Code)
	}
}
