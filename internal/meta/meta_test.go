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

	scwConfigigFile := `
	access_key: SCWXXXXXXXXXXXXXXXXX
	secret_key: 866F4A9A-D058-4D3C-A39F-86930849CCC0
	`

	dir := t.TempDir()

	err := os.WriteFile(path.Join(dir, "config.yaml"), []byte(scwConfigigFile), 0o644)
	require.NoError(t, err)

	t.Setenv("SCW_CONFIG_FILE", path.Join(dir, "config.yaml"))
	t.Setenv("SCW_ACCESS_KEY", "SCWXXXXXXXXXXXXXXXXX")

	m, err := meta.NewMeta(ctx, cfg)
	require.NoError(t, err)

	ok, message := m.HasMultipleVariableSources()
	assert.True(t, ok)

	expectedMessage := `Variable	AvailableSources					Using
SCW_ACCESS_KEY	Active Profile in config.yaml, Environment variable	Environment variable
`
	assert.Equal(t, expectedMessage, message)
}
