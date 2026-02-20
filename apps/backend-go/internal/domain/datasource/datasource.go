package datasource

const (
	StatusSuccess = "Success"
	StatusError   = "Error"

	TypeFolder = "folder"
	TypeExcel  = "Excel"
)

type CoreDatasource struct {
	ID             int64   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PID            *int64  `gorm:"column:pid" json:"pid"`
	Name           string  `gorm:"column:name" json:"name"`
	Description    *string `gorm:"column:description" json:"description"`
	Type           string  `gorm:"column:type" json:"type"`
	EditType       *string `gorm:"column:edit_type" json:"editType"`
	Configuration  *string `gorm:"column:configuration" json:"configuration"`
	Status         *string `gorm:"column:status" json:"status"`
	QrtzInstance   *string `gorm:"column:qrtz_instance" json:"qrtzInstance"`
	TaskStatus     *string `gorm:"column:task_status" json:"taskStatus"`
	EnableDataFill *bool   `gorm:"column:enable_data_fill" json:"enableDataFill"`
	CreateTime     *int64  `gorm:"column:create_time" json:"createTime"`
	UpdateTime     *int64  `gorm:"column:update_time" json:"updateTime"`
	UpdateBy       *int64  `gorm:"column:update_by" json:"updateBy"`
	CreateBy       *string `gorm:"column:create_by" json:"createBy"`
	DelFlag        *int    `gorm:"column:del_flag" json:"delFlag"`
}

func (CoreDatasource) TableName() string {
	return "core_datasource"
}

type ListRequest struct {
	Keyword *string `json:"keyword"`
	Current int     `json:"current"`
	Size    int     `json:"size"`
}

type ListResponse struct {
	List    []*CoreDatasource `json:"list"`
	Total   int64             `json:"total"`
	Current int               `json:"current"`
	Size    int               `json:"size"`
}

type ValidateRequest struct {
	DatasourceID  *int64  `json:"datasourceId"`
	Type          *string `json:"type"`
	Configuration *string `json:"configuration"`
}

type ValidateResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type TableRequest struct {
	DatasourceID int64  `json:"datasourceId"`
	TableName    string `json:"tableName"`
	Limit        int    `json:"limit"`
}

type TableInfo struct {
	ID           int64  `json:"id"`
	DatasourceID int64  `json:"datasourceId"`
	Name         string `json:"name"`
	TableName    string `json:"tableName"`
	Type         string `json:"type"`
	Status       string `json:"status,omitempty"`
	LastUpdate   int64  `json:"lastUpdateTime,omitempty"`
}

type TableField struct {
	OriginName string `json:"originName"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	DeType     int    `json:"deType"`
}

type PreviewDataResponse struct {
	Fields []TableField             `json:"fields"`
	Data   []map[string]interface{} `json:"data"`
	Total  int64                    `json:"total"`
}

type WriteRequest struct {
	ID             int64   `json:"id"`
	PID            *int64  `json:"pid"`
	Name           string  `json:"name"`
	Description    *string `json:"description"`
	Type           string  `json:"type"`
	NodeType       string  `json:"nodeType"`
	EditType       *string `json:"editType"`
	Configuration  *string `json:"configuration"`
	EnableDataFill *bool   `json:"enableDataFill"`
}

type ConnectionConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	JDBCUrl  string `json:"jdbcUrl"`
	Database string `json:"dataBase"`
	Schema   string `json:"schema"`
}
