package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
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
			Create:  schema.DefaultTimeout(defaultContainerDomainTimeout),
			Read:    schema.DefaultTimeout(defaultContainerDomainTimeout),
			Update:  schema.DefaultTimeout(defaultContainerDomainTimeout),
			Delete:  schema.DefaultTimeout(defaultContainerDomainTimeout),
			Default: schema.DefaultTimeout(defaultContainerDomainTimeout),
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
			"region": regional.Schema(),
		},
		CustomizeDiff: CustomizeDiffLocalityCheck("container_id"),
	}
}

func resourceScalewayContainerDomainCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := containerAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	hostname := d.Get("hostname").(string)
	containerID := locality.ExpandID(d.Get("container_id"))

	_, err = waitForContainer(ctx, api, containerID, region, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	req := &container.CreateDomainRequest{
		Region:      region,
		Hostname:    hostname,
		ContainerID: containerID,
	}

	domain, err := retryCreateContainerDomain(ctx, api, req, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForContainerDomain(ctx, api, domain.ID, region, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, domain.ID))

	return resourceScalewayContainerDomainRead(ctx, d, m)
}

func resourceScalewayContainerDomainRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, domainID, err := containerAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	domain, err := waitForContainerDomain(ctx, api, domainID, region, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		if httperrors.Is404(err) {
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

func resourceScalewayContainerDomainDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, domainID, err := containerAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForContainerDomain(ctx, api, domainID, region, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_, err = api.DeleteDomain(&container.DeleteDomainRequest{
		Region:   region,
		DomainID: domainID,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
