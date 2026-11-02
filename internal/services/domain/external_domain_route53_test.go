package domain_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func requireAWSCredsForExternalDomain(t *testing.T, dnsZone string) {
	t.Helper()

	if !*acctest.UpdateCassettes {
		return
	}

	ctx := t.Context()

	client, err := newRoute53Client(ctx)
	if err != nil {
		t.Skipf("Test skipped when recording: no usable AWS credentials for Route53 (%v)", err)
	}

	_, err = route53HostedZoneID(ctx, client, dnsZone)
	if err != nil {
		t.Skipf("Test skipped when recording: cannot access Route53 zone %q (%v)", dnsZone, err)
	}
}

func newRoute53Client(ctx context.Context) (*route53.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	return route53.NewFromConfig(cfg), nil
}

func route53HostedZoneID(ctx context.Context, client *route53.Client, zoneName string) (string, error) {
	name := strings.TrimSuffix(zoneName, ".") + "."

	out, err := client.ListHostedZonesByName(ctx, &route53.ListHostedZonesByNameInput{
		DNSName:  aws.String(name),
		MaxItems: aws.Int32(1),
	})
	if err != nil {
		return "", err
	}

	if len(out.HostedZones) == 0 || aws.ToString(out.HostedZones[0].Name) != name {
		return "", fmt.Errorf("route53 hosted zone %q not found", zoneName)
	}

	id := aws.ToString(out.HostedZones[0].Id)

	return strings.TrimPrefix(id, "/hostedzone/"), nil
}

func route53ChallengeRecordName(subdomain, dnsZone string) string {
	return "_scaleway-challenge." + subdomain + "." + strings.TrimSuffix(dnsZone, ".") + "."
}

func changeRoute53ChallengeTXT(ctx context.Context, client *route53.Client, zoneID, subdomain, dnsZone, token string, action types.ChangeAction) error {
	name := route53ChallengeRecordName(subdomain, dnsZone)

	_, err := client.ChangeResourceRecordSets(ctx, &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{
				{
					Action: action,
					ResourceRecordSet: &types.ResourceRecordSet{
						Name: aws.String(name),
						Type: types.RRTypeTxt,
						TTL:  aws.Int64(300),
						ResourceRecords: []types.ResourceRecord{
							{Value: aws.String(`"` + token + `"`)},
						},
					},
				},
			},
		},
	})

	return err
}

// testAccPublishExternalDomainChallengeTXT reads validation_token from state and
// upserts the ownership TXT on Route53. No-op when replaying cassettes.
func testAccPublishExternalDomainChallengeTXT(t *testing.T, resourceName, subdomain, dnsZone string) resource.TestCheckFunc {
	t.Helper()

	return func(s *terraform.State) error {
		if !*acctest.UpdateCassettes {
			return nil
		}

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		token := rs.Primary.Attributes["validation_token"]
		if token == "" {
			return fmt.Errorf("validation_token is empty on %s", resourceName)
		}

		ctx := t.Context()

		client, err := newRoute53Client(ctx)
		if err != nil {
			return err
		}

		zoneID, err := route53HostedZoneID(ctx, client, dnsZone)
		if err != nil {
			return err
		}

		if err := changeRoute53ChallengeTXT(ctx, client, zoneID, subdomain, dnsZone, token, types.ChangeActionUpsert); err != nil {
			return fmt.Errorf("upsert route53 challenge TXT: %w", err)
		}

		// Fixed DNS propagation window before the validation waiter starts (record-only).
		time.Sleep(15 * time.Second) // lintignore:R018

		t.Cleanup(func() {
			cleanupCtx := context.Background()
			if err := changeRoute53ChallengeTXT(cleanupCtx, client, zoneID, subdomain, dnsZone, token, types.ChangeActionDelete); err != nil {
				t.Logf("cleanup route53 challenge TXT: %v", err)
			}
		})

		return nil
	}
}
