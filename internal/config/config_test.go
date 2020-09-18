package config

import (
	"errors"
	"fmt"
	"github.com/ca-gip/artifactory-operator/internal/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_validateClusterLocation_ValidValues(t *testing.T) {
	for _, value := range utils.AllowedClusterLocations {
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

func TestExtractLDAPCN(t *testing.T) {

	t.Run("CustomerOps group CN  should be returned in lowercase", func(t *testing.T) {
		//Prepare
		expectedCN := "dl_kub_caahp_ops"

		//Assert
		cn, _ := extractLDAPCN("CN=DL_KUB_CAAHP_OPS,OU=HPRD,OU=CAA,OU=PAAS_CONTAINER,OU=Applications,OU=Groupes,O=CA")

		//Test
		assert.Equal(t, expectedCN, cn)
	})

	t.Run("CustomerOps group CN should not be uppercase", func(t *testing.T) {
		//Prepare
		expectedCN := "DL_KUB_CAAHP_OPS"

		//Assert
		cn, _ := extractLDAPCN("CN=DL_KUB_CAAHP_OPS,OU=HPRD,OU=CAA,OU=PAAS_CONTAINER,OU=Applications,OU=Groupes,O=CA")

		//Test
		assert.NotEqual(t, expectedCN, cn)
	})

	t.Run("An error should be returned if CN does not exist in customer OPS GroupBase", func(t *testing.T) {
		//Prepare
		expectedCN := ""
		DN := "OU=HPRD,OU=CAA,OU=PAAS_CONTAINER,OU=Applications,OU=Groupes,O=CA"
		ExpectedError := errors.New(fmt.Sprintf("LDAP CN cannot be extracted from the DN: %s", DN))

		//Assert
		cn, err := extractLDAPCN("OU=HPRD,OU=CAA,OU=PAAS_CONTAINER,OU=Applications,OU=Groupes,O=CA")

		//Test
		assert.Equal(t, expectedCN, cn)
		assert.Equal(t, ExpectedError, err)

	})

	t.Run("ViewerOps group CN  should be returned in lowercase", func(t *testing.T) {
		//Prepare
		expectedCN := "dl_kub_caahp_view"

		//Assert
		cn, _ := extractLDAPCN("CN=DL_KUB_CAAHP_VIEW,OU=HPRD,OU=CAA,OU=PAAS_CONTAINER,OU=Applications,OU=Groupes,O=CA")

		//Test
		assert.Equal(t, expectedCN, cn)
	})


}
