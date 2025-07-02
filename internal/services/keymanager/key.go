package keymanager

import (
	"context"

	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"

	key_manager "github.com/scaleway/scaleway-sdk-go/api/key_manager/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
)

func ResourceKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeyCreate,
		ReadContext:   resourceKeyRead,
		UpdateContext: resourceKeyUpdate,
		DeleteContext: resourceKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Read:    schema.DefaultTimeout(defaultKeyTimeout),
			Update:  schema.DefaultTimeout(defaultKeyTimeout),
			Delete:  schema.DefaultTimeout(defaultKeyTimeout),
			Default: schema.DefaultTimeout(defaultKeyTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the key.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the key.",
			},
			"state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The state of the key (enabled, disabled, pending_key_material)",
			},
			"locked": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "The locked state of the key.",
			},
			"protected": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Returns `true` if key protection is applied to the key",
			},
			"rotation_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of times the key has been rotated",
			},
			"rotated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last rotation of the key",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the key",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last update of the key",
			},
			"tags": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of tags [\"tag1\", \"tag2\", ...] attached to a key",
			},
			"origin": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The origin of the key (scaleway_kms, external)",
			},
			// Computed
			"project_id": account.ProjectIDSchema(),
			"region":     regional.ComputedSchema(),
		},
	}
}

func resourceKeyCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, err := newKeyManagerAPI(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	keyCreateRequest := &key_manager.CreateKeyRequest{
		Region:      region,
		ProjectID:   d.Get("project_id").(string),
		Name:        types.ExpandStringPtr(d.Get("name")),
		Description: types.ExpandStringPtr(d.Get("description")),
		Unprotected: d.Get("protected").(bool),
	}

	rawTag, tagExist := d.GetOk("tags")
	if tagExist {
		keyCreateRequest.Tags = types.ExpandStrings(rawTag)
	}

	rawDescription, descriptionExist := d.GetOk("description")
	if descriptionExist {
		keyCreateRequest.Description = types.ExpandStringPtr(rawDescription)
	}

	keyResponse, err := api.CreateKey(keyCreateRequest, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, keyResponse.ID))

	return resourceKeyRead(ctx, d, meta)
}

func resourceKeyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, id, err := NewKeyManagerAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	key, err := api.GetKey(&key_manager.GetKeyRequest{
		Region: region,
		KeyID:  id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("region", key.Region)
	_ = d.Set("project_id", key.ProjectID)
	_ = d.Set("description", key.Description)
	_ = d.Set("name", key.Name)
	_ = d.Set("state", key.State)
	_ = d.Set("rotation_count", key.RotationCount)
	_ = d.Set("created_at", types.FlattenTime(key.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(key.UpdatedAt))
	_ = d.Set("origin", key.Origin)
	_ = d.Set("rotated_at", types.FlattenTime(key.RotatedAt))
	_ = d.Set("locked", key.Locked)
	_ = d.Set("protected", key.Protected)

	return nil
}

func resourceKeyUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	keyManagerAPI, region, ID, err := NewKeyManagerAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	shouldUpdate := false
	req := &key_manager.UpdateKeyRequest{
		Region: region,
		KeyID:  ID,
	}

	if d.HasChange("name") {
		req.Name = types.ExpandStringPtr(d.Get("name"))
		shouldUpdate = true
	}

	if d.HasChange("tags") {
		req.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
	}

	if shouldUpdate {
		_, err = keyManagerAPI.UpdateKey(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceKeyRead(ctx, d, meta)
}

func resourceKeyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	keyManagerAPI, region, ID, err := NewKeyManagerAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = keyManagerAPI.DeleteKey(&key_manager.DeleteKeyRequest{
		KeyID:  ID,
		Region: region,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
