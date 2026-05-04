package datafilling

import datasourcedomain "dataease/backend/internal/domain/datasource"

// DataFillingForm is the GORM model for data_filling_forms table.
type DataFillingForm struct {
	ID                int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name              string `gorm:"column:name;type:varchar(255)" json:"name"`
	PID               int64  `gorm:"column:pid" json:"pid"`
	Level             int    `gorm:"column:level" json:"level"`
	NodeType          string `gorm:"column:node_type;type:varchar(50)" json:"nodeType"`
	PhysicalTableName string `gorm:"column:table_name;type:varchar(255)" json:"tableName"`
	DatasourceID      int64  `gorm:"column:datasource_id" json:"datasourceId"`
	Forms             string `gorm:"column:forms;type:longtext" json:"forms"`
	CreateIndex       bool   `gorm:"column:create_index" json:"createIndex"`
	TableIndexes      string `gorm:"column:table_indexes;type:longtext" json:"tableIndexes"`
	CreateBy          int64  `gorm:"column:create_by" json:"createBy"`
	CreateTime        int64  `gorm:"column:create_time" json:"createTime"`
	UpdateBy          int64  `gorm:"column:update_by" json:"updateBy"`
	UpdateTime        int64  `gorm:"column:update_time" json:"updateTime"`
	Creator           string `gorm:"-" json:"creator,omitempty"`
	Updater           string `gorm:"-" json:"updater,omitempty"`
	DatasourceName    string `gorm:"-" json:"datasourceName,omitempty"`
	UseExistsTable    bool   `gorm:"-" json:"useExistsTable,omitempty"`
}

func (DataFillingForm) TableName() string { return "data_filling_forms" }

// DfCommitLog is the GORM model for df_commit_log table.
type DfCommitLog struct {
	ID         int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	FormID     int64  `gorm:"column:form_id;index" json:"formId"`
	DataID     string `gorm:"column:data_id;type:varchar(64)" json:"dataId"`
	Operate    int    `gorm:"column:operate" json:"operate"`
	CommitBy   int64  `gorm:"column:commit_by" json:"commitBy"`
	Committer  string `gorm:"column:committer;type:varchar(255)" json:"committer"`
	CommitTime int64  `gorm:"column:commit_time" json:"commitTime"`
	Count      int    `gorm:"column:count" json:"count"`
}

func (DfCommitLog) TableName() string { return "df_commit_log" }

const (
	NodeTypeFolder = "folder"
	NodeTypeForm   = "form"
)

type BaseType string

const (
	BaseTypeNvarchar BaseType = "nvarchar"
	BaseTypeText     BaseType = "text"
	BaseTypeNumber   BaseType = "number"
	BaseTypeDecimal  BaseType = "decimal"
	BaseTypeDatetime BaseType = "datetime"
)

type ExtTableField struct {
	Type     string               `json:"type"`
	TypeName string               `json:"typeName"`
	Icon     string               `json:"icon"`
	ID       string               `json:"id"`
	Settings ExtTableFieldSetting `json:"settings"`
	Removed  bool                 `json:"removed,omitempty"`
}

type ExtTableFieldSetting struct {
	Name             string               `json:"name"`
	Required         bool                 `json:"required"`
	Mapping          ExtTableFieldMapping `json:"mapping"`
	Unique           bool                 `json:"unique"`
	InputType        string               `json:"inputType"`
	Placeholder      string               `json:"placeholder,omitempty"`
	OptionSourceType int                  `json:"optionSourceType,omitempty"`
	OptionDatasource int64                `json:"optionDatasource,omitempty"`
	OptionTable      string               `json:"optionTable,omitempty"`
	OptionColumn     string               `json:"optionColumn,omitempty"`
	OptionOrder      string               `json:"optionOrder,omitempty"`
	Multiple         bool                 `json:"multiple,omitempty"`
	Options          []FieldOption        `json:"options,omitempty"`
}

type ExtTableFieldMapping struct {
	ColumnName string   `json:"columnName"`
	Type       BaseType `json:"type"`
	Size       int      `json:"size,omitempty"`
	Accuracy   int      `json:"accuracy,omitempty"`
}

type FieldOption struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ExtIndexField struct {
	ID      string           `json:"id"`
	Name    string           `json:"name"`
	Columns []ExtIndexColumn `json:"columns"`
}

type ExtIndexColumn struct {
	ColumnName string `json:"columnName"`
	Order      string `json:"order"`
}

type CreateFormRequest struct {
	Name           string `json:"name"`
	PID            int64  `json:"pid"`
	NodeType       string `json:"nodeType"`
	TableName      string `json:"tableName,omitempty"`
	DatasourceID   int64  `json:"datasourceId,omitempty"`
	Forms          string `json:"forms,omitempty"`
	CreateIndex    bool   `json:"createIndex,omitempty"`
	TableIndexes   string `json:"tableIndexes,omitempty"`
	UseExistsTable bool   `json:"useExistsTable,omitempty"`
}

type UpdateFormRequest struct {
	ID int64 `json:"id"`
	CreateFormRequest
}

type TableDataRequest struct {
	ID           int64         `json:"id"`
	CurrentPage  int64         `json:"currentPage"`
	PageSize     int64         `json:"pageSize"`
	SearchParams []SearchParam `json:"searchParams"`
}

type SearchParam struct {
	Term     string        `json:"term"`
	Field    string        `json:"field"`
	Value    interface{}   `json:"value"`
	Values   []interface{} `json:"values"`
	Multiple bool          `json:"multiple"`
}

type TableDataResponse struct {
	Data        []map[string]interface{} `json:"data"`
	Fields      string                   `json:"fields"`
	Total       int64                    `json:"total"`
	CurrentPage int64                    `json:"currentPage"`
	PageSize    int64                    `json:"pageSize"`
	Key         string                   `json:"key"`
}

type BatchDeleteRowDataRequest struct {
	IDs []string `json:"ids"`
}

type ListColumnDataRequest struct {
	ColumnName string `json:"columnName"`
}

type CommitLogPageRequest struct {
	FormID  int64 `json:"formId"`
	Operate *int  `json:"operate,omitempty"`
}

type ClearCommitLogRequest struct {
	FormID    int64  `json:"formId"`
	ClearType string `json:"clearType,omitempty"`
}

type RenameRequest struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type MoveRequest struct {
	ID  int64 `json:"id"`
	PID int64 `json:"pid"`
}

type TreeRequest struct {
	Keyword *string `json:"keyword"`
}

type TreeNode struct {
	ID       int64       `json:"id"`
	Name     string      `json:"name"`
	PID      int64       `json:"pid"`
	NodeType string      `json:"nodeType"`
	Children []*TreeNode `json:"children,omitempty"`
}

type TreeResponse = []*TreeNode

type DatasourceSummary struct {
	ID             int64  `json:"id"`
	PID            int64  `json:"pid"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	Status         string `json:"status,omitempty"`
	EnableDataFill bool   `json:"enableDataFill"`
}

type BuiltInTable = datasourcedomain.TableInfo
