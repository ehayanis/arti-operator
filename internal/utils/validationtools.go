package utils

import (
	"errors"
	"github.com/ca-gip/kubi/pkg/apis/cagip/v1"
	"strings"
)

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
