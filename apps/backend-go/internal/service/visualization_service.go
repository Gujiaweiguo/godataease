package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"dataease/backend/internal/domain/audit"
	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/domain/visualization"
	"dataease/backend/internal/repository"

	"github.com/google/uuid"
)

type VisualizationService struct {
	repo                   *repository.VisualizationRepository
	datasetRepo            *repository.DatasetRepository
	resourcePermService    *ResourcePermissionService
	staticService          *StaticService
	templateService        *TemplateService
	templateExtendDataRepo *repository.TemplateExtendDataRepository
	auditService           *AuditService
	thresholdService       *ThresholdService
}

type marketTemplateDTO struct {
	Name            string          `json:"name"`
	DvType          string          `json:"dvType"`
	Version         int             `json:"version"`
	CanvasStyleData json.RawMessage `json:"canvasStyleData"`
	ComponentData   json.RawMessage `json:"componentData"`
	DynamicData     json.RawMessage `json:"dynamicData"`
	AppData         json.RawMessage `json:"appData"`
	StaticResource  json.RawMessage `json:"staticResource"`
}

const (
	visualizationNodeTypePanel      = "panel"
	visualizationNodeTypeFolder     = "folder"
	visualizationTypeDashboard      = "dashboard"
	visualizationTypeDataV          = "dataV"
	visualizationBusiFlagDashboardV = "dashboard-dataV"
	jsonNullLiteral                 = "null"
)

func NewVisualizationService(repo *repository.VisualizationRepository) *VisualizationService {
	return &VisualizationService{repo: repo}
}

func (s *VisualizationService) SetResourcePermissionService(resourcePermSvc *ResourcePermissionService) {
	s.resourcePermService = resourcePermSvc
}

func (s *VisualizationService) SetTemplateService(ts *TemplateService) {
	s.templateService = ts
}

func (s *VisualizationService) SetStaticService(ss *StaticService) {
	s.staticService = ss
}

func (s *VisualizationService) SetTemplateExtendDataRepo(r *repository.TemplateExtendDataRepository) {
	s.templateExtendDataRepo = r
}

func (s *VisualizationService) SetDatasetRepository(r *repository.DatasetRepository) {
	s.datasetRepo = r
}

func (s *VisualizationService) SetAuditService(auditSvc *AuditService) {
	s.auditService = auditSvc
}

func (s *VisualizationService) SetThresholdService(ts *ThresholdService) {
	s.thresholdService = ts
}

func (s *VisualizationService) Save(req *visualization.SaveRequest, updateBy string) (int64, error) {
	if err := s.processAppImport(req, updateBy); err != nil {
		return 0, fmt.Errorf("app import: %w", err)
	}

	now := time.Now().UnixMilli()
	nodeType := visualizationNodeTypePanel
	if req.NodeType != nil && *req.NodeType != "" {
		nodeType = *req.NodeType
	}
	status := 0
	if nodeType == visualizationNodeTypeFolder {
		status = 1
	}

	v := &visualization.DataVisualizationInfo{
		Name:            req.Name,
		PID:             req.PID,
		Type:            req.Type,
		NodeType:        &nodeType,
		CanvasStyleData: req.CanvasStyleData,
		ComponentData:   req.ComponentData,
		MobileLayout:    req.MobileLayout,
		ContentID:       req.ContentID,
		CheckVersion:    req.CheckVersion,
		Status:          &status,
		CreateTime:      &now,
		UpdateTime:      &now,
		CreateBy:        &updateBy,
		UpdateBy:        &updateBy,
	}

	if err := s.repo.Create(v); err != nil {
		return 0, err
	}
	s.saveChartViewsFromVisualization(req.ComponentData, v.ID, req.CanvasViewInfo)
	if err := s.applyInheritedPermissionsOnCreate(v.ID, req.Name, req.PID, req.Type); err != nil {
		_ = s.repo.DeleteLogic(v.ID, updateBy)
		return 0, err
	}
	return v.ID, nil
}

func (s *VisualizationService) processAppImport(req *visualization.SaveRequest, updateBy string) error {
	if req == nil || req.AppData == "" {
		return nil
	}
	if req.DataType != nil && *req.DataType == "dataset" {
		return nil
	}
	if s.datasetRepo == nil {
		return nil
	}

	appData, err := s.parseAppDataJSON(req.AppData)
	if err != nil {
		return nil
	}

	datasourceIDMap := s.buildDatasourceIDMap(appData)
	now := time.Now().UnixMilli()

	folderID, err := s.createDatasetFolder(req, updateBy, now)
	if err != nil {
		return err
	}

	datasetGroupsInfo, err := parseJSONFieldFromAppData[[]map[string]interface{}](appData, "datasetGroupsInfo")
	if err != nil {
		return fmt.Errorf("parse dataset groups: %w", err)
	}

	datasetGroupIDMap, newDatasetGroups := s.buildAppDatasetGroups(datasetGroupsInfo, folderID, updateBy, now)

	datasetTablesInfo, err := parseJSONFieldFromAppData[[]map[string]interface{}](appData, "datasetTablesInfo")
	if err != nil {
		return fmt.Errorf("parse dataset tables: %w", err)
	}

	datasetTableIDMap, err := s.createAppDatasetTables(datasetTablesInfo, datasetGroupIDMap, datasourceIDMap, folderID)
	if err != nil {
		return err
	}

	datasetTableFieldsInfo, err := parseJSONFieldFromAppData[[]map[string]interface{}](appData, "datasetTableFieldsInfo")
	if err != nil {
		return fmt.Errorf("parse dataset table fields: %w", err)
	}

	datasetTableFieldIDMap, err := s.createAppDatasetFields(datasetTableFieldsInfo, datasetGroupIDMap, datasetTableIDMap, datasourceIDMap, folderID)
	if err != nil {
		return err
	}

	if err := s.persistAppDatasetGroups(newDatasetGroups, datasetGroupIDMap, datasetTableIDMap, datasetTableFieldIDMap, datasourceIDMap); err != nil {
		return err
	}

	s.remapAppImportIDs(req, datasetGroupIDMap, datasetTableIDMap, datasetTableFieldIDMap, datasourceIDMap)

	return nil
}

func (s *VisualizationService) parseAppDataJSON(appDataStr string) (map[string]json.RawMessage, error) {
	var appData map[string]json.RawMessage
	if err := json.Unmarshal([]byte(appDataStr), &appData); err != nil {
		return nil, err
	}
	return appData, nil
}

func (s *VisualizationService) buildDatasourceIDMap(appData map[string]json.RawMessage) map[int64]int64 {
	datasourceIDMap := make(map[int64]int64)
	raw, ok := appData["datasourceInfo"]
	if !ok {
		return datasourceIDMap
	}
	var datasourceInfo []map[string]interface{}
	if err := json.Unmarshal(raw, &datasourceInfo); err != nil {
		return datasourceIDMap
	}
	for _, ds := range datasourceInfo {
		oldID, okOld := extractInt64Value(ds["id"])
		systemID, okSystem := extractInt64Value(ds["systemDatasourceId"])
		if okOld && okSystem && oldID > 0 && systemID > 0 {
			datasourceIDMap[oldID] = systemID
		}
	}
	return datasourceIDMap
}

