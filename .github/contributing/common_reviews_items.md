# Common Review Items

The Terraform Scaleway Provider follows common practices to ensure consistent and reliable implementations across all resources.
While there may be older resources and testing code that predates these guidelines, new submissions are generally expected to adhere to these items to maintain the quality of the Terraform provider.
For any guidelines listed, contributors are encouraged to ask any questions and community reviewers are encouraged to provide review suggestions based on these guidelines to speed up the review and merge process.

## Go Coding Style

The following Go language resources provide common coding preferences that may be referenced during review, if not automatically handled by the project's linting tools.

- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

## Resource Contribution Guidelines

The following resource checks need to be addressed before your contribution can be merged.
The exclusion of any applicable check may result in a delayed time to merge.

- [ ] __Passes Testing__: All code and documentation changes must pass unit testing, code linting, and website link testing.
  Resource code changes must pass all acceptance testing for the resource.
- [ ] __Avoids API Calls Across Account, Region, and Service Boundaries__: Resources should not implement cross-account, cross-region, or cross-service API calls.
- [ ] __Avoids Optional and Required for Non-Configurable Attributes__: Resource schema definitions for read-only attributes should not include `Optional: true` or `Required: true`.
- [ ] __Avoids resource.Retry() without resource.RetryableError()__: Resource logic should only implement [`resource.Retry()`](https://godoc.org/github.com/hashicorp/terraform/helper/resource#Retry) if there is a retryable condition (e.g. `return resource.RetryableError(err)`).
- [ ] __Avoids Reading Schema Structure in Resource Code__: The resource `Schema` should not be read in resource `Create`/`Read`/`Update`/`Delete` functions to perform looping or otherwise complex attribute logic.
  Use [`d.Get()`](https://godoc.org/github.com/hashicorp/terraform/helper/schema#ResourceData.Get) and [`d.Set()`](https://godoc.org/github.com/hashicorp/terraform/helper/schema#ResourceData.Set) directly with individual attributes instead.
- [ ] __Avoids ResourceData.GetOkExists()__: Resource logic should avoid using [`ResourceData.GetOkExists()`](https://godoc.org/github.com/hashicorp/terraform/helper/schema#ResourceData.GetOkExists) as its expected functionality is not guaranteed in all scenarios.
- [ ] __Implements Read After Create and Update__: Except where API eventual consistency prohibits immediate reading of resources or updated attributes, resource `Create` and `Update` functions should return the resource `Read` function.
- [ ] __Implements Immediate Resource ID Set During Create__: Immediately after calling the API creation function, the resource ID should be set with [`d.SetId()`](https://godoc.org/github.com/hashicorp/terraform/helper/schema#ResourceData.SetId) before other API operations or returning the `Read` function.
- [ ] __Implements Attribute Refreshes During Read__: All attributes available in the API should have [`d.Set()`](https://godoc.org/github.com/hashicorp/terraform/helper/schema#ResourceData.Set) called their values in the Terraform state during the `Read` function.
- [ ] __Implements Error Checks with Non-Primative Attribute Refreshes__: When using [`d.Set()`](https://godoc.org/github.com/hashicorp/terraform/helper/schema#ResourceData.Set) with non-primative types (`schema.TypeList`, `schema.TypeSet`, or `schema.TypeMap`), perform error checking to [prevent issues where the code is not properly able to refresh the Terraform state](https://www.terraform.io/docs/extend/best-practices/detecting-drift.html#error-checking-aggregate-types).
- [ ] __Implements Import Acceptance Testing and Documentation__: Support for resource import (`Importer` in resource schema) must include `ImportState` acceptance testing (see also the [Acceptance Testing Guidelines](#acceptance-testing-guidelines) below) and `## Import` section in resource documentation.
- [ ] __Implements State Migration When Adding New Virtual Attribute__: For new "virtual" attributes (those only in Terraform and not in the API), the schema should implement [State Migration](https://www.terraform.io/docs/extend/resources.html#state-migrations) to prevent differences for existing configurations that upgrade.
- [ ] __Re-Use Resource Read Function in Data Source Read Function__: When possible, re-use the read data-source function between the resource and the data-source.
- [ ] __Uses Scaleway Go SDK Constants__: Many Scaleway services provide string constants for value enumerations, error codes, and status types.
  See also the "Constants" sections under each of the service packages in the [Scaleway Go SDK documentation](https://pkg.go.dev/github.com/scaleway/scaleway-sdk-go).
- [ ] __Uses Scaleway Go SDK Types__: Use available SDK structs instead of implementing custom types with indirection.
- [ ] __Uses TypeList and MaxItems: 1__: Configuration block attributes (e.g. `Type: schema.TypeList` or `Type: schema.TypeSet` with `Elem: &schema.Resource{...}`) that can only have one block should use `Type: schema.TypeList` and `MaxItems: 1` in the schema definition.
- [ ] __Uses Existing Validation Functions__: Schema definitions including `ValidateFunc` for attribute validation should use available [Terraform `helper/validation` package](https://godoc.org/github.com/hashicorp/terraform/helper/validation) functions. `All()`/`Any()` can be used for combining multiple validation function behaviors.
- [ ] __Uses resource.NotFoundError__: Custom errors for missing resources should use [`resource.NotFoundError`](https://godoc.org/github.com/hashicorp/terraform/helper/resource#NotFoundError).
- [ ] __Skips Exists Function__: Implementing a resource `Exists` function is extraneous as it often duplicates resource `Read` functionality. Ensure `d.SetId("")` is used to appropriately trigger resource recreation in the resource `Read` function.
- [ ] __Skips id Attribute__: The `id` attribute is implicit for all Terraform resources and does not need to be defined in the schema.

Below are style-based items that _may_ be noted during review and are recommended for simplicity, consistency, and quality assurance:

- [ ] __Avoids CustomizeDiff__: Usage of `CustomizeDiff` is generally discouraged.
- [ ] __Implements Error Message Context__: Returning errors from resource `Create`, `Read`, `Update`, and `Delete` functions should include additional messaging about the location or cause of the error for operators and code maintainers by wrapping with [`fmt.Errorf()`](https://godoc.org/golang.org/x/exp/errors/fmt#Errorf).
  - An example `Delete` API error: `return fmt.Errorf("error deleting {SERVICE} {THING} (%s): %s", d.Id(), err)`
  - An example `d.Set()` error: `return fmt.Errorf("error setting {ATTRIBUTE}: %w", err)`
- [ ] __Implements Warning Logging With Resource State Removal__: If a resource is removed outside of Terraform (e.g. via different tool, API, or web UI), `d.SetId("")` and `return nil` can be used in the resource `Read` function to trigger resource recreation.
  When this occurs, a warning log message should be printed beforehand: `log.Printf("[WARN] {SERVICE} {THING} (%s) not found, removing from state", d.Id())`
- [ ] __Uses Elem with TypeMap__: While provider schema validation does not error when the `Elem` configuration is not present with `Type: schema.TypeMap` attributes, including the explicit `Elem: &schema.Schema{Type: schema.TypeString}` is recommended.
- [ ] __Uses American English for Attribute Naming__: For any ambiguity with attribute naming, prefer American English over British English. e.g. `color` instead of `colour`.
- [ ] __Skips Timestamp Attributes__: Generally, creation and modification dates from the API should be omitted from the schema.

#### Acceptance Testing Guidelines

Below are required items that will be noted during submission review and prevent immediate merging:

- [ ] __Implements CheckDestroy__: Resource testing should include a `CheckDestroy` function (typically named `testAccCheckScaleway{SERVICE}{RESOURCE}Destroy`) that calls the API to verify that the Terraform resource has been deleted or disassociated as appropriate.
  More information about `CheckDestroy` functions can be found in the [Extending Terraform TestCase documentation](https://www.terraform.io/docs/extend/testing/acceptance-tests/testcase.html#checkdestroy).
- [ ] __Implements Exists Check Function__: Resource testing should include a `TestCheckFunc` function (typically named `testAccCheckScaleway{SERVICE}{RESOURCE}Exists`) that calls the API to verify that the Terraform resource has been created or associated as appropriate.
  Preferably, this function will also accept a pointer to an API object representing the Terraform resource from the API response that can be set for potential usage in later `TestCheckFunc`.
  More information about these functions can be found in the [Extending Terraform Custom Check Functions documentation](https://www.terraform.io/docs/extend/testing/acceptance-tests/testcase.html#checkdestroy).
- [ ] __Excludes Provider Declarations__: Test configurations should not include `provider "scaleway" {...}` declarations.
  If necessary, only the provider declarations in `provider_test.go` should be used for multiple account/region or otherwise specialized testing.
- [ ] __Passes in fr-par-1 Region__: Tests default to running in `fr-par-1` and at a minimum should pass in that region or include necessary `PreCheck` functions to skip the test when ran outside an expected environment.
- [ ] __Uses resource.ParallelTest__: Tests should utilize [`resource.ParallelTest()`](https://godoc.org/github.com/hashicorp/terraform/helper/resource#ParallelTest) instead of [`resource.Test()`](https://godoc.org/github.com/hashicorp/terraform/helper/resource#Test) except where serialized testing is absolutely required.
- [ ] __Uses inline test configuration__: Test configurations preferably should be an inline string directly in the test case with no or as few variables as possible.
- [ ] __Uses Randomized Infrastructure Naming__: Test configurations that utilize resources where a unique name is required should generate a random name.
  Typically this is created via `rName := acctest.RandomWithPrefix("tf-acc-test")` in the acceptance test function before generating the configuration.

For resources that support import, the additional item below is required that will be noted during submission review and prevent immediate merging:

- [ ] __Implements ImportState Testing__: Tests should include an additional `TestStep` configuration that verifies resource import via `ImportState: true` and `ImportStateVerify: true`.
  This `TestStep` should be added to all possible tests for the resource to ensure that all infrastructure configurations are properly imported into Terraform.

Below are style-based items that _may_ be noted during review and are recommended for simplicity, consistency, and quality assurance:

- [ ] __Uses Builtin Check Functions__: Tests should utilize already available check functions, e.g. `resource.TestCheckResourceAttr()`, to verify values in the Terraform state over creating custom `TestCheckFunc`.
  More information about these functions can be found in the [Extending Terraform Builtin Check Functions documentation](https://www.terraform.io/docs/extend/testing/acceptance-tests/teststep.html#builtin-check-functions).
- [ ] __Uses TestCheckResoureAttrPair() for Data Sources__: Tests should utilize [`resource.TestCheckResourceAttrPair()`](https://godoc.org/github.com/hashicorp/terraform/helper/resource#TestCheckResourceAttrPair) to verify values in the Terraform state for data sources attributes to compare them with their expected resource attributes.
- [ ] __Implements Default and Zero Value Validation__: The basic test for a resource (typically named `TestAccScaleway{SERVICE}{RESOURCE}_basic`) should utilize available check functions, e.g. `resource.TestCheckResourceAttr()`, to verify default and zero values in the Terraform state for all attributes.
  Empty/missing configuration blocks can be verified with `resource.TestCheckResourceAttr(resourceName, "{ATTRIBUTE}.#", "0")` and empty maps with `resource.TestCheckResourceAttr(resourceName, "{ATTRIBUTE}.%", "0")`
