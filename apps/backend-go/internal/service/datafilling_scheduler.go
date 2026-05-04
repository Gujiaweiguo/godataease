package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	datafillingdomain "dataease/backend/internal/domain/datafilling"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"dataease/backend/internal/pkg/logger"
)

const (
	dataFillingExecStatusFailed  = 0
	dataFillingExecStatusSuccess = 1
)

var scheduleNumberPattern = regexp.MustCompile(`\d+`)

type DataFillingScheduler struct {
	cron            *cron.Cron
	entries         map[int64]cron.EntryID
	mu              sync.RWMutex
	taskRepo        TaskRepository
	subTaskRepo     SubTaskRepository
	subInstanceRepo SubInstanceRepository
	formRepo        DataFillingRepo
}

func NewDataFillingScheduler(taskRepo TaskRepository, subTaskRepo SubTaskRepository, subInstanceRepo SubInstanceRepository, formRepo DataFillingRepo) *DataFillingScheduler {
	return &DataFillingScheduler{
		cron:            cron.New(cron.WithSeconds()),
		entries:         make(map[int64]cron.EntryID),
		taskRepo:        taskRepo,
		subTaskRepo:     subTaskRepo,
		subInstanceRepo: subInstanceRepo,
		formRepo:        formRepo,
	}
}

func (s *DataFillingScheduler) Start(ctx context.Context) {
	if s == nil || s.cron == nil {
		return
	}
	s.cron.Start()
	if err := s.loadAndRegisterStartedTasks(ctx); err != nil {
		logger.Warn("failed to start data filling scheduler", zap.Error(err))
	}
}

func (s *DataFillingScheduler) Stop() {
	if s == nil || s.cron == nil {
		return
	}
	ctx := s.cron.Stop()
	select {
	case <-ctx.Done():
	default:
	}
}

func (s *DataFillingScheduler) loadAndRegisterStartedTasks(ctx context.Context) error {
	if s.taskRepo == nil {
		return nil
	}
	tasks, err := s.taskRepo.GetStartedTasks(ctx)
	if err != nil {
		return err
	}
	now := time.Now().UnixMilli()
	for _, task := range tasks {
		if task == nil {
			continue
		}
		if err := s.RegisterTask(ctx, task.ID); err != nil {
			logger.Warn("failed to register data filling task", zap.Int64("taskID", task.ID), zap.Error(err))
			continue
		}
		if task.NextExecTime > 0 && task.NextExecTime <= now {
			if err := s.FireTask(ctx, task.ID); err != nil {
				logger.Warn("failed to fire overdue data filling task", zap.Int64("taskID", task.ID), zap.Error(err))
			}
		}
	}
	return nil
}

func (s *DataFillingScheduler) RegisterTask(ctx context.Context, taskID int64) error {
	if s == nil || s.taskRepo == nil {
		return fmt.Errorf("task scheduler not configured")
	}
	task, err := s.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return err
	}
	spec, err := buildTaskCronSpec(task)
	if err != nil {
		return err
	}
	s.UnregisterTask(taskID)
	entryID, err := s.cron.AddFunc(spec, func() {
		if err := s.FireTask(context.Background(), taskID); err != nil {
			logger.Warn("data filling task execution failed", zap.Int64("taskID", taskID), zap.Error(err))
		}
	})
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.entries[taskID] = entryID
	s.mu.Unlock()
	return nil
}

func (s *DataFillingScheduler) UnregisterTask(taskID int64) {
	if s == nil || s.cron == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	entryID, ok := s.entries[taskID]
	if !ok {
		return
	}
	s.cron.Remove(entryID)
	delete(s.entries, taskID)
}

func (s *DataFillingScheduler) validateFireTaskPreconditions(ctx context.Context, taskID int64) (*datafillingdomain.DataFillingTask, error) {
	if s == nil || s.taskRepo == nil || s.subTaskRepo == nil || s.subInstanceRepo == nil {
		return nil, fmt.Errorf("task scheduler repositories not configured")
	}
	task, err := s.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task.FormID <= 0 {
		return nil, gorm.ErrInvalidData
	}
	if s.formRepo != nil {
		if _, err := s.formRepo.GetByID(ctx, task.FormID); err != nil {
			return nil, err
		}
	}
	return task, nil
}

func (s *DataFillingScheduler) resolveUserIDs(ctx context.Context, task *datafillingdomain.DataFillingTask) ([]int64, error) {
	uidList, err := parseJSONInt64List(task.UIDList)
	if err != nil {
		return nil, err
	}
	ridList, err := parseJSONInt64List(task.RIDList)
	if err != nil {
		return nil, err
	}
	return s.resolveRecipients(ctx, uidList, ridList)
}

