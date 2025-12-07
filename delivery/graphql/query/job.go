package query

import (
	"context"

	_dataloader "jobqueue/delivery/graphql/dataloader"
	"jobqueue/delivery/graphql/resolver"
	_interface "jobqueue/interface"
)

type JobQuery struct {
	jobService _interface.JobService
	dataloader *_dataloader.GeneralDataloader
}

// GetAllJobs
func (q JobQuery) Jobs(ctx context.Context) ([]resolver.JobResolver, error) {
	jobs, err := q.jobService.GetAllJobs(ctx)
	if err != nil {
		return nil, err
	}

	resolvers := make([]resolver.JobResolver, 0, len(jobs))
	for _, job := range jobs {
		resolvers = append(resolvers, resolver.JobResolver{
			Data:       job,
			JobService: q.jobService,
			Dataloader: q.dataloader,
		})
	}
	return resolvers, nil
}

// GetJobById
func (q JobQuery) Job(ctx context.Context, args struct {
	ID string
}) (*resolver.JobResolver, error) {
	job, err := q.jobService.GetJobByID(ctx, args.ID)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, nil
	}

	return &resolver.JobResolver{
		Data:       job,
		JobService: q.jobService,
		Dataloader: q.dataloader,
	}, nil
}

// GetAllJobStatus
func (q JobQuery) JobStatus(ctx context.Context) (resolver.JobStatusResolver, error) {
	status := q.jobService.GetAllJobStatus(ctx)

	return resolver.JobStatusResolver{
		Data: status,
	}, nil
}

// constructor
func NewJobQuery(
	jobService _interface.JobService,
	dataloader *_dataloader.GeneralDataloader,
) JobQuery {
	return JobQuery{
		jobService: jobService,
		dataloader: dataloader,
	}
}
