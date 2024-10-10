package cockpit

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceCockpitSource() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceCockpitSourceCreate,
		ReadContext:   ResourceCockpitSourceRead,
		DeleteContext: ResourceCockpitSourceDelete,
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(DefaultCockpitTimeout),
			Read:    schema.DefaultTimeout(DefaultCockpitTimeout),
			Delete:  schema.DefaultTimeout(DefaultCockpitTimeout),
			Default: schema.DefaultTimeout(DefaultCockpitTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Name of the datasource",
			},
			"type": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Description:      "The type of the datasource",
				ValidateDiagFunc: verify.ValidateEnum[cockpit.DataSourceType](),
			},
			// computed
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The URL of the datasource",
			},
			"origin": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The origin of the datasource",
			},
			"synchronized_with_grafana": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether the data source is synchronized with Grafana",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the cockpit datasource",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last update of the cockpit datasource",
			},
			"push_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The URL endpoint used for pushing data to the cockpit data source.",
			},
			"project_id": account.ProjectIDSchema(),
			"region":     regional.Schema(),
		},
	}
}

func ResourceCockpitSourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := cockpitAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := api.CreateDataSource(&cockpit.RegionalAPICreateDataSourceRequest{
		Region:    region,
		ProjectID: d.Get("project_id").(string),
		Name:      d.Get("name").(string),
		Type:      cockpit.DataSourceType(d.Get("type").(string)),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, res.ID))
	return ResourceCockpitSourceRead(ctx, d, meta)
}

func ResourceCockpitSourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := api.GetDataSource(&cockpit.RegionalAPIGetDataSourceRequest{
		Region:       region,
		DataSourceID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	pushURL, err := createCockpitPushURL(res.Type, res.URL)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("name", res.Name)
	_ = d.Set("type", res.Type.String())
	_ = d.Set("url", res.URL)
	_ = d.Set("origin", res.Origin)
	_ = d.Set("synchronized_with_grafana", res.SynchronizedWithGrafana)
	_ = d.Set("region", res.Region)
	_ = d.Set("created_at", types.FlattenTime(res.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(res.UpdatedAt))
	_ = d.Set("project_id", res.ProjectID)
	_ = d.Set("push_url", pushURL)

	return nil
}

func ResourceCockpitSourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = api.DeleteDataSource(&cockpit.RegionalAPIDeleteDataSourceRequest{
		DataSourceID: id,
		Region:       region,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
