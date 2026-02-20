package service

import (
	"dataease/backend/internal/domain/msgcenter"
	"dataease/backend/internal/repository"
)

type MsgCenterService struct {
	repo *repository.MsgCenterRepository
}

func NewMsgCenterService(repo *repository.MsgCenterRepository) *MsgCenterService {
	return &MsgCenterService{repo: repo}
}

func (s *MsgCenterService) Count(_ *msgcenter.CountRequest) int64 {
	return 0
}

func (s *MsgCenterService) List(req *msgcenter.ListRequest) *msgcenter.ListResponse {
	current := req.Current
	if current < 1 {
		current = 1
	}
	size := req.Size
	if size < 1 {
		size = 10
	}

	return &msgcenter.ListResponse{
		List:    make([]msgcenter.Message, 0),
		Total:   0,
		Current: current,
		Size:    size,
	}
}

func (s *MsgCenterService) Read(req *msgcenter.ReadRequest, userID int64) *msgcenter.ReadResponse {
	isRead, err := s.repo.IsRead(req.ID, userID)
	if err != nil {
		return &msgcenter.ReadResponse{Success: false, AlreadyRead: false}
	}

	if isRead {
		return &msgcenter.ReadResponse{Success: true, AlreadyRead: true}
	}

	if err := s.repo.MarkAsRead(req.ID, userID); err != nil {
		return &msgcenter.ReadResponse{Success: false, AlreadyRead: false}
	}

	return &msgcenter.ReadResponse{Success: true, AlreadyRead: false}
}

func (s *MsgCenterService) ReadBatch(req *msgcenter.ReadBatchRequest, userID int64) *msgcenter.ReadBatchResponse {
	updated, err := s.repo.MarkBatchAsRead(req.IDs, userID)
	if err != nil {
		return &msgcenter.ReadBatchResponse{Success: false, Updated: 0}
	}

	return &msgcenter.ReadBatchResponse{Success: true, Updated: updated}
}
