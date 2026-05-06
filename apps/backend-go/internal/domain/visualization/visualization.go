package visualization

import (
	"encoding/json"
	"reflect"
	"strconv"
)

// Node type constants for visualization nodes.
const (
	NodeTypePanel  = "panel"
	NodeTypeFolder = "folder"
	NodeTypeLeaf   = "leaf"
)

// Visualization type constants.
const (
	TypeDashboard = "dashboard"
	TypeDataV     = "dataV"
)

// Workbranch resource type aliases (used in findRecent queries).
const (
	ResourceAliasPanel      = "panel"
	ResourceAliasScreen     = "screen"
	ResourceAliasDataset    = "dataset"
	ResourceAliasDatasource = "datasource"
)

// FlexInt decodes from JSON number or quoted string (frontend sends id as "2").
type FlexInt int64

func (f *FlexInt) UnmarshalJSON(data []byte) error {
	var n json.Number
	if err := json.Unmarshal(data, &n); err == nil {
		if i, err := n.Int64(); err == nil {
			*f = FlexInt(i)
			return nil
		}
		if parsed, err := strconv.ParseInt(n.String(), 10, 64); err == nil {
			*f = FlexInt(parsed)
			return nil
		}
	}
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		if parsed, err := strconv.ParseInt(s, 10, 64); err == nil {
			*f = FlexInt(parsed)
			return nil
		}
	}
	return &json.UnmarshalTypeError{Value: "flexint", Type: reflect.TypeOf(int64(0))}
}

func (f FlexInt) Int64() int64 { return int64(f) }

type DataVisualizationInfo struct {
	ID              int64   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name            string  `gorm:"column:name" json:"name"`
	PID             *int64  `gorm:"column:pid" json:"pid"`
	OrgID           *int64  `gorm:"column:org_id" json:"orgId"`
	Level           *int    `gorm:"column:level" json:"level"`
	NodeType        *string `gorm:"column:node_type" json:"nodeType"`
	Type            *string `gorm:"column:type" json:"type"`
	CanvasStyleData *string `gorm:"column:canvas_style_data" json:"canvasStyleData"`
	ComponentData   *string `gorm:"column:component_data" json:"componentData"`
	MobileLayout    *bool   `gorm:"column:mobile_layout" json:"mobileLayout"`
	Status          *int    `gorm:"column:status" json:"status"`
	Sort            *int    `gorm:"column:sort" json:"sort"`
	CreateTime      *int64  `gorm:"column:create_time" json:"createTime"`
	CreateBy        *string `gorm:"column:create_by" json:"createBy"`
	UpdateTime      *int64  `gorm:"column:update_time" json:"updateTime"`
	UpdateBy        *string `gorm:"column:update_by" json:"updateBy"`
	DeleteFlag      *bool   `gorm:"column:delete_flag" json:"deleteFlag"`
	DeleteTime      *int64  `gorm:"column:delete_time" json:"deleteTime"`
	DeleteBy        *string `gorm:"column:delete_by" json:"deleteBy"`
	Version         *int    `gorm:"column:version" json:"version"`
	ContentID       *string `gorm:"column:content_id" json:"contentId"`
	CheckVersion    *string `gorm:"column:check_version" json:"checkVersion"`
}

func (DataVisualizationInfo) TableName() string {
	return "data_visualization_info"
}

type SaveRequest struct {
	Name              string                            `json:"name" binding:"required"`
	PID               *int64                            `json:"pid"`
	Type              *string                           `json:"type"`
	NodeType          *string                           `json:"nodeType"`
	CanvasStyleData   *string                           `json:"canvasStyleData"`
	ComponentData     *string                           `json:"componentData"`
	MobileLayout      *bool                             `json:"mobileLayout"`
	ContentID         *string                           `json:"contentId"`
	CheckVersion      *string                           `json:"checkVersion"`
	CanvasViewInfo    map[string]map[string]interface{} `json:"canvasViewInfo"`
	AppData           string                            `json:"appData"`
	DataType          *string                           `json:"dataType"`
	DatasetFolderPID  *int64                            `json:"datasetFolderPid"`
	DatasetFolderName *string                           `json:"datasetFolderName"`
}

type CopyRequest struct {
	ID           int64   `json:"id" binding:"required"`
	Name         string  `json:"name" binding:"required"`
	PID          *int64  `json:"pid"`
	Type         *string `json:"type"`
	NodeType     *string `json:"nodeType"`
	MobileLayout *bool   `json:"mobileLayout"`
}

type UpdateRequest struct {
	ID              int64                             `json:"id" binding:"required"`
	Name            *string                           `json:"name"`
	PID             *int64                            `json:"pid"`
	Type            *string                           `json:"type"`
	CanvasStyleData *string                           `json:"canvasStyleData"`
	ComponentData   *string                           `json:"componentData"`
	MobileLayout    *bool                             `json:"mobileLayout"`
	Status          *int                              `json:"status"`
	ContentID       *string                           `json:"contentId"`
	CheckVersion    *string                           `json:"checkVersion"`
	CanvasViewInfo  map[string]map[string]interface{} `json:"canvasViewInfo"`
}

type NameCheckRequest struct {
	ID   int64   `json:"id"`
	Name string  `json:"name" binding:"required"`
	PID  *int64  `json:"pid"`
	Opt  *string `json:"opt"`
	Type *string `json:"type"`
}

