package audittrail

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	audittrailSDK "github.com/scaleway/scaleway-sdk-go/api/audit_trail/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceEvent() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourceEventsRead,
		Schema: map[string]*schema.Schema{
			"organization_id": {
				Type:             schema.TypeString,
				Description:      "ID of the organization containing the Audit Trail events.",
				Optional:         true,
				ValidateDiagFunc: verify.IsUUID(),
			},
			"region": regional.Schema(),
			"project_id": {
				Type:             schema.TypeString,
				Description:      "ID of the project containing the Audit Trail events.",
				Optional:         true,
				ValidateDiagFunc: verify.IsUUID(),
			},
			"resource_type": {
				Type:        schema.TypeString,
				Description: "Type of the scaleway resources associated with the listed events",
				Optional:    true,
				ValidateDiagFunc: func(i any, p cty.Path) diag.Diagnostics {
					resourceTypeValues := audittrailSDK.ResourceType("").Values()

					resourceTypeStringValues := make([]string, 0, len(resourceTypeValues))
					for _, resourceTypeValue := range resourceTypeValues {
						resourceTypeStringValues = append(resourceTypeStringValues, resourceTypeValue.String())
					}

					return verify.ValidateStringInSliceWithWarning(resourceTypeStringValues, "resourceType")(i, p)
				},
			},
			"resource_id": {
				Type:             schema.TypeString,
				Description:      "ID of the Scaleway resource associated with the listed events",
				Optional:         true,
				ValidateDiagFunc: verify.IsUUID(),
			},
			"product_name": {
				Type:        schema.TypeString,
				Description: "Scaleway product associated with the listed events in a hyphenated format",
				Optional:    true,
			},
			"events": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of Audit Trail events",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Description: "ID of the event",
							Computed:    true,
						},
						"recorded_at": {
							Type:        schema.TypeString,
							Description: "Timestamp of the event",
							Computed:    true,
						},
						"locality": {
							Type:        schema.TypeString,
							Description: "Locality of the resource attached to the event",
							Computed:    true,
						},
						"principal_id": {
							Type:        schema.TypeString,
							Description: "ID of the user or IAM application at the origin of the event",
							Computed:    true,
						},
						"organization_id": {
							Type:        schema.TypeString,
							Description: "Organization of the resource attached to the event",
							Computed:    true,
						},
						"project_id": {
							Type:        schema.TypeString,
							Description: "Project of the resource attached to the event",
							Computed:    true,
						},
						"source_ip": {
							Type:        schema.TypeString,
							Description: "IP address at the origin of the event",
							Computed:    true,
						},
						"user_agent": {
							Type:        schema.TypeString,
							Description: "User Agent at the origin of the event",
							Computed:    true,
						},
						"product_name": {
							Type:        schema.TypeString,
							Description: "Product name of the resource attached to the event",
							Computed:    true,
						},
						"service_name": {
							Type:        schema.TypeString,
							Description: "API name called to trigger the event",
							Computed:    true,
						},
						"method_name": {
							Type:        schema.TypeString,
							Description: "API method called to trigger the event",
							Computed:    true,
						},
						"resources": {
							Type:        schema.TypeList,
							Description: "List of resources attached to the event",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "ID of the resource attached to the event",
									},
									"type": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Type of the Scaleway resource",
									},
									"name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Name of the Scaleway resource",
									},
								},
							},
						},
						"request_id": {
							Type:        schema.TypeString,
							Description: "Unique identifier of the request at the origin of the event",
							Computed:    true,
						},
						"request_body": {
							Type:        schema.TypeString,
							Description: "Request at the origin of the event",
							Computed:    true,
						},
						"status_code": {
							Type:        schema.TypeString,
							Description: "HTTP status code resulting of the API call",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func DataSourceEventsRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	auditTrailAPI, region, orgID, err := newAPIWithRegionAndOrgID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := audittrailSDK.ListEventsRequest{
		OrganizationID: orgID,
		Region:         region,
	}

	if projectID, ok := d.GetOk("project_id"); ok {
		req.ProjectID = types.ExpandStringPtr(projectID)
	}

	if resourceType, ok := d.GetOk("resource_type"); ok {
		req.ResourceType = audittrailSDK.ResourceType(resourceType.(string))
	}

	if productName, ok := d.GetOk("product_name"); ok {
		req.ProductName = types.ExpandStringPtr(productName)
	}

	if resourceID, ok := d.GetOk("resource_id"); ok {
		req.ResourceID = types.ExpandStringPtr(resourceID)
	}

	res, err := auditTrailAPI.ListEvents(&req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uuid.New().String())
	_ = d.Set("organization_id", orgID)
	_ = d.Set("region", region)

	flattenedEvents, err := flattenEvents(res.Events)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("events", flattenedEvents)

	return nil
}

func flattenEvents(events []*audittrailSDK.Event) ([]map[string]any, error) {
	flattenedEvents := make([]map[string]any, len(events))
	for i, event := range events {
		var principalID string

		if event.Principal != nil {
			principalID = event.Principal.ID
		}

		requestBody, err := scw.EncodeJSONObject(*event.RequestBody, scw.NoEscape)
		if err != nil {
			return nil, err
		}

		flattenedEvents[i] = map[string]any{
			"id":              event.ID,
			"recorded_at":     event.RecordedAt.String(),
			"locality":        event.Locality,
			"principal_id":    principalID,
			"organization_id": event.OrganizationID,
			"project_id":      event.ProjectID,
			"source_ip":       event.SourceIP.String(),
			"user_agent":      event.UserAgent,
			"product_name":    event.ProductName,
			"service_name":    event.ServiceName,
			"method_name":     event.MethodName,
			"resources":       flattenResources(event.Resources),
			"request_id":      event.RequestID,
			"request_body":    requestBody,
			"status_code":     strconv.FormatUint(uint64(event.StatusCode), 10),
		}
	}

	return flattenedEvents, nil
}

func flattenResources(resources []*audittrailSDK.Resource) []map[string]any {
	flattenedResources := make([]map[string]any, len(resources))
	for i, r := range resources {
		flattenedResources[i] = map[string]any{
			"id":   r.ID,
			"type": string(r.Type),
			"name": r.Name,
		}
	}

	return flattenedResources
}
