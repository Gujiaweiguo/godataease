package service

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"dataease/backend/internal/domain/audit"
	"dataease/backend/internal/pkg/logger"
	"dataease/backend/internal/repository"

	"go.uber.org/zap"
)

const DefaultRetentionDays = 90

type AuditService struct {
	auditLogRepo       *repository.AuditLogRepository
	loginFailureRepo   *repository.LoginFailureRepository
	auditLogDetailRepo *repository.AuditLogDetailRepository
}

func NewAuditService(
	auditLogRepo *repository.AuditLogRepository,
	loginFailureRepo *repository.LoginFailureRepository,
	auditLogDetailRepo *repository.AuditLogDetailRepository,
) *AuditService {
	return &AuditService{
		auditLogRepo:       auditLogRepo,
		loginFailureRepo:   loginFailureRepo,
		auditLogDetailRepo: auditLogDetailRepo,
	}
}

type PaginatedResult struct {
	List    interface{} `json:"list"`
	Total   int64       `json:"total"`
	Current int         `json:"current"`
	Size    int         `json:"size"`
}

func (s *AuditService) CreateAuditLog(req *audit.AuditLogCreateRequest) (*audit.AuditLog, error) {
	log := &audit.AuditLog{
		UserID:         req.UserID,
		Username:       req.Username,
		ActionType:     req.ActionType,
		ActionName:     req.ActionName,
		ResourceType:   req.ResourceType,
		ResourceID:     req.ResourceID,
		ResourceName:   req.ResourceName,
		Operation:      req.Operation,
		IPAddress:      req.IPAddress,
		UserAgent:      req.UserAgent,
		BeforeValue:    req.BeforeValue,
		AfterValue:     req.AfterValue,
		OrganizationID: req.OrganizationID,
		CreateTime:     time.Now(),
	}

	if req.Status != nil {
		log.Status = *req.Status
	} else {
		log.Status = audit.StatusSuccess
	}

	if req.FailureReason != nil {
		log.FailureReason = req.FailureReason
	}

	if err := s.auditLogRepo.Create(log); err != nil {
		logger.Error("Failed to create audit log", zap.Error(err))
		return nil, fmt.Errorf("failed to create audit log: %w", err)
	}

	logger.Info("Audit log created", zap.Int64("id", log.ID), zap.String("action", string(log.ActionType)))
	return log, nil
}

func (s *AuditService) GetAuditLogByID(id int64) (*audit.AuditLog, error) {
	return s.auditLogRepo.GetByID(id)
}

func (s *AuditService) GetAuditLogsByUserID(userID int64, page, pageSize int) (*PaginatedResult, error) {
	logs, total, err := s.auditLogRepo.GetByUserID(userID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}

	return &PaginatedResult{
		List:    logs,
		Total:   total,
		Current: page,
		Size:    pageSize,
	}, nil
}

func (s *AuditService) QueryAuditLogs(query *audit.AuditLogQuery) (*PaginatedResult, error) {
	logs, total, err := s.auditLogRepo.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}

	page := query.Page
	pageSize := query.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	return &PaginatedResult{
		List:    logs,
		Total:   total,
		Current: page,
		Size:    pageSize,
	}, nil
}

func (s *AuditService) RecordLoginFailure(req *audit.LoginFailureRequest) (*audit.LoginFailure, error) {
	failure := &audit.LoginFailure{
		Username:      req.Username,
		IPAddress:     req.IPAddress,
		FailureReason: req.FailureReason,
		UserAgent:     req.UserAgent,
		CreateTime:    time.Now(),
	}

	if err := s.loginFailureRepo.Create(failure); err != nil {
		logger.Error("Failed to record login failure", zap.Error(err))
		return nil, fmt.Errorf("failed to record login failure: %w", err)
	}

	logger.Info("Login failure recorded", zap.String("username", req.Username))
	return failure, nil
}

func (s *AuditService) DeleteAuditLogsBeforeDate(days int) (int64, error) {
	if days <= 0 {
		days = DefaultRetentionDays
	}

	beforeTime := time.Now().AddDate(0, 0, -days)
	affected, err := s.auditLogRepo.DeleteBeforeDate(beforeTime)
	if err != nil {
		return 0, fmt.Errorf("failed to delete audit logs: %w", err)
	}

	logger.Info("Audit logs deleted", zap.Int64("count", affected), zap.Int("days", days))
	return affected, nil
}

func (s *AuditService) ExportAuditLogs(ids []int64, format string) (string, error) {
	logs, err := s.auditLogRepo.GetByIDs(ids)
	if err != nil {
		return "", fmt.Errorf("failed to get audit logs: %w", err)
	}

	if len(logs) == 0 {
		return "", fmt.Errorf("no audit logs found")
	}

	timestamp := time.Now().Format("20060102150405")
	var filePath string

	switch format {
	case "csv":
		filePath = fmt.Sprintf("/tmp/audit_logs_%s.csv", timestamp)
		if err := s.exportToCSV(logs, filePath); err != nil {
			return "", err
		}
	case "json":
		filePath = fmt.Sprintf("/tmp/audit_logs_%s.json", timestamp)
		if err := s.exportToJSON(logs, filePath); err != nil {
			return "", err
		}
	default:
		filePath = fmt.Sprintf("/tmp/audit_logs_%s.csv", timestamp)
		if err := s.exportToCSV(logs, filePath); err != nil {
			return "", err
		}
	}

	return filePath, nil
}

func (s *AuditService) exportToCSV(logs []*audit.AuditLog, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"ID", "User ID", "Username", "Action Type", "Action Name", "Resource Type", "Resource ID", "Operation", "Status", "IP Address", "Create Time"}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	for _, log := range logs {
		var userID, resourceID string
		if log.UserID != nil {
			userID = strconv.FormatInt(*log.UserID, 10)
		}
		if log.ResourceID != nil {
			resourceID = strconv.FormatInt(*log.ResourceID, 10)
		}

		var ipAddress string
		if log.IPAddress != nil {
			ipAddress = *log.IPAddress
		}

		var username, resourceType string
		if log.Username != nil {
			username = *log.Username
		}
		if log.ResourceType != nil {
			resourceType = *log.ResourceType
		}

		record := []string{
			strconv.FormatInt(log.ID, 10),
			userID,
			username,
			string(log.ActionType),
			log.ActionName,
			resourceType,
			resourceID,
			string(log.Operation),
			string(log.Status),
			ipAddress,
			log.CreateTime.Format(time.RFC3339),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}

func (s *AuditService) exportToJSON(logs []*audit.AuditLog, filePath string) error {
	data, err := json.MarshalIndent(logs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal logs: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (s *AuditService) CheckSuspiciousLoginActivity(username string, threshold int, duration time.Duration) (bool, error) {
	since := time.Now().Add(-duration)
	count, err := s.loginFailureRepo.CountSinceTime(username, since)
	if err != nil {
		return false, err
	}
	return count >= int64(threshold), nil
}
