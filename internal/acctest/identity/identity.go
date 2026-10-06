package identity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

type identity struct {
	values          map[string]any
	resourceAddress string
}

func Identity() identity {
	return identity{}
}

// GetIdentity sets the resource address to check and stores the identity values.
func (v *identity) GetIdentity(resourceAddress string) statecheck.StateCheck {
	v.resourceAddress = resourceAddress

	return newIdentityStateChecker(v)
}

type identityStateChecker struct {
	identity *identity
	base     Base
}

func newIdentityStateChecker(identity *identity) identityStateChecker {
	return identityStateChecker{
		base:     NewBase(identity.resourceAddress),
		identity: identity,
	}
}

func (vc identityStateChecker) CheckState(ctx context.Context, request statecheck.CheckStateRequest, response *statecheck.CheckStateResponse) {
	resource, ok := vc.base.ResourceFromState(request, response)
	if !ok {
		return
	}

	if resource.IdentitySchemaVersion == nil || len(resource.IdentityValues) == 0 {
		response.Error = fmt.Errorf("%s - Identity not found in state. Either the resource does not support identity or the Terraform version running the test does not support identity. (must be v1.12+)", vc.base.resourceAddress)

		return
	}

	vc.identity.values = maps.Collect(maps.All(resource.IdentityValues))
}

func (v *identity) Checks() func() map[string]knownvalue.Check {
	return func() map[string]knownvalue.Check {
		checks := make(map[string]knownvalue.Check, len(v.values))

		for k, val := range v.values {
			if val == nil {
				checks[k] = knownvalue.Null()
			} else {
				switch v := val.(type) {
				case string:
					checks[k] = knownvalue.StringExact(v)
				case json.Number:
					if i, err := v.Int64(); err == nil {
						checks[k] = knownvalue.Int64Exact(i)
					} else {
						checks[k] = knownvalue.StringExact(v.String())
					}
				default:
					checks[k] = knownvalue.StringExact(fmt.Sprintf("%v", val))
				}
			}
		}

		return checks
	}
}

type Base struct {
	resourceAddress string
}

func NewBase(resourceAddress string) Base {
	return Base{
		resourceAddress: resourceAddress,
	}
}

func (b Base) ResourceFromState(req statecheck.CheckStateRequest, resp *statecheck.CheckStateResponse) (*tfjson.StateResource, bool) {
	var resource *tfjson.StateResource

	if req.State == nil {
		resp.Error = errors.New("state is nil")

		return nil, false
	}

	if req.State.Values == nil {
		resp.Error = errors.New("state does not contain any state values")

		return nil, false
	}

	if req.State.Values.RootModule == nil {
		resp.Error = errors.New("state does not contain a root module")

		return nil, false
	}

	for _, r := range req.State.Values.RootModule.Resources {
		if b.resourceAddress == r.Address {
			resource = r

			break
		}
	}

	if resource == nil {
		resp.Error = fmt.Errorf("%s - Resource not found in state", b.resourceAddress)

		return nil, false
	}

	return resource, true
}

func (b Base) ResourceAddress() string {
	return b.resourceAddress
}
