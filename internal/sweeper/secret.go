package sweeper

import (
	"context"
	"fmt"

	secret "github.com/scaleway/scaleway-sdk-go/api/secret/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	listscw "github.com/scaleway/terraform-provider-scaleway/v2/internal/list"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

// secretSweeper sweeps scaleway_secret resources.
//
// Secrets are regional and taggable: the Scaleway List API supports filtering
// by tags server-side.
type secretSweeper struct{}

func (s *secretSweeper) TypeName() string        { return "scaleway_secret" }
func (s *secretSweeper) Locality() Locality      { return LocalityRegional }
func (s *secretSweeper) SupportsTagFilter() bool { return true }

func (s *secretSweeper) List(ctx context.Context, m *meta.Meta, filter Filter) ([]Resource, error) {
	api := secret.NewAPI(m.ScwClient())

	targets := listscw.RegionalProjectTargets(filter.Regions, filter.ProjectIDs)

	resources, err := listscw.FetchConcurrently(ctx, targets,
		func(ctx context.Context, target listscw.RegionalFetchTarget) ([]Resource, error) {
			req := &secret.ListSecretsRequest{
				Region:    target.Region,
				ProjectID: &target.ProjectID,
			}

			if len(filter.Tags) > 0 {
				req.Tags = filter.Tags
			}

			resp, err := api.ListSecrets(req, scw.WithContext(ctx), scw.WithAllPages())
			if err != nil {
				return nil, err
			}

			items := make([]Resource, 0, len(resp.Secrets))
			for _, sec := range resp.Secrets {
				items = append(items, Resource{
					ID:          sec.ID,
					DisplayName: sec.Name,
					Region:      sec.Region,
					Tags:        sec.Tags,
				})
			}

			return items, nil
		},
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	return resources, nil
}

func (s *secretSweeper) Delete(ctx context.Context, m *meta.Meta, resource Resource) error {
	api := secret.NewAPI(m.ScwClient())

	err := transport.RetryOn403(ctx, func() error {
		return api.DeleteSecret(&secret.DeleteSecretRequest{
			Region:   resource.Region,
			SecretID: resource.ID,
		}, scw.WithContext(ctx))
	})
	if err != nil && !httperrors.Is404(err) {
		return err
	}

	return nil
}
