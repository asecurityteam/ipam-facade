package v1

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDepCheckHandleSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dependencyChecker := NewMockDependencyCheck(ctrl)
	dependencyChecker.EXPECT().CheckDependencies(context.Background()).Return(nil)
	handler := &DependencyCheckHandler{
		DependencyChecker: dependencyChecker,
	}
	err := handler.Handle(context.Background())
	assert.Nil(t, err)
}

func TestDepCheckHandleError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dependencyChecker := NewMockDependencyCheck(ctrl)
	dependencyChecker.EXPECT().CheckDependencies(context.Background()).Return(fmt.Errorf("error"))
	handler := &DependencyCheckHandler{
		DependencyChecker: dependencyChecker,
	}
	err := handler.Handle(context.Background())
	assert.NotNil(t, err)
}
