# Testing the module

## Unit testing

You can test the provider, by running `make test`.

```sh
make test
```

## Acceptance testing

Acceptance test are made to test the terraform module with real API calls so they will create real resources that will be invoiced.
But in order to run faster tests and avoid bad surprise at the end of the month we shipped the project with mocks (recorded with [go-vcr](https://github.com/dnaeon/go-vcr)).

### Running the acceptance tests with mocks

By default, mocks are used during acceptance tests.

```sh
make testacc
```

### Running the acceptance tests on real resources

:warning: This will cost money.

```sh
export TF_UPDATE_CASSETTES=true
make testacc
```

It's also required to have Scaleway environment vars available:

```sh
export SCW_ACCESS_KEY=SCWXXXXXXXXXXXXXXXXX
export SCW_SECRET_KEY=XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX
export SCW_DEFAULT_PROJECT_ID=XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX
```

For testing the domain API, it will use the first available domain in your domains list. You need to have a valid domain.

You can force the test domain with an environment var:

```sh
export TF_TEST_DOMAIN=your-domain.tld
```

For testing a domain zone you can force the following environment var:

```sh
export TF_TEST_DOMAIN_ZONE=your-zone
```

To ease debugging you can also set:
```sh
export TF_LOG=DEBUG
export SCW_DEBUG=1
```

Running a single test:
```sh
TF_UPDATE_CASSETTES=true TF_LOG=DEBUG SCW_DEBUG=1 TF_ACC=1 go test ./scaleway -v -run=TestAccScalewayDataSourceRDBInstance_Basic -timeout=120m -parallel=10
```

## Compressing the cassettes

We record interactions with the Scaleway API in cassettes, which are stored in the `testdata` directory of each service.
Each wait function used in the resources will perform several requests to the API for pulling a resource state, which can lead to large cassettes.
We use a compressor to reduce the size of these cassettes once they are recorded.
By doing so, tests can run faster and the cassettes are easier to read.

To use the compressor on a given cassette, run the following command:

```sh
go run -v ./cmd/vcr-compressor internal/services/rdb/testdata/acl-basic.cassette
```
