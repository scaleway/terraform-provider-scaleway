package scaleway

import (
	"bytes"
	"fmt"
	"io/ioutil"
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
			"image_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "The base image of the server", // TODO: add in doc example with UUID
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
				Description: "The security group the server is attached to",
			},
			"root_volume": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Description: "Root volume attached to the server on creation",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"size_in_gb": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
							ForceNew: true, // TODO: don't force new but stop server and create new volume instead
						},
						"delete_on_termination": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"volume_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validationUUID(),
						},
					},
				},
			},
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
				Default:     ServerStateStarted,
				Description: "the state of the server should be: started, stopped, standby",
				ValidateFunc: validation.StringInSlice([]string{
					ServerStateStarted,
					ServerStateStopped,
					ServerStateStandby,
				}, false),
			},
			"cloud_init": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "the cloud init script associated with this server",
				ValidateFunc: validation.StringLenBetween(0, 127998),
			},
			"user_data": {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				MaxItems:    98,
				Description: "the user data associated with the server", // TODO: document reserved keys (`cloud-init`)
				Set:         schemaSetUserData,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validationStringNotInSlice([]string{"cloud-init"}, true),
						},
						"value": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 127998),
						},
					},
				},
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

	////
	// Create the server
	////
	name, ok := d.GetOk("name")
	if !ok {
		name = getRandomName("srv")
	}
	req := &instance.CreateServerRequest{
		Zone:           zone,
		Name:           name.(string),
		Organization:   d.Get("project_id").(string),
		Image:          d.Get("image_id").(string),
		CommercialType: d.Get("type").(string),
		EnableIPv6:     d.Get("enable_ipv6").(bool),
		SecurityGroup:  d.Get("security_group_id").(string),
	}

	if raw, ok := d.GetOk("tags"); ok {
		for _, tag := range raw.([]interface{}) {
			req.Tags = append(req.Tags, tag.(string))
		}
	}

	if size, ok := d.GetOk("root_volume.0.size_in_gb"); ok {
		req.Volumes = make(map[string]*instance.VolumeTemplate)
		req.Volumes["0"] = &instance.VolumeTemplate{
			Size: uint64(size.(int)) * gb,
		}
	}

	res, err := instanceApi.CreateServer(req)
	if err != nil {
		return err
	}

	d.SetId(newZonedId(zone, res.Server.ID))

	////
	// Set user data
	////
	var userDataRequests []*instance.SetServerUserDataRequest

	// cloud init script is set in user data
	if cloudInit, ok := d.GetOk("cloud_init"); ok {
		userDataRequests = append(userDataRequests, &instance.SetServerUserDataRequest{
			Zone:     zone,
			ServerID: res.Server.ID,
			Key:      "cloud-init",
			Content:  bytes.NewBufferString(cloudInit.(string)),
		})
	}

	if allUserData, ok := d.GetOk("user_data"); ok {
		userDataSet := allUserData.(*schema.Set)
		for _, rawUserData := range userDataSet.List() {
			userData := rawUserData.(map[string]interface{})
			userDataRequests = append(userDataRequests, &instance.SetServerUserDataRequest{
				Zone:     zone,
				ServerID: res.Server.ID,
				Key:      userData["key"].(string),
				Content:  bytes.NewBufferString(userData["value"].(string)),
			})
		}
	}

	for _, userDataRequest := range userDataRequests {
		err := instanceApi.SetServerUserData(userDataRequest)
		if err != nil {
			return err
		}
	}

	////
	// Execute server action to reach the expected state
	////
	for _, action := range stateToAction(ServerStateStopped, d.Get("state").(string)) {
		err = instanceApi.ServerActionAndWait(&instance.ServerActionAndWaitRequest{
			Zone:     zone,
			ServerID: res.Server.ID,
			Action:   action,
			Timeout:  time.Minute * 10,
		})
		if err != nil {
			return err
		}
	}

	return resourceScalewayComputeInstanceServerRead(d, m)
}

