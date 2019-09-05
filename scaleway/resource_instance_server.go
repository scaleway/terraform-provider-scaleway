package scaleway

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/api/marketplace/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayInstanceServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayInstanceServerCreate,
		Read:   resourceScalewayInstanceServerRead,
		Update: resourceScalewayInstanceServerUpdate,
		Delete: resourceScalewayInstanceServerDelete,
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
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The UUID or the label of the base image used by the server",
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
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: suppressLocality,
				Description:      "The security group the server is attached to",
			},
			"placement_group_id": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressLocality,
				Description:      "The placement group the server is attached to",
			},
			"placement_group_policy_respected": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True when the placement group policy is respected",
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
							Type:        schema.TypeInt,
							Optional:    true,
							Computed:    true,
							ForceNew:    true, // todo: don't force new but stop server and create new volume instead
							Description: "Size of the root volume in gigabytes",
						},
						"delete_on_termination": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Force deletion of the root volume on instance termination",
						},
						"volume_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Volume ID of the root volume",
						},
					},
				},
			},
			"additional_volume_ids": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateFunc:     validationUUIDorUUIDWithLocality(),
					DiffSuppressFunc: suppressLocality,
				},
				Optional:    true,
				Description: "The additional volumes attached to the server",
			},
			"enable_ipv6": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Determines if IPv6 is enabled for the server",
			},
			"private_ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Scaleway internal IP address of the server",
			},
			"public_ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The public IPv4 address of the server",
			},
			"disable_public_ip": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Disable dynamic ip on the server",
			},
			"state": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     ServerStateStarted,
				Description: "The state of the server should be: started, stopped, standby",
				ValidateFunc: validation.StringInSlice([]string{
					ServerStateStarted,
					ServerStateStopped,
					ServerStateStandby,
				}, false),
			},
			"cloud_init": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The cloud init script associated with this server",
				ValidateFunc: validation.StringLenBetween(0, 127998),
			},
			"user_data": {
				Type:        schema.TypeSet,
				Optional:    true,
				MaxItems:    98,
				Description: "The user data associated with the server", // TODO: document reserved keys (`cloud-init`)
				Set:         userDataHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validationStringNotInSlice([]string{"cloud-init"}, true),
							Description:  "A user data key, the value \"cloud-init\" is not allowed",
						},
						"value": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 127998),
							Description:  "A user value",
						},
					},
				},
			},
			"zone":            zoneSchema(),
			"organization_id": organizationIDSchema(),
		},
	}
}

func resourceScalewayInstanceServerCreate(d *schema.ResourceData, m interface{}) error {
	instanceAPI, zone, err := getInstanceAPIWithZone(d, m)
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

	commercialType := d.Get("type").(string)

	image := expandID(d.Get("image"))
	if !isUUID(image) {
		instanceAPI := marketplace.NewAPI(m.(*Meta).scwClient)
		imageUUID, err := instanceAPI.GetLocalImageIDByLabel(&marketplace.GetLocalImageIDByLabelRequest{
			CommercialType: commercialType,
			Zone:           zone,
			ImageLabel:     image,
		})
		if err != nil {
			return fmt.Errorf("invalid image '%s': it must be an UUID or a valid image label", image)
		}
		image = imageUUID
	}

	req := &instance.CreateServerRequest{
		Zone:              zone,
		Name:              name.(string),
		Organization:      d.Get("organization_id").(string),
		Image:             image,
		CommercialType:    commercialType,
		EnableIPv6:        d.Get("enable_ipv6").(bool),
		SecurityGroup:     expandID(d.Get("security_group_id")),
		DynamicIPRequired: Bool(!d.Get("disable_public_ip").(bool)),
	}

	if placementGroupID, ok := d.GetOk("placement_group_id"); ok {
		req.ComputeCluster = expandID(placementGroupID)
	}

	if raw, ok := d.GetOk("tags"); ok {
		for _, tag := range raw.([]interface{}) {
			req.Tags = append(req.Tags, tag.(string))
		}
	}

	req.Volumes = make(map[string]*instance.VolumeTemplate)
	if size, ok := d.GetOk("root_volume.0.size_in_gb"); ok {
		req.Volumes["0"] = &instance.VolumeTemplate{
			Size: scw.Size(uint64(size.(int)) * gb),
		}
	}

	if raw, ok := d.GetOk("additional_volume_ids"); ok {
		for i, volumeID := range raw.([]interface{}) {
			req.Volumes[strconv.Itoa(i+1)] = &instance.VolumeTemplate{
				ID:   expandID(volumeID),
				Name: getRandomName("vol"),
			}
		}
	}

	res, err := instanceAPI.CreateServer(req)
	if err != nil {
		return err
	}

	d.SetId(newZonedId(zone, res.Server.ID))

	////
	// Set user data
	////
	userDataRequests := &instance.SetAllServerUserDataRequest{
		Zone:     zone,
		ServerID: res.Server.ID,
		UserData: make(map[string]io.Reader),
	}

	if allUserData, ok := d.GetOk("user_data"); ok {
		userDataSet := allUserData.(*schema.Set)
		for _, rawUserData := range userDataSet.List() {
			userData := rawUserData.(map[string]interface{})
			userDataRequests.UserData[userData["key"].(string)] = bytes.NewBufferString(userData["value"].(string))
		}
	}

	// cloud init script is set in user data
	if cloudInit, ok := d.GetOk("cloud_init"); ok {
		userDataRequests.UserData["cloud-init"] = bytes.NewBufferString(cloudInit.(string))
	}

	if len(userDataRequests.UserData) > 0 {
		err := instanceAPI.SetAllServerUserData(userDataRequests)
		if err != nil {
			return err
		}
	}

	err = reachState(instanceAPI, zone, res.Server.ID, ServerStateStopped, d.Get("state").(string), false)
	if err != nil {
		return err
	}

	return resourceScalewayInstanceServerRead(d, m)
}

