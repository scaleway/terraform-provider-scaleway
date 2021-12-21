package scaleway

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	DefaultWaitLBRetryInterval = 30 * time.Second
)

func resourceScalewayLb() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayLbCreate,
		ReadContext:   resourceScalewayLbRead,
		UpdateContext: resourceScalewayLbUpdate,
		DeleteContext: resourceScalewayLbDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultLbLbTimeout),
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{Version: 0, Type: lbUpgradeV1SchemaType(), Upgrade: lbUpgradeV1SchemaUpgradeFunc},
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of the lb",
			},
			"type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: diffSuppressFuncIgnoreCase,
				Description:      "The type of load-balancer you want to create",
			},
			"tags": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Array of tags to associate with the load-balancer",
			},
			"ip_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				Description:      "The load-balance public IP ID",
				DiffSuppressFunc: diffSuppressFuncLocality,
			},
			"ip_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The load-balance public IP address",
			},
			"release_ip": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Release the IPs related to this load-balancer",
			},
			"private_network": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    8,
				Description: "List of private network to connect with your load balancer",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"private_network_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validationUUIDorUUIDWithLocality(),
							Description:  "The Private Network ID",
						},
						"static_config": {
							Description: "Define two IP addresses in the subnet of your private network that will be assigned for the principal and standby node of your load balancer.",
							Type:        schema.TypeList,
							Optional:    true,
							Computed:    true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.IsIPAddress,
							},
						},
						"dhcp_config": {
							Description: "Set to true if you want to let DHCP assign IP addresses",
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
						},
						// Readonly attributes
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The status of private network connection",
						},
						"zone": zoneSchema(),
					},
				},
			},
			"region":          regionComputedSchema(),
			"zone":            zoneSchema(),
			"organization_id": organizationIDSchema(),
			"project_id":      projectIDSchema(),
		},
	}
}

func resourceScalewayLbCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lbAPI, zone, err := lbAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	createReq := &lb.ZonedAPICreateLBRequest{
		Zone:      zone,
		IPID:      expandStringPtr(expandID(d.Get("ip_id"))),
		ProjectID: expandStringPtr(d.Get("project_id")),
		Name:      expandOrGenerateString(d.Get("name"), "lb"),
		Type:      d.Get("type").(string),
	}

	if raw, ok := d.GetOk("tags"); ok {
		for _, tag := range raw.([]interface{}) {
			createReq.Tags = append(createReq.Tags, tag.(string))
		}
	}
	res, err := lbAPI.CreateLB(createReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newZonedIDString(zone, res.ID))
	// wait for lb
	retryInterval := DefaultWaitLBRetryInterval
	_, err = lbAPI.WaitForLb(&lb.ZonedAPIWaitForLBRequest{
		Zone:          zone,
		LBID:          res.ID,
		Timeout:       scw.TimeDurationPtr(defaultInstanceServerWaitTimeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
	// check err waiting process
	if err != nil {
		return diag.FromErr(err)
	}

	//attach private network
	pnConfigs, pnExist := d.GetOk("private_network")
	if pnExist {
		pnConfigs, err := expandPrivateNetworks(pnConfigs, res.ID)
		if err != nil {
			return diag.FromErr(err)
		}

		for _, config := range pnConfigs {
			_, err := lbAPI.AttachPrivateNetwork(config, scw.WithContext(ctx))
			if err != nil && !is404Error(err) {
				return diag.FromErr(err)
			}
		}
		_, err = lbAPI.WaitForLb(&lb.ZonedAPIWaitForLBRequest{
			Zone:          zone,
			LBID:          res.ID,
			Timeout:       scw.TimeDurationPtr(defaultInstanceServerWaitTimeout),
			RetryInterval: &retryInterval,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceScalewayLbRead(ctx, d, meta)
}

func resourceScalewayLbRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lbAPI, zone, ID, err := lbAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	retryInterval := DefaultWaitLBRetryInterval
	res, err := lbAPI.WaitForLb(&lb.ZonedAPIWaitForLBRequest{
		Zone:          zone,
		LBID:          ID,
		Timeout:       scw.TimeDurationPtr(defaultInstanceServerWaitTimeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) || is403Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	// set the region from zone
	region, err := zone.Region()
	if err != nil {
		return diag.FromErr(err)
	}

	var relaseIPValue bool
	releaseIPAddress, releaseIPPExist := d.GetOk("release_ip")
	if releaseIPPExist {
		relaseIPValue = *expandBoolPtr(releaseIPAddress)
	}

	_ = d.Set("release_ip", relaseIPValue)
	_ = d.Set("name", res.Name)
	_ = d.Set("zone", zone.String())
	_ = d.Set("region", region.String())
	_ = d.Set("organization_id", res.OrganizationID)
	_ = d.Set("project_id", res.ProjectID)
	_ = d.Set("tags", res.Tags)
	// For now API return lowercase lb type. This should be fixed in a near future on the API side
	_ = d.Set("type", strings.ToUpper(res.Type))
	_ = d.Set("ip_id", newZonedIDString(zone, res.IP[0].ID))
	_ = d.Set("ip_address", res.IP[0].IPAddress)

	// retrieve attached private networks
	resPN, err := lbAPI.ListLBPrivateNetworks(&lb.ZonedAPIListLBPrivateNetworksRequest{
		Zone: zone,
		LBID: ID,
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			return nil
		}
		return diag.FromErr(err)
	}
	_ = d.Set("private_network", flattenPrivateNetworkConfigs(resPN))

	return nil
}

func resourceScalewayLbUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lbAPI, zone, ID, err := lbAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChanges("name", "tags") {
		req := &lb.ZonedAPIUpdateLBRequest{
			Zone: zone,
			LBID: ID,
			Name: d.Get("name").(string),
			Tags: expandStrings(d.Get("tags")),
		}

		_, err = lbAPI.WaitForLb(&lb.ZonedAPIWaitForLBRequest{
			LBID:          ID,
			Zone:          zone,
			Timeout:       scw.TimeDurationPtr(LbWaitForTimeout),
			RetryInterval: DefaultWaitRetryInterval,
		}, scw.WithContext(ctx))

		if err != nil && !is404Error(err) {
			return diag.FromErr(err)
		}

		_, err = lbAPI.UpdateLB(req, scw.WithContext(ctx))
		if err != nil && !is404Error(err) {
			return diag.FromErr(err)
		}
	}
	////
	// Attach / Detach Private Networks
	////
	if d.HasChangesExcept("private_network") {
		retryInterval := DefaultWaitLBRetryInterval
		// check that pns are in a stable state
		pns, err := lbAPI.WaitForLBPN(&lb.ZonedAPIWaitForLBPNRequest{
			Zone:          zone,
			LBID:          ID,
			Timeout:       scw.TimeDurationPtr(defaultInstanceServerWaitTimeout),
			RetryInterval: &retryInterval},
			scw.WithContext(ctx))
		if err != nil && !is404Error(err) {
			return diag.FromErr(err)
		}
		// select only private networks that has change
		pnToDetach, err := privateNetworksToDetach(pns, d.Get("private_network"))
		if err != nil {
			diag.FromErr(err)
		}
		// detach private networks
		for pnID, detach := range pnToDetach {
			if detach {
				err = lbAPI.DetachPrivateNetwork(&lb.ZonedAPIDetachPrivateNetworkRequest{
					Zone:             zone,
					LBID:             ID,
					PrivateNetworkID: pnID,
				})
				if err != nil && !is404Error(err) {
					return diag.FromErr(err)
				}
			}
		}

		//attach private network
		pnConfigs, pnExist := d.GetOk("private_network")
		if pnExist {
			pnConfigs, err := expandPrivateNetworks(pnConfigs, ID)
			if err != nil {
				return diag.FromErr(err)
			}

			for _, config := range pnConfigs {
				// check private network is already in config
				if detach, exist := pnToDetach[config.PrivateNetworkID]; exist && !detach {
					continue
				}
				// check load balancer state
				_, err = lbAPI.WaitForLb(&lb.ZonedAPIWaitForLBRequest{
					Zone:          zone,
					LBID:          ID,
					Timeout:       scw.TimeDurationPtr(defaultInstanceServerWaitTimeout),
					RetryInterval: &retryInterval,
				}, scw.WithContext(ctx))
				if err != nil && !is404Error(err) {
					return diag.FromErr(err)
				}
				// attach updated private networks
				_, err := lbAPI.AttachPrivateNetwork(config, scw.WithContext(ctx))
				if err != nil && !is404Error(err) {
					return diag.FromErr(err)
				}
			}

			_, err = lbAPI.WaitForLBPN(&lb.ZonedAPIWaitForLBPNRequest{
				Zone:          zone,
				LBID:          ID,
				Timeout:       scw.TimeDurationPtr(defaultInstanceServerWaitTimeout),
				RetryInterval: &retryInterval},
				scw.WithContext(ctx))
			if err != nil && !is404Error(err) {
				return diag.FromErr(err)
			}
		}
	}

	return resourceScalewayLbRead(ctx, d, meta)
}

func resourceScalewayLbDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lbAPI, zone, ID, err := lbAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// check if current lb is on stable state
	retryInterval := DefaultWaitLBRetryInterval
	currentLB, err := lbAPI.WaitForLb(&lb.ZonedAPIWaitForLBRequest{
		LBID:          ID,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(LbWaitForTimeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	if currentLB.PrivateNetworkCount != 0 {
		lbPNs, err := lbAPI.ListLBPrivateNetworks(&lb.ZonedAPIListLBPrivateNetworksRequest{
			Zone: zone,
			LBID: ID,
		}, scw.WithContext(ctx))
		if err != nil && !is404Error(err) {
			return diag.FromErr(err)
		}

		// detach private networks
		for _, pn := range lbPNs.PrivateNetwork {
			err = lbAPI.DetachPrivateNetwork(&lb.ZonedAPIDetachPrivateNetworkRequest{
				Zone:             zone,
				LBID:             ID,
				PrivateNetworkID: pn.PrivateNetworkID,
			})
			if err != nil && !is404Error(err) {
				return diag.FromErr(err)
			}
		}

		_, err = lbAPI.WaitForLb(&lb.ZonedAPIWaitForLBRequest{
			LBID:          ID,
			Zone:          zone,
			Timeout:       scw.TimeDurationPtr(LbWaitForTimeout),
			RetryInterval: &retryInterval,
		}, scw.WithContext(ctx))
		if err != nil && !is404Error(err) {
			return diag.FromErr(err)
		}
	}

	var releaseAddressValue bool
	releaseIPAddress, releaseIPPExist := d.GetOk("release_ip")
	if releaseIPPExist {
		releaseAddressValue = *expandBoolPtr(releaseIPAddress)
	}

	err = lbAPI.DeleteLB(&lb.ZonedAPIDeleteLBRequest{
		Zone:      zone,
		LBID:      ID,
		ReleaseIP: releaseAddressValue,
	}, scw.WithContext(ctx))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	_, err = lbAPI.WaitForLb(&lb.ZonedAPIWaitForLBRequest{
		LBID:          ID,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(LbWaitForTimeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
