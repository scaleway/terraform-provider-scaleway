package messageq

import (
	"context"
	"fmt"
	"io"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	messageqapi "github.com/scaleway/scaleway-sdk-go/api/messageq/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceCertificateAuthority() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourceCertificateAuthorityRead,
		Schema: map[string]*schema.Schema{
			"deployment_id": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
				Description:      "The ID of the MessageQ deployment",
			},
			"region": regional.Schema(),
			"pem": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "PEM-encoded certificate authority content",
			},
		},
	}
}

func DataSourceCertificateAuthorityRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	region, deploymentID := regionAndIDFromAttr(d.Get("deployment_id").(string), region)

	file, err := api.DownloadDeploymentCertificateAuthority(&messageqapi.DownloadDeploymentCertificateAuthorityRequest{
		Region:       region,
		DeploymentID: deploymentID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if file == nil || file.Content == nil {
		return diag.FromErr(fmt.Errorf("certificate authority content is empty for deployment %s", deploymentID))
	}

	content, err := io.ReadAll(file.Content)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to read certificate authority content: %w", err))
	}

	d.SetId(regional.NewIDString(region, deploymentID))
	_ = d.Set("region", region.String())
	_ = d.Set("deployment_id", regional.NewIDString(region, deploymentID))
	_ = d.Set("pem", string(content))

	return nil
}
