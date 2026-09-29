package acctest

import (
	"net/url"
	"reflect"
	"slices"
	"sort"
	"strings"
)

// CompareJSONFields is the entry point for comparing two interface values
// handle string with special cases, map[string]interface{} and []interface{} or any other primitive type
func CompareJSONFields(requestValue, cassetteValue any, strict bool) bool {
	// Tolerate JSON null vs empty array [] mismatches caused by SDK
	// serialization drift (unset slices may serialize as either).
	if IsNilOrEmptySlice(requestValue) && IsNilOrEmptySlice(cassetteValue) {
		return true
	}

	// If either side is nil and they were not both nil/empty above, they are a
	// genuine mismatch. Returning here also guards the type assertions below
	// against panicking on a nil value.
	if requestValue == nil || cassetteValue == nil {
		return false
	}

	switch requestValue := requestValue.(type) {
	case string:
		return CompareFieldsStrings(requestValue, cassetteValue.(string))
	case map[string]any:
		return CompareJSONBodies(requestValue, cassetteValue.(map[string]any), strict)
	case []any:
		return CompareSlices(requestValue, cassetteValue.([]any))
	default:
		return reflect.DeepEqual(requestValue, cassetteValue)
	}
}

// IsNilOrEmptySlice reports whether v is nil or an empty []any.
// Used to tolerate JSON null vs [] mismatches caused by SDK serialization drift,
// where an unset slice may serialize as either null or [].
func IsNilOrEmptySlice(v any) bool {
	if v == nil {
		return true
	}

	s, ok := v.([]any)

	return ok && len(s) == 0
}

// CompareJSONBodies compare two given maps that represent json bodies
// returns true if both json are equivalent
func CompareJSONBodies(request, cassette map[string]any, strict bool) bool {
	for key, requestValue := range request {
		cassetteValue, ok := cassette[key]
		if !ok {
			if strict {
				return false
			}

			continue
		}

		// Tolerate JSON null vs empty array [] mismatches (see IsNilOrEmptySlice).
		// Without this, reflect.TypeOf reports different types (nil vs []any)
		// and the request never matches the recorded cassette interaction.
		if reflect.TypeOf(cassetteValue) != reflect.TypeOf(requestValue) &&
			(!IsNilOrEmptySlice(requestValue) || !IsNilOrEmptySlice(cassetteValue)) {
			return false
		}

		if !CompareJSONFields(requestValue, cassetteValue, strict) {
			return false
		}
	}

	for key, cassetteValue := range cassette {
		if _, ok := request[key]; !ok && cassetteValue != nil {
			// Fails match if cassettes contains a field not in actual requests
			// Fields should not disappear from requests unless a sdk breaking change
			// We ignore if field is nil in cassette as it could be an old deprecated and unused field
			return false
		}
	}

	return true
}

// compareFormBodies compare two given url.Values
// returns true if both url.Values are equivalent
func compareFormBodies(request, cassette url.Values) bool {
	// Check for each key in actual requests
	// Compare its value to cassette content if marshal-able to string
	for key := range request {
		requestValue, exists := request[key]
		if !exists {
			// Actual request may contain a field that does not exist in cassette
			// New fields can appear in requests with new api features
			// We do not want to generate new cassettes for each new features
			continue
		}

		if !CompareStringSlices(requestValue, cassette[key]) {
			return false
		}
	}

	for key, cassetteValue := range cassette {
		if _, exists := request[key]; !exists && cassetteValue != nil {
			// Fails match if cassettes contains a field not in actual requests
			// Fields should not disappear from requests unless a sdk breaking change
			// We ignore if field is nil in cassette as it could be an old deprecated and unused field
			return false
		}
	}

	return true
}

