package mnq

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceSQSCredentials() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceMNQSQSCredentialsCreate,
		ReadContext:   ResourceMNQSQSCredentialsRead,
		UpdateContext: ResourceMNQSQSCredentialsUpdate,
		DeleteContext: ResourceMNQSQSCredentialsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		SchemaFunc:    sqsCredentialsSchema,
		Identity:      identity.DefaultRegional(),
	}
}

func sqsCredentialsSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			Description: "The credentials name",
		},
		"permissions": {
			Type:        schema.TypeList,
			Description: "The permissions attached to the credentials",
			MaxItems:    1,
			Optional:    true,
			Computed:    true,
			Elem: &schema.Resource{
				SchemaFunc: func() map[string]*schema.Schema {
					return map[string]*schema.Schema{
						"can_publish": {
							Type:        schema.TypeBool,
							Computed:    true,
							Optional:    true,
							Description: "Allow publish messages to the service",
						},
						"can_receive": {
							Type:        schema.TypeBool,
							Computed:    true,
							Optional:    true,
							Description: "Allow receive messages from the service",
						},
						"can_manage": {
							Type:        schema.TypeBool,
							Computed:    true,
							Optional:    true,
							Description: "Allow manage the associated resource",
						},
					}
				},
			},
		},
		"region":     regional.Schema(),
		"project_id": account.ProjectIDSchema(),

		// Computed
		"access_key": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "SQS credentials access key",
			Sensitive:   true,
		},
		"secret_key": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "SQS credentials secret key",
			Sensitive:   true,
		},
	}
}

func ResourceMNQSQSCredentialsCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := newSQSAPI(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	credentials, err := api.CreateSqsCredentials(&mnq.SqsAPICreateSqsCredentialsRequest{
		Region:    region,
		ProjectID: d.Get("project_id").(string),
		Name:      types.ExpandOrGenerateString(d.Get("name").(string), "sqs-credentials"),
		Permissions: &mnq.SqsPermissions{
			CanPublish: types.ExpandBoolPtr(d.Get("permissions.0.can_publish")),
			CanReceive: types.ExpandBoolPtr(d.Get("permissions.0.can_receive")),
			CanManage:  types.ExpandBoolPtr(d.Get("permissions.0.can_manage")),
		},
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	err = identity.SetRegionalIdentity(d, credentials.Region, credentials.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("access_key", credentials.AccessKey)
	_ = d.Set("secret_key", credentials.SecretKey)

	return ResourceMNQSQSCredentialsRead(ctx, d, m)
}

func ResourceMNQSQSCredentialsRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewSQSAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	credentials, err := api.GetSqsCredentials(&mnq.SqsAPIGetSqsCredentialsRequest{
		Region:           region,
		SqsCredentialsID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("name", credentials.Name)
	_ = d.Set("region", credentials.Region)
	_ = d.Set("project_id", credentials.ProjectID)

	if credentials.Permissions != nil {
		_ = d.Set("permissions", []map[string]any{{
			"can_publish": credentials.Permissions.CanPublish,
			"can_receive": credentials.Permissions.CanReceive,
			"can_manage":  credentials.Permissions.CanManage,
		}})
	}

	return nil
}

func ResourceMNQSQSCredentialsUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewSQSAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	req := &mnq.SqsAPIUpdateSqsCredentialsRequest{
		Region:           region,
		SqsCredentialsID: id,
	}

	if d.HasChange("name") {
		req.Name = types.ExpandUpdatedStringPtr(d.Get("name"))
	}

	if d.HasChange("permissions.0") {
		req.Permissions = &mnq.SqsPermissions{}

		if d.HasChange("permissions.0.can_publish") {
			req.Permissions.CanPublish = types.ExpandBoolPtr(d.Get("permissions.0.can_publish"))
		}

		if d.HasChange("permissions.0.can_receive") {
			req.Permissions.CanReceive = types.ExpandBoolPtr(d.Get("permissions.0.can_receive"))
		}

		if d.HasChange("permissions.0.can_manage") {
			req.Permissions.CanManage = types.ExpandBoolPtr(d.Get("permissions.0.can_manage"))
		}
	}

	if _, err := api.UpdateSqsCredentials(req, scw.WithContext(ctx)); err != nil {
		return diag.FromErr(err)
	}

	return ResourceMNQSQSCredentialsRead(ctx, d, m)
}

func ResourceMNQSQSCredentialsDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewSQSAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = api.DeleteSqsCredentials(&mnq.SqsAPIDeleteSqsCredentialsRequest{
		Region:           region,
		SqsCredentialsID: id,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
