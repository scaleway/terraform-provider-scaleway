package types_test

import (
	"strings"
	"testing"

	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestGetRandomName(t *testing.T) {
	name := types.NewRandomName("test")
	assert.True(t, strings.HasPrefix(name, "tf-test-"))
}
