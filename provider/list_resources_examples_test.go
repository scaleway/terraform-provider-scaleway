package provider_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/list"
	providerFramework "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/provider"
)

// extractListResourceTypeName extracts the type name from a list resource
func extractListResourceTypeName(t *testing.T, resourceFunc func() list.ListResource) string {
	t.Helper()

	r := resourceFunc()
	resp := &resource.MetadataResponse{}
	r.Metadata(t.Context(), resource.MetadataRequest{
		ProviderTypeName: "scaleway",
	}, resp)

	return resp.TypeName
}

// extractResourceTypeName extracts the type name from a framework resource
func extractResourceTypeName(t *testing.T, resourceFunc func() resource.Resource) string {
	t.Helper()

	r := resourceFunc()
	resp := &resource.MetadataResponse{}
	r.Metadata(t.Context(), resource.MetadataRequest{
		ProviderTypeName: "scaleway",
	}, resp)

	return resp.TypeName
}

// extractDataSourceTypeName extracts the type name from a framework data source
func extractDataSourceTypeName(t *testing.T, dataSourceFunc func() datasource.DataSource) string {
	t.Helper()

	d := dataSourceFunc()
	resp := &datasource.MetadataResponse{}
	d.Metadata(t.Context(), datasource.MetadataRequest{
		ProviderTypeName: "scaleway",
	}, resp)

	return resp.TypeName
}

// extractSDKv2ResourceTypeNames extracts the type names from an SDKv2 provider resource map
func extractSDKv2ResourceTypeNames(_ *testing.T, provider *schema.Provider) []string {
	typeNames := make([]string, 0, len(provider.ResourcesMap))
	for typeName := range provider.ResourcesMap {
		typeNames = append(typeNames, typeName)
	}

	return typeNames
}

// extractSDKv2DataSourceTypeNames extracts the type names from an SDKv2 provider data source map
func extractSDKv2DataSourceTypeNames(_ *testing.T, provider *schema.Provider) []string {
	typeNames := make([]string, 0, len(provider.DataSourcesMap))
	for typeName := range provider.DataSourcesMap {
		typeNames = append(typeNames, typeName)
	}

	return typeNames
}

// extractActionTypeName extracts the type name from an action
func extractActionTypeName(t *testing.T, actionFunc func() action.Action) string {
	t.Helper()

	a := actionFunc()
	resp := &action.MetadataResponse{}
	a.Metadata(t.Context(), action.MetadataRequest{
		ProviderTypeName: "scaleway",
	}, resp)

	return resp.TypeName
}

// extractEphemeralResourceTypeName extracts the type name from an ephemeral resource
func extractEphemeralResourceTypeName(t *testing.T, ephemeralFunc func() ephemeral.EphemeralResource) string {
	t.Helper()

	e := ephemeralFunc()
	resp := &ephemeral.MetadataResponse{}
	e.Metadata(t.Context(), ephemeral.MetadataRequest{
		ProviderTypeName: "scaleway",
	}, resp)

	return resp.TypeName
}

// extractFunctionTypeName extracts the type name from a function
func extractFunctionTypeName(t *testing.T, functionFunc func() function.Function) string {
	t.Helper()

	f := functionFunc()
	resp := &function.MetadataResponse{}
	f.Metadata(t.Context(), function.MetadataRequest{}, resp)

	return resp.Name
}

// checkExamplesExist checks if examples exist for a given type name in the specified directory
func checkExamplesExist(t *testing.T, typeName string, examplesDir string) (exists bool, hasFiles bool) {
	t.Helper()

	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		// Directory doesn't exist - no examples can exist
		if os.IsNotExist(err) {
			return false, false
		}

		t.Fatalf("Failed to read examples directory: %v", err)
	}

	// Check if a subdirectory exists
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() == typeName {
			return true, false
		}
	}

	// Check if files with prefix exist directly in the directory
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), typeName+".") {
			return false, true
		}
	}

	return false, false
}