func (s *VisualizationService) createDatasetFolder(req *visualization.SaveRequest, updateBy string, now int64) (int64, error) {
	if req.DatasetFolderName == nil || strings.TrimSpace(*req.DatasetFolderName) == "" {
		return 0, nil
	}
	folderID := int64(uuid.New().ID())
	folderNodeType := dataset.NodeTypeFolder
	folderLevel := 0
	folder := &dataset.CoreDatasetGroup{
		ID:             folderID,
		Name:           strings.TrimSpace(*req.DatasetFolderName),
		PID:            req.DatasetFolderPID,
		Level:          &folderLevel,
		NodeType:       &folderNodeType,
		CreateBy:       updateBy,
		CreateTime:     now,
		UpdateBy:       updateBy,
		LastUpdateTime: now,
	}
	if err := s.datasetRepo.CreateGroup(folder); err != nil {
		return 0, fmt.Errorf("create dataset folder: %w", err)
	}
	return folderID, nil
}

func (s *VisualizationService) buildAppDatasetGroups(
	groupsInfo []map[string]interface{},
	folderID int64,
	updateBy string,
	now int64,
) (map[int64]int64, []dataset.CoreDatasetGroup) {
	datasetGroupIDMap := make(map[int64]int64)
	newGroups := make([]dataset.CoreDatasetGroup, 0, len(groupsInfo))
	for _, groupInfo := range groupsInfo {
		nodeType, _ := groupInfo["nodeType"].(string)
		if nodeType != dataset.NodeTypeDataset {
			continue
		}
		oldID, ok := extractInt64Value(groupInfo["id"])
		if !ok || oldID <= 0 {
			continue
		}
		newID := int64(uuid.New().ID())
		datasetGroupIDMap[oldID] = newID
		name, _ := groupInfo["name"].(string)
		groupType, _ := groupInfo["type"].(string)
		info, _ := groupInfo["info"].(string)
		groupNodeType := dataset.NodeTypeDataset
		groupLevel := 0
		parentID := folderID
		newGroups = append(newGroups, dataset.CoreDatasetGroup{
			ID:             newID,
			Name:           name,
			PID:            &parentID,
			Level:          &groupLevel,
			NodeType:       &groupNodeType,
			Type:           stringPtrIfNotEmpty(groupType),
			Info:           stringPtrAllowEmpty(info),
			CreateBy:       updateBy,
			CreateTime:     now,
			UpdateBy:       updateBy,
			LastUpdateTime: now,
		})
	}
	return datasetGroupIDMap, newGroups
}

func (s *VisualizationService) createAppDatasetTables(
	tablesInfo []map[string]interface{},
	datasetGroupIDMap map[int64]int64,
	datasourceIDMap map[int64]int64,
	folderID int64,
) (map[int64]int64, error) {
	datasetTableIDMap := make(map[int64]int64)
	for _, tableInfo := range tablesInfo {
		oldID, ok := extractInt64Value(tableInfo["id"])
		if !ok || oldID <= 0 {
			continue
		}
		newID := int64(uuid.New().ID())
		datasetTableIDMap[oldID] = newID

		newGroupID := resolveMappedID(tableInfo, "datasetGroupId", datasetGroupIDMap, folderID)
		newDatasourceID := resolveRemappedID(tableInfo, "datasourceId", datasourceIDMap)

		name, _ := tableInfo["name"].(string)
		physicalTable, _ := tableInfo["tableName"].(string)
		tableType, _ := tableInfo["type"].(string)
		info, _ := tableInfo["info"].(string)
		sqlVariables, _ := tableInfo["sqlVariableDetails"].(string)
		table := &dataset.CoreDatasetTable{
			ID:             newID,
			Name:           stringPtrIfNotEmpty(name),
			DatasourceID:   int64PtrIfPositive(newDatasourceID),
			DatasetGroupID: newGroupID,
			PhysicalTable:  stringPtrIfNotEmpty(physicalTable),
			Type:           stringPtrIfNotEmpty(tableType),
			Info:           stringPtrIfNotEmpty(info),
			SQLVariables:   stringPtrIfNotEmpty(sqlVariables),
		}
		if err := s.datasetRepo.CreateTable(table); err != nil {
			return nil, fmt.Errorf("create dataset table: %w", err)
		}
	}
	return datasetTableIDMap, nil
}

func (s *VisualizationService) createAppDatasetFields(
	fieldsInfo []map[string]interface{},
	datasetGroupIDMap map[int64]int64,
	datasetTableIDMap map[int64]int64,
	datasourceIDMap map[int64]int64,
	folderID int64,
) (map[int64]int64, error) {
	datasetTableFieldIDMap := make(map[int64]int64)
	fields := make([]dataset.CoreDatasetTableField, 0, len(fieldsInfo))
	for _, fieldInfo := range fieldsInfo {
		oldID, ok := extractInt64Value(fieldInfo["id"])
		if !ok || oldID <= 0 {
			continue
		}
		newID := int64(uuid.New().ID())
		datasetTableFieldIDMap[oldID] = newID

		newGroupID := resolveMappedID(fieldInfo, "datasetGroupId", datasetGroupIDMap, folderID)
		newTableID := resolveRemappedID(fieldInfo, "datasetTableId", datasetTableIDMap)
		newDatasourceID := resolveRemappedID(fieldInfo, "datasourceId", datasourceIDMap)

		originName, _ := fieldInfo["originName"].(string)
		name, _ := fieldInfo["name"].(string)
		dataeaseName, _ := fieldInfo["dataeaseName"].(string)
		fieldShortName, _ := fieldInfo["fieldShortName"].(string)
		groupType, _ := fieldInfo["groupType"].(string)
		fieldType, _ := fieldInfo["type"].(string)
		params, _ := fieldInfo["params"].(string)

		deType := extractIntValueOrDefault(fieldInfo["deType"], 0)
		deExtractType := extractIntValueOrDefault(fieldInfo["deExtractType"], 0)
		extField := extractIntValueOrDefault(fieldInfo["extField"], 0)
		checked := true

		fields = append(fields, dataset.CoreDatasetTableField{
			ID:             newID,
			DatasourceID:   int64PtrIfPositive(newDatasourceID),
			DatasetTableID: int64PtrIfPositive(newTableID),
			DatasetGroupID: newGroupID,
			OriginName:     stringPtrAllowEmpty(originName),
			Name:           stringPtrIfNotEmpty(name),
			DataeaseName:   stringPtrIfNotEmpty(dataeaseName),
			FieldShortName: stringPtrIfNotEmpty(fieldShortName),
			GroupType:      stringPtrIfNotEmpty(groupType),
			Type:           stringPtrIfNotEmpty(fieldType),
			DeType:         &deType,
			DeExtractType:  &deExtractType,
			ExtField:       &extField,
			Checked:        &checked,
			Params:         stringPtrIfNotEmpty(params),
		})
	}

	for i := range fields {
		if fields[i].OriginName == nil {
			continue
		}
		origin := replaceIDs(*fields[i].OriginName, datasetTableFieldIDMap)
		fields[i].OriginName = &origin
	}
	if err := s.datasetRepo.BatchCreateFields(fields); err != nil {
		return nil, fmt.Errorf("create dataset fields: %w", err)
	}
	return datasetTableFieldIDMap, nil
}

func (s *VisualizationService) persistAppDatasetGroups(
	groups []dataset.CoreDatasetGroup,
	datasetGroupIDMap, datasetTableIDMap, datasetTableFieldIDMap, datasourceIDMap map[int64]int64,
) error {
	for i := range groups {
		if groups[i].Info != nil {
			info := replaceIDs(*groups[i].Info, datasetGroupIDMap)
			info = replaceIDs(info, datasetTableIDMap)
			info = replaceIDs(info, datasetTableFieldIDMap)
			info = replaceIDs(info, datasourceIDMap)
			groups[i].Info = &info
		}
		if err := s.datasetRepo.CreateGroup(&groups[i]); err != nil {
			return fmt.Errorf("create dataset group: %w", err)
		}
	}
	return nil
}

