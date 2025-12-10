package instance

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	product_catalog "github.com/scaleway/scaleway-sdk-go/api/product_catalog/v2alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

func DataSourceServerType() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourceInstanceServerTypeRead,
		SchemaFunc:  serverTypeSchema,
	}
}

func serverTypeSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the server type",
		},
		"arch": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The architecture of the server type",
		},
		"cpu": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The number of CPU cores of the server type",
		},
		"ram": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The number of bytes of RAM of the server type",
		},
		"gpu": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The number of GPUs of the server type",
		},
		"volumes": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "The specifications of volumes allowed for the server type",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"min_size_total": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "The minimum total size in bytes of volumes allowed on the server type",
					},
					"max_size_total": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "The maximum total size in bytes of volumes allowed on the server type",
					},
					"min_size_per_local_volume": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "The minimum size in bytes per local volume allowed on the server type",
					},
					"max_size_per_local_volume": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "The maximum size in bytes per local volume allowed on the server type",
					},
					"scratch_storage_max_size": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "The maximum size in bytes of the scratch volume allowed on the server type",
					},
					"block_storage": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: "Whether block storage is allowed on the server type",
					},
				},
			},
		},
		"capabilities": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "The specific capabilities of the server type",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"boot_types": {
						Type:        schema.TypeList,
						Computed:    true,
						Description: "The list of boot types allowed for the server type",
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"max_file_systems": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "The maximum number of file systems that can be attached to the server type",
					},
				},
			},
		},
		"network": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "The network specifications of the server type",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"internal_bandwidth": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "The internal bandwidth of the server type",
					},
					"public_bandwidth": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "The public bandwidth of the server type",
					},
					"block_bandwidth": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "The block bandwidth of the server type",
					},
				},
			},
		},
		"hourly_price": {
			Type:        schema.TypeFloat,
			Computed:    true,
			Description: "The hourly price of the server type in euro",
		},
		"end_of_service": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Whether the server type will soon reach End Of Service",
		},
		"availability": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Whether the server type is available in the zone",
		},
		"zone": zonal.Schema(),
	}
}

func DataSourceInstanceServerTypeRead(ctx context.Context, d *schema.ResourceData, i any) diag.Diagnostics {
	instanceAPI, zone, err := newAPIWithZone(d, i)
	if err != nil {
		return diag.FromErr(err)
	}

	serverTypes, err := instanceAPI.ListServersTypes(&instance.ListServersTypesRequest{
		Zone: zone,
	}, scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	name := d.Get("name").(string)

	serverType, ok := serverTypes.Servers[name]
	if !ok {
		return diag.Errorf("server type %s not found", name)
	}

	d.SetId(name)
	_ = d.Set("name", name)
	_ = d.Set("zone", zone)
	_ = d.Set("arch", serverType.Arch)
	_ = d.Set("cpu", int(serverType.Ncpus))
	_ = d.Set("ram", int(serverType.RAM))
	_ = d.Set("volumes", flattenVolumeConstraints(serverType))
	_ = d.Set("capabilities", flattenCapabilities(serverType.Capabilities))
	_ = d.Set("network", flattenNetwork(serverType))
	_ = d.Set("end_of_service", serverType.EndOfService)

	if serverType.Gpu != nil {
		_ = d.Set("gpu", int(*serverType.Gpu))
	}

	// Price (needs to be fetched from the Product Catalog)
	pcuAPI := product_catalog.NewPublicCatalogAPI(meta.ExtractScwClient(i))

	pcuInstances, err := pcuAPI.ListPublicCatalogProducts(&product_catalog.PublicCatalogAPIListPublicCatalogProductsRequest{
		ProductTypes: []product_catalog.ListPublicCatalogProductsRequestProductType{
			product_catalog.ListPublicCatalogProductsRequestProductTypeInstance,
		},
		Zone: &zone,
	}, scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	for _, pcuInstance := range pcuInstances.Products {
		if pcuInstance.Properties.Instance.OfferID != name {
			continue
		}

		_ = d.Set("hourly_price", pcuInstance.Price.RetailPrice.ToFloat())
	}

	// Availability
	availabilitiesResponse, err := instanceAPI.GetServerTypesAvailability(&instance.GetServerTypesAvailabilityRequest{
		Zone: zone,
	}, scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if availability, exists := availabilitiesResponse.Servers[name]; exists {
		_ = d.Set("availability", availability.Availability.String())
	}

	return nil
}

func flattenVolumeConstraints(serverType *instance.ServerType) []map[string]any {
	flattened := map[string]any{}

	if serverType.VolumesConstraint != nil {
		flattened["min_size_total"] = serverType.VolumesConstraint.MinSize
		flattened["max_size_total"] = serverType.VolumesConstraint.MaxSize
	}

	if serverType.PerVolumeConstraint != nil && serverType.PerVolumeConstraint.LSSD != nil {
		flattened["min_size_per_local_volume"] = serverType.PerVolumeConstraint.LSSD.MinSize
		flattened["max_size_per_local_volume"] = serverType.PerVolumeConstraint.LSSD.MaxSize
	}

	if serverType.ScratchStorageMaxSize != nil {
		flattened["scratch_storage_max_size"] = serverType.ScratchStorageMaxSize
	}

	if serverType.Capabilities != nil && serverType.Capabilities.BlockStorage != nil {
		flattened["block_storage"] = *serverType.Capabilities.BlockStorage
	}

	return []map[string]any{flattened}
}

func flattenCapabilities(capabilities *instance.ServerTypeCapabilities) []map[string]any {
	if capabilities == nil {
		return nil
	}

	bootTypes := []string(nil)
	for _, bootType := range capabilities.BootTypes {
		bootTypes = append(bootTypes, bootType.String())
	}

	flattened := map[string]any{
		"max_file_systems": capabilities.MaxFileSystems,
		"boot_types":       bootTypes,
	}

	return []map[string]any{flattened}
}

func flattenNetwork(serverType *instance.ServerType) []map[string]any {
	if serverType.Network == nil {
		return nil
	}

	flattened := map[string]any{}

	if serverType.Network.SumInternalBandwidth != nil {
		flattened["internal_bandwidth"] = *serverType.Network.SumInternalBandwidth
	}

	if serverType.Network.SumInternetBandwidth != nil {
		flattened["public_bandwidth"] = *serverType.Network.SumInternetBandwidth
	}

	if serverType.BlockBandwidth != nil {
		flattened["block_bandwidth"] = *serverType.BlockBandwidth
	}

	return []map[string]any{flattened}
}
