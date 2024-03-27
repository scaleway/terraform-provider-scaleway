package scaleway_test

import (
	"testing"

	"github.com/scaleway/terraform-provider-scaleway/v2/scaleway"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_resourceScalewayDocumentDBDatabaseName(t *testing.T) {
	localizedInstanceID, databaseName, err := scaleway.ResourceScalewayDocumentDBDatabaseName("fr-par/uuid/name")
	require.NoError(t, err)
	assert.Equal(t, "fr-par/uuid", localizedInstanceID)
	assert.Equal(t, "name", databaseName)

	_, _, err = scaleway.ResourceScalewayDocumentDBDatabaseName("fr-par/uuid")
	require.Error(t, err)
}
