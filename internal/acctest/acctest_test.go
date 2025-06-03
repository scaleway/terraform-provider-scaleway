package acctest_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
)

var barMemberCreationBody = `{
  "organization_id": "6867048b-fe12-4e96-835e-41c79a39604b",
  "tags": [
    "bar"
  ],
  "member": {
    "email": "bar@scaleway.com",
    "send_password_email": false,
    "send_welcome_email": false,
    "username": "bar",
    "password": "",
    "first_name": "",
    "last_name": "",
    "phone_number": "",
    "locale": ""
  }
}
`

var fooMemberCreationBody = `{
  "organization_id": "6867048b-fe12-4e96-835e-41c79a39604b",
  "tags": [
    "foo"
  ],
  "member": {
    "email": "foo@scaleway.com",
    "send_password_email": false,
    "send_welcome_email": false,
    "username": "foo",
    "password": "",
    "first_name": "",
    "last_name": "",
    "phone_number": "",
    "locale": ""
  }
}
`

var secretPatchBodyCassette = `{
  "environment_variables": {
    "foo": "bar"
  },
  "privacy": "unknown_privacy",
  "protocol": "unknown_protocol",
  "secret_environment_variables": [
    {
      "key": "foo_secret",
      "value": "bar_secret"
    },
    {
      "key": "test_secret",
      "value": "updated_secret"
    },
    {
      "key": "first_secret",
      "value": null
    }
  ],
  "http_option": "unknown_http_option",
  "sandbox": "unknown_sandbox"
}
`

var secretPatchBodyRequest = `{
  "environment_variables": {
    "foo": "bar"
  },
  "privacy": "unknown_privacy",
  "protocol": "unknown_protocol",
  "secret_environment_variables": [
    {
      "key": "first_secret",
      "value": null
    },
    {
      "key": "foo_secret",
      "value": "bar_secret"
    },
    {
      "key": "test_secret",
      "value": "updated_secret"
    }
  ],
  "http_option": "unknown_http_option",
  "sandbox": "unknown_sandbox"
}
`

var integertestBodyRequest = `{
  "akey": "avalue",
   "integers": [
     1,
     2,
     3
   ]
}
`

var integertestBodyCassette = `{
  "akey": "avalue",
   "integers": [
     4,
     5,
     6
   ]
}
`

var integerBodyRequestOutOfOrder = `{
  "akey": "avalue",
   "integers": [
     2,
     1,
     3
   ]
}
`

var nestedSliceOfSlicesRequest = `{
  "akey": "avalue",
  "nested_lists": [
    [
      "1",
      "2",
      "3"
    ],
    [
      "4",
      "5",
      "6"
    ]
  }
}
`

var nestedSliceOfSlicesCassette = `{
  "akey": "avalue",
  "nested_slice_of_slices": {
	"integers_array": [
   		[
   		  "4",
   		  "5",
   		  "6"
   		],
   		[
   		  "1",
   		  "2",
   		  "3"
   		]
	]
  },
}
`

var simpleSliceOfStringsRequest = `{
  "strings": [
    "1",
    "2",
    "3"
  ]
}
`
var simpleSliceOfStringsCassette = `{
  "strings": [
    "3",
    "2",
    "1"
  ]
}
`

// we don't use httptest.NewRequest because it does not set the GetBody func
func newRequest(method, url string, body io.Reader) *http.Request {
	req, err := http.NewRequestWithContext(context.Background(), method, url, body)
	if err != nil {
		panic(err) // lintignore: R009
	}

	return req
}

