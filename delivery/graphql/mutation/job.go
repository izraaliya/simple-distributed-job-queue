package mutation

import (
	"context"

	_dataloader "jobqueue/delivery/graphql/dataloader"
	"jobqueue/delivery/graphql/resolver"
	_interface "jobqueue/interface"

	"jobqueue/entity"
)

type JobMutation struct {
	jobService _interface.JobService
	dataloader *_dataloader.GeneralDataloader
}

// SimultaneousCreateJob / Enqueue
func (q JobMutation) Enqueue(ctx context.Context, args entity.Job) (*resolver.JobResolver, error) {
	jobID, err := q.jobService.Enqueue(ctx, args.Task)
	if err != nil {
		return nil, err
	}

	job, err := q.jobService.GetJobByID(ctx, jobID)
	if err != nil {
		return nil, err
	}

	return &resolver.JobResolver{
		Data:       job,
		JobService: q.jobService,
		Dataloader: q.dataloader,
	}, nil
}

// constructor
func NewJobMutation(
	jobService _interface.JobService,
	dataloader *_dataloader.GeneralDataloader,
) JobMutation {
	return JobMutation{
		jobService: jobService,
		dataloader: dataloader,
	}
}
