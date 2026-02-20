package share

import "time"

// ShareTicket represents a share ticket for resource access
type ShareTicket struct {
	ID         int64      `json:"id"`
	UUID       string     `json:"uuid"`
	Ticket     string     `json:"ticket"`
	Exp        int64      `json:"exp"`
	Args       string     `json:"args"`
	AccessTime *time.Time `json:"accessTime,omitempty"`
}

// Share represents a share record for a resource
type Share struct {
	ID            int64      `json:"id"`
	Creator       int64      `json:"creator"`
	ResourceID    int64      `json:"resourceId"`
	ResourceType  string     `json:"resourceType"`
	Time          *time.Time `json:"time,omitempty"`
	Exp           int64      `json:"exp"`
	UUID          string     `json:"uuid"`
	Pwd           string     `json:"pwd,omitempty"`
	AutoPwd       bool       `json:"autoPwd"`
	TicketRequire bool       `json:"ticketRequire"`
}

// ShareCreateRequest represents request to create a share
type ShareCreateRequest struct {
	ResourceID   int64  `json:"resourceId" binding:"required"`
	ResourceType string `json:"resourceType" binding:"required"`
	Exp          int64  `json:"exp"`
	AutoPwd      bool   `json:"autoPwd"`
}

// ShareCreateResponse represents response after creating a share
type ShareCreateResponse struct {
	ID      int64  `json:"id"`
	UUID    string `json:"uuid"`
	Pwd     string `json:"pwd,omitempty"`
	AutoPwd bool   `json:"autoPwd"`
}

// ShareValidateRequest represents request to validate a share
type ShareValidateRequest struct {
	UUID string `json:"uuid" binding:"required"`
	Pwd  string `json:"pwd"`
}

// ShareValidateResponse represents response for share validation
type ShareValidateResponse struct {
	Valid         bool   `json:"valid"`
	ResourceID    int64  `json:"resourceId,omitempty"`
	ResourceType  string `json:"resourceType,omitempty"`
	TicketRequire bool   `json:"ticketRequire"`
}

// ShareRevokeRequest represents request to revoke a share
type ShareRevokeRequest struct {
	ID int64 `json:"id" binding:"required"`
}

// ShareRevokeResponse represents response for share revocation
type ShareRevokeResponse struct {
	Success bool `json:"success"`
}

// TicketCreateRequest represents request to create a ticket
type TicketCreateRequest struct {
	Ticket      string `json:"ticket" binding:"required"`
	Exp         int64  `json:"exp"`
	Args        string `json:"args"`
	UUID        string `json:"uuid" binding:"required"`
	GenerateNew bool   `json:"generateNew"`
}

// TicketValidateRequest represents request to validate a ticket
type TicketValidateRequest struct {
	Ticket string `json:"ticket" binding:"required"`
	UUID   string `json:"uuid" binding:"required"`
}

// TicketValidateResponse represents response for ticket validation
type TicketValidateResponse struct {
	TicketValid bool   `json:"ticketValid"`
	TicketExp   bool   `json:"ticketExp"`
	Args        string `json:"args,omitempty"`
}

// ShareDetailResponse represents detailed share information
type ShareDetailResponse struct {
	ID            int64  `json:"id"`
	Exp           int64  `json:"exp"`
	UUID          string `json:"uuid"`
	Pwd           string `json:"pwd,omitempty"`
	AutoPwd       bool   `json:"autoPwd"`
	TicketRequire bool   `json:"ticketRequire"`
}
