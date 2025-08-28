package acctest_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/scaleway/scaleway-sdk-go/vcr"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
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

var ludacrisBodyRequest = `{
  "payload": {
    "artists": [
		{
			"name": "Ludacris",
			"age": 45,
			"songs": ["Ludacris", "Ludacris", "Ludacris"]
		}
	}
  }
}
`

var jdillaBodyCassette = `{
  "payload": {
    "artists": [
		{
			"name": "Jdilla",
			"age": 54,
			"songs": ["this", "is", "jdilla"]
		}
	]
  }
}
`

var requestInstanceSettings = `{
  "settings": [
    {
      "name": "max_connections",
      "value": "200"
    },
    {
      "name": "max_parallel_workers",
      "value": "2"
    },
    {
      "name": "effective_cache_size",
      "value": "1300"
    },
    {
      "name": "maintenance_work_mem",
      "value": "150"
    },
    {
      "name": "max_parallel_workers_per_gather",
      "value": "2"
    },
    {
      "name": "work_mem",
      "value": "4"
    }
  ]
}
`

var cassetteInstanceSettings = `{
  "settings": [
    {
      "name": "maintenance_work_mem",
      "value": "150"
    },
    {
      "name": "effective_cache_size",
      "value": "1300"
    },
    {
      "name": "work_mem",
      "value": "4"
    },
    {
      "name": "max_parallel_workers",
      "value": "2"
    },
    {
      "name": "max_parallel_workers_per_gather",
      "value": "2"
    },
    {
      "name": "max_connections",
      "value": "200"
    }
  ]
}
`

var objectBodyRequest = `{
  "Id": "MyPolicy",
  "Statement": [
    {
      "Action": [
        "s3:ListBucket",
        "s3:GetObject"
      ],
      "Effect": "Allow",
      "Principal": {
        "SCW": "*"
      },
      "Resource": [
        "tf-tests-scw-obp-basic-4713290580220176511",
        "tf-tests-scw-obp-basic-4713290580220176511/*"
      ],
      "Sid": "GrantToEveryone"
    },
    {
      "Action": [
        "s3:ListBucket",
        "s3:GetObject"
      ],
      "Effect": "Allow",
      "Principal": {
        "SCW": "*"
      },
      "Sid": "GrantToEveryone",
	  "project_id": "1234567890"
    }
  ],
  "Version": "2012-10-17"
}
`

var objectBodyCassette = `{
  "Id": "MyPolicy",
  "Statement": [
    {
      "Action": [
        "s3:ListBucket",
        "s3:GetObject"
      ],
      "Effect": "Allow",
      "Principal": {
        "SCW": "*"
      },
      "Sid": "GrantToEveryone",
      "project_id": "9876543210"
    },
    {
      "Action": [
        "s3:ListBucket",
        "s3:GetObject"
      ],
      "Effect": "Allow",
      "Principal": {
        "SCW": "*"
      },
      "Sid": "GrantToEveryone",
      "Resource": [
        "tf-tests-scw-obp-basic-1234567890",
        "tf-tests-scw-obp-basic-1234567890/*"
      ]
    }
  ],
  "Version": "2012-10-17"
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
	// bar does not match foo
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
	// bar matches bar
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
	{
		requestBody: newRequest(http.MethodPost, "https://api.scaleway.com/iam/v1alpha1/users", strings.NewReader(requestInstanceSettings)),
		cassetteBody: &cassette.Request{
			URL:           "https://api.scaleway.com/iam/v1alpha1/users",
			Method:        http.MethodPost,
			Body:          cassetteInstanceSettings,
			ContentLength: int64(len(cassetteInstanceSettings)),
		},
		shouldMatch: true,
	},
	// complex slice of maps case
	{
		requestBody: newRequest(http.MethodPost, "https://api.scaleway.com/iam/v1alpha1/policies", strings.NewReader(objectBodyRequest)),
		cassetteBody: &cassette.Request{
			URL:           "https://api.scaleway.com/iam/v1alpha1/policies",
			Method:        http.MethodPost,
			Body:          objectBodyCassette,
			ContentLength: int64(len(objectBodyCassette)),
		},
		shouldMatch: true,
	},
	// compare slices of different integers
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
	// compare slices of same integers in different order
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
	// compare slices of slices of strings in different order
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
	// compare slices of strings in different order
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
	// ludacris does not match jdilla
	{
		requestBody: newRequest(http.MethodPost, "https://api.scaleway.com/iam/v1alpha1/users", strings.NewReader(ludacrisBodyRequest)),
		cassetteBody: &cassette.Request{
			URL:           "https://api.scaleway.com/iam/v1alpha1/users",
			Method:        http.MethodPost,
			Body:          jdillaBodyCassette,
			ContentLength: int64(len(jdillaBodyCassette)),
		},
		shouldMatch: false,
	},
}

func TestCassetteMatcher(t *testing.T) {
	for i, test := range testBodyMatcherCases {
		shouldMatch := vcr.CassetteMatcher(test.requestBody, *test.cassetteBody)
		if shouldMatch != test.shouldMatch {
			t.Errorf("test %d: expected %v, got %v", i, test.shouldMatch, shouldMatch)
			t.Errorf("requestBody: %s", test.requestBody.Body)
			t.Errorf("cassetteBody: %s", test.cassetteBody.Body)
		}
	}
}
