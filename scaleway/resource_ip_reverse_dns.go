package scaleway

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	api "github.com/nicolai86/scaleway-sdk"
)

func resourceScalewayIPReverseDNS() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: `This resource is deprecated and will be removed in the next major version.
 Please use scaleway_compute_instance_ip instead.`,

		Create: resourceScalewayIPReverseDNSCreate,
		Read:   resourceScalewayIPReverseDNSRead,
		Update: resourceScalewayIPReverseDNSUpdate,
		Delete: resourceScalewayIPReverseDNSDelete,

		Schema: map[string]*schema.Schema{
			"ip": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ipv4 address of the ip, or IP ID",
			},
			"reverse": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ipv4 reverse dns",
			},
		},
	}
}

func resourceScalewayIPReverseDNSCreate(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient

	ips, err := scaleway.GetIPS()
	if err != nil {
		return err
	}

	pk := d.Get("ip").(string)
	for _, ip := range ips {
		if ip.ID == pk {
			d.SetId(fmt.Sprintf("ip-reverse-dns/%s", ip.ID))
			return resourceScalewayIPReverseDNSUpdate(d, m)
		}
	}

	return fmt.Errorf("Unable to find IP with Address/ID %q", pk)
}

func resourceScalewayIPReverseDNSRead(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient

	ip, err := scaleway.GetIP(d.Get("ip").(string))
	if err != nil {
		if serr, ok := err.(api.APIError); ok {
			if serr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}
	if ip.Reverse != nil {
		d.Set("reverse", *ip.Reverse)
	}
	return nil
}

func resourceScalewayIPReverseDNSUpdate(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient

	ip, err := scaleway.UpdateIP(api.UpdateIPRequest{
		ID:      d.Get("ip").(string),
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

	return resourceScalewayIPReverseDNSRead(d, m)
}

func resourceScalewayIPReverseDNSDelete(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient

	_, err := scaleway.UpdateIP(api.UpdateIPRequest{
		ID:      d.Get("ip").(string),
		Reverse: "",
	})
	if err != nil {
		if serr, ok := err.(api.APIError); ok {
			if serr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}
	d.SetId("")
	return nil
}