func buildSubInstances(taskID, subTaskID, formID int64, userIDs []int64) []*datafillingdomain.DataFillingSubInstance {
	instances := make([]*datafillingdomain.DataFillingSubInstance, 0, len(userIDs))
	for _, uid := range userIDs {
		instances = append(instances, &datafillingdomain.DataFillingSubInstance{
			TaskID: taskID,
			PID:    subTaskID,
			UID:    uid,
			FormID: formID,
			Status: datafillingdomain.SubInstanceStatusOpen,
		})
	}
	return instances
}

func (s *DataFillingScheduler) FireTask(ctx context.Context, taskID int64) error {
	task, err := s.validateFireTaskPreconditions(ctx, taskID)
	if err != nil {
		return err
	}
	userIDs, err := s.resolveUserIDs(ctx, task)
	if err != nil {
		return err
	}
	now := time.Now()
	subTask := &datafillingdomain.DataFillingSubTask{
		TaskID:              task.ID,
		StartTime:           task.StartTime,
		EndTime:             task.EndTime,
		ExecStatus:          dataFillingExecStatusSuccess,
		Status:              computeSubTaskStatus(task, now),
		TotalCount:          len(userIDs),
		UnfinishedCount:     len(userIDs),
		TotalUserCount:      len(userIDs),
		UnfinishedUserCount: len(userIDs),
		FillType:            task.FillType,
	}
	if err := s.subTaskRepo.CreateSubTask(ctx, subTask); err != nil {
		return err
	}
	instances := buildSubInstances(task.ID, subTask.ID, task.FormID, userIDs)
	if err := s.subInstanceRepo.BatchCreateSubInstances(ctx, instances); err != nil {
		return err
	}
	if err := s.subTaskRepo.UpdateSubTaskCounts(ctx, subTask.ID, len(instances), len(instances), len(userIDs), len(userIDs)); err != nil {
		return err
	}
	nextExecTime, err := s.computeNextExecTime(task)
	if err != nil {
		return err
	}
	task.LastExecStatus = dataFillingExecStatusSuccess
	task.LastExecTime = now.UnixMilli()
	task.NextExecTime = nextExecTime
	return s.taskRepo.UpdateTask(ctx, task)
}

