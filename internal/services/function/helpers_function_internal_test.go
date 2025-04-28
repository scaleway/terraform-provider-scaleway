package function

import (
	"testing"

	functionSDK "github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/stretchr/testify/assert"
)

func TestFilterSecretEnvsToPatch(t *testing.T) {
	testSecret := "test_secret"
	secretToDelete := "secret_to_delete"
	updatedSecret := "updated_secret"
	newSecret := "new_secret"

	oldEnv := []*functionSDK.Secret{
		{Key: testSecret, Value: &testSecret},
		{Key: secretToDelete, Value: &secretToDelete},
	}
	newEnv := []*functionSDK.Secret{
		{Key: testSecret, Value: &updatedSecret},
		{Key: newSecret, Value: &newSecret},
	}

	toPatch := filterSecretEnvsToPatch(oldEnv, newEnv)
	assert.Equal(t, []*functionSDK.Secret{
		{Key: testSecret, Value: &updatedSecret},
		{Key: newSecret, Value: &newSecret},
		{Key: secretToDelete, Value: nil},
	}, toPatch)
}
