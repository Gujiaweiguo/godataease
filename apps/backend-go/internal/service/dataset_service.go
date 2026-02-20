package service

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"

	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/repository"

	"gorm.io/gorm"
)

type DatasetService struct {
	repo *repository.DatasetRepository
}

type sqlVariableDetailRaw struct {
	VariableName string        `json:"variableName"`
	Type         []string      `json:"type"`
	Params       []interface{} `json:"params"`
}

func NewDatasetService(repo *repository.DatasetRepository) *DatasetService {
	return &DatasetService{repo: repo}
}

func (s *DatasetService) Tree(req *dataset.TreeRequest) ([]dataset.TreeNode, error) {
	groups, err := s.repo.ListGroups(req.Keyword)
	if err != nil {
		return nil, err
	}

	nodesByID := make(map[int64]*dataset.TreeNode)
	childrenByPID := make(map[int64][]*dataset.TreeNode)

	for _, g := range groups {
		nodeType := "dataset"
		if g.NodeType != nil && *g.NodeType != "" {
			nodeType = *g.NodeType
		}
		n := &dataset.TreeNode{
			ID:       g.ID,
			Name:     g.Name,
			NodeType: nodeType,
		}
		nodesByID[g.ID] = n
		pid := int64(0)
		if g.PID != nil {
			pid = *g.PID
		}
		childrenByPID[pid] = append(childrenByPID[pid], n)
	}

	for id, node := range nodesByID {
		children := childrenByPID[id]
		if len(children) == 0 {
			continue
		}
		sort.Slice(children, func(i, j int) bool {
			return children[i].ID < children[j].ID
		})
		node.Children = make([]dataset.TreeNode, 0, len(children))
		for _, c := range children {
			node.Children = append(node.Children, *c)
		}
	}

	rootChildren := childrenByPID[0]
	sort.Slice(rootChildren, func(i, j int) bool {
		return rootChildren[i].ID < rootChildren[j].ID
	})
	roots := make([]dataset.TreeNode, 0, len(rootChildren))
	for _, r := range rootChildren {
		roots = append(roots, *r)
	}
	return roots, nil
}

func (s *DatasetService) Fields(req *dataset.FieldsRequest) ([]*dataset.CoreDatasetTableField, error) {
	return s.repo.ListFields(req.DatasetGroupID)
}

func (s *DatasetService) Preview(req *dataset.PreviewRequest) (*dataset.PreviewResponse, error) {
	limit := req.Limit
	if limit < 1 {
		limit = 100
	}

	tableName, err := s.repo.FindPrimaryTableName(req.DatasetGroupID)
	if err != nil {
		return nil, err
	}

	rows, err := s.repo.PreviewRows(tableName, limit)
	if err != nil {
		return nil, err
	}
	total, err := s.repo.CountRows(tableName)
	if err != nil {
		return nil, err
	}

	columns := make([]string, 0)
	if len(rows) > 0 {
		for k := range rows[0] {
			columns = append(columns, k)
		}
		sort.Strings(columns)
	}

	return &dataset.PreviewResponse{
		Columns: columns,
		Rows:    rows,
		Total:   total,
	}, nil
}

func (s *DatasetService) PreviewSQL(req *dataset.SQLPreviewRequest) (map[string]interface{}, error) {
	empty := map[string]interface{}{
		"data": dataset.SQLPreviewData{
			Fields: []dataset.SQLPreviewField{},
			Data:   []map[string]interface{}{},
		},
		"sql": "",
	}

	if req == nil {
		return empty, nil
	}

	rawSQL := strings.TrimSpace(req.SQL)
	if rawSQL == "" {
		return empty, nil
	}
	decoded, decodeErr := base64.StdEncoding.DecodeString(rawSQL)
	if decodeErr == nil {
		rawSQL = strings.TrimSpace(string(decoded))
	}
	if rawSQL == "" {
		return empty, nil
	}

	if err := validatePreviewSQL(rawSQL); err != nil {
		return nil, err
	}

	rows, err := s.repo.PreviewSQL(rawSQL, 100)
	if err != nil {
		return nil, err
	}
	for i := range rows {
		rows[i] = normalizePreviewRow(rows[i])
	}

	fields := buildPreviewFields(rows)
	return map[string]interface{}{
		"data": dataset.SQLPreviewData{
			Fields: fields,
			Data:   rows,
		},
		"sql": base64.StdEncoding.EncodeToString([]byte(rawSQL)),
	}, nil
}

