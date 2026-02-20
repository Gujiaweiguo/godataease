package msgcenter

type CountRequest struct{}

type ListRequest struct {
	Current    int    `json:"current"`
	Size       int    `json:"size"`
	ReadStatus string `json:"readStatus"`
	Type       string `json:"type"`
}

type Message struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Type       string `json:"type"`
	Level      string `json:"level"`
	Read       bool   `json:"read"`
	CreateTime int64  `json:"createTime"`
	ReadTime   int64  `json:"readTime,omitempty"`
}

type ListResponse struct {
	List    []Message `json:"list"`
	Total   int64     `json:"total"`
	Current int       `json:"current"`
	Size    int       `json:"size"`
}

type ReadRequest struct {
	ID string `json:"id" binding:"required"`
}

type ReadBatchRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

type ReadResponse struct {
	Success     bool `json:"success"`
	AlreadyRead bool `json:"alreadyRead"`
}

type ReadBatchResponse struct {
	Success bool `json:"success"`
	Updated int  `json:"updated"`
}
