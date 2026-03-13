package acctest_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/stretchr/testify/require"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
)

// TestAccCassettes_NoSensitiveLeak ensures no sensitive data in cassette requests.
func TestAccCassettes_NoSensitiveLeak(t *testing.T) {
	t.Parallel()

	paths, err := getTestFiles(true)
	require.NoError(t, err)

	for path := range paths {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			c, err := cassette.Load(path)
			if err != nil {
				t.Skipf("cannot load cassette: %v", err)

				return
			}

			for i, inter := range c.Interactions {
				checkInteractionForLeaks(t, path, i, inter)
			}
		})
	}
}

func checkInteractionForLeaks(t *testing.T, cassettePath string, idx int, inter *cassette.Interaction) {
	t.Helper()

	checkBodyForLeaks(t, cassettePath, idx, "request", inter.Request.Body)
	checkHeadersForLeaks(t, cassettePath, idx, "request", inter.Request.Headers)
}

func checkBodyForLeaks(t *testing.T, cassettePath string, idx int, part, body string) {
	t.Helper()

	trimmed := strings.TrimSpace(body)
	if trimmed == "" || (!strings.HasPrefix(trimmed, "{") && !strings.HasPrefix(trimmed, "[")) {
		return
	}

	var v any
	if err := json.Unmarshal([]byte(body), &v); err != nil {
		return
	}

	checkJSONForLeaks(t, cassettePath, idx, part, v, "")
}

func checkJSONForLeaks(t *testing.T, cassettePath string, idx int, part string, v any, path string) {
	t.Helper()

	switch x := v.(type) {
	case map[string]any:
		for key, val := range x {
			keyLower := strings.ToLower(key)
			if placeholder, isSensitive := acctest.LeakCheckFields[keyLower]; isSensitive {
				if s, ok := val.(string); ok && s != "" {
					expected, _ := placeholder.(string)
					if s != expected {
						t.Errorf("cassette %s interaction %d: %s.body%s.%s contains potential leak (expected placeholder %q, got %q)",
							cassettePath, idx, part, path, key, expected, s)
					}
				}
			} else {
				checkJSONForLeaks(t, cassettePath, idx, part, val, path+"."+key)
			}
		}
	case []any:
		for i, item := range x {
			checkJSONForLeaks(t, cassettePath, idx, part, item, path+fmt.Sprintf("[%d]", i))
		}
	}
}

func checkHeadersForLeaks(t *testing.T, cassettePath string, idx int, part string, headers map[string][]string) {
	t.Helper()

	if headers == nil {
		return
	}

	for name, vals := range headers {
		nameLower := strings.ToLower(name)

		expected, ok := acctest.HeaderPlaceholders[nameLower]
		if !ok || len(vals) == 0 {
			continue
		}

		val := vals[0]
		if val != "" && val != expected {
			t.Errorf("cassette %s interaction %d: %s.headers.%s contains potential leak (expected placeholder %q, got %q)",
				cassettePath, idx, part, name, expected, val)
		}
	}
}
