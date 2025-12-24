package acctest_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/mnq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
)

const servicesDir = "../services"

func exceptionsCassettesCases() map[string]struct{} {
	return map[string]struct{}{
		"../services/mnq/testdata/sns-topic-basic.cassette.yaml":                                           {},
		"../services/mnq/testdata/sns-topic-subscription-basic.cassette.yaml":                              {},
		"../services/mnq/testdata/sqs-already-activated.cassette.yaml":                                     {},
		"../services/object/testdata/bucket-cors-empty-origin.cassette.yaml":                               {},
		"../services/object/testdata/bucket-destroy-force.cassette.yaml":                                   {},
		"../services/rdb/testdata/data-source-privilege-basic.cassette.yaml":                               {},
		"../services/rdb/testdata/privilege-basic.cassette.yaml":                                           {},
		"../services/object/testdata/object-bucket-destroy-force.cassette.yaml":                            {},
		"../services/secret/testdata/secret-protected.cassette.yaml":                                       {},
		"../services/secret/testdata/secret-version-type.cassette.yaml":                                    {},
		"../services/file/testdata/file-system-invalid-size-granularity-fails.cassette.yaml":               {},
		"../services/file/testdata/file-system-size-too-small-fails.cassette.yaml":                         {},
		"../services/container/testdata/namespace-vpc-integration.cassette.yaml":                           {},
		"../services/function/testdata/function-namespace-vpc-integration.cassette.yaml":                   {},
		"../services/baremetal/testdata/server-cloud-init-not-compatible-offer.cassette.yaml":              {},
		"../services/keymanager/testdata/decrypt-ephemeral-resource-invalid-associated-data.cassette.yaml": {},
	}
}

// getTestFiles returns a map of cassettes files
func getTestFiles(includeExceptions bool) (map[string]struct{}, error) {
	filesMap := make(map[string]struct{})
	exceptions := exceptionsCassettesCases()

	err := filepath.WalkDir(servicesDir, func(path string, _ fs.DirEntry, _ error) error {
		if isCassette := strings.Contains(path, "cassette"); isCassette {
			_, isException := exceptions[path]
			if !isException || includeExceptions {
				filesMap[fileNameWithoutExtSuffix(path)] = struct{}{}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return filesMap, nil
}

func TestAccCassettes_Validator(t *testing.T) {
	paths, err := getTestFiles(false)
	require.NoError(t, err)

	for path := range paths {
		c, err := cassette.Load(path)
		require.NoError(t, err)
		assert.NoError(t, checkErrorCode(c))
	}
}

func checkErrorCode(c *cassette.Cassette) error {
	for _, i := range c.Interactions {
		if !checkErrCodeExcept(i, c, http.StatusBadRequest, http.StatusNotFound, http.StatusTooManyRequests, http.StatusForbidden, http.StatusGone) &&
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
	if strings.Contains(i.Response.Body, mnq.AWSErrNonExistentQueue) && i.Response.Code == 400 {
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

func listAccTestFunctions() (map[string]string, error) {
	fset := token.NewFileSet()
	testFuncs := map[string]string{}

	err := filepath.WalkDir(servicesDir, func(path string, _ fs.DirEntry, _ error) error {
		if strings.HasSuffix(path, "_test.go") {
			pkgFolder := filepath.Base(filepath.Dir(path))

			node, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				return err
			}

			for _, decl := range node.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok {
					if strings.HasPrefix(fn.Name.Name, "Test") && fn.Name.Name != "TestMain" && fn.Recv == nil {
						expectedCassettePath := fmt.Sprintf("%s/%s", servicesDir, acctest.BuildCassetteName(fn.Name.Name, pkgFolder, ".cassette"))
						testFuncs[expectedCassettePath] = fmt.Sprintf("%s/%s", pkgFolder, fn.Name.Name)
					}
				}
			}
		}

		return nil
	})

	return testFuncs, err
}

func TestAccCassettes_CheckOrphans(t *testing.T) {
	// List actual cassettes
	actualCassettesPaths, err := getTestFiles(true)
	if err != nil {
		t.Fatalf("Failed to list cassettes: %v", err)
	}

	// List actual acceptance tests functions and their expected cassettes' paths
	expectedCassettesPaths, err := listAccTestFunctions()
	if err != nil {
		t.Fatalf("Failed to list acceptance tests: %v", err)
	}

	testWithNoCassetteErrs := []string(nil)
	cassetteWithNoTestErrs := []error(nil)

	// Look for tests with no matching cassette
	for expectedCassettePath, testName := range expectedCassettesPaths {
		if _, ok := actualCassettesPaths[expectedCassettePath]; !ok {
			testWithNoCassetteErrs = append(testWithNoCassetteErrs, fmt.Sprintf("- %s has no matching cassette", testName))
		}
	}

	// Look for cassettes with no matching test
	for actualCassettePath := range actualCassettesPaths {
		if _, ok := expectedCassettesPaths[actualCassettePath]; !ok {
			cassetteWithNoTestErrs = append(cassetteWithNoTestErrs, fmt.Errorf("+ cassette [%s] has no matching test", actualCassettePath))
		}
	}

	// Print results:
	// If a cassette has no test, it should result in an error, but if a test has no cassette, it should only result in
	// a warning (e.g. for tests that are currently skipped and which cassette had to be removed because of a 500, or else)
	sort.Strings(testWithNoCassetteErrs)
	t.Log("WARNING:\n", strings.Join(testWithNoCassetteErrs, "\n"))

	if len(cassetteWithNoTestErrs) > 0 {
		sort.Slice(cassetteWithNoTestErrs, func(i, j int) bool {
			return cassetteWithNoTestErrs[i].Error() < cassetteWithNoTestErrs[j].Error()
		})
		t.Error(errors.Join(cassetteWithNoTestErrs...))
	}
}