var testBodyMatcherCases = []struct {
	requestBody  *http.Request
	cassetteBody *cassette.Request
	shouldMatch  bool
}{
	// create bar compare with foo
	{
		requestBody: newRequest(http.MethodPost, "https://api.scaleway.com/iam/v1alpha1/users", strings.NewReader(barMemberCreationBody)),
		cassetteBody: &cassette.Request{
			URL:           "https://api.scaleway.com/iam/v1alpha1/users",
			Method:        http.MethodPost,
			Body:          fooMemberCreationBody,
			ContentLength: int64(len(fooMemberCreationBody)),
		},
		shouldMatch: false,
	},
	// create bar compare with bar
	{
		requestBody: newRequest(http.MethodPost, "https://api.scaleway.com/iam/v1alpha1/users", strings.NewReader(barMemberCreationBody)),
		cassetteBody: &cassette.Request{
			URL:           "https://api.scaleway.com/iam/v1alpha1/users",
			Method:        http.MethodPost,
			Body:          barMemberCreationBody,
			ContentLength: int64(len(barMemberCreationBody)),
		},
		shouldMatch: true,
	},
	// simple http get
	{
		requestBody: newRequest(http.MethodGet, "https://api.scaleway.com/iam/v1alpha1/users/6867048b-fe12-4e96-835e-41c79a39604b", nil),
		cassetteBody: &cassette.Request{
			URL:           "https://api.scaleway.com/iam/v1alpha1/users/6867048b-fe12-4e96-835e-41c79a39604b",
			Method:        http.MethodGet,
			Body:          "",
			ContentLength: 0,
		},
		shouldMatch: true,
	},
	// patch secret with nested slices of map[string]interface{} in different order
	// we cannot user deep equal because the order of the slices is different although the values are the same
	// it is not possible to sort them because they are not comparable (map[string]interface{})
	{
		requestBody: newRequest(http.MethodPatch, "https://api.scaleway.com/secrets/v1/secrets/123", strings.NewReader(secretPatchBodyRequest)),
		cassetteBody: &cassette.Request{
			URL:           "https://api.scaleway.com/secrets/v1/secrets/123",
			Method:        http.MethodPatch,
			Body:          secretPatchBodyCassette,
			ContentLength: int64(len(secretPatchBodyCassette)),
		},
		shouldMatch: true,
	},
	// compare nested slices of different integers
	{
		requestBody: newRequest(http.MethodPost, "https://api.scaleway.com/iam/v1alpha1/users", strings.NewReader(integertestBodyRequest)),
		cassetteBody: &cassette.Request{
			URL:           "https://api.scaleway.com/iam/v1alpha1/users",
			Method:        http.MethodPost,
			Body:          integertestBodyCassette,
			ContentLength: int64(len(integertestBodyCassette)),
		},
		shouldMatch: false,
	},
	// compare nested slices of same integers in different order
	{
		requestBody: newRequest(http.MethodPost, "https://api.scaleway.com/iam/v1alpha1/users", strings.NewReader(integerBodyRequestOutOfOrder)),
		cassetteBody: &cassette.Request{
			URL:           "https://api.scaleway.com/iam/v1alpha1/users",
			Method:        http.MethodPost,
			Body:          integertestBodyRequest,
			ContentLength: int64(len(integertestBodyRequest)),
		},
		shouldMatch: true,
	},
	// compare nested slices of slices of strings
	{
		requestBody: newRequest(http.MethodPost, "https://api.scaleway.com/iam/v1alpha1/users", strings.NewReader(nestedSliceOfSlicesRequest)),
		cassetteBody: &cassette.Request{
			URL:           "https://api.scaleway.com/iam/v1alpha1/users",
			Method:        http.MethodPost,
			Body:          nestedSliceOfSlicesCassette,
			ContentLength: int64(len(nestedSliceOfSlicesCassette)),
		},
		shouldMatch: false,
	},
	{
		requestBody: newRequest(http.MethodPost, "https://api.scaleway.com/iam/v1alpha1/users", strings.NewReader(nestedSliceOfSlicesRequest)),
		cassetteBody: &cassette.Request{
			URL:           "https://api.scaleway.com/iam/v1alpha1/users",
			Method:        http.MethodPost,
			Body:          nestedSliceOfSlicesRequest,
			ContentLength: int64(len(nestedSliceOfSlicesRequest)),
		},
		shouldMatch: true,
	},
	// compare simple slice of strings
	{
		requestBody: newRequest(http.MethodPost, "https://api.scaleway.com/iam/v1alpha1/users", strings.NewReader(simpleSliceOfStringsRequest)),
		cassetteBody: &cassette.Request{
			URL:           "https://api.scaleway.com/iam/v1alpha1/users",
			Method:        http.MethodPost,
			Body:          simpleSliceOfStringsCassette,
			ContentLength: int64(len(simpleSliceOfStringsCassette)),
		},
		shouldMatch: true,
	},
}

func TestCassetteMatcher(t *testing.T) {
	for i, test := range testBodyMatcherCases {
		shouldMatch := acctest.CassetteMatcher(test.requestBody, *test.cassetteBody)
		if shouldMatch != test.shouldMatch {
			t.Errorf("test %d: expected %v, got %v", i, test.shouldMatch, shouldMatch)
		}
	}
}