func (s *DatasetService) GetSQLParams(ids []int64) ([]dataset.SQLVariableDetails, error) {
	if len(ids) == 0 {
		return []dataset.SQLVariableDetails{}, nil
	}

	result := make([]dataset.SQLVariableDetails, 0)
	for _, datasetGroupID := range ids {
		if datasetGroupID <= 0 {
			continue
		}

		tables, err := s.repo.ListTablesByDatasetGroupID(datasetGroupID)
		if err != nil {
			return nil, err
		}
		if len(tables) == 0 {
			continue
		}

		fullName, err := s.datasetFullName(datasetGroupID)
		if err != nil {
			return nil, err
		}

		for _, table := range tables {
			if table == nil || table.SQLVariables == nil || strings.TrimSpace(*table.SQLVariables) == "" {
				continue
			}

			rawList := make([]sqlVariableDetailRaw, 0)
			if err = json.Unmarshal([]byte(*table.SQLVariables), &rawList); err != nil {
				continue
			}

			for _, raw := range rawList {
				name := strings.TrimSpace(raw.VariableName)
				if name == "" {
					continue
				}

				item := dataset.SQLVariableDetails{
					ID:              fmt.Sprintf("%d|DE|%s", table.ID, name),
					VariableName:    name,
					Type:            raw.Type,
					Params:          raw.Params,
					DatasetGroupID:  datasetGroupID,
					DatasetTableID:  table.ID,
					DatasetFullName: fullName,
					DeType:          inferSQLVariableDeType(raw.Type),
				}
				result = append(result, item)
			}
		}
	}

	return result, nil
}

func (s *DatasetService) GetFieldEnum(req *dataset.MultFieldValuesRequest) ([]string, error) {
	if req == nil || len(req.FieldIDs) == 0 {
		return []string{}, nil
	}

	limit := 1000
	if req.ResultMode == 1 {
		limit = 5000
	}

	uniqField := make(map[int64]struct{}, len(req.FieldIDs))
	result := make([]string, 0)
	seen := make(map[string]struct{})
	for _, fieldID := range req.FieldIDs {
		if fieldID <= 0 {
			continue
		}
		if _, ok := uniqField[fieldID]; ok {
			continue
		}
		uniqField[fieldID] = struct{}{}

		field, tableName, columnName, err := s.resolveEnumFieldTarget(fieldID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return nil, err
		}

		filters, err := s.buildEnumFilterClauses(req.Filter, tableName)
		if err != nil {
			return nil, err
		}
		values, err := s.repo.QueryDistinctValues(tableName, columnName, filters, limit)
		if err != nil {
			return nil, err
		}

		for _, value := range values {
			normalized := normalizeEnumValue(value, field.DeType)
			if normalized == "" {
				continue
			}
			if _, ok := seen[normalized]; ok {
				continue
			}
			seen[normalized] = struct{}{}
			result = append(result, normalized)
		}
	}
	return result, nil
}

