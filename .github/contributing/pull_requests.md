# Pull Requests

## Pull Request Guidelines

The [maintainers](MAINTAINERS.md) of `scaleway-sdk-go` use a "Let's Get This Merged" (LGTM) message in the pull request to note that the commits are ready to merge.
After one or more maintainer states LGTM, we will merge.
If you have questions or comments on your code, feel free to correct these in your branch through new commits.


The goal of the following guidelines is to have Pull Requests (PRs) that are fairly easy to review and comprehend, and code that is easy to maintain in the future.

- **Pull Request title should be clear** on what is being fixed or added to the code base.
  If you are addressing an open issue, please start the title with "fix: #XXX" or "feature: #XXX"
- **Keep it readable for human reviewers** and prefer a subset of functionality (code) with tests and documentation over delivering them separately
- **Don't forget commenting code** to help reviewers understand and to keep [our Go Report Card](https://goreportcard.com/report/github.com/scaleway/scaleway-sdk-go) at A+
- **Notify Work In Progress PRs** by prefixing the title with `[WIP]`
- **Please, keep us updated.**
  We will try our best to merge your PR, but please notice that PRs may be closed after 30 days of inactivity.

Your pull request should be rebased against the current master branch. Please do not merge
the current master branch in with your topic branch, nor use the Update Branch button provided
by GitHub on the pull request page.

Keep in mind only the **Pull Request Title** will be used as commit message as we stash all commits on merge.


We appreciate direct contributions to the provider codebase.
Here's what to expect:

* For pull requests that follow the guidelines, we will proceed to reviewing and merging, following the provider team's review schedule.
  There may be some internal or community discussion needed before we can complete this.
* Pull requests that don't follow the guidelines will be commented with what they're missing.
  The person who submits the pull request or another community member will need to address those requests before they move forward.

## Pull Request Lifecycle

