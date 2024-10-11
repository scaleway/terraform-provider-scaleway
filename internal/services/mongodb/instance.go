package mongodb

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	mongodb "github.com/scaleway/scaleway-sdk-go/api/mongodb/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/cdf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceInstanceCreate,
		ReadContext:   ResourceInstanceRead,
		UpdateContext: ResourceInstanceUpdate,
		DeleteContext: ResourceInstanceDelete,
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultMongodbInstanceTimeout),
			Update:  schema.DefaultTimeout(defaultMongodbInstanceTimeout),
			Delete:  schema.DefaultTimeout(defaultMongodbInstanceTimeout),
			Default: schema.DefaultTimeout(defaultMongodbInstanceTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of the mongoDB cluster",
			},
			"version": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Mongodb version of the instance",
			},
			"node_number": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "number of node in the instance",
			},
			"node_type": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "Type of node to use for the instance",
				DiffSuppressFunc: dsf.IgnoreCase,
			},
			"user_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the user created when the cluster is created",
			},
			"password": {
				Type:        schema.TypeString,
				Sensitive:   true,
				Optional:    true,
				Description: "Password of the user",
			},
			// volume
			"volume_type": {
				Type:        schema.TypeString,
				Default:     mongodb.VolumeTypeSbs5k,
				Optional:    true,
				Description: "Volume size of instance.",
			},
			"volume_size_in_gb": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "Volume size (in GB)",
				ValidateFunc: validation.IntDivisibleBy(5),
			},
			"snapshot_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Snapshot id",
			},
			//endpoint
			"private_network": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Private network specs details",
				Set:         privateNetworkSetHash,
				DiffSuppressFunc: func(k, oldValue, newValue string, _ *schema.ResourceData) bool {
					// Check if the key is for the 'id' attribute
					if strings.HasSuffix(k, "id") {
						return locality.ExpandID(oldValue) == locality.ExpandID(newValue)
					}
					// For all other attributes, don't suppress the diff
					return false
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
							Description:      "UUID of the private network to be connected to the cluster",
						},
						"ips": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.IsCIDR,
							},
							Description: "List of IPv4 addresses of the private network with a CIDR notation",
						},
						"port": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The port of your load balancer service",
						},
						"dns_records": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The dns_record of your endpoint",
						},
						// computed
						"endpoint_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "UUID of the endpoint to be connected to the cluster",
						},
						"zone": zonal.ComputedSchema(),
					},
				},
			},
			// Computed
			"public_network": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Public network specs details",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"port": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "TCP port of the endpoint",
						},
						"dns_record": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The dns_record of your endpoint",
						},
					},
				},
			},
			"tags": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of tags [\"tag1\", \"tag2\", ...] attached to a Mongodb instance",
			},
			"settings": {
				Type:        schema.TypeMap,
				Description: "Map of settings to define for the instance.",
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the Mongodb instance",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last update of the Mongodb instance",
			},
			// Common
			"region":     regional.Schema(),
			"project_id": account.ProjectIDSchema(),
		},
		CustomizeDiff: customdiff.All(
			cdf.LocalityCheck("private_network.#.id"),
			func(ctx context.Context, diff *schema.ResourceDiff, v interface{}) error {
				snapshotID := diff.Get("snapshot_id")

				if snapshotID == nil || snapshotID == "" {
					if diff.Get("user_name") == nil || diff.Get("user_name") == "" {
						return fmt.Errorf("`user_name` must be provided when `snapshot_id` is not set")
					}
					if diff.Get("password") == nil || diff.Get("password") == "" {
						return fmt.Errorf("`password` must be provided when `snapshot_id` is not set")
					}
					if diff.Get("version") == nil || diff.Get("version") == "" {
						return fmt.Errorf("`version` must be provided when `snapshot_id` is not set")
					}
					if diff.Get("node_number") == nil || diff.Get("node_number") == "" {
						return fmt.Errorf("`node_number` must be provided when `snapshot_id` is not set")
					}
				}
				return nil
			},
		),
	}
}

func ResourceInstanceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	mongodbAPI, zone, err := newAPIWithZone(d, m)

	if err != nil {
		return diag.FromErr(err)
	}

	nodeNumber := scw.Uint32Ptr(uint32(d.Get("node_number").(int)))

	snapshotID, exist := d.GetOk("snapshot_id")
	res := &mongodb.Instance{}
	if exist {
		volume := &mongodb.RestoreSnapshotRequestVolumeDetails{
			VolumeType: mongodb.VolumeType(d.Get("volume_type").(string)),
		}
		restoreSnapshotRequest := &mongodb.RestoreSnapshotRequest{
			SnapshotID:   snapshotID.(string),
			InstanceName: types.ExpandOrGenerateString(d.Get("name"), "mongodb"),
			NodeType:     d.Get("node_type").(string),
			Volume:       volume,
		}
		res, err = mongodbAPI.RestoreSnapshot(restoreSnapshotRequest, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	} else {

		createReq := &mongodb.CreateInstanceRequest{
			ProjectID:  d.Get("project_id").(string),
			Name:       types.ExpandOrGenerateString(d.Get("name"), "mongodb"),
			Version:    d.Get("version").(string),
			NodeType:   d.Get("node_type").(string),
			NodeNumber: *nodeNumber,
			UserName:   d.Get("user_name").(string),
			Password:   d.Get("password").(string),
		}

		volumeRequestDetails := &mongodb.CreateInstanceRequestVolumeDetails{
			VolumeType: mongodb.VolumeType(d.Get("volume_type").(string)),
		}
		volumeSize, volumeSizeExist := d.GetOk("volume_size_in_gb")
		if volumeSizeExist {
			volumeRequestDetails.VolumeSize = scw.Size(uint64(volumeSize.(int)) * uint64(scw.GB))
		} else {
			volumeRequestDetails.VolumeSize = scw.Size(defaultVolumeSize * uint64(scw.GB))
		}
		createReq.Volume = volumeRequestDetails

		tags, tagsExist := d.GetOk("tags")
		if tagsExist {
			createReq.Tags = types.ExpandStrings(tags)
		}

		pn, pnExists := d.GetOk("private_network")
		if pnExists {
			pnSpecs, err := expandPrivateNetwork(pn.(*schema.Set).List())
			if err != nil {
				return diag.FromErr(err)
			}
			createReq.Endpoints = pnSpecs
		} else {
			epSpecs := make([]*mongodb.EndpointSpec, 0, 1)
			spec := &mongodb.EndpointSpecPublicDetails{}
			createReq.Endpoints = append(epSpecs, &mongodb.EndpointSpec{Public: spec})
		}

		res, err = mongodbAPI.CreateInstance(createReq, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	d.SetId(zonal.NewIDString(zone, res.ID))
	_, err = waitForInstance(ctx, mongodbAPI, res.Region, res.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceInstanceRead(ctx, d, m)
}

func ResourceInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	mongodbAPI, region, ID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	getReq := &mongodb.GetInstanceRequest{
		Region:     region,
		InstanceID: ID,
	}

	instance, err := mongodbAPI.GetInstance(getReq, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("name", instance.Name)
	_ = d.Set("version", instance.Version)
	_ = d.Set("node_number", instance.NodeNumber)
	_ = d.Set("node_type", instance.NodeType)
	_ = d.Set("project_id", instance.ProjectID)
	_ = d.Set("tags", instance.Tags)
	_ = d.Set("created_at", instance.CreatedAt.Format(time.RFC3339))
	_ = d.Set("region", instance.Region.String())

	if instance.Volume != nil {
		_ = d.Set("volume_type", instance.Volume.Type)
		_ = d.Set("volume_size_in_gb", int(instance.Volume.Size/scw.GB))
	}

	privateNetworkEndpoints, privateNetworkExists := flattenPrivateNetwork(instance.Endpoints)
	if privateNetworkExists {
		_ = d.Set("private_network", privateNetworkEndpoints)
	}
	publicNetworkEndpoint, publicNetworkExists := flattenPublicNetwork(instance.Endpoints)
	if publicNetworkExists {
		_ = d.Set("public_network", publicNetworkEndpoint)
	}

	if len(instance.Settings) > 0 {
		settingsMap := make(map[string]string)
		for _, setting := range instance.Settings {
			settingsMap[setting.Name] = setting.Value
		}
		_ = d.Set("settings", settingsMap)
	}

	return nil
}

func ResourceInstanceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	mongodbAPI, region, ID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	////////////////////
	// Upgrade instance
	////////////////////

	if d.HasChange("volume_size_in_gb") {
		oldSizeInterface, newSizeInterface := d.GetChange("volume_size_in_gb")
		oldSize := uint64(oldSizeInterface.(int))
		newSize := uint64(newSizeInterface.(int))
		if newSize < oldSize {
			return diag.FromErr(errors.New("volume_size_in_gb cannot be decreased"))
		}

		if newSize%5 != 0 {
			return diag.FromErr(errors.New("volume_size_in_gb must be a multiple of 5"))
		}
		size := scw.Size(newSize * uint64(scw.GB))

		upgradeInstanceRequests := mongodb.UpgradeInstanceRequest{
			InstanceID: ID,
			Region:     region,
			VolumeSize: &size,
		}

		_, err = mongodbAPI.UpgradeInstance(&upgradeInstanceRequests, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		_, err = waitForInstance(ctx, mongodbAPI, region, ID, d.Timeout(schema.TimeoutUpdate))

	}

	req := &mongodb.UpdateInstanceRequest{
		Region:     region,
		InstanceID: ID,
	}

	if d.HasChange("name") {
		req.Name = types.ExpandStringPtr(d.Get("name"))
	}

	if d.HasChange("tags") {
		if tags := types.ExpandUpdatedStringsPtr(d.Get("tags")); tags != nil {
			req.Tags = *tags
		}
	}

	_, err = mongodbAPI.UpdateInstance(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	////////////////////
	// Update user
	////////////////////

	updateUserRequest := mongodb.UpdateUserRequest{
		Region:     region,
		InstanceID: ID,
	}
	if d.HasChange("user_name") {
		updateUserRequest.Name = d.Get("user_name").(string)
	}

	if d.HasChange("password") {
		password := d.Get("password").(string)
		updateUserRequest.Password = &password
	}

	_, err = mongodbAPI.UpdateUser(&updateUserRequest, scw.WithContext(ctx))

	_, err = waitForInstance(ctx, mongodbAPI, region, ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceInstanceRead(ctx, d, m)
}

func ResourceInstanceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	mongodbAPI, region, ID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForInstance(ctx, mongodbAPI, region, ID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = mongodbAPI.DeleteInstance(&mongodb.DeleteInstanceRequest{
		Region:     region,
		InstanceID: ID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForInstance(ctx, mongodbAPI, region, ID, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
