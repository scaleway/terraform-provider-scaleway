package scaleway

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	api "github.com/nicolai86/scaleway-sdk"
)

func resourceScalewayIP() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayIPCreate,
		Read:   resourceScalewayIPRead,
		Update: resourceScalewayIPUpdate,
		Delete: resourceScalewayIPDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"server": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The server associated with the ip",
			},
			"ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ipv4 address of the ip",
			},
			"reverse": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ipv4 reverse dns",
			},
		},
	}
}

func resourceScalewayIPCreate(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway

	ip, err := scaleway.CreateIP()
	if err != nil {
		return err
	}

	d.SetId(ip.ID)
	return resourceScalewayIPUpdate(d, m)
}

func resourceScalewayIPRead(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway
	log.Printf("[DEBUG] Reading IP\n")

	ip, err := scaleway.GetIP(d.Id())
	if err != nil {
		log.Printf("[DEBUG] Error reading ip: %q\n", err)
		if serr, ok := err.(api.APIError); ok {
			if serr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.Set("ip", ip.Address)
	if ip.Server != nil {
		d.Set("server", ip.Server.Identifier)
	}
	if ip.Reverse != nil {
		d.Set("reverse", *ip.Reverse)
	}
	return nil
}

func resourceScalewayIPUpdate(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway

	if d.HasChange("reverse") {
		log.Printf("[DEBUG] Updating IP %q reverse to %q\n", d.Id(), d.Get("reverse").(string))
		ip, err := scaleway.UpdateIP(api.UpdateIPRequest{
			ID:      d.Id(),
			Reverse: d.Get("reverse").(string),
		})
		if err != nil {
			return err
		}
		if ip.Reverse != nil {
			d.Set("reverse", *ip.Reverse)
		} else {
			d.Set("reverse", "")
		}
	}

	if d.HasChange("server") {
		if d.Get("server").(string) != "" {
			log.Printf("[DEBUG] Attaching IP %q to server %q\n", d.Id(), d.Get("server").(string))
			if err := scaleway.AttachIP(d.Id(), d.Get("server").(string)); err != nil {
				return err
			}
		} else {
			log.Printf("[DEBUG] Detaching IP %q\n", d.Id())
			return scaleway.DetachIP(d.Id())
		}
	}

	return resourceScalewayIPRead(d, m)
}

func resourceScalewayIPDelete(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway

	err := scaleway.DeleteIP(d.Id())
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}
