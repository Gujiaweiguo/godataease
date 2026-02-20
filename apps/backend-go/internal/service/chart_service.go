package service

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/dataset"
)

type ChartRepository interface {
	GetByID(id int64) (*chart.CoreChartView, error)
	Update(view *chart.CoreChartView) error
	QueryRows(chartID int64, limit int) ([]map[string]interface{}, int64, error)
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
	repo ChartRepository
}

func NewChartService(repo ChartRepository) *ChartService {
	return &ChartService{repo: repo}
}

func (s *ChartService) Query(req *chart.ChartQueryRequest) (*chart.CoreChartView, error) {
	return s.repo.GetByID(req.ID)
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

	return &chart.ChartDataResponse{
		ChartID: req.ID,
		Columns: columns,
		Rows:    rows,
		Total:   total,
	}, nil
}

func (s *ChartService) SaveFromMap(body map[string]interface{}) (*chart.CoreChartView, error) {
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
		name = "field"
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
		Summary:        "count",
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
		summary = "count"
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
