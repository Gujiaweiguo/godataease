package dataset

const (
	NodeTypeFolder  = "folder"
	NodeTypeDataset = "dataset"
)

type CoreDatasetGroup struct {
	ID             int64   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name           string  `gorm:"column:name" json:"name"`
	PID            *int64  `gorm:"column:pid" json:"pid"`
	Level          *int    `gorm:"column:level" json:"level"`
	NodeType       *string `gorm:"column:node_type" json:"nodeType"`
	Type           *string `gorm:"column:type" json:"type"`
	Info           *string `gorm:"column:info" json:"info"`
	DelFlag        *int    `gorm:"column:del_flag" json:"delFlag"`
	CreateBy       string  `gorm:"column:create_by" json:"createBy"`
	CreateTime     int64   `gorm:"column:create_time" json:"createTime"`
	UpdateBy       string  `gorm:"column:update_by" json:"updateBy"`
	LastUpdateTime int64   `gorm:"column:last_update_time" json:"lastUpdateTime"`
}

func (CoreDatasetGroup) TableName() string {
	return "core_dataset_group"
}

type CoreDatasetTable struct {
	ID             int64   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name           *string `gorm:"column:name" json:"name"`
	DatasourceID   *int64  `gorm:"column:datasource_id" json:"datasourceId"`
	DatasetGroupID int64   `gorm:"column:dataset_group_id" json:"datasetGroupId"`
	PhysicalTable  *string `gorm:"column:table_name" json:"tableName"`
	Type           *string `gorm:"column:type" json:"type"`
	Info           *string `gorm:"column:info" json:"info"`
	SQLVariables   *string `gorm:"column:sql_variable_details" json:"sqlVariableDetails"`
}

func (CoreDatasetTable) TableName() string {
	return "core_dataset_table"
}

type CoreDatasetTableField struct {
	ID             int64   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	DatasourceID   *int64  `gorm:"column:datasource_id" json:"datasourceId"`
	DatasetTableID *int64  `gorm:"column:dataset_table_id" json:"datasetTableId"`
	DatasetGroupID int64   `gorm:"column:dataset_group_id" json:"datasetGroupId"`
	ChartID        *int64  `gorm:"column:chart_id" json:"chartId"`
	OriginName     *string `gorm:"column:origin_name" json:"originName"`
	Name           *string `gorm:"column:name" json:"name"`
	DataeaseName   *string `gorm:"column:dataease_name" json:"dataeaseName"`
	FieldShortName *string `gorm:"column:field_short_name" json:"fieldShortName"`
	GroupType      *string `gorm:"column:group_type" json:"groupType"`
	Type           *string `gorm:"column:type" json:"type"`
	DeType         *int    `gorm:"column:de_type" json:"deType"`
	DeExtractType  *int    `gorm:"column:de_extract_type" json:"deExtractType"`
	ExtField       *int    `gorm:"column:ext_field" json:"extField"`
	Checked        *bool   `gorm:"column:checked" json:"checked"`
	Params         *string `gorm:"column:params" json:"params"`
}

func (CoreDatasetTableField) TableName() string {
	return "core_dataset_table_field"
}

type TreeRequest struct {
	Keyword *string `json:"keyword"`
}

type TreeNode struct {
	ID       int64      `json:"id"`
	Name     string     `json:"name"`
	NodeType string     `json:"nodeType"`
	Children []TreeNode `json:"children,omitempty"`
}

type FieldsRequest struct {
	DatasetGroupID int64 `json:"datasetGroupId" binding:"required"`
}

type PreviewRequest struct {
	DatasetGroupID int64 `json:"datasetGroupId" binding:"required"`
	Limit          int   `json:"limit"`
}

type PreviewResponse struct {
	Columns []string                 `json:"columns"`
	Rows    []map[string]interface{} `json:"rows"`
	Total   int64                    `json:"total"`
}

type ExportDatasetRequest struct {
	ID             int64           `json:"id"`
	ViewName       string          `json:"viewName"`
	Header         []string        `json:"header"`
	Details        [][]interface{} `json:"details"`
	Filename       string          `json:"filename"`
	ExpressionTree string          `json:"expressionTree"`
	Row            int64           `json:"row"`
	DataEaseBi     bool            `json:"dataEaseBi"`
}

type ExportDatasetResponse struct {
	TaskID         string `json:"taskId"`
	Status         string `json:"status"`
	ExportFrom     int64  `json:"exportFrom"`
	ExportFromType string `json:"exportFromType"`
	ExportFromName string `json:"exportFromName"`
}

type WriteRequest struct {
	ID       int64   `json:"id"`
	PID      *int64  `json:"pid"`
	Name     string  `json:"name"`
	NodeType string  `json:"nodeType"`
	Type     *string `json:"type"`
	IsCross  *bool   `json:"isCross"`
}

type SQLPreviewRequest struct {
	DatasourceID       int64  `json:"datasourceId"`
	SQL                string `json:"sql"`
	IsCross            bool   `json:"isCross"`
	SQLVariableDetails string `json:"sqlVariableDetails"`
}

type SQLPreviewField struct {
	OriginName string `json:"originName"`
	DeType     int    `json:"deType"`
}

type SQLPreviewData struct {
	Fields []SQLPreviewField        `json:"fields"`
	Data   []map[string]interface{} `json:"data"`
}

type SQLVariableDetails struct {
	ID              string        `json:"id"`
	VariableName    string        `json:"variableName"`
	Type            []string      `json:"type"`
	Params          []interface{} `json:"params,omitempty"`
	DatasetGroupID  int64         `json:"datasetGroupId"`
	DatasetTableID  int64         `json:"datasetTableId"`
	DatasetFullName string        `json:"datasetFullName"`
	DeType          int           `json:"deType"`
}

type EnumFilter struct {
	FieldID  string        `json:"fieldId"`
	Operator string        `json:"operator"`
	Value    []interface{} `json:"value"`
}

type EnumValueRequest struct {
	QueryID    int64        `json:"queryId"`
	DisplayID  int64        `json:"displayId"`
	SortID     int64        `json:"sortId"`
	Sort       string       `json:"sort"`
	SearchText string       `json:"searchText"`
	Filter     []EnumFilter `json:"filter"`
	ResultMode int          `json:"resultMode"`
}

type MultFieldValuesRequest struct {
	FieldIDs   []int64      `json:"fieldIds"`
	Filter     []EnumFilter `json:"filter"`
	ResultMode int          `json:"resultMode"`
}

type BaseTreeNode struct {
	ID       string         `json:"id"`
	Pid      string         `json:"pid,omitempty"`
	Text     string         `json:"text"`
	NodeType string         `json:"nodeType,omitempty"`
	Children []BaseTreeNode `json:"children,omitempty"`
}

type EnumFilterClause struct {
	Column string
	Values []string
}

type EnumObjectColumn struct {
	Column string
	Alias  string
}

type BarInfo struct {
	ID                int64            `json:"id"`
	Name              string           `json:"name"`
	NodeType          string           `json:"nodeType"`
	CreateBy          string           `json:"createBy"`
	CreateTime        int64            `json:"createTime"`
	Creator           string           `json:"creator"`
	LastUpdateTime    int64            `json:"lastUpdateTime"`
	UpdateBy          string           `json:"updateBy"`
	Updater           string           `json:"updater"`
	DatasourceDTOList []DatasourceInfo `json:"datasourceDTOList"`
	IsCross           bool             `json:"isCross"`
}

type DatasourceInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}
