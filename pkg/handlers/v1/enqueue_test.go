package v1

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestEnqueue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jobID := "f613056d-9b3e-4d69-888f-9f56c1ee8093"
	mockRandomNumberGenerator := NewMockRandomNumberGenerator(ctrl)
	mockRandomNumberGenerator.EXPECT().NewRandom().Return(uuid.MustParse(jobID), nil)
	mockProducer := NewMockProducer(ctrl)
	mockProducer.EXPECT().Produce(gomock.Any(), JobMetadata{JobID: jobID}).Return(JobMetadata{JobID: jobID}, nil)

	h := &EnqueueHandler{
		RandomNumberGenerator: mockRandomNumberGenerator,
		Producer:              mockProducer,
		LogFn:                 testLogFn,
	}
	resp, err := h.Handle(context.Background())
	assert.Equal(t, JobMetadata{JobID: jobID}, resp)
	assert.Nil(t, err)
}

func TestEnqueueProducerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jobID := "f613056d-9b3e-4d69-888f-9f56c1ee8093"
	mockRandomNumberGenerator := NewMockRandomNumberGenerator(ctrl)
	mockRandomNumberGenerator.EXPECT().NewRandom().Return(uuid.MustParse(jobID), nil)
	mockProducer := NewMockProducer(ctrl)
	mockProducer.EXPECT().Produce(gomock.Any(), JobMetadata{JobID: jobID}).Return(uuid.New(), errors.New(""))

	h := &EnqueueHandler{
		RandomNumberGenerator: mockRandomNumberGenerator,
		Producer:              mockProducer,
		LogFn:                 testLogFn,
	}
	_, err := h.Handle(context.Background())
	assert.Error(t, err)
}

func TestEnqueueUUIDGenerationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRandomNumberGenerator := NewMockRandomNumberGenerator(ctrl)
	mockRandomNumberGenerator.EXPECT().NewRandom().Return(uuid.New(), errors.New(""))
	mockProducer := NewMockProducer(ctrl)

	h := &EnqueueHandler{
		RandomNumberGenerator: mockRandomNumberGenerator,
		Producer:              mockProducer,
		LogFn:                 testLogFn,
	}
	_, err := h.Handle(context.Background())
	assert.Error(t, err)
}