func (s *VisualizationService) remapAppImportIDs(
	req *visualization.SaveRequest,
	datasetGroupIDMap, datasetTableIDMap, datasetTableFieldIDMap, datasourceIDMap map[int64]int64,
) {
	if req.ComponentData != nil {
		componentData := *req.ComponentData
		componentData = replaceIDs(componentData, datasetGroupIDMap)
		componentData = replaceIDs(componentData, datasetTableIDMap)
		componentData = replaceIDs(componentData, datasetTableFieldIDMap)
		componentData = replaceIDs(componentData, datasourceIDMap)
		req.ComponentData = &componentData
	}

	if req.CanvasViewInfo == nil {
		return
	}
	for viewID, viewData := range req.CanvasViewInfo {
		viewJSON, err := json.Marshal(viewData)
		if err != nil {
			continue
		}
		viewStr := string(viewJSON)
		viewStr = replaceIDs(viewStr, datasourceIDMap)
		viewStr = replaceIDs(viewStr, datasetTableIDMap)
		viewStr = replaceIDs(viewStr, datasetTableFieldIDMap)
		viewStr = replaceIDs(viewStr, datasetGroupIDMap)

		var newViewData map[string]interface{}
		if err := json.Unmarshal([]byte(viewStr), &newViewData); err != nil {
			continue
		}
		newViewData["dataFrom"] = "dataset"
		if tableID, ok := newViewData["tableId"]; !ok || tableID == nil || tableID == float64(0) {
			if sourceTableID, ok := newViewData["sourceTableId"]; ok {
				newViewData["tableId"] = sourceTableID
			}
		}
		req.CanvasViewInfo[viewID] = newViewData
	}
}

func parseJSONFieldFromAppData[T any](appData map[string]json.RawMessage, key string) (T, error) {
	var zero T
	raw, ok := appData[key]
	if !ok {
		return zero, nil
	}
	var result T
	if err := json.Unmarshal(raw, &result); err != nil {
		return zero, err
	}
	return result, nil
}

func resolveMappedID(info map[string]interface{}, key string, idMap map[int64]int64, fallback int64) int64 {
	oldID, _ := extractInt64Value(info[key])
	if mappedID, ok := idMap[oldID]; ok {
		return mappedID
	}
	return fallback
}

func resolveRemappedID(info map[string]interface{}, key string, idMap map[int64]int64) int64 {
	oldID, _ := extractInt64Value(info[key])
	if mappedID, ok := idMap[oldID]; ok {
		return mappedID
	}
	return oldID
}

func (s *VisualizationService) applyInheritedPermissionsOnCreate(resourceID int64, resourceName string, pid *int64, visualizationType *string) error {
	if s.resourcePermService == nil || pid == nil || *pid <= 0 {
		return nil
	}
	return s.resourcePermService.InheritParentResourcePermissions(*pid, resourceID, resourceName, normalizeVisualizationResourceType(visualizationType))
}

func (s *VisualizationService) BackfillGovernedResources() (*VisualizationGovernanceBackfillReport, error) {
	return s.BackfillGovernedVisualizationResourcesWithOptions(nil)
}

func (s *VisualizationService) BackfillGovernedVisualizationResourcesWithOptions(options *GovernanceBackfillOptions) (*VisualizationGovernanceBackfillReport, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("visualization repository not initialized")
	}
	if s.resourcePermService == nil {
		return nil, fmt.Errorf("resource permission service not initialized")
	}

	normalized := normalizeGovernanceBackfillOptions(options)
	items, err := s.repo.ListAllByTypesBatch(nil, normalized.AfterID, normalized.Limit, normalized.OrgID)
	if err != nil {
		return nil, err
	}

	report := newGovernanceBackfillReport("visualization", normalized)
	for _, item := range items {
		if item == nil || item.ID <= 0 {
			continue
		}
		resourceType := normalizeVisualizationResourceType(item.Type)
		report.observe(item.ID)
		if item.PID == nil || *item.PID <= 0 {
			report.addSkipped(item.ID, resourceType, 0, GovernanceBackfillSkipReasonMissingParent)
			continue
		}
		inherited, err := s.resourcePermService.TryInheritParentResourcePermissions(*item.PID, item.ID, item.Name, resourceType)
		if err != nil {
			return nil, err
		}
		if !inherited {
			report.addSkipped(item.ID, resourceType, *item.PID, GovernanceBackfillSkipReasonParentNotGoverned)
			continue
		}
		report.addGoverned(item.ID)
	}

	return report, nil
}

func normalizeVisualizationResourceType(visualizationType *string) string {
	if visualizationType != nil {
		switch *visualizationType {
		case visualizationTypeDataV, permission.ResourceTypeScreen:
			return permission.ResourceTypeScreen
		}
	}
	return permission.ResourceTypeDashboard
}

func (s *VisualizationService) Copy(req *visualization.CopyRequest, updateBy string) (int64, error) {
	if req == nil {
		return 0, fmt.Errorf("copy request is required")
	}
	if req.ID <= 0 {
		return 0, fmt.Errorf("source id is required")
	}
	if req.Name == "" {
		return 0, fmt.Errorf("name is required")
	}

	source, err := s.repo.GetByID(req.ID)
	if err != nil {
		return 0, fmt.Errorf("source visualization not found: %w", err)
	}

	newDvID := int64(uuid.New().ID())
	copyID := int64(uuid.New().ID()) / 100

	componentData, err := s.copyVisualizationChartViewsAndRemapComponentData(req.ID, newDvID, copyID, source.ComponentData)
	if err != nil {
		return 0, err
	}
	s.copyVisualizationRelations(copyID)

	nodeType := resolveCopiedVisualizationNodeType(source.NodeType, req.NodeType)
	visualizationType := resolveCopiedVisualizationType(source.Type, req.Type)
	mobileLayout := resolveCopiedVisualizationMobileLayout(source.MobileLayout, req.MobileLayout)

	now := time.Now().UnixMilli()
	status := resolveCopiedVisualizationStatus(source.Status, nodeType)

	newViz := &visualization.DataVisualizationInfo{
		ID:              newDvID,
		Name:            req.Name,
		PID:             req.PID,
		NodeType:        nodeType,
		Type:            visualizationType,
		CanvasStyleData: source.CanvasStyleData,
		ComponentData:   componentData,
		MobileLayout:    mobileLayout,
		Status:          status,
		ContentID:       source.ContentID,
		CheckVersion:    source.CheckVersion,
		CreateTime:      &now,
		UpdateTime:      &now,
		CreateBy:        &updateBy,
		UpdateBy:        &updateBy,
	}

	if err := s.repo.Create(newViz); err != nil {
		return 0, fmt.Errorf("create copied visualization: %w", err)
	}
	if err := s.applyInheritedPermissionsOnCreate(newViz.ID, req.Name, req.PID, visualizationType); err != nil {
		_ = s.repo.DeleteLogic(newViz.ID, updateBy)
		return 0, err
	}

	return newViz.ID, nil
}

