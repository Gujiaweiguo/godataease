package scheduler

import (
	"context"
	"time"

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
	wrappedCmd := func() {
		if s.redis != nil {
			lockKey := s.prefix + name + ":lock"
			acquired, err := s.redis.SetNX(context.Background(), lockKey, "1", 30*time.Second).Result()
			if err != nil || !acquired {
				return
			}
			defer s.redis.Del(context.Background(), lockKey)
		}
		cmd()
	}
	return s.AddFunc(spec, wrappedCmd)
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
