package visualization

type CanvasChartView struct {
	ID                int64   `gorm:"column:id;primaryKey" json:"id"`
	Title             *string `gorm:"column:title" json:"title"`
	SceneID           *int64  `gorm:"column:scene_id" json:"sceneId"`
	TableID           *int64  `gorm:"column:table_id" json:"tableId"`
	Type              *string `gorm:"column:type" json:"type"`
	Render            *string `gorm:"column:render" json:"render"`
	ResultCount       *int    `gorm:"column:result_count" json:"resultCount"`
	ResultMode        *string `gorm:"column:result_mode" json:"resultMode"`
	XAxis             *string `gorm:"column:x_axis" json:"xAxis"`
	XAxisExt          *string `gorm:"column:x_axis_ext" json:"xAxisExt"`
	YAxis             *string `gorm:"column:y_axis" json:"yAxis"`
	YAxisExt          *string `gorm:"column:y_axis_ext" json:"yAxisExt"`
	ExtStack          *string `gorm:"column:ext_stack" json:"extStack"`
	ExtBubble         *string `gorm:"column:ext_bubble" json:"extBubble"`
	ExtLabel          *string `gorm:"column:ext_label" json:"extLabel"`
	ExtTooltip        *string `gorm:"column:ext_tooltip" json:"extTooltip"`
	CustomAttr        *string `gorm:"column:custom_attr" json:"customAttr"`
	CustomAttrMobile  *string `gorm:"column:custom_attr_mobile" json:"customAttrMobile"`
	CustomStyle       *string `gorm:"column:custom_style" json:"customStyle"`
	CustomStyleMobile *string `gorm:"column:custom_style_mobile" json:"customStyleMobile"`
	CustomFilter      *string `gorm:"column:custom_filter" json:"customFilter"`
	DrillFields       *string `gorm:"column:drill_fields" json:"drillFields"`
	Senior            *string `gorm:"column:senior" json:"senior"`
	CreateBy          *string `gorm:"column:create_by" json:"createBy"`
	CreateTime        *int64  `gorm:"column:create_time" json:"createTime"`
	UpdateTime        *int64  `gorm:"column:update_time" json:"updateTime"`
	Snapshot          *string `gorm:"column:snapshot" json:"snapshot"`
	StylePriority     *string `gorm:"column:style_priority" json:"stylePriority"`
	ChartType         *string `gorm:"column:chart_type" json:"chartType"`
	IsPlugin          *bool   `gorm:"column:is_plugin" json:"isPlugin"`
	DataFrom          *string `gorm:"column:data_from" json:"dataFrom"`
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

func (CanvasChartView) TableName() string {
	return "core_chart_view"
}

type SnapshotCanvasChartView CanvasChartView

func (SnapshotCanvasChartView) TableName() string {
	return "snapshot_core_chart_view"
}
