package syncmodule

type PageResult[T any] struct {
	Records []T   `json:"records"`
	Total   int64 `json:"total"`
	Current int   `json:"current"`
	Size    int   `json:"size"`
}

type SyncDatasourceDTO struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Desc             string `json:"desc"`
	Type             string `json:"type"`
	Configuration    string `json:"configuration"`
	CreateTime       int64  `json:"createTime"`
	UpdateTime       int64  `json:"updateTime"`
	CreateBy         int64  `json:"createBy"`
	CreateByUserName string `json:"createByUserName,omitempty"`
	CreateByName     string `json:"createByName,omitempty"`
	Status           string `json:"status"`
	StatusRemark     string `json:"statusRemark,omitempty"`
}

type SyncDatasourceFieldRequest struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Desc          string `json:"desc"`
	Type          string `json:"type"`
	Configuration string `json:"configuration"`
	Query         string `json:"query"`
	Table         string `json:"table"`
	TableExtract  bool   `json:"tableExtract"`
	TargetDBID    string `json:"targetDbId"`
}

type DBTableDTO struct {
	DatasourceID string `json:"datasourceId"`
	Name         string `json:"name"`
	Remark       string `json:"remark"`
	EnableCheck  bool   `json:"enableCheck"`
	DatasetPath  string `json:"datasetPath"`
}

type TableField struct {
	ID             string `json:"id,omitempty"`
	FieldSource    string `json:"fieldSource"`
	FieldName      string `json:"fieldName"`
	Remarks        string `json:"remarks"`
	FieldType      string `json:"fieldType"`
	FieldSize      int    `json:"fieldSize"`
	FieldPrecision int    `json:"fieldPrecision"`
	FieldPk        bool   `json:"fieldPk"`
	FieldIndex     bool   `json:"fieldIndex"`
}

type SchedulerOption struct {
	Interval int    `json:"interval"`
	Unit     string `json:"unit"`
}

type Source struct {
	Type                string       `json:"type"`
	Query               string       `json:"query"`
	Tables              string       `json:"tables"`
	DatasourceID        string       `json:"datasourceId"`
	TableExtract        string       `json:"tableExtract"`
	DsTableList         []DBTableDTO `json:"dsTableList,omitempty"`
	FieldList           []TableField `json:"fieldList,omitempty"`
	TargetFieldTypeList []string     `json:"targetFieldTypeList,omitempty"`
	IncrementCheckbox   string       `json:"incrementCheckbox,omitempty"`
	IncrementField      string       `json:"incrementField,omitempty"`
	ESQuery             string       `json:"esQuery,omitempty"`
}

type TargetProperty struct {
	PartitionEnable            string `json:"partitionEnable,omitempty"`
	PartitionType              string `json:"partitionType,omitempty"`
	DynamicPartitionEnable     string `json:"dynamicPartitionEnable,omitempty"`
	DynamicPartitionEnd        int64  `json:"dynamicPartitionEnd,omitempty"`
	DynamicPartitionTimeUnit   string `json:"dynamicPartitionTimeUnit,omitempty"`
	ManualPartitionColumnValue string `json:"manualPartitionColumnValue,omitempty"`
	ManualPartitionStart       int64  `json:"manualPartitionStart,omitempty"`
	ManualPartitionEnd         int64  `json:"manualPartitionEnd,omitempty"`
	ManualPartitionInterval    int64  `json:"manualPartitionInterval,omitempty"`
	ManualPartitionTimeRange   string `json:"manualPartitionTimeRange,omitempty"`
	ManualPartitionTimeUnit    string `json:"manualPartitionTimeUnit,omitempty"`
	PartitionColumn            string `json:"partitionColumn,omitempty"`
}

