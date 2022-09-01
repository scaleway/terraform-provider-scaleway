package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayInstanceSecurityGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayInstanceSecurityGroupCreate,
		ReadContext:   resourceScalewayInstanceSecurityGroupRead,
		UpdateContext: resourceScalewayInstanceSecurityGroupUpdate,
		DeleteContext: resourceScalewayInstanceSecurityGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultInstanceSecurityGroupTimeout),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of the security group",
			},
			"stateful": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "The stateful value of the security group",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the security group",
			},
			"inbound_default_policy": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "accept",
				Description: "Default inbound traffic policy for this security group",
				ValidateFunc: validation.StringInSlice([]string{
					instance.SecurityGroupPolicyAccept.String(),
					instance.SecurityGroupPolicyDrop.String(),
				}, false),
			},
			"outbound_default_policy": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "accept",
				Description: "Default outbound traffic policy for this security group",
				ValidateFunc: validation.StringInSlice([]string{
					instance.SecurityGroupPolicyAccept.String(),
					instance.SecurityGroupPolicyDrop.String(),
				}, false),
			},
			"enable_default_security": {
				Type:        schema.TypeBool,
				Description: "Enable blocking of SMTP on IPv4 and IPv6",
				Optional:    true,
				Default:     true,
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "The tags associated with the security group",
			},
			"zone":            zoneSchema(),
			"organization_id": organizationIDSchema(),
			"project_id":      projectIDSchema(),
		},
	}
}

func resourceScalewayInstanceSecurityGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceAPI, zone, err := instanceAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &instance.CreateSecurityGroupRequest{
		Name:                  expandOrGenerateString(d.Get("name"), "sg"),
		Zone:                  zone,
		Project:               expandStringPtr(d.Get("project_id")),
		Description:           d.Get("description").(string),
		Stateful:              d.Get("stateful").(bool),
		EnableDefaultSecurity: expandBoolPtr(d.Get("enable_default_security")),
	}
	tags := expandStrings(d.Get("tags"))
	if len(tags) > 0 {
		req.Tags = tags
	}

	inboundDefaultPolicy := instance.SecurityGroupPolicy("")
	if d.Get("inbound_default_policy") != nil {
		inboundDefaultPolicy = instance.SecurityGroupPolicy(d.Get("inbound_default_policy").(string))
	}
	outboundDefaultPolicy := instance.SecurityGroupPolicy("")
	if d.Get("outbound_default_policy") != nil {
		outboundDefaultPolicy = instance.SecurityGroupPolicy(d.Get("outbound_default_policy").(string))
	}

	req.InboundDefaultPolicy = inboundDefaultPolicy
	req.OutboundDefaultPolicy = outboundDefaultPolicy

	res, err := instanceAPI.CreateSecurityGroup(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newZonedIDString(zone, res.SecurityGroup.ID))

	return resourceScalewayInstanceSecurityGroupRead(ctx, d, meta)
}

func resourceScalewayInstanceSecurityGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceAPI, zone, ID, err := instanceAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := instanceAPI.GetSecurityGroup(&instance.GetSecurityGroupRequest{
		SecurityGroupID: ID,
		Zone:            zone,
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("zone", zone)
	_ = d.Set("organization_id", res.SecurityGroup.Organization)
	_ = d.Set("project_id", res.SecurityGroup.Project)
	_ = d.Set("name", res.SecurityGroup.Name)
	_ = d.Set("stateful", res.SecurityGroup.Stateful)
	_ = d.Set("description", res.SecurityGroup.Description)
	_ = d.Set("inbound_default_policy", res.SecurityGroup.InboundDefaultPolicy.String())
	_ = d.Set("outbound_default_policy", res.SecurityGroup.OutboundDefaultPolicy.String())
	_ = d.Set("enable_default_security", res.SecurityGroup.EnableDefaultSecurity)
	_ = d.Set("tags", res.SecurityGroup.Tags)
	return nil
}

func resourceScalewayInstanceSecurityGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceAPI, _, err := instanceAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	zone, ID, err := parseZonedID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	updateReq := &instance.UpdateSecurityGroupRequest{
		Zone:            zone,
		SecurityGroupID: ID,
	}

	if d.HasChange("description") {
		updateReq.Description = expandStringPtr(d.Get("description").(string))
	}

	if d.HasChange("stateful") {
		updateReq.Stateful = scw.BoolPtr(d.Get("stateful").(bool))
	}

	if d.HasChange("inbound_default_policy") {
		inboundDefaultPolicy := instance.SecurityGroupPolicy(d.Get("inbound_default_policy").(string))
		updateReq.InboundDefaultPolicy = &inboundDefaultPolicy
	}
	if d.HasChange("outbound_default_policy") {
		outboundDefaultPolicy := instance.SecurityGroupPolicy(d.Get("outbound_default_policy").(string))
		updateReq.OutboundDefaultPolicy = &outboundDefaultPolicy
	}

	if d.HasChange("tags") {
		updateReq.Tags = expandUpdatedStringsPtr(d.Get("tags"))
	}

	if d.HasChange("enable_default_security") {
		updateReq.EnableDefaultSecurity = expandBoolPtr(d.Get("enable_default_security"))
	}

	// Only update name if one is provided in the state
	if d.HasChange("name") && d.Get("name") != nil && d.Get("name").(string) != "" {
		updateReq.Name = expandStringPtr(d.Get("name"))
	}

	_, err = instanceAPI.UpdateSecurityGroup(updateReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayInstanceSecurityGroupRead(ctx, d, meta)
}

func resourceScalewayInstanceSecurityGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceAPI, _, err := instanceAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	zone, ID, err := parseZonedID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = instanceAPI.DeleteSecurityGroup(&instance.DeleteSecurityGroupRequest{
		SecurityGroupID: ID,
		Zone:            zone,
	}, scw.WithContext(ctx))

	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
