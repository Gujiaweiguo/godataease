package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"dataease/backend/internal/domain/auto"
	thresholddomain "dataease/backend/internal/domain/threshold"
	"gorm.io/gorm"
)

type ThresholdRepo interface {
	Create(ctx context.Context, info *auto.XpackThresholdInfo) error
	Update(ctx context.Context, info *auto.XpackThresholdInfo) error
	GetByID(ctx context.Context, id int64) (*auto.XpackThresholdInfo, error)
	DeleteByIDs(ctx context.Context, ids []int64) error
	DeleteByChartID(ctx context.Context, chartID int64) error
	UpdateEnable(ctx context.Context, id int64, enable bool) error
	UpdateRecipients(ctx context.Context, ids []int64, users, roles, emails, larkGroups, larksuiteGroups, webhooks string) error
	Pager(ctx context.Context, req *thresholddomain.GridRequest, goPage, pageSize int) ([]*thresholddomain.GridVO, int64, error)
	ExistsByChartID(ctx context.Context, chartID int64) (bool, error)
	InstancePager(ctx context.Context, req *thresholddomain.InstanceRequest, goPage, pageSize int) ([]*thresholddomain.InstanceVO, int64, error)
}

var errThresholdRepoNotReady = errors.New("threshold repository not initialized")

type ThresholdService struct {
	repo ThresholdRepo
}

func NewThresholdService(repo ThresholdRepo) *ThresholdService {
	return &ThresholdService{repo: repo}
}

