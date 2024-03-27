package acctest_test

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/dnaeon/go-vcr/cassette"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
)

func exceptionsCassettesCases() map[string]struct{} {
	return map[string]struct{}{
		"../services/documentdb/testdata/privilege-basic.cassette.yaml":         {},
		"../services/mnq/testdata/sns-topic-basic.cassette.yaml":                {},
		"../services/mnq/testdata/sns-topic-subscription-basic.cassette.yaml":   {},
		"../services/mnq/testdata/sqs-already-activated.cassette.yaml":          {},
		"../services/object/testdata/bucket-cors-empty-origin.cassette.yaml":    {},
		"../services/object/testdata/bucket-destroy-force.cassette.yaml":        {},
		"../services/rdb/testdata/data-source-privilege-basic.cassette.yaml":    {},
		"../services/rdb/testdata/privilege-basic.cassette.yaml":                {},
		"../services/object/testdata/object-bucket-destroy-force.cassette.yaml": {},
	}
}

// getTestFiles returns a map of cassettes files
func getTestFiles() (map[string]struct{}, error) {
	filesMap := make(map[string]struct{})
	exceptions := exceptionsCassettesCases()
	err := filepath.WalkDir("../services", func(path string, d fs.DirEntry, err error) error {
		isCassette := strings.Contains(path, "cassette")
		_, isException := exceptions[path]
		if isCassette && !isException {
			filesMap[fileNameWithoutExtSuffix(path)] = struct{}{}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return filesMap, nil
}

func TestAccCassettes_Validator(t *testing.T) {
	paths, err := getTestFiles()
	require.NoError(t, err)

	for path := range paths {
		c, err := cassette.Load(path)
		require.NoError(t, err)
		assert.NoError(t, checkErrorCode(c))
	}
}

func checkErrorCode(c *cassette.Cassette) error {
	for _, i := range c.Interactions {
		if !checkErrCodeExcept(i, c, http.StatusNotFound, http.StatusTooManyRequests, http.StatusForbidden, http.StatusGone) &&
			!isTransientStateError(i) {
			return fmt.Errorf("status: %v found on %s. method: %s, url %s\nrequest body = %v\nresponse body = %v", i.Response.Code, c.Name, i.Request.Method, i.Request.URL, i.Request.Body, i.Response.Body)
		}
	}

	return nil
}

func checkErrCodeExcept(i *cassette.Interaction, c *cassette.Cassette, codes ...int) bool {
	exceptions := exceptionsCassettesCases()
	_, isException := exceptions[c.File]
	if isException {
		return isException
	}

	// SQS returns 400 when the queue does not exist
	if strings.Contains(i.Response.Body, sqs.ErrCodeQueueDoesNotExist) && i.Response.Code == 400 {
		return true
	}

	if i.Response.Code >= 400 {
		for _, httpCode := range codes {
			if i.Response.Code == httpCode {
				return true
			}
		}
		return false
	}

	return true
}

func fileNameWithoutExtSuffix(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

// isTransientStateError checks if the interaction response is a transient state error
// Transient state error are expected when creating resource linked to each other
// example:
// creating a gateway_network will set its public gateway to a transient state
// when creating 2 gateway_network, one will fail with a transient state error
// but the transient state error will be caught, it will wait again for the resource to be ready
func isTransientStateError(i *cassette.Interaction) bool {
	if i.Response.Code != 409 {
		return false
	}

	scwError := struct {
		Type string `json:"type"`
	}{}

	err := json.Unmarshal([]byte(i.Response.Body), &scwError)
	if err != nil {
		return false
	}

	return scwError.Type == "transient_state"
}
