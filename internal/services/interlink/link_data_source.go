package interlink

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	interlink "github.com/scaleway/scaleway-sdk-go/api/interlink/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/link_data_source.md
var linkDataSourceDescription string

func DataSourceLink() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceLink().SchemaFunc())

	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "region", "project_id")

	dsSchema["name"].ConflictsWith = []string{"link_id"}
	dsSchema["link_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the link",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith:    []string{"name"},
	}

	return &schema.Resource{
		Schema:      dsSchema,
		ReadContext: DataSourceLinkRead,
		Description: linkDataSourceDescription,
	}
}

func DataSourceLinkRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := interlinkAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	linkID, ok := d.GetOk("link_id")
	if !ok {
		linkName := d.Get("name").(string)

		res, err := api.ListLinks(&interlink.ListLinksRequest{
			Region:    region,
			Name:      types.ExpandStringPtr(linkName),
			ProjectID: types.ExpandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundLink, err := datasource.FindExact(
			res.Links,
			func(s *interlink.Link) bool { return s.Name == linkName },
			linkName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		linkID = foundLink.ID
	}

	regionalID := datasource.NewRegionalID(linkID, region)
	d.SetId(regionalID)
	_ = d.Set("link_id", regionalID)

	link, err := api.GetLink(&interlink.GetLinkRequest{
		LinkID: locality.ExpandID(linkID),
		Region: region,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return setLinkState(d, link)
}
