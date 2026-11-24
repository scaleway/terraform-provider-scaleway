package secret_test

import (
	"encoding/base64"
	"reflect"
	"testing"

	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/secret"
)

func TestVersionV1SchemaUpgradeFunc(t *testing.T) {
	raw := "my-secret-payload"
	encoded := base64.StdEncoding.EncodeToString([]byte(raw))

	tests := map[string]struct {
		in   map[string]any
		want map[string]any
	}{
		"decodes legacy base64 data and adds type": {
			in:   map[string]any{"data": encoded, "revision": "1", "region": "fr-par"},
			want: map[string]any{"data": raw, "revision": "1", "region": "fr-par", "type": "string"},
		},
		"empty data is left untouched and type is added": {
			in:   map[string]any{"data": "", "revision": "1"},
			want: map[string]any{"data": "", "revision": "1", "type": "string"},
		},
		"absent data is left untouched and type is added": {
			in:   map[string]any{"revision": "1"},
			want: map[string]any{"revision": "1", "type": "string"},
		},
		"non-base64 data is left untouched and type is added": {
			in:   map[string]any{"data": "this is not base64", "revision": "1"},
			want: map[string]any{"data": "this is not base64", "revision": "1", "type": "string"},
		},
		"base64 data decoding to invalid UTF-8 is kept verbatim and marked base64": {
			in:   map[string]any{"data": "//4=", "revision": "1"},
			want: map[string]any{"data": "//4=", "revision": "1", "type": "base64"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := secret.VersionV1SchemaUpgradeFunc(t.Context(), tc.in, nil)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !reflect.DeepEqual(tc.want, got) {
				t.Fatalf("\nwant:\n%#v\ngot:\n%#v", tc.want, got)
			}
		})
	}
}
