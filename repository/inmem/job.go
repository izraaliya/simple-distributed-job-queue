package inmemrepo

import (
	"context"
	"errors"
	"sync"

	"jobqueue/entity"
	_interface "jobqueue/interface"
)

type jobRepository struct {
	mu         sync.RWMutex
	inMemDb    map[string]*entity.Job
	tokenIndex map[string]string // token → jobID (idempotency)
}

// Save Job
func (t *jobRepository) Save(ctx context.Context, job *entity.Job) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.inMemDb[job.ID] = job
	if job.Token != nil {
		t.tokenIndex[*job.Token] = job.ID
	}
	return nil
}

// Update Job
func (t *jobRepository) Update(ctx context.Context, job *entity.Job) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.inMemDb[job.ID]; !exists {
		return errors.New("job not found")
	}
	t.inMemDb[job.ID] = job
	return nil
}

// Find Job By ID
func (t *jobRepository) FindByID(ctx context.Context, id string) (*entity.Job, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	job, exists := t.inMemDb[id]
	if !exists {
		return nil, errors.New("job not found")
	}
	return job, nil
}

// Find Job By Token (Idempotency)
func (t *jobRepository) FindByToken(ctx context.Context, token string) (*entity.Job, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	jobID, ok := t.tokenIndex[token]
	if !ok {
		return nil, errors.New("job not found")
	}
	return t.inMemDb[jobID], nil
}

// FindAll Job
func (t *jobRepository) FindAll(ctx context.Context) ([]*entity.Job, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var jobs []*entity.Job
	for _, job := range t.inMemDb {
		jobs = append(jobs, job)
	}
	return jobs, nil
}

// Get Job Status Summary
func (t *jobRepository) GetStatusSummary(ctx context.Context) entity.JobStatusSummary {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var summary entity.JobStatusSummary
	for _, job := range t.inMemDb {
		switch job.Status {
		case entity.StatusPending:
			summary.Pending++
		case entity.StatusRunning:
			summary.Running++
		case entity.StatusFailed:
			summary.Failed++
		case entity.StatusCompleted:
			summary.Completed++
		}
	}
	return summary
}


// Initiator ...
type Initiator func(s *jobRepository) *jobRepository

// NewJobRepository ...
func NewJobRepository() Initiator {
	return func(q *jobRepository) *jobRepository {
		q.inMemDb = make(map[string]*entity.Job)
		q.tokenIndex = make(map[string]string)
		return q
	}
}

// SetInMemConnection set database client connection
func (i Initiator) SetInMemConnection(inMemDb map[string]*entity.Job) Initiator {
	return func(s *jobRepository) *jobRepository {
		i(s).inMemDb = inMemDb
		return s
	}
}

// Build ...
func (i Initiator) Build() _interface.JobRepository {
	return i(&jobRepository{})
}
