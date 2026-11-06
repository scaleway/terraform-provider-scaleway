package dataorchestrator

import (
	"errors"
	"fmt"

	dataorchestrator "github.com/scaleway/scaleway-sdk-go/api/dataorchestrator/v1alpha2"
)

func currentWorkflowVersion(def *dataorchestrator.WorkflowDefinition) (*dataorchestrator.WorkflowVersion, error) {
	if def == nil {
		return nil, errors.New("workflow definition is nil")
	}

	var current *dataorchestrator.WorkflowVersion

	for _, version := range def.Versions {
		if version == nil {
			continue
		}

		if version.IsCurrent {
			return version, nil
		}

		if current == nil {
			current = version
		}
	}

	if current == nil {
		return nil, fmt.Errorf("workflow definition %s has no versions", def.ID)
	}

	if len(def.Versions) > 1 {
		return nil, fmt.Errorf("workflow definition %s has multiple versions and none marked current", def.ID)
	}

	return current, nil
}
