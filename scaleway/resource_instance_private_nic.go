package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayInstancePrivateNIC() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayInstancePrivateNICCreate,
		ReadContext:   resourceScalewayInstancePrivateNICRead,
		UpdateContext: resourceScalewayInstancePrivateNICUpdate,
		DeleteContext: resourceScalewayInstancePrivateNICDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultInstancePrivateNICWaitTimeout),
			Read:    schema.DefaultTimeout(defaultInstancePrivateNICWaitTimeout),
			Update:  schema.DefaultTimeout(defaultInstancePrivateNICWaitTimeout),
			Delete:  schema.DefaultTimeout(defaultInstancePrivateNICWaitTimeout),
			Default: schema.DefaultTimeout(defaultInstancePrivateNICWaitTimeout),
		},
		Schema: map[string]*schema.Schema{
			"server_id": {
				Type:        schema.TypeString,
				Description: "The server ID",
				Required:    true,
				ForceNew:    true,
			},
			"private_network_id": {
				Type:             schema.TypeString,
				Description:      "The private network ID",
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: diffSuppressFuncLocality,
			},
			"mac_address": {
				Type:        schema.TypeString,
				Description: "MAC address of the NIC",
				Computed:    true,
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "The tags associated with the private-nic",
			},
			"ip_ids": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "IPAM ip list, should be for internal use only",
				ForceNew:    true,
			},
			"zone": zoneSchema(),
		},
		CustomizeDiff: customizeDiffLocalityCheck("server_id", "private_network_id"),
	}
}

func resourceScalewayInstancePrivateNICCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceAPI, zone, err := instanceAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForInstanceServer(ctx, instanceAPI, zone, expandID(d.Get("server_id")), d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	createPrivateNICRequest := &instance.CreatePrivateNICRequest{
		Zone:             zone,
		ServerID:         expandZonedID(d.Get("server_id").(string)).ID,
		PrivateNetworkID: expandRegionalID(d.Get("private_network_id").(string)).ID,
		Tags:             expandStrings(d.Get("tags")),
		IPIDs:            expandStrings(d.Get("ip_ids")),
	}

	privateNIC, err := instanceAPI.CreatePrivateNIC(
		createPrivateNICRequest,
		scw.WithContext(ctx),
	)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForPrivateNIC(ctx, instanceAPI, zone, privateNIC.PrivateNic.ServerID, privateNIC.PrivateNic.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(
		newZonedNestedIDString(
			zone,
			privateNIC.PrivateNic.ServerID,
			privateNIC.PrivateNic.ID,
		),
	)

	return resourceScalewayInstancePrivateNICRead(ctx, d, meta)
}

func resourceScalewayInstancePrivateNICRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceAPI, _, err := instanceAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	zone, privateNICID, serverID, err := parseZonedNestedID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	privateNIC, err := waitForPrivateNIC(ctx, instanceAPI, zone, serverID, privateNICID, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	fetchRegion, err := zone.Region()
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("zone", zone)
	_ = d.Set("server_id", newZonedID(zone, privateNIC.ServerID).String())
	_ = d.Set("private_network_id", newRegionalIDString(fetchRegion, privateNIC.PrivateNetworkID))
	_ = d.Set("mac_address", privateNIC.MacAddress)

	if len(privateNIC.Tags) > 0 {
		_ = d.Set("tags", privateNIC.Tags)
	}

	return nil
}

func resourceScalewayInstancePrivateNICUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceAPI, _, err := instanceAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	zone, privateNICID, serverID, err := parseZonedNestedID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("tags") {
		_, err := instanceAPI.UpdatePrivateNIC(
			&instance.UpdatePrivateNICRequest{
				Zone:         zone,
				ServerID:     serverID,
				PrivateNicID: privateNICID,
				Tags:         expandUpdatedStringsPtr(d.Get("tags")),
			},
			scw.WithContext(ctx),
		)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceScalewayInstancePrivateNICRead(ctx, d, meta)
}

func resourceScalewayInstancePrivateNICDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceAPI, _, err := instanceAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	zone, privateNICID, serverID, err := parseZonedNestedID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForPrivateNIC(ctx, instanceAPI, zone, serverID, privateNICID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		if is404Error(err) {
			return nil
		}
		return diag.FromErr(err)
	}

	err = instanceAPI.DeletePrivateNIC(&instance.DeletePrivateNICRequest{
		ServerID:     serverID,
		PrivateNicID: privateNICID,
		Zone:         zone,
	}, scw.WithContext(ctx))

	if err != nil {
		if is404Error(err) {
			return nil
		}
		return diag.FromErr(err)
	}

	_, err = waitForPrivateNIC(ctx, instanceAPI, zone, serverID, privateNICID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		if is404Error(err) {
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}
