package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_validateClusterLocation_ValidValues(t *testing.T) {
	for _, value := range AllowedClusterLocations {
		result, err := validateClusterLocation(value)

		assert.Equal(t, value, result)
		assert.Nil(t, err)
	}
}

func Test_validateClusterLocation_InvalidValue(t *testing.T) {
	result, err := validateClusterLocation("ThisIsNotAClusterLocation")

	assert.Empty(t, result)
	assert.Error(t, err)
}
