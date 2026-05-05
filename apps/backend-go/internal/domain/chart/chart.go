package chart

type CoreChartView struct {
	ID                int64   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Title             *string `gorm:"column:title" json:"title"`
	SceneID           *int64  `gorm:"column:scene_id" json:"sceneId"`
	TableID           *int64  `gorm:"column:table_id" json:"tableId"`
	Type              *string `gorm:"column:type" json:"type"`
	Render            *string `gorm:"column:render" json:"render"`
	ResultCount       *int    `gorm:"column:result_count" json:"resultCount"`
	ResultMode        *string `gorm:"column:result_mode" json:"resultMode"`
	XAxis             *string `gorm:"column:x_axis" json:"xAxis"`
	YAxis             *string `gorm:"column:y_axis" json:"yAxis"`
	CustomAttr        *string `gorm:"column:custom_attr" json:"customAttr"`
	CustomStyle       *string `gorm:"column:custom_style" json:"customStyle"`
	CustomFilter      *string `gorm:"column:custom_filter" json:"customFilter"`
	CreateBy          *string `gorm:"column:create_by" json:"createBy"`
	CreateTime        *int64  `gorm:"column:create_time" json:"createTime"`
	UpdateTime        *int64  `gorm:"column:update_time" json:"updateTime"`
	DataFrom          *string `gorm:"column:data_from" json:"dataFrom"`
	XAxisExt          *string `gorm:"column:x_axis_ext" json:"xAxisExt"`
	YAxisExt          *string `gorm:"column:y_axis_ext" json:"yAxisExt"`
	ExtStack          *string `gorm:"column:ext_stack" json:"extStack"`
	ExtBubble         *string `gorm:"column:ext_bubble" json:"extBubble"`
	ExtLabel          *string `gorm:"column:ext_label" json:"extLabel"`
	ExtTooltip        *string `gorm:"column:ext_tooltip" json:"extTooltip"`
	CustomAttrMobile  *string `gorm:"column:custom_attr_mobile" json:"customAttrMobile"`
	CustomStyleMobile *string `gorm:"column:custom_style_mobile" json:"customStyleMobile"`
	DrillFields       *string `gorm:"column:drill_fields" json:"drillFields"`
	Senior            *string `gorm:"column:senior" json:"senior"`
	Snapshot          *string `gorm:"column:snapshot" json:"snapshot"`
	StylePriority     *string `gorm:"column:style_priority" json:"stylePriority"`
	ChartType         *string `gorm:"column:chart_type" json:"chartType"`
	IsPlugin          *bool   `gorm:"column:is_plugin" json:"isPlugin"`
	ViewFields        *string `gorm:"column:view_fields" json:"viewFields"`
	RefreshViewEnable *bool   `gorm:"column:refresh_view_enable" json:"refreshViewEnable"`
	RefreshUnit       *string `gorm:"column:refresh_unit" json:"refreshUnit"`
	RefreshTime       *int    `gorm:"column:refresh_time" json:"refreshTime"`
	LinkageActive     *bool   `gorm:"column:linkage_active" json:"linkageActive"`
	JumpActive        *bool   `gorm:"column:jump_active" json:"jumpActive"`
	CopyFrom          *int64  `gorm:"column:copy_from" json:"copyFrom"`
	CopyID            *int64  `gorm:"column:copy_id" json:"copyId"`
	Aggregate         *bool   `gorm:"column:aggregate" json:"aggregate"`
	FlowMapStartName  *string `gorm:"column:flow_map_start_name" json:"flowMapStartName"`
	FlowMapEndName    *string `gorm:"column:flow_map_end_name" json:"flowMapEndName"`
	ExtColor          *string `gorm:"column:ext_color" json:"extColor"`
	SortPriority      *string `gorm:"column:sort_priority" json:"sortPriority"`
}

func (CoreChartView) TableName() string {
	return "core_chart_view"
}

type ChartQueryRequest struct {
	ID int64 `json:"id" binding:"required"`
}

type ChartDataRequest struct {
	ID          int64                  `json:"id" binding:"required"`
	ResultCount *int                   `json:"resultCount"`
	ResultMode  string                 `json:"resultMode"`
	Payload     map[string]interface{} `json:"-"`
}

type ChartDataResponse struct {
	ChartID      int64                    `json:"chartId,omitempty"`
	Columns      []string                 `json:"columns,omitempty"`
	Rows         []map[string]interface{} `json:"rows,omitempty"`
	Total        int64                    `json:"total,omitempty"`
	Data         []ChartDataPoint         `json:"data,omitempty"`
	Fields       []map[string]interface{} `json:"fields,omitempty"`
	TableRow     []map[string]interface{} `json:"tableRow,omitempty"`
	SourceFields []map[string]interface{} `json:"sourceFields,omitempty"`
}

type ChartDataPoint struct {
	Field         string               `json:"field,omitempty"`
	Name          string               `json:"name,omitempty"`
	Category      string               `json:"category,omitempty"`
	Value         float64              `json:"value"`
	DimensionList []ChartDataFieldItem `json:"dimensionList,omitempty"`
	QuotaList     []ChartDataFieldItem `json:"quotaList,omitempty"`
}

type ChartDataFieldItem struct {
	ID    string      `json:"id"`
	Value interface{} `json:"value,omitempty"`
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

// ViewSelectorVO is returned by /chart/viewOption — a lightweight chart selector item
type ViewSelectorVO struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Type  string `json:"type"`
	PID   *int64 `json:"pid"`
}

// ChartBaseVO is returned by /chart/chartBaseInfo — full chart metadata with parsed axis fields
type ChartBaseVO struct {
	ChartID          int64                    `json:"chartId"`
	ChartType        string                   `json:"chartType"`
	ChartName        string                   `json:"chartName"`
	ResourceID       int64                    `json:"resourceId"`
	ResourceType     string                   `json:"resourceType"`
	ResourceName     string                   `json:"resourceName"`
	TableID          *int64                   `json:"tableId"`
	XAxis            []map[string]interface{} `json:"xAxis"`
	XAxisExt         []map[string]interface{} `json:"xAxisExt"`
	YAxis            []map[string]interface{} `json:"yAxis"`
	YAxisExt         []map[string]interface{} `json:"yAxisExt"`
	ExtStack         []map[string]interface{} `json:"extStack"`
	ExtBubble        []map[string]interface{} `json:"extBubble"`
	FlowMapStartName []map[string]interface{} `json:"flowMapStartName"`
	FlowMapEndName   []map[string]interface{} `json:"flowMapEndName"`
	ExtColor         []map[string]interface{} `json:"extColor"`
	ExtLabel         []map[string]interface{} `json:"extLabel"`
	ExtTooltip       []map[string]interface{} `json:"extTooltip"`
}
