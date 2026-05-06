package service

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/domain/permission"
)

const chartDefaultFieldName = "field"

type ChartRepository interface {
	GetByID(id int64) (*chart.CoreChartView, error)
	Update(view *chart.CoreChartView) error
	QueryRows(chartID int64, limit int) ([]map[string]interface{}, int64, error)
	QueryViewOption(resourceId int64) ([]chart.ViewSelectorVO, error)
	GetVisualizationComponentData(resourceId int64) (string, error)
	QueryChartBaseInfo(id int64, resourceTable string) (*chart.ChartBaseVO, error)
	ListDatasetFieldsByGroup(datasetGroupID int64) ([]*dataset.CoreDatasetTableField, error)
	ListDatasetFieldsByChart(chartID int64) ([]*dataset.CoreDatasetTableField, error)
	GetDatasetFieldByID(id int64) (*dataset.CoreDatasetTableField, error)
	CountDatasetFieldName(datasetGroupID int64, name string) (int64, error)
	CreateDatasetField(field *dataset.CoreDatasetTableField) error
	UpdateDatasetFieldNames(id int64, dataeaseName string, fieldShortName string) error
	DeleteDatasetField(id int64) error
	DeleteDatasetFieldsByChart(chartID int64) error
}

type ChartService struct {
	repo                    ChartRepository
	rowPermissionService    *RowPermissionService
	columnPermissionService *ColumnPermissionService
}

type permissionAwareChartRepo interface {
	QueryRowsWithFilter(chartID int64, selectColumns string, whereClause string, whereArgs []interface{}, limit int) ([]map[string]interface{}, int64, error)
	GetDatasetGroupIDByChartID(chartID int64) (int64, error)
}

func NewChartService(repo ChartRepository) *ChartService {
	return &ChartService{repo: repo}
}

func (s *ChartService) SetRowPermissionService(rowPermSvc *RowPermissionService) {
	s.rowPermissionService = rowPermSvc
}

func (s *ChartService) SetColumnPermissionService(columnPermSvc *ColumnPermissionService) {
	s.columnPermissionService = columnPermSvc
}

func (s *ChartService) Query(req *chart.ChartQueryRequest) (*chart.CoreChartView, error) {
	return s.repo.GetByID(req.ID)
}

func (s *ChartService) ViewOption(resourceId int64) ([]chart.ViewSelectorVO, error) {
	views, err := s.repo.QueryViewOption(resourceId)
	if err != nil {
		return nil, err
	}
	componentData, err := s.repo.GetVisualizationComponentData(resourceId)
	if err != nil || componentData == "" {
		return views, nil
	}
	filtered := make([]chart.ViewSelectorVO, 0, len(views))
	for _, v := range views {
		if strings.Contains(componentData, strconv.FormatInt(v.ID, 10)) {
			filtered = append(filtered, v)
		}
	}
	return filtered, nil
}

func (s *ChartService) ChartBaseInfo(id int64, resourceTable string) (*chart.ChartBaseVO, error) {
	return s.repo.QueryChartBaseInfo(id, resourceTable)
}

func (s *ChartService) QueryData(req *chart.ChartDataRequest) (*chart.ChartDataResponse, error) {
	limit := 100
	if req.ResultCount != nil && *req.ResultCount > 0 {
		limit = *req.ResultCount
	}

	rows, total, err := s.repo.QueryRows(req.ID, limit)
	if err != nil {
		return nil, err
	}

	columns := make([]string, 0)
	if len(rows) > 0 {
		for col := range rows[0] {
			columns = append(columns, col)
		}
		sort.Strings(columns)
	}

	return s.buildChartDataResponse(req, rows, columns, total)
}