func (s *VisualizationService) copyVisualizationChartViewsAndRemapComponentData(sourceID, newDvID, copyID int64, componentData *string) (*string, error) {
	if err := s.repo.CopyChartViews(sourceID, newDvID, copyID, "core"); err != nil {
		return nil, fmt.Errorf("copy chart views (core): %w", err)
	}
	// Snapshot table copy is best-effort; the source may have no snapshot rows.
	_ = s.repo.CopyChartViews(sourceID, newDvID, copyID, "snapshot")

	viewMapping, err := s.repo.GetCopiedChartViewMapping(copyID)
	if err != nil || len(viewMapping) == 0 || componentData == nil {
		return componentData, nil
	}

	remapped := *componentData
	for oldID, newID := range viewMapping {
		remapped = strings.ReplaceAll(remapped, strconv.FormatInt(oldID, 10), strconv.FormatInt(newID, 10))
	}

	return &remapped, nil
}

func (s *VisualizationService) copyVisualizationRelations(copyID int64) {
	_ = s.repo.CopyLinkages(copyID)
	_ = s.repo.CopyLinkageFields(copyID)
	_ = s.repo.CopyLinkJumps(copyID)
	_ = s.repo.CopyLinkJumpInfos(copyID)
	_ = s.repo.CopyLinkJumpTargetInfos(copyID)
}

func resolveCopiedVisualizationNodeType(sourceNodeType, requestedNodeType *string) *string {
	nodeType := sourceNodeType
	if requestedNodeType != nil && *requestedNodeType != "" {
		nodeType = requestedNodeType
	}
	if nodeType != nil && *nodeType != visualizationNodeTypeFolder {
		leaf := "leaf"
		return &leaf
	}
	return nodeType
}

func resolveCopiedVisualizationType(sourceType, requestedType *string) *string {
	if requestedType != nil && *requestedType != "" {
		return requestedType
	}
	return sourceType
}

func resolveCopiedVisualizationMobileLayout(sourceMobileLayout, requestedMobileLayout *bool) *bool {
	if requestedMobileLayout != nil {
		return requestedMobileLayout
	}
	return sourceMobileLayout
}

func resolveCopiedVisualizationStatus(sourceStatus *int, nodeType *string) *int {
	if nodeType != nil && *nodeType == visualizationNodeTypeFolder {
		folderStatus := 1
		return &folderStatus
	}
	return sourceStatus
}

func (s *VisualizationService) Update(req *visualization.UpdateRequest, updateBy string) error {
	v, err := s.repo.GetByID(req.ID)
	if err != nil {
		return fmt.Errorf("visualization not found: %w", err)
	}

	if req.Name != nil {
		v.Name = *req.Name
	}
	if req.PID != nil {
		v.PID = req.PID
	}
	if req.Type != nil {
		v.Type = req.Type
	}
	if req.CanvasStyleData != nil {
		v.CanvasStyleData = req.CanvasStyleData
	}
	if req.ComponentData != nil {
		v.ComponentData = req.ComponentData
	}
	if req.MobileLayout != nil {
		v.MobileLayout = req.MobileLayout
	}
	if req.ContentID != nil {
		v.ContentID = req.ContentID
	}
	if req.CheckVersion != nil {
		v.CheckVersion = req.CheckVersion
	}
	if req.Status != nil {
		v.Status = req.Status
	}
	now := time.Now().UnixMilli()
	v.UpdateTime = &now
	v.UpdateBy = &updateBy
	s.saveChartViewsFromVisualization(req.ComponentData, req.ID, req.CanvasViewInfo)

	return s.repo.Update(v)
}

func (s *VisualizationService) saveChartViewsFromVisualization(componentData *string, dvID int64, canvasViewInfo map[string]map[string]interface{}) {
	if len(canvasViewInfo) == 0 || componentData == nil {
		return
	}

	componentDataStr := *componentData
	now := time.Now().UnixMilli()

	for chartIDStr, chartData := range canvasViewInfo {
		if !strings.Contains(componentDataStr, chartIDStr) {
			continue
		}

		chartID, err := strconv.ParseInt(chartIDStr, 10, 64)
		if err != nil || chartID <= 0 {
			continue
		}

		view := buildSnapshotChartViewFromMap(chartData)
		if view == nil {
			continue
		}

		view.ID = chartID
		view.SceneID = &dvID
		view.UpdateTime = &now
		if view.CreateTime == nil || *view.CreateTime == 0 {
			view.CreateTime = &now
		}

		_ = s.repo.SaveSnapshotChartView(view)
	}
}

func buildSnapshotChartViewFromMap(data map[string]interface{}) *visualization.SnapshotCanvasChartView { //nolint:gocyclo // mirrors chart view DTO-to-record mapping
	if len(data) == 0 {
		return nil
	}

	view := &visualization.SnapshotCanvasChartView{}

	if v, ok := stringFromAnyMap(data, "title"); ok {
		view.Title = &v
	}
	if v, ok := stringFromAnyMap(data, "type"); ok {
		view.Type = &v
	}
	if v, ok := stringFromAnyMap(data, "render"); ok {
		view.Render = &v
	}
	if v, ok := stringFromAnyMap(data, "resultMode"); ok {
		view.ResultMode = &v
	}
	if v, ok := stringFromAnyMap(data, "dataFrom"); ok {
		view.DataFrom = &v
	}
	if v, ok := stringFromAnyMap(data, "stylePriority"); ok {
		view.StylePriority = &v
	}
	if v, ok := stringFromAnyMap(data, "chartType"); ok {
		view.ChartType = &v
	}
	if v, ok := stringFromAnyMap(data, "refreshUnit"); ok {
		view.RefreshUnit = &v
	}
	if v, ok := stringFromAnyMap(data, "flowMapStartName"); ok {
		view.FlowMapStartName = &v
	}
	if v, ok := stringFromAnyMap(data, "flowMapEndName"); ok {
		view.FlowMapEndName = &v
	}
	if v, ok := stringFromAnyMap(data, "createBy"); ok {
		view.CreateBy = &v
	}

	if v, ok := int64FromAnyMap(data, "tableId"); ok {
		view.TableID = &v
	}
	if v, ok := int64FromAnyMap(data, "createTime"); ok {
		view.CreateTime = &v
	}

	if v, ok := intFromAnyMap(data, "resultCount"); ok {
		view.ResultCount = &v
	}
	if v, ok := intFromAnyMap(data, "refreshTime"); ok {
		view.RefreshTime = &v
	}

	if v, ok := data["isPlugin"]; ok {
		b := boolFromAnyMap(v)
		view.IsPlugin = &b
	}
	if v, ok := data["refreshViewEnable"]; ok {
		b := boolFromAnyMap(v)
		view.RefreshViewEnable = &b
	}
	if v, ok := data["linkageActive"]; ok {
		b := boolFromAnyMap(v)
		view.LinkageActive = &b
	}
	if v, ok := data["jumpActive"]; ok {
		b := boolFromAnyMap(v)
		view.JumpActive = &b
	}
	if v, ok := data["aggregate"]; ok {
		b := boolFromAnyMap(v)
		view.Aggregate = &b
	}

	if v, ok := marshalJSONFieldFromMap(data, "xAxis"); ok {
		view.XAxis = &v
	}
	if v, ok := marshalJSONFieldFromMap(data, "xAxisExt"); ok {
		view.XAxisExt = &v
	}
	if v, ok := marshalJSONFieldFromMap(data, "yAxis"); ok {
		view.YAxis = &v
	}
	if v, ok := marshalJSONFieldFromMap(data, "yAxisExt"); ok {
		view.YAxisExt = &v
	}
	if v, ok := marshalJSONFieldFromMap(data, "extStack"); ok {
		view.ExtStack = &v
	}
	if v, ok := marshalJSONFieldFromMap(data, "extBubble"); ok {
		view.ExtBubble = &v
	}
	if v, ok := marshalJSONFieldFromMap(data, "extLabel"); ok {
		view.ExtLabel = &v
	}
	if v, ok := marshalJSONFieldFromMap(data, "extTooltip"); ok {
		view.ExtTooltip = &v
	}
	if v, ok := marshalJSONFieldFromMap(data, "customAttr"); ok {
		view.CustomAttr = &v
	}
	if v, ok := marshalJSONFieldFromMap(data, "customAttrMobile"); ok {
		view.CustomAttrMobile = &v
	}
	if v, ok := marshalJSONFieldFromMap(data, "customStyle"); ok {
		view.CustomStyle = &v
	}
	if v, ok := marshalJSONFieldFromMap(data, "customStyleMobile"); ok {
		view.CustomStyleMobile = &v
	}
	if v, ok := marshalJSONFieldFromMap(data, "customFilter"); ok {
		view.CustomFilter = &v
	}
	if v, ok := marshalJSONFieldFromMap(data, "drillFields"); ok {
		view.DrillFields = &v
	}
	if v, ok := marshalJSONFieldFromMap(data, "senior"); ok {
		view.Senior = &v
	}
	if v, ok := marshalJSONFieldFromMap(data, "snapshot"); ok {
		view.Snapshot = &v
	}
	if v, ok := marshalJSONFieldFromMap(data, "viewFields"); ok {
		view.ViewFields = &v
	}
	if v, ok := marshalJSONFieldFromMap(data, "extColor"); ok {
		view.ExtColor = &v
	}
	if v, ok := marshalJSONFieldFromMap(data, "sortPriority"); ok {
		view.SortPriority = &v
	}

	return view
}

