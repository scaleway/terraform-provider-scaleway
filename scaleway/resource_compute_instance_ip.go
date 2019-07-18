package scaleway

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
)

func resourceScalewayComputeInstanceIP() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayComputeInstanceIPCreate,
		Read:   resourceScalewayComputeInstanceIPRead,
		Update: resourceScalewayComputeInstanceIPUpdate,
		Delete: resourceScalewayComputeInstanceIPDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The IP address",
			},
			"reverse": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The reverse DNS for this IP",
			},
			"server_id": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "The server associated with this IP",
				ValidateFunc:     validationUUIDorUUIDWithLocality(),
				DiffSuppressFunc: suppressLocality,
			},
			"zone":       zoneSchema(),
			"project_id": projectIDSchema(),
		},
	}
}

func resourceScalewayComputeInstanceIPCreate(d *schema.ResourceData, m interface{}) error {
	instanceAPI, zone, err := getInstanceAPIWithZone(d, m)
	if err != nil {
		return err
	}

	res, err := instanceAPI.CreateIP(&instance.CreateIPRequest{
		Zone:         zone,
		Organization: d.Get("project_id").(string),
	})
	if err != nil {
		return err
	}

	reverse := d.Get("reverse").(string)
	if reverse != "" {
		_, err = instanceAPI.UpdateIP(&instance.UpdateIPRequest{
			Zone:    zone,
			IPID:    res.IP.ID,
			Reverse: &instance.NullableStringValue{Value: reverse},
		})
		if err != nil {
			return err
		}
	}

	d.SetId(newZonedId(zone, res.IP.ID))

	serverID := expandID(d.Get("server_id"))
	if serverID != "" {
		_, err = instanceAPI.AttachIP(&instance.AttachIPRequest{
			Zone:     zone,
			IPID:     res.IP.ID,
			ServerID: serverID,
		})
		if err != nil {
			return err
		}

	}
	return resourceScalewayComputeInstanceIPRead(d, m)
}

func resourceScalewayComputeInstanceIPRead(d *schema.ResourceData, m interface{}) error {
	instanceAPI, zone, ID, err := getInstanceAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return err
	}

	res, err := instanceAPI.GetIP(&instance.GetIPRequest{
		IPID: ID,
		Zone: zone,
	})

	if err != nil {
		// We check for 403 because instance API returns 403 for a deleted IP
		if is404Error(err) || is403Error(err) {
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("address", res.IP.Address.String())
	d.Set("zone", string(zone))
	d.Set("project_id", res.IP.Organization)
	d.Set("reverse", res.IP.Reverse)
	
	if res.IP.Server != nil {
		d.Set("server_id", res.IP.Server.ID)
	}

	return nil
}

func resourceScalewayComputeInstanceIPUpdate(d *schema.ResourceData, m interface{}) error {
	instanceAPI, zone, ID, err := getInstanceAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return err
	}

	if d.HasChange("reverse") {
		l.Debugf("updating IP %q reverse to %q\n", d.Id(), d.Get("reverse"))

		updateReverseReq := &instance.UpdateIPRequest{
			Zone: zone,
			IPID: ID,
		}

		reverse := d.Get("reverse").(string)
		if reverse == "" {
			updateReverseReq.Reverse = &instance.NullableStringValue{Null: true}
		} else {
			updateReverseReq.Reverse = &instance.NullableStringValue{Value: reverse}
		}
		_, err = instanceAPI.UpdateIP(updateReverseReq)
		if err != nil {
			return err
		}
	}

	if d.HasChange("server_id") {
		serverID := expandID(d.Get("server_id"))
		if serverID != "" {
			_, err = instanceAPI.AttachIP(&instance.AttachIPRequest{
				Zone:     zone,
				IPID:     ID,
				ServerID: serverID,
			})
		} else {
			_, err = instanceAPI.DetachIP(&instance.DetachIPRequest{
				Zone: zone,
				IPID: ID,
			})
		}
		if err != nil {
			return err
		}
	}

	return resourceScalewayComputeInstanceIPRead(d, m)
}

func resourceScalewayComputeInstanceIPDelete(d *schema.ResourceData, m interface{}) error {
	instanceAPI, zone, ID, err := getInstanceAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return err
	}

	err = instanceAPI.DeleteIP(&instance.DeleteIPRequest{
		IPID: ID,
		Zone: zone,
	})

	if err != nil && !is404Error(err) {
		return err
	}

	return nil
}