func (s *ChartService) QueryDataWithPermission(req *chart.ChartDataRequest, userID int64) (*chart.ChartDataResponse, error) {
	permissionRepo, ok := s.repo.(permissionAwareChartRepo)
	if !ok || userID <= 0 {
		return s.QueryData(req)
	}

	datasetGroupID, err := permissionRepo.GetDatasetGroupIDByChartID(req.ID)
	if err != nil {
		return nil, err
	}

	limit := 100
	if req.ResultCount != nil && *req.ResultCount > 0 {
		limit = *req.ResultCount
	}

	selectColumns := "*"
	whereClause := ""
	var whereArgs []interface{}
	if s.rowPermissionService != nil {
		selectColumns, err = s.rowPermissionService.BuildSelectColumns(datasetGroupID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to build row permission select columns: %w", err)
		}
		whereResult, err := s.rowPermissionService.BuildWhereClause(datasetGroupID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to build row permission where clause: %w", err)
		}
		if whereResult != nil {
			whereClause = whereResult.Clause
			whereArgs = whereResult.Args
		}
	}

	rows, total, err := permissionRepo.QueryRowsWithFilter(req.ID, selectColumns, whereClause, whereArgs, limit)
	if err != nil {
		return nil, err
	}
	rows, err = s.applyColumnRules(datasetGroupID, userID, rows)
	if err != nil {
		return nil, err
	}

	columns := make([]string, 0)
	if len(rows) > 0 {
		for col := range rows[0] {
			columns = append(columns, col)
		}
		sort.Strings(columns)
	}

	return s.buildChartDataResponse(req, rows, columns, total)
}

func (s *ChartService) buildChartDataResponse(req *chart.ChartDataRequest, rows []map[string]interface{}, columns []string, total int64) (*chart.ChartDataResponse, error) {
	chartType, xAxis, yAxis, err := s.resolveChartDataConfig(req)
	if err != nil {
		return nil, err
	}

	resp := &chart.ChartDataResponse{
		ChartID:      req.ID,
		Columns:      columns,
		Rows:         rows,
		Total:        total,
		Fields:       cloneFieldList(xAxis),
		SourceFields: append(cloneFieldList(xAxis), cloneFieldList(yAxis)...),
	}

	switch strings.ToLower(strings.TrimSpace(chartType)) {
	case "table-info", "table-normal":
		resp.TableRow = buildTableRows(rows, resp.SourceFields)
	default:
		if len(xAxis) == 0 && len(yAxis) == 0 {
			resp.TableRow = cloneRows(rows)
			break
		}
		resp.Data = buildSeriesData(rows, xAxis, yAxis)
	}

	return resp, nil
}

func (s *ChartService) resolveChartDataConfig(req *chart.ChartDataRequest) (string, []map[string]interface{}, []map[string]interface{}, error) {
	payload := req.Payload
	chartType := strings.TrimSpace(anyToString(payload["type"]))
	xAxis := fieldListFromAny(payload["xAxis"])
	yAxis := fieldListFromAny(payload["yAxis"])
	if chartType != "" && (len(xAxis) > 0 || len(yAxis) > 0) {
		return chartType, xAxis, yAxis, nil
	}

	view, err := s.repo.GetByID(req.ID)
	if err != nil || view == nil {
		return chartType, xAxis, yAxis, err
	}

	if chartType == "" {
		chartType = stringValue(view.Type)
	}
	if len(xAxis) == 0 {
		xAxis = fieldListFromJSONString(view.XAxis)
	}
	if len(yAxis) == 0 {
		yAxis = fieldListFromJSONString(view.YAxis)
	}

	return chartType, xAxis, yAxis, nil
}

func fieldListFromJSONString(raw *string) []map[string]interface{} {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil
	}
	var parsed interface{}
	if err := json.Unmarshal([]byte(*raw), &parsed); err != nil {
		return nil
	}
	return fieldListFromAny(parsed)
}

func fieldListFromAny(value interface{}) []map[string]interface{} {
	items, ok := value.([]interface{})
	if !ok {
		mapped, ok := value.([]map[string]interface{})
		if !ok {
			return nil
		}
		return cloneFieldList(mapped)
	}

	result := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		mapped, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		result = append(result, cloneMap(mapped))
	}
	return result
}

func buildTableRows(rows []map[string]interface{}, sourceFields []map[string]interface{}) []map[string]interface{} {
	if len(rows) == 0 {
		return []map[string]interface{}{}
	}
	if len(sourceFields) == 0 {
		return cloneRows(rows)
	}

	result := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		item := make(map[string]interface{}, len(sourceFields))
		for _, field := range sourceFields {
			key := preferredFieldKey(field)
			if key == "" {
				continue
			}
			item[key] = fieldRowValue(row, field)
		}
		result = append(result, item)
	}
	return result
}