func (s *DatasetService) GetFieldEnumObj(req *dataset.EnumValueRequest) ([]map[string]interface{}, error) {
	if req == nil || req.QueryID <= 0 {
		return []map[string]interface{}{}, nil
	}

	queryField, tableName, queryColumn, err := s.resolveEnumFieldTarget(req.QueryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []map[string]interface{}{}, nil
		}
		return nil, err
	}

	displayID := req.DisplayID
	if displayID <= 0 {
		displayID = req.QueryID
	}
	displayField, _, displayColumn, err := s.resolveEnumFieldTarget(displayID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			displayID = req.QueryID
			displayField = queryField
			displayColumn = queryColumn
		} else {
			return nil, err
		}
	}

	if req.SortID > 0 {
		_, sortTableName, _, sortErr := s.resolveEnumFieldTarget(req.SortID)
		if sortErr == nil && sortTableName != tableName {
			req.SortID = 0
		}
	}

	if displayID != req.QueryID {
		_, displayTableName, _, displayErr := s.resolveEnumFieldTarget(displayID)
		if displayErr == nil && displayTableName != tableName {
			displayID = req.QueryID
			displayField = queryField
			displayColumn = queryColumn
		}
	}

	columns := []dataset.EnumObjectColumn{{Column: queryColumn, Alias: enumAlias(req.QueryID)}}
	if displayID != req.QueryID {
		columns = append(columns, dataset.EnumObjectColumn{Column: displayColumn, Alias: enumAlias(displayID)})
	}

	filters, err := s.buildEnumFilterClauses(req.Filter, tableName)
	if err != nil {
		return nil, err
	}

	limit := 1000
	if req.ResultMode == 1 {
		limit = 5000
	}

	searchColumn := displayColumn
	if displayID == req.QueryID {
		searchColumn = queryColumn
	}

	sortColumn := ""
	if req.SortID > 0 {
		_, _, resolvedSortColumn, sortErr := s.resolveEnumFieldTarget(req.SortID)
		if sortErr == nil {
			sortColumn = resolvedSortColumn
		}
	}
	if sortColumn == "" {
		sortColumn = searchColumn
	}

	rows, err := s.repo.QueryDistinctObjectValues(
		tableName,
		columns,
		filters,
		searchColumn,
		req.SearchText,
		sortColumn,
		req.Sort,
		limit,
	)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(rows))
	seen := make(map[string]struct{})
	for _, row := range rows {
		item := make(map[string]interface{}, len(columns))
		hasEmpty := false
		for _, column := range columns {
			rawValue, exists := row[column.Alias]
			if !exists {
				hasEmpty = true
				break
			}
			fieldID := enumFieldIDFromAlias(column.Alias)
			deType := queryField.DeType
			if fieldID == displayID {
				deType = displayField.DeType
			}
			normalized := normalizeEnumValue(fmt.Sprintf("%v", normalizePreviewValue(rawValue)), deType)
			if normalized == "" {
				hasEmpty = true
				break
			}
			item[strconv.FormatInt(fieldID, 10)] = normalized
		}
		if hasEmpty {
			continue
		}
		keyBytes, marshalErr := json.Marshal(item)
		if marshalErr != nil {
			continue
		}
		key := string(keyBytes)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, item)
	}

	return result, nil
}

func (s *DatasetService) GetFieldEnumDs(fieldID int64) ([]string, error) {
	if fieldID <= 0 {
		return []string{}, nil
	}
	return s.GetFieldEnum(&dataset.MultFieldValuesRequest{FieldIDs: []int64{fieldID}, ResultMode: 0})
}

func (s *DatasetService) PerDelete(id int64) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("dataset id is required")
	}
	count, err := s.repo.CountChartRelations(id)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *DatasetService) resolveEnumFieldTarget(fieldID int64) (*dataset.CoreDatasetTableField, string, string, error) {
	field, err := s.repo.GetFieldByID(fieldID)
	if err != nil {
		return nil, "", "", err
	}

	columnName := ""
	if field.OriginName != nil {
		columnName = strings.TrimSpace(*field.OriginName)
	}
	if columnName == "" && field.DataeaseName != nil {
		columnName = strings.TrimSpace(*field.DataeaseName)
	}
	if columnName == "" && field.Name != nil {
		columnName = strings.TrimSpace(*field.Name)
	}
	if columnName == "" {
		return nil, "", "", fmt.Errorf("dataset field origin name is required")
	}

	tableName := ""
	if field.DatasetTableID != nil && *field.DatasetTableID > 0 {
		table, tableErr := s.repo.GetTableByID(*field.DatasetTableID)
		if tableErr == nil && table.PhysicalTable != nil {
			tableName = strings.TrimSpace(*table.PhysicalTable)
		}
	}
	if tableName == "" {
		tableName, err = s.repo.FindPrimaryTableName(field.DatasetGroupID)
		if err != nil {
			return nil, "", "", err
		}
	}

	return field, tableName, columnName, nil
}

