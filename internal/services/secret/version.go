package secret

import (
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	secret "github.com/scaleway/scaleway-sdk-go/api/secret/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

//go:embed descriptions/secret_version_resource.md
var versionResourceDescription string

func ResourceVersion() *schema.Resource {
	return &schema.Resource{
		Description:   versionResourceDescription,
		CreateContext: ResourceVersionCreate,
		ReadContext:   ResourceVersionRead,
		UpdateContext: ResourceVersionUpdate,
		DeleteContext: ResourceVersionDelete,
		Identity:      secretVersionIdentity(),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultSecretTimeout),
		},
		SchemaVersion: 1,
		SchemaFunc:    versionSchema,
		CustomizeDiff: customdiff.All(resourceVersionTypeCustomizeDiff),
		StateUpgraders: []schema.StateUpgrader{
			{
				Version: 0,
				Type:    secretVersionUpgradeV1SchemaType(),
				Upgrade: VersionV1SchemaUpgradeFunc,
			},
		},
	}
}

func resourceVersionTypeCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ any) error {
	_, newType := d.GetChange("type")

	if newType.(string) != "base64" {
		return nil
	}

	for _, attribute := range []string{"data", "data_wo"} {
		_, raw := d.GetChange(attribute)

		value, _ := raw.(string)
		if value == "" {
			continue
		}

		if _, err := base64.StdEncoding.DecodeString(value); err != nil {
			return fmt.Errorf("%s must be a valid base64-encoded string when `type` is set to \"base64\": %w", attribute, err)
		}
	}

	return nil
}

func versionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"secret_id": {
			Type:             schema.TypeString,
			Required:         true,
			Description:      "The secret ID associated with this version",
			ForceNew:         true,
			DiffSuppressFunc: dsf.Locality,
		},
		"data": {
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "The data payload of the secret version, encoded as defined by the `type` attribute. Must not exceed 64KiB in size. Only one of `data` or `data_wo` should be specified.",
			Sensitive:    true,
			ForceNew:     true,
			ExactlyOneOf: []string{"data", "data_wo"},
		},
		"data_wo": {
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "The data payload of your secret version in [write-only](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/guides/using-write-only-arguments) mode, encoded as defined by the `type` attribute. Must not exceed 64KiB in size. Only one of `data` or `data_wo` should be specified. `data_wo` will not be set in the Terraform state. To update the `data_wo`, you must also update the `data_wo_version`.",
			Sensitive:    true,
			WriteOnly:    true,
			ExactlyOneOf: []string{"data", "data_wo"},
		},
		"type": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "string",
			Description:  "The encoding of the `data` or `data_wo` attribute. Can be either `string` (default) or `base64`.",
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice([]string{"base64", "string"}, false),
		},
		"data_wo_version": {
			Type:        schema.TypeInt,
			Optional:    true,
			ForceNew:    true,
			Description: "The version of the [write-only](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/guides/using-write-only-arguments) data. To update the `data_wo`, you must also update the `data_wo_version`.",
		},
		"description": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Description of the secret version",
		},
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Status of the secret version",
		},
		"revision": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The revision of secret version.",
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Date and time of secret version's creation (RFC 3339 format)",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Date and time of secret version's creation (RFC 3339 format)",
		},
		"region": regional.Schema(),
		"srn": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The Scaleway Resource Name (SRN) of the secret version",
		},
	}
}

func secretVersionIdentity() *schema.ResourceIdentity {
	return identity.WrapSchemaMap(map[string]*schema.Schema{
		"region": identity.DefaultRegionAttribute(),
		"secret_id": {
			Type:              schema.TypeString,
			Description:       "The secret ID",
			RequiredForImport: true,
		},
		"revision": {
			Type:              schema.TypeString,
			Description:       "The revision of the secret version",
			RequiredForImport: true,
		},
	})
}

func ResourceVersionCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	secretID := locality.ExpandID(d.Get("secret_id").(string))

	var payloadSecretRaw []byte

	data, hasData := d.GetOk("data")

	var (
		payload          string
		payloadAttribute string
	)

	if hasData {
		payload = data.(string)
		payloadAttribute = "data"
	} else {
		payload = d.GetRawConfig().GetAttr("data_wo").AsString()
		payloadAttribute = "data_wo"
	}

	if d.Get("type").(string) == "base64" {
		decoded, err := base64.StdEncoding.DecodeString(payload)
		if err != nil {
			return diag.FromErr(fmt.Errorf("%s must be a valid base64-encoded string when `type` is set to \"base64\": %w", payloadAttribute, err))
		}

		payloadSecretRaw = decoded
	} else {
		payloadSecretRaw = []byte(payload)
	}

	secretCreateVersionRequest := &secret.CreateSecretVersionRequest{
		Region:      region,
		SecretID:    secretID,
		Data:        payloadSecretRaw,
		Description: types.ExpandStringPtr(d.Get("description")),
	}

	var secretResponse *secret.SecretVersion

	err = transport.RetryOnStatus(ctx, transport.IAMPropagation, func(ctx context.Context) error {
		var err error

		secretResponse, err = api.CreateSecretVersion(secretCreateVersionRequest, scw.WithContext(ctx))

		return err
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = identity.SetMultiPartIdentity(d, map[string]string{
		"region":    region.String(),
		"secret_id": secretResponse.SecretID,
		"revision":  strconv.FormatUint(uint64(secretResponse.Revision), 10),
	}, "region", "secret_id", "revision")
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceVersionRead(ctx, d, m)
}

func ResourceVersionRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, revision, err := NewVersionAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Retry on 404 to handle eventual consistency issues
	secretResponse, err := transport.RetryOnStatusValue(ctx, transport.NotFound, func(ctx context.Context) (*secret.SecretVersion, error) {
		return api.GetSecretVersion(&secret.GetSecretVersionRequest{
			Region:   region,
			SecretID: id,
			Revision: revision,
		}, scw.WithContext(ctx))
	})
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	setVersionState(d, secretResponse)

	if d.Get("type").(string) == "" {
		_ = d.Set("type", "string")
	}

	err = identity.SetMultiPartIdentity(d, map[string]string{
		"region":    region.String(),
		"secret_id": id,
		"revision":  revision,
	}, "region", "secret_id", "revision")
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func ResourceVersionUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, revision, err := NewVersionAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	updateRequest := &secret.UpdateSecretVersionRequest{
		Region:   region,
		SecretID: id,
		Revision: revision,
	}

	hasChanged := false

	if d.HasChange("description") {
		updateRequest.Description = types.ExpandUpdatedStringPtr(d.Get("description"))
		hasChanged = true
	}

	if hasChanged {
		_, err := api.UpdateSecretVersion(updateRequest, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceVersionRead(ctx, d, m)
}

func ResourceVersionDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, revision, err := NewVersionAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = transport.RetryOnStatus(ctx, transport.IAMPropagation, func(ctx context.Context) error {
		return api.DeleteSecretVersion(&secret.DeleteSecretVersionRequest{
			Region:   region,
			SecretID: id,
			Revision: revision,
		}, scw.WithContext(ctx))
	})
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
