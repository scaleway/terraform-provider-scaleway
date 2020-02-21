package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	baremetal "github.com/scaleway/scaleway-sdk-go/api/baremetal/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	sdkValidation "github.com/scaleway/scaleway-sdk-go/validation"
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
		Timeouts: &schema.ResourceTimeout{
			Create: &baremetalServerResourceTimeout,
			Delete: &baremetalServerResourceTimeout,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of the server",
			},
			"offer": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				Description:      "ID of the server type",
				DiffSuppressFunc: diffSuppressFuncLabelUUID,
			},
			"os_id": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The base image of the server",
				ValidateFunc: validationUUID(),
			},
			"ssh_key_ids": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validationUUID(),
				},
				Required:    true,
				Description: "Array of SSH key IDs allowed to SSH to the server",
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
				Description:  "Some description to associate to the server, max 255 characters",
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "Array of tags to associate with the server",
			},
			"zone":            zoneSchema(),
			"organization_id": organizationIDSchema(),

			// TODO: Remove deleted attributes at the end of the beta.
			"offer_id": {
				Type:     schema.TypeString,
				Optional: true,
				Removed:  "Please use offer instead",
			},
			"ips": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the IP",
						},
						"address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The IP address of the IP",
						},
						"reverse": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The Reverse of the IP",
						},
					},
				},
			},
			"domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceScalewayBaremetalServerBetaCreate(d *schema.ResourceData, m interface{}) error {
	baremetalAPI, zone, err := baremetalAPIWithZone(d, m)
	if err != nil {
		return err
	}

	offer := d.Get("offer").(string)
	if !sdkValidation.IsUUID(offer) {
		o, err := baremetalOfferByName(baremetalAPI, zone, offer)
		if err != nil {
			return err
		}
		offer = o.ID
	}

	createReq := &baremetal.CreateServerRequest{
		Zone:           zone,
		Name:           expandOrGenerateString(d.Get("name"), "bm"),
		OrganizationID: d.Get("organization_id").(string),
		Description:    d.Get("description").(string),
		OfferID:        offer,
	}
	if raw, ok := d.GetOk("tags"); ok {
		for _, tag := range raw.([]interface{}) {
			createReq.Tags = append(createReq.Tags, tag.(string))
		}
	}
	res, err := baremetalAPI.CreateServer(createReq)
	if err != nil {
		return err
	}

	d.SetId(newZonedId(zone, res.ID))

	_, err = baremetalAPI.WaitForServer(&baremetal.WaitForServerRequest{
		Zone:     zone,
		ServerID: res.ID,
		Timeout:  baremetalServerWaitForTimeout,
	})
	if err != nil {
		return err
	}

	installReq := &baremetal.InstallServerRequest{
		Zone:     zone,
		ServerID: res.ID,
		OsID:     d.Get("os_id").(string),
		Hostname: res.Name,
	}

	for _, sshKeyID := range d.Get("ssh_key_ids").([]interface{}) {
		installReq.SSHKeyIDs = append(installReq.SSHKeyIDs, sshKeyID.(string))
	}

	_, err = baremetalAPI.InstallServer(installReq)
	if err != nil {
		return err
	}

	_, err = baremetalAPI.WaitForServerInstall(&baremetal.WaitForServerInstallRequest{
		Zone:     zone,
		ServerID: res.ID,
		Timeout:  baremetalServerWaitForTimeout,
	})
	if err != nil {
		return err
	}

	return resourceScalewayBaremetalServerBetaRead(d, m)
}

func resourceScalewayBaremetalServerBetaRead(d *schema.ResourceData, m interface{}) error {
	baremetalAPI, zone, ID, err := baremetalAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return err
	}

	res, err := baremetalAPI.GetServer(&baremetal.GetServerRequest{
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

	offer, err := baremetalOfferByID(baremetalAPI, zone, res.OfferID)
	if err != nil {
		return err
	}

	_ = d.Set("name", res.Name)
	_ = d.Set("zone", string(zone))
	_ = d.Set("organization_id", res.OrganizationID)
	_ = d.Set("offer", flattenLabelUUID(offer.Name, offer.ID))
	_ = d.Set("tags", res.Tags)
	_ = d.Set("domain", res.Domain)
	_ = d.Set("ips", flattenBaremetalIPs(res.IPs))
	if res.Install != nil {
		_ = d.Set("os_id", res.Install.OsID)
		_ = d.Set("ssh_key_ids", res.Install.SSHKeyIDs)
	}
	_ = d.Set("description", res.Description)

	return nil
}

func resourceScalewayBaremetalServerBetaUpdate(d *schema.ResourceData, m interface{}) error {
	baremetalAPI, zone, ID, err := baremetalAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return err
	}
	req := &baremetal.UpdateServerRequest{
		Zone:     zone,
		ServerID: ID,
	}

	needUpdate := false

	if d.HasChange("name") {
		req.Name = scw.StringPtr(d.Get("name").(string))
		needUpdate = true
	}

	if d.HasChange("description") {
		req.Description = scw.StringPtr(d.Get("description").(string))
		needUpdate = true
	}

	if d.HasChange("tags") {
		var tags []string
		for _, tag := range d.Get("tags").([]interface{}) {
			tags = append(tags, tag.(string))
		}
		req.Tags = &tags
		needUpdate = true
	}

	if needUpdate {
		_, err = baremetalAPI.UpdateServer(req)
		if err != nil {
			return err
		}
	}

	if d.HasChange("os_id") || d.HasChange("ssh_key_ids") {
		installReq := &baremetal.InstallServerRequest{
			Zone:     zone,
			ServerID: ID,
			OsID:     d.Get("os_id").(string),
			Hostname: d.Get("name").(string),
		}

		for _, sshKeyID := range d.Get("ssh_key_ids").([]interface{}) {
			installReq.SSHKeyIDs = append(installReq.SSHKeyIDs, sshKeyID.(string))
		}

		_, err := baremetalAPI.InstallServer(installReq)
		if err != nil {
			return err
		}
	}

	return resourceScalewayBaremetalServerBetaRead(d, m)
}

func resourceScalewayBaremetalServerBetaDelete(d *schema.ResourceData, m interface{}) error {
	baremetalAPI, zone, ID, err := baremetalAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return err
	}

	_, err = baremetalAPI.DeleteServer(&baremetal.DeleteServerRequest{
		Zone:     zone,
		ServerID: ID,
	})

	if err != nil && !is404Error(err) {
		return err
	}

	_, err = baremetalAPI.WaitForServer(&baremetal.WaitForServerRequest{
		Zone:     zone,
		ServerID: ID,
		Timeout:  baremetalServerWaitForTimeout,
	})

	if is404Error(err) {
		return nil
	}

	return err
}