func buildSeriesData(rows []map[string]interface{}, xAxis []map[string]interface{}, yAxis []map[string]interface{}) []chart.ChartDataPoint {
	if len(rows) == 0 || len(yAxis) == 0 {
		return []chart.ChartDataPoint{}
	}

	groups := make(map[string]*seriesGroup)
	order := make([]string, 0)
	for _, row := range rows {
		key, category, dimensions := buildDimensionKey(row, xAxis)
		group, ok := groups[key]
		if !ok {
			group = &seriesGroup{category: category, dimensions: dimensions, metrics: make(map[string]*metricAccumulator)}
			groups[key] = group
			order = append(order, key)
		}
		for _, quota := range yAxis {
			quotaID := fieldIDString(quota)
			if quotaID == "" {
				quotaID = preferredFieldKey(quota)
			}
			acc, ok := group.metrics[quotaID]
			if !ok {
				acc = &metricAccumulator{summary: strings.ToLower(strings.TrimSpace(anyToString(quota["summary"])))}
				if acc.summary == "" {
					acc.summary = "sum"
				}
				group.metrics[quotaID] = acc
			}
			acc.add(fieldRowValue(row, quota), isCountField(quota))
		}
	}

	result := make([]chart.ChartDataPoint, 0, len(order)*len(yAxis))
	for _, key := range order {
		group := groups[key]
		for _, quota := range yAxis {
			quotaID := fieldIDString(quota)
			if quotaID == "" {
				quotaID = preferredFieldKey(quota)
			}
			acc := group.metrics[quotaID]
			if acc == nil {
				continue
			}
			result = append(result, chart.ChartDataPoint{
				Field:         group.category,
				Name:          group.category,
				Category:      group.category,
				Value:         acc.value(),
				DimensionList: cloneDimensionItems(group.dimensions),
				QuotaList:     []chart.ChartDataFieldItem{{ID: quotaID}},
			})
		}
	}
	return result
}

type seriesGroup struct {
	category   string
	dimensions []chart.ChartDataFieldItem
	metrics    map[string]*metricAccumulator
}

const (
	summaryCount = "count"
	summaryAvg   = "avg"
)

type metricAccumulator struct {
	summary string
	sum     float64
	count   int
	min     float64
	max     float64
	seeded  bool
}

func (m *metricAccumulator) add(value interface{}, countOnly bool) {
	if countOnly || m.summary == summaryCount {
		m.count++
		return
	}

	number, ok := toFloat64(value)
	if !ok {
		return
	}
	if !m.seeded {
		m.min = number
		m.max = number
		m.seeded = true
	}
	if number < m.min {
		m.min = number
	}
	if number > m.max {
		m.max = number
	}
	m.sum += number
	m.count++
}

func (m *metricAccumulator) value() float64 {
	switch m.summary {
	case summaryCount, "count_distinct", "countdistinct":
		return float64(m.count)
	case summaryAvg, "average":
		if m.count == 0 {
			return 0
		}
		return m.sum / float64(m.count)
	case "min":
		if !m.seeded {
			return 0
		}
		return m.min
	case "max":
		if !m.seeded {
			return 0
		}
		return m.max
	default:
		return m.sum
	}
}

func buildDimensionKey(row map[string]interface{}, xAxis []map[string]interface{}) (string, string, []chart.ChartDataFieldItem) {
	if len(xAxis) == 0 {
		return "__all__", "全部", nil
	}

	parts := make([]string, 0, len(xAxis))
	labels := make([]string, 0, len(xAxis))
	dimensions := make([]chart.ChartDataFieldItem, 0, len(xAxis))
	for _, field := range xAxis {
		value := fieldRowValue(row, field)
		text := anyToString(value)
		parts = append(parts, text)
		labels = append(labels, text)
		dimensions = append(dimensions, chart.ChartDataFieldItem{ID: fieldIDString(field), Value: value})
	}
	label := strings.Join(labels, " / ")
	if strings.TrimSpace(label) == "" {
		label = "全部"
	}
	return strings.Join(parts, "\x1f"), label, dimensions
}

func fieldRowValue(row map[string]interface{}, field map[string]interface{}) interface{} {
	if isCountField(field) {
		return 1
	}
	for _, key := range fieldLookupKeys(field) {
		if value, ok := row[key]; ok {
			return value
		}
	}
	return nil
}

func fieldLookupKeys(field map[string]interface{}) []string {
	keys := make([]string, 0, 4)
	for _, candidate := range []string{anyToString(field["dataeaseName"]), anyToString(field["originName"]), anyToString(field["name"]), anyToString(field["fieldShortName"])} {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" || candidate == "*" {
			continue
		}
		alreadyExists := false
		for _, existing := range keys {
			if existing == candidate {
				alreadyExists = true
				break
			}
		}
		if !alreadyExists {
			keys = append(keys, candidate)
		}
	}
	return keys
}

