package scaleway

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_resourceScalewayDocumentDBDatabaseName(t *testing.T) {
	localizedInstanceID, databaseName, err := resourceScalewayDocumentDBDatabaseName("fr-par/uuid/name")
	assert.Nil(t, err)
	assert.Equal(t, "fr-par/uuid", localizedInstanceID)
	assert.Equal(t, "name", databaseName)

	_, _, err = resourceScalewayDocumentDBDatabaseName("fr-par/uuid")
	assert.NotNil(t, err)
}
