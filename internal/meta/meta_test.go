package meta_test

import (
	"os"
	"path"
	"testing"

	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasMultipleVariableSources(t *testing.T) {
	ctx := t.Context()

	cfg := &meta.Config{}

	scwConfigigFile := `profiles:
  test:
    access_key: SCWXXXXXXXXXXXXXXXXX
    secret_key: 866F4A9A-D058-4D3C-A39F-86930849CCC0
    default_project_id: 866F4A9A-D058-4D3C-A39F-86930849CCC0
`

	dir := t.TempDir()

	err := os.WriteFile(path.Join(dir, "config.yaml"), []byte(scwConfigigFile), 0o644)
	require.NoError(t, err)

	t.Setenv("SCW_CONFIG_PATH", path.Join(dir, "config.yaml"))
	t.Setenv("SCW_PROFILE", "test")
	t.Setenv("SCW_ACCESS_KEY", "SCWXXXXXXXXXXXXXXXXX")
	t.Setenv("SCW_SECRET_KEY", "866F4A9A-D058-4D3C-A39F-86930849CCC0")
	t.Setenv("SCW_DEFAULT_PROJECT_ID", "866F4A9A-D058-4D3C-A39F-86930849CCC0")

	m, err := meta.NewMeta(ctx, cfg)
	require.NoError(t, err)

	ok, message, err := m.HasMultipleVariableSources()
	require.NoError(t, err)
	assert.True(t, ok)

	expectedMessage := `Variable		AvailableSources					Using
SCW_ACCESS_KEY		Active Profile in config.yaml, Environment variable	Environment variable
SCW_SECRET_KEY		Active Profile in config.yaml, Environment variable	Environment variable
SCW_DEFAULT_PROJECT_ID	Active Profile in config.yaml, Environment variable	Environment variable
`
	assert.Equal(t, expectedMessage, message)
}
