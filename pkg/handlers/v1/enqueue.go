package v1

import (
	"context"

	producer "github.com/asecurityteam/component-producer"
	"github.com/asecurityteam/ipam-facade/pkg/domain"
	"github.com/asecurityteam/ipam-facade/pkg/logs"
)

// EnqueueHandler enqueues IPAM data sync requests
type EnqueueHandler struct {
	Producer      producer.Producer
	UUIDGenerator domain.UUIDGenerator
	LogFn         domain.LogFn
}

// Handle creates a job ID and enqueues the sync request with that ID
func (h *EnqueueHandler) Handle(ctx context.Context) (JobMetadata, error) {
	jobID, err := h.UUIDGenerator.NewUUIDString()
	if err != nil {
		h.LogFn(ctx).Error(logs.SyncError{Reason: err.Error()})
		return JobMetadata{}, err
	}
	jobMetadata := JobMetadata{JobID: jobID}

	_, err = h.Producer.Produce(ctx, jobMetadata)
	if err != nil {
		h.LogFn(ctx).Error(logs.ProducerError{Reason: err.Error()})
		return JobMetadata{}, err
	}

	return jobMetadata, nil
}
