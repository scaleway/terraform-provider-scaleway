package documentdb_test

import (
	"testing"

	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/documentdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ResourceDatabaseName(t *testing.T) {
	localizedInstanceID, databaseName, err := documentdb.ResourceDocumentDBDatabaseName("fr-par/uuid/name")
	require.NoError(t, err)
	assert.Equal(t, "fr-par/uuid", localizedInstanceID)
	assert.Equal(t, "name", databaseName)

	_, _, err = documentdb.ResourceDocumentDBDatabaseName("fr-par/uuid")
	require.Error(t, err)
}
