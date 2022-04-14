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
			Default: schema.DefaultTimeout(defaultInstancePrivateNICWaitTimeout),
		},
		Schema: map[string]*schema.Schema{
			"server_id": {
				Type:        schema.TypeString,
				Description: "The server ID",
				Required:    true,
			},
			"private_network_id": {
				Type:        schema.TypeString,
				Description: "The private network ID",
				Required:    true,
			},
			"mac_address": {
				Type:        schema.TypeString,
				Description: "MAC address of the NIC",
				Computed:    true,
			},
			"zone": zoneSchema(),
		},
	}
}

func resourceScalewayInstancePrivateNICCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceAPI, zone, err := instanceAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	createPrivateNICRequest := &instance.CreatePrivateNICRequest{
		Zone:             zone,
		ServerID:         expandZonedID(d.Get("server_id").(string)).ID,
		PrivateNetworkID: expandZonedID(d.Get("private_network_id").(string)).ID,
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

	_ = d.Set("zone", zone)
	_ = d.Set("server_id", newZonedID(zone, privateNIC.ServerID).String())
	_ = d.Set("private_network_id", newZonedID(zone, privateNIC.PrivateNetworkID).String())
	_ = d.Set("mac_address", privateNIC.MacAddress)

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

	if d.HasChanges("private_network_id", "server_id") {
		_, err = waitForPrivateNIC(ctx, instanceAPI, zone, serverID, privateNICID, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}

		// delete previous private NIC
		err = instanceAPI.DeletePrivateNIC(&instance.DeletePrivateNICRequest{
			ServerID:     serverID,
			PrivateNicID: privateNICID,
			Zone:         zone,
		}, scw.WithContext(ctx))

		if err != nil && !is404Error(err) {
			return diag.FromErr(err)
		}

		_, err = waitForPrivateNIC(ctx, instanceAPI, zone, serverID, privateNICID, d.Timeout(schema.TimeoutUpdate))
		if err != nil && !is404Error(err) {
			return diag.FromErr(err)
		}

		// create the new one
		createPrivateNICRequest := &instance.CreatePrivateNICRequest{
			Zone:             zone,
			ServerID:         expandZonedID(d.Get("server_id").(string)).ID,
			PrivateNetworkID: expandZonedID(d.Get("private_network_id").(string)).ID,
		}

		privateNIC, err := instanceAPI.CreatePrivateNIC(
			createPrivateNICRequest,
			scw.WithContext(ctx),
		)
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = waitForPrivateNIC(ctx, instanceAPI, zone, serverID, privateNICID, d.Timeout(schema.TimeoutUpdate))
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
		return diag.FromErr(err)
	}

	err = instanceAPI.DeletePrivateNIC(&instance.DeletePrivateNICRequest{
		ServerID:     serverID,
		PrivateNicID: privateNICID,
		Zone:         zone,
	}, scw.WithContext(ctx))

	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	_, err = waitForPrivateNIC(ctx, instanceAPI, zone, serverID, privateNICID, d.Timeout(schema.TimeoutDelete))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
