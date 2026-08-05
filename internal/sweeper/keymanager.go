package sweeper

import (
	"context"
	"fmt"

	key_manager "github.com/scaleway/scaleway-sdk-go/api/key_manager/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	listscw "github.com/scaleway/terraform-provider-scaleway/v2/internal/list"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

// keyManagerKeySweeper sweeps scaleway_key_manager_key resources.
//
// Keys are regional and taggable: the Scaleway List API supports filtering by
// tags server-side.
type keyManagerKeySweeper struct{}

func (s *keyManagerKeySweeper) TypeName() string        { return "scaleway_key_manager_key" }
func (s *keyManagerKeySweeper) Locality() Locality      { return LocalityRegional }
func (s *keyManagerKeySweeper) SupportsTagFilter() bool { return true }

func (s *keyManagerKeySweeper) List(ctx context.Context, m *meta.Meta, filter Filter) ([]Resource, error) {
	api := key_manager.NewAPI(m.ScwClient())

	targets := listscw.RegionalProjectTargets(filter.Regions, filter.ProjectIDs)

	resources, err := listscw.FetchConcurrently(ctx, targets,
		func(ctx context.Context, target listscw.RegionalFetchTarget) ([]Resource, error) {
			req := &key_manager.ListKeysRequest{
				Region:    target.Region,
				ProjectID: &target.ProjectID,
			}

			if len(filter.Tags) > 0 {
				req.Tags = filter.Tags
			}

			resp, err := api.ListKeys(req, scw.WithContext(ctx), scw.WithAllPages())
			if err != nil {
				return nil, err
			}

			items := make([]Resource, 0, len(resp.Keys))
			for _, key := range resp.Keys {
				items = append(items, Resource{
					ID:          key.ID,
					DisplayName: key.Name,
					Region:      key.Region,
					Tags:        key.Tags,
				})
			}

			return items, nil
		},
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list key manager keys: %w", err)
	}

	return resources, nil
}

func (s *keyManagerKeySweeper) Delete(ctx context.Context, m *meta.Meta, resource Resource) error {
	api := key_manager.NewAPI(m.ScwClient())

	err := transport.RetryOn403(ctx, func() error {
		return api.DeleteKey(&key_manager.DeleteKeyRequest{
			Region: resource.Region,
			KeyID:  resource.ID,
		})
	})
	if err != nil && !httperrors.Is404(err) {
		return err
	}

	return nil
}