type Target struct {
	CreateTable         string         `json:"createTable,omitempty"`
	Type                string         `json:"type"`
	FieldList           []TableField   `json:"fieldList,omitempty"`
	TableName           string         `json:"tableName"`
	DatasourceID        string         `json:"datasourceId"`
	TargetProperty      string         `json:"targetProperty,omitempty"`
	Property            TargetProperty `json:"property,omitempty"`
	IncrementSync       string         `json:"incrementSync,omitempty"`
	IncrementField      string         `json:"incrementField,omitempty"`
	IncrementFieldType  string         `json:"incrementFieldType,omitempty"`
	Remarks             string         `json:"remarks,omitempty"`
	FaultToleranceRate  float64        `json:"faultToleranceRate,omitempty"`
	IncrementOffset     int64          `json:"incrementOffset,omitempty"`
	IncrementOffsetUnit string         `json:"incrementOffsetUnit,omitempty"`
}

type TaskInfo struct {
	ID                     string          `json:"id"`
	Name                   string          `json:"name"`
	SchedulerType          string          `json:"schedulerType,omitempty"`
	SchedulerConf          string          `json:"schedulerConf,omitempty"`
	SchedulerOption        SchedulerOption `json:"schedulerOption,omitempty"`
	TaskKey                string          `json:"taskKey,omitempty"`
	Desc                   string          `json:"desc,omitempty"`
	ExecutorTimeout        int64           `json:"executorTimeout,omitempty"`
	ExecutorFailRetryCount int64           `json:"executorFailRetryCount,omitempty"`
	Source                 Source          `json:"source,omitempty"`
	Target                 Target          `json:"target,omitempty"`
	Status                 string          `json:"status,omitempty"`
	StartTime              string          `json:"startTime,omitempty"`
	StopTime               string          `json:"stopTime,omitempty"`
	LastExecuteStatus      string          `json:"lastExecuteStatus,omitempty"`
	LogID                  string          `json:"logId,omitempty"`
	ExecutorStartTime      int64           `json:"executorStartTime,omitempty"`
	ExecutorEndTime        int64           `json:"executorEndTime,omitempty"`
	ExecutorMsg            string          `json:"executorMsg,omitempty"`
	LogStatus              string          `json:"logStatus,omitempty"`
}

type TaskLog struct {
	ID                string `json:"id"`
	JobID             string `json:"jobId,omitempty"`
	JobName           string `json:"jobName,omitempty"`
	JobDesc           string `json:"jobDesc,omitempty"`
	ExecutorStartTime int64  `json:"executorStartTime,omitempty"`
	ExecutorEndTime   int64  `json:"executorEndTime,omitempty"`
	Status            string `json:"status,omitempty"`
	ExecutorMsg       string `json:"executorMsg,omitempty"`
	ExecutorAddress   string `json:"executorAddress,omitempty"`
	ClearType         string `json:"clearType,omitempty"`
}

type LogResult struct {
	FromLineNum int    `json:"fromLineNum"`
	ToLineNum   int    `json:"toLineNum"`
	LogContent  string `json:"logContent"`
	IsEnd       bool   `json:"isEnd"`
}

type ResourceCount struct {
	JobCount        int64 `json:"jobCount"`
	DatasourceCount int64 `json:"datasourceCount"`
	JobLogCount     int64 `json:"jobLogCount"`
}

type ChartPoint struct {
	Day   string `json:"day"`
	Count int64  `json:"count"`
}

type TaskGridRequest struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type TaskLogGridRequest struct {
	JobID string `json:"jobId,omitempty"`
}

type TaskPersistedData struct {
	TaskKey                string          `json:"taskKey,omitempty"`
	Desc                   string          `json:"desc,omitempty"`
	ExecutorTimeout        int64           `json:"executorTimeout,omitempty"`
	ExecutorFailRetryCount int64           `json:"executorFailRetryCount,omitempty"`
	SchedulerType          string          `json:"schedulerType,omitempty"`
	SchedulerConf          string          `json:"schedulerConf,omitempty"`
	SchedulerOption        SchedulerOption `json:"schedulerOption,omitempty"`
	Source                 Source          `json:"source,omitempty"`
	Target                 Target          `json:"target,omitempty"`
	StartTime              string          `json:"startTime,omitempty"`
	StopTime               string          `json:"stopTime,omitempty"`
	LastTaskID             string          `json:"lastTaskId,omitempty"`
}
