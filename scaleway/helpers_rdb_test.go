package scaleway

import (
	"context"
	"net"
	"reflect"
	"testing"

	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/stretchr/testify/assert"
)

func TestIsEndPointEqual(t *testing.T) {
	tests := []struct {
		name     string
		A        *rdb.EndpointPrivateNetworkDetails
		B        *rdb.EndpointPrivateNetworkDetails
		expected bool
	}{
		{
			name: "isEqualPrivateNetworkDetails",
			A: &rdb.EndpointPrivateNetworkDetails{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", ServiceIP: scw.IPNet{IPNet: net.IPNet{
				IP: net.IPv4(1, 1, 1, 1), Mask: net.CIDRMask(24, 32),
			}}, Zone: scw.ZoneFrPar1},
			B: &rdb.EndpointPrivateNetworkDetails{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", ServiceIP: scw.IPNet{IPNet: net.IPNet{
				IP: net.IPv4(1, 1, 1, 1), Mask: net.CIDRMask(24, 32),
			}}, Zone: scw.ZoneFrPar1},
			expected: true,
		},
		{
			name: "notEqualIP",
			A: &rdb.EndpointPrivateNetworkDetails{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", ServiceIP: scw.IPNet{IPNet: net.IPNet{
				IP: net.IPv4(1, 1, 1, 1), Mask: net.CIDRMask(24, 32),
			}}, Zone: scw.ZoneFrPar1},
			B: &rdb.EndpointPrivateNetworkDetails{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", ServiceIP: scw.IPNet{IPNet: net.IPNet{
				IP: net.IPv4(1, 1, 1, 2), Mask: net.CIDRMask(24, 32),
			}}, Zone: scw.ZoneFrPar1},
			expected: false,
		},
		{
			name: "notEqualZone",
			A: &rdb.EndpointPrivateNetworkDetails{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", ServiceIP: scw.IPNet{IPNet: net.IPNet{
				IP: net.IPv4(1, 1, 1, 1), Mask: net.CIDRMask(24, 32),
			}}, Zone: scw.ZoneFrPar1},
			B: &rdb.EndpointPrivateNetworkDetails{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", ServiceIP: scw.IPNet{IPNet: net.IPNet{
				IP: net.IPv4(1, 1, 1, 1), Mask: net.CIDRMask(24, 32),
			}}, Zone: scw.ZoneFrPar2},
			expected: false,
		},
		{
			name: "notEqualMask",
			A: &rdb.EndpointPrivateNetworkDetails{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", ServiceIP: scw.IPNet{IPNet: net.IPNet{
				IP: net.IPv4(1, 1, 1, 1), Mask: net.CIDRMask(25, 32),
			}}, Zone: scw.ZoneFrPar1},
			B: &rdb.EndpointPrivateNetworkDetails{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", ServiceIP: scw.IPNet{IPNet: net.IPNet{
				IP: net.IPv4(1, 1, 1, 1), Mask: net.CIDRMask(24, 32),
			}}, Zone: scw.ZoneFrPar1},
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isEndPointEqual(tt.A, tt.B))
		})
	}
}

func TestEndpointsToRemove(t *testing.T) {
	tests := []struct {
		name      string
		Endpoints []*rdb.Endpoint
		Updates   []interface{}
		Expected  map[string]bool
	}{
		{
			name: "removeAll",
			Endpoints: []*rdb.Endpoint{{
				ID: "6ba7b810-9dad-11d1-80b4-00c04fd430c1",
				PrivateNetwork: &rdb.EndpointPrivateNetworkDetails{
					PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
					ServiceIP: scw.IPNet{IPNet: net.IPNet{
						IP: net.IPv4(1, 1, 1, 1), Mask: net.CIDRMask(24, 32),
					}},
					Zone: scw.ZoneFrPar1,
				},
			}},
			Expected: map[string]bool{
				"6ba7b810-9dad-11d1-80b4-00c04fd430c1": true,
			},
		},
		{
			name: "shouldUpdatePrivateNetwork",
			Endpoints: []*rdb.Endpoint{{
				ID: "6ba7b810-9dad-11d1-80b4-00c04fd430c1",
				PrivateNetwork: &rdb.EndpointPrivateNetworkDetails{
					PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
					ServiceIP: scw.IPNet{IPNet: net.IPNet{
						IP: net.IPv4(1, 1, 1, 1), Mask: net.CIDRMask(24, 32),
					}},
					Zone: scw.ZoneFrPar1,
				},
			}},
			Updates: []interface{}{map[string]interface{}{"pn_id": "fr-par-1/6ba7b810-9dad-11d1-80b4-00c04fd430c8", "ip_net": "192.168.1.43/24"}},
			Expected: map[string]bool{
				"6ba7b810-9dad-11d1-80b4-00c04fd430c1": true,
			},
		},
		{
			name: "shouldNotUpdatePrivateNetwork",
			Endpoints: []*rdb.Endpoint{{
				ID: "6ba7b810-9dad-11d1-80b4-00c04fd430c1",
				PrivateNetwork: &rdb.EndpointPrivateNetworkDetails{
					PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
					ServiceIP: scw.IPNet{IPNet: net.IPNet{
						IP: net.IPv4(1, 1, 1, 1), Mask: net.CIDRMask(24, 32),
					}},
					Zone: scw.ZoneFrPar1,
				},
			}},
			Updates: []interface{}{map[string]interface{}{"pn_id": "fr-par-1/6ba7b810-9dad-11d1-80b4-00c04fd430c8", "ip_net": "1.1.1.1/24"}},
			Expected: map[string]bool{
				"6ba7b810-9dad-11d1-80b4-00c04fd430c1": false,
			},
		},
		{
			name:     "shouldAddPrivateNetwork",
			Updates:  []interface{}{map[string]interface{}{"pn_id": "fr-par-1/6ba7b810-9dad-11d1-80b4-00c04fd430c8", "ip_net": "1.1.1.1/24"}},
			Expected: map[string]bool{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := endpointsToRemove(tt.Endpoints, tt.Updates)
			assert.NoError(t, err)
			assert.Equal(t, tt.Expected, result)
		})
	}
}

func TestRDBPrivilegeV1SchemaUpgradeFunc(t *testing.T) {
	v0Schema := map[string]interface{}{
		"id":            "fr-par/11111111-1111-1111-1111-111111111111",
		"region":        "fr-par",
		"database_name": "database",
		"user_name":     "username",
	}
	v1Schema := map[string]interface{}{
		"id":            "fr-par/11111111-1111-1111-1111-111111111111/database/username",
		"region":        "fr-par",
		"database_name": "database",
		"user_name":     "username",
	}

	actual, err := rdbPrivilegeV1SchemaUpgradeFunc(context.Background(), v0Schema, nil)
	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(v1Schema, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", v1Schema, actual)
	}
}
