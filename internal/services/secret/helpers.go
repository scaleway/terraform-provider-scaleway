package secret

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	secret "github.com/scaleway/scaleway-sdk-go/api/secret/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

const (
	defaultSecretTimeout = 5 * time.Minute
)

var ErrCannotDeleteProtectedSecret = errors.New("cannot delete a protected secret")

// newAPIWithRegion returns a new Secret API and the region for a Create request
func newAPIWithRegion(d *schema.ResourceData, m any) (*secret.API, scw.Region, error) {
	api := secret.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

// newAPIWithRegionAndDefault returns a new Secret API and the region for a Create request
func newAPIWithRegionOptionalProjectIDAndDefault(d *schema.ResourceData, m any, defaultRegion scw.Region) (*secret.API, scw.Region, *string, error) {
	api := secret.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegionWithDefault(d, m, defaultRegion)
	if err != nil {
		return nil, "", nil, err
	}

	var projectIDPtr *string

	projectID, _, err := meta.ExtractProjectID(d, m)
	if err == nil {
		projectIDPtr = &projectID
	}

	return api, region, projectIDPtr, nil
}

// newAPIWithRegionAndProjectID returns a new Secret API, with region and projectID
func newAPIWithRegionAndProjectID(d *schema.ResourceData, m any) (*secret.API, scw.Region, *string, error) {
	api := secret.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", nil, err
	}

	var projectIDPtr *string

	projectID, _, err := meta.ExtractProjectID(d, m)
	if err == nil {
		projectIDPtr = &projectID
	}

	return api, region, projectIDPtr, nil
}

// NewAPIWithRegionAndID returns a Secret API with locality and ID extracted from the state
func NewAPIWithRegionAndID(m any, id string) (*secret.API, scw.Region, string, error) {
	api := secret.NewAPI(meta.ExtractScwClient(m))

	region, id, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}

	return api, region, id, nil
}

// NewVersionAPIWithRegionAndID returns a Secret API with locality and Nested ID extracted from the state
func NewVersionAPIWithRegionAndID(m any, id string) (*secret.API, scw.Region, string, string, error) {
	region, id, revision, err := locality.ParseLocalizedNestedID(id)
	if err != nil {
		return nil, "", "", "", err
	}

	api := secret.NewAPI(meta.ExtractScwClient(m))

	return api, scw.Region(region), id, revision, nil
}

// CreateSecretVersion creates a new version of the secret referenced by secretID,
// containing value. secretID can be a UUID or a regional ID (region/UUID); when
// it is a raw UUID, the client default region is used. The secret itself must
// already exist and is not managed by this function: the created version is
// removed when the secret is deleted.
// This function is specifically used from other resources to allow setting
// a sensitive value in a secret instead of in the terraform state.
func CreateSecretVersion(ctx context.Context, m any, secretID string, value string, description string) (*secret.SecretVersion, error) {
	client := meta.ExtractScwClient(m)

	region, id, err := regional.ResolveID(secretID, "", client)
	if err != nil {
		return nil, err
	}

	api := secret.NewAPI(client)

	version, err := transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(_ context.Context) (*secret.SecretVersion, error) {
		return api.CreateSecretVersion(&secret.CreateSecretVersionRequest{
			Region:      region,
			SecretID:    id,
			Data:        []byte(value),
			Description: &description,
		}, scw.WithContext(ctx))
	})
	if err != nil {
		return nil, fmt.Errorf("error creating a new version of the secret %s: %w", secretID, err)
	}

	return version, nil
}

// updateSecretProtection sets the protected value of a secret to requested one.
func updateSecretProtection(api *secret.API, region scw.Region, secretID string, protected bool) error {
	s, err := api.GetSecret(&secret.GetSecretRequest{
		Region:   region,
		SecretID: secretID,
	})
	if err != nil {
		return err
	}

	if s.Protected == protected {
		return nil
	}

	if protected {
		_, err = api.ProtectSecret(&secret.ProtectSecretRequest{
			Region:   region,
			SecretID: secretID,
		})
		if err != nil {
			return fmt.Errorf("failed to protect secret %s: %w", secretID, err)
		}
	} else {
		_, err = api.UnprotectSecret(&secret.UnprotectSecretRequest{
			Region:   region,
			SecretID: secretID,
		})
		if err != nil {
			return fmt.Errorf("failed to unprotect secret %s: %w", secretID, err)
		}
	}

	return nil
}

func expandEphemeralPolicy(rawSchemaPolicy any) (*secret.EphemeralPolicy, error) {
	rawList := rawSchemaPolicy.([]any)
	if len(rawList) != 1 {
		return nil, fmt.Errorf("expected 1 policy, found %d", len(rawList))
	}

	rawPolicy := rawList[0].(map[string]any)

	ttl, err := types.ExpandDuration(rawPolicy["ttl"])
	if err != nil {
		return nil, fmt.Errorf("error parsing ttl: %w", err)
	}

	policy := &secret.EphemeralPolicy{
		ExpiresOnceAccessed: types.ExpandBoolPtr(rawPolicy["expires_once_accessed"]),
		Action:              secret.EphemeralPolicyAction(rawPolicy["action"].(string)),
	}

	if ttl != nil {
		policy.TimeToLive = scw.NewDurationFromTimeDuration(*ttl)
	}

	return policy, nil
}