func preferredFieldKey(field map[string]interface{}) string {
	for _, key := range []string{anyToString(field["dataeaseName"]), anyToString(field["originName"]), anyToString(field["name"]), anyToString(field["fieldShortName"])} {
		trimmed := strings.TrimSpace(key)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func fieldIDString(field map[string]interface{}) string {
	return strings.TrimSpace(anyToString(field["id"]))
}

func isCountField(field map[string]interface{}) bool {
	return strings.TrimSpace(anyToString(field["dataeaseName"])) == "*" || strings.EqualFold(strings.TrimSpace(anyToString(field["summary"])), summaryCount) && preferredFieldKey(field) == "*"
}

func cloneRows(rows []map[string]interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		result = append(result, cloneMap(row))
	}
	return result
}

func cloneFieldList(fields []map[string]interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(fields))
	for _, field := range fields {
		result = append(result, cloneMap(field))
	}
	return result
}

func cloneMap(src map[string]interface{}) map[string]interface{} {
	if src == nil {
		return map[string]interface{}{}
	}
	cloned := make(map[string]interface{}, len(src))
	for k, v := range src {
		cloned[k] = v
	}
	return cloned
}

func cloneDimensionItems(items []chart.ChartDataFieldItem) []chart.ChartDataFieldItem {
	result := make([]chart.ChartDataFieldItem, len(items))
	copy(result, items)
	return result
}

func anyToString(v interface{}) string {
	switch value := v.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(value)
	case json.Number:
		return value.String()
	case float64:
		if math.Mod(value, 1) == 0 {
			return fmt.Sprintf("%.0f", value)
		}
		return fmt.Sprintf("%v", value)
	case float32:
		f := float64(value)
		if math.Mod(f, 1) == 0 {
			return fmt.Sprintf("%.0f", f)
		}
		return fmt.Sprintf("%v", value)
	default:
		return fmt.Sprintf("%v", value)
	}
}

func toFloat64(v interface{}) (float64, bool) {
	switch value := v.(type) {
	case nil:
		return 0, false
	case float64:
		return value, true
	case float32:
		return float64(value), true
	case json.Number:
		n, err := value.Float64()
		if err != nil {
			return 0, false
		}
		return n, true
	case string:
		n, err := json.Number(strings.TrimSpace(value)).Float64()
		if err != nil {
			return 0, false
		}
		return n, true
	default:
		return intLikeToFloat(value)
	}
}

func intLikeToFloat(v interface{}) (float64, bool) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(rv.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(rv.Uint()), true
	default:
		return 0, false
	}
}

