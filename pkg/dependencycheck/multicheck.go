package dependencycheck

import (
	"context"

	"github.com/asecurityteam/ipam-facade/pkg/domain"
)

// MultiDependencyCheck is an implementation of DependencyCheck that
// checks multiple dependencies
type MultiDependencyCheck struct {
	DependencyCheckList []domain.DependencyCheck
}

// CheckDependencies loops through DependencyCheckList and errors out on the first
// failure of a DependencyCheck
func (m *MultiDependencyCheck) CheckDependencies(ctx context.Context) error {

	for _, dependencyCheck := range m.DependencyCheckList {
		err := dependencyCheck.CheckDependencies(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