func (s *DataFillingScheduler) resolveRecipients(ctx context.Context, uidList, ridList []int64) ([]int64, error) {
	_ = ctx
	_ = ridList
	uniq := make(map[int64]struct{}, len(uidList))
	for _, uid := range uidList {
		if uid > 0 {
			uniq[uid] = struct{}{}
		}
	}
	result := make([]int64, 0, len(uniq))
	for uid := range uniq {
		result = append(result, uid)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result, nil
}

func (s *DataFillingScheduler) computeNextExecTime(task *datafillingdomain.DataFillingTask) (int64, error) {
	if task == nil {
		return 0, gorm.ErrInvalidData
	}
	next, err := computeTaskNextExecTime(task, time.Now())
	if err != nil {
		return 0, err
	}
	return next.UnixMilli(), nil
}

func computeTaskNextExecTime(task *datafillingdomain.DataFillingTask, now time.Time) (time.Time, error) {
	base := now
	if task.LastExecTime > 0 {
		base = time.UnixMilli(task.LastExecTime)
	}
	if task.RateType == 0 {
		parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
		parsed, err := parser.Parse(strings.TrimSpace(task.RateVal))
		if err != nil {
			return time.Time{}, err
		}
		return parsed.Next(base), nil
	}
	return computeSimpleNextTime(task.RateType, strings.TrimSpace(task.RateVal), now)
}

func buildTaskCronSpec(task *datafillingdomain.DataFillingTask) (string, error) {
	if task == nil {
		return "", gorm.ErrInvalidData
	}
	if task.RateType == 0 {
		return strings.TrimSpace(task.RateVal), nil
	}
	nums := extractScheduleNumbers(task.RateVal)
	switch task.RateType {
	case 1:
		hour, minute, second, err := extractTimeParts(nums)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d %d %d * * *", second, minute, hour), nil
	case 2:
		if len(nums) < 3 {
			return "", fmt.Errorf("invalid weekly rate value")
		}
		weekday, err := normalizeWeekday(nums[0])
		if err != nil {
			return "", err
		}
		hour, minute, second, err := extractTimeParts(nums[1:])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d %d %d * * %d", second, minute, hour, weekday), nil
	case 3:
		if len(nums) < 3 {
			return "", fmt.Errorf("invalid monthly rate value")
		}
		day := nums[0]
		if day < 1 || day > 31 {
			return "", fmt.Errorf("invalid monthly day")
		}
		hour, minute, second, err := extractTimeParts(nums[1:])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d %d %d %d * *", second, minute, hour, day), nil
	default:
		return "", fmt.Errorf("unsupported rate type: %d", task.RateType)
	}
}

func computeSimpleNextTime(rateType int, rateVal string, now time.Time) (time.Time, error) {
	nums := extractScheduleNumbers(rateVal)
	switch rateType {
	case 1:
		hour, minute, second, err := extractTimeParts(nums)
		if err != nil {
			return time.Time{}, err
		}
		candidate := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, second, 0, now.Location())
		if !candidate.After(now) {
			candidate = candidate.Add(24 * time.Hour)
		}
		return candidate, nil
	case 2:
		if len(nums) < 3 {
			return time.Time{}, fmt.Errorf("invalid weekly rate value")
		}
		weekday, err := normalizeWeekday(nums[0])
		if err != nil {
			return time.Time{}, err
		}
		hour, minute, second, err := extractTimeParts(nums[1:])
		if err != nil {
			return time.Time{}, err
		}
		daysUntil := (weekday - int(now.Weekday()) + 7) % 7
		candidate := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, second, 0, now.Location()).AddDate(0, 0, daysUntil)
		if !candidate.After(now) {
			candidate = candidate.AddDate(0, 0, 7)
		}
		return candidate, nil
	case 3:
		if len(nums) < 3 {
			return time.Time{}, fmt.Errorf("invalid monthly rate value")
		}
		day := nums[0]
		if day < 1 || day > 31 {
			return time.Time{}, fmt.Errorf("invalid monthly day")
		}
		hour, minute, second, err := extractTimeParts(nums[1:])
		if err != nil {
			return time.Time{}, err
		}
		candidate := clampMonthlyTime(now.Year(), now.Month(), day, hour, minute, second, now.Location())
		if !candidate.After(now) {
			nextMonth := now.AddDate(0, 1, 0)
			candidate = clampMonthlyTime(nextMonth.Year(), nextMonth.Month(), day, hour, minute, second, now.Location())
		}
		return candidate, nil
	default:
		return time.Time{}, fmt.Errorf("unsupported rate type: %d", rateType)
	}
}

func computeSubTaskStatus(task *datafillingdomain.DataFillingTask, now time.Time) int {
	if task == nil {
		return datafillingdomain.SubTaskStatusExpired
	}
	if task.EndTime > 0 && now.UnixMilli() > task.EndTime {
		return datafillingdomain.SubTaskStatusExpired
	}
	return datafillingdomain.SubTaskStatusActive
}

func parseJSONInt64List(raw string) ([]int64, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return []int64{}, nil
	}
	var values []int64
	if err := json.Unmarshal([]byte(trimmed), &values); err == nil {
		return values, nil
	}
	var ints []int
	if err := json.Unmarshal([]byte(trimmed), &ints); err == nil {
		values = make([]int64, 0, len(ints))
		for _, v := range ints {
			values = append(values, int64(v))
		}
		return values, nil
	}
	return nil, fmt.Errorf("parse int64 list failed")
}

func extractScheduleNumbers(raw string) []int {
	matches := scheduleNumberPattern.FindAllString(raw, -1)
	values := make([]int, 0, len(matches))
	for _, match := range matches {
		parsed, err := strconv.Atoi(match)
		if err == nil {
			values = append(values, parsed)
		}
	}
	return values
}

func extractTimeParts(nums []int) (int, int, int, error) {
	if len(nums) < 2 {
		return 0, 0, 0, fmt.Errorf("invalid time value")
	}
	hour, minute, second := nums[0], nums[1], 0
	if len(nums) > 2 {
		second = nums[2]
	}
	if hour < 0 || hour > 23 || minute < 0 || minute > 59 || second < 0 || second > 59 {
		return 0, 0, 0, fmt.Errorf("invalid time value")
	}
	return hour, minute, second, nil
}

func normalizeWeekday(day int) (int, error) {
	if day == 7 {
		return 0, nil
	}
	if day < 0 || day > 6 {
		return 0, fmt.Errorf("invalid weekday")
	}
	return day, nil
}

func clampMonthlyTime(year int, month time.Month, day, hour, minute, second int, location *time.Location) time.Time {
	lastDay := time.Date(year, month+1, 0, hour, minute, second, 0, location).Day()
	if day > lastDay {
		day = lastDay
	}
	return time.Date(year, month, day, hour, minute, second, 0, location)
}