func (s *ChartService) SaveFromMap(body map[string]interface{}) (*chart.CoreChartView, error) { //nolint:gocyclo // chart view construction with multiple field types
	id, ok := int64FromAny(body["id"])
	if !ok || id <= 0 {
		return nil, fmt.Errorf("chart id is required")
	}

	view, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if title, ok := stringFromAny(body["title"]); ok {
		view.Title = &title
	}
	if tableID, ok := int64FromAny(body["tableId"]); ok {
		view.TableID = &tableID
	}
	if sceneID, ok := int64FromAny(body["sceneId"]); ok {
		view.SceneID = &sceneID
	}
	if chartType, ok := stringFromAny(body["type"]); ok {
		view.Type = &chartType
	}
	if render, ok := stringFromAny(body["render"]); ok {
		view.Render = &render
	}
	if resultMode, ok := stringFromAny(body["resultMode"]); ok {
		view.ResultMode = &resultMode
	}
	if resultCount, ok := intFromAny(body["resultCount"]); ok {
		view.ResultCount = &resultCount
	}
	if dataFrom, ok := stringFromAny(body["dataFrom"]); ok {
		view.DataFrom = &dataFrom
	}

	if v, ok := marshalJSONField(body, "xAxis"); ok {
		view.XAxis = &v
	}
	if v, ok := marshalJSONField(body, "yAxis"); ok {
		view.YAxis = &v
	}
	if v, ok := marshalJSONField(body, "customAttr"); ok {
		view.CustomAttr = &v
	}
	if v, ok := marshalJSONField(body, "customStyle"); ok {
		view.CustomStyle = &v
	}
	if v, ok := marshalJSONField(body, "customFilter"); ok {
		view.CustomFilter = &v
	}

	// Extended axis and chart config fields
	if v, ok := marshalJSONField(body, "xAxisExt"); ok {
		view.XAxisExt = &v
	}
	if v, ok := marshalJSONField(body, "yAxisExt"); ok {
		view.YAxisExt = &v
	}
	if v, ok := marshalJSONField(body, "extStack"); ok {
		view.ExtStack = &v
	}
	if v, ok := marshalJSONField(body, "extBubble"); ok {
		view.ExtBubble = &v
	}
	if v, ok := marshalJSONField(body, "extLabel"); ok {
		view.ExtLabel = &v
	}
	if v, ok := marshalJSONField(body, "extTooltip"); ok {
		view.ExtTooltip = &v
	}
	if v, ok := marshalJSONField(body, "customAttrMobile"); ok {
		view.CustomAttrMobile = &v
	}
	if v, ok := marshalJSONField(body, "customStyleMobile"); ok {
		view.CustomStyleMobile = &v
	}
	if v, ok := marshalJSONField(body, "drillFields"); ok {
		view.DrillFields = &v
	}
	if v, ok := marshalJSONField(body, "senior"); ok {
		view.Senior = &v
	}
	if v, ok := marshalJSONField(body, "snapshot"); ok {
		view.Snapshot = &v
	}
	if v, ok := marshalJSONField(body, "viewFields"); ok {
		view.ViewFields = &v
	}
	if v, ok := marshalJSONField(body, "extColor"); ok {
		view.ExtColor = &v
	}
	if v, ok := marshalJSONField(body, "sortPriority"); ok {
		view.SortPriority = &v
	}

	// Simple string fields
	if v, ok := stringFromAny(body["stylePriority"]); ok {
		view.StylePriority = &v
	}
	if v, ok := stringFromAny(body["chartType"]); ok {
		view.ChartType = &v
	}
	if v, ok := stringFromAny(body["refreshUnit"]); ok {
		view.RefreshUnit = &v
	}
	if v, ok := stringFromAny(body["flowMapStartName"]); ok {
		view.FlowMapStartName = &v
	}
	if v, ok := stringFromAny(body["flowMapEndName"]); ok {
		view.FlowMapEndName = &v
	}

	// Bool fields
	if v, ok := body["isPlugin"]; ok {
		b := boolFromAny(v)
		view.IsPlugin = &b
	}
	if v, ok := body["refreshViewEnable"]; ok {
		b := boolFromAny(v)
		view.RefreshViewEnable = &b
	}
	if v, ok := body["linkageActive"]; ok {
		b := boolFromAny(v)
		view.LinkageActive = &b
	}
	if v, ok := body["jumpActive"]; ok {
		b := boolFromAny(v)
		view.JumpActive = &b
	}
	if v, ok := body["aggregate"]; ok {
		b := boolFromAny(v)
		view.Aggregate = &b
	}

	// Int fields
	if v, ok := intFromAny(body["refreshTime"]); ok {
		view.RefreshTime = &v
	}

	now := time.Now().UnixMilli()
	view.UpdateTime = &now
	if err = s.repo.Update(view); err != nil {
		return nil, err
	}
	return view, nil
}

func (s *ChartService) ListByDQ(datasetGroupID int64, chartID int64) (*chart.ChartFieldListResponse, error) {
	if datasetGroupID <= 0 {
		return &chart.ChartFieldListResponse{DimensionList: []chart.ChartField{}, QuotaList: []chart.ChartField{}}, nil
	}

	baseFields, err := s.repo.ListDatasetFieldsByGroup(datasetGroupID)
	if err != nil {
		return nil, err
	}

	all := make([]chart.ChartField, 0, len(baseFields)+8)
	for _, field := range baseFields {
		if field == nil {
			continue
		}
		all = append(all, convertToChartField(field))
	}
	all = append(all, countChartField(datasetGroupID))

	if chartID > 0 {
		chartFields, chartErr := s.repo.ListDatasetFieldsByChart(chartID)
		if chartErr != nil {
			return nil, chartErr
		}
		for _, field := range chartFields {
			if field == nil {
				continue
			}
			all = append(all, convertToChartField(field))
		}
	}

	dimensionList := make([]chart.ChartField, 0)
	quotaList := make([]chart.ChartField, 0)
	for _, field := range all {
		if strings.EqualFold(field.GroupType, "d") {
			dimensionList = append(dimensionList, field)
		} else {
			quotaList = append(quotaList, field)
		}
	}

	return &chart.ChartFieldListResponse{DimensionList: dimensionList, QuotaList: quotaList}, nil
}