1. [Fork the GitHub repository](https://help.github.com/en/articles/fork-a-repo), modify the code, and [create a pull request](https://help.github.com/en/articles/creating-a-pull-request-from-a-fork).
   You are welcome to submit your pull request for commentary or review before it is fully completed by creating a [draft pull request](https://help.github.com/en/articles/about-pull-requests#draft-pull-requests) or adding `[WIP]` to the beginning of the pull request title.
   Please include specific questions or items you'd like feedback on.

1. Once you believe your pull request is ready to be reviewed, ensure the pull request is not a draft pull request by [marking it ready for review](https://help.github.com/en/articles/changing-the-stage-of-a-pull-request) or removing `[WIP]` from the pull request title if necessary, and a maintainer will review it.
   Follow [the checklists below](#checklists-for-contribution) to help ensure that your contribution can be easily reviewed and potentially merged.

1. One of Terraform's provider team members will look over your contribution and either approve it or provide comments letting you know if there is anything left to do.
   We do our best to keep up with the volume of PRs waiting for review, but it may take some time depending on the complexity of the work.

1. Once all outstanding comments and checklist items have been addressed, your contribution will be merged!
   Merged PRs will be included in the next Terraform release.
   The provider team takes care of updating the CHANGELOG as they merge.

1. In some cases, we might decide that a PR should be closed without merging.
   We'll make sure to provide clear reasoning when this happens.

## Checklists for Contribution

There are several different kinds of contribution, each of which has its own standards for a speedy review.
The following sections describe guidelines for each type of contribution.

### Documentation Update

The [Terraform Scaleway Provider's website source][website] is in this repository along with the code and tests.
Below are some common items that will get flagged during documentation reviews:

- [ ] __Reasoning for Change__: Documentation updates should include an explanation for why the update is needed.
- [ ] __Prefer Scaleway Documentation__: Documentation about Scaleway service features and valid argument values that are likely to update over time should link to Scaleway service user guides and API references where possible.
  You can find all those document at <https://developers.scaleway.com>
- [ ] __Large Example Configurations__: Example Terraform configuration that includes multiple resource definitions should be added to the repository `examples` directory instead of an individual resource documentation page.
  Each directory under `examples` should be self-contained to call `terraform apply` without special configuration.
- [ ] __Terraform Configuration Language Features__: Individual resource documentation pages and examples should refrain from highlighting particular Terraform configuration language syntax workarounds or features such as `variable`, `local`, `count`, and built-in functions.

### Enhancement/Bugfix to a Resource

Working on existing resources is a great way to get started as a Terraform contributor because you can work within existing code and tests to get a feel for what to do.

In addition to the below checklist, please see the [Common Review Items](common_reviews_items.md) sections for more specific coding and testing guidelines.

 - [ ] __Acceptance test coverage of new behavior__: Existing resources each have a set of [acceptance tests][acctests] covering their functionality.
   These tests should exercise all the behavior of the resource.
   Whether you are adding something or fixing a bug, the idea is to have an acceptance test that fails if your code were to be removed.
   Sometimes it is sufficient to "enhance" an existing test by adding an assertion or tweaking the config that is used, but it's often better to add a new test.
   You can copy/paste an existing test and follow the conventions you see there, modifying the test to exercise the behavior of your code.
 - [ ] __Documentation updates__: If your code makes any changes that need to be documented, you should include those doc updates in the same PR.
   This includes things like new resource attributes or changes in default values.
   The [Terraform website][website] source is in this repo and includes instructions for getting a local copy of the site up and running if you'd like to preview your changes.
 - [ ] __Well-formed Code__: Do your best to follow existing conventions you see in the codebase, and ensure your code is formatted with `goimports`.
   (The Travis CI build will fail if `goimports` has not been run on incoming code.)
   The PR reviewers can help out on this front, and may provide comments with suggestions on how to improve the code.
 - [ ] __Vendor additions__: Create a separate PR if you are updating the vendor folder.
   This is to avoid conflicts as the vendor versions tend to be fast-moving targets.
   We will plan to merge the PR with this change first.

#### Adding Resource Import Support

Adding import support for Terraform resources will allow existing infrastructure to be managed within Terraform.
This type of enhancement generally requires a small to moderate amount of code changes.

Comprehensive code examples and information about resource import support can be found in the [Extending Terraform documentation](https://www.terraform.io/docs/extend/resources/import.html).

In addition to the below checklist and the items noted in the Extending Terraform documentation, please see the [Common Review Items](#common-review-items) sections for more specific coding and testing guidelines.

- [ ] _Resource Code Implementation_: In the resource code (e.g. `scaleway/resource_scaleway_service_thing.go`), implementation of `Importer` `State` function
- [ ] _Resource Acceptance Testing Implementation_: In the resource acceptance testing (e.g. `scaleway/resource_scaleway_service_thing_test.go`), implementation of `TestStep`s with `ImportState: true`
- [ ] _Resource Documentation Implementation_: In the resource documentation (e.g. `website/docs/r/service_thing.html.markdown`), addition of `Import` documentation section at the bottom of the page

#### Adding Resource Name Generation Support

Terraform Scaleway Provider resources can use shared logic to support and test name generation, where the operator can choose between an expected naming value, a generated naming value with a prefix, or a fully generated name.

Implementing name generation support for Terraform Scaleway Provider resources requires the following, each with its own section below:

- [ ] _Resource Name Generation Code Implementation_: In the resource code (e.g. `scaleway/resource_instance_server.go`), implementation of `name_prefix` attribute, along with handling in `Create` function.
- [ ] _Resource Name Generation Testing Implementation_: In the resource acceptance testing (e.g. `scaleway/resource_scaleway_service_thing_test.go`), implementation of new acceptance test functions and configurations to exercise new naming logic.
- [ ] _Resource Name Generation Documentation Implementation_: In the resource documentation (e.g. `website/docs/r/service_thing.html.markdown`), addition of `name_prefix` argument and update of `name` argument description.

##### Resource Code Generation Documentation Implementation

- In the resource documentation (e.g. `website/docs/r/service_thing.html.markdown`), add the following to the arguments reference:

```markdown
* `name_prefix` - (Optional) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
```

- Adjust the existing `name` argument reference to ensure its denoted as `Optional`, includes a mention that it can be generated, and that it conflicts with `name_prefix`:

```markdown
* `name` - (Optional) Name of the thing. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`.
```

#### New Resource

Implementing a new resource is a good way to learn more about how Terraform interacts with upstream APIs.
There are plenty of examples to draw from in the existing resources, but you still get to implement something completely new.

In addition to the below checklist, please see the [Common Review Items](#common-review-items) sections for more specific coding and testing guidelines.

 - [ ] __Minimal LOC__: It's difficult for both the reviewer and author to go through long feedback cycles on a big PR with many resources.
   We ask you to only submit **1 resource at a time**.
 - [ ] __Acceptance tests__: New resources should include acceptance tests covering their behavior.
   See [Writing Acceptance Tests](#writing-acceptance-tests) below for a detailed guide on how to approach these.
 - [ ] __Resource Naming__: Resources should be named `scaleway_<service>_<name>`, using underscores (`_`) as the separator.
   Resources are namespaced with the service name to allow easier searching of related resources, to align the resource naming with the service for [Customizing Endpoints](https://www.terraform.io/docs/providers/scaleway/guides/custom-service-endpoints.html#available-endpoint-customizations), and to prevent future conflicts with new Scaleway services/resources.
   For reference:

   - `service` is the Scaleway short service name that matches the entry in `endpointServiceNames` (created via the [New Service](#new-service) section)
   - `name` represents the conceptual infrastructure represented by the create, read, update, and delete methods of the service API.
     It should be a singular noun.
     For example, in an API that has methods such as `CreateThing`, `DeleteThing`, `DescribeThing`, and `ModifyThing` the name of the resource would end in `_thing`.

 - [ ] __Arguments_and_Attributes__: The HCL for arguments and attributes should mimic the types and structs presented by the Scaleway API.
   API arguments should be converted from `CamelCase` to `camel_case`.
 - [ ] __Documentation__: Each resource gets a page in the Terraform documentation.
   The [Terraform website][website] source is in this repo and includes instructions for getting a local copy of the site up and running if you'd like to preview your changes.
   For a resource, you'll want to add a new file in the appropriate place and add a link to the sidebar for that page.
 - [ ] __Well-formed Code__: Do your best to follow existing conventions you see in the codebase, and ensure your code is formatted with `goimports`.
   (The Travis CI build will fail if `goimports` has not been run on incoming code.)
   The PR reviewers can help out on this front, and may provide comments with suggestions on how to improve the code.
 - [ ] __Vendor updates__: Create a separate PR if you are adding to the vendor folder.
   This is to avoid conflicts as the vendor versions tend to be fast-moving targets.
   We will plan to merge the PR with this change first.

#### New Service

Implementing a new Scaleway service gives Terraform the ability to manage resources in a whole new API.
It's a larger undertaking, but brings major new functionality into Terraform.

- [ ] __Service Client__: Before new resources are submitted, we request a separate pull request containing just the new Scaleway Go SDK service client.
  Doing so will pull the Scaleway Go SDK service code into the project at the current version.
  Since the Scaleway Go SDK is updated frequently, these pull requests can easily have merge conflicts or be out of date.
  The maintainers prioritize reviewing and merging these quickly to prevent those situations.

  To add the Scaleway Go SDK service client:

  - In `scaleway/provider.go` Add a new service entry to `endpointServiceNames`. 
    This service name should match the Scaleway sdk go or Scaleway CLI service name.
  - In `scaleway/config.go`: Add a new import for the Scaleway Go SDK code. e.g. `github.com/scaleway/scaleway-sdk-go/service/quicksight`
  - In `scaleway/config.go`: Add a new `{SERVICE}conn` field to the `ScalewayClient` struct for the service client.
    The service name should match the name in `endpointServiceNames`. e.g. `k8sconn *k8s.K8S`
  - In `scaleway/config.go`: Create the new service client in the `{SERVICE}conn` field in the `ScalewayClient` instantiation within `Client()`.
    e.g. `k8sconn: k8s.New(sess.Copy(&scaleway.Config{Endpoint: scaleway.String(c.Endpoints["k8s"])})),`
  - In `website/allowed-subcategories.txt`: Add a name acceptable for the documentation navigation.
  - In `website/docs/guides/custom-service-endpoints.html.md`: Add the service name in the list of customizable endpoints.
  - In `.hashibot.hcl`: Add the new service to automated issue and pull request labeling. e.g. with the `quicksight` service

  ```hcl
  behavior "regexp_issue_labeler_v2" "service_labels" {
    # ... other configuration ...

    label_map = {
      # ... other services ...
      "service/k8s" = [
        "scaleway_k8s_",
      ],
      # ... other services ...
    }
  }

  behavior "pull_request_path_labeler" "service_labels"
    # ... other configuration ...

    label_map = {
      # ... other services ...
      "service/quicksight" = [
        "**/*_k8s_*",
        "**/k8s_*",
      ],
      # ... other services ...
    }
  }
  ```

  - Run the following then submit the pull request:

  ```sh
  go test ./scaleway
  go mod tidy
  go mod vendor
  ```

- [ ] __Initial Resource__: Some services can be big and it can be difficult for both reviewer & author to go through long feedback cycles on a big PR with many resources.
  Often feedback items in one resource will also need to be applied in other resources.
  We prefer you to submit the necessary minimum in a single PR, ideally **just the first resource** of the service.

The initial resource and changes afterwards should follow the other sections of this guide as appropriate.

