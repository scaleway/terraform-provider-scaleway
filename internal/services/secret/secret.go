package secret

import (
	"context"
	"path/filepath"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	secret "github.com/scaleway/scaleway-sdk-go/api/secret/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceSecret() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceSecretCreate,
		ReadContext:   ResourceSecretRead,
		UpdateContext: ResourceSecretUpdate,
		DeleteContext: ResourceSecretDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultSecretTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The secret name",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the secret",
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "List of tags [\"tag1\", \"tag2\", ...] associated to secret",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the secret",
			},
			"version_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of versions for this Secret",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date and time of secret's creation (RFC 3339 format)",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date and time of secret's creation (RFC 3339 format)",
			},
			"path": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Location of the secret in the directory structure.",
				Default:     "/",
				DiffSuppressFunc: func(_, oldValue, newValue string, _ *schema.ResourceData) bool {
					return filepath.Clean(oldValue) == filepath.Clean(newValue)
				},
			},
			"type": {
				ForceNew:         true,
				Type:             schema.TypeString,
				Optional:         true,
				Default:          secret.SecretTypeOpaque,
				ValidateDiagFunc: verify.ValidateEnum[secret.SecretType](),
			},
			"protected": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "True if secret protection is enabled on a given secret. A protected secret cannot be deleted.",
			},
			"ephemeral_policy": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ttl": {
							Optional:         true,
							Type:             schema.TypeString,
							DiffSuppressFunc: dsf.Duration,
							ValidateDiagFunc: verify.IsDuration(),
							Description:      "Time frame, from one second and up to one year, during which the secret's versions are valid. Has to be specified in Go Duration format",
						},
						"expires_once_accessed": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "True if the secret version expires after a single user access.",
						},
						"action": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: verify.ValidateEnum[secret.EphemeralPolicyAction](),
							Description:      "Action to perform when the version of a secret expires.",
						},
					},
				},
			},
			"versions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"revision": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The revision of secret version",
						},
						"secret_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The secret ID associated with this version",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status of the secret version",
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
						"description": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Description of the secret version",
						},
						"latest": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Returns true if the version is the latest.",
						},
					},
				},
			},
			"region":     regional.Schema(),
			"project_id": account.ProjectIDSchema(),
		},
	}
}

func ResourceSecretCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	secretCreateRequest := &secret.CreateSecretRequest{
		Region:    region,
		ProjectID: d.Get("project_id").(string),
		Name:      d.Get("name").(string),
		Protected: d.Get("protected").(bool),
		Type:      secret.SecretType(d.Get("type").(string)),
	}

	rawTag, tagExist := d.GetOk("tags")
	if tagExist {
		secretCreateRequest.Tags = types.ExpandStrings(rawTag)
	}

	rawDescription, descriptionExist := d.GetOk("description")
	if descriptionExist {
		secretCreateRequest.Description = types.ExpandStringPtr(rawDescription)
	}

	rawPath, pathExist := d.GetOk("path")
	if pathExist {
		secretCreateRequest.Path = types.ExpandStringPtr(rawPath)
	}

	rawEphemeralPolicy, policyExists := d.GetOk("ephemeral_policy")
	if policyExists {
		secretCreateRequest.EphemeralPolicy, err = expandEphemeralPolicy(rawEphemeralPolicy)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	secretResponse, err := api.CreateSecret(secretCreateRequest, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, secretResponse.ID))

	return ResourceSecretRead(ctx, d, m)
}

func ResourceSecretRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	secretResponse, err := api.GetSecret(&secret.GetSecretRequest{
		Region:   region,
		SecretID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	if len(secretResponse.Tags) > 0 {
		_ = d.Set("tags", types.FlattenSliceString(secretResponse.Tags))
	}

	versions, err := api.ListSecretVersions(&secret.ListSecretVersionsRequest{
		Region:   region,
		SecretID: id,
	}, scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("name", secretResponse.Name)
	_ = d.Set("description", types.FlattenStringPtr(secretResponse.Description))
	_ = d.Set("created_at", types.FlattenTime(secretResponse.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(secretResponse.UpdatedAt))
	_ = d.Set("status", secretResponse.Status.String())
	_ = d.Set("version_count", int(versions.TotalCount))
	_ = d.Set("region", string(region))
	_ = d.Set("project_id", secretResponse.ProjectID)
	_ = d.Set("path", secretResponse.Path)
	_ = d.Set("protected", secretResponse.Protected)
	_ = d.Set("ephemeral_policy", flattenEphemeralPolicy(secretResponse.EphemeralPolicy))
	_ = d.Set("type", secretResponse.Type)

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

	_ = d.Set("versions", versionsList)

	return nil
}

func ResourceSecretUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	updateRequest := &secret.UpdateSecretRequest{
		Region:   region,
		SecretID: id,
	}

	hasChanged := false

	if d.HasChange("description") {
		updateRequest.Description = types.ExpandUpdatedStringPtr(d.Get("description"))
		hasChanged = true
	}

	if d.HasChange("name") {
		updateRequest.Name = types.ExpandUpdatedStringPtr(d.Get("name"))
		hasChanged = true
	}

	if d.HasChange("tags") {
		updateRequest.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
		hasChanged = true
	}

	if d.HasChange("path") {
		updateRequest.Path = types.ExpandUpdatedStringPtr(d.Get("path"))
		hasChanged = true
	}

	if d.HasChange("ephemeral_policy") {
		updateRequest.EphemeralPolicy, err = expandEphemeralPolicy(d.Get("ephemeral_policy"))
		if err != nil {
			return diag.FromErr(err)
		}

		hasChanged = true
	}

	if hasChanged {
		_, err := api.UpdateSecret(updateRequest, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("protected") {
		err = updateSecretProtection(api, region, id, d.Get("protected").(bool))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceSecretRead(ctx, d, m)
}

func ResourceSecretDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = api.DeleteSecret(&secret.DeleteSecretRequest{
		Region:   region,
		SecretID: id,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
