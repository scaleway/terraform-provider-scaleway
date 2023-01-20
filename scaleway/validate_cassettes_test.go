package scaleway

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/stretchr/testify/assert"
)

const testDirectory = "testdata/"

// getTestFiles returns a map of cassettes files
func getTestFiles() (map[string]struct{}, error) {
	filesMap := make(map[string]struct{})
	files, err := ioutil.ReadDir(testDirectory)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		filesMap[fileNameWithoutExtSuffix(file.Name())] = struct{}{}
	}

	return filesMap, nil
}

func TestAccScalewayCassettes_Validator(t *testing.T) {
	files, err := getTestFiles()
	assert.NoError(t, err)

	for name := range files {
		c, err := cassette.Load(fmt.Sprintf("%s%s", testDirectory, name))
		assert.NoError(t, err)
		assert.NoError(t, checkErrorCode(c))
	}
}

func checkErrorCode(c *cassette.Cassette) error {
	for _, i := range c.Interactions {
		if !checkErrCodeExcept(i, c, http.StatusNotFound, http.StatusTooManyRequests, http.StatusForbidden) &&
			!isTransientStateError(i) {
			return fmt.Errorf("status: %v found on %s. method: %s, url %s\nrequest body = %v\nresponse body = %v", i.Code, c.Name, i.Request.Method, i.Request.URL, i.Request.Body, i.Response.Body)
		}
	}

	return nil
}

func exceptionsCassettesCases() map[string]struct{} {
	return map[string]struct{}{
		"testdata/object-bucket-destroy-force.cassette.yaml":     {},
		"testdata/object-bucket-cors-empty-origin.cassette.yaml": {},
		"testdata/rdb-privilege-basic.cassette.yaml":             {},
		"testdata/data-source-rdb-privilege-basic.cassette.yaml": {},
	}
}

func checkErrCodeExcept(i *cassette.Interaction, c *cassette.Cassette, codes ...int) bool {
	exceptions := exceptionsCassettesCases()
	_, isException := exceptions[c.File]
	if isException {
		return isException
	}

	if i.Code >= 400 {
		for _, httpCode := range codes {
			if i.Code == httpCode {
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

func isTransientStateError(i *cassette.Interaction) bool {
	if i.Code != 409 {
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