func stringFromAnyMap(m map[string]interface{}, key string) (string, bool) {
	v, exists := m[key]
	if !exists {
		return "", false
	}
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

func int64FromAnyMap(m map[string]interface{}, key string) (int64, bool) {
	v, exists := m[key]
	if !exists {
		return 0, false
	}
	return int64FromVisualizationAny(v)
}

func intFromAnyMap(m map[string]interface{}, key string) (int, bool) {
	v, ok := int64FromAnyMap(m, key)
	if !ok {
		return 0, false
	}
	return int(v), true
}

func boolFromAnyMap(v interface{}) bool {
	switch b := v.(type) {
	case bool:
		return b
	case string:
		return strings.EqualFold(strings.TrimSpace(b), "true") || strings.TrimSpace(b) == "1"
	case float64:
		return b != 0
	case int:
		return b != 0
	case int64:
		return b != 0
	case json.Number:
		s := strings.ToLower(strings.TrimSpace(b.String()))
		return s == "true" || s == "1"
	default:
		return false
	}
}

func marshalJSONFieldFromMap(m map[string]interface{}, key string) (string, bool) {
	v, exists := m[key]
	if !exists {
		return "", false
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "", false
	}
	return string(b), true
}

func int64FromVisualizationAny(v interface{}) (int64, bool) {
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

func (s *VisualizationService) Detail(req *visualization.DetailRequest) (*visualization.DataVisualizationInfo, error) {
	return s.repo.GetByID(req.ID.Int64())
}

func (s *VisualizationService) List(req *visualization.ListRequest) (*visualization.ListResponse, error) {
	list, total, err := s.repo.Query(req)
	if err != nil {
		return nil, err
	}

	current := req.Current
	if current < 1 {
		current = 1
	}
	size := req.Size
	if size < 1 {
		size = 10
	}

	return &visualization.ListResponse{
		List:    list,
		Total:   total,
		Current: current,
		Size:    size,
	}, nil
}

func (s *VisualizationService) FindRecent(req *visualization.WorkbranchQueryRequest, uid int64) ([]visualization.VisualizationResourceVO, error) {
	results, err := s.repo.FindRecent(uid, req)
	if err != nil {
		return nil, err
	}
	// The VO fields are already strings from SQL scan.
	return results, nil
}

func (s *VisualizationService) InteractiveTree(busiFlag string) ([]*visualization.DataVisualizationInfo, error) {
	types, err := resolveInteractiveVisualizationTypes(busiFlag)
	if err != nil {
		return nil, err
	}
	return s.repo.ListAllByTypes(types)
}

func (s *VisualizationService) DeleteLogic(id int64, updateBy string) error {
	var chartIDs []int64
	if s.thresholdService != nil {
		viz, err := s.repo.GetByID(id)
		if err == nil && viz != nil && viz.ComponentData != nil {
			chartIDs = extractChartIDsFromComponentData(json.RawMessage(*viz.ComponentData))
		}
	}

	if err := s.repo.DeleteLogic(id, updateBy); err != nil {
		return err
	}

	if s.thresholdService != nil && len(chartIDs) > 0 {
		for _, chartID := range chartIDs {
			_ = s.thresholdService.DeleteWithChart(context.Background(), chartID, "core")
		}
	}

	return nil
}

// extractChartIDsFromComponentData parses ComponentData JSON and returns chart component IDs.
func extractChartIDsFromComponentData(componentData json.RawMessage) []int64 {
	if len(componentData) == 0 {
		return nil
	}
	var components []struct {
		Component string `json:"component"`
		ID        int64  `json:"id"`
		PropValue []struct {
			ID        int64  `json:"id"`
			Component string `json:"component"`
		} `json:"propValue"`
	}
	if err := json.Unmarshal(componentData, &components); err != nil {
		return nil
	}
	seen := make(map[int64]struct{})
	chartIDs := make([]int64, 0)
	appendChartID := func(component string, id int64) {
		if component != "UserView" || id <= 0 {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		chartIDs = append(chartIDs, id)
	}
	for _, c := range components {
		appendChartID(c.Component, c.ID)
		for _, pv := range c.PropValue {
			appendChartID(pv.Component, pv.ID)
		}
	}
	return chartIDs
}

func (s *VisualizationService) FindDvType(id int64) (string, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		return "", err
	}
	if item.Type == nil {
		return "", nil
	}
	return *item.Type, nil
}

func resolveInteractiveVisualizationTypes(busiFlag string) ([]string, error) {
	flag := busiFlag
	switch flag {
	case "", visualizationBusiFlagDashboardV:
		return []string{visualizationTypeDashboard, visualizationTypeDataV}, nil
	case visualizationNodeTypePanel, visualizationTypeDashboard:
		return []string{visualizationTypeDashboard}, nil
	case "screen", visualizationTypeDataV:
		return []string{visualizationTypeDataV}, nil
	default:
		return nil, fmt.Errorf("unsupported busiFlag: %s", flag)
	}
}

func (s *VisualizationService) NameCheck(req *visualization.NameCheckRequest) (string, error) {
	if req == nil {
		return compatResultSuccess, nil
	}
	var excludeID *int64
	if req.ID > 0 {
		excludeID = &req.ID
	}
	count, err := s.repo.CountByNameAndPID(req.Name, req.PID, excludeID)
	if err != nil {
		return "", err
	}
	if count > 0 {
		return "repeat", nil
	}
	return compatResultSuccess, nil
}

func (s *VisualizationService) CheckCanvasChange(req *visualization.CanvasChangeRequest) (string, error) {
	if req == nil || req.ID <= 0 {
		return "", nil
	}
	item, err := s.repo.GetByID(req.ID)
	if err != nil {
		return "", err
	}
	if req.ContentID != nil && item.ContentID != nil && *req.ContentID != "" && *item.ContentID != "" && *req.ContentID != *item.ContentID {
		return "Repeat", nil
	}
	if req.CheckVersion != nil && item.CheckVersion != nil && *req.CheckVersion != "" && *item.CheckVersion != "" && *req.CheckVersion != *item.CheckVersion {
		return "Repeat", nil
	}
	return "", nil
}

func (s *VisualizationService) UpdateBase(req *visualization.UpdateRequest, updateBy string) error {
	return s.Update(req, updateBy)
}

func (s *VisualizationService) Move(req *visualization.MoveRequest, updateBy string) error {
	if req == nil {
		return nil
	}
	return s.Update(&visualization.UpdateRequest{ID: req.ID, PID: req.PID}, updateBy)
}

func (s *VisualizationService) UpdatePublishStatus(req *visualization.UpdateRequest, updateBy string) (*visualization.DataVisualizationInfo, error) {
	if err := s.Update(req, updateBy); err != nil {
		return nil, err
	}
	return s.repo.GetByID(req.ID)
}

func (s *VisualizationService) RecoverToPublished(id int64, updateBy string) (*visualization.DataVisualizationInfo, error) {
	status := 1
	if err := s.Update(&visualization.UpdateRequest{ID: id, Status: &status}, updateBy); err != nil {
		return nil, err
	}
	return s.repo.GetByID(id)
}

func (s *VisualizationService) ViewDetailList(dvID int64) ([]chart.CoreChartView, error) {
	rows, err := s.repo.GetChartViewsBySceneID(dvID)
	if err != nil {
		return nil, err
	}
	if len(rows) > 0 {
		return rows, nil
	}
	return s.repo.GetSnapshotChartViewsBySceneID(dvID)
}

func (s *VisualizationService) AppCanvasNameCheck(req *visualization.AppCanvasNameCheckRequest) (string, error) {
	if req == nil {
		return compatResultSuccess, nil
	}
	if s.datasetRepo == nil {
		return compatResultSuccess, nil
	}
	name := strings.TrimSpace(req.DatasetFolderName)
	if name == "" {
		return compatResultSuccess, nil
	}
	var pid int64
	if req.DatasetFolderPid != nil {
		pid = *req.DatasetFolderPid
	}
	count, err := s.datasetRepo.CountFolderByNameAndPID(name, pid)
	if err != nil {
		return "", err
	}
	if count > 0 {
		return "repeat", nil
	}
	return compatResultSuccess, nil
}

func (s *VisualizationService) RecordExportLog(req *visualization.ExportLogRequest, userID *int64, username *string, ipAddress *string, userAgent *string, logType string) error {
	if req == nil || req.ID == nil || *req.ID <= 0 || s.auditService == nil {
		return nil
	}
	actionName := "导出资源"
	resourceTypeValue := "DASHBOARD"
	switch logType {
	case "app":
		actionName = "导出应用模板"
	case "template":
		actionName = "导出样式模板"
	case "pdf":
		actionName = "导出PDF"
	case "img":
		actionName = "导出图片"
	}
	if strings.EqualFold(req.Type, "screen") || strings.EqualFold(req.Type, visualizationTypeDataV) {
		resourceTypeValue = "SCREEN"
	}
	_, err := s.auditService.CreateAuditLog(&audit.AuditLogCreateRequest{
		UserID:       userID,
		Username:     username,
		ActionType:   audit.ActionTypeDataAccess,
		ActionName:   actionName,
		ResourceType: &resourceTypeValue,
		ResourceID:   req.ID,
		Operation:    audit.OperationExport,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	})
	return err
}

func (s *VisualizationService) Export2AppCheck(req *visualization.Export2AppCheckRequest) (*visualization.Export2AppCheckResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("export2AppCheck request is required")
	}

	data, err := s.loadExport2AppCheckData(req)
	if err != nil {
		return nil, err
	}
	if err := validateExport2AppDatasources(data.datasources); err != nil {
		return nil, err
	}

	return &visualization.Export2AppCheckResponse{
		CheckStatus:            true,
		CheckMes:               compatResultSuccess,
		ChartViewsInfo:         stringifyExportPayload(data.chartViews),
		DatasetGroupsInfo:      stringifyExportPayload(data.datasetGroups),
		DatasetTablesInfo:      stringifyExportPayload(data.datasetTables),
		DatasetTableFieldsInfo: stringifyExportPayload(data.datasetTableFields),
		DatasourceInfo:         stringifyExportPayload(data.datasources),
		DatasourceTaskInfo:     stringifyExportPayload(data.datasourceTasks),
		LinkJumps:              stringifyExportPayload(data.linkJumps),
		LinkJumpInfos:          stringifyExportPayload(data.linkJumpInfos),
		LinkJumpTargetInfos:    stringifyExportPayload(data.linkJumpTargets),
		Linkages:               stringifyExportPayload(data.linkages),
		LinkageFields:          stringifyExportPayload(data.linkageFields),
	}, nil
}

type export2AppCheckData struct {
	chartViews         []map[string]interface{}
	datasetGroups      []map[string]interface{}
	datasetTables      []map[string]interface{}
	datasetTableFields []map[string]interface{}
	datasources        []map[string]interface{}
	datasourceTasks    []map[string]interface{}
	linkages           []map[string]interface{}
	linkageFields      []map[string]interface{}
	linkJumps          []map[string]interface{}
	linkJumpInfos      []map[string]interface{}
	linkJumpTargets    []map[string]interface{}
}

func (s *VisualizationService) loadExport2AppCheckData(req *visualization.Export2AppCheckRequest) (*export2AppCheckData, error) {
	chartViews, err := s.repo.FindChartViewsByIDs(req.ViewIDs)
	if err != nil {
		return nil, fmt.Errorf("query chart views: %w", err)
	}
	datasetGroups, err := s.repo.FindDatasetGroupsByIDs(req.DsIDs)
	if err != nil {
		return nil, fmt.Errorf("query dataset groups: %w", err)
	}
	datasetTables, err := s.repo.FindDatasetTablesByGroupIDs(req.DsIDs)
	if err != nil {
		return nil, fmt.Errorf("query dataset tables: %w", err)
	}
	datasetTableFields, err := s.repo.FindDatasetTableFieldsByGroupIDs(req.DsIDs)
	if err != nil {
		return nil, fmt.Errorf("query dataset table fields: %w", err)
	}
	datasources, err := s.repo.FindDatasourcesByGroupIDs(req.DsIDs)
	if err != nil {
		return nil, fmt.Errorf("query datasources: %w", err)
	}
	datasourceTasks, err := s.repo.FindDatasourceTasksByGroupIDs(req.DsIDs)
	if err != nil {
		return nil, fmt.Errorf("query datasource tasks: %w", err)
	}
	linkages, err := s.repo.FindLinkagesByDvID(req.DvID)
	if err != nil {
		return nil, fmt.Errorf("query linkages: %w", err)
	}
	linkageFields, err := s.repo.FindLinkageFieldsByDvID(req.DvID)
	if err != nil {
		return nil, fmt.Errorf("query linkage fields: %w", err)
	}
	linkJumps, err := s.repo.FindLinkJumpsByDvID(req.DvID)
	if err != nil {
		return nil, fmt.Errorf("query link jumps: %w", err)
	}
	linkJumpInfos, err := s.repo.FindLinkJumpInfosByDvID(req.DvID)
	if err != nil {
		return nil, fmt.Errorf("query link jump infos: %w", err)
	}
	linkJumpTargets, err := s.repo.FindLinkJumpTargetViewInfosByDvID(req.DvID)
	if err != nil {
		return nil, fmt.Errorf("query link jump target view infos: %w", err)
	}
	return &export2AppCheckData{
		chartViews:         chartViews,
		datasetGroups:      datasetGroups,
		datasetTables:      datasetTables,
		datasetTableFields: datasetTableFields,
		datasources:        datasources,
		datasourceTasks:    datasourceTasks,
		linkages:           linkages,
		linkageFields:      linkageFields,
		linkJumps:          linkJumps,
		linkJumpInfos:      linkJumpInfos,
		linkJumpTargets:    linkJumpTargets,
	}, nil
}

func validateExport2AppDatasources(datasources []map[string]interface{}) error {
	if len(datasources) == 0 {
		return fmt.Errorf("当前不存在数据源无法导出")
	}
	for _, ds := range datasources {
		if dsType, ok := ds["type"]; ok {
			typeStr := fmt.Sprintf("%v", dsType)
			if strings.Contains(strings.ToUpper(typeStr), "API") {
				return fmt.Errorf("包含API数据源不支持导出")
			}
		}
	}
	return nil
}

func stringifyExportPayload(rows []map[string]interface{}) []map[string]interface{} {
	if rows == nil {
		return []map[string]interface{}{}
	}
	return stringifyExportIDs(rows)
}

func stringifyExportIDs(rows []map[string]interface{}) []map[string]interface{} {
	for _, row := range rows {
		for key, value := range row {
			lowerKey := strings.ToLower(key)
			if !strings.Contains(lowerKey, "id") {
				continue
			}
			switch v := value.(type) {
			case int64:
				row[key] = fmt.Sprintf("%d", v)
			case int32:
				row[key] = fmt.Sprintf("%d", v)
			case int:
				row[key] = fmt.Sprintf("%d", v)
			case uint64:
				row[key] = fmt.Sprintf("%d", v)
			case uint32:
				row[key] = fmt.Sprintf("%d", v)
			case uint:
				row[key] = fmt.Sprintf("%d", v)
			}
		}
	}
	return rows
}

const (
	newFromInnerTemplate  = "new_inner_template"
	newFromOuterTemplate  = "new_outer_template"
	newFromMarketTemplate = "new_market_template"
	newFromLocalFile      = "localFile"
	compatResultSuccess   = "success"
)

type localTemplateFileDTO struct {
	Name            json.RawMessage `json:"name"`
	DvType          string          `json:"dvType"`
	Version         int             `json:"version"`
	CanvasStyleData json.RawMessage `json:"canvasStyleData"`
	ComponentData   json.RawMessage `json:"componentData"`
	DynamicData     json.RawMessage `json:"dynamicData"`
	StaticResource  json.RawMessage `json:"staticResource"`
	AppData         json.RawMessage `json:"appData"`
}

func (s *VisualizationService) Decompression(req *visualization.DecompressionRequest) (*visualization.DecompressionResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("decompression request is required")
	}
	newDvID := int64(uuid.New().ID())

	source, err := s.loadDecompressionTemplateSource(req)
	if err != nil {
		return nil, err
	}

	appDataStr := processAppData(source.appData, newDvID)

	hasAppData := strings.TrimSpace(appDataStr) != ""
	canvasViewInfo, err := s.processDynamicData(source.dynamicData, newDvID, &source.templateData, &appDataStr, hasAppData)
	if err != nil {
		return nil, err
	}

	resp := &visualization.DecompressionResponse{
		ID:              fmt.Sprintf("%d", newDvID),
		Name:            source.name,
		Type:            source.dvType,
		Version:         source.version,
		CanvasStyleData: source.templateStyle,
		ComponentData:   source.templateData,
		AppData:         appDataStr,
		CanvasViewInfo:  canvasViewInfo,
	}

	if s.staticService != nil && strings.TrimSpace(source.staticResource) != "" {
		if err := s.staticService.SaveFilesToServe(source.staticResource); err != nil {
			return nil, err
		}
	}

	return resp, nil
}

