package cockpit

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

// SourceSchema is the SDKv2 attribute map still shared by cockpit source data sources.
func SourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "Name of the datasource",
		},
		"type": {
			Type:             schema.TypeString,
			Optional:         true,
			ForceNew:         true,
			Description:      "The type of the datasource",
			ValidateDiagFunc: verify.ValidateEnum[cockpit.DataSourceType](),
		},
		"retention_days": {
			Type:         schema.TypeInt,
			Required:     true,
			ValidateFunc: validation.IntBetween(1, 1825),
			Description:  "The number of days to retain data. Use scaleway_cockpit_config data source to read allowed min, max, and default values for each data source type.",
		},
		"url": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The URL of the datasource",
		},
		"origin": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The origin of the datasource",
		},
		"synchronized_with_grafana": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Indicates whether the data source is synchronized with Grafana",
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the creation of the cockpit datasource",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the last update of the cockpit datasource",
		},
		"push_url": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The URL endpoint used for pushing data to the cockpit data source.",
		},
		"srn": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The Scaleway Resource Name (SRN) of the data source",
		},
		"project_id": account.ProjectIDSchema(),
		"region":     regional.Schema(),
	}
}
