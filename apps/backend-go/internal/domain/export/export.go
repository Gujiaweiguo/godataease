package export

type ExportTask struct {
	ID                string  `json:"id"`
	UserID            int64   `json:"userId"`
	FileName          string  `json:"fileName"`
	FileSize          float64 `json:"fileSize"`
	FileSizeUnit      string  `json:"fileSizeUnit"`
	ExportFrom        int64   `json:"exportFrom"`
	ExportStatus      string  `json:"exportStatus"`
	Msg               string  `json:"msg"`
	ExportFromType    string  `json:"exportFromType"`
	ExportTime        int64   `json:"exportTime"`
	ExportProgress    string  `json:"exportProgress"`
	ExportMachineName string  `json:"exportMachineName"`
	ExportFromName    string  `json:"exportFromName"`
	OrgName           string  `json:"orgName"`
}

type ExportTasksRequest struct{}

type ExportTasksResponse map[string]int64

type PagerRequest struct {
	GoPage   int    `json:"goPage" form:"goPage"`
	PageSize int    `json:"pageSize" form:"pageSize"`
	Status   string `json:"status" form:"status"`
}

type PagerResponse struct {
	List     []ExportTask `json:"records"`
	Total    int64        `json:"total"`
	PageNum  int          `json:"current"`
	PageSize int          `json:"size"`
}

type DeleteRequest struct {
	IDs []string `json:"ids"`
}

type DeleteAllRequest struct {
	Type string `json:"type" form:"type"`
}

type DownloadRequest struct {
	ID string `json:"id" form:"id"`
}

type DownloadResponse struct {
	URL string `json:"url"`
}

type RetryRequest struct {
	ID string `json:"id" form:"id"`
}

type ExportLimitResponse struct {
	Limit string `json:"limit"`
}
