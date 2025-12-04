package applesilicon

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	applesilicon "github.com/scaleway/scaleway-sdk-go/api/applesilicon/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceRunner() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceAppleSiliconRunnerCreate,
		ReadContext:   ResourceAppleSiliconRunnerRead,
		UpdateContext: ResourceAppleSiliconRunnerUpdate,
		DeleteContext: ResourceAppleSiliconRunnerDelete,
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(5 * time.Minute),
			Default: schema.DefaultTimeout(5 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the runner",
				Computed:    true,
				Optional:    true,
			},
			"ci_provider": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The CI/CD provider for the runner. Must be either 'github' or 'gitlab'",
				ValidateDiagFunc: verify.ValidateEnum[applesilicon.RunnerConfigurationProvider](),
			},
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The URL of the runner to run",
			},
			"token": {
				Type:        schema.TypeString,
				Sensitive:   true,
				Required:    true,
				Description: "The token used to authenticate the runner to run",
			},
			"labels": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Computed:    true,
				Description: "A list of labels that should be applied to the runner.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the runner",
			},
			"zone":       zonal.Schema(),
			"project_id": account.ProjectIDSchema(),
		},
	}
}

func ResourceAppleSiliconRunnerCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	asAPI, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	provider := d.Get("ci_provider").(string)

	runnerConfig := &applesilicon.RunnerConfigurationV2{
		Name:                d.Get("name").(string),
		Provider:            applesilicon.RunnerConfigurationV2Provider(provider),
		GithubConfiguration: nil,
		GitlabConfiguration: nil,
	}

	if provider == "github" {
		runnerConfig.GithubConfiguration = &applesilicon.GithubRunnerConfiguration{
			URL:    d.Get("url").(string),
			Token:  d.Get("token").(string),
			Labels: types.ExpandStrings(d.Get("labels")),
		}
	}

	if provider == "gitlab" {
		runnerConfig.GitlabConfiguration = &applesilicon.GitlabRunnerConfiguration{
			URL:   d.Get("url").(string),
			Token: d.Get("token").(string),
		}
	}

	createRunnerReq := &applesilicon.CreateRunnerRequest{
		Zone:                zone,
		ProjectID:           d.Get("project_id").(string),
		RunnerConfiguration: runnerConfig,
	}

	runner, err := asAPI.CreateRunner(createRunnerReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(zonal.NewIDString(zone, runner.ID))

	return ResourceAppleSiliconRunnerRead(ctx, d, m)
}

func ResourceAppleSiliconRunnerRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	asAPI, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	runner, err := asAPI.GetRunner(&applesilicon.GetRunnerRequest{
		Zone:     zone,
		RunnerID: ID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("name", runner.ID)
	_ = d.Set("name", runner.Configuration.Name)
	_ = d.Set("ci_provider", runner.Configuration.Provider)
	_ = d.Set("status", runner.Status)

	if runner.Configuration.Provider == "github" {
		_ = d.Set("token", runner.Configuration.GithubConfiguration.Token)
		_ = d.Set("url", runner.Configuration.GithubConfiguration.URL)
		_ = d.Set("labels", runner.Configuration.GithubConfiguration.Labels)
	}

	return nil
}

func ResourceAppleSiliconRunnerUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	asAPI, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	provider := d.Get("ci_provider").(string)

	runnerConfig := &applesilicon.RunnerConfigurationV2{
		Name:                d.Get("name").(string),
		Provider:            applesilicon.RunnerConfigurationV2Provider(provider),
		GithubConfiguration: nil,
		GitlabConfiguration: nil,
	}

	if provider == "github" {
		runnerConfig.GithubConfiguration = &applesilicon.GithubRunnerConfiguration{
			URL:    d.Get("url").(string),
			Token:  d.Get("token").(string),
			Labels: types.ExpandStrings(d.Get("labels")),
		}
	}

	if provider == "gitlab" {
		runnerConfig.GitlabConfiguration = &applesilicon.GitlabRunnerConfiguration{
			URL:   d.Get("url").(string),
			Token: d.Get("token").(string),
		}
	}

	updateRunnerReq := &applesilicon.UpdateRunnerRequest{
		Zone:                zone,
		RunnerID:            ID,
		RunnerConfiguration: runnerConfig,
	}

	_, err = asAPI.UpdateRunner(updateRunnerReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceAppleSiliconRunnerRead(ctx, d, m)
}

func ResourceAppleSiliconRunnerDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	asAPI, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	runnerDeleteReq := &applesilicon.DeleteRunnerRequest{
		Zone:     zone,
		RunnerID: ID,
	}
	err = asAPI.DeleteRunner(runnerDeleteReq, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}
	return nil
}
