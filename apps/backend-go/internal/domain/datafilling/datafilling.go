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

// DataFillingTask is the GORM model for data_filling_task table.
type DataFillingTask struct {
	ID                   int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	FormID               int64  `gorm:"column:form_id;index" json:"formId"`
	Name                 string `gorm:"column:name;type:varchar(255)" json:"name"`
	ReciFlagList         string `gorm:"column:reci_flag_list;type:longtext" json:"reciFlagList"`
	UIDList              string `gorm:"column:uid_list;type:longtext" json:"uidList"`
	RIDList              string `gorm:"column:rid_list;type:longtext" json:"ridList"`
	FillType             int    `gorm:"column:fill_type" json:"fillType"`
	FitType              int    `gorm:"column:fit_type" json:"fitType"`
	FitColumn            string `gorm:"column:fit_column;type:varchar(255)" json:"fitColumn"`
	RateType             int    `gorm:"column:rate_type" json:"rateType"`
	RateVal              string `gorm:"column:rate_val;type:varchar(255)" json:"rateVal"`
	OneTimeType          int    `gorm:"column:one_time_type" json:"oneTimeType"`
	StartTime            int64  `gorm:"column:start_time" json:"startTime"`
	EndTime              int64  `gorm:"column:end_time" json:"endTime"`
	PublishRangeTime     int    `gorm:"column:publish_range_time" json:"publishRangeTime"`
	PublishRangeTimeType int    `gorm:"column:publish_range_time_type" json:"publishRangeTimeType"`
	Status               int    `gorm:"column:status" json:"status"`
	LastExecStatus       int    `gorm:"column:last_exec_status" json:"lastExecStatus"`
	LastExecTime         int64  `gorm:"column:last_exec_time" json:"lastExecTime"`
	NextExecTime         int64  `gorm:"column:next_exec_time" json:"nextExecTime"`
	CreateBy             int64  `gorm:"column:create_by" json:"createBy"`
	CreateTime           int64  `gorm:"column:create_time" json:"createTime"`
	UpdateBy             int64  `gorm:"column:update_by" json:"updateBy"`
	UpdateTime           int64  `gorm:"column:update_time" json:"updateTime"`
	FormExtSetting       string `gorm:"column:form_ext_setting;type:longtext" json:"formExtSetting"`
	FormFilterSetting    string `gorm:"column:form_filter_setting;type:longtext" json:"formFilterSetting"`
}

func (DataFillingTask) TableName() string { return "data_filling_task" }

// DataFillingSubTask is the GORM model for data_filling_sub_task table.
type DataFillingSubTask struct {
	ID                  int64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TaskID              int64 `gorm:"column:task_id;index" json:"taskId"`
	StartTime           int64 `gorm:"column:start_time" json:"startTime"`
	EndTime             int64 `gorm:"column:end_time" json:"endTime"`
	ExecStatus          int   `gorm:"column:exec_status" json:"execStatus"`
	Status              int   `gorm:"column:status" json:"status"`
	TotalCount          int   `gorm:"column:total_count" json:"totalCount"`
	UnfinishedCount     int   `gorm:"column:unfinished_count" json:"unfinishedCount"`
	TotalUserCount      int   `gorm:"column:total_user_count" json:"totalUserCount"`
	UnfinishedUserCount int   `gorm:"column:unfinished_user_count" json:"unfinishedUserCount"`
	FillType            int   `gorm:"column:fill_type" json:"fillType"`
}

func (DataFillingSubTask) TableName() string { return "data_filling_sub_task" }

// DataFillingSubInstance is the GORM model for data_filling_sub_instance table.
type DataFillingSubInstance struct {
	ID         int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TaskID     int64  `gorm:"column:task_id;index:idx_task_pid" json:"taskId"`
	PID        int64  `gorm:"column:pid;index:idx_task_pid" json:"pid"`
	UID        int64  `gorm:"column:uid;index" json:"uid"`
	FormID     int64  `gorm:"column:form_id" json:"formId"`
	DataID     string `gorm:"column:data_id;type:varchar(64)" json:"dataId"`
	FinishTime int64  `gorm:"column:finish_time" json:"finishTime"`
	Status     int    `gorm:"column:status" json:"status"`
}

