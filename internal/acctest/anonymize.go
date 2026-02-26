package acctest

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// sensitiveKeys lists JSON keys whose values should be anonymized.
var sensitiveKeys = map[string]string{
	"api_key":       "00000000000000000000000000000000",
	"secret_key":    "00000000-0000-0000-0000-000000000000",
	"secret":        "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
	"token":         "xxxxxxxx-xxxx-xxxx-xxxxxxxxxxxxxxxx",
	"password":      "xxxxxxxx",
	"authorization": "Bearer xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
}

// headerKeys lists HTTP header names to anonymize (case-insensitive).
// Placeholder values used for cassette anonymization, not real credentials.
var headerKeys = map[string]string{ //nolint: gosec // G101: placeholder values for anonymization
	"x-auth-token":  "2b8d6113-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
	"authorization": "Bearer xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
}

// AnonymizeCassetteForTest anonymizes the cassette file for the given test when
// recording. Call via t.Cleanup() at the start of tests that record sensitive
// data (e.g. API keys). Runs after the test completes and the cassette is saved.
// If pkgFolder is empty, derives it from the caller's file location.
func AnonymizeCassetteForTest(t *testing.T, pkgFolder string) error {
	t.Helper()

	if pkgFolder == "" {
		_, f, _, ok := runtime.Caller(1)
		if !ok {
			return errors.New("cannot get caller for cassette path")
		}

		pkgFolder = filepath.Dir(f)
	}

	path := BuildCassetteName(t.Name(), pkgFolder, ".cassette") + ".yaml"

	return AnonymizeCassetteFile(path)
}

// AnonymizeCassetteFile anonymizes sensitive values in a cassette YAML file.
func AnonymizeCassetteFile(path string) error {
	data, err := os.ReadFile(path) //nolint: gosec // G304: path is from BuildCassetteName, not user input
	if err != nil {
		return err
	}

	var doc map[string]any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return err
	}

	interactions, _ := doc["interactions"].([]any)
	if interactions == nil {
		return nil
	}

	for _, it := range interactions {
		inter, _ := it.(map[string]any)
		if inter == nil {
			continue
		}

		if req, ok := inter["request"].(map[string]any); ok {
			anonymizeBodyInMap(req, "body")
			anonymizeHeadersInMap(req, "headers")
		}

		if resp, ok := inter["response"].(map[string]any); ok {
			anonymizeBodyInMap(resp, "body")
			anonymizeHeadersInMap(resp, "headers")
		}
	}

	out, err := yaml.Marshal(&doc)
	if err != nil {
		return err
	}

	return os.WriteFile(path, out, 0o600)
}

func anonymizeBodyInMap(m map[string]any, key string) {
	body, ok := m[key].(string)
	if !ok || body == "" {
		return
	}

	trimmed := strings.TrimSpace(body)
	if !strings.HasPrefix(trimmed, "{") && !strings.HasPrefix(trimmed, "[") {
		return
	}

	var v any
	if err := json.Unmarshal([]byte(body), &v); err != nil {
		return
	}

	if anonymizeJSON(v) {
		b, err := json.Marshal(v)
		if err != nil {
			return
		}

		m[key] = string(b)
	}
}

func anonymizeHeadersInMap(m map[string]any, key string) {
	headers, ok := m[key].(map[string]any)
	if !ok {
		return
	}

	for name, val := range headers {
		nameLower := strings.ToLower(name)
		if placeholder, ok := headerKeys[nameLower]; ok {
			if arr, ok := val.([]any); ok && len(arr) > 0 {
				if s, ok := arr[0].(string); ok && s != "" && s != placeholder {
					headers[name] = []any{placeholder}
				}
			}
		}
	}
}

func anonymizeJSON(v any) bool {
	modified := false

	switch x := v.(type) {
	case map[string]any:
		for key, val := range x {
			keyLower := strings.ToLower(key)
			if placeholder, ok := sensitiveKeys[keyLower]; ok {
				if s, ok := val.(string); ok && s != "" && s != placeholder {
					x[key] = placeholder
					modified = true
				}
			} else {
				if anonymizeJSON(val) {
					modified = true
				}
			}
		}
	case []any:
		for _, item := range x {
			if anonymizeJSON(item) {
				modified = true
			}
		}
	}

	return modified
}
