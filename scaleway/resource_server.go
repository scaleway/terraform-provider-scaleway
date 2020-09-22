package scaleway

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	api "github.com/nicolai86/scaleway-sdk"
)

var commercialServerTypes []string

func resourceScalewayServer() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: `This resource is deprecated and will be removed in the next major version.
 Please use scaleway_instance_server instead.`,

		Create: resourceScalewayServerCreate,
		Read:   resourceScalewayServerRead,
		Update: resourceScalewayServerUpdate,
		Delete: resourceScalewayServerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the server",
			},
			"image": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The base image of the server",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "The instance type of the server",
				ValidateFunc: validateServerType,
			},
			"boot_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "The boot_type of the server",
				ValidateFunc: validation.StringInSlice([]string{
					"bootscript",
					"local",
				}, false),
			},
			"bootscript": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The boot configuration of the server",
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "The tags associated with the server",
			},
			"security_group": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The security group the server is attached to",
			},
			"volume": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"size_in_gb": {
							Type:     schema.TypeInt,
							Required: true,
							ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
								value := v.(int)
								if value > 300 {
									errors = append(errors, fmt.Errorf("%q needs to be less than 300", k))
								}
								return
							},
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateVolumeType,
						},
						"volume_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Description: "Volumes attached to the server on creation",
			},
			"enable_ipv6": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "determines if IPv6 is enabled for the server",
			},
			"dynamic_ip_required": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "determines if a public IP address should be allocated for the server",
			},
			"private_ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "the Scaleway internal IP address of the server",
			},
			"public_ip": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "the public IPv4 address of the server",
			},
			"cloudinit": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "the cloudinit script associated with this server",
			},
			"public_ipv6": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "the public IPv6 address of the server, if enabled",
			},
			"state": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "the server state (running, stopped)",
			},
			"state_detail": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "scaleway description of the server state",
			},
		},
	}
}

func attachIP(scaleway *api.API, serverID, IPAddress string) error {
	ips, err := scaleway.GetIPS()
	if err != nil {
		return err
	}
	for _, ip := range ips {
		if ip.Address == IPAddress {
			log.Printf("[DEBUG] Attaching IP %q to server %q\n", ip.ID, serverID)
			return scaleway.AttachIP(ip.ID, serverID)
		}
	}
	return fmt.Errorf("Failed to find IP with ip %q to attach", IPAddress)
}

func resourceScalewayServerCreate(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient

	image := d.Get("image").(string)
	var req = api.ServerDefinition{
		Name:          d.Get("name").(string),
		Image:         String(image),
		Organization:  scaleway.Organization,
		EnableIPV6:    d.Get("enable_ipv6").(bool),
		SecurityGroup: d.Get("security_group").(string),
	}
	bootType, ok := d.GetOk("boot_type")
	if ok {
		req.BootType = bootType.(string)
	}

	req.DynamicIPRequired = Bool(d.Get("dynamic_ip_required").(bool))
	req.CommercialType = d.Get("type").(string)

	availabilities, err := scaleway.GetServerAvailabilities()
	if err != nil {
		log.Printf("[DEBUG] Unable to fetch server availability; won't validate availability: %q\n", err.Error())
	} else {
		typeAvailability, ok := availabilities[req.CommercialType]
		if !ok {
			// this will most likely happen for new instance types
			log.Printf("[DEBUG] no server availability for type %q. Ignoring\n", req.CommercialType)
		}
		if typeAvailability.Availability == api.InstanceTypeShortage {
			return fmt.Errorf("InstanceType %s is currently out of stock", req.CommercialType)
		}
	}

	if bootscript, ok := d.GetOk("bootscript"); ok {
		req.Bootscript = String(bootscript.(string))
	}

	if vs, ok := d.GetOk("volume"); ok {
		req.Volumes = make(map[string]string)

		volumes := vs.([]interface{})
		for i, v := range volumes {
			volume := v.(map[string]interface{})
			sizeInGB := uint64(volume["size_in_gb"].(int))

			if sizeInGB > 0 {
				v, err := scaleway.CreateVolume(api.VolumeDefinition{
					Size: sizeInGB * gb,
					Type: volume["type"].(string),
					Name: fmt.Sprintf("%s-%d", req.Name, sizeInGB),
				})
				if err != nil {
					return err
				}
				volume["volume_id"] = v.Identifier
				req.Volumes[fmt.Sprintf("%d", i+1)] = v.Identifier
			}
			volumes[i] = volume
		}
		_ = d.Set("volume", volumes)
	}

	if raw, ok := d.GetOk("tags"); ok {
		for _, tag := range raw.([]interface{}) {
			req.Tags = append(req.Tags, tag.(string))
		}
	}

	server, err := scaleway.CreateServer(req)
	if err != nil {
		return err
	}

	if v, ok := d.GetOk("cloudinit"); ok {
		if err := scaleway.PatchUserdata(server.Identifier, "cloud-init", []byte(v.(string)), false); err != nil {
			return err
		}
	}

	d.SetId(server.Identifier)
	if d.Get("state").(string) != "stopped" {
		err := startServer(scaleway, server)
		if err != nil {
			return err
		}

		if v, ok := d.GetOk("public_ip"); ok {
			if err := attachIP(scaleway, d.Id(), v.(string)); err != nil {
				return err
			}
		}
	}

	if err != nil {
		return err
	}

	return resourceScalewayServerRead(d, m)
}