func (s *ChartService) ListByDQWithPermission(datasetGroupID int64, chartID int64, userID int64) (*chart.ChartFieldListResponse, error) {
	resp, err := s.ListByDQ(datasetGroupID, chartID)
	if err != nil {
		return nil, err
	}
	if s.columnPermissionService == nil || userID <= 0 {
		return resp, nil
	}
	if s.rowPermissionService != nil && s.rowPermissionService.IsAdmin(userID) {
		return resp, nil
	}

	disabledColumns, err := s.columnPermissionService.GetDisabledColumns(datasetGroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to load disabled columns: %w", err)
	}
	maskRules, err := s.columnPermissionService.GetMaskRules(datasetGroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to load mask rules: %w", err)
	}
	return &chart.ChartFieldListResponse{
		DimensionList: s.filterChartFields(resp.DimensionList, disabledColumns, maskRules),
		QuotaList:     s.filterChartFields(resp.QuotaList, disabledColumns, maskRules),
	}, nil
}

func (s *ChartService) CopyField(id int64, chartID int64) error {
	if id <= 0 || chartID <= 0 {
		return fmt.Errorf("field id and chart id are required")
	}
	field, err := s.repo.GetDatasetFieldByID(id)
	if err != nil {
		return err
	}

	originName := stringValue(field.Name)
	if originName == "" {
		originName = stringValue(field.OriginName)
	}
	if originName == "" {
		originName = fmt.Sprintf("field_%d", field.ID)
	}

	newName, err := s.nextCopyName(field.DatasetGroupID, originName)
	if err != nil {
		return err
	}

	copied := *field
	copied.ID = 0
	copied.ChartID = &chartID
	copied.Name = &newName
	origin := fmt.Sprintf("[%d]", field.ID)
	copied.OriginName = &origin
	extField := 2
	copied.ExtField = &extField
	copied.DataeaseName = nil
	copied.FieldShortName = nil

	if err = s.repo.CreateDatasetField(&copied); err != nil {
		return err
	}
	aliasSeed := fmt.Sprintf("%d_%s", copied.ID, origin)
	alias := fieldNameShort(aliasSeed)
	if err = s.repo.UpdateDatasetFieldNames(copied.ID, alias, alias); err != nil {
		return err
	}
	return nil
}

func (s *ChartService) DeleteField(id int64) error {
	if id <= 0 {
		return fmt.Errorf("field id is required")
	}
	return s.repo.DeleteDatasetField(id)
}

func (s *ChartService) DeleteFieldByChart(chartID int64) error {
	if chartID <= 0 {
		return fmt.Errorf("chart id is required")
	}
	return s.repo.DeleteDatasetFieldsByChart(chartID)
}

func (s *ChartService) nextCopyName(datasetGroupID int64, source string) (string, error) {
	name := strings.TrimSpace(source)
	if name == "" {
		name = chartDefaultFieldName
	}
	for {
		name = name + "_copy"
		count, err := s.repo.CountDatasetFieldName(datasetGroupID, name)
		if err != nil {
			return "", err
		}
		if count == 0 {
			return name, nil
		}
	}
}

func countChartField(datasetGroupID int64) chart.ChartField {
	return chart.ChartField{
		ID:             -1,
		DatasetGroupID: datasetGroupID,
		OriginName:     "*",
		Name:           "记录数*",
		DataeaseName:   "*",
		FieldShortName: "*",
		GroupType:      "q",
		Type:           "INT",
		DeType:         2,
		DeExtractType:  2,
		ExtField:       1,
		Checked:        true,
		Desensitized:   false,
		Summary:        summaryCount,
	}
}

