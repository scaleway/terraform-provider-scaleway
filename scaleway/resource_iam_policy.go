package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
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
			"organization_id": organizationIDOptionalSchema(),
			"user_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "User id",
				ValidateFunc: verify.IsUUID(),
				ExactlyOneOf: []string{"group_id", "application_id", "no_principal"},
			},
			"group_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Group id",
				ValidateFunc: verify.IsUUID(),
				ExactlyOneOf: []string{"user_id", "application_id", "no_principal"},
			},
			"application_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Application id",
				ValidateFunc: verify.IsUUID(),
				ExactlyOneOf: []string{"user_id", "group_id", "no_principal"},
			},
			"no_principal": {
				Type:         schema.TypeBool,
				Optional:     true,
				Description:  "Deactivate policy to a principal",
				ExactlyOneOf: []string{"user_id", "group_id", "application_id"},
			},
			"rule": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Rules of the policy to create",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"organization_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "ID of organization scoped to the rule. Only one of project_ids and organization_id may be set.",
							ValidateFunc: verify.IsUUID(),
						},
						"project_ids": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "List of project IDs scoped to the rule. Only one of project_ids and organization_id may be set.",
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: verify.IsUUID(),
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
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "The tags associated with the policy",
			},
		},
	}
}

func resourceScalewayIamPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := iamAPI(m)

	pol, err := api.CreatePolicy(&iam.CreatePolicyRequest{
		Name:           types.ExpandOrGenerateString(d.Get("name"), "policy"),
		Description:    d.Get("description").(string),
		Rules:          expandPolicyRuleSpecs(d.Get("rule")),
		UserID:         types.ExpandStringPtr(d.Get("user_id")),
		GroupID:        types.ExpandStringPtr(d.Get("group_id")),
		ApplicationID:  types.ExpandStringPtr(d.Get("application_id")),
		NoPrincipal:    types.ExpandBoolPtr(getBool(d, "no_principal")),
		OrganizationID: d.Get("organization_id").(string),
		Tags:           types.ExpandStrings(d.Get("tags")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(pol.ID)

	return resourceScalewayIamPolicyRead(ctx, d, m)
}

func resourceScalewayIamPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := iamAPI(m)
	pol, err := api.GetPolicy(&iam.GetPolicyRequest{
		PolicyID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	_ = d.Set("name", pol.Name)
	_ = d.Set("description", pol.Description)
	_ = d.Set("created_at", types.FlattenTime(pol.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(pol.UpdatedAt))
	_ = d.Set("organization_id", pol.OrganizationID)
	_ = d.Set("editable", pol.Editable)
	_ = d.Set("tags", types.FlattenSliceString(pol.Tags))

	if pol.UserID != nil {
		_ = d.Set("user_id", types.FlattenStringPtr(pol.UserID))
	}
	if pol.GroupID != nil {
		_ = d.Set("group_id", types.FlattenStringPtr(pol.GroupID))
	}
	if pol.ApplicationID != nil {
		_ = d.Set("application_id", types.FlattenStringPtr(pol.ApplicationID))
	}

	_ = d.Set("no_principal", types.FlattenBoolPtr(pol.NoPrincipal))

	listRules, err := api.ListRules(&iam.ListRulesRequest{
		PolicyID: pol.ID,
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to list policy's rules: %w", err))
	}

	_ = d.Set("rule", flattenPolicyRules(listRules.Rules))

	return nil
}

func resourceScalewayIamPolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := iamAPI(m)

	req := &iam.UpdatePolicyRequest{
		PolicyID: d.Id(),
	}

	hasUpdated := false

	if d.HasChange("name") {
		hasUpdated = true
		req.Name = types.ExpandStringPtr(d.Get("name"))
	}
	if d.HasChange("description") {
		hasUpdated = true
		req.Description = types.ExpandUpdatedStringPtr(d.Get("description"))
	}
	if d.HasChange("tags") {
		hasUpdated = true
		req.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
	}
	if d.HasChange("user_id") {
		hasUpdated = true
		req.UserID = types.ExpandStringPtr(d.Get("user_id"))
	}
	if d.HasChange("group_id") {
		hasUpdated = true
		req.GroupID = types.ExpandStringPtr(d.Get("group_id"))
	}
	if d.HasChange("application_id") {
		hasUpdated = true
		req.ApplicationID = types.ExpandStringPtr(d.Get("application_id"))
	}
	if noPrincipal := d.Get("no_principal"); d.HasChange("no_principal") && noPrincipal.(bool) {
		hasUpdated = true
		req.NoPrincipal = types.ExpandBoolPtr(noPrincipal)
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

	return resourceScalewayIamPolicyRead(ctx, d, m)
}

func resourceScalewayIamPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := iamAPI(m)

	err := api.DeletePolicy(&iam.DeletePolicyRequest{
		PolicyID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}
