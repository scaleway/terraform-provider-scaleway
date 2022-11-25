package scaleway

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"time"
)

const (
	defaultMNQNamespaceTimeout = 5 * time.Minute
)

func resourceScalewayMNQNamespace() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayMNQNamespaceCreate,
		ReadContext:   resourceScalewayMNQNamespaceRead,
		UpdateContext: resourceScalewayMNQNamespaceUpdate,
		DeleteContext: resourceScalewayMNQNamespaceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultMNQNamespaceTimeout),
			Read:    schema.DefaultTimeout(defaultMNQNamespaceTimeout),
			Update:  schema.DefaultTimeout(defaultMNQNamespaceTimeout),
			Delete:  schema.DefaultTimeout(defaultMNQNamespaceTimeout),
			Default: schema.DefaultTimeout(defaultMNQNamespaceTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The name of the mnq namespace",
			},
			"protocol": {
				Type:        schema.TypeString,
				Required:    true,
				Optional:    true,
				Description: "The Namespace protocol",
				ForceNew:    true,
				ValidateFunc: func(i interface{}, s string) (strings []string, errors []error) {
					str, isStr := i.(string)
					if !isStr {
						return nil, []error{fmt.Errorf("%v is not a string", i)}
					}
					_, err := mnq.NamespaceProtocol(str).MarshalJSON()
					if err != nil {
						return nil, []error{fmt.Errorf("is not a supported namespace protocol %s", str)}
					}
					return nil, nil
				},
			},
			// computed
			"region":     regionSchema(),
			"project_id": projectIDSchema(),
			"endpoint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The endpoint of the service matching the Namespace protocol",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the mnq Namespace",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last update of the mnq Namespace",
			},
		},
	}
}
func resourceScalewayMNQNamespaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := newMNQAPI(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	request := &mnq.CreateNamespaceRequest{
		Name:      expandOrGenerateString(d.Get("name").(string), "ns"),
		ProjectID: d.Get("project_id").(string),
		Region:    region,
		Protocol:  mnq.NamespaceProtocol(d.Get("protocol").(string)),
	}
	if regionRaw, ok := d.GetOk("region"); ok {
		request.Region = scw.Region(regionRaw.(string))
	}
	namespace, err := api.CreateNamespace(request, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newRegionalID(namespace.Region, namespace.ID).String())

	return resourceScalewayMNQNamespaceRead(ctx, d, meta)
}

func resourceScalewayMNQNamespaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := mnqAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	request := &mnq.GetNamespaceRequest{
		NamespaceID: id,
		Region:      region,
	}

	namespace, err := api.GetNamespace(request, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("name", namespace.Name)
	_ = d.Set("project_id", namespace.ProjectID)
	_ = d.Set("endpoint", namespace.Endpoint)
	_ = d.Set("protocol", namespace.Protocol.String())
	_ = d.Set("created_at", flattenTime(namespace.CreatedAt))
	_ = d.Set("updated_at", flattenTime(namespace.UpdatedAt))

	return nil
}

func resourceScalewayMNQNamespaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := mnqAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	request := &mnq.UpdateNamespaceRequest{
		NamespaceID: id,
		Region:      region,
	}

	if d.HasChange("name") {
		request.Name = scw.StringPtr(d.Get("name").(string))
	}

	_, err = api.UpdateNamespace(request, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayMNQNamespaceRead(ctx, d, meta)
}

func resourceScalewayMNQNamespaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := mnqAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	request := &mnq.DeleteNamespaceRequest{
		NamespaceID: id,
		Region:      region,
	}
	err = api.DeleteNamespace(request, scw.WithContext(ctx))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
