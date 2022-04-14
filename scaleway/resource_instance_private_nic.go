package scaleway

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	retryInstancePrivateNICInterval = 30 * time.Second
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

	res, err := instanceAPI.CreatePrivateNIC(
		createPrivateNICRequest,
		scw.WithContext(ctx),
	)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = instanceAPI.WaitForPrivateNIC(&instance.WaitForPrivateNICRequest{
		ServerID:      res.PrivateNic.ServerID,
		PrivateNicID:  res.PrivateNic.ID,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(d.Timeout(schema.TimeoutCreate)),
		RetryInterval: scw.TimeDurationPtr(retryInstancePrivateNICInterval),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(
		newZonedNestedIDString(
			zone,
			res.PrivateNic.ServerID,
			res.PrivateNic.ID,
		),
	)

	return resourceScalewayInstancePrivateNICRead(ctx, d, meta)
}

func resourceScalewayInstancePrivateNICRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceAPI, _, err := instanceAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	zone, innerID, outerID, err := parseZonedNestedID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := instanceAPI.WaitForPrivateNIC(&instance.WaitForPrivateNICRequest{
		ServerID:      outerID,
		PrivateNicID:  innerID,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(d.Timeout(schema.TimeoutRead)),
		RetryInterval: scw.TimeDurationPtr(retryInstancePrivateNICInterval),
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("zone", zone)
	_ = d.Set("server_id", newZonedID(zone, res.ServerID).String())
	_ = d.Set("private_network_id", newZonedID(zone, res.PrivateNetworkID).String())
	_ = d.Set("mac_address", res.MacAddress)

	return nil
}

func resourceScalewayInstancePrivateNICUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceAPI, _, err := instanceAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	zone, innerID, outerID, err := parseZonedNestedID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChanges("private_network_id", "server_id") {
		_, err := instanceAPI.WaitForPrivateNIC(&instance.WaitForPrivateNICRequest{
			ServerID:      outerID,
			PrivateNicID:  innerID,
			Zone:          zone,
			Timeout:       scw.TimeDurationPtr(d.Timeout(schema.TimeoutUpdate)),
			RetryInterval: scw.TimeDurationPtr(retryInstancePrivateNICInterval),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		// delete previous private NIC
		err = instanceAPI.DeletePrivateNIC(&instance.DeletePrivateNICRequest{
			ServerID:     outerID,
			PrivateNicID: innerID,
			Zone:         zone,
		}, scw.WithContext(ctx))

		if err != nil && !is404Error(err) {
			return diag.FromErr(err)
		}

		_, err = instanceAPI.WaitForPrivateNIC(&instance.WaitForPrivateNICRequest{
			ServerID:      outerID,
			PrivateNicID:  innerID,
			Zone:          zone,
			Timeout:       scw.TimeDurationPtr(d.Timeout(schema.TimeoutDelete)),
			RetryInterval: scw.TimeDurationPtr(retryInstancePrivateNICInterval),
		}, scw.WithContext(ctx))
		if err != nil && !is404Error(err) {
			return diag.FromErr(err)
		}

		// create the new one
		createPrivateNICRequest := &instance.CreatePrivateNICRequest{
			Zone:             zone,
			ServerID:         expandZonedID(d.Get("server_id").(string)).ID,
			PrivateNetworkID: expandZonedID(d.Get("private_network_id").(string)).ID,
		}

		res, err := instanceAPI.CreatePrivateNIC(
			createPrivateNICRequest,
			scw.WithContext(ctx),
		)
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = instanceAPI.WaitForPrivateNIC(&instance.WaitForPrivateNICRequest{
			ServerID:      res.PrivateNic.ServerID,
			PrivateNicID:  res.PrivateNic.ID,
			Zone:          zone,
			Timeout:       scw.TimeDurationPtr(d.Timeout(schema.TimeoutUpdate)),
			RetryInterval: scw.TimeDurationPtr(retryInstancePrivateNICInterval),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(
			newZonedNestedIDString(
				zone,
				res.PrivateNic.ServerID,
				res.PrivateNic.ID,
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
	zone, innerID, outerID, err := parseZonedNestedID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = instanceAPI.WaitForPrivateNIC(&instance.WaitForPrivateNICRequest{
		ServerID:      outerID,
		PrivateNicID:  innerID,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(d.Timeout(schema.TimeoutRead)),
		RetryInterval: scw.TimeDurationPtr(retryInstancePrivateNICInterval),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	err = instanceAPI.DeletePrivateNIC(&instance.DeletePrivateNICRequest{
		ServerID:     outerID,
		PrivateNicID: innerID,
		Zone:         zone,
	}, scw.WithContext(ctx))

	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	_, err = instanceAPI.WaitForPrivateNIC(&instance.WaitForPrivateNICRequest{
		ServerID:      outerID,
		PrivateNicID:  innerID,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(d.Timeout(schema.TimeoutDelete)),
		RetryInterval: scw.TimeDurationPtr(retryInstancePrivateNICInterval),
	}, scw.WithContext(ctx))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
