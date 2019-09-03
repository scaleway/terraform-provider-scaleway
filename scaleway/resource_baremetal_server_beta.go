package scaleway

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	baremetal "github.com/scaleway/scaleway-sdk-go/api/baremetal/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayBaremetalServerBeta() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayBaremetalServerBetaCreate,
		Read:   resourceScalewayBaremetalServerBetaRead,
		Update: resourceScalewayBaremetalServerBetaUpdate,
		Delete: resourceScalewayBaremetalServerBetaDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of the server.",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "ID of the server type.",
				ValidateFunc: validationUUID(),
			},
			"image_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "The base image of the server", // TODO: add in doc example with UUID
				ValidateFunc: validationUUID(),
			},
			"ssh_key_ids": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validationUUID(),
				},
				Optional:    true,
				Description: "Array of SSH key IDs allowed to SSH to the server.",
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
				Description:  "Some description to associate to the server, max 255 characters.",
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "Array of tags to associate with the server.",
			},
			"zone":            zoneSchema(),
			"organization_id": organizationIDSchema(),
		},
	}
}

func resourceScalewayBaremetalServerBetaCreate(d *schema.ResourceData, m interface{}) error {
	baremetalApi, zone, err := getBaremetalAPIWithZone(d, m)
	if err != nil {
		return err
	}

	name, ok := d.GetOk("name")
	if !ok {
		name = getRandomName("bm")
	}
	createReq := &baremetal.CreateServerRequest{
		Zone:           zone,
		Name:           name.(string),
		OrganizationID: d.Get("project_id").(string),
		Description:    d.Get("description").(string),
		OfferID:        d.Get("type").(string),
	}
	if raw, ok := d.GetOk("tags"); ok {
		for _, tag := range raw.([]interface{}) {
			createReq.Tags = append(createReq.Tags, tag.(string))
		}
	}
	res, err := baremetalApi.CreateServer(createReq)
	if err != nil {
		return err
	}

	d.SetId(newZonedId(zone, res.ID))

	_, err = baremetalApi.WaitForServer(&baremetal.WaitForServerRequest{
		Zone:     zone,
		ServerID: res.ID,
		Timeout:  BaremetalServerWaitForTimeout,
	})
	if err != nil {
		return err
	}

	installReq := &baremetal.InstallServerRequest{
		Zone:     zone,
		ServerID: res.ID,
		OsID:     d.Get("image_id").(string),
		Hostname: res.Name,
	}

	if raw, ok := d.GetOk("ssh_key_ids"); ok {
		for _, sshKeyID := range raw.([]interface{}) {
			installReq.SSHKeyIds = append(installReq.SSHKeyIds, sshKeyID.(string))
		}
	} else {
		// TODO: pull all user ssh keys
	}

	_, err = baremetalApi.InstallServer(installReq)
	if err != nil {
		return err
	}

	_, err = baremetalApi.WaitForServer(&baremetal.WaitForServerRequest{
		Zone:     zone,
		ServerID: res.ID,
		Timeout:  BaremetalServerWaitForTimeout,
	})
	if err != nil {
		return err
	}

	return resourceScalewayBaremetalServerBetaRead(d, m)
}

func resourceScalewayBaremetalServerBetaRead(d *schema.ResourceData, m interface{}) error {
	baremetalApi, zone, ID, err := getBaremetalAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return err
	}

	res, err := baremetalApi.GetServer(&baremetal.GetServerRequest{
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

	d.Set("name", res.Name)
	d.Set("zone", string(zone))
	d.Set("project_id", res.OrganizationID)
	d.Set("tags", res.Tags)
	d.Set("type", res.OfferID)
	d.Set("image_id", res.Install.OsID)
	d.Set("ssh_key_ids", res.Install.SSHKeyIds)
	d.Set("description", res.Description)

	return nil
}

func resourceScalewayBaremetalServerBetaUpdate(d *schema.ResourceData, m interface{}) error {
	baremetalApi, zone, ID, err := getBaremetalAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return err
	}
	req := &baremetal.UpdateServerRequest{
		Zone:     zone,
		ServerID: ID,
	}

	hasChanged := false

	if d.HasChange("name") {
		req.Name = scw.StringPtr(d.Get("name").(string))
		hasChanged = true
	}

	if d.HasChange("description") {
		req.Description = scw.StringPtr(d.Get("description").(string))
		hasChanged = true
	}

	if d.HasChange("tags") {
		req.Tags = scw.StringsPtr(d.Get("tags").([]string))
		hasChanged = true
	}

	if hasChanged {
		_, err = baremetalApi.UpdateServer(req)
		if err != nil {
			return err
		}
	}

	return resourceScalewayBaremetalServerBetaRead(d, m)
}

func resourceScalewayBaremetalServerBetaDelete(d *schema.ResourceData, m interface{}) error {
	baremetalApi, zone, ID, err := getBaremetalAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return err
	}

	_, err = baremetalApi.DeleteServer(&baremetal.DeleteServerRequest{
		Zone:     zone,
		ServerID: ID,
	})

	if err != nil && !is404Error(err) {
		return err
	}

	_, err = baremetalApi.WaitForServer(&baremetal.WaitForServerRequest{
		Zone:     zone,
		ServerID: ID,
		Timeout:  BaremetalServerWaitForTimeout,
	})

	return err
}
