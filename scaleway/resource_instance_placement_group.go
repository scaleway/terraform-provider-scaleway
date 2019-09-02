package scaleway

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
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
				Default:     instance.ComputeClusterPolicyTypeLowLatency.String(),
				Description: "The operating mode is selected by a policy_type",
			},
			"policy_mode": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     instance.ComputeClusterPolicyModeOptional,
				Description: "One of the two policy_mode may be selected: enforced or optional.",
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
	instanceApi, zone, err := getInstanceAPIWithZone(d, m)
	if err != nil {
		return err
	}

	name, ok := d.GetOk("name")
	if !ok {
		name = getRandomName("pg")
	}
	res, err := instanceApi.CreateComputeCluster(&instance.CreateComputeClusterRequest{
		Zone:         zone,
		Name:         name.(string),
		Organization: d.Get("organization_id").(string),
		PolicyMode:   instance.ComputeClusterPolicyMode(d.Get("policy_mode").(string)),
		PolicyType:   instance.ComputeClusterPolicyType(d.Get("policy_type").(string)),
	})
	if err != nil {
		return err
	}

	d.SetId(newZonedId(zone, res.ComputeCluster.ID))
	return resourceScalewayInstancePlacementGroupRead(d, m)
}

func resourceScalewayInstancePlacementGroupRead(d *schema.ResourceData, m interface{}) error {
	instanceApi, zone, ID, err := getInstanceAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return err
	}

	res, err := instanceApi.GetComputeCluster(&instance.GetComputeClusterRequest{
		Zone:             zone,
		ComputeClusterID: ID,
	})

	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("name", res.ComputeCluster.Name)
	d.Set("zone", string(zone))
	d.Set("organization_id", res.ComputeCluster.Organization)
	d.Set("policy_mode", res.ComputeCluster.PolicyMode.String())
	d.Set("policy_type", res.ComputeCluster.PolicyType.String())
	d.Set("policy_respected", res.ComputeCluster.PolicyRespected)

	return nil
}

func resourceScalewayInstancePlacementGroupUpdate(d *schema.ResourceData, m interface{}) error {
	instanceApi, zone, ID, err := getInstanceAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return err
	}
	req := &instance.UpdateComputeClusterRequest{
		Zone:             zone,
		ComputeClusterID: ID,
		PolicyMode:       instance.ComputeClusterPolicyMode(d.Get("policy_mode").(string)),
		PolicyType:       instance.ComputeClusterPolicyType(d.Get("policy_type").(string)),
	}

	hasChanged := d.HasChange("policy_mode") || d.HasChange("policy_type")

	if d.HasChange("name") {
		req.Name = String(d.Get("name").(string))
		hasChanged = true
	}
	if hasChanged {
		_, err = instanceApi.UpdateComputeCluster(req)
		if err != nil {
			return err
		}
	}

	return resourceScalewayInstancePlacementGroupRead(d, m)
}

func resourceScalewayInstancePlacementGroupDelete(d *schema.ResourceData, m interface{}) error {
	instanceApi, zone, ID, err := getInstanceAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return err
	}

	err = instanceApi.DeleteComputeCluster(&instance.DeleteComputeClusterRequest{
		Zone:             zone,
		ComputeClusterID: ID,
	})

	if err != nil && !is404Error(err) {
		return err
	}

	return nil
}