// checkExampleDirectoryHasTfqueryFiles checks if a list-resources example directory contains .tfquery.hcl files
func checkExampleDirectoryHasTfqueryFiles(t *testing.T, examplesDir string) []string {
	t.Helper()

	var examplesWithoutFiles []string

	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		// Directory doesn't exist - no directories to check
		if os.IsNotExist(err) {
			return examplesWithoutFiles
		}

		t.Fatalf("Failed to read examples directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		exampleDir := filepath.Join(examplesDir, entry.Name())
		hasTfqueryFile := false

		dirEntries, err := os.ReadDir(exampleDir)
		if err != nil {
			t.Errorf("Failed to read example directory %s: %v", exampleDir, err)

			continue
		}

		for _, dirEntry := range dirEntries {
			if !dirEntry.IsDir() && strings.HasSuffix(dirEntry.Name(), ".tfquery.hcl") {
				hasTfqueryFile = true

				break
			}
		}

		if !hasTfqueryFile {
			examplesWithoutFiles = append(examplesWithoutFiles, entry.Name())
		}
	}

	return examplesWithoutFiles
}

// checkExampleDirectoryHasTfFiles checks if an example directory contains .tf files with a given prefix
func checkExampleDirectoryHasTfFiles(t *testing.T, examplesDir string, filePrefix string) []string {
	t.Helper()

	var examplesWithoutFiles []string

	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		// Directory doesn't exist - no directories to check
		if os.IsNotExist(err) {
			return examplesWithoutFiles
		}

		t.Fatalf("Failed to read examples directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		exampleDir := filepath.Join(examplesDir, entry.Name())
		hasTfFile := false

		dirEntries, err := os.ReadDir(exampleDir)
		if err != nil {
			t.Errorf("Failed to read example directory %s: %v", exampleDir, err)

			continue
		}

		for _, dirEntry := range dirEntries {
			if !dirEntry.IsDir() && strings.HasPrefix(dirEntry.Name(), filePrefix) && strings.HasSuffix(dirEntry.Name(), ".tf") {
				hasTfFile = true

				break
			}
		}

		if !hasTfFile {
			examplesWithoutFiles = append(examplesWithoutFiles, entry.Name())
		}
	}

	return examplesWithoutFiles
}

func TestProviderExamples_Resources(t *testing.T) {
	// Test Resources (SDKv2 + Framework)
	t.Run("Resources", func(t *testing.T) {
		examplesDir := filepath.Join("..", "examples", "resources")

		// Get all existing example directories
		existingExamples := make(map[string]bool)

		entries, err := os.ReadDir(examplesDir)
		if err != nil {
			t.Fatalf("Failed to read examples directory: %v", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				existingExamples[entry.Name()] = true
			}
		}

		// Collect all resources missing examples
		var missingExamples []string

		// SDKv2 resources
		sdkProvider := provider.SDKProvider(nil)()
		for _, typeName := range extractSDKv2ResourceTypeNames(t, sdkProvider) {
			exists, hasFiles := checkExamplesExist(t, typeName, examplesDir)

			if !exists && !hasFiles {
				missingExamples = append(missingExamples, typeName)
			}
		}

		// Framework resources
		pResource := provider.NewFrameworkProvider(nil)()
		for _, resourceFunc := range pResource.Resources(t.Context()) {
			typeName := extractResourceTypeName(t, resourceFunc)

			exists, hasFiles := checkExamplesExist(t, typeName, examplesDir)

			if !exists && !hasFiles {
				missingExamples = append(missingExamples, typeName)
			}
		}

		// Collect all example directories missing resource*.tf files
		examplesWithoutFiles := checkExampleDirectoryHasTfFiles(t, examplesDir, "resource")

		if len(missingExamples) > 0 {
			t.Errorf("Found %d resource(s) without example files:\n%s\n"+
				"Please add example .tf files in examples/resources/<resource-name>/ (e.g., resource-basic.tf)",
				len(missingExamples), strings.Join(missingExamples, "\n"))
		}

		if len(examplesWithoutFiles) > 0 {
			t.Errorf("Found %d resources example directory(ies) without resource*.tf files:\n%s\n"+
				"Please add .tf files named resource*.tf to these directories",
				len(examplesWithoutFiles), strings.Join(examplesWithoutFiles, "\n"))
		}
	})
}

