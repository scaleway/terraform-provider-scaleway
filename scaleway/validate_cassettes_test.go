package scaleway

import (
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
		switch i.Code {
		case http.StatusInternalServerError:
			return fmt.Errorf("status: %v founded on %s. method: %s, url %s", i.Code, c.Name, i.Request.Method, i.Request.URL)
		case http.StatusConflict:
			if !checkExceptions(c) {
				return fmt.Errorf("status: %v founded on %s. method: %s, url %s", i.Code, c.Name, i.Request.Method, i.Request.URL)
			}
		}
	}

	return nil
}

func exceptionsCassettesCases() map[string]struct{} {
	return map[string]struct{}{
		"testdata/object-bucket-destroy-force.cassette.yaml":     {},
		"testdata/rdb-privilege-basic.cassette.yaml":             {},
		"testdata/data-source-rdb-privilege-basic.cassette.yaml": {}}
}

func checkExceptions(c *cassette.Cassette) bool {
	exceptions := exceptionsCassettesCases()
	_, exist := exceptions[c.File]
	return exist
}
func fileNameWithoutExtSuffix(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}
