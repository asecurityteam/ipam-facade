package v1

import (
	"context"

	"github.com/asecurityteam/ipam-facade/pkg/domain"
)

// DependencyCheckHandler takes in a domain.DependencyChecker to check external dependencies
type DependencyCheckHandler struct {
	DependencyCheckList []domain.DependencyCheck
}

// Handle makes a call CheckDependencies from DependencyChecker that verifies this
// app can talk to it's external dependencies
func (h *DependencyCheckHandler) Handle(ctx context.Context) error {
	for _, dependencyCheck := range h.DependencyCheckList {
		err := dependencyCheck.CheckDependencies(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