// Test Data Sources (SDKv2 + Framework)
func TestProviderExamples_DataSources(t *testing.T) {
	examplesDir := filepath.Join("..", "examples", "data-sources")

	// Get all existing example directories
	existingExamples := make(map[string]bool)

	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		t.Fatalf("Failed to read examples directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			existingExamples[entry.Name()] = true
		}
	}

	// Collect all data sources missing examples
	var missingExamples []string

	// SDKv2 data sources
	sdkProvider := provider.SDKProvider(nil)()
	for _, typeName := range extractSDKv2DataSourceTypeNames(t, sdkProvider) {
		exists, hasFiles := checkExamplesExist(t, typeName, examplesDir)

		if !exists && !hasFiles {
			missingExamples = append(missingExamples, typeName)
		}
	}

	// Framework data sources
	pDataSource := provider.NewFrameworkProvider(nil)()
	for _, dataSourceFunc := range pDataSource.DataSources(t.Context()) {
		typeName := extractDataSourceTypeName(t, dataSourceFunc)

		exists, hasFiles := checkExamplesExist(t, typeName, examplesDir)

		if !exists && !hasFiles {
			missingExamples = append(missingExamples, typeName)
		}
	}

	// Collect all example directories missing data-source*.tf files
	examplesWithoutFiles := checkExampleDirectoryHasTfFiles(t, examplesDir, "data-source")

	if len(missingExamples) > 0 {
		t.Errorf("Found %d data source(s) without example files:\n%s\n"+
			"Please add example .tf files in examples/data-sources/<data-source-name>/ (e.g., data-source-basic.tf)",
			len(missingExamples), strings.Join(missingExamples, "\n"))
	}

	if len(examplesWithoutFiles) > 0 {
		t.Errorf("Found %d data-sources example directory(ies) without data-source*.tf files:\n%s\n"+
			"Please add .tf files named data-source*.tf to these directories",
			len(examplesWithoutFiles), strings.Join(examplesWithoutFiles, "\n"))
	}
}

func TestProviderExamples_ListResources(t *testing.T) {
	p := provider.NewFrameworkProvider(nil)().(providerFramework.ProviderWithListResources)
	listResources := p.ListResources(t.Context())
	examplesDir := filepath.Join("..", "examples", "list-resources")

	// Get all existing example directories
	existingExamples := make(map[string]bool)

	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		t.Fatalf("Failed to read examples directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			existingExamples[entry.Name()] = true
		}
	}

	// Collect all resources missing examples
	var missingExamples []string

	for _, resourceFunc := range listResources {
		typeName := extractListResourceTypeName(t, resourceFunc)

		exists, hasFiles := checkExamplesExist(t, typeName, examplesDir)

		if !exists && !hasFiles {
			missingExamples = append(missingExamples, typeName)
		}
	}

	// Collect all example directories missing .tfquery.hcl files
	examplesWithoutFiles := checkExampleDirectoryHasTfqueryFiles(t, examplesDir)

	if len(missingExamples) > 0 {
		t.Errorf("Found %d list resource(s) without example files:\n%s\n"+
			"Please add example .tfquery.hcl files in examples/list-resources/<resource-name>/",
			len(missingExamples), strings.Join(missingExamples, "\n"))
	}

	if len(examplesWithoutFiles) > 0 {
		t.Errorf("Found %d list-resources example directory(ies) without .tfquery.hcl files:\n%s\n"+
			"Please add .tfquery.hcl files to these directories",
			len(examplesWithoutFiles), strings.Join(examplesWithoutFiles, "\n"))
	}
}