func resourceScalewayInstanceServerRead(d *schema.ResourceData, m interface{}) error {
	instanceAPI, zone, ID, err := getInstanceAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return err
	}

	////
	// Read Server
	////
	response, err := instanceAPI.GetServer(&instance.GetServerRequest{
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
	state, err := serverStateFlatten(response.Server.State)
	if err != nil {
		return err
	} else {
		d.Set("state", state)
	}

	d.Set("zone", string(zone))
	d.Set("name", response.Server.Name)
	d.Set("type", response.Server.CommercialType)
	d.Set("tags", response.Server.Tags)
	d.Set("security_group_id", response.Server.SecurityGroup.ID)
	d.Set("enable_ipv6", response.Server.EnableIPv6)
	d.Set("disable_public_ip", !response.Server.DynamicIPRequired)

	// TODO: If image is a label, check that response.Server.Image.ID match the label.
	// It could be useful if the user edit the image with another tool.
	if response.Server.Image != nil && isUUID(d.Get("image").(string)) {
		d.Set("image", response.Server.Image.ID)
	}

	if response.Server.ComputeCluster != nil {
		d.Set("placement_group_policy_respected", response.Server.ComputeCluster.PolicyRespected)
	}

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

	var additionalVolumesIDs []string
	for i, volume := range orderVolumes(response.Server.Volumes) {
		if i == 0 {
			rootVolume := map[string]interface{}{}

			vs, ok := d.Get("root_volume").([]map[string]interface{})
			if ok && len(vs) > 0 {
				rootVolume = vs[0]
			}

			rootVolume["volume_id"] = volume.ID
			rootVolume["size_in_gb"] = int(uint64(volume.Size) / gb)

			if _, exist := rootVolume["delete_on_termination"]; !exist {
				rootVolume["delete_on_termination"] = true // default value does not work on list
			}

			d.Set("root_volume", []map[string]interface{}{rootVolume})
		} else {
			additionalVolumesIDs = append(additionalVolumesIDs, volume.ID)
		}
	}
	d.Set("additional_volume_ids", additionalVolumesIDs)

	////
	// Read server user data
	////
	allUserData, err := instanceAPI.GetAllServerUserData(&instance.GetAllServerUserDataRequest{
		Zone:     zone,
		ServerID: ID,
	})

	var userDataList []interface{}
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
	if len(userDataList) > 0 {
		d.Set("user_data", schema.NewSet(userDataHash, userDataList))
	}

	return nil
}

func resourceScalewayInstanceServerUpdate(d *schema.ResourceData, m interface{}) error {
	instanceAPI, zone, ID, err := getInstanceAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return err
	}

	var forceReboot bool

	////
	// Update the server
	////
	previousState, nextState := d.GetChange("state") // the previous state of the server might change when updating volumes
	updateRequest := &instance.UpdateServerRequest{
		Zone:     zone,
		ServerID: ID,
	}

	if d.HasChange("name") {
		updateRequest.Name = scw.StringPtr(d.Get("name").(string))
	}

	if d.HasChange("tags") {
		updateRequest.Tags = scw.StringsPtr(d.Get("tags").([]string))
	}

	if d.HasChange("security_group_id") {
		updateRequest.SecurityGroup = &instance.SecurityGroupTemplate{
			ID:   expandID(d.Get("security_group_id")),
			Name: getRandomName("sg"), // this value will be ignored by the API
		}
	}

	if d.HasChange("enable_ipv6") {
		updateRequest.EnableIPv6 = scw.BoolPtr(d.Get("enable_ipv6").(bool))
	}

	if d.HasChange("disable_public_ip") {
		updateRequest.DynamicIPRequired = scw.BoolPtr(!d.Get("disable_public_ip").(bool))
	}

	volumes := map[string]*instance.VolumeTemplate{}

	if raw, ok := d.GetOk("additional_volume_ids"); d.HasChange("additional_volume_ids") && ok {
		volumes["0"] = &instance.VolumeTemplate{ID: d.Get("root_volume.0.volume_id").(string), Name: getRandomName("vol")} // name is ignored by the API, any name will work here

		for i, volumeID := range raw.([]interface{}) {
			volumes[strconv.Itoa(i+1)] = &instance.VolumeTemplate{
				ID:   expandID(volumeID),
				Name: getRandomName("vol"), // name is ignored by the API, any name will work here
			}
		}

		updateRequest.Volumes = &volumes
	}

	if d.HasChange("placement_group_id") {
		placementGroupID := expandID(d.Get("placement_group_id"))
		if placementGroupID == "" {
			updateRequest.ComputeCluster = &instance.NullableStringValue{Null: true}
		} else {
			forceReboot = true
			updateRequest.ComputeCluster = &instance.NullableStringValue{Value: placementGroupID}
		}
	}

	var updateResponse *instance.UpdateServerResponse

	err = resource.Retry(ServerRetryFuncTimeout, func() *resource.RetryError {
		updateResponse, err = instanceAPI.UpdateServer(updateRequest)
		if isSDKResponseError(err, http.StatusBadRequest, "Instance must be powered off to change local volumes") {
			err = reachState(instanceAPI, zone, ID, previousState.(string), ServerStateStopped, false)
			if err != nil && !isSDKResponseError(err, http.StatusBadRequest, "server should be running") {
				return resource.NonRetryableError(err)
			}
			return resource.RetryableError(fmt.Errorf("server is being powered off"))
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	previousState, err = serverStateFlatten(updateResponse.Server.State)
	if err != nil {
		return err
	}

	////
	// Update server user data
	////
	if d.HasChange("cloud_init") || d.HasChange("user_data") {

		userDataRequests := &instance.SetAllServerUserDataRequest{
			Zone:     zone,
			ServerID: ID,
			UserData: make(map[string]io.Reader),
		}

		if allUserData, ok := d.GetOk("user_data"); ok {
			userDataSet := allUserData.(*schema.Set)
			for _, rawUserData := range userDataSet.List() {
				userData := rawUserData.(map[string]interface{})
				userDataRequests.UserData[userData["key"].(string)] = bytes.NewBufferString(userData["value"].(string))
			}
		}

		// cloud init script is set in user data
		if cloudInit, ok := d.GetOk("cloud_init"); ok {
			userDataRequests.UserData["cloud-init"] = bytes.NewBufferString(cloudInit.(string))
			forceReboot = true // instance must reboot when cloud init script change
		}

		err := instanceAPI.SetAllServerUserData(userDataRequests)
		if err != nil {
			return err
		}

	}

	// reach expected state
	err = reachState(instanceAPI, zone, ID, previousState.(string), nextState.(string), forceReboot)
	if err != nil {
		return err
	}

	return resourceScalewayInstanceServerRead(d, m)
}

func resourceScalewayInstanceServerDelete(d *schema.ResourceData, m interface{}) error {
	instanceAPI, zone, ID, err := getInstanceAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return err
	}

	// reach stopped state
	err = reachState(instanceAPI, zone, ID, d.Get("state").(string), ServerStateStopped, false)
	if is404Error(err) {
		return nil
	}
	if err != nil {
		return err
	}

	err = instanceAPI.DeleteServer(&instance.DeleteServerRequest{
		Zone:     zone,
		ServerID: ID,
	})

	if err != nil && !is404Error(err) {
		return err
	}

	if d.Get("root_volume.0.delete_on_termination").(bool) {
		err = instanceAPI.DeleteVolume(&instance.DeleteVolumeRequest{
			Zone:     zone,
			VolumeID: d.Get("root_volume.0.volume_id").(string),
		})
		if err != nil && !is404Error(err) {
			return err
		}
	}

	return nil
}
