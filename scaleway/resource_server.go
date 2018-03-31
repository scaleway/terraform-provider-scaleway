package scaleway

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	api "github.com/nicolai86/scaleway-sdk"
)

var commercialServerTypes []string

func resourceScalewayServer() *schema.Resource {
	return &schema.Resource{
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
								if value > 150 {
									errors = append(errors, fmt.Errorf("%q needs to be less than 150", k))
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
	scaleway := m.(*Client).scaleway

	mu.Lock()
	defer mu.Unlock()

	image := d.Get("image").(string)
	var req = api.ServerDefinition{
		Name:          d.Get("name").(string),
		Image:         String(image),
		Organization:  scaleway.Organization,
		EnableIPV6:    d.Get("enable_ipv6").(bool),
		SecurityGroup: d.Get("security_group").(string),
	}

	req.DynamicIPRequired = Bool(d.Get("dynamic_ip_required").(bool))
	req.CommercialType = d.Get("type").(string)

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
		d.Set("volume", volumes)
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

	d.SetId(server.Identifier)
	if d.Get("state").(string) != "stopped" {
		_, err = scaleway.PostServerAction(server.Identifier, "poweron")
		if err != nil {
			return err
		}

		err = waitForServerStartup(scaleway, server.Identifier)

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
	scaleway := m.(*Client).scaleway
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

	d.Set("name", server.Name)
	d.Set("image", server.Image.Identifier)
	d.Set("type", server.CommercialType)
	d.Set("enable_ipv6", server.EnableIPV6)
	d.Set("private_ip", server.PrivateIP)
	d.Set("public_ip", server.PublicAddress.IP)

	if server.EnableIPV6 && server.IPV6 != nil {
		d.Set("public_ipv6", server.IPV6.Address)
	}

	d.Set("state", server.State)
	d.Set("state_detail", server.StateDetail)
	d.Set("tags", server.Tags)

	d.SetConnInfo(map[string]string{
		"type": "ssh",
		"host": server.PublicAddress.IP,
	})

	return nil
}

func resourceScalewayServerUpdate(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway

	mu.Lock()
	defer mu.Unlock()

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

	if err := scaleway.PatchServer(d.Id(), req); err != nil {
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
	scaleway := m.(*Client).scaleway

	mu.Lock()
	defer mu.Unlock()

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
