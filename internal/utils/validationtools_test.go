package utils

import (
"errors"
"github.com/ca-gip/kubi/pkg/apis/ca-gip/v1"
"github.com/stretchr/testify/assert"
"testing"
)

func TestCheckMandatoryParametersArePresents(t *testing.T) {

	project := &v1.Project{
		Spec: v1.ProjectSpec{
			Tenant:      "tenantTest",
			Environment: "",
			Project:     "projectTest",
			Stages:      []string{"stages"},
		},
	}
	errorExpected := errors.New("")
	errorExpected = nil
	error := CheckMandatoryParameters(project)

	assert.Equal(t, errorExpected, error)

}

func TestCheckMandatoryParametersMissTenant(t *testing.T) {

	project := &v1.Project{
		Spec: v1.ProjectSpec{
			Environment: "",
			Project:     "projectTest",
			Stages:      []string{"stages"},
		},
	}
	errorExpected := errors.New("Spec.Tenant empty.")
	error := CheckMandatoryParameters(project)

	assert.Equal(t, errorExpected, error)

}

func TestCheckMandatoryParametersMissStages(t *testing.T) {

	project := &v1.Project{
		Spec: v1.ProjectSpec{
			Tenant:      "tenantTest",
			Environment: "",
			Project:     "projectTest",
		},
	}
	errorExpected := errors.New("Spec.Stages empty.")
	error := CheckMandatoryParameters(project)

	assert.Equal(t, errorExpected, error)

}

func TestCheckMandatoryParametersMissProject(t *testing.T) {

	project := &v1.Project{
		Spec: v1.ProjectSpec{
			Tenant:      "tenantTest",
			Environment: "",
			Stages:      []string{"stages"},
		},
	}
	errorExpected := errors.New("Spec.Project empty.")
	error := CheckMandatoryParameters(project)

	assert.Equal(t, errorExpected, error)

}

