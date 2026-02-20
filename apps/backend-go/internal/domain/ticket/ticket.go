package ticket

// Ticket represents a share ticket entity
type Ticket struct {
	ID         int64  `json:"id"`
	UUID       string `json:"uuid"`
	Ticket     string `json:"ticket"`
	Exp        int64  `json:"exp"`
	Args       string `json:"args"`
	AccessTime int64  `json:"accessTime"`
}

// TicketCreateRequest represents the request to create a ticket
type TicketCreateRequest struct {
	Ticket      string `json:"ticket" binding:"required"`
	Exp         int64  `json:"exp" binding:"required"`
	Args        string `json:"args"`
	UUID        string `json:"uuid" binding:"required"`
	GenerateNew bool   `json:"generateNew"`
}

// TicketCreateResponse represents the response after creating a ticket
type TicketCreateResponse struct {
	Ticket string `json:"ticket"`
}

// TicketDeleteRequest represents the request to delete a ticket
type TicketDeleteRequest struct {
	Ticket string `json:"ticket" binding:"required"`
}

// TicketValidateResponse represents the validation result of a ticket
type TicketValidateResponse struct {
	TicketValid bool   `json:"ticketValid"`
	TicketExp   bool   `json:"ticketExp"`
	Args        string `json:"args"`
}

// TicketListResponse represents a paginated list of tickets
type TicketListResponse struct {
	List    []Ticket `json:"list"`
	Total   int64    `json:"total"`
	Current int      `json:"current"`
	Size    int      `json:"size"`
}
