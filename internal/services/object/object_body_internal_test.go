package object

import (
	"encoding/base64"
	"io"
	"maps"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readObjectBody(t *testing.T, d *schema.ResourceData) []byte {
	t.Helper()

	body, err := objectBody(d)
	require.NoError(t, err)

	if closer, ok := body.(io.Closer); ok {
		defer func() { require.NoError(t, closer.Close()) }()
	}

	data, err := io.ReadAll(body)
	require.NoError(t, err)

	return data
}

// TestObjectBodyContentBase64 checks that content_base64 is uploaded as exactly the decoded bytes, whatever the
// content's length: a buffer sized by DecodedLen is one or two bytes longer than the content when its length isn't a
// multiple of 3, and those bytes must not reach the object.
func TestObjectBodyContentBase64(t *testing.T) {
	for _, content := range []string{
		"",
		"a",
		"ab",
		"abc",
		"User-agent: *\nDisallow: /\n",
		"\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR",
	} {
		d := schema.TestResourceDataRaw(t, objectSchema(), map[string]any{
			"bucket":         "bucket",
			"key":            "key",
			"content_base64": base64.StdEncoding.EncodeToString([]byte(content)),
		})

		assert.Equal(t, []byte(content), readObjectBody(t, d), "content of %d bytes", len(content))
	}
}

func TestObjectBodyInvalidContentBase64(t *testing.T) {
	d := schema.TestResourceDataRaw(t, objectSchema(), map[string]any{
		"bucket":         "bucket",
		"key":            "key",
		"content_base64": "not base64",
	})

	_, err := objectBody(d)
	require.ErrorContains(t, err, "illegal base64 data at input byte 3")
}

func TestObjectBodyContentAndFile(t *testing.T) {
	d := schema.TestResourceDataRaw(t, objectSchema(), map[string]any{
		"bucket":  "bucket",
		"key":     "key",
		"content": "plain text",
	})
	assert.Equal(t, []byte("plain text"), readObjectBody(t, d))

	path := filepath.Join(t.TempDir(), "file.txt")
	require.NoError(t, os.WriteFile(path, []byte("from a file"), 0o600))

	d = schema.TestResourceDataRaw(t, objectSchema(), map[string]any{
		"bucket": "bucket",
		"key":    "key",
		"file":   path,
	})
	assert.Equal(t, []byte("from a file"), readObjectBody(t, d))

	d = schema.TestResourceDataRaw(t, objectSchema(), map[string]any{
		"bucket": "bucket",
		"key":    "key",
	})
	assert.Empty(t, readObjectBody(t, d))
}

// TestObjectBodyChanged checks that an update uploads the object again when its content changes, from whichever
// argument it comes, and only then.
func TestObjectBodyChanged(t *testing.T) {
	old := map[string]any{
		"bucket":         "bucket",
		"key":            "key",
		"content_base64": base64.StdEncoding.EncodeToString([]byte("old")),
		"hash":           "old-hash",
		"content_type":   "text/plain",
	}
	state := &terraform.InstanceState{ID: "fr-par/bucket/key", Attributes: map[string]string{}}

	for k, v := range old {
		state.Attributes[k] = v.(string)
	}

	for name, change := range map[string]struct {
		key, value string
		changed    bool
	}{
		"content_base64": {"content_base64", base64.StdEncoding.EncodeToString([]byte("new")), true},
		"hash":           {"hash", "new-hash", true},
		"content_type":   {"content_type", "application/json", false},
	} {
		config := maps.Clone(old)

		config[change.key] = change.value

		diff, err := ResourceObject().Diff(t.Context(), state, terraform.NewResourceConfigRaw(config), nil)
		require.NoError(t, err)

		d, err := schema.InternalMap(objectSchema()).Data(state, diff)
		require.NoError(t, err)
		assert.Equal(t, change.changed, objectBodyChanged(d), name)
	}
}
