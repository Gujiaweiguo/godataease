package template

import "time"

// Template represents a visualization template
type Template struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	Pid           int64      `json:"pid"`
	Level         int        `json:"level"`
	DvType        string     `json:"dvType"`
	NodeType      string     `json:"nodeType"`
	CreateBy      string     `json:"createBy"`
	CreateTime    *time.Time `json:"createTime,omitempty"`
	Snapshot      string     `json:"snapshot"`
	TemplateType  string     `json:"templateType"`
	TemplateStyle string     `json:"templateStyle"`
	TemplateData  string     `json:"templateData"`
	DynamicData   string     `json:"dynamicData"`
	AppData       string     `json:"appData"`
	UseCount      int        `json:"useCount"`
	Version       int        `json:"version"`
}

// TemplateListRequest represents request to list templates
type TemplateListRequest struct {
	Pid        string `json:"pid"`
	DvType     string `json:"dvType"`
	WithBlobs  string `json:"withBlobs"`
	Keyword    string `json:"keyword"`
	CategoryID string `json:"categoryId"`
	Sort       string `json:"sort"`
}

// TemplateListResponse represents response for template list
type TemplateListResponse struct {
	List  []Template `json:"list"`
	Total int64      `json:"total"`
}

// TemplateCreateRequest represents request to create a template
type TemplateCreateRequest struct {
	Name          string `json:"name" binding:"required"`
	Pid           int64  `json:"pid"`
	DvType        string `json:"dvType"`
	NodeType      string `json:"nodeType"`
	Snapshot      string `json:"snapshot"`
	TemplateType  string `json:"templateType"`
	TemplateStyle string `json:"templateStyle"`
	TemplateData  string `json:"templateData"`
	DynamicData   string `json:"dynamicData"`
	AppData       string `json:"appData"`
}

// TemplateUpdateRequest represents request to update a template
type TemplateUpdateRequest struct {
	ID            int64  `json:"id" binding:"required"`
	Name          string `json:"name"`
	Snapshot      string `json:"snapshot"`
	TemplateStyle string `json:"templateStyle"`
	TemplateData  string `json:"templateData"`
	DynamicData   string `json:"dynamicData"`
	AppData       string `json:"appData"`
}

// TemplateDeleteRequest represents request to delete a template
type TemplateDeleteRequest struct {
	ID         int64 `json:"id" binding:"required"`
	CategoryID int64 `json:"categoryId"`
}
