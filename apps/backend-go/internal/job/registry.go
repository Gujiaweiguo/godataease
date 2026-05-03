package scheduler

import (
	"context"
	"fmt"
	"strings"
)

type RunFunc func(context.Context) error

type Metadata struct {
	Key         string
	Spec        string
	Description string
	Enabled     bool
	Distributed bool
}

type Definition struct {
	Metadata Metadata
	Run      RunFunc
}

type Registry struct {
	jobs []Definition
}

func NewRegistry() *Registry {
	return &Registry{jobs: make([]Definition, 0)}
}

func (r *Registry) Register(def Definition) error {
	if err := validateDefinition(def); err != nil {
		return err
	}
	for _, existing := range r.jobs {
		if existing.Metadata.Key == def.Metadata.Key {
			return fmt.Errorf("job key %s already registered", def.Metadata.Key)
		}
	}
	r.jobs = append(r.jobs, def)
	return nil
}

func (r *Registry) Jobs() []Definition {
	if r == nil {
		return nil
	}
	jobs := make([]Definition, len(r.jobs))
	copy(jobs, r.jobs)
	return jobs
}

func (r *Registry) EnabledJobs() []Definition {
	if r == nil {
		return nil
	}
	enabled := make([]Definition, 0, len(r.jobs))
	for _, job := range r.jobs {
		if job.Metadata.Enabled {
			enabled = append(enabled, job)
		}
	}
	return enabled
}

func validateDefinition(def Definition) error {
	if strings.TrimSpace(def.Metadata.Key) == "" {
		return fmt.Errorf("job key is required")
	}
	if strings.TrimSpace(def.Metadata.Spec) == "" {
		return fmt.Errorf("job spec is required")
	}
	if strings.TrimSpace(def.Metadata.Description) == "" {
		return fmt.Errorf("job description is required")
	}
	if def.Run == nil {
		return fmt.Errorf("job run function is required")
	}
	return nil
}
