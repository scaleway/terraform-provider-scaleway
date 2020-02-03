package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
)

func resourceScalewayInstanceIP() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayInstanceIPCreate,
		Read:   resourceScalewayInstanceIPRead,
		Delete: resourceScalewayInstanceIPDelete,

		// Because of removed attribute server_id we must add an update func that does nothing. This could be removed on
		// next major release.
		Update: resourceScalewayInstanceIPRead,

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
				Computed:    true,
				Description: "The reverse DNS for this IP",
			},
			"server_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Removed:     "server_id has been removed in favor of scaleway_instance_server.ip_id",
				Description: "The server associated with this IP",
			},
			"zone":            zoneSchema(),
			"organization_id": organizationIDSchema(),
		},
	}
}

func resourceScalewayInstanceIPCreate(d *schema.ResourceData, m interface{}) error {
	instanceAPI, zone, err := instanceAPIWithZone(d, m)
	if err != nil {
		return err
	}

	res, err := instanceAPI.CreateIP(&instance.CreateIPRequest{
		Zone:         zone,
		Organization: d.Get("organization_id").(string),
	})
	if err != nil {
		return err
	}

	d.SetId(newZonedId(zone, res.IP.ID))
	return resourceScalewayInstanceIPRead(d, m)
}

func resourceScalewayInstanceIPRead(d *schema.ResourceData, m interface{}) error {
	instanceAPI, zone, ID, err := instanceAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return err
	}

	res, err := instanceAPI.GetIP(&instance.GetIPRequest{
		IP:   ID,
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

	_ = d.Set("address", res.IP.Address.String())
	_ = d.Set("zone", string(zone))
	_ = d.Set("organization_id", res.IP.Organization)
	_ = d.Set("reverse", res.IP.Reverse)

	return nil
}

func resourceScalewayInstanceIPDelete(d *schema.ResourceData, m interface{}) error {
	instanceAPI, zone, ID, err := instanceAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return err
	}

	err = instanceAPI.DeleteIP(&instance.DeleteIPRequest{
		IP:   ID,
		Zone: zone,
	})

	if err != nil && !is404Error(err) {
		return err
	}

	return nil
}