func (s *DatasetService) buildEnumFilterClauses(filters []dataset.EnumFilter, targetTableName string) ([]dataset.EnumFilterClause, error) {
	clauses := make([]dataset.EnumFilterClause, 0)
	for _, filter := range filters {
		if strings.TrimSpace(filter.Operator) != "" && !strings.EqualFold(strings.TrimSpace(filter.Operator), "in") {
			continue
		}
		ids := parseFilterFieldIDs(filter.FieldID)
		if len(ids) == 0 {
			continue
		}
		values := extractFilterValues(filter.Value)
		if len(values) == 0 {
			continue
		}

		for _, id := range ids {
			_, tableName, columnName, err := s.resolveEnumFieldTarget(id)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return nil, err
			}
			if tableName != targetTableName {
				continue
			}
			clauses = append(clauses, dataset.EnumFilterClause{Column: columnName, Values: values})
		}
	}
	return clauses, nil
}

func parseFilterFieldIDs(fieldID string) []int64 {
	text := strings.TrimSpace(fieldID)
	if text == "" {
		return []int64{}
	}
	parts := strings.Split(text, ",")
	result := make([]int64, 0, len(parts))
	seen := make(map[int64]struct{}, len(parts))
	for _, part := range parts {
		id, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err != nil || id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func extractFilterValues(values []interface{}) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		text := strings.TrimSpace(fmt.Sprintf("%v", value))
		if text == "" || text == "<nil>" {
			continue
		}
		if _, ok := seen[text]; ok {
			continue
		}
		seen[text] = struct{}{}
		result = append(result, text)
	}
	return result
}

func normalizeEnumValue(value string, deType *int) string {
	text := strings.TrimSpace(value)
	if text == "" || strings.EqualFold(text, "<nil>") {
		return ""
	}
	if deType != nil && *deType == 3 && strings.Contains(strings.ToUpper(text), "E") {
		if f, _, err := big.ParseFloat(text, 10, 128, big.ToNearestEven); err == nil {
			return strings.TrimRight(strings.TrimRight(f.Text('f', 8), "0"), ".")
		}
	}
	return text
}

func enumAlias(fieldID int64) string {
	return fmt.Sprintf("f_%d", fieldID)
}

func enumFieldIDFromAlias(alias string) int64 {
	trimmed := strings.TrimSpace(alias)
	if !strings.HasPrefix(trimmed, "f_") {
		return 0
	}
	id, err := strconv.ParseInt(strings.TrimPrefix(trimmed, "f_"), 10, 64)
	if err != nil {
		return 0
	}
	return id
}

func (s *DatasetService) Save(req *dataset.WriteRequest) (*dataset.CoreDatasetGroup, error) {
	if req == nil {
		return nil, fmt.Errorf("dataset request is required")
	}
	if req.ID <= 0 {
		return s.Create(req)
	}

	existing, err := s.repo.GetGroupByID(req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("dataset not found")
		}
		return nil, err
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = existing.Name
	}
	pid := normalizedDatasetPID(req.PID)
	if req.PID == nil && existing.PID != nil {
		pid = *existing.PID
	}

	count, err := s.repo.CountGroupByNameAndPID(name, pid, &req.ID)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("dataset name already exists")
	}

	existing.Name = name
	existing.PID = &pid
	nodeType := normalizedDatasetNodeType(req.NodeType)
	if nodeType == "" {
		if existing.NodeType != nil {
			nodeType = strings.TrimSpace(*existing.NodeType)
		}
	}
	if nodeType != "" {
		existing.NodeType = &nodeType
	}
	if req.Type != nil {
		existing.Type = req.Type
	}

	if err = s.repo.UpdateGroup(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *DatasetService) Create(req *dataset.WriteRequest) (*dataset.CoreDatasetGroup, error) {
	if req == nil {
		return nil, fmt.Errorf("dataset request is required")
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("dataset name is required")
	}

	pid := normalizedDatasetPID(req.PID)
	level := 0
	if pid > 0 {
		parent, err := s.repo.GetGroupByID(pid)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("destination folder not found")
			}
			return nil, err
		}
		if parent.Level != nil {
			level = *parent.Level + 1
		} else {
			level = 1
		}
	}

	count, err := s.repo.CountGroupByNameAndPID(name, pid, nil)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("dataset name already exists")
	}

	nodeType := normalizedDatasetNodeType(req.NodeType)
	if nodeType == "" {
		nodeType = dataset.NodeTypeFolder
	}
	delFlag := 0
	group := &dataset.CoreDatasetGroup{
		Name:     name,
		PID:      &pid,
		Level:    &level,
		NodeType: &nodeType,
		Type:     req.Type,
		DelFlag:  &delFlag,
	}

	if err = s.repo.CreateGroup(group); err != nil {
		return nil, err
	}

	return group, nil
}

