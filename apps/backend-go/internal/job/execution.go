package scheduler

import (
	"context"
	"fmt"
	"time"

	applogger "dataease/backend/internal/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Outcome string

const (
	OutcomeSuccess Outcome = "success"
	OutcomeSkipped Outcome = "skipped"
	OutcomeFailed  Outcome = "failed"
)

const defaultDistributedLockTTL = 30 * time.Second

const releaseLockScript = `if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("del", KEYS[1]) else return 0 end`

type Result struct {
	JobKey      string
	Spec        string
	Description string
	Outcome     Outcome
	Distributed bool
	StartedAt   time.Time
	FinishedAt  time.Time
	Duration    time.Duration
	Err         error
}

type Reporter func(Result)

func DefaultReporter(result Result) {
	fields := []zap.Field{
		zap.String("job_key", result.JobKey),
		zap.String("job_spec", result.Spec),
		zap.String("job_description", result.Description),
		zap.String("outcome", string(result.Outcome)),
		zap.Bool("distributed", result.Distributed),
		zap.Time("started_at", result.StartedAt),
		zap.Time("finished_at", result.FinishedAt),
		zap.Duration("duration", result.Duration),
	}
	if result.Err != nil {
		fields = append(fields, zap.String("error", result.Err.Error()))
	}

	switch result.Outcome {
	case OutcomeSuccess:
		applogger.Info("Scheduled job completed", fields...)
	case OutcomeSkipped:
		applogger.Info("Scheduled job skipped", fields...)
	default:
		applogger.Error("Scheduled job failed", fields...)
	}
}

func (s *Scheduler) AddDefinition(def Definition, reporter Reporter) error {
	if err := validateDefinition(def); err != nil {
		return err
	}
	if !def.Metadata.Enabled {
		return nil
	}

	_, err := s.cron.AddFunc(def.Metadata.Spec, func() {
		s.runDefinition(context.Background(), def, reporter)
	})
	return err
}

func (s *Scheduler) AddRegistry(registry *Registry, reporter Reporter) error {
	if registry == nil {
		return nil
	}
	for _, def := range registry.EnabledJobs() {
		if err := s.AddDefinition(def, reporter); err != nil {
			return err
		}
	}
	return nil
}

func (s *Scheduler) runDefinition(ctx context.Context, def Definition, reporter Reporter) (result Result) {
	if reporter == nil {
		reporter = DefaultReporter
	}

	result = Result{
		JobKey:      def.Metadata.Key,
		Spec:        def.Metadata.Spec,
		Description: def.Metadata.Description,
		Distributed: def.Metadata.Distributed,
		StartedAt:   time.Now(),
	}

	finalize := func() Result {
		result.FinishedAt = time.Now()
		result.Duration = result.FinishedAt.Sub(result.StartedAt)
		reporter(result)
		return result
	}

	if def.Metadata.Distributed && s.redis != nil {
		lockKey := s.prefix + def.Metadata.Key + ":lock"
		lockOwner := uuid.NewString()
		acquired, err := s.redis.SetNX(ctx, lockKey, lockOwner, defaultDistributedLockTTL).Result()
		if err != nil {
			result.Outcome = OutcomeFailed
			result.Err = fmt.Errorf("acquire distributed lock: %w", err)
			return finalize()
		}
		if !acquired {
			result.Outcome = OutcomeSkipped
			return finalize()
		}
		defer func() {
			if err := releaseDistributedLock(ctx, s, lockKey, lockOwner); err != nil {
				result.Outcome = OutcomeFailed
				if result.Err == nil {
					result.Err = fmt.Errorf("release distributed lock: %w", err)
				}
			}
		}()
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			result.Outcome = OutcomeFailed
			result.Err = fmt.Errorf("panic executing scheduled job: %v", recovered)
			finalize()
		}
	}()

	if err := def.Run(ctx); err != nil {
		result.Outcome = OutcomeFailed
		result.Err = err
		return finalize()
	}

	result.Outcome = OutcomeSuccess
	return finalize()
}

func releaseDistributedLock(ctx context.Context, s *Scheduler, lockKey, owner string) error {
	if s == nil || s.redis == nil {
		return nil
	}
	return s.redis.Eval(ctx, releaseLockScript, []string{lockKey}, owner).Err()
}
