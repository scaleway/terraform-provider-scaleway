package scaleway

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayRdbUserBeta() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayRdbUserBetaCreate,
		Read:   resourceScalewayRdbUserBetaRead,
		Update: resourceScalewayRdbUserBetaUpdate,
		Delete: resourceScalewayRdbUserBetaDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validationUUIDorUUIDWithLocality(),
				Description:  "Instance on which the user is created",
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Database user name",
				Required:    true,
				ForceNew:    true,
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Database user password",
			},
			"is_admin": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Grant admin permissions to database user",
			},
			// Common
			"region": regionSchema(),
		},
	}
}

func resourceScalewayRdbUserBetaCreate(d *schema.ResourceData, m interface{}) error {
	rdbAPI, region, err := rdbAPIWithRegion(d, m)
	if err != nil {
		return err
	}
	instanceId := d.Get("instance_id").(string)
	createReq := &rdb.CreateUserRequest{
		Region:     region,
		InstanceID: expandID(instanceId),
		Name:       d.Get("name").(string),
		Password:   d.Get("password").(string),
		IsAdmin:    d.Get("is_admin").(bool),
	}

	res, err := rdbAPI.CreateUser(createReq)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s/%s/%s", region, expandID(instanceId), res.Name))

	return resourceScalewayRdbUserBetaRead(d, m)
}

func resourceScalewayRdbUserBetaRead(d *schema.ResourceData, m interface{}) error {
	rdbAPI, region, err := rdbAPIWithRegion(d, m)
	if err != nil {
		return err
	}

	regionName, instanceId, userName, err := resourceScalewayRdbUserBetaParseId(d.Id())

	if err != nil {
		return err
	}

	res, err := rdbAPI.ListUsers(&rdb.ListUsersRequest{
		Region:     region,
		InstanceID: instanceId,
		Name:       &userName,
	})

	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return err
	}

	var user = res.Users[0]
	_ = d.Set("instance_id", fmt.Sprintf("%s/%s", region, instanceId))
	_ = d.Set("name", user.Name)
	_ = d.Set("password", d.Get("password").(string)) // password are immutable
	_ = d.Set("is_admin", user.IsAdmin)

	d.SetId(fmt.Sprintf("%s/%s/%s", regionName, instanceId, user.Name))

	return nil
}

func resourceScalewayRdbUserBetaUpdate(d *schema.ResourceData, m interface{}) error {
	rdbAPI, region, err := rdbAPIWithRegion(d, m)
	if err != nil {
		return err
	}

	_, instanceId, userName, err := resourceScalewayRdbUserBetaParseId(d.Id())

	if err != nil {
		return err
	}

	req := &rdb.UpdateUserRequest{
		Region:     region,
		InstanceID: instanceId,
		Name:       userName,
	}

	if d.HasChange("password") {
		req.Password = scw.StringPtr(d.Get("password").(string))
	}
	if d.HasChange("is_admin") {
		req.IsAdmin = scw.BoolPtr(d.Get("is_admin").(bool))
	}

	_, err = rdbAPI.UpdateUser(req)
	if err != nil {
		return err
	}

	return resourceScalewayRdbUserBetaRead(d, m)
}

func resourceScalewayRdbUserBetaDelete(d *schema.ResourceData, m interface{}) error {
	rdbAPI, region, err := rdbAPIWithRegion(d, m)
	if err != nil {
		return err
	}

	_, instanceId, userName, err := resourceScalewayRdbUserBetaParseId(d.Id())

	if err != nil {
		return err
	}

	err = rdbAPI.DeleteUser(&rdb.DeleteUserRequest{
		Region:     region,
		InstanceID: instanceId,
		Name:       userName,
	})

	if err != nil && !is404Error(err) {
		return err
	}

	return nil
}

func resourceScalewayRdbUserBetaParseId(resourceId string) (region string, instanceId string, userName string, err error) {
	idParts := strings.Split(resourceId, "/")
	if len(idParts) != 3 {
		return "", "", "", fmt.Errorf("can't parse user resource id: %s", resourceId)
	}
	region, instanceId, userName = idParts[0], idParts[1], idParts[2]

	return
}
