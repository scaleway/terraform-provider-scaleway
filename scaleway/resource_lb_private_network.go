package scaleway

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayLbPrivateNetwork() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayLbPrivateNetworkCreate,
		ReadContext:   resourceScalewayLbPrivateNetworkRead,
		DeleteContext: resourceScalewayLbPrivateNetworkDelete,
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
			"lb_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validationUUIDorUUIDWithLocality(),
				Description:  "The load-balancer ID",
			},
			"private_network_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validationUUIDorUUIDWithLocality(),
				Description:  "The Private Network ID",
			},
			"static_config": {
				ConflictsWith: []string{"dhcp_config"},
				Description:   "Define two local IP addresses of your choice for each load balancer instance",
				Type:          schema.TypeList,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsIPAddress,
				},
			},
			"dhcp_config": {
				ConflictsWith: []string{"static_config"},
				Description:   "Set to true if you want to let DHCP assign IP addresses",
				Default:       false,
				Type:          schema.TypeBool,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
			},
			// Readonly attributes
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of private network connection",
			},
		},
	}
}

func resourceScalewayLbPrivateNetworkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lbAPI, _, err := lbAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	zone, lbID, err := parseZonedID(d.Get("lb_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	zonePN, pnID, err := parseZonedID(d.Get("private_network_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	if zonePN != zone {
		return diag.Errorf("LB and Private Network must be in the same zone (got %s and %s)", zone, zonePN)
	}

	retryInterval := DefaultWaitLBRetryInterval
	_, err = lbAPI.WaitForLb(&lb.ZonedAPIWaitForLBRequest{
		Zone:          zone,
		LBID:          lbID,
		Timeout:       scw.TimeDurationPtr(defaultInstanceServerWaitTimeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	createReq := &lb.ZonedAPIAttachPrivateNetworkRequest{
		Zone:             zone,
		LBID:             lbID,
		PrivateNetworkID: pnID,
	}

	dhcpConfig, dhcpConfigExist := d.GetOk("dhcp_config")
	if dhcpConfigExist {
		createReq.DHCPConfig = expandLbPrivateNetworkDHCPConfig(dhcpConfig)
	}

	staticConfig, staticConfigExist := d.GetOk("static_config")
	if staticConfigExist {
		createReq.StaticConfig = expandLbPrivateNetworkStaticConfig(staticConfig)
	}

	res, err := lbAPI.AttachPrivateNetwork(createReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = lbAPI.WaitForLb(&lb.ZonedAPIWaitForLBRequest{
		Zone:          zone,
		LBID:          lbID,
		Timeout:       scw.TimeDurationPtr(defaultInstanceServerWaitTimeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	d.SetId(newZonedIDString(zone, res.PrivateNetworkID))
	_ = d.Set("status", res.Status)

	return resourceScalewayLbPrivateNetworkRead(ctx, d, meta)
}

func resourceScalewayLbPrivateNetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lbAPI, zone, pnID, err := lbAPIWithZoneAndID(meta, d.Get("private_network_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	_, lbID, err := parseZonedID(d.Get("lb_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := lbAPI.ListLBPrivateNetworks(&lb.ZonedAPIListLBPrivateNetworksRequest{
		Zone: zone,
		LBID: lbID,
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			_ = d.Set("lb_id", "")
			return nil
		}
		return diag.FromErr(err)
	}

	pn := findPn(res.PrivateNetwork, pnID)
	if pn == nil {
		d.SetId("")
		return nil
	}

	if pn.DHCPConfig != nil {
		_ = d.Set("dhcp_config", true)
	} else {
		staticConfig := flattenLbPrivateNetworkStaticConfig(pn.StaticConfig).([]string)
		_ = d.Set("static_config", staticConfig)
	}

	_ = d.Set("lb_id", newZonedIDString(zone, pn.LB.ID))
	_ = d.Set("private_network_id", newZonedIDString(zone, pn.PrivateNetworkID))
	_ = d.Set("status", pn.Status)

	return nil
}

func findPn(privateNetworks []*lb.PrivateNetwork, id string) *lb.PrivateNetwork {
	for _, pn := range privateNetworks {
		if pn.PrivateNetworkID == id {
			return pn
		}
	}
	return nil
}

func resourceScalewayLbPrivateNetworkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lbAPI, zone, pnID, err := lbAPIWithZoneAndID(meta, d.Get("private_network_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	_, lbID, err := parseZonedID(d.Get("lb_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	err = lbAPI.DetachPrivateNetwork(&lb.ZonedAPIDetachPrivateNetworkRequest{
		Zone:             zone,
		LBID:             lbID,
		PrivateNetworkID: pnID,
	})
	if err != nil {
		if is404Error(err) || is403Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	retryInterval := DefaultWaitLBRetryInterval
	_, err = lbAPI.WaitForLb(&lb.ZonedAPIWaitForLBRequest{
		Zone:          zone,
		LBID:          lbID,
		Timeout:       scw.TimeDurationPtr(defaultInstanceServerWaitTimeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
