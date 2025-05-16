package container

import (
	"testing"

	containerSDK "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/stretchr/testify/assert"
)

func TestFilterSecretEnvsToPatch(t *testing.T) {
	testSecret := "test_secret"
	secretToDelete := "secret_to_delete"
	updatedSecret := "updated_secret"
	newSecret := "new_secret"

	oldEnv := []*containerSDK.Secret{
		{Key: testSecret, Value: &testSecret},
		{Key: secretToDelete, Value: &secretToDelete},
	}
	newEnv := []*containerSDK.Secret{
		{Key: testSecret, Value: &updatedSecret},
		{Key: newSecret, Value: &newSecret},
	}

	toPatch := filterSecretEnvsToPatch(oldEnv, newEnv)
	assert.Equal(t, []*containerSDK.Secret{
		{Key: testSecret, Value: &updatedSecret},
		{Key: newSecret, Value: &newSecret},
		{Key: secretToDelete, Value: nil},
	}, toPatch)
}