func flattenEphemeralPolicy(policy *secret.EphemeralPolicy) []map[string]any {
	if policy == nil {
		return nil
	}

	policyElem := map[string]any{}
	if policy.TimeToLive != nil {
		policyElem["ttl"] = types.FlattenDuration(policy.TimeToLive.ToTimeDuration())
	}

	if policy.ExpiresOnceAccessed != nil {
		policyElem["expires_once_accessed"] = types.FlattenBoolPtr(policy.ExpiresOnceAccessed)
	}

	policyElem["action"] = policy.Action

	return []map[string]any{policyElem}
}

func setSecretState(d *schema.ResourceData, secret *secret.Secret, versions *secret.ListSecretVersionsResponse) {
	_ = d.Set("name", secret.Name)
	_ = d.Set("description", types.FlattenStringPtr(secret.Description))
	_ = d.Set("created_at", types.FlattenTime(secret.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(secret.UpdatedAt))
	_ = d.Set("status", secret.Status.String())
	_ = d.Set("region", string(secret.Region))
	_ = d.Set("project_id", secret.ProjectID)
	_ = d.Set("path", secret.Path)
	_ = d.Set("protected", secret.Protected)
	_ = d.Set("ephemeral_policy", flattenEphemeralPolicy(secret.EphemeralPolicy))
	_ = d.Set("type", secret.Type)
	_ = d.Set("key_id", types.FlattenStringPtr(secret.KeyID))
	_ = d.Set("tags", types.FlattenSliceString(secret.Tags))
	_ = d.Set("srn", secret.Srn)

	// versions is not populated for list secret
	if versions != nil {
		versionsList := make([]map[string]any, 0, len(versions.Versions))
		for _, version := range versions.Versions {
			versionsList = append(versionsList, map[string]any{
				"revision":    strconv.Itoa(int(version.Revision)),
				"secret_id":   version.SecretID,
				"status":      version.Status.String(),
				"created_at":  types.FlattenTime(version.CreatedAt),
				"updated_at":  types.FlattenTime(version.UpdatedAt),
				"description": types.FlattenStringPtr(version.Description),
				"latest":      types.FlattenBoolPtr(&version.Latest),
			})
		}

		_ = d.Set("version_count", int(versions.TotalCount))
		_ = d.Set("versions", versionsList)
	}
}

func setVersionState(d *schema.ResourceData, version *secret.SecretVersion) {
	revisionStr := strconv.Itoa(int(version.Revision))
	_ = d.Set("revision", revisionStr)
	_ = d.Set("secret_id", regional.NewIDString(version.Region, version.SecretID))
	_ = d.Set("description", types.FlattenStringPtr(version.Description))
	_ = d.Set("created_at", types.FlattenTime(version.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(version.UpdatedAt))
	_ = d.Set("status", version.Status.String())
	_ = d.Set("region", string(version.Region))
	_ = d.Set("srn", version.Srn)
}

func secretVersionUpgradeV1SchemaType() cty.Type {
	return cty.Object(map[string]cty.Type{
		"id":              cty.String,
		"secret_id":       cty.String,
		"data":            cty.String,
		"data_wo_version": cty.Number,
		"description":     cty.String,
		"status":          cty.String,
		"revision":        cty.String,
		"created_at":      cty.String,
		"updated_at":      cty.String,
		"srn":             cty.String,
		"region":          cty.String,
	})
}

// VersionV1SchemaUpgradeFunc migrates a V0 state to V1.
//
// Before the `type` attribute existed, the value stored in state for `data`
// was the result of the Base64Encoded helper, which base64-encodes its input
// unless it is already a valid base64 string:
//   - a raw (non-base64) configuration value was stored as base64(config);
//   - a configuration value that was itself a valid base64 string was stored
//     verbatim.
//
// In the first (common) case, decoding the state value recovers the exact
// configuration value, so no ForceNew-based replacement occurs on upgrade. In
// the second case, decoding yields a value different from the configuration,
// so the first plan after the upgrade may recreate the secret version (the
// secret payload itself is preserved). The two cases are indistinguishable
// from the stored state alone and the upgrader cannot access the
// configuration, so we decode by default to keep the common case correct.
//
// If the state value decodes to bytes that are not valid UTF-8, storing the
// decoded value would corrupt the state, so we keep the value as-is and set
// `type` to "base64" to preserve the payload sent to the API.
func VersionV1SchemaUpgradeFunc(_ context.Context, rawState map[string]any, _ any) (map[string]any, error) {
	rawState["type"] = "string"

	raw, ok := rawState["data"].(string)
	if !ok || raw == "" {
		return rawState, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err == nil && utf8.Valid(decoded) {
		rawState["data"] = string(decoded)
	} else if err == nil {
		rawState["type"] = "base64"
	}

	return rawState, nil
}