type CanvasChangeRequest struct {
	ID           int64   `json:"id" binding:"required"`
	ContentID    *string `json:"contentId"`
	CheckVersion *string `json:"checkVersion"`
}

type MoveRequest struct {
	ID  int64  `json:"id" binding:"required"`
	PID *int64 `json:"pid"`
}

type DetailRequest struct {
	ID FlexInt `json:"id" binding:"required"`
}

type ListRequest struct {
	Keyword *string `json:"keyword"`
	Type    *string `json:"type"`
	Current int     `json:"current"`
	Size    int     `json:"size"`
}

type ListResponse struct {
	List    []*DataVisualizationInfo `json:"list"`
	Total   int64                    `json:"total"`
	Current int                      `json:"current"`
	Size    int                      `json:"size"`
}

// DecompressionRequest represents the template import request sent by the frontend.
type DecompressionRequest struct {
	NewFrom         string `json:"newFrom"`         // "new_inner_template" | "new_outer_template" | "new_market_template" | "localFile"
	TemplateID      *int64 `json:"templateId"`      // required for new_inner_template
	ResourceName    string `json:"resourceName"`    // used by new_market_template
	TemplateURL     string `json:"templateUrl"`     // used by new_market_template
	Name            string `json:"name"`            // used by new_outer_template
	Type            string `json:"type"`            // visualization type (dashboard/dataV)
	Version         int    `json:"version"`         // used by local template file import
	CanvasStyleData string `json:"canvasStyleData"` // used by new_outer_template
	ComponentData   string `json:"componentData"`   // used by new_outer_template
	DynamicData     string `json:"dynamicData"`     // used by new_outer_template
	AppData         string `json:"appData"`         // optional app import payload
	StaticResource  string `json:"staticResource"`  // static resource info
}

// Export2AppCheckRequest is the request body for POST /dataVisualization/export2AppCheck.
type Export2AppCheckRequest struct {
	DvID    int64   `json:"dvId"`
	ViewIDs []int64 `json:"viewIds"`
	DsIDs   []int64 `json:"dsIds"`
}

type AppCanvasNameCheckRequest struct {
	DatasetFolderPid  *int64 `json:"datasetFolderPid"`
	DatasetFolderName string `json:"datasetFolderName"`
}

type ExportLogRequest struct {
	ID   *int64 `json:"id"`
	Type string `json:"type"`
}

// Export2AppCheckResponse mirrors Java VisualizationExport2AppVO for non-market template app export.
type Export2AppCheckResponse struct {
	CheckStatus            bool                     `json:"checkStatus"`
	CheckMes               string                   `json:"checkMes"`
	ChartViewsInfo         []map[string]interface{} `json:"chartViewsInfo"`
	DatasetGroupsInfo      []map[string]interface{} `json:"datasetGroupsInfo"`
	DatasetTablesInfo      []map[string]interface{} `json:"datasetTablesInfo"`
	DatasetTableFieldsInfo []map[string]interface{} `json:"datasetTableFieldsInfo"`
	DatasourceInfo         []map[string]interface{} `json:"datasourceInfo"`
	DatasourceTaskInfo     []map[string]interface{} `json:"datasourceTaskInfo"`
	LinkJumps              []map[string]interface{} `json:"linkJumps"`
	LinkJumpInfos          []map[string]interface{} `json:"linkJumpInfos"`
	LinkJumpTargetInfos    []map[string]interface{} `json:"linkJumpTargetInfos"`
	Linkages               []map[string]interface{} `json:"linkages"`
	LinkageFields          []map[string]interface{} `json:"linkageFields"`
}

// DecompressionResponse is the Java-compatible response consumed by frontend canvasUtils.ts.
type DecompressionResponse struct {
	ID              string                            `json:"id"` // new visualization ID as string
	Name            string                            `json:"name"`
	Type            string                            `json:"type"` // "dashboard" or "dataV"
	Version         int                               `json:"version"`
	CanvasStyleData string                            `json:"canvasStyleData"` // JSON string (frontend does JSON.parse)
	ComponentData   string                            `json:"componentData"`   // JSON string (frontend does JSON.parse)
	AppData         string                            `json:"appData"`         // JSON string or empty
	CanvasViewInfo  map[string]map[string]interface{} `json:"canvasViewInfo"`  // keyed by new view ID string
}

// WorkbranchQueryRequest is the request body for /dataVisualization/findRecent
type WorkbranchQueryRequest struct {
	Type      string `json:"type"`      // "panel" | "screen" | "dataset" | "datasource"
	Keyword   string `json:"keyword"`   // optional search filter
	QueryFrom string `json:"queryFrom"` // server sets to "recent"
	Asc       bool   `json:"asc"`       // sort direction (false = DESC)
}

// VisualizationResourceVO is a resource item for the workbranch recent list
type VisualizationResourceVO struct {
	ID           string `json:"id"`
	ResourceID   string `json:"resourceId"`
	Name         string `json:"name"`
	Type         string `json:"type"` // "panel" | "screen" | "dataset" | "datasource"
	Creator      string `json:"creator"`
	LastEditor   string `json:"lastEditor"`
	LastEditTime int64  `json:"lastEditTime"`
	Favorite     bool   `json:"favorite"`
	Weight       int    `json:"weight"`
	ExtFlag      int    `json:"extFlag"`
}
