package dependencycheck

import (
	"context"
	"fmt"
	"testing"

	"github.com/asecurityteam/ipam-facade/pkg/domain"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDepCheckHandleSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dependencyChecker1 := NewMockDependencyCheck(ctrl)
	dependencyChecker1.EXPECT().CheckDependencies(context.Background()).Return(nil)
	dependencyChecker2 := NewMockDependencyCheck(ctrl)
	dependencyChecker2.EXPECT().CheckDependencies(context.Background()).Return(nil)
	multi := &MultiDependencyCheck{
		DependencyCheckList: []domain.DependencyCheck{dependencyChecker1, dependencyChecker2},
	}
	err := multi.CheckDependencies(context.Background())
	assert.Nil(t, err)
}

func TestDepCheckHandleError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dependencyChecker1 := NewMockDependencyCheck(ctrl)
	dependencyChecker1.EXPECT().CheckDependencies(context.Background()).Return(nil)
	dependencyChecker2 := NewMockDependencyCheck(ctrl)
	dependencyChecker2.EXPECT().CheckDependencies(context.Background()).Return(fmt.Errorf("error"))
	multi := &MultiDependencyCheck{
		DependencyCheckList: []domain.DependencyCheck{dependencyChecker1, dependencyChecker2},
	}
	err := multi.CheckDependencies(context.Background())
	assert.NotNil(t, err)
}