func resourceScalewayComputeInstanceServerRead(d *schema.ResourceData, m interface{}) error {
	instanceApi, zone, ID, err := getInstanceAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return err
	}

	d.Set("zone", string(zone))

	////
	// Read Server
	////
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
	switch response.Server.State {
	case instance.ServerStateStopped:
		d.Set("state", ServerStateStopped)
	case instance.ServerStateStoppedInPlace:
		d.Set("state", ServerStateStandby)
	case instance.ServerStateRunning:
		d.Set("state", ServerStateStarted)
	case instance.ServerStateLocked:
		return fmt.Errorf("server is locked, please contact Scaleway support: https://console.scaleway.com/support/tickets")
	default:
		return fmt.Errorf("server is in an invalid state, someone else might be executing action at the same time")
	}

	d.Set("name", response.Server.Name)
	d.Set("image_id", response.Server.Image.ID)
	d.Set("type", response.Server.CommercialType)
	d.Set("tags", response.Server.Tags)
	d.Set("security_group_id", response.Server.SecurityGroup.ID)
	d.Set("enable_ipv6", response.Server.EnableIPv6)

	if response.Server.PrivateIP != nil {
		d.Set("private_ip", *response.Server.PrivateIP)
	}

	if response.Server.PublicIP != nil {
		d.Set("public_ip", response.Server.PublicIP.Address.String())
		d.SetConnInfo(map[string]string{
			"type": "ssh",
			"host": response.Server.PublicIP.Address.String(),
		})
	}

	if response.Server.EnableIPv6 && response.Server.IPv6 != nil {
		d.Set("public_ipv6", response.Server.IPv6.Address.String())
	}

	if vs, ok := response.Server.Volumes["0"]; ok {
		rootVolume := expandRootVolume(d.Get("root_volume"))
		rootVolume["volume_id"] = vs.ID
		rootVolume["size_in_gb"] = int(vs.Size / gb)
		d.Set("root_volume", []map[string]interface{}{rootVolume})
	}

	////
	// Read server user data
	////
	allUserData, err := instanceApi.GetAllServerUserData(&instance.GetAllServerUserDataRequest{
		Zone:     zone,
		ServerID: ID,
	})

	userDataList := []interface{}{}
	for key, value := range allUserData.UserData {
		userData, err := ioutil.ReadAll(value)
		if err != nil {
			return err
		}
		if key != "cloud-init" {
			userDataList = append(userDataList, map[string]interface{}{
				"key":   key,
				"value": string(userData),
			})
		} else {
			d.Set("cloud_init", string(userData))
		}
	}
	d.Set("user_data", schema.NewSet(schemaSetUserData, userDataList))

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
			Name: getRandomName("sg"), // this value will be ignored by the API
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
		previousState, nextState := d.GetChange("state")
		for _, action := range stateToAction(previousState.(string), nextState.(string)) {
			err = instanceApi.ServerActionAndWait(&instance.ServerActionAndWaitRequest{
				Zone:     zone,
				ServerID: ID,
				Action:   action,
				Timeout:  ServerWaitForTimeout,
			})
			if err != nil && !is404Error(err) {
				return err
			}
		}
	}

	return resourceScalewayComputeInstanceServerRead(d, m)
}

func resourceScalewayComputeInstanceServerDelete(d *schema.ResourceData, m interface{}) error {
	instanceApi, zone, ID, err := getInstanceAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return err
	}

	if d.Get("state").(string) != ServerStateStopped {
		err = instanceApi.ServerActionAndWait(&instance.ServerActionAndWaitRequest{
			Zone:     zone,
			ServerID: ID,
			Action:   instance.ServerActionPoweroff,
			Timeout:  ServerWaitForTimeout,
		})
		if is404Error(err) {
			return nil
		}
		if err != nil {
			return err
		}
	}

	err = instanceApi.DeleteServer(&instance.DeleteServerRequest{
		Zone:     zone,
		ServerID: ID,
	})

	if err != nil && !is404Error(err) {
		return err
	}

	if d.Get("root_volume.0.delete_on_termination").(bool) {
		err = instanceApi.DeleteVolume(&instance.DeleteVolumeRequest{
			Zone:     zone,
			VolumeID: d.Get("root_volume.0.volume_id").(string),
		})
		if err != nil && !is404Error(err) {
			return err
		}
	}

	return nil
}

func stateToAction(previousState, nextState string) []instance.ServerAction {
	transitionMap := map[[2]string][]instance.ServerAction{
		{ServerStateStopped, ServerStateStopped}: {},
		{ServerStateStopped, ServerStateStarted}: {instance.ServerActionPoweron},
		{ServerStateStopped, ServerStateStandby}: {instance.ServerActionPoweron, instance.ServerActionStopInPlace},
		{ServerStateStarted, ServerStateStopped}: {instance.ServerActionPoweroff},
		{ServerStateStarted, ServerStateStarted}: {},
		{ServerStateStarted, ServerStateStandby}: {instance.ServerActionStopInPlace},
		{ServerStateStandby, ServerStateStopped}: {instance.ServerActionPoweroff},
		{ServerStateStandby, ServerStateStarted}: {instance.ServerActionPoweron},
		{ServerStateStandby, ServerStateStandby}: {},
	}

	return transitionMap[[2]string{previousState, nextState}]
}
