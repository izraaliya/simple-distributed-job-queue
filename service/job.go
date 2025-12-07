package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"jobqueue/entity"
	_interface "jobqueue/interface"
)

type jobService struct {
	jobRepo _interface.JobRepository
	queue   chan *entity.Job

	mu             sync.Mutex
	unstableCounts map[string]int
}


var idCounter int64

func generateJobID() string {
	counter := atomic.AddInt64(&idCounter, 1)
	return fmt.Sprintf("job-%d-%d", time.Now().UnixNano(), counter)
}

// Initiator ...
type Initiator func(s *jobService) *jobService


// GetAllJobs
func (q jobService) GetAllJobs(ctx context.Context) ([]*entity.Job, error) {
	return q.jobRepo.FindAll(ctx)
}

// GetJobByID
func (q jobService) GetJobByID(ctx context.Context, id string) (*entity.Job, error) {
	return q.jobRepo.FindByID(ctx, id)
}

// GetAllJobStatus
func (q jobService) GetAllJobStatus(ctx context.Context) entity.JobStatusSummary {
	jobs, _ := q.jobRepo.FindAll(ctx)
var status entity.JobStatusSummary

	for _, job := range jobs {
		switch job.Status {
		case entity.StatusPending:
			status.Pending++
		case entity.StatusRunning:
			status.Running++
		case entity.StatusFailed:
			status.Failed++
		case entity.StatusCompleted:
			status.Completed++
		}
	}
	return status
}

// Enqueue
func (q jobService) Enqueue(ctx context.Context, taskName string) (string, error) {
	job := &entity.Job{
		ID:          generateJobID(),
		Task:        taskName,
		Status:      entity.StatusPending,
		Attempts:    0,
		MaxAttempts: 3,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := q.jobRepo.Save(ctx, job); err != nil {
		return "", err
	}

	q.queue <- job
	log.Println("📥 enqueue job:", job.ID)

	return job.ID, nil
}


func (q *jobService) startWorkers(n int) {
	for i := 0; i < n; i++ {
		go q.worker(i)
	}
}

func (q *jobService) worker(id int) {
	log.Println("👷 worker started:", id)
	for job := range q.queue {
		q.processJob(job)
	}
}

func (q *jobService) processJob(job *entity.Job) {
	for job.Attempts < job.MaxAttempts {
		job.Attempts++
		job.Status = entity.StatusRunning
		job.UpdatedAt = time.Now()
		_ = q.jobRepo.Save(context.Background(), job)

		err := q.execute(job)
		if err == nil {
			job.Status = entity.StatusCompleted
			job.UpdatedAt = time.Now()
			_ = q.jobRepo.Save(context.Background(), job)
			log.Println("✅ job completed:", job.ID)
			return
		}

		job.Status = entity.StatusFailed
		job.UpdatedAt = time.Now()
		_ = q.jobRepo.Save(context.Background(), job)

		log.Println("❌ job failed:", job.ID, "attempt:", job.Attempts)
		time.Sleep(time.Duration(job.Attempts) * 300 * time.Millisecond)
	}
}

// unstable-job fails twice
func (q *jobService) execute(job *entity.Job) error {
	if job.Task == "unstable-job" {
		q.mu.Lock()
		q.unstableCounts[job.ID]++
		count := q.unstableCounts[job.ID]
		q.mu.Unlock()

		if count <= 2 {
			return errors.New("unstable job failed")
		}
	}
	return nil
}


// NewJobService
func NewJobService() Initiator {
	return func(s *jobService) *jobService {
		s.queue = make(chan *entity.Job, 200)
		s.unstableCounts = make(map[string]int)
		s.startWorkers(5)
		return s
	}
}

// SetJobRepository
func (i Initiator) SetJobRepository(jobRepository _interface.JobRepository) Initiator {
	return func(s *jobService) *jobService {
		i(s).jobRepo = jobRepository
		return s
	}
}

// Build
func (i Initiator) Build() _interface.JobService {
	return i(&jobService{})
}