func (s *ThresholdService) Create(ctx context.Context, req *thresholddomain.CreateRequest, userID int64, userName string, oid int64) (*auto.XpackThresholdInfo, error) {
	if s.repo == nil {
		return nil, errThresholdRepoNotReady
	}
	if err := validateThresholdCreateRequest(req); err != nil {
		return nil, err
	}

	item := buildThresholdInfo(req)
	item.ID = 0
	item.Creator = userID
	item.CreatorName = userName
	item.CreateTime = time.Now().UnixMilli()
	item.Oid = oid

	if err := s.repo.Create(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *ThresholdService) Edit(ctx context.Context, req *thresholddomain.CreateRequest) (*auto.XpackThresholdInfo, error) {
	if s.repo == nil {
		return nil, errThresholdRepoNotReady
	}
	if req == nil || req.ID <= 0 {
		return nil, gorm.ErrInvalidData
	}
	if err := validateThresholdCreateRequest(req); err != nil {
		return nil, err
	}

	current, err := s.repo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	updated := buildThresholdInfo(req)
	current.Name = updated.Name
	current.Enable = updated.Enable
	current.RateType = updated.RateType
	current.RateValue = updated.RateValue
	current.ResourceID = updated.ResourceID
	current.ResourceType = updated.ResourceType
	current.ChartID = updated.ChartID
	current.ChartType = updated.ChartType
	current.ThresholdRules = updated.ThresholdRules
	current.MsgType = updated.MsgType
	current.MsgTitle = updated.MsgTitle
	current.MsgContent = updated.MsgContent
	current.RepeatSend = updated.RepeatSend
	current.ReciUsers = updated.ReciUsers
	current.ReciRoles = updated.ReciRoles
	current.ReciEmails = updated.ReciEmails
	current.ReciLarkGroups = updated.ReciLarkGroups
	current.ReciLarksuiteGroups = updated.ReciLarksuiteGroups
	current.ReciWebhooks = updated.ReciWebhooks
	current.Recisetting = updated.Recisetting

	if err = s.repo.Update(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *ThresholdService) FormInfo(ctx context.Context, id int64, resourceTable string) (*thresholddomain.CreateRequest, error) {
	if isSnapshotResourceTable(resourceTable) {
		return &thresholddomain.CreateRequest{}, nil
	}
	if s.repo == nil {
		return nil, errThresholdRepoNotReady
	}

	info, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	result := &thresholddomain.CreateRequest{
		BaseReciDTO:    thresholddomain.BaseReciDTO{},
		ID:             info.ID,
		Name:           info.Name,
		Enable:         thresholdBoolPtr(info.Enable),
		RateType:       thresholdIntPtr(int(info.RateType)),
		RateValue:      info.RateValue,
		ResourceID:     info.ResourceID,
		ResourceType:   info.ResourceType,
		ChartID:        info.ChartID,
		ChartType:      info.ChartType,
		ThresholdRules: info.ThresholdRules,
		MsgType:        thresholdIntPtr(int(info.MsgType)),
		MsgTitle:       info.MsgTitle,
		MsgContent:     info.MsgContent,
		RepeatSend:     thresholdBoolPtr(info.RepeatSend),
		ShowFieldValue: thresholdBoolPtr(false),
		ResourceTable:  "core",
	}

	thresholdUnmarshalJSON(info.ReciUsers, &result.UIDList)
	thresholdUnmarshalJSON(info.ReciRoles, &result.RIDList)
	thresholdUnmarshalJSON(info.ReciEmails, &result.EmailList)
	thresholdUnmarshalJSON(info.ReciLarkGroups, &result.LarkGroupList)
	thresholdUnmarshalJSON(info.ReciLarksuiteGroups, &result.LarksuiteGroupList)
	thresholdUnmarshalJSON(info.ReciWebhooks, &result.WebhookList)
	thresholdUnmarshalJSON(info.Recisetting, &result.ReciFlagList)

	return result, nil
}

func (s *ThresholdService) SwitchEnable(ctx context.Context, req *thresholddomain.SwitchRequest) error {
	if req != nil && isSnapshotResourceTable(req.ResourceTable) {
		return nil
	}
	if s.repo == nil {
		return errThresholdRepoNotReady
	}
	if req == nil || req.ID <= 0 || req.Enable == nil {
		return gorm.ErrInvalidData
	}
	return s.repo.UpdateEnable(ctx, req.ID, *req.Enable)
}

func (s *ThresholdService) Delete(ctx context.Context, ids []int64, resourceTable string) error {
	if isSnapshotResourceTable(resourceTable) {
		return nil
	}
	if s.repo == nil {
		return errThresholdRepoNotReady
	}
	return s.repo.DeleteByIDs(ctx, ids)
}

func (s *ThresholdService) DeleteWithChart(ctx context.Context, chartID int64, resourceTable string) error {
	if isSnapshotResourceTable(resourceTable) {
		return nil
	}
	if s.repo == nil {
		return errThresholdRepoNotReady
	}
	return s.repo.DeleteByChartID(ctx, chartID)
}

func (s *ThresholdService) BatchReci(ctx context.Context, req *thresholddomain.BatchReciRequest) error {
	if s.repo == nil {
		return errThresholdRepoNotReady
	}
	if req == nil {
		return gorm.ErrInvalidData
	}
	if len(req.IDList) == 0 {
		return nil
	}
	return s.repo.UpdateRecipients(
		ctx,
		req.IDList,
		thresholdMarshalJSON(req.UIDList),
		thresholdMarshalJSON(req.RIDList),
		thresholdMarshalJSON(req.EmailList),
		thresholdMarshalJSON(req.LarkGroupList),
		thresholdMarshalJSON(req.LarksuiteGroupList),
		thresholdMarshalJSON(req.WebhookList),
	)
}

func (s *ThresholdService) Pager(ctx context.Context, req *thresholddomain.GridRequest, goPage, pageSize int) (*thresholddomain.PageResult, error) {
	if s.repo == nil {
		return nil, errThresholdRepoNotReady
	}
	rows, total, err := s.repo.Pager(ctx, req, goPage, pageSize)
	if err != nil {
		return nil, err
	}
	return &thresholddomain.PageResult{List: rows, Total: total, Current: goPage, Size: pageSize}, nil
}

func (s *ThresholdService) AnyThreshold(ctx context.Context, chartID int64, resourceTable string) (bool, error) {
	if isSnapshotResourceTable(resourceTable) {
		return false, nil
	}
	if s.repo == nil {
		return false, errThresholdRepoNotReady
	}
	return s.repo.ExistsByChartID(ctx, chartID)
}

func (s *ThresholdService) InstancePager(ctx context.Context, req *thresholddomain.InstanceRequest, goPage, pageSize int) (*thresholddomain.PageResult, error) {
	if s.repo == nil {
		return nil, errThresholdRepoNotReady
	}
	rows, total, err := s.repo.InstancePager(ctx, req, goPage, pageSize)
	if err != nil {
		return nil, err
	}
	return &thresholddomain.PageResult{List: rows, Total: total, Current: goPage, Size: pageSize}, nil
}

func (s *ThresholdService) Preview(ctx context.Context, req *thresholddomain.PreviewRequest) (string, error) {
	return "", errors.New("preview requires chart data access - not yet implemented")
}

func validateThresholdCreateRequest(req *thresholddomain.CreateRequest) error {
	if req == nil {
		return gorm.ErrInvalidData
	}
	if strings.TrimSpace(req.Name) == "" || req.ChartID <= 0 || req.ResourceID <= 0 || strings.TrimSpace(req.ThresholdRules) == "" {
		return gorm.ErrInvalidData
	}
	return nil
}

func buildThresholdInfo(req *thresholddomain.CreateRequest) *auto.XpackThresholdInfo {
	enable := true
	if req.Enable != nil {
		enable = *req.Enable
	}
	rateType := 1
	if req.RateType != nil {
		rateType = *req.RateType
	}
	msgType := 0
	if req.MsgType != nil {
		msgType = *req.MsgType
	}
	repeatSend := true
	if req.RepeatSend != nil {
		repeatSend = *req.RepeatSend
	}

	return &auto.XpackThresholdInfo{
		ID:                  req.ID,
		Name:                strings.TrimSpace(req.Name),
		Enable:              enable,
		RateType:            int32(rateType),
		RateValue:           req.RateValue,
		ResourceType:        req.ResourceType,
		ResourceID:          req.ResourceID,
		ChartType:           req.ChartType,
		ChartID:             req.ChartID,
		ThresholdRules:      req.ThresholdRules,
		Recisetting:         thresholdMarshalJSON(req.ReciFlagList),
		ReciUsers:           thresholdMarshalJSON(req.UIDList),
		ReciRoles:           thresholdMarshalJSON(req.RIDList),
		ReciEmails:          thresholdMarshalJSON(req.EmailList),
		ReciLarkGroups:      thresholdMarshalJSON(req.LarkGroupList),
		ReciLarksuiteGroups: thresholdMarshalJSON(req.LarksuiteGroupList),
		ReciWebhooks:        thresholdMarshalJSON(req.WebhookList),
		MsgTitle:            req.MsgTitle,
		MsgType:             int32(msgType),
		MsgContent:          req.MsgContent,
		RepeatSend:          repeatSend,
	}
}

func isSnapshotResourceTable(resourceTable string) bool {
	return strings.EqualFold(strings.TrimSpace(resourceTable), "snapshot")
}

func thresholdMarshalJSON(v any) string {
	if v == nil {
		return ""
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

func thresholdUnmarshalJSON(s string, target any) {
	if s == "" {
		return
	}
	_ = json.Unmarshal([]byte(s), target)
}

func thresholdBoolPtr(b bool) *bool { return &b }

func thresholdIntPtr(i int) *int { return &i }
