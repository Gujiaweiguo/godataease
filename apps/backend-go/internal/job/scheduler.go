package scheduler

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron   *cron.Cron
	redis  *redis.Client
	prefix string
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		cron:   cron.New(cron.WithSeconds()),
		prefix: "dataease:scheduler:",
	}
}

func (s *Scheduler) SetRedis(client *redis.Client) {
	s.redis = client
}

func (s *Scheduler) AddFunc(spec string, cmd func()) error {
	_, err := s.cron.AddFunc(spec, cmd)
	return err
}

func (s *Scheduler) AddJob(spec string, job cron.Job) error {
	_, err := s.cron.AddJob(spec, job)
	return err
}

func (s *Scheduler) AddDistributedFunc(name, spec string, cmd func()) error {
	return s.AddDefinition(Definition{
		Metadata: Metadata{
			Key:         name,
			Spec:        spec,
			Description: name,
			Enabled:     true,
			Distributed: true,
		},
		Run: func(context.Context) error {
			cmd()
			return nil
		},
	}, nil)
}

func (s *Scheduler) Start() {
	s.cron.Start()
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
}

func (s *Scheduler) Remove(id cron.EntryID) {
	s.cron.Remove(id)
}

func (s *Scheduler) Entries() []cron.Entry {
	return s.cron.Entries()
}