func (s *DatasetService) Rename(id int64, name string) (*dataset.CoreDatasetGroup, error) {
	if id <= 0 {
		return nil, fmt.Errorf("dataset id is required")
	}

	existing, err := s.repo.GetGroupByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("dataset not found")
		}
		return nil, err
	}

	newName := strings.TrimSpace(name)
	if newName == "" {
		return nil, fmt.Errorf("dataset name is required")
	}

	pid := int64(0)
	if existing.PID != nil {
		pid = *existing.PID
	}
	count, err := s.repo.CountGroupByNameAndPID(newName, pid, &id)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("dataset name already exists")
	}

	existing.Name = newName
	if err = s.repo.UpdateGroup(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *DatasetService) Move(id int64, pid int64) (*dataset.CoreDatasetGroup, error) {
	if id <= 0 {
		return nil, fmt.Errorf("dataset id is required")
	}
	if id == pid {
		return nil, fmt.Errorf("destination folder cannot be itself")
	}

	existing, err := s.repo.GetGroupByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("dataset not found")
		}
		return nil, err
	}

	if pid > 0 {
		if _, err = s.repo.GetGroupByID(pid); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("destination folder not found")
			}
			return nil, err
		}
		isDescendant, checkErr := s.isDescendant(id, pid)
		if checkErr != nil {
			return nil, checkErr
		}
		if isDescendant {
			return nil, fmt.Errorf("destination folder cannot be child of current dataset")
		}
	}

	count, err := s.repo.CountGroupByNameAndPID(existing.Name, pid, &id)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("dataset name already exists")
	}

	existing.PID = &pid
	if err = s.repo.UpdateGroup(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *DatasetService) Delete(id int64) error {
	if id <= 0 {
		return fmt.Errorf("dataset id is required")
	}
	return s.deleteRecursive(id)
}

func (s *DatasetService) deleteRecursive(id int64) error {
	children, err := s.repo.ListGroupChildren(id)
	if err != nil {
		return err
	}
	for _, child := range children {
		if err = s.deleteRecursive(child.ID); err != nil {
			return err
		}
	}
	return s.repo.SoftDeleteGroup(id)
}

func (s *DatasetService) isDescendant(rootID int64, targetID int64) (bool, error) {
	children, err := s.repo.ListGroupChildren(rootID)
	if err != nil {
		return false, err
	}
	for _, child := range children {
		if child.ID == targetID {
			return true, nil
		}
		descendant, innerErr := s.isDescendant(child.ID, targetID)
		if innerErr != nil {
			return false, innerErr
		}
		if descendant {
			return true, nil
		}
	}
	return false, nil
}

func normalizedDatasetPID(pid *int64) int64 {
	if pid == nil {
		return 0
	}
	if *pid < 0 {
		return 0
	}
	return *pid
}

func normalizedDatasetNodeType(nodeType string) string {
	nt := strings.TrimSpace(nodeType)
	if nt == "" {
		return ""
	}
	switch nt {
	case dataset.NodeTypeFolder, dataset.NodeTypeDataset:
		return nt
	default:
		return dataset.NodeTypeDataset
	}
}

func (s *DatasetService) datasetFullName(datasetGroupID int64) (string, error) {
	if datasetGroupID <= 0 {
		return "", nil
	}

	names := make([]string, 0)
	visited := make(map[int64]struct{})
	currentID := datasetGroupID
	for currentID > 0 {
		if _, ok := visited[currentID]; ok {
			break
		}
		visited[currentID] = struct{}{}

		group, err := s.repo.GetGroupByID(currentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break
			}
			return "", err
		}
		n := strings.TrimSpace(group.Name)
		if n != "" {
			names = append(names, n)
		}
		if group.PID == nil || *group.PID <= 0 || *group.PID == currentID {
			break
		}
		currentID = *group.PID
	}

	for i, j := 0, len(names)-1; i < j; i, j = i+1, j-1 {
		names[i], names[j] = names[j], names[i]
	}
	return strings.Join(names, "/"), nil
}

