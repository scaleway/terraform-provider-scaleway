package scaleway

import (
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/utils"
)

func resourceScalewayComputeInstanceServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayComputeInstanceServerCreate,
		Read:   resourceScalewayComputeInstanceServerRead,
		Update: resourceScalewayComputeInstanceServerUpdate,
		Delete: resourceScalewayComputeInstanceServerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of the server",
			},
			"image": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "The base image of the server", // TODO: add in doc example with UUID or distrib name
				ValidateFunc: validationUUID(),
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The instance type of the server", // TODO: link to scaleway pricing in the doc
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "The tags associated with the server",
			},
			"security_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The security group the server is attached to", // TODO: add this field in CreateServerRequest (proto)
			},
			//"root_volume": {
			//	Type:     schema.TypeMap,
			//	Optional: true,
			//	ForceNew: true,
			//	Elem: &schema.Resource{
			//		Schema: map[string]*schema.Schema{
			//			"size": {
			//				Type:     schema.TypeString,
			//				Optional: true,
			//			},
			//			"id": {
			//				Type:     schema.TypeString,
			//				Computed: true,
			//			},
			//		},
			//	},
			//	Description: "Root volume attached to the server on creation",
			//},
			"enable_ipv6": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "determines if IPv6 is enabled for the server",
			},
			"private_ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "the Scaleway internal IP address of the server", // TODO: add this field in CreateServerRequest (proto)
			},
			"public_ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "the public IPv4 address of the server",
			},
			"state": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "started",
				Description: "the state of the server should be (started, stopped, standby)",
				ValidateFunc: validation.StringInSlice([]string{
					"started",
					"stopped",
					"standby",
				}, false),
			},
			"user_data": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "The user data associated with the server", // TODO: document reserved keys (`cloud-init`)
			},
			"zone":       zoneSchema(),
			"project_id": projectIDSchema(),
		},
	}
}

func resourceScalewayComputeInstanceServerCreate(d *schema.ResourceData, m interface{}) error {
	instanceApi, zone, err := getInstanceAPIWithZone(d, m)
	if err != nil {
		return err
	}

	name, ok := d.GetOk("name")
	if !ok {
		name = getRandomName("srv")
	}
	req := &instance.CreateServerRequest{
		Zone:           zone,
		Name:           name.(string),
		Organization:   d.Get("project_id").(string),
		Image:          d.Get("image").(string),
		CommercialType: d.Get("type").(string),
		EnableIPv6:     d.Get("enable_ipv6").(bool),
		SecurityGroup:  d.Get("security_group_id").(string),
	}

	if raw, ok := d.GetOk("tags"); ok {
		for _, tag := range raw.([]interface{}) {
			req.Tags = append(req.Tags, tag.(string))
		}
	}

	//if vs, ok := d.GetOk("root_volume"); ok {
	//	req.Volumes = make(map[string]*instance.VolumeTemplate)
	//	req.Volumes["0"] = &instance.VolumeTemplate{
	//		ID:   v.(string),
	//		Name: namesgenerator.GetRandomName(),
	//	}
	//}

	res, err := instanceApi.CreateServer(req)
	if err != nil {
		return err
	}

	d.SetId(newZonedId(zone, res.Server.ID))

	// todo: add userdata

	// room for improvement: start the action in Create / Update, wait for the state in Read
	err = instanceApi.ServerActionAndWait(&instance.ServerActionAndWaitRequest{
		Zone:     zone,
		ServerID: res.Server.ID,
		Action:   stateToAction(d.Get("state").(string)),
		Timeout:  time.Minute * 10,
	})
	if err != nil {
		return err
	}

	return resourceScalewayComputeInstanceServerRead(d, m)
}

func resourceScalewayComputeInstanceServerRead(d *schema.ResourceData, m interface{}) error {
	instanceApi, zone, ID, err := getInstanceAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return err
	}

	response, err := instanceApi.GetServer(&instance.GetServerRequest{
		Zone:     zone,
		ServerID: ID,
	})
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("name", response.Server.Name)
	d.Set("image", response.Server.Image.ID)
	d.Set("type", response.Server.CommercialType)
	d.Set("tags", response.Server.Tags)
	d.Set("security_group_id", response.Server.SecurityGroup.ID)
	d.Set("enable_ipv6", response.Server.EnableIPv6)

	if response.Server.PrivateIP != nil {
		d.Set("private_ip", *response.Server.PrivateIP)
	}

	if response.Server.PublicIP != nil {
		d.Set("public_ip", response.Server.PublicIP.Address.String())
	}

	if response.Server.EnableIPv6 && response.Server.IPv6 != nil {
		d.Set("public_ipv6", response.Server.IPv6.Address.String())
	}

	// todo: set user data

	d.SetConnInfo(map[string]string{
		"type": "ssh",
		"host": response.Server.PublicIP.Address.String(),
	})

	return nil
}

func resourceScalewayComputeInstanceServerUpdate(d *schema.ResourceData, m interface{}) error {
	instanceApi, zone, ID, err := getInstanceAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return err
	}

	updateRequest := &instance.UpdateServerRequest{
		Zone:     zone,
		ServerID: ID,
	}

	if d.HasChange("name") {
		updateRequest.Name = utils.String(d.Get("name").(string))
	}

	if d.HasChange("tags") {
		updateRequest.Tags = utils.Strings(d.Get("tags").([]string))
	}

	if d.HasChange("security_group_id") {
		updateRequest.SecurityGroup = &instance.SecurityGroupSummary{
			ID:   d.Get("security_group_id").(string),
			Name: getRandomName("sg"),
		}
	}

	if d.HasChange("enable_ipv6") {
		updateRequest.EnableIPv6 = utils.Bool(d.Get("enable_ipv6").(bool))
	}

	_, err = instanceApi.UpdateServer(updateRequest)
	if err != nil {
		return err
	}

	if d.HasChange("state") {
		// room for improvement: start the action in Create / Update, wait for the state in Read
		err = instanceApi.ServerActionAndWait(&instance.ServerActionAndWaitRequest{
			Zone:     zone,
			ServerID: ID,
			Action:   stateToAction(d.Get("state").(string)),
			Timeout:  time.Minute * 10,
		})
		if err != nil {
			return err
		}
	}

	return resourceScalewayComputeInstanceServerRead(d, m)
}

func resourceScalewayComputeInstanceServerDelete(d *schema.ResourceData, m interface{}) error {
	instanceApi, zone, ID, err := getInstanceAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return err
	}

	err = instanceApi.ServerActionAndWait(&instance.ServerActionAndWaitRequest{
		Zone:     zone,
		ServerID: ID,
		Action:   instance.ServerActionTerminate,
		Timeout:  time.Minute * 10,
	})

	if err != nil && !is404Error(err) {
		return err
	}

	// TODO: if root volume remove_on_terminate

	return nil
}

func stateToAction(state string) instance.ServerAction {
	action := instance.ServerActionPoweron
	switch state {
	case "stopped":
		action = instance.ServerActionPoweroff
	case "standby":
		action = instance.ServerActionStopInPlace
	}

	return action
}
