package jobs

import (
	"context"

	"dataease/backend/internal/app"
	scheduler "dataease/backend/internal/job"
	applogger "dataease/backend/internal/pkg/logger"

	"go.uber.org/zap"
)

const sampleHeartbeatJobKey = "scheduler-foundation-sample-heartbeat"

func NewRegistry(cfg app.SchedulerConfig) (*scheduler.Registry, error) {
	registry := scheduler.NewRegistry()
	if err := registry.Register(sampleHeartbeatDefinition(cfg.SampleJobEnabled)); err != nil {
		return nil, err
	}
	return registry, nil
}

func sampleHeartbeatDefinition(enabled bool) scheduler.Definition {
	return scheduler.Definition{
		Metadata: scheduler.Metadata{
			Key:         sampleHeartbeatJobKey,
			Spec:        "0 */5 * * * *",
			Description: "Low-risk heartbeat job used to validate centralized scheduler registration and execution diagnostics",
			Enabled:     enabled,
			Distributed: true,
		},
		Run: func(context.Context) error {
			applogger.Info("Scheduled job foundation heartbeat",
				zap.String("job_key", sampleHeartbeatJobKey),
			)
			return nil
		},
	}
}
