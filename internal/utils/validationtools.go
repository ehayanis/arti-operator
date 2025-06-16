package utils

import (
	"errors"
	"strings"

	v1 "github.com/ca-gip/kubi/pkg/apis/cagip/v1"
)

// CheckMandatoryParameters validates that a v1 Project resource has the required fields
func CheckMandatoryParameters(project *v1.Project) error {
	if strings.TrimSpace(project.Spec.Project) == "" {
		return errors.New("Spec.Project empty.")
	}

	if strings.TrimSpace(project.Spec.Tenant) == "" {
		return errors.New("Spec.Tenant empty.")
	}

	if len(project.Spec.Stages) == 0 {
		return errors.New("Spec.Stages empty.")
	}

	return nil
}

// CheckMandatoryParametersV2 validates that a v2 Project resource has the required fields
// For v2, we only need tenant and project fields
func CheckMandatoryParametersV2(project *v1.Project) error {
	if strings.TrimSpace(project.Spec.Project) == "" {
		return errors.New("Spec.Project empty.")
	}

	if strings.TrimSpace(project.Spec.Tenant) == "" {
		return errors.New("Spec.Tenant empty.")
	}

	return nil
}

// IsV2Project determines if a Project resource is v2 based on its API version
func IsV2Project(project *v1.Project) bool {
	// Log the API version for debugging
	Log.Debug().Msgf("Project %s has APIVersion: '%s'", project.Name, project.APIVersion)

	// Check if the APIVersion contains "v2" (more flexible matching)
	return strings.Contains(project.APIVersion, "v2")
}