func validatePreviewSQL(rawSQL string) error {
	text := strings.TrimSpace(strings.TrimSuffix(rawSQL, ";"))
	lower := strings.ToLower(text)
	if text == "" {
		return fmt.Errorf("sql is required")
	}
	if !(strings.HasPrefix(lower, "select") || strings.HasPrefix(lower, "with")) {
		return fmt.Errorf("only select query is supported")
	}
	if strings.Contains(text, ";") {
		return fmt.Errorf("only single select statement is supported")
	}
	blocked := []string{" insert ", " update ", " delete ", " drop ", " alter ", " truncate ", " create "}
	padded := " " + lower + " "
	for _, token := range blocked {
		if strings.Contains(padded, token) {
			return fmt.Errorf("unsupported sql statement")
		}
	}
	return nil
}

func buildPreviewFields(rows []map[string]interface{}) []dataset.SQLPreviewField {
	if len(rows) == 0 {
		return []dataset.SQLPreviewField{}
	}

	keys := make([]string, 0, len(rows[0]))
	for k := range rows[0] {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fields := make([]dataset.SQLPreviewField, 0, len(keys))
	for _, k := range keys {
		fields = append(fields, dataset.SQLPreviewField{
			OriginName: k,
			DeType:     inferPreviewDeType(rows[0][k]),
		})
	}
	return fields
}

func normalizePreviewRow(row map[string]interface{}) map[string]interface{} {
	if row == nil {
		return map[string]interface{}{}
	}
	result := make(map[string]interface{}, len(row))
	for key, val := range row {
		result[key] = normalizePreviewValue(val)
	}
	return result
}

func normalizePreviewValue(v interface{}) interface{} {
	switch value := v.(type) {
	case []byte:
		return string(value)
	case time.Time:
		return value.Format("2006-01-02 15:04:05")
	default:
		return value
	}
}

func inferPreviewDeType(v interface{}) int {
	switch value := v.(type) {
	case bool:
		return 4
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return 2
	case float32, float64:
		return 3
	case time.Time:
		return 1
	case string:
		text := strings.TrimSpace(value)
		if text == "" {
			return 0
		}
		if isDateTimeText(text) {
			return 1
		}
		if _, err := strconv.ParseInt(text, 10, 64); err == nil {
			return 2
		}
		if _, err := strconv.ParseFloat(text, 64); err == nil {
			return 3
		}
		return 0
	default:
		return 0
	}
}

func inferSQLVariableDeType(typeList []string) int {
	if len(typeList) == 0 {
		return 0
	}
	typeText := strings.ToUpper(strings.TrimSpace(typeList[0]))
	if typeText == "" {
		return 0
	}
	if strings.Contains(typeText, "DATETIME") || strings.Contains(typeText, "TIMESTAMP") || strings.Contains(typeText, "DATE") || strings.Contains(typeText, "TIME") || strings.Contains(typeText, "YEAR") {
		return 1
	}
	if strings.Contains(typeText, "DOUBLE") || strings.Contains(typeText, "FLOAT") || strings.Contains(typeText, "DECIMAL") || strings.Contains(typeText, "NUMERIC") || strings.Contains(typeText, "REAL") {
		return 3
	}
	if strings.Contains(typeText, "INT") || strings.Contains(typeText, "LONG") || strings.Contains(typeText, "SHORT") || strings.Contains(typeText, "BIGINT") || strings.Contains(typeText, "SMALLINT") || strings.Contains(typeText, "TINYINT") {
		return 2
	}
	if strings.Contains(typeText, "BOOL") {
		return 4
	}
	return 0
}

func isDateTimeText(text string) bool {
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
		"2006/01/02 15:04:05",
		"2006/01/02",
	}
	for _, layout := range layouts {
		if _, err := time.Parse(layout, text); err == nil {
			return true
		}
	}
	return false
}
