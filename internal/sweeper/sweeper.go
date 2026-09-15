package sweeper

import (
	"context"

	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

// Locality describes the scope at which a resource type lives.
type Locality string

const (
	// LocalityGlobal means the resource is not tied to a region or zone
	// (e.g. IAM resources, scoped to an organization).
	LocalityGlobal Locality = "global"
	// LocalityRegional means the resource lives in a specific region.
	LocalityRegional Locality = "regional"
)

// Filter holds the criteria used to select resources to sweep.
type Filter struct {
	// OrganizationID scopes the sweep to an organization. Used by global
	// sweepers. If empty, the provider default is used.
	OrganizationID string
	// Tags filters resources by Scaleway tags. Only resources that carry at
	// least one of the given tags are swept. Sweepers that do not support
	// tag filtering ignore this field.
	Tags []string
	// Regions is the list of regions to scan. Used by regional sweepers.
	Regions []scw.Region
	// ProjectIDs scopes the sweep to the given projects. Use ["*"] to sweep
	// across all projects. Used by regional sweepers.
	ProjectIDs []string
}

// Resource represents a resource discovered by a Sweeper that may be deleted.
type Resource struct {
	// ID is the identifier used to delete the resource.
	ID string
	// DisplayName is a human-readable name shown in progress events.
	DisplayName string
	// Region is the region the resource lives in (empty for global resources).
	Region scw.Region
	// Tags are the tags attached to the resource (when available).
	Tags []string
}

// Sweeper knows how to list and delete a specific Scaleway resource type,
// even when the resource is not managed by Terraform.
type Sweeper interface {
	// TypeName is the Terraform resource type (e.g. "scaleway_iam_api_key").
	TypeName() string
	// Locality returns the scope at which the resource lives.
	Locality() Locality
	// SupportsTagFilter reports whether the resource type can be filtered by
	// tags. When false and tags are provided, the action warns the user and
	// sweeps all matching resources regardless of tags.
	SupportsTagFilter() bool
	// List returns the resources matching the filter.
	List(ctx context.Context, m *meta.Meta, filter Filter) ([]Resource, error)
	// Delete deletes a single resource.
	Delete(ctx context.Context, m *meta.Meta, resource Resource) error
}

// All returns every registered sweeper.
func All() []Sweeper {
	return []Sweeper{
		&iamAPIKeySweeper{},
		&iamScimTokenSweeper{},
		&secretSweeper{},
		&keyManagerKeySweeper{},
	}
}

// SupportedTypes returns the Terraform type names of all registered sweepers.
func SupportedTypes() []string {
	all := All()
	types := make([]string, 0, len(all))

	for _, s := range all {
		types = append(types, s.TypeName())
	}

	return types
}
