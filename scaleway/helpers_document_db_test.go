package scaleway

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_resourceScalewayDocumentDBDatabaseName(t *testing.T) {
	localizedInstanceID, databaseName, err := resourceScalewayDocumentDBDatabaseName("fr-par/uuid/name")
	require.NoError(t, err)
	assert.Equal(t, "fr-par/uuid", localizedInstanceID)
	assert.Equal(t, "name", databaseName)

	_, _, err = resourceScalewayDocumentDBDatabaseName("fr-par/uuid")
	require.Error(t, err)
}