func TestProviderExamples_Actions(t *testing.T) {
	pAction := provider.NewFrameworkProvider(nil)().(providerFramework.ProviderWithActions)
	actions := pAction.Actions(t.Context())
	examplesDir := filepath.Join("..", "examples", "actions")

	// Get all existing example directories
	existingExamples := make(map[string]bool)

	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		t.Fatalf("Failed to read examples directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			existingExamples[entry.Name()] = true
		}
	}

	// Collect all actions missing examples
	var missingExamples []string

	for _, actionFunc := range actions {
		typeName := extractActionTypeName(t, actionFunc)

		exists, hasFiles := checkExamplesExist(t, typeName, examplesDir)

		if !exists && !hasFiles {
			missingExamples = append(missingExamples, typeName)
		}
	}

	// Collect all example directories missing action*.tf files
	examplesWithoutFiles := checkExampleDirectoryHasTfFiles(t, examplesDir, "action")

	if len(missingExamples) > 0 {
		t.Errorf("Found %d action(s) without example files:\n%s\n"+
			"Please add example .tf files in examples/actions/<action-name>/ (e.g., action_example.tf)",
			len(missingExamples), strings.Join(missingExamples, "\n"))
	}

	if len(examplesWithoutFiles) > 0 {
		t.Errorf("Found %d actions example directory(ies) without action*.tf files:\n%s\n"+
			"Please add .tf files named action*.tf to these directories",
			len(examplesWithoutFiles), strings.Join(examplesWithoutFiles, "\n"))
	}
}

func TestProviderExamples_EphemeralResources(t *testing.T) {
	pEphemeral := provider.NewFrameworkProvider(nil)().(providerFramework.ProviderWithEphemeralResources)
	ephemeralResources := pEphemeral.EphemeralResources(t.Context())
	examplesDir := filepath.Join("..", "examples", "ephemeral-resources")

	// Get all existing example directories
	existingExamples := make(map[string]bool)

	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		t.Fatalf("Failed to read examples directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			existingExamples[entry.Name()] = true
		}
	}

	// Collect all ephemeral resources missing examples
	var missingExamples []string

	for _, ephemeralFunc := range ephemeralResources {
		typeName := extractEphemeralResourceTypeName(t, ephemeralFunc)

		exists, hasFiles := checkExamplesExist(t, typeName, examplesDir)

		if !exists && !hasFiles {
			missingExamples = append(missingExamples, typeName)
		}
	}

	// Collect all example directories missing ephemeral-resource*.tf files
	examplesWithoutFiles := checkExampleDirectoryHasTfFiles(t, examplesDir, "ephemeral-resource")

	if len(missingExamples) > 0 {
		t.Errorf("Found %d ephemeral resource(s) without example files:\n%s\n"+
			"Please add example .tf files in examples/ephemeral-resources/<resource-name>/ (e.g., ephemeral-resource_example.tf)",
			len(missingExamples), strings.Join(missingExamples, "\n"))
	}

	if len(examplesWithoutFiles) > 0 {
		t.Errorf("Found %d ephemeral-resources example directory(ies) without ephemeral-resource*.tf files:\n%s\n"+
			"Please add .tf files named ephemeral-resource*.tf to these directories",
			len(examplesWithoutFiles), strings.Join(examplesWithoutFiles, "\n"))
	}
}

func TestProviderExamples_Functions(t *testing.T) {
	pFunctions := provider.NewFrameworkProvider(nil)().(providerFramework.ProviderWithFunctions)
	functions := pFunctions.Functions(t.Context())
	examplesDir := filepath.Join("..", "examples", "functions")

	// Get all existing example directories
	existingExamples := make(map[string]bool)

	entries, err := os.ReadDir(examplesDir)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("Failed to read examples directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			existingExamples[entry.Name()] = true
		}
	}

	// Collect all functions missing examples
	var missingExamples []string

	for _, functionFunc := range functions {
		typeName := extractFunctionTypeName(t, functionFunc)

		exists, hasFiles := checkExamplesExist(t, typeName, examplesDir)

		if !exists && !hasFiles {
			missingExamples = append(missingExamples, typeName)
		}
	}

	// Collect all example directories missing function*.tf files
	examplesWithoutFiles := checkExampleDirectoryHasTfFiles(t, examplesDir, "function")

	if len(missingExamples) > 0 {
		t.Errorf("Found %d function(s) without example files:\n%s\n"+
			"Please add example .tf files in examples/functions/<function-name>/ (e.g., function_example.tf)",
			len(missingExamples), strings.Join(missingExamples, "\n"))
	}

	if len(examplesWithoutFiles) > 0 {
		t.Errorf("Found %d functions example directory(ies) without function*.tf files:\n%s\n"+
			"Please add .tf files named function*.tf to these directories",
			len(examplesWithoutFiles), strings.Join(examplesWithoutFiles, "\n"))
	}

}
