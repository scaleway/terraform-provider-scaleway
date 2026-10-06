# Acceptance Tests

Terraform includes an acceptance test harness that does most of the repetitive work involved in testing a resource.
For additional information about testing Terraform Providers, see the [Extending Terraform documentation](https://www.terraform.io/docs/extend/testing/index.html).

## Acceptance Tests Often Cost Money to Run 💸

Because acceptance tests create real resources, they often cost money to run.
Because the resources only exist for a short period of time, the total amount of money required is usually relatively small.
Nevertheless, we don't want financial limitations to be a barrier to contribution, so if you are unable to pay to run acceptance tests for your contribution, mention this in your pull request.
We will happily accept "best effort" implementations of acceptance tests and run them for you on our side.
This might mean that your PR will take a bit longer to merge, but it most definitely is not a blocker for contributions.

## Running an Acceptance Test

Acceptance tests can be run with the `test:provider` task of [`mise.toml`](../../mise.toml).
The individual tests to run can be controlled using a regular expression.
By default the tests run against the recorded mocks; add `--cassettes` to hit the real API and record new cassettes.
Prior to running the tests, provider configuration details such as access keys must be made available (the provider reads the `scw` CLI config in `~/.config/scw/config.yaml`, falling back to the `SCW_*` environment variables below).

For example, to run an acceptance test against the Scaleway provider, the following environment variables must be set:

```sh
# Using a profile
export SCW_PROFILE=...
# Otherwise
export SCW_ACCESS_KEY=...
export SCW_SECRET_KEY=...
export SCW_DEFAULT_REGION=...
export SCW_DEFAULT_ZONE=...
```

Please note that the default zone for the testing is `fr-par-1` and the default region is `fr-par`.
If needed, you can override via the `SCW_DEFAULT_REGION`, `SCW_DEFAULT_ZONE` environment variable.

Tests can then be run by specifying a regular expression defining the tests to run:

```sh
$ mise run test:provider --acceptance --run 'TestAccScalewayInstanceServerBasic1'
```

Entire resource test suites can be targeted by using the naming convention to write the regular expression.
For example, to run all tests of the `scaleway_instance_security_group` resource rather than just the update test, you can start testing like this:

```sh
$ mise run test:provider --acceptance --run 'TestAccScalewayInstanceSecurityGroup'
```

#### Writing an Acceptance Test

Terraform has a framework for writing acceptance tests which minimises the amount of boilerplate code necessary to use common testing patterns.
The entry point to the framework is the `resource.ParallelTest()` function.

Tests are divided into `TestStep`s.
Each `TestStep` proceeds by applying some Terraform configuration using the provider under test, and then verifying that results are as expected by making assertions using the provider API.
It is common for a single test function to exercise both the creation of and updates to a single resource.
Most tests follow a similar structure.

First, pre-flight checks are made to ensure that sufficient provider configuration is available to be able to proceed.
For example, in an acceptance test targeting Scaleway, `SCW_ACCESS_KEY_ID` and `SCW_SECRET_ACCESS_KEY` must be set prior to running acceptance tests.
This is common to all tests exercising a single provider.

Each `TestStep` is defined in the call to `resource.ParallelTest()`.
Most assertion functions are defined out of band with the tests.
This keeps the tests readable, and allows reuse of assertion functions across different tests of the same type of resource.
The definition of a complete test looks like this:

```go
func TestAccServerImport(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: instancechecks.CheckServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_server" "server01" {
						type  = "DEV1-S"
						image = "f974feac-abae-4365-b988-8ec7d1cec10d"
						state = "stopped"
					}
				`,
			},
			{
				ResourceName:      "scaleway_instance_server.server01",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
```

When executing the test, the following steps are taken for each `TestStep`:

1. The Terraform configuration required for the test is applied.
   This is responsible for configuring the resource under test, and any dependencies it may have.
   For example, to test the `scaleway_instance_server` resource, a valid configuration with the requisite fields is required.
   This results in configuration which looks like this:

    ```terraform
    resource "scaleway_instance_server" "server01" {
        type  = "DEV1-S"
        image = "f974feac-abae-4365-b988-8ec7d1cec10d"
        state = "stopped"
    }
    ```

1. Assertions are run using the provider API.
   These use the provider API directly rather than asserting against the resource state.
   For example, to verify that the `scaleway_instance_server` described above was created successfully, a test function like this is used:

    ```go
    func testAccCheckInstanceServerExists(tt *TestTools, n string) resource.TestCheckFunc {
    	return func(state *terraform.State) error {
    		rs, ok := state.RootModule().Resources[n]
    		if !ok {
    			return fmt.Errorf("resource not found: %s", n)
    		}
    
    		instanceAPI, zone, ID, err := InstanceAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
    		if err != nil {
    			return err
    		}
    
    		_, err = instanceAPI.GetServer(&instance.GetServerRequest{ServerID: ID, Zone: zone})
    		if err != nil {
    			return err
    		}
    
    		return nil
    	}
    }
    ```

   Notice that the only information used from the Terraform state is the ID of the resource.
   For computed properties, we instead assert that the value saved in the Terraform state was the expected value if possible.
   The testing framework provides helper functions for several common types of check - for example:

    ```go
    resource.TestCheckResourceAttr("scaleway_instance_server.foobar", "server_name", testAccScalewayInstanceServerName(rInt)),
    ```

1. The resources created by the test are destroyed.
   This step happens automatically, and is the equivalent of calling `terraform destroy`.

1. Assertions are made against the provider API to verify that the resources have indeed been removed.
   If these checks fail, the test fails and reports "dangling resources".
   The code to ensure that the `scaleway_instance_server` shown above has been destroyed looks like this:

    ```go
    func instancechecks.CheckServerDestroy(tt *TestTools) resource.TestCheckFunc {
       return func(state *terraform.State) error {
            for _, rs := range s.RootModule().Resources {
                if rs.Type != "scaleway_instance_server" {
                    continue
                }
        
                instanceAPI, zone, ID, err := InstanceAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
                if err != nil {
                    return err
                }
        
                _, err = instanceAPI.GetServer(&instance.GetServerRequest{
                    ServerID: ID,
                    Zone:     zone,
                })
        
                // If no error resource still exist
                if err == nil {
                    return fmt.Errorf("Server (%s) still exists", rs.Primary.ID)
                }
        
                // Unexpected api error we return it
                if !is404Error(err) {
                    return err
                }
            } 
            return nil
           }
    }
    ```

   These functions usually test only for the resource used during a specific test.
