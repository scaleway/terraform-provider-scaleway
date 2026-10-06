// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package identity

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
)

var _ querycheck.QueryResultCheck = expectNoIdentityFunc{}

type expectNoIdentityFunc struct {
	identityFunc        func() map[string]knownvalue.Check
	listResourceAddress string
}

// Asserts that the given list resource does NOT contain a resource with an identity matching
// the identity checks returned by the identityFunc.
// This query check can only be used with managed resources that support resource identity and query.
func ExpectNoIdentityFunc(resourceAddress string, identityFunc func() map[string]knownvalue.Check) querycheck.QueryResultCheck {
	return expectNoIdentityFunc{
		listResourceAddress: resourceAddress,
		identityFunc:        identityFunc,
	}
}

func (e expectNoIdentityFunc) CheckQuery(_ context.Context, req querycheck.CheckQueryRequest, resp *querycheck.CheckQueryResponse) {
	checks := e.identityFunc()

	for _, res := range req.Query {
		if e.listResourceAddress != strings.TrimPrefix(res.Address, "list.") {
			continue
		}

		identityMatches := true

		for attr, check := range checks {
			actualIdentityVal, ok := res.Identity[attr]
			if !ok {
				identityMatches = false

				break
			}

			if err := check.CheckValue(actualIdentityVal); err != nil {
				identityMatches = false

				break
			}
		}

		if identityMatches {
			errCollection := make([]error, 0, len(checks)+2)

			errCollection = append(errCollection, errors.New("identity with the following attributes was found, but was expected to be absent"))

			for attr, check := range checks {
				errCollection = append(errCollection, fmt.Errorf("attribute %q: %s", attr, check))
			}

			errCollection = append(errCollection, fmt.Errorf("address: %s", e.listResourceAddress))
			resp.Error = errors.Join(errCollection...)

			return
		}
	}
}
