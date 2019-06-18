package scaleway

import (
	"strconv"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
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

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: defaultFuncRandomName("srv"),
				Description: "The name of the server",
			},
			"image": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The base image of the server",
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
				Description: "the server action (poweron, poweroff)",
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
	meta := m.(*Meta)
	instanceApi := instance.NewAPI(meta.scwClient)

	zone, err := getZone(d, meta)
	if err != nil {
		return err
	}

	req := &instance.CreateServerRequest{
		Zone:           zone,
		Organization:   d.Get("project_id").(string),
		CommercialType: d.Get("commercial_type").(string),
		EnableIPv6:     d.Get("enable_ipv6").(bool),
		SecurityGroup:  d.Get("security_group_id").(string),
	}

	name, ok := d.GetOk("name")
	if !ok {
		name = namesgenerator.GetRandomName()
	}
	req.Name = name.(string)

	if raw, ok := d.GetOk("tags"); ok {
		for _, tag := range raw.([]interface{}) {
			req.Tags = append(req.Tags, tag.(string))
		}
	}

	if vs, ok := d.GetOk("volumes"); ok {
		req.Volumes = make(map[string]*instance.VolumeTemplate)

		for i, v := range vs.([]interface{}) {
			req.Volumes[strconv.Itoa(i)] = &instance.VolumeTemplate{
				ID:   v.(string),
				Name: namesgenerator.GetRandomName(),
			}
		}
	}

	res, err := instanceApi.CreateServer(req)
	if err != nil {
		return err
	}

	d.SetId(res.Server.ID)

	// todo: add userdata

	action := instance.ServerActionPoweron
	switch d.Get("state").(string) {
	case "stopped":
		action = instance.ServerActionPoweroff
	case "standby":
		action = instance.ServerActionStopInPlace
	}

	err = instanceApi.ServerActionAndWait(&instance.ServerActionAndWaitRequest{
		Zone:     zone,
		ServerID: res.Server.ID,
		Action:   action,
		Timeout:  time.Minute * 10,
	})
	if err != nil {
		return err
	}

	return nil // todo: read
}
