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
	// Always log at Info level to ensure visibility
	Log.Info().Msgf("IsV2Project: Project %s has APIVersion: '%s'", project.Name, project.APIVersion)

	// Check for v2 in multiple ways to be more robust
	isV2 := strings.Contains(project.APIVersion, "v2") ||
		strings.Contains(project.APIVersion, "V2") ||
		project.APIVersion == "cagip.github.com/v2"

	Log.Info().Msgf("IsV2Project: Project %s is v2: %v", project.Name, isV2)

	return isV2
}
