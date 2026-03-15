package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"dataease/backend/internal/domain/share"
	"dataease/backend/internal/repository"

	"gorm.io/gorm"
)

var shareUUIDPattern = regexp.MustCompile(`^[a-zA-Z0-9]{8,16}$`)

type ShareService struct {
	repo *repository.ShareRepository
}

func NewShareService(repo *repository.ShareRepository) *ShareService {
	return &ShareService{repo: repo}
}

func (s *ShareService) CreateShare(req *share.ShareCreateRequest, creator int64) (*share.ShareCreateResponse, error) {
	uuid, err := generateUUID()
	if err != nil {
		return nil, err
	}

	pwd := ""
	autoPwd := req.AutoPwd
	if autoPwd {
		pwd, err = generatePassword(4)
		if err != nil {
			return nil, err
		}
	}

	sh := &share.Share{
		Creator:       creator,
		ResourceID:    req.ResourceID,
		ResourceType:  req.ResourceType,
		Exp:           req.Exp,
		UUID:          uuid,
		Pwd:           pwd,
		AutoPwd:       autoPwd,
		TicketRequire: false,
	}

	if err := s.repo.Create(sh); err != nil {
		return nil, err
	}

	return &share.ShareCreateResponse{
		ID:      sh.ID,
		UUID:    sh.UUID,
		Pwd:     sh.Pwd,
		AutoPwd: sh.AutoPwd,
	}, nil
}

func (s *ShareService) ValidateShare(req *share.ShareValidateRequest) (*share.ShareValidateResponse, error) {
	sh, err := s.repo.GetByUUID(req.UUID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &share.ShareValidateResponse{Valid: false}, nil
		}
		return nil, err
	}

	if isShareExpired(sh.Exp) {
		return &share.ShareValidateResponse{Valid: false}, nil
	}

	if sh.Pwd != "" && req.Pwd != sh.Pwd {
		return &share.ShareValidateResponse{Valid: false}, nil
	}

	return &share.ShareValidateResponse{
		Valid:         true,
		ResourceID:    sh.ResourceID,
		ResourceType:  sh.ResourceType,
		TicketRequire: sh.TicketRequire,
	}, nil
}

func (s *ShareService) RevokeShare(id int64, creator int64) (*share.ShareRevokeResponse, error) {
	sh, err := s.repo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &share.ShareRevokeResponse{Success: false}, nil
		}
		return nil, err
	}

	if sh.Creator != creator {
		return &share.ShareRevokeResponse{Success: false}, nil
	}

	if err := s.repo.Delete(id); err != nil {
		return nil, err
	}

	return &share.ShareRevokeResponse{Success: true}, nil
}

func (s *ShareService) ensureShareOwner(sh *share.Share, creator int64) error {
	if sh == nil {
		return gorm.ErrRecordNotFound
	}
	if creator <= 0 || sh.Creator != creator {
		return fmt.Errorf("forbidden")
	}
	return nil
}

func (s *ShareService) GetDetail(resourceID int64) (*share.ShareDetailResponse, error) {
	sh, err := s.repo.GetByResourceID(resourceID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &share.ShareDetailResponse{
		ID:            sh.ID,
		Exp:           sh.Exp,
		UUID:          sh.UUID,
		Pwd:           sh.Pwd,
		AutoPwd:       sh.AutoPwd,
		TicketRequire: sh.TicketRequire,
	}, nil
}

func (s *ShareService) SwitchStatus(resourceID int64, creator int64) error {
	sh, err := s.repo.GetByResourceID(resourceID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			_, err := s.CreateShare(&share.ShareCreateRequest{
				ResourceID:   resourceID,
				ResourceType: "dashboard",
				AutoPwd:      true,
			}, creator)
			return err
		}
		return err
	}

	return s.repo.Delete(sh.ID)
}