// CompareFieldsStrings compare two strings from request JSON bodies
// has special case when string are terraform generated names
func CompareFieldsStrings(expected, actual string) bool {
	if expected == actual {
		return true
	}

	// Action=DeleteTopic&TopicArn=arn%3Ascw%3Asns%3Afr-par%3Aproject-1a080a81-67b6-476d-80b4-f3bb9184e318%3Atest-mnq-sns-topic-basic20250603151943185500000004&Version=2010-03-31
	snsPrefix := "test-mnq-sns-topic-basic"
	if strings.HasPrefix(actual, snsPrefix) && strings.HasPrefix(expected, snsPrefix) {
		return true
	}

	if strings.HasPrefix(actual, "arn:scw:sns:") && strings.HasPrefix(expected, "arn:scw:sns:") {
		return true
	}

	expectedHandled := expected
	actualHandled := actual

	// Remove s3 url suffix to allow comparison
	if before, ok := strings.CutSuffix(actual, ".s3-website.fr-par.scw.cloud"); ok {
		actual = before
		expected = strings.TrimSuffix(expected, ".s3-website.fr-par.scw.cloud")
	}

	// Try to parse test generated name
	if strings.Contains(actual, "-") {
		expectedHandled = extractTestGeneratedNamePrefix(expected)
		actualHandled = extractTestGeneratedNamePrefix(actual)
	}

	// Try provider generated name
	if actualHandled == actual && strings.HasPrefix(actual, "tf-") {
		expectedHandled = extractGeneratedNamePrefix(expected)
		actualHandled = extractGeneratedNamePrefix(actual)
	}

	return expectedHandled == actualHandled
}

func CompareStringSlices(request, cassette []string) bool {
	if len(request) != len(cassette) {
		return false
	}

	slices.Sort(request)
	slices.Sort(cassette)

	for i, v := range request {
		if !CompareFieldsStrings(v, cassette[i]) {
			return false
		}
	}

	return true
}

// CompareSlices compares two slices of interface{}
// in case of slice of map[string]interface{}, it will attempt to find a match in the other slice without taking into account the order
func CompareSlices(request, cassette []any) bool {
	if len(request) != len(cassette) {
		return false
	}

	if len(request) == 0 {
		return true
	}

	switch request[0].(type) {
	case string:
		requestStrings := make([]string, len(request))
		for i, v := range request {
			requestStrings[i] = v.(string)
		}

		cassetteStrings := make([]string, len(cassette))
		for i, v := range cassette {
			cassetteStrings[i] = v.(string)
		}

		return CompareStringSlices(requestStrings, cassetteStrings)
	case float64:
		sort.Slice(request, func(i, j int) bool {
			return request[i].(float64) < request[j].(float64)
		})
		sort.Slice(cassette, func(i, j int) bool {
			return cassette[i].(float64) < cassette[j].(float64)
		})

		for i := range request {
			if request[i] != cassette[i] {
				return false
			}
		}

		return true
	case map[string]any:
		// first compare assuming that the order is the same, tolerating missing keys in the cassette
		matched := 0

		for i := range request {
			// cleanup ignored keys
			for _, key := range BodyMatcherIgnore {
				RemoveKeyRecursive(request[i].(map[string]any), key)
			}

			for _, key := range BodyMatcherIgnore {
				RemoveKeyRecursive(cassette[i].(map[string]any), key)
			}

			if CompareJSONFields(request[i], cassette[i], false) {
				matched++
			}
		}

		if matched == len(request) {
			return true
		}

		// if no match try to compare out of order
		matched = 0
		reqVisited := make([]bool, len(request))
		casVisited := make([]bool, len(cassette))

		for i := range request {
			if reqVisited[i] {
				continue
			}

			for j := range cassette {
				if casVisited[j] {
					continue
				}

				if CompareJSONFields(request[i], cassette[j], true) {
					matched++
					reqVisited[i] = true
					casVisited[j] = true

					break
				}
			}
		}

		return matched == len(request)
	default:
		return reflect.DeepEqual(request, cassette)
	}
}