type decompressionTemplateSource struct {
	templateStyle  string
	templateData   string
	dynamicData    string
	name           string
	dvType         string
	appData        string
	staticResource string
	version        int
}

func (s *VisualizationService) loadDecompressionTemplateSource(req *visualization.DecompressionRequest) (*decompressionTemplateSource, error) {
	switch req.NewFrom {
	case newFromInnerTemplate:
		return s.loadInnerDecompressionTemplateSource(req)
	case newFromOuterTemplate:
		return buildDirectDecompressionTemplateSource(req, 3), nil
	case newFromLocalFile:
		version := req.Version
		if version <= 0 {
			version = 3
		}
		return buildDirectDecompressionTemplateSource(req, version), nil
	case newFromMarketTemplate:
		return loadMarketDecompressionTemplateSource(req)
	default:
		return nil, fmt.Errorf("unsupported newFrom: %s", req.NewFrom)
	}
}

func (s *VisualizationService) loadInnerDecompressionTemplateSource(req *visualization.DecompressionRequest) (*decompressionTemplateSource, error) {
	if req.TemplateID == nil || *req.TemplateID <= 0 {
		return nil, fmt.Errorf("templateId is required for new_inner_template")
	}
	if s.templateService == nil {
		return nil, fmt.Errorf("template service is not initialized")
	}

	tmpl, err := s.templateService.GetTemplate(*req.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}
	_ = s.templateService.IncrementUseCount(tmpl.ID)

	return &decompressionTemplateSource{
		templateStyle: tmpl.TemplateStyle,
		templateData:  tmpl.TemplateData,
		dynamicData:   tmpl.DynamicData,
		name:          tmpl.Name,
		dvType:        tmpl.DvType,
		appData:       tmpl.AppData,
		version:       tmpl.Version,
	}, nil
}