func convertToChartField(field *dataset.CoreDatasetTableField) chart.ChartField {
	deType := intPointerValue(field.DeType)
	deExtractType := intPointerValue(field.DeExtractType)
	if deExtractType == 0 {
		deExtractType = deType
	}
	groupType := strings.TrimSpace(stringValue(field.GroupType))
	if groupType == "" {
		if deType == 2 || deType == 3 {
			groupType = "q"
		} else {
			groupType = "d"
		}
	}
	summary := "sum"
	if field.ID == -1 || deType == 0 || deType == 1 || deType == 7 {
		summary = summaryCount
	}
	return chart.ChartField{
		ID:             field.ID,
		DatasourceID:   field.DatasourceID,
		DatasetTableID: field.DatasetTableID,
		DatasetGroupID: field.DatasetGroupID,
		ChartID:        field.ChartID,
		OriginName:     stringValue(field.OriginName),
		Name:           stringValue(field.Name),
		DataeaseName:   stringValue(field.DataeaseName),
		FieldShortName: stringValue(field.FieldShortName),
		GroupType:      groupType,
		Type:           stringValue(field.Type),
		DeType:         deType,
		DeExtractType:  deExtractType,
		ExtField:       intPointerValue(field.ExtField),
		Checked:        boolPointerValue(field.Checked),
		Desensitized:   false,
		Summary:        summary,
	}
}

func (s *ChartService) applyColumnRules(datasetGroupID int64, userID int64, rows []map[string]interface{}) ([]map[string]interface{}, error) {
	if len(rows) == 0 || s.columnPermissionService == nil {
		return rows, nil
	}
	if s.rowPermissionService != nil && s.rowPermissionService.IsAdmin(userID) {
		return rows, nil
	}
	disabledColumns, err := s.columnPermissionService.GetDisabledColumns(datasetGroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to load disabled columns: %w", err)
	}
	maskRules, err := s.columnPermissionService.GetMaskRules(datasetGroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to load mask rules: %w", err)
	}
	for i := range rows {
		rows[i] = s.columnPermissionService.FilterDisabledColumns(rows[i], disabledColumns)
		rows[i] = s.columnPermissionService.MaskRowData(rows[i], maskRules)
	}
	return rows, nil
}

func (s *ChartService) filterChartFields(fields []chart.ChartField, disabledColumns map[string]bool, maskRules map[string]*permission.DesensitizationRule) []chart.ChartField {
	result := make([]chart.ChartField, 0, len(fields))
	for _, field := range fields {
		if field.ID == -1 {
			result = append(result, field)
			continue
		}
		fieldKey := chartFieldPermissionKey(field)
		if fieldKey != "" && disabledColumns[fieldKey] {
			continue
		}
		if fieldKey != "" {
			if _, ok := maskRules[fieldKey]; ok {
				field.Desensitized = true
			}
		}
		result = append(result, field)
	}
	return result
}

func chartFieldPermissionKey(field chart.ChartField) string {
	for _, key := range []string{field.OriginName, field.Name, field.DataeaseName, field.FieldShortName} {
		trimmed := strings.TrimSpace(key)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func marshalJSONField(body map[string]interface{}, key string) (string, bool) {
	v, exists := body[key]
	if !exists {
		return "", false
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "", false
	}
	return string(b), true
}

func stringFromAny(v interface{}) (string, bool) {
	s, ok := v.(string)
	if !ok {
		return "", false
	}
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return "", false
	}
	return trimmed, true
}

func int64FromAny(v interface{}) (int64, bool) {
	switch n := v.(type) {
	case int64:
		return n, true
	case int:
		return int64(n), true
	case float64:
		return int64(n), true
	case json.Number:
		parsed, err := n.Int64()
		if err != nil {
			return 0, false
		}
		return parsed, true
	case string:
		parsed, err := json.Number(strings.TrimSpace(n)).Int64()
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}

func intFromAny(v interface{}) (int, bool) {
	parsed, ok := int64FromAny(v)
	if !ok {
		return 0, false
	}
	return int(parsed), true
}

func stringValue(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}

func intPointerValue(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

func boolPointerValue(v *bool) bool {
	if v == nil {
		return false
	}
	return *v
}

func fieldNameShort(seed string) string {
	sum := md5.Sum([]byte(seed))
	hex := fmt.Sprintf("%x", sum)
	if len(hex) < 24 {
		return "f_" + hex
	}
	return "f_" + hex[8:24]
}

func boolFromAny(v interface{}) bool {
	switch b := v.(type) {
	case bool:
		return b
	case string:
		return strings.EqualFold(b, "true") || b == "1"
	case float64:
		return b != 0
	case int:
		return b != 0
	case json.Number:
		s := strings.ToLower(string(b))
		return s == "true" || s == "1"
	default:
		return false
	}
}
