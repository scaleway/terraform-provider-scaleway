package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayInstancePlacementGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayInstancePlacementGroupCreate,
		Read:   resourceScalewayInstancePlacementGroupRead,
		Update: resourceScalewayInstancePlacementGroupUpdate,
		Delete: resourceScalewayInstancePlacementGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of the placement group",
			},
			"policy_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     instance.PlacementGroupPolicyTypeMaxAvailability.String(),
				Description: "The operating mode is selected by a policy_type",
				ValidateFunc: validation.StringInSlice([]string{
					instance.PlacementGroupPolicyTypeLowLatency.String(),
					instance.PlacementGroupPolicyTypeMaxAvailability.String(),
				}, false),
			},
			"policy_mode": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     instance.PlacementGroupPolicyModeOptional,
				Description: "One of the two policy_mode may be selected: enforced or optional.",
				ValidateFunc: validation.StringInSlice([]string{
					instance.PlacementGroupPolicyModeOptional.String(),
					instance.PlacementGroupPolicyModeEnforced.String(),
				}, false),
			},
			"policy_respected": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Is true when the policy is respected.",
			},
			"zone":            zoneSchema(),
			"organization_id": organizationIDSchema(),
		},
	}
}

func resourceScalewayInstancePlacementGroupCreate(d *schema.ResourceData, m interface{}) error {
	instanceAPI, zone, err := instanceAPIWithZone(d, m)
	if err != nil {
		return err
	}

	res, err := instanceAPI.CreatePlacementGroup(&instance.CreatePlacementGroupRequest{
		Zone:         zone,
		Name:         expandOrGenerateString(d.Get("name"), "pg"),
		Organization: expandStringPtr(d.Get("organization_id")),
		PolicyMode:   instance.PlacementGroupPolicyMode(d.Get("policy_mode").(string)),
		PolicyType:   instance.PlacementGroupPolicyType(d.Get("policy_type").(string)),
	})
	if err != nil {
		return err
	}

	d.SetId(newZonedIDString(zone, res.PlacementGroup.ID))
	return resourceScalewayInstancePlacementGroupRead(d, m)
}

func resourceScalewayInstancePlacementGroupRead(d *schema.ResourceData, m interface{}) error {
	instanceAPI, zone, ID, err := instanceAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return err
	}

	res, err := instanceAPI.GetPlacementGroup(&instance.GetPlacementGroupRequest{
		Zone:             zone,
		PlacementGroupID: ID,
	})

	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return err
	}

	_ = d.Set("name", res.PlacementGroup.Name)
	_ = d.Set("zone", string(zone))
	_ = d.Set("organization_id", res.PlacementGroup.Organization)
	_ = d.Set("policy_mode", res.PlacementGroup.PolicyMode.String())
	_ = d.Set("policy_type", res.PlacementGroup.PolicyType.String())
	_ = d.Set("policy_respected", res.PlacementGroup.PolicyRespected)

	return nil
}

func resourceScalewayInstancePlacementGroupUpdate(d *schema.ResourceData, m interface{}) error {
	instanceAPI, zone, ID, err := instanceAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return err
	}
	req := &instance.UpdatePlacementGroupRequest{
		Zone:             zone,
		PlacementGroupID: ID,
		PolicyMode:       instance.PlacementGroupPolicyMode(d.Get("policy_mode").(string)),
		PolicyType:       instance.PlacementGroupPolicyType(d.Get("policy_type").(string)),
	}

	hasChanged := d.HasChange("policy_mode") || d.HasChange("policy_type")

	if d.HasChange("name") {
		req.Name = scw.StringPtr(d.Get("name").(string))
		hasChanged = true
	}
	if hasChanged {
		_, err = instanceAPI.UpdatePlacementGroup(req)
		if err != nil {
			return err
		}
	}

	return resourceScalewayInstancePlacementGroupRead(d, m)
}

func resourceScalewayInstancePlacementGroupDelete(d *schema.ResourceData, m interface{}) error {
	instanceAPI, zone, ID, err := instanceAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return err
	}

	err = instanceAPI.DeletePlacementGroup(&instance.DeletePlacementGroupRequest{
		Zone:             zone,
		PlacementGroupID: ID,
	})

	if err != nil && !is404Error(err) {
		return err
	}

	return nil
}
