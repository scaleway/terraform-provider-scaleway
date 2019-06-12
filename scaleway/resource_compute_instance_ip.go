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
				Description: "The ip address",
			},
			"reverse": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The reverse dns for this IP",
			},
			"server_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The server associated with this ip",
			},
			"zone":       zoneSchema(),
			"project_id": projectIDSchema(),
		},
	}
}

func resourceScalewayComputeInstanceIPCreate(d *schema.ResourceData, m interface{}) error {
	meta := m.(*Meta)
	instanceApi := instance.NewAPI(meta.scwClient)

	zone, err := getZone(d, meta)
	if err != nil {
		return err
	}

	res, err := instanceApi.CreateIP(&instance.CreateIPRequest{
		Zone:         zone,
		Organization: d.Get("project_id").(string),
	})
	if err != nil {
		return err
	}

	reverse := d.Get("reverse").(string)
	if reverse != "" {
		_, err = instanceApi.UpdateIP(&instance.UpdateIPRequest{
			Zone:    zone,
			IPID:    res.IP.ID,
			Reverse: &reverse,
		})
		if err != nil {
			return err
		}
	}

	d.SetId(newZonedId(zone, res.IP.ID))
	return resourceScalewayComputeInstanceIPRead(d, m)
}

func resourceScalewayComputeInstanceIPRead(d *schema.ResourceData, m interface{}) error {
	meta := m.(*Meta)
	instanceApi := instance.NewAPI(meta.scwClient)

	zone, ID, err := parseZonedID(d.Id())
	if err != nil {
		return err
	}

	res, err := instanceApi.GetIP(&instance.GetIPRequest{
		IPID: ID,
		Zone: zone,
	})

	if err != nil {
		// We check for 403 because instance API return 403 for deleted IP
		if is404Error(err) || is403Error(err) {
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("address", res.IP.Address)
	d.Set("zone", string(zone))
	d.Set("project_id", res.IP.Organization)
	d.Set("reverse", res.IP.Reverse)

	if res.IP.Server != nil {
		d.Set("server_id", res.IP.Server.ID)
	} else {
		d.Set("server_id", nil)
	}

	return nil
}

func resourceScalewayComputeInstanceIPUpdate(d *schema.ResourceData, m interface{}) error {
	meta := m.(*Meta)
	instanceApi := instance.NewAPI(meta.scwClient)

	zone, ID, err := parseZonedID(d.Id())
	if err != nil {
		return err
	}

	if d.HasChange("reverse") {
		l.Debugf("updating IP %q reverse to %s\n", d.Id(), d.Get("reverse"))

		reverse := d.Get("reverse").(string)
		_, err = instanceApi.UpdateIP(&instance.UpdateIPRequest{
			Zone:    zone,
			IPID:    ID,
			Reverse: &reverse,
		})
		if err != nil {
			return err
		}
	}

	return resourceScalewayComputeInstanceIPRead(d, m)
}

func resourceScalewayComputeInstanceIPDelete(d *schema.ResourceData, m interface{}) error {
	meta := m.(*Meta)
	instanceApi := instance.NewAPI(meta.scwClient)

	zone, ID, err := parseZonedID(d.Id())
	if err != nil {
		return err
	}

	err = instanceApi.DeleteIP(&instance.DeleteIPRequest{
		IPID: ID,
		Zone: zone,
	})

	if err != nil && !is404Error(err) {
		return err
	}

	return nil
}
