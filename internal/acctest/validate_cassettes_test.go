package acctest_test

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"net/http"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/mnq"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/vcr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
)

const servicesDir = "../services"

func exceptionsCassettesCases() map[string]struct{} {
	return map[string]struct{}{
		"../services/baremetal/testdata/server-cloud-init-not-compatible-offer.cassette.yaml":              {},
		"../services/block/testdata/action-snapshot-export-basic.cassette.yaml":                            {},
		"../services/container/testdata/container-private-endpoint.cassette.yaml":                          {},
		"../services/container/testdata/data-source-container-private-endpoint.cassette.yaml":              {},
		"../services/container/testdata/namespace-vpc-integration.cassette.yaml":                           {},
		"../services/file/testdata/file-system-invalid-size-granularity-fails.cassette.yaml":               {},
		"../services/file/testdata/file-system-size-too-small-fails.cassette.yaml":                         {},
		"../services/function/testdata/function-namespace-vpc-integration.cassette.yaml":                   {},
		"../services/iam/testdata/data-source-scim-invalid-deactivated.cassette.yaml":                      {},
		"../services/keymanager/testdata/decrypt-ephemeral-resource-invalid-associated-data.cassette.yaml": {},
		"../services/mnq/testdata/sns-topic-basic.cassette.yaml":                                           {},
		"../services/mnq/testdata/sns-topic-subscription-basic.cassette.yaml":                              {},
		"../services/mnq/testdata/sqs-already-activated.cassette.yaml":                                     {},
		"../services/object/testdata/bucket-cors-empty-origin.cassette.yaml":                               {},
		"../services/object/testdata/bucket-destroy-force.cassette.yaml":                                   {},
		"../services/object/testdata/object-bucket-destroy-force.cassette.yaml":                            {},
		"../services/rdb/testdata/data-source-privilege-basic.cassette.yaml":                               {},
		"../services/rdb/testdata/privilege-basic.cassette.yaml":                                           {},
		"../services/secret/testdata/secret-protected.cassette.yaml":                                       {},
		"../services/secret/testdata/secret-version-type.cassette.yaml":                                    {},
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
		return slices.Contains(codes, i.Response.Code)
	}

	return true
}

func fileNameWithoutExtSuffix(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

// isTransientStateError checks if the interaction response is a transient error that
// the provider retries and that succeeds later in the same cassette.
//
// Three shapes are treated as transient:
//   - 409 transient_state: a resource is in a transient state during creation (e.g.
//     creating a gateway_network puts its public gateway in a transient state, so a
//     parallel creation fails and is retried).
//   - 412 precondition_failed: project DELETE before the API finishes garbage-collecting
//     the project's resources; the provider retries and the project is deleted once empty.
//   - 409 BucketNotEmpty: S3 bucket DELETE while objects still exist; the provider clears
//     them and retries the DELETE.
func isTransientStateError(i *cassette.Interaction) bool {
	switch i.Response.Code {
	case 409:
		// S3 returns XML, not JSON: <Error><Code>BucketNotEmpty</Code>...</Error>
		if strings.Contains(i.Response.Body, "<Code>BucketNotEmpty</Code>") {
			return true
		}

		scwError := struct {
			Type string `json:"type"`
		}{}
		if err := json.Unmarshal([]byte(i.Response.Body), &scwError); err == nil && scwError.Type == "transient_state" {
			return true
		}
	case 412:
		// ponytail: 412 precondition_failed today only comes from account/v3 project
		// DELETE cleanup. If another endpoint starts returning it, narrow this check.
		scwError := struct {
			Type string `json:"type"`
		}{}
		if err := json.Unmarshal([]byte(i.Response.Body), &scwError); err == nil && scwError.Type == "precondition_failed" {
			return true
		}
	}

	return false
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
						expectedCassettePath := fmt.Sprintf("%s/%s", servicesDir, vcr.BuildCassetteName(fn.Name.Name, pkgFolder, ".cassette"))
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
