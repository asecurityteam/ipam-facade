package v1

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestEnqueue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uuid := "f613056d-9b3e-4d69-888f-9f56c1ee8093"
	mockRandomNumberGenerator := NewMockRandomNumberGenerator(ctrl)
	mockRandomNumberGenerator.EXPECT().NewRandom().Return(uuid, nil)
	mockProducer := NewMockProducer(ctrl)
	mockProducer.EXPECT().Produce(gomock.Any(), jobMetadata).Return(nil, nil)

	h := &EnqueueHandler{
		RandomNumberGenerator: mockRandomNumberGenerator,
		Producer:              mockProducer,
		LogFn:                 testLogFn,
	}
	resp, err := h.Handle(context.Background())
	assert.Equal(&JobMetadata{JobID: uuid}, resp)
	assert.Nil(err)
}

func TestEnqueueProducerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uuid := "f613056d-9b3e-4d69-888f-9f56c1ee8093"
	mockRandomNumberGenerator := NewMockRandomNumberGenerator(ctrl)
	mockRandomNumberGenerator.EXPECT().NewRandom().Return(uuid, nil)
	mockProducer := NewMockProducer(ctrl)
	mockProducer.EXPECT().Produce(gomock.Any(), jobMetadata).Return(nil, errors.New(""))

	h := &EnqueueHandler{
		RandomNumberGenerator: mockRandomNumberGenerator,
		Producer:              mockProducer,
		LogFn:                 testLogFn,
	}
	assert.Error(t, h.Handle(context.Background()))
}

func TestEnqueueUUIDGenerationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRandomNumberGenerator := NewMockRandomNumberGenerator(ctrl)
	mockRandomNumberGenerator.EXPECT().NewRandom().Return(nil, errors.New(""))
	mockProducer := NewMockProducer(ctrl)
	mockProducer.EXPECT().Produce(gomock.Any(), jobMetadata).Return(nil, nil)

	h := &EnqueueHandler{
		RandomNumberGenerator: mockRandomNumberGenerator,
		Producer:              mockProducer,
		LogFn:                 testLogFn,
	}
	assert.Error(t, h.Handle(context.Background()))
}
