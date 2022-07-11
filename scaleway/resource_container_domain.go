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
			Default: schema.DefaultTimeout(defaultContainerTimeout),
			Read:    schema.DefaultTimeout(defaultContainerTimeout),
			Update:  schema.DefaultTimeout(defaultContainerTimeout),
			Delete:  schema.DefaultTimeout(defaultContainerTimeout),
			Create:  schema.DefaultTimeout(defaultContainerTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"container_id": {
				Type:        schema.TypeString,
				Description: "The ID of the container",
				Required:    true,
				ForceNew:    true,
			},
			"hostname": {
				Type:        schema.TypeString,
				Description: "The hostname that should be redirected to the container",
				Required:    true,
				ForceNew:    true,
			},
			"url": {
				Type:        schema.TypeString,
				Description: "URL to use to trigger the container",
				Computed:    true,
			},
			"region": regionSchema(),
		},
	}
}

func resourceScalewayContainerDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := containerAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	containerID := expandRegionalID(d.Get("container_id").(string)).ID
	_, err = waitForContainer(ctx, api, containerID, region, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	req := &container.CreateDomainRequest{
		Region:      region,
		ContainerID: containerID,
		Hostname:    d.Get("hostname").(string),
	}

	domain, err := api.CreateDomain(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newRegionalIDString(region, domain.ID))

	_, err = waitForContainerDomain(ctx, api, region, domain.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayContainerDomainRead(ctx, d, meta)
}

func resourceScalewayContainerDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := containerAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	domain, err := waitForContainerDomain(ctx, api, region, id, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("hostname", domain.Hostname)
	_ = d.Set("container_id", newRegionalIDString(region, domain.ContainerID))
	_ = d.Set("url", domain.URL)
	_ = d.Set("region", region)

	return nil
}

func resourceScalewayContainerDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := containerAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForContainerDomain(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return nil
	}

	_, err = api.DeleteDomain(&container.DeleteDomainRequest{
		DomainID: id,
		Region:   region,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForContainerDomain(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
