package mongodb

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	ipamAPI "github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	mongodb "github.com/scaleway/scaleway-sdk-go/api/mongodb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/ipam"
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
				Description: "Name of the MongoDB cluster",
			},
			"version": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "MongoDB version of the instance",
				ConflictsWith: []string{
					"snapshot_id",
				},
			},
			"node_number": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
				Description:  "Number of nodes in the instance",
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
				ConflictsWith: []string{
					"snapshot_id",
				},
			},
			"password": {
				Type:        schema.TypeString,
				Sensitive:   true,
				Optional:    true,
				Description: "Password of the user",
				ConflictsWith: []string{
					"snapshot_id",
				},
			},
			// volume
			"volume_type": {
				Type:        schema.TypeString,
				Default:     mongodb.VolumeTypeSbs5k,
				Optional:    true,
				Description: "Volume type of the instance",
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
				Description: "Snapshot ID to restore the MongoDB instance from",
				ConflictsWith: []string{
					"user_name",
					"password",
					"version",
				},
			},
			"private_network": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Private network to expose your mongodb instance",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pn_id": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
							DiffSuppressFunc: dsf.Locality,
							Description:      "The private network ID",
						},
						// Computed
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The private network ID",
						},
						"port": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "TCP port of the endpoint",
						},
						"dns_records": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of DNS records for your endpoint",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},

						"ips": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of IP addresses for your endpoint",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			// Computed
			"private_ip": {
				Type:        schema.TypeList,
				Computed:    true,
				Optional:    true,
				Description: "The private IPv4 address associated with the resource",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the IPv4 address resource",
						},
						"address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The private IPv4 address",
						},
					},
				},
			},
			"public_network": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Public network specs details",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the public network",
						},
						"port": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "TCP port of the endpoint",
						},
						"dns_record": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The DNS record of your endpoint",
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
				Description: "List of tags [\"tag1\", \"tag2\", ...] attached to a MongoDB instance",
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
				Description: "The date and time of the creation of the MongoDB instance",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last update of the MongoDB instance",
			},
			// Common
			"region":     regional.Schema(),
			"project_id": account.ProjectIDSchema(),
			"tls_certificate": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "PEM-encoded TLS certificate for MongoDB",
			},
		},
		CustomizeDiff: customdiff.All(
			func(ctx context.Context, d *schema.ResourceDiff, meta any) error {
				if d.HasChange("version") {
					v := d.Get("version").(string)
					parts := strings.Split(v, ".")
					if len(parts) > 2 {
						majorMinor := parts[0] + "." + parts[1]
						if err := d.SetNew("version", majorMinor); err != nil {
							return err
						}
					}
				}

				return nil
			},
		),
	}
}

func ResourceInstanceCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	mongodbAPI, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	nodeNumber := scw.Uint32Ptr(uint32(d.Get("node_number").(int)))

	snapshotID, exist := d.GetOk("snapshot_id")

	var res *mongodb.Instance

	if exist {
		restoreSnapshotRequest := &mongodb.RestoreSnapshotRequest{
			SnapshotID:   regional.ExpandID(snapshotID.(string)).ID,
			InstanceName: types.ExpandOrGenerateString(d.Get("name"), "mongodb"),
			NodeAmount:   *nodeNumber,
			NodeType:     d.Get("node_type").(string),
			VolumeType:   mongodb.VolumeType(d.Get("volume_type").(string)),
		}

		res, err = mongodbAPI.RestoreSnapshot(restoreSnapshotRequest, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		version := d.Get("version").(string)
		normalizeVersion := NormalizeMongoDBVersion(version)

		createReq := &mongodb.CreateInstanceRequest{
			ProjectID:  d.Get("project_id").(string),
			Name:       types.ExpandOrGenerateString(d.Get("name"), "mongodb"),
			Version:    normalizeVersion,
			NodeType:   d.Get("node_type").(string),
			NodeAmount: *nodeNumber,
			UserName:   d.Get("user_name").(string),
			Password:   d.Get("password").(string),
		}

		volumeRequestDetails := &mongodb.Volume{
			Type: mongodb.VolumeType(d.Get("volume_type").(string)),
		}
		volumeSize, volumeSizeExist := d.GetOk("volume_size_in_gb")

		if volumeSizeExist {
			volumeRequestDetails.SizeBytes = scw.Size(uint64(volumeSize.(int)) * uint64(scw.GB))
		} else {
			volumeRequestDetails.SizeBytes = scw.Size(defaultVolumeSize * uint64(scw.GB))
		}

		createReq.Volume = volumeRequestDetails

		tags, tagsExist := d.GetOk("tags")
		if tagsExist {
			createReq.Tags = types.ExpandStrings(tags)
		}

		var eps []*mongodb.EndpointSpec

		if privateNetworkList, ok := d.GetOk("private_network"); ok {
			privateNetworks := privateNetworkList.([]any)

			if len(privateNetworks) > 0 {
				pn := privateNetworks[0].(map[string]any)
				privateNetworkID := locality.ExpandID(pn["pn_id"].(string))

				if privateNetworkID != "" {
					eps = append(eps, &mongodb.EndpointSpec{
						PrivateNetwork: &mongodb.EndpointSpecPrivateNetworkDetails{
							PrivateNetworkID: privateNetworkID,
						},
					})
				}
			}
		}

		if pubList, ok := d.GetOk("public_network"); ok {
			items := pubList.([]any)
			if len(items) > 0 {
				eps = append(eps, &mongodb.EndpointSpec{
					PublicNetwork: &mongodb.EndpointSpecPublicNetworkDetails{},
				})
			}
		}

		if len(eps) == 0 {
			eps = append(eps, &mongodb.EndpointSpec{
				PublicNetwork: &mongodb.EndpointSpecPublicNetworkDetails{},
			})
		}

		createReq.Endpoints = eps

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

func ResourceInstanceRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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
	_ = d.Set("node_number", int(instance.NodeAmount))
	_ = d.Set("node_type", instance.NodeType)
	_ = d.Set("project_id", instance.ProjectID)
	_ = d.Set("tags", instance.Tags)
	_ = d.Set("created_at", instance.CreatedAt.Format(time.RFC3339))
	_ = d.Set("region", instance.Region.String())

	if instance.Volume != nil {
		_ = d.Set("volume_type", instance.Volume.Type)
		_ = d.Set("volume_size_in_gb", int(instance.Volume.SizeBytes/scw.GB))
	}

	publicNetworkEndpoint, publicNetworkExists := flattenPublicNetwork(instance.Endpoints)
	if publicNetworkExists {
		_ = d.Set("public_network", publicNetworkEndpoint)
	}

	diags := diag.Diagnostics{}
	privateIPs := []map[string]any(nil)
	authorized := true

	privateNetworkEndpoint, privateNetworkExists := flattenPrivateNetwork(instance.Endpoints)

	if privateNetworkExists {
		_ = d.Set("private_network", privateNetworkEndpoint)

		for _, endpoint := range instance.Endpoints {
			if endpoint.PrivateNetwork == nil {
				continue
			}

			resourceType := ipamAPI.ResourceTypeMgdbInstance
			opts := &ipam.GetResourcePrivateIPsOptions{
				ResourceID:       &instance.ID,
				ResourceType:     &resourceType,
				PrivateNetworkID: &endpoint.PrivateNetwork.PrivateNetworkID,
				ProjectID:        &instance.ProjectID,
			}

			endpointPrivateIPs, err := ipam.GetResourcePrivateIPs(ctx, m, region, opts)

			switch {
			case err == nil:
				privateIPs = append(privateIPs, endpointPrivateIPs...)
			case httperrors.Is403(err):
				authorized = false

				diags = append(diags, diag.Diagnostic{
					Severity:      diag.Warning,
					Summary:       "Unauthorized to read MongoDB Instance's private IP, please check your IAM permissions",
					Detail:        err.Error(),
					AttributePath: cty.GetAttrPath("private_ip"),
				})
			default:
				diags = append(diags, diag.Diagnostic{
					Severity:      diag.Warning,
					Summary:       fmt.Sprintf("Unable to get private IP for instance %q", instance.Name),
					Detail:        err.Error(),
					AttributePath: cty.GetAttrPath("private_ip"),
				})
			}

			if !authorized {
				break
			}
		}
	}

	if authorized {
		_ = d.Set("private_ip", privateIPs)
	}

	_ = d.Set("settings", map[string]string{})

	cert, err := mongodbAPI.GetInstanceCertificate(&mongodb.GetInstanceCertificateRequest{
		Region:     region,
		InstanceID: ID,
	}, scw.WithContext(ctx))

	var buf bytes.Buffer

	if err == nil && cert != nil {
		_, copyErr := io.Copy(&buf, cert.Content)
		if copyErr == nil {
			_ = d.Set("tls_certificate", buf.String())
		}
	}

	return diags
}

func ResourceInstanceUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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
			InstanceID:      ID,
			Region:          region,
			VolumeSizeBytes: &size,
		}

		_, err = mongodbAPI.UpgradeInstance(&upgradeInstanceRequests, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = waitForInstance(ctx, mongodbAPI, region, ID, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	////////////////////
	// Update instance
	////////////////////

	shouldUpdateInstance := false
	req := &mongodb.UpdateInstanceRequest{
		Region:     region,
		InstanceID: ID,
	}

	if d.HasChange("name") {
		req.Name = types.ExpandStringPtr(d.Get("name"))
		shouldUpdateInstance = true
	}

	if d.HasChange("tags") {
		if tags := types.ExpandUpdatedStringsPtr(d.Get("tags")); tags != nil {
			req.Tags = tags
			shouldUpdateInstance = true
		}
	}

	if shouldUpdateInstance {
		_, err = mongodbAPI.UpdateInstance(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	////////////////////
	// Update user
	////////////////////

	shouldUpdateUser := false
	updateUserRequest := mongodb.UpdateUserRequest{
		Name:       d.Get("user_name").(string),
		Region:     region,
		InstanceID: ID,
	}

	var diags diag.Diagnostics

	if d.HasChange("user_name") {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Change in 'user_name' detected",
			Detail: "[WARN] The 'user_name' field was changed, but this functionality is not supported yet. " +
				"As a result, no changes were applied to the instance. " +
				"Please be aware that changing the 'user_name' will not modify the instance at this time.",
		})
	}

	if d.HasChange("password") {
		password := d.Get("password").(string)
		updateUserRequest.Password = &password
		shouldUpdateUser = true
	}

	if shouldUpdateUser {
		_, err = mongodbAPI.UpdateUser(&updateUserRequest, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	////////////////////
	// Endpoints
	////////////////////

	if d.HasChange("private_network") {
		res, err := waitForInstance(ctx, mongodbAPI, region, ID, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}

		for _, e := range res.Endpoints {
			if e.PrivateNetwork != nil {
				err := mongodbAPI.DeleteEndpoint(
					&mongodb.DeleteEndpointRequest{
						EndpointID: e.ID, Region: region,
					},
					scw.WithContext(ctx))
				if err != nil {
					diag.FromErr(err)
				}
			}
		}

		_, err = waitForInstance(ctx, mongodbAPI, region, ID, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}

		var eps []*mongodb.EndpointSpec

		if privateNetworkList, ok := d.GetOk("private_network"); ok {
			privateNetworks := privateNetworkList.([]any)
			if len(privateNetworks) > 0 {
				pn := privateNetworks[0].(map[string]any)
				privateNetworkID := locality.ExpandID(pn["pn_id"].(string))

				if privateNetworkID != "" {
					eps = append(eps, &mongodb.EndpointSpec{
						PrivateNetwork: &mongodb.EndpointSpecPrivateNetworkDetails{
							PrivateNetworkID: privateNetworkID,
						},
					})
				}
			}

			if len(eps) != 0 {
				_, err = mongodbAPI.CreateEndpoint(&mongodb.CreateEndpointRequest{
					InstanceID: ID,
					Endpoint:   eps[0],
					Region:     region,
				}, scw.WithContext(ctx))
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}

	_, err = waitForInstance(ctx, mongodbAPI, region, ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return append(diags, ResourceInstanceRead(ctx, d, m)...)
}

func ResourceInstanceDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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