func (DataFillingSubInstance) TableName() string { return "data_filling_sub_instance" }

const (
	NodeTypeFolder = "folder"
	NodeTypeForm   = "form"
)

const (
	TaskStatusStopped = 0
	TaskStatusStarted = 1

	SubTaskStatusExpired = 0
	SubTaskStatusActive  = 1

	SubInstanceStatusOpen     = 0
	SubInstanceStatusFinished = 1
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

type TaskSaveRequest struct {
	ID                   *int64  `json:"id,omitempty"`
	FormID               int64   `json:"formId"`
	Name                 string  `json:"name"`
	ReciFlagList         []int   `json:"reciFlagList"`
	UIDList              []int64 `json:"uidList"`
	RIDList              []int64 `json:"ridList"`
	FillType             int     `json:"fillType"`
	FitType              int     `json:"fitType"`
	FitColumn            string  `json:"fitColumn"`
	RateType             int     `json:"rateType"`
	RateVal              string  `json:"rateVal"`
	OneTimeType          int     `json:"oneTimeType"`
	StartTime            int64   `json:"startTime"`
	EndTime              int64   `json:"endTime"`
	PublishRangeTime     int     `json:"publishRangeTime"`
	PublishRangeTimeType int     `json:"publishRangeTimeType"`
	FormExtSetting       string  `json:"formExtSetting"`
	FormFilterSetting    string  `json:"formFilterSetting"`
}

type TaskInfoVO struct {
	ID                   int64   `json:"id"`
	FormID               int64   `json:"formId"`
	Name                 string  `json:"name"`
	ReciFlagList         []int   `json:"reciFlagList"`
	UIDList              []int64 `json:"uidList"`
	RIDList              []int64 `json:"ridList"`
	FillType             int     `json:"fillType"`
	FitType              int     `json:"fitType"`
	FitColumn            string  `json:"fitColumn"`
	RateType             int     `json:"rateType"`
	RateVal              string  `json:"rateVal"`
	OneTimeType          int     `json:"oneTimeType"`
	StartTime            int64   `json:"startTime"`
	EndTime              int64   `json:"endTime"`
	PublishRangeTime     int     `json:"publishRangeTime"`
	PublishRangeTimeType int     `json:"publishRangeTimeType"`
	Status               int     `json:"status"`
	LastExecStatus       int     `json:"lastExecStatus"`
	LastExecTime         int64   `json:"lastExecTime"`
	NextExecTime         int64   `json:"nextExecTime"`
	CreateBy             int64   `json:"createBy"`
	CreateTime           int64   `json:"createTime"`
	UpdateBy             int64   `json:"updateBy"`
	UpdateTime           int64   `json:"updateTime"`
	FormExtSetting       string  `json:"formExtSetting"`
	FormFilterSetting    string  `json:"formFilterSetting"`
}

type TaskPageRequest struct {
	TaskID  *int64  `json:"taskId,omitempty"`
	Keyword *string `json:"keyword,omitempty"`
}

type TaskPageResponse struct {
	Records []*TaskInfoVO `json:"records"`
	Total   int64         `json:"total"`
	Current int           `json:"current"`
	Size    int           `json:"size"`
}

type BatchDeleteTaskRequest struct {
	IDs []int64 `json:"ids"`
}

type BatchDeleteSubTaskRequest struct {
	IDs []int64 `json:"ids"`
}

type SubTaskPageRequest struct {
	TaskID int64 `json:"taskId"`
}

type SubTaskPageResponse struct {
	Records []*DataFillingSubTask `json:"records"`
	Total   int64                 `json:"total"`
	Current int                   `json:"current"`
	Size    int                   `json:"size"`
}

type SubTaskUsersRequest struct {
	TaskID int64 `json:"taskId"`
}

type SubTaskUserItem struct {
	ID         int64  `json:"id"`
	TaskID     int64  `json:"taskId"`
	PID        int64  `json:"pid"`
	UID        int64  `json:"uid"`
	FormID     int64  `json:"formId"`
	DataID     string `json:"dataId"`
	FinishTime int64  `json:"finishTime"`
	Status     int    `json:"status"`
}

type ExecuteNowRequest struct {
	TaskID int64 `json:"taskId"`
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