func buildDirectDecompressionTemplateSource(req *visualization.DecompressionRequest, version int) *decompressionTemplateSource {
	return &decompressionTemplateSource{
		templateStyle:  req.CanvasStyleData,
		templateData:   req.ComponentData,
		dynamicData:    req.DynamicData,
		name:           req.Name,
		dvType:         req.Type,
		appData:        req.AppData,
		staticResource: req.StaticResource,
		version:        version,
	}
}

func loadMarketDecompressionTemplateSource(req *visualization.DecompressionRequest) (*decompressionTemplateSource, error) {
	tmpl, err := fetchMarketTemplate(req.TemplateURL)
	if err != nil {
		return nil, err
	}

	return &decompressionTemplateSource{
		templateStyle:  normalizeJSONPayload(tmpl.CanvasStyleData),
		templateData:   normalizeJSONPayload(tmpl.ComponentData),
		dynamicData:    normalizeJSONPayload(tmpl.DynamicData),
		name:           tmpl.Name,
		dvType:         tmpl.DvType,
		appData:        normalizeJSONPayload(tmpl.AppData),
		staticResource: normalizeJSONPayload(tmpl.StaticResource),
		version:        tmpl.Version,
	}, nil
}

func (s *VisualizationService) DecompressionLocalFile(fileContent []byte) (*visualization.DecompressionResponse, error) {
	if len(fileContent) == 0 {
		return nil, fmt.Errorf("upload file is empty")
	}

	var payload localTemplateFileDTO
	if err := json.Unmarshal(fileContent, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse template file JSON: %w", err)
	}

	return s.Decompression(&visualization.DecompressionRequest{
		NewFrom:         newFromLocalFile,
		Name:            normalizeTemplateFileName(payload.Name),
		Type:            payload.DvType,
		Version:         payload.Version,
		CanvasStyleData: normalizeJSONPayload(payload.CanvasStyleData),
		ComponentData:   normalizeJSONPayload(payload.ComponentData),
		DynamicData:     normalizeJSONPayload(payload.DynamicData),
		AppData:         normalizeJSONPayload(payload.AppData),
		StaticResource:  normalizeJSONPayload(payload.StaticResource),
	})
}

