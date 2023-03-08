package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayContainerDomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayContainerDomainCreate,
		ReadContext:   resourceScalewayContainerDomainRead,
		DeleteContext: resourceScalewayContainerDomainDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultContainerTimeout),
			Read:    schema.DefaultTimeout(defaultContainerTimeout),
			Update:  schema.DefaultTimeout(defaultContainerTimeout),
			Delete:  schema.DefaultTimeout(defaultContainerTimeout),
			Default: schema.DefaultTimeout(defaultContainerTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"hostname": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Domain's hostname",
			},
			"container_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				Description:      "Container the domain will be bound to",
				ValidateFunc:     validationUUIDorUUIDWithLocality(),
				DiffSuppressFunc: diffSuppressFuncLocality,
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL used to query the container",
			},
			"region": regionSchema(),
		},
		CustomizeDiff: customizeDiffLocalityCheck("container_id"),
	}
}

func resourceScalewayContainerDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := containerAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	domain, err := api.CreateDomain(&container.CreateDomainRequest{
		Region:      region,
		Hostname:    d.Get("hostname").(string),
		ContainerID: expandID(d.Get("container_id")),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForContainerDomain(ctx, api, domain.ID, region, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newRegionalIDString(region, domain.ID))

	return resourceScalewayContainerDomainRead(ctx, d, meta)
}

func resourceScalewayContainerDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, domainID, err := containerAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	domain, err := waitForContainerDomain(ctx, api, domainID, region, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("hostname", domain.Hostname)
	_ = d.Set("container_id", domain.ContainerID)
	_ = d.Set("url", domain.URL)
	_ = d.Set("region", region)

	return nil
}

func resourceScalewayContainerDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, domainID, err := containerAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForContainerDomain(ctx, api, domainID, region, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_, err = api.DeleteDomain(&container.DeleteDomainRequest{
		Region:   region,
		DomainID: domainID,
	}, scw.WithContext(ctx))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
