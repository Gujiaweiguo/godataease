package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"dataease/backend/internal/domain/audit"
	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/domain/visualization"
	"dataease/backend/internal/repository"

	"github.com/google/uuid"
)

type VisualizationService struct {
	repo                   *repository.VisualizationRepository
	datasetRepo            *repository.DatasetRepository
	resourcePermService    *ResourcePermissionService
	templateService        *TemplateService
	templateExtendDataRepo *repository.TemplateExtendDataRepository
	auditService           *AuditService
}

const (
	visualizationNodeTypePanel      = "panel"
	visualizationNodeTypeFolder     = "folder"
	visualizationTypeDashboard      = "dashboard"
	visualizationTypeDataV          = "dataV"
	visualizationBusiFlagDashboardV = "dashboard-dataV"
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

func (s *VisualizationService) SetTemplateExtendDataRepo(r *repository.TemplateExtendDataRepository) {
	s.templateExtendDataRepo = r
}

func (s *VisualizationService) SetDatasetRepository(r *repository.DatasetRepository) {
	s.datasetRepo = r
}

func (s *VisualizationService) SetAuditService(auditSvc *AuditService) {
	s.auditService = auditSvc
}

func (s *VisualizationService) Save(req *visualization.SaveRequest, updateBy string) (int64, error) {
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
	if err := s.applyInheritedPermissionsOnCreate(v.ID, req.Name, req.PID, req.Type); err != nil {
		_ = s.repo.DeleteLogic(v.ID, updateBy)
		return 0, err
	}
	return v.ID, nil
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
		return 0, err
	}

	nodeType := source.NodeType
	if req.NodeType != nil && *req.NodeType != "" {
		nodeType = req.NodeType
	}
	typ := source.Type
	if req.Type != nil && *req.Type != "" {
		typ = req.Type
	}
	mobileLayout := source.MobileLayout
	if req.MobileLayout != nil {
		mobileLayout = req.MobileLayout
	}

	return s.Save(&visualization.SaveRequest{
		Name:            req.Name,
		PID:             req.PID,
		Type:            typ,
		NodeType:        nodeType,
		CanvasStyleData: source.CanvasStyleData,
		ComponentData:   source.ComponentData,
		MobileLayout:    mobileLayout,
		ContentID:       source.ContentID,
		CheckVersion:    source.CheckVersion,
	}, updateBy)
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

	return s.repo.Update(v)
}

func (s *VisualizationService) Detail(req *visualization.DetailRequest) (*visualization.DataVisualizationInfo, error) {
	return s.repo.GetByID(req.ID)
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

func (s *VisualizationService) InteractiveTree(busiFlag string) ([]*visualization.DataVisualizationInfo, error) {
	types, err := resolveInteractiveVisualizationTypes(busiFlag)
	if err != nil {
		return nil, err
	}
	return s.repo.ListAllByTypes(types)
}

func (s *VisualizationService) DeleteLogic(id int64, updateBy string) error {
	return s.repo.DeleteLogic(id, updateBy)
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
		return "success", nil
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
	return "success", nil
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

func (s *VisualizationService) ViewDetailList(dvID int64) ([]map[string]interface{}, error) {
	return s.repo.GetChartViewsBySceneID(dvID)
}

func (s *VisualizationService) AppCanvasNameCheck(req *visualization.AppCanvasNameCheckRequest) (string, error) {
	if req == nil {
		return "success", nil
	}
	if s.datasetRepo == nil {
		return "success", nil
	}
	name := strings.TrimSpace(req.DatasetFolderName)
	if name == "" {
		return "success", nil
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
	return "success", nil
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

	ensureEmpty := func(m []map[string]interface{}) []map[string]interface{} {
		if m == nil {
			return []map[string]interface{}{}
		}
		return stringifyExportIDs(m)
	}

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

	if len(datasources) == 0 {
		return nil, fmt.Errorf("当前不存在数据源无法导出")
	}

	for _, ds := range datasources {
		if dsType, ok := ds["type"]; ok {
			typeStr := fmt.Sprintf("%v", dsType)
			if strings.Contains(strings.ToUpper(typeStr), "API") {
				return nil, fmt.Errorf("包含API数据源不支持导出")
			}
		}
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

	return &visualization.Export2AppCheckResponse{
		CheckStatus:            true,
		CheckMes:               "success",
		ChartViewsInfo:         ensureEmpty(chartViews),
		DatasetGroupsInfo:      ensureEmpty(datasetGroups),
		DatasetTablesInfo:      ensureEmpty(datasetTables),
		DatasetTableFieldsInfo: ensureEmpty(datasetTableFields),
		DatasourceInfo:         ensureEmpty(datasources),
		DatasourceTaskInfo:     ensureEmpty(datasourceTasks),
		LinkJumps:              ensureEmpty(linkJumps),
		LinkJumpInfos:          ensureEmpty(linkJumpInfos),
		LinkJumpTargetInfos:    ensureEmpty(linkJumpTargets),
		Linkages:               ensureEmpty(linkages),
		LinkageFields:          ensureEmpty(linkageFields),
	}, nil
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
)

func (s *VisualizationService) Decompression(req *visualization.DecompressionRequest) (*visualization.DecompressionResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("decompression request is required")
	}
	newDvID := int64(uuid.New().ID())

	var templateStyle, templateData, dynamicData, name, dvType, appDataStr string
	var version int

	switch req.NewFrom {
	case newFromInnerTemplate:
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
		templateStyle = tmpl.TemplateStyle
		templateData = tmpl.TemplateData
		dynamicData = tmpl.DynamicData
		name = tmpl.Name
		dvType = tmpl.DvType
		version = tmpl.Version
		appDataStr = tmpl.AppData
		_ = s.templateService.IncrementUseCount(tmpl.ID)

	case newFromOuterTemplate:
		templateStyle = req.CanvasStyleData
		templateData = req.ComponentData
		dynamicData = req.DynamicData
		appDataStr = req.AppData
		name = req.Name
		dvType = req.Type
		version = 3

	case newFromMarketTemplate:
		return nil, fmt.Errorf("new_market_template is not yet supported: templateUrl=%s", req.TemplateURL)

	default:
		return nil, fmt.Errorf("unsupported newFrom: %s", req.NewFrom)
	}

	appDataStr = processAppData(appDataStr, newDvID)

	hasAppData := strings.TrimSpace(appDataStr) != ""
	canvasViewInfo, err := s.processDynamicData(dynamicData, newDvID, &templateData, &appDataStr, hasAppData)
	if err != nil {
		return nil, err
	}

	return &visualization.DecompressionResponse{
		ID:              fmt.Sprintf("%d", newDvID),
		Name:            name,
		Type:            dvType,
		Version:         version,
		CanvasStyleData: templateStyle,
		ComponentData:   templateData,
		AppData:         appDataStr,
		CanvasViewInfo:  canvasViewInfo,
	}, nil
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
	}
	return 0, false
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
