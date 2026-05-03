package threshold

// BaseReciDTO mirrors Java BaseReciDTO — recipient configuration.
type BaseReciDTO struct {
	ReciFlagList       []int    `json:"reciFlagList"`
	UIDList            []string `json:"uidList"`
	RIDList            []string `json:"ridList"`
	EmailList          []string `json:"emailList"`
	LarkGroupList      []string `json:"larkGroupList"`
	LarksuiteGroupList []string `json:"larksuiteGroupList"`
	WebhookList        []string `json:"webhookList"`
}

// CreateRequest mirrors Java ThresholdCreator.
type CreateRequest struct {
	BaseReciDTO
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Enable         *bool  `json:"enable"`
	RateType       *int   `json:"rateType"`
	RateValue      string `json:"rateValue"`
	ResourceID     int64  `json:"resourceId"`
	ResourceType   string `json:"resourceType"`
	ChartID        int64  `json:"chartId"`
	ChartType      string `json:"chartType"`
	ThresholdRules string `json:"thresholdRules"`
	MsgType        *int   `json:"msgType"`
	MsgTitle       string `json:"msgTitle"`
	MsgContent     string `json:"msgContent"`
	RepeatSend     *bool  `json:"repeatSend"`
	ShowFieldValue *bool  `json:"showFieldValue"`
	ResourceTable  string `json:"resourceTable"`
}

// GridRequest mirrors Java ThresholdGridRequest.
type GridRequest struct {
	Keyword          string   `json:"keyword"`
	ResourceTable    string   `json:"resourceTable"`
	ResourceTypeList []string `json:"resourceTypeList"`
	StatusList       []int    `json:"statusList"`
	EnableList       []int    `json:"enableList"`
	TimeList         []int64  `json:"timeList"`
	ChartID          *int64   `json:"chartId,omitempty"`
}

// GridVO mirrors Java ThresholdGridVO.
type GridVO struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	ResourceID   int64  `json:"resourceId"`
	ResourceType string `json:"resourceType"`
	ResourceName string `json:"resourceName"`
	ChartID      int64  `json:"chartId"`
	ChartType    string `json:"chartType"`
	ChartName    string `json:"chartName"`
	Status       bool   `json:"status"`
	Enable       bool   `json:"enable"`
	Creator      int64  `json:"creator"`
	CreateName   string `json:"createName"`
	CreateTime   int64  `json:"createTime"`
}

// SwitchRequest mirrors Java ThresholdSwitchRequest.
type SwitchRequest struct {
	ID            int64  `json:"id"`
	Enable        *bool  `json:"enable"`
	ResourceTable string `json:"resourceTable"`
}

// BatchReciRequest mirrors Java ThresholdBatchReciRequest.
type BatchReciRequest struct {
	BaseReciDTO
	IDList []int64 `json:"idList"`
}

// PreviewRequest mirrors Java ThresholdPreviewRequest.
type PreviewRequest struct {
	ChartID        int64  `json:"chartId"`
	ThresholdRules string `json:"thresholdRules"`
	MsgContent     string `json:"msgContent"`
	ResourceTable  string `json:"resourceTable"`
}

// PreviewResponse wraps preview result.
type PreviewResponse struct {
	Content string `json:"content"`
}

// InstanceRequest mirrors Java ThresholdInstanceRequest.
type InstanceRequest struct {
	Keyword     string `json:"keyword"`
	ThresholdID *int64 `json:"thresholdId,omitempty"`
}

// InstanceVO mirrors Java ThresholdInstanceVO.
type InstanceVO struct {
	ID       int64  `json:"id"`
	TaskID   int64  `json:"taskId"`
	Name     string `json:"name"`
	ExecTime int64  `json:"execTime"`
	Status   bool   `json:"status"`
	Content  string `json:"content"`
	Msg      string `json:"msg"`
}

// PageResult is a generic paginated response matching Java IPage contract.
type PageResult struct {
	List    any   `json:"list"`
	Total   int64 `json:"total"`
	Current int   `json:"current"`
	Size    int   `json:"size"`
}