func resourceScalewayServerRead(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient

	server, err := scaleway.GetServer(d.Id())
	if err != nil {
		if serr, ok := err.(api.APIError); ok {
			log.Printf("[DEBUG] Error reading server: %q\n", serr.APIMessage)

			if serr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}

		return err
	}

	cloudinit, err := scaleway.GetUserdata(server.Identifier, "cloud-init", false)
	if err == nil {
		_ = d.Set("cloudinit", cloudinit.String())
	} else {
		fmt.Printf("[DEBUG] unable to retrieve cloudinit configuration for server %q\n", d.Get("server"))
	}

	_ = d.Set("name", server.Name)
	_ = d.Set("image", server.Image.Identifier)
	_ = d.Set("type", server.CommercialType)
	_ = d.Set("enable_ipv6", server.EnableIPV6)
	_ = d.Set("private_ip", server.PrivateIP)
	_ = d.Set("public_ip", server.PublicAddress.IP)
	_ = d.Set("boot_type", server.BootType)

	if server.EnableIPV6 && server.IPV6 != nil {
		_ = d.Set("public_ipv6", server.IPV6.Address)
	}

	_ = d.Set("state", server.State)
	_ = d.Set("state_detail", server.StateDetail)
	_ = d.Set("tags", server.Tags)

	d.SetConnInfo(map[string]string{
		"type": "ssh",
		"host": server.PublicAddress.IP,
	})

	return nil
}

func resourceScalewayServerUpdate(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient

	var req api.ServerPatchDefinition
	if d.HasChange("name") {
		name := d.Get("name").(string)
		req.Name = &name
	}

	if d.HasChange("tags") {
		if raw, ok := d.GetOk("tags"); ok {
			var tags []string
			for _, tag := range raw.([]interface{}) {
				tags = append(tags, tag.(string))
			}
			req.Tags = &tags
		}
	}

	if d.HasChange("cloudinit") {
		if err := scaleway.PatchUserdata(d.Id(), "cloud-init", []byte(d.Get("cloudinit").(string)), false); err != nil {
			fmt.Printf("[DEBUG] unable to update cloud-init for server %q\n", d.Id())
		}
	}

	if d.HasChange("enable_ipv6") {
		req.EnableIPV6 = Bool(d.Get("enable_ipv6").(bool))
	}

	if d.HasChange("dynamic_ip_required") {
		req.DynamicIPRequired = Bool(d.Get("dynamic_ip_required").(bool))
	}

	if d.HasChange("security_group") {
		req.SecurityGroup = &api.SecurityGroupRef{
			Identifier: d.Get("security_group").(string),
		}
	}

	err := scaleway.PatchServer(d.Id(), req)

	if err != nil {
		return fmt.Errorf("Failed patching scaleway server: %q", err)
	}

	if d.HasChange("public_ip") {
		ips, err := scaleway.GetIPS()
		if err != nil {
			return err
		}
		if v, ok := d.GetOk("public_ip"); ok {
			for _, ip := range ips {
				if ip.Address == v.(string) {
					log.Printf("[DEBUG] Attaching IP %q to server %q\n", ip.ID, d.Id())
					if err := scaleway.AttachIP(ip.ID, d.Id()); err != nil {
						return err
					}
					break
				}
			}
		} else {
			for _, ip := range ips {
				if ip.Server != nil && ip.Server.Identifier == d.Id() {
					log.Printf("[DEBUG] Detaching IP %q to server %q\n", ip.ID, d.Id())
					if err := scaleway.DetachIP(ip.ID); err != nil {
						return err
					}
					break
				}
			}
		}
	}

	return resourceScalewayServerRead(d, m)
}

func resourceScalewayServerDelete(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient

	s, err := scaleway.GetServer(d.Id())
	if err != nil {
		return err
	}

	if s.State == "stopped" {
		return deleteStoppedServer(scaleway, s)
	}

	err = deleteRunningServer(scaleway, s)
	if err == nil {
		d.SetId("")
	}

	return err
}
