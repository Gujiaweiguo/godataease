package chart

type CoreChartView struct {
	ID           int64   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Title        *string `gorm:"column:title" json:"title"`
	SceneID      *int64  `gorm:"column:scene_id" json:"sceneId"`
	TableID      *int64  `gorm:"column:table_id" json:"tableId"`
	Type         *string `gorm:"column:type" json:"type"`
	Render       *string `gorm:"column:render" json:"render"`
	ResultCount  *int    `gorm:"column:result_count" json:"resultCount"`
	ResultMode   *string `gorm:"column:result_mode" json:"resultMode"`
	XAxis        *string `gorm:"column:x_axis" json:"xAxis"`
	YAxis        *string `gorm:"column:y_axis" json:"yAxis"`
	CustomAttr   *string `gorm:"column:custom_attr" json:"customAttr"`
	CustomStyle  *string `gorm:"column:custom_style" json:"customStyle"`
	CustomFilter *string `gorm:"column:custom_filter" json:"customFilter"`
	CreateBy     *string `gorm:"column:create_by" json:"createBy"`
	CreateTime   *int64  `gorm:"column:create_time" json:"createTime"`
	UpdateTime   *int64  `gorm:"column:update_time" json:"updateTime"`
	DataFrom     *string `gorm:"column:data_from" json:"dataFrom"`
}

func (CoreChartView) TableName() string {
	return "core_chart_view"
}

type ChartQueryRequest struct {
	ID int64 `json:"id" binding:"required"`
}

type ChartDataRequest struct {
	ID          int64  `json:"id" binding:"required"`
	ResultCount *int   `json:"resultCount"`
	ResultMode  string `json:"resultMode"`
}

type ChartDataResponse struct {
	ChartID int64                    `json:"chartId"`
	Columns []string                 `json:"columns"`
	Rows    []map[string]interface{} `json:"rows"`
	Total   int64                    `json:"total"`
}

type ChartField struct {
	ID             int64  `json:"id"`
	DatasourceID   *int64 `json:"datasourceId,omitempty"`
	DatasetTableID *int64 `json:"datasetTableId,omitempty"`
	DatasetGroupID int64  `json:"datasetGroupId"`
	ChartID        *int64 `json:"chartId,omitempty"`
	OriginName     string `json:"originName"`
	Name           string `json:"name"`
	DataeaseName   string `json:"dataeaseName"`
	FieldShortName string `json:"fieldShortName"`
	GroupType      string `json:"groupType"`
	Type           string `json:"type"`
	DeType         int    `json:"deType"`
	DeExtractType  int    `json:"deExtractType"`
	ExtField       int    `json:"extField"`
	Checked        bool   `json:"checked"`
	Desensitized   bool   `json:"desensitized"`
	Summary        string `json:"summary"`
}

type ChartFieldListResponse struct {
	DimensionList []ChartField `json:"dimensionList"`
	QuotaList     []ChartField `json:"quotaList"`
}
