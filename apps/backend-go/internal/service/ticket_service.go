package service

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"dataease/backend/internal/domain/ticket"
	"dataease/backend/internal/repository"
)

type TicketService struct {
	repo *repository.TicketRepository
}

func NewTicketService(repo *repository.TicketRepository) *TicketService {
	return &TicketService{repo: repo}
}

func (s *TicketService) CreateTicket(req *ticket.TicketCreateRequest) (*ticket.TicketCreateResponse, error) {
	ticketStr := req.Ticket
	if ticketStr == "" || req.GenerateNew {
		ticketStr = s.generateTicket()
	}

	t := &ticket.Ticket{
		UUID:   req.UUID,
		Ticket: ticketStr,
		Exp:    req.Exp,
		Args:   req.Args,
	}

	if err := s.repo.Create(t); err != nil {
		return nil, err
	}

	return &ticket.TicketCreateResponse{Ticket: ticketStr}, nil
}

func (s *TicketService) ValidateTicket(ticketStr string) *ticket.TicketValidateResponse {
	t, err := s.repo.FindByTicket(ticketStr)
	if err != nil || t == nil {
		return &ticket.TicketValidateResponse{
			TicketValid: false,
			TicketExp:   false,
		}
	}

	now := time.Now().Unix()
	isExpired := t.Exp > 0 && now > t.Exp

	if isExpired {
		return &ticket.TicketValidateResponse{
			TicketValid: false,
			TicketExp:   true,
		}
	}

	_ = s.repo.UpdateAccessTime(ticketStr, now)

	return &ticket.TicketValidateResponse{
		TicketValid: true,
		TicketExp:   false,
		Args:        t.Args,
	}
}

func (s *TicketService) DeleteTicket(req *ticket.TicketDeleteRequest) error {
	return s.repo.Delete(req.Ticket)
}

func (s *TicketService) ListTickets(uuid string, page, pageSize int) *ticket.TicketListResponse {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	list, total, err := s.repo.ListByUUID(uuid, page, pageSize)
	if err != nil {
		return &ticket.TicketListResponse{
			List:    []ticket.Ticket{},
			Total:   0,
			Current: page,
			Size:    pageSize,
		}
	}

	return &ticket.TicketListResponse{
		List:    list,
		Total:   total,
		Current: page,
		Size:    pageSize,
	}
}

func (s *TicketService) TempTicket() string {
	return s.generateTicket()
}

func (s *TicketService) generateTicket() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