func (s *ShareService) CreateTicket(req *share.TicketCreateRequest) (*share.ShareTicket, error) {
	ticket := req.Ticket
	if req.GenerateNew || ticket == "" {
		var err error
		ticket, err = generateUUID()
		if err != nil {
			return nil, err
		}
	}

	t := &share.ShareTicket{
		UUID:   req.UUID,
		Ticket: ticket,
		Exp:    req.Exp,
		Args:   req.Args,
	}

	if err := s.repo.CreateTicket(t); err != nil {
		return nil, err
	}

	return t, nil
}

func (s *ShareService) ValidateTicket(req *share.TicketValidateRequest) (*share.TicketValidateResponse, error) {
	t, err := s.repo.GetTicketByTicket(req.Ticket)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &share.TicketValidateResponse{TicketValid: false, TicketExp: false}, nil
		}
		return nil, err
	}

	if t.UUID != req.UUID {
		return &share.TicketValidateResponse{TicketValid: false, TicketExp: false}, nil
	}

	if isShareExpired(t.Exp) {
		return &share.TicketValidateResponse{TicketValid: true, TicketExp: true}, nil
	}

	_ = s.repo.UpdateTicketAccessTime(req.Ticket)

	return &share.TicketValidateResponse{
		TicketValid: true,
		TicketExp:   false,
		Args:        t.Args,
	}, nil
}

func (s *ShareService) EditUUID(req *share.ShareEditUUIDRequest, creator int64) (string, error) {
	sh, err := s.repo.GetByResourceID(req.ResourceID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "share not found", nil
		}
		return "", err
	}
	if err = s.ensureShareOwner(sh, creator); err != nil {
		return "", err
	}

	candidate := strings.TrimSpace(req.UUID)
	if candidate == "" {
		return "uuid cannot be empty", nil
	}
	if candidate == sh.UUID {
		return "", nil
	}
	if !shareUUIDPattern.MatchString(candidate) {
		return "invalid uuid format", nil
	}
	exists, err := s.repo.ExistsByUUID(candidate, sh.ID)
	if err != nil {
		return "", err
	}
	if exists {
		return "uuid already exists", nil
	}
	oldUUID := sh.UUID
	if oldUUID != candidate {
		if err = s.repo.UpdateUUIDWithTickets(sh.ID, oldUUID, candidate); err != nil {
			return "", err
		}
	}
	return "", nil
}

func (s *ShareService) EditExp(req *share.ShareEditExpRequest, creator int64) error {
	sh, err := s.repo.GetByResourceID(req.ResourceID)
	if err != nil {
		return err
	}
	if err = s.ensureShareOwner(sh, creator); err != nil {
		return err
	}
	if req.Exp > 0 && req.Exp < time.Now().UnixMilli() {
		return fmt.Errorf("invalid expiration")
	}
	sh.Exp = req.Exp
	return s.repo.Update(sh)
}

func (s *ShareService) EditPwd(req *share.ShareEditPwdRequest, creator int64) error {
	sh, err := s.repo.GetByResourceID(req.ResourceID)
	if err != nil {
		return err
	}
	if err = s.ensureShareOwner(sh, creator); err != nil {
		return err
	}
	pwd := strings.TrimSpace(req.Pwd)
	if pwd == "" {
		sh.Pwd = ""
		sh.AutoPwd = req.AutoPwd
		return s.repo.Update(sh)
	}
	if !req.AutoPwd && !isValidSharePassword(pwd) {
		return fmt.Errorf("invalid password format")
	}
	sh.Pwd = pwd
	sh.AutoPwd = req.AutoPwd
	return s.repo.Update(sh)
}

func generateUUID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func generatePassword(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

func isShareExpired(exp int64) bool {
	if exp <= 0 {
		return false
	}
	if exp >= 1_000_000_000_000 {
		return time.Now().UnixMilli() > exp
	}
	return time.Now().Unix() > exp
}

func isValidSharePassword(pwd string) bool {
	if len(pwd) < 4 || len(pwd) > 10 {
		return false
	}
	hasLetter := false
	hasDigit := false
	hasSpecial := false
	for _, r := range pwd {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+", r):
			hasSpecial = true
		default:
			return false
		}
	}
	return hasLetter && hasDigit && hasSpecial
}
