package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayIamPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayIamPolicyCreate,
		ReadContext:   resourceScalewayIamPolicyRead,
		UpdateContext: resourceScalewayIamPolicyUpdate,
		DeleteContext: resourceScalewayIamPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The name of the iam policy",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the iam policy",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the policy",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last update of the policy",
			},
			"editable": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether or not the policy is editable.",
			},
			"organization_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "ID of organization the policy is linked to.",
			},
			"user_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "User id",
				ExactlyOneOf: []string{"group_id", "application_id", "no_principal"},
			},
			"group_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Group id",
				ExactlyOneOf: []string{"user_id", "application_id", "no_principal"},
			},
			"application_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Application id",
				ExactlyOneOf: []string{"user_id", "group_id", "no_principal"},
			},
			"no_principal": {
				Type:         schema.TypeBool,
				Optional:     true,
				Description:  "Deactivate policy to a principal",
				ExactlyOneOf: []string{"user_id", "group_id", "application_id"},
				ForceNew:     true,
			},
			"rule": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Rules of the policy to create",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"organization_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "ID of organization scoped to the rule. Only one of project_ids and organization_id may be set.",
						},
						"project_ids": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "List of project IDs scoped to the rule. Only one of project_ids and organization_id may be set.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"permission_set_names": {
							Type:        schema.TypeSet,
							Required:    true,
							Description: "Names of permission sets bound to the rule.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func resourceScalewayIamPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := iamAPI(meta)

	pol, err := api.CreatePolicy(&iam.CreatePolicyRequest{
		Name:           expandOrGenerateString(d.Get("name"), "policy-"),
		Description:    d.Get("description").(string),
		Rules:          expandPolicyRuleSpecs(d.Get("rule")),
		UserID:         expandStringPtr(d.Get("user_id")),
		GroupID:        expandStringPtr(d.Get("group_id")),
		ApplicationID:  expandStringPtr(d.Get("application_id")),
		NoPrincipal:    expandBoolPtr(getBool(d, "no_principal")),
		OrganizationID: d.Get("organization_id").(string),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(pol.ID)

	return resourceScalewayIamPolicyRead(ctx, d, meta)
}

func resourceScalewayIamPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := iamAPI(meta)
	pol, err := api.GetPolicy(&iam.GetPolicyRequest{
		PolicyID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	_ = d.Set("name", pol.Name)
	_ = d.Set("description", pol.Description)
	_ = d.Set("created_at", flattenTime(pol.CreatedAt))
	_ = d.Set("updated_at", flattenTime(pol.UpdatedAt))
	_ = d.Set("organization_id", pol.OrganizationID)
	_ = d.Set("editable", pol.Editable)

	if pol.UserID != nil {
		_ = d.Set("user_id", flattenStringPtr(pol.UserID))
	}
	if pol.GroupID != nil {
		_ = d.Set("group_id", flattenStringPtr(pol.GroupID))
	}
	if pol.ApplicationID != nil {
		_ = d.Set("application_id", flattenStringPtr(pol.ApplicationID))
	}
	_ = d.Set("no_principal", flattenBoolPtr(pol.NoPrincipal))

	listRules, err := api.ListRules(&iam.ListRulesRequest{
		PolicyID: &pol.ID,
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to list policy's rules: %w", err))
	}

	_ = d.Set("rule", flattenPolicyRules(listRules.Rules))

	return nil
}

func resourceScalewayIamPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := iamAPI(meta)

	req := &iam.UpdatePolicyRequest{
		PolicyID: d.Id(),
	}

	hasUpdated := false

	if d.HasChange("name") {
		hasUpdated = true
		req.Name = expandStringPtr(d.Get("name"))
	}
	if d.HasChange("description") {
		hasUpdated = true
		req.Description = expandUpdatedStringPtr(d.Get("description"))
	}
	if d.HasChange("user_id") {
		hasUpdated = true
		req.UserID = expandStringPtr(d.Get("user_id"))
	}
	if d.HasChange("group_id") {
		hasUpdated = true
		req.GroupID = expandStringPtr(d.Get("group_id"))
	}
	if d.HasChange("application_id") {
		hasUpdated = true
		req.ApplicationID = expandStringPtr(d.Get("application_id"))
	}
	if d.HasChange("no_principal") {
		hasUpdated = true
		req.NoPrincipal = expandBoolPtr(getBool(d, "no_principal"))
	}
	if hasUpdated {
		_, err := api.UpdatePolicy(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("rule") {
		_, err := api.SetRules(&iam.SetRulesRequest{
			PolicyID: d.Id(),
			Rules:    expandPolicyRuleSpecs(d.Get("rule")),
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceScalewayIamPolicyRead(ctx, d, meta)
}

func resourceScalewayIamPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := iamAPI(meta)

	err := api.DeletePolicy(&iam.DeletePolicyRequest{
		PolicyID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}