func processAppData(appDataStr string, newDvID int64) string {
	if len(appDataStr) <= 10 {
		return appDataStr
	}
	var parsed map[string]json.RawMessage
	if err := json.Unmarshal([]byte(appDataStr), &parsed); err != nil {
		return appDataStr
	}
	visInfoRaw, ok := parsed["visualizationInfo"]
	if !ok {
		return appDataStr
	}
	var baseInfo struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(visInfoRaw, &baseInfo); err != nil {
		return appDataStr
	}
	if baseInfo.ID <= 0 {
		return appDataStr
	}
	return strings.ReplaceAll(appDataStr, fmt.Sprintf("%d", baseInfo.ID), fmt.Sprintf("%d", newDvID))
}

func fetchMarketTemplate(templateURL string) (*marketTemplateDTO, error) {
	if strings.TrimSpace(templateURL) == "" {
		return nil, fmt.Errorf("templateUrl is required for new_market_template")
	}
	if err := marketTemplateURLValidator(templateURL); err != nil {
		return nil, err
	}

	resp, err := marketTemplateHTTPClient.Get(templateURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch market template: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("market template fetch returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, marketTemplateMaxResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read market template response: %w", err)
	}
	if int64(len(body)) > marketTemplateMaxResponseBytes {
		return nil, fmt.Errorf("market template response exceeds %d bytes", marketTemplateMaxResponseBytes)
	}

	var tmpl marketTemplateDTO
	if err := json.Unmarshal(body, &tmpl); err != nil {
		return nil, fmt.Errorf("failed to parse market template JSON: %w", err)
	}

	return &tmpl, nil
}

func normalizeJSONPayload(raw json.RawMessage) string {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == jsonNullLiteral {
		return ""
	}

	var plain string
	if err := json.Unmarshal(raw, &plain); err == nil {
		return plain
	}

	return trimmed
}

func normalizeTemplateFileName(raw json.RawMessage) string {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == jsonNullLiteral {
		return ""
	}

	var plain string
	if err := json.Unmarshal(raw, &plain); err == nil {
		return plain
	}

	var wrapped struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(raw, &wrapped); err == nil {
		return wrapped.Name
	}

	return trimmed
}

func (s *VisualizationService) processDynamicData(dynamicData string, newDvID int64, templateData *string, appDataStr *string, hasAppData bool) (map[string]map[string]interface{}, error) {
	canvasViewInfo := make(map[string]map[string]interface{})
	if strings.TrimSpace(dynamicData) == "" {
		return canvasViewInfo, nil
	}

	dynamicMap, err := parseDynamicData(dynamicData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dynamicData: %w", err)
	}

	var extendRecords []auto.VisualizationTemplateExtendDatum

	for originViewIDStr, rawViewData := range dynamicMap {
		newViewID := int64(uuid.New().ID())
		originalViewJSON := rawViewData

		var viewMap map[string]interface{}
		if err := json.Unmarshal([]byte(rawViewData), &viewMap); err != nil {
			return nil, fmt.Errorf("failed to parse dynamicData view %s: %w", originViewIDStr, err)
		}

		if cf, ok := viewMap["customFilter"]; ok {
			if _, isSlice := cf.([]interface{}); isSlice {
				viewMap["customFilter"] = map[string]interface{}{}
			}
		}

		viewMap["id"] = newViewID
		viewMap["sceneId"] = newDvID
		viewMap["dataFrom"] = "template"

		if tableID, ok := extractInt64Value(viewMap["tableId"]); ok {
			viewMap["sourceTableId"] = tableID
			viewMap["tableId"] = nil
			if hasAppData {
				viewMap["tableId"] = tableID
			}
		}

		if _, err := json.Marshal(viewMap); err != nil {
			return nil, fmt.Errorf("failed to marshal dynamicData view %s: %w", originViewIDStr, err)
		}

		extendRecords = append(extendRecords, auto.VisualizationTemplateExtendDatum{
			ID:          int64(uuid.New().ID()),
			DvID:        newDvID,
			ViewID:      newViewID,
			ViewDetails: originalViewJSON,
			CopyFrom:    originViewIDStr,
			CopyID:      "",
		})

		*templateData = strings.ReplaceAll(*templateData, originViewIDStr, fmt.Sprintf("%d", newViewID))
		if *appDataStr != "" {
			*appDataStr = strings.ReplaceAll(*appDataStr, originViewIDStr, fmt.Sprintf("%d", newViewID))
		}

		canvasViewInfo[fmt.Sprintf("%d", newViewID)] = viewMap
	}

	if s.templateExtendDataRepo != nil && len(extendRecords) > 0 {
		if err := s.templateExtendDataRepo.BatchCreate(extendRecords); err != nil {
			return nil, err
		}
	}

	return canvasViewInfo, nil
}

func extractInt64Value(value interface{}) (int64, bool) {
	switch v := value.(type) {
	case float64:
		return int64(v), true
	case int64:
		return v, true
	case int:
		return int64(v), true
	case json.Number:
		parsed, err := v.Int64()
		if err == nil {
			return parsed, true
		}
	case string:
		parsed, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		if err == nil {
			return parsed, true
		}
	}
	return 0, false
}

func extractIntValueOrDefault(value interface{}, defaultValue int) int {
	if parsed, ok := extractInt64Value(value); ok {
		return int(parsed)
	}
	return defaultValue
}

func stringPtrIfNotEmpty(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func stringPtrAllowEmpty(value string) *string {
	return &value
}

func int64PtrIfPositive(value int64) *int64 {
	if value <= 0 {
		return nil
	}
	return &value
}

func replaceIDs(value string, idMap map[int64]int64) string {
	type idReplacement struct {
		old string
		new string
	}

	replacements := make([]idReplacement, 0, len(idMap))
	for oldID, newID := range idMap {
		replacements = append(replacements, idReplacement{
			old: strconv.FormatInt(oldID, 10),
			new: strconv.FormatInt(newID, 10),
		})
	}

	sort.Slice(replacements, func(i, j int) bool {
		return len(replacements[i].old) > len(replacements[j].old)
	})

	for _, replacement := range replacements {
		re := regexp.MustCompile(`(^|[^0-9])` + regexp.QuoteMeta(replacement.old) + `($|[^0-9])`)
		value = re.ReplaceAllString(value, `${1}`+replacement.new+`${2}`)
	}
	return value
}

func parseDynamicData(raw string) (map[string]string, error) {
	result := make(map[string]string)

	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &rawMap); err != nil {
		return nil, err
	}

	for key, rawVal := range rawMap {
		valStr := string(rawVal)
		if strings.HasPrefix(strings.TrimSpace(valStr), "\"") {
			var unquoted string
			if err := json.Unmarshal(rawVal, &unquoted); err == nil {
				valStr = unquoted
			}
		}
		result[key] = valStr
	}

	return result, nil
}
