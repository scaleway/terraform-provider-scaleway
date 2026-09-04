package instance

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	instanceSDK "github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	ipamAPI "github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/api/marketplace/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	scwvalidation "github.com/scaleway/scaleway-sdk-go/validation"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/cdf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/instancehelpers"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/ipam"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpc"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceServer() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceInstanceServerCreate,
		ReadContext:   ResourceInstanceServerRead,
		UpdateContext: ResourceInstanceServerUpdate,
		DeleteContext: ResourceInstanceServerDelete,
		Importer:      identity.DefaultZonalImporter(),
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(DefaultInstanceServerWaitTimeout),
			Read:    schema.DefaultTimeout(DefaultInstanceServerWaitTimeout),
			Update:  schema.DefaultTimeout(DefaultInstanceServerWaitTimeout),
			Delete:  schema.DefaultTimeout(DefaultInstanceServerWaitTimeout),
			Default: schema.DefaultTimeout(DefaultInstanceServerWaitTimeout),
		},
		SchemaVersion: 0,
		SchemaFunc:    serverSchema,
		Identity:      identity.DefaultZonal(),
		CustomizeDiff: customdiff.All(
			cdf.LocalityCheck(
				"placement_group_id",
				"additional_volume_ids.#",
				"ip_id",
				"ip_ids.#",
			),
			customDiffInstanceServerType,
			customDiffInstanceServerImage,
			customDiffInstanceRootVolumeSize,
		),
	}
}

func serverSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "The name of the server",
		},
		"image": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "The UUID or the label of the base image used by the server",
			DiffSuppressFunc: dsf.Locality,
			ExactlyOneOf:     []string{"image", "root_volume.0.volume_id"},
		},
		"type": {
			Type:             schema.TypeString,
			Required:         true,
			Description:      "The instance type of the server", // TODO: link to scaleway pricing in the doc
			DiffSuppressFunc: dsf.IgnoreCase,
		},
		"protected": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "If true, the instance is protected against accidental deletion via the Scaleway API.",
		},
		"replace_on_type_change": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Delete and re-create server if type change",
		},
		"tags": {
			Type: schema.TypeList,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Optional:    true,
			Description: "The tags associated with the server",
		},
		"security_group_id": {
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			DiffSuppressFunc: dsf.Locality,
			Description:      "The security group the server is attached to",
		},
		"placement_group_id": {
			Type:             schema.TypeString,
			Optional:         true,
			DiffSuppressFunc: dsf.Locality,
			Description:      "The placement group the server is attached to",
		},
		"placement_group_policy_respected": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "True when the placement group policy is respected",
		},
		"root_volume": {
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Computed:    true,
			Description: "Root volume attached to the server on creation",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:        schema.TypeString,
						Computed:    true,
						Optional:    true,
						Description: "Name of the root volume",
					},
					"size_in_gb": {
						Type:        schema.TypeInt,
						Optional:    true,
						Computed:    true,
						Description: "Size of the root volume in gigabytes",
					},
					"volume_type": {
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
						ForceNew:    true,
						Description: "Volume type of the root volume",
						ValidateDiagFunc: func(i any, path cty.Path) diag.Diagnostics {
							diags := verify.ValidateEnum[instanceSDK.VolumeVolumeType]()(i, path)
							if i.(string) == "b_ssd" {
								diags = append(diags, diag.Diagnostic{
									Severity:      diag.Error,
									Summary:       "b_ssd volumes are not supported anymore",
									Detail:        "Remove explicit b_ssd volume_type, migrate to sbs or downgrade terraform.\nLearn more about migration: https://www.scaleway.com/en/docs/instances/how-to/migrate-volumes-snapshots-to-sbs/",
									AttributePath: path,
								})
							}

							return diags
						},
					},
					"delete_on_termination": {
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     true,
						Description: "Force deletion of the root volume on instance termination",
					},
					"boot": {
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
						Description: "Set the volume where the boot the server",
					},
					"volume_id": {
						Type:         schema.TypeString,
						Computed:     true,
						Optional:     true,
						Description:  "Volume ID of the root volume",
						ExactlyOneOf: []string{"image", "root_volume.0.volume_id"},
					},
					"sbs_iops": {
						Type:        schema.TypeInt,
						Computed:    true,
						Optional:    true,
						Description: "SBS Volume IOPS, only with volume_type as sbs_volume",
					},
				},
			},
		},
		"additional_volume_ids": {
			Type: schema.TypeList,
			Elem: &schema.Schema{
				Type:             schema.TypeString,
				ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
				DiffSuppressFunc: dsf.Locality,
			},
			Optional:    true,
			Description: "The additional volumes attached to the server",
		},
		"filesystems": {
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			Description: "Filesystems attach to the server",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"filesystem_id": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The filesystem ID attached to the server",
					},
					"status": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The state of the filesystem",
					},
				},
			},
		},
		"ip_id": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "The ID of the reserved IP for the server",
			DiffSuppressFunc: dsf.Locality,
			ConflictsWith:    []string{"ip_ids"},
		},
		"ip_ids": {
			Type:          schema.TypeList,
			Description:   "The IDs of the reserved IP for the server",
			Optional:      true,
			ConflictsWith: []string{"ip_id"},
			Elem: &schema.Schema{
				Type:             schema.TypeString,
				Description:      "ID of the reserved IP for the server",
				DiffSuppressFunc: dsf.Locality,
			},
		},
		"enable_dynamic_ip": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Enable dynamic IP on the server",
		},
		"state": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "The state of the server should be: started, stopped, standby",
			ValidateFunc: validation.StringInSlice([]string{
				InstanceServerStateStarted,
				InstanceServerStateStopped,
				InstanceServerStateStandby,
			}, false),
		},
		"boot_type": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "The boot type of the server",
			Default:          instanceSDK.BootTypeLocal,
			ValidateDiagFunc: verify.ValidateEnum[instanceSDK.BootType](),
		},
		"bootscript_id": {
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			Description:      "ID of the target bootscript (set boot_type to bootscript)",
			ValidateDiagFunc: verify.IsUUID(),
			Deprecated:       "bootscript is not supported anymore.",
		},
		"cloud_init": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			Description:  "The cloud init script associated with this server",
			ValidateFunc: validation.StringLenBetween(0, 127998),
		},
		"user_data": {
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Description: "The user data associated with the server",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			DiffSuppressFunc: func(k, _, _ string, _ *schema.ResourceData) bool {
				return k == "user_data.ssh-host-fingerprints"
			},
		},
		"private_network": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    8,
			Description: "List of private network to connect with your instance",
			Elem: &schema.Resource{
				Timeouts: &schema.ResourceTimeout{
					Default: schema.DefaultTimeout(defaultInstancePrivateNICWaitTimeout),
				},
				Schema: map[string]*schema.Schema{
					"pn_id": {
						Type:             schema.TypeString,
						Required:         true,
						ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
						Description:      "The Private Network ID",
						DiffSuppressFunc: dsf.Locality,
					},
					// Computed
					"mac_address": {
						Type:        schema.TypeString,
						Description: "MAC address of the NIC",
						Computed:    true,
					},
					"status": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The private NIC state",
					},
					"pnic_id": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The ID of the NIC",
					},
					"zone": zonal.Schema(),
				},
			},
		},
		"public_ips": {
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			Description: "List of private IPv4 and IPv6 addresses attached to your instance",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "ID of the IP",
					},
					"address": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "IP Address",
					},
					"gateway": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Gateway's IP address",
					},
					"netmask": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "CIDR netmask",
					},
					"family": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "IP address family (inet or inet6)",
					},
					"dynamic": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: "Whether the IP is dynamic",
					},
					"provisioning_mode": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Provisioning mode of the IP address",
					},
				},
			},
		},
		"private_ips": {
			Type:        schema.TypeList,
			Computed:    true,
			Optional:    true,
			Description: "List of private IPv4 and IPv6 addresses associated with the resource",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The ID of the IP address resource",
					},
					"address": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The private IP address",
					},
				},
			},
		},
		"admin_password_encryption_ssh_key_id": {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: verify.IsUUIDOrEmpty(),
			Description:      "The ID of the IAM SSH key used to encrypt the initial admin password on a Windows server",
		},
		"zone":            zonal.Schema(),
		"organization_id": account.OrganizationIDSchema(),
		"project_id":      account.ProjectIDSchema(),
	}
}

//gocyclo:ignore
func ResourceInstanceServerCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := instancehelpers.InstanceAndBlockAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	////
	// Create the server
	////

	commercialType := d.Get("type").(string)

	imageUUID := locality.ExpandID(d.Get("image"))

	req := &instanceSDK.CreateServerRequest{
		Zone:              zone,
		Name:              types.ExpandOrGenerateString(d.Get("name"), "srv"),
		Project:           types.ExpandStringPtr(d.Get("project_id")),
		CommercialType:    commercialType,
		SecurityGroup:     types.ExpandStringPtr(zonal.ExpandID(d.Get("security_group_id")).ID),
		DynamicIPRequired: new(d.Get("enable_dynamic_ip").(bool)),
		Tags:              types.ExpandStrings(d.Get("tags")),
		Protected:         d.Get("protected").(bool),
	}

	if bootType, ok := d.GetOk("boot_type"); ok {
		req.BootType = new(instanceSDK.BootType(bootType.(string)))
	}

	if ipID, ok := d.GetOk("ip_id"); ok {
		req.PublicIPs = &[]string{zonal.ExpandID(ipID).ID}
	} else if ipIDs, ok := d.GetOk("ip_ids"); ok {
		req.PublicIPs = types.ExpandSliceIDsPtr(ipIDs)
	}

	if placementGroupID, ok := d.GetOk("placement_group_id"); ok {
		req.PlacementGroup = types.ExpandStringPtr(zonal.ExpandID(placementGroupID).ID)
	}

	if adminPasswordEncryptionSSHKeyID, ok := d.GetOk("admin_password_encryption_ssh_key_id"); ok {
		req.AdminPasswordEncryptionSSHKeyID = types.ExpandStringPtr(adminPasswordEncryptionSSHKeyID)
	}

	serverType := getServerType(ctx, api.API, req.Zone, req.CommercialType)
	if serverType == nil {
		return diag.Diagnostics{{
			Severity:      diag.Error,
			Summary:       fmt.Sprintf("could not find a server type associated with %s in zone %s", req.CommercialType, req.Zone),
			Detail:        "Ensure that the server type is correct, and that it does exist in this zone.",
			AttributePath: cty.GetAttrPath("type"),
		}}
	}

	req.Volumes = make(map[string]*instanceSDK.VolumeServerTemplate)
	rootVolume := d.Get("root_volume.0").(map[string]any)

	req.Volumes["0"] = prepareRootVolume(rootVolume, serverType, imageUUID).VolumeTemplate()

	if raw, ok := d.GetOk("additional_volume_ids"); ok {
		for i, volumeID := range raw.([]any) {
			// We have to get the volume to know whether it is a local or a block volume
			volumeTemplate, err := instanceServerAdditionalVolumeTemplate(api, zone, volumeID.(string))
			if err != nil {
				return diag.FromErr(fmt.Errorf("failed to get additional volume: %w", err))
			}

			req.Volumes[strconv.Itoa(i+1)] = volumeTemplate
		}
	}

	// Validate total local volume sizes.
	if err = validateLocalVolumeSizes(req.Volumes, serverType, req.CommercialType); err != nil {
		return diag.FromErr(err)
	}

	if imageUUID != "" && !scwvalidation.IsUUID(imageUUID) {
		// Replace dashes with underscores ubuntu-focal -> ubuntu_focal
		imageLabel := formatImageLabel(imageUUID)

		marketPlaceAPI := marketplace.NewAPI(meta.ExtractScwClient(m))

		image, err := marketPlaceAPI.GetLocalImageByLabel(&marketplace.GetLocalImageByLabelRequest{
			CommercialType: commercialType,
			Zone:           zone,
			ImageLabel:     imageLabel,
			Type:           volumeTypeToMarketplaceFilter(req.Volumes["0"].VolumeType),
		})
		if err != nil {
			return diag.FromErr(fmt.Errorf("could not get image '%s': %w", zonal.NewID(zone, imageLabel), err))
		}

		imageUUID = image.ID
	}

	if imageUUID != "" {
		req.Image = new(imageUUID)
	}

	res, err := api.CreateServer(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	err = identity.SetZonalIdentity(d, res.Server.Zone, res.Server.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForServer(ctx, api.API, zone, res.Server.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	////
	// Configure Volumes
	////
	var diags diag.Diagnostics

	if iops, ok := d.GetOk("root_volume.0.sbs_iops"); ok {
		updateDiags := ResourceInstanceServerUpdateRootVolumeIOPS(ctx, api, zone, res.Server.ID, types.ExpandUint32Ptr(iops))
		if len(updateDiags) > 0 {
			diags = append(diags, updateDiags...)
		}
	}

	err = renameRootVolumeIfNeeded(d, api, zone, res.Server.Volumes)
	if err != nil {
		return diag.FromErr(err)
	}

	////
	// Set user data
	////
	userDataRequests := &instanceSDK.SetAllServerUserDataRequest{
		Zone:     zone,
		ServerID: res.Server.ID,
		UserData: make(map[string]io.Reader),
	}

	if rawUserData, ok := d.GetOk("user_data"); ok {
		for key, value := range rawUserData.(map[string]any) {
			userDataRequests.UserData[key] = bytes.NewBufferString(value.(string))
		}
	}

	// cloud init script is set in user data
	if cloudInit, ok := d.GetOk("cloud_init"); ok {
		userDataRequests.UserData["cloud-init"] = bytes.NewBufferString(cloudInit.(string))
	}

	if len(userDataRequests.UserData) > 0 {
		_, err := waitForServer(ctx, api.API, zone, res.Server.ID, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.FromErr(err)
		}

		err = api.SetAllServerUserData(userDataRequests)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	targetState, err := serverStateExpand(d.Get("state").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	err = reachState(ctx, api, zone, res.Server.ID, targetState)
	if err != nil {
		return diag.FromErr(err)
	}

	///
	// Attach Filesystem
	///

	if filesystems, ok := d.GetOk("filesystems"); ok {
		for _, filesystem := range filesystems.([]any) {
			fs := filesystem.(map[string]any)
			filesystemID := fs["filesystem_id"]

			_, err := api.AttachServerFileSystem(&instanceSDK.AttachServerFileSystemRequest{
				Zone:         zone,
				FilesystemID: regional.ExpandID(filesystemID.(string)).ID,
				ServerID:     res.Server.ID,
			})
			if err != nil {
				return diag.FromErr(err)
			}

			_, err = waitForFilesystems(ctx, api.API, zone, res.Server.ID, DefaultInstanceServerWaitTimeout)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	////
	// Private Network
	////
	if rawPNICs, ok := d.GetOk("private_network"); ok {
		vpcAPI, err := vpc.NewAPI(m)
		if err != nil {
			return diag.FromErr(err)
		}

		pnRequest, err := preparePrivateNIC(ctx, rawPNICs, res.Server, vpcAPI)
		if err != nil {
			return diag.FromErr(err)
		}
		// compute attachment
		for _, q := range pnRequest {
			_, err := waitForServer(ctx, api.API, zone, res.Server.ID, d.Timeout(schema.TimeoutCreate))
			if err != nil {
				return diag.FromErr(err)
			}

			pn, err := api.InstanceV2API.CreatePrivateNetworkInterface(q, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}

			tflog.Debug(ctx, fmt.Sprintf("private network created (ID: %s, status: %s)", pn.ID, pn.Status))

			_, err = waitForPrivateNIC(ctx, api.InstanceV2API, zone, pn.ID, d.Timeout(schema.TimeoutCreate))
			if err != nil {
				return diag.FromErr(err)
			}

			err = waitForMACAddress(ctx, api.API, zone, res.Server.ID, pn.ID, d.Timeout(schema.TimeoutCreate))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return append(diags, setServerState(ctx, d, m, api, res.Server.Zone, res.Server.ID)...)
}

//gocyclo:ignore
func setServerState(ctx context.Context, d *schema.ResourceData, m any, api *instancehelpers.BlockAndInstanceAPI, zone scw.Zone, id string) diag.Diagnostics {
	server, err := waitForServer(ctx, api.API, zone, id, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if errorCheck(err, "is not found") {
			log.Printf("[WARN] instance %s not found droping from state", d.Id())
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	state, err := serverStateFlatten(server.State)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("state", state)
	_ = d.Set("zone", string(zone))
	_ = d.Set("name", server.Name)
	_ = d.Set("boot_type", server.BootType)

	_ = d.Set("type", server.CommercialType)
	if len(server.Tags) > 0 {
		_ = d.Set("tags", server.Tags)
	}

	if server.Filesystems != nil {
		_ = d.Set("filesystems", flattenServerFileSystem(server.Zone, server.Filesystems))
	}

	_ = d.Set("security_group_id", zonal.NewID(zone, server.SecurityGroup.ID).String())
	_ = d.Set("enable_dynamic_ip", server.DynamicIPRequired)
	_ = d.Set("organization_id", server.Organization)
	_ = d.Set("project_id", server.Project)
	_ = d.Set("protected", server.Protected)

	// Image could be empty in an import context.
	image := regional.ExpandID(d.Get("image").(string))
	if server.Image != nil && (image.ID == "" || scwvalidation.IsUUID(image.ID)) {
		_ = d.Set("image", zonal.NewID(zone, server.Image.ID).String())
	}

	if server.PlacementGroup != nil {
		_ = d.Set("placement_group_id", zonal.NewID(zone, server.PlacementGroup.ID).String())
		_ = d.Set("placement_group_policy_respected", server.PlacementGroup.PolicyRespected)
	}

	////
	// Read server's public IPs
	////
	if len(server.PublicIPs) > 0 {
		_ = d.Set("public_ips", flattenServerPublicIPs(server.Zone, server.PublicIPs))

		var ipIDToSet string

		var ipIDsToSet []any

		ipID, hasIPID := d.GetOk("ip_id")
		_, hasIPIDs := d.GetOk("ip_ids")

		switch {
		case hasIPID:
			publicIP := FindIPInList(ipID.(string), server.PublicIPs)
			if publicIP != nil && !publicIP.Dynamic {
				ipIDToSet = zonal.NewID(zone, publicIP.ID).String()
			}
		case hasIPIDs:
			ipIDsToSet = flattenServerIPIDs(server.PublicIPs, server.Zone)
		default:
			// In import context, we don't know if the field that was used is 'ip_id' or 'ip_ids', so we set them both
			for _, publicIP := range server.PublicIPs {
				if !publicIP.Dynamic {
					ipIDToSet = zonal.NewID(zone, publicIP.ID).String()
					ipIDsToSet = append(ipIDsToSet, zonal.NewID(zone, publicIP.ID).String())
				}
			}
		}

		_ = d.Set("ip_id", ipIDToSet)
		_ = d.Set("ip_ids", ipIDsToSet)

		d.SetConnInfo(map[string]string{
			"type": "ssh",
			"host": server.PublicIPs[0].Address.String(),
		})
	} else {
		_ = d.Set("public_ips", []any{})
		_ = d.Set("ip_ids", []any{})
		_ = d.Set("ip_id", "")
		d.SetConnInfo(nil)
	}

	if server.AdminPasswordEncryptionSSHKeyID != nil {
		_ = d.Set("admin_password_encryption_ssh_key_id", server.AdminPasswordEncryptionSSHKeyID)
	}

	////
	// Read server's volumes
	////
	rootVolume := make(map[string]any, 1)
	additionalVolumesIDs := make([]string, 0, len(server.Volumes)-1)

	for i, serverVolume := range sortVolumeServer(server.Volumes) {
		if i == 0 {
			rootVolume, err = flattenServerVolume(api, serverVolume, zone)
			if err != nil {
				return diag.FromErr(err)
			}

			_, rootVolumeAttributeSet := d.GetOk("root_volume") // Related to https://github.com/hashicorp/terraform-plugin-sdk/issues/142
			rootVolume["delete_on_termination"] = d.Get("root_volume.0.delete_on_termination").(bool) || !rootVolumeAttributeSet
		} else {
			additionalVolumesIDs = append(additionalVolumesIDs, zonal.NewID(zone, serverVolume.ID).String())
		}
	}

	_ = d.Set("root_volume", []map[string]any{rootVolume})
	_ = d.Set("additional_volume_ids", additionalVolumesIDs)

	////
	// Read server user data
	////
	allUserData, err := api.GetAllServerUserData(&instanceSDK.GetAllServerUserDataRequest{
		Zone:     zone,
		ServerID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	userData := make(map[string]any)

	for key, value := range allUserData.UserData {
		userDataValue, err := io.ReadAll(value)
		if err != nil {
			return diag.FromErr(err)
		}

		userData[key] = string(userDataValue)
	}

	_ = d.Set("user_data", userData)

	////
	// Display warning if server will soon reach End of Service
	////
	diags := diag.Diagnostics{}

	if server.EndOfService {
		eosDate, err := GetEndOfServiceDate(ctx, meta.ExtractScwClient(m), server.Zone, server.CommercialType)
		if err != nil {
			return diag.FromErr(err)
		}

		compatibleTypes, err := api.GetServerCompatibleTypes(&instanceSDK.GetServerCompatibleTypesRequest{
			Zone:     zone,
			ServerID: id,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		mostRelevantTypes := compatibleTypes.CompatibleTypes[:5]

		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Detail:   fmt.Sprintf("Instance type %q will soon reach End of Service", server.CommercialType),
			Summary: fmt.Sprintf(`Your Instance will reach End of Service by %s. We recommend that you migrate your Instance before that.
Here are the %d best options for %q, ordered by relevance: [%s]

You can check the full list of compatible server types:
	- on the Scaleway console
	- using the CLI command 'scw instance server get-compatible-types %s zone=%s'`,
				eosDate,
				len(mostRelevantTypes),
				server.CommercialType,
				strings.Join(mostRelevantTypes, ", "),
				server.ID,
				server.Zone,
			),
			AttributePath: cty.GetAttrPath("type"),
		})
	}

	////
	// Read server private networks
	////
	ph, err := newPrivateNICHandler(api.InstanceV2API, api.API, id, zone, server.Project)
	if err != nil {
		return diag.FromErr(err)
	}

	// set private networks
	err = ph.set(d)
	if err != nil {
		return diag.FromErr(err)
	}

	privateNICIDs := []string(nil)
	for _, nic := range ph.privateNICsMap {
		privateNICIDs = append(privateNICIDs, nic.ID)
	}

	// Read server's private IPs if possible
	allPrivateIPs := []map[string]any(nil)
	resourceType := ipamAPI.ResourceTypeInstancePrivateNic

	region, err := zone.Region()
	if err != nil {
		return append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Unable to get server's private IPs",
			Detail:   err.Error(),
		})
	}

	for _, nicID := range privateNICIDs {
		opts := &ipam.GetResourcePrivateIPsOptions{
			ResourceType: &resourceType,
			ResourceID:   &nicID,
			ProjectID:    &server.Project,
		}

		privateIPs, err := ipam.GetResourcePrivateIPs(ctx, m, region, opts)

		switch {
		case err == nil:
			allPrivateIPs = append(allPrivateIPs, privateIPs...)
		case httperrors.Is403(err):
			return append(diags, diag.Diagnostic{
				Severity:      diag.Warning,
				Summary:       "Unauthorized to read server's private IPs, please check your IAM permissions",
				Detail:        err.Error(),
				AttributePath: cty.GetAttrPath("private_ips"),
			})
		default:
			diags = append(diags, diag.Diagnostic{
				Severity:      diag.Warning,
				Summary:       fmt.Sprintf("Unable to get private IPs for server %s (pnic_id: %s)", server.ID, nicID),
				Detail:        err.Error(),
				AttributePath: cty.GetAttrPath("private_ips"),
			})
		}
	}

	_ = d.Set("private_ips", allPrivateIPs)

	return diags
}

func ResourceInstanceServerRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, id, err := instancehelpers.InstanceAndBlockAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = identity.SetZonalIdentity(d, zone, id)
	if err != nil {
		return diag.FromErr(err)
	}

	return setServerState(ctx, d, m, api, zone, id)
}

//gocyclo:ignore
func ResourceInstanceServerUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, id, err := instancehelpers.InstanceAndBlockAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	wantedState := d.Get("state").(string)
	isStopped := wantedState == InstanceServerStateStopped

	var warnings diag.Diagnostics

	server, err := waitForServer(ctx, api.API, zone, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}
	////
	// Construct UpdateServerRequest
	////
	serverShouldUpdate := false
	updateRequest := &instanceSDK.UpdateServerRequest{
		Zone:     zone,
		ServerID: server.ID,
	}

	if d.HasChange("name") {
		serverShouldUpdate = true
		updateRequest.Name = types.ExpandStringPtr(d.Get("name"))
	}

	if d.HasChange("tags") {
		serverShouldUpdate = true
		updateRequest.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
	}

	if d.HasChange("security_group_id") {
		serverShouldUpdate = true
		updateRequest.SecurityGroup = &instanceSDK.SecurityGroupTemplate{
			ID:   zonal.ExpandID(d.Get("security_group_id")).ID,
			Name: types.NewRandomName("sg"), // this value will be ignored by the API
		}
	}

	if d.HasChange("enable_dynamic_ip") {
		serverShouldUpdate = true
		updateRequest.DynamicIPRequired = new(d.Get("enable_dynamic_ip").(bool))
	}

	if d.HasChange("protected") {
		serverShouldUpdate = true
		updateRequest.Protected = types.ExpandBoolPtr(d.Get("protected").(bool))
	}

	if d.HasChanges("additional_volume_ids", "root_volume") {
		volumes, err := instanceServerVolumesUpdate(ctx, d, api, zone, isStopped)
		if err != nil {
			return diag.FromErr(err)
		}

		serverShouldUpdate = true
		updateRequest.Volumes = &volumes
	}

	if d.HasChange("placement_group_id") {
		serverShouldUpdate = true

		placementGroupID := zonal.ExpandID(d.Get("placement_group_id")).ID
		if placementGroupID == "" {
			updateRequest.PlacementGroup = &instanceSDK.NullableStringValue{Null: true}
		} else {
			if !isStopped {
				return diag.FromErr(errors.New("instance must be stopped to change placement group"))
			}

			updateRequest.PlacementGroup = &instanceSDK.NullableStringValue{Value: placementGroupID}
		}
	}

	if d.HasChange("admin_password_encryption_ssh_key_id") {
		serverShouldUpdate = true
		updateRequest.AdminPasswordEncryptionSSHKeyID = types.ExpandUpdatedStringPtr(d.Get("admin_password_encryption_ssh_key_id").(string))
	}

	////
	// Update reserved IP
	////
	if d.HasChange("ip_id") && !instanceIPHasMigrated(d) {
		ipID := zonal.ExpandID(d.Get("ip_id")).ID
		if ipID == "" {
			updateRequest.PublicIPs = new(make([]string, 0))
			serverShouldUpdate = true
		} else {
			err := ResourceInstanceServerUpdateIPs(ctx, d, api.API, zone, id, "ip_id")
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("ip_ids") {
		err := ResourceInstanceServerUpdateIPs(ctx, d, api.API, zone, id, "ip_ids")
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChanges("boot_type") {
		serverShouldUpdate = true
		updateRequest.BootType = new(instanceSDK.BootType(d.Get("boot_type").(string)))

		if !isStopped {
			warnings = append(warnings, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "instance may need to be rebooted to use the new boot type",
			})
		}
	}

	////
	// Update server user data
	////
	if d.HasChanges("user_data") {
		userDataRequests := &instanceSDK.SetAllServerUserDataRequest{
			Zone:     zone,
			ServerID: id,
			UserData: make(map[string]io.Reader),
		}

		if allUserData, ok := d.GetOk("user_data"); ok {
			userDataMap := allUserData.(map[string]any)
			for key, value := range userDataMap {
				userDataRequests.UserData[key] = bytes.NewBufferString(value.(string))
			}

			if !isStopped && d.HasChange("user_data.cloud-init") {
				warnings = append(warnings, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "instance may need to be rebooted to use the new cloud init config",
				})
			}
		}

		_, err := waitForServer(ctx, api.API, zone, id, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}

		err = api.SetAllServerUserData(userDataRequests)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	////
	// Update server filesystems
	///

	if d.HasChange("filesystems") {
		oldRaw, newRaw := d.GetChange("filesystems")

		oldList := oldRaw.([]any)
		newList := newRaw.([]any)

		oldIDs := make(map[string]struct{})
		newIDs := make(map[string]struct{})

		collectFilesystemIDs(oldList, oldIDs)
		collectFilesystemIDs(newList, newIDs)

		err := detachOldFileSystem(ctx, oldIDs, newIDs, api.API, zone, server)
		if err != nil {
			return diag.FromErr(err)
		}

		err = attachNewFileSystem(ctx, newIDs, oldIDs, api.API, zone, server)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	////
	// Update server private network
	////
	if d.HasChanges("private_network") {
		ph, err := newPrivateNICHandler(api.InstanceV2API, api.API, id, zone, server.Project)
		if err != nil {
			diag.FromErr(err)
		}

		if raw, ok := d.GetOk("private_network"); ok {
			// retrieve all current private network interfaces
			for index := range raw.([]any) {
				pnKey := fmt.Sprintf("private_network.%d.pn_id", index)
				if d.HasChange(pnKey) {
					o, n := d.GetChange(pnKey)
					if !cmp.Equal(n, o) {
						_, err := waitForServer(ctx, api.API, zone, id, d.Timeout(schema.TimeoutUpdate))
						if err != nil {
							return diag.FromErr(err)
						}

						err = ph.detach(ctx, o, d.Timeout(schema.TimeoutUpdate))
						if err != nil {
							return diag.FromErr(err)
						}

						err = ph.attach(ctx, n, d.Timeout(schema.TimeoutUpdate))
						if err != nil {
							return diag.FromErr(err)
						}
					}
				}
			}
		} else {
			// retrieve old private network config
			o, _ := d.GetChange("private_network")
			for _, raw := range o.([]any) {
				pn, pnExist := raw.(map[string]any)
				if pnExist {
					_, err := waitForServer(ctx, api.API, zone, id, d.Timeout(schema.TimeoutUpdate))
					if err != nil {
						return diag.FromErr(err)
					}

					err = ph.detach(ctx, pn["pn_id"], d.Timeout(schema.TimeoutUpdate))
					if err != nil {
						return diag.FromErr(err)
					}
				}
			}
		}
	}
	////
	// Apply changes
	////

	if d.HasChange("state") {
		targetState, err := serverStateExpand(d.Get("state").(string))
		if err != nil {
			return diag.FromErr(err)
		}
		// reach expected state
		err = reachState(ctx, api, zone, id, targetState)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if serverShouldUpdate {
		_, err = api.UpdateServer(updateRequest)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	server, err = waitForServer(ctx, api.API, zone, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("type") {
		err := ResourceInstanceServerMigrate(ctx, d, api, zone, id)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChanges("root_volume.0.sbs_iops") {
		warnings = append(warnings, ResourceInstanceServerUpdateRootVolumeIOPS(ctx, api, zone, id, types.ExpandUint32Ptr(d.Get("root_volume.0.sbs_iops")))...)
	}

	err = renameRootVolumeIfNeeded(d, api, zone, server.Volumes)
	if err != nil {
		return diag.FromErr(err)
	}

	return append(warnings, ResourceInstanceServerRead(ctx, d, m)...)
}

func ResourceInstanceServerDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, id, err := instancehelpers.InstanceAndBlockAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	// detach eip to ensure to free eip even if instance won't stop
	if ipID, ok := d.GetOk("ip_id"); ok {
		_, err := api.UpdateIP(&instanceSDK.UpdateIPRequest{
			Zone:   zone,
			IP:     zonal.ExpandID(ipID).ID,
			Server: &instanceSDK.NullableStringValue{Null: true},
		}, scw.WithContext(ctx))
		if err != nil {
			log.Print("[WARN] Failed to detach eip of server")
		}
	} else if ipIDs, ok := d.GetOk("ip_ids"); ok {
		for _, ipID := range ipIDs.([]any) {
			_, err := api.UpdateIP(&instanceSDK.UpdateIPRequest{
				Zone:   zone,
				IP:     zonal.ExpandID(ipID).ID,
				Server: &instanceSDK.NullableStringValue{Null: true},
			}, scw.WithContext(ctx))
			if err != nil {
				log.Print("[WARN] Failed to detach eip of server")
			}
		}
	}

	// Remove instance from placement group to free it even if instance won't stop
	if _, ok := d.GetOk("placement_group_id"); ok {
		_, err := api.UpdateServer(&instanceSDK.UpdateServerRequest{
			Zone:           zone,
			PlacementGroup: &instanceSDK.NullableStringValue{Null: true},
			ServerID:       id,
		}, scw.WithContext(ctx))
		if err != nil {
			log.Print("[WARN] Failed remove server from instance group")
		}
	}

	// Detach filesystem
	if filesystems, ok := d.GetOk("filesystems"); ok {
		diags := detachFileSystemDelete(ctx, filesystems, api.API, zone, id)
		if diags.HasError() {
			return diags
		}
	}

	server, err := api.GetServer(&instanceSDK.GetServerRequest{
		Zone:     zone,
		ServerID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	// Delete private-nic if managed by instance_server resource
	if pnRaw, ok := d.GetOk("private_network"); ok {
		diags := detachPrivateNetworkDelete(ctx, d, pnRaw, api.InstanceV2API, api.API, zone, id, server.Server.Project)
		if diags.HasError() {
			return diags
		}
	}

	_, err = waitForServer(ctx, api.API, zone, id, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	err = terminateServer(ctx, api, zone, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		err = deleteServer(ctx, api, zone, id, d.Timeout(schema.TimeoutDelete))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Related to https://github.com/hashicorp/terraform-plugin-sdk/issues/142
	_, rootVolumeAttributeSet := d.GetOk("root_volume")
	if d.Get("root_volume.0.delete_on_termination").(bool) || !rootVolumeAttributeSet {
		volumeID, volumeExist := d.GetOk("root_volume.0.volume_id")
		if !volumeExist {
			return diag.Errorf("volume ID not found")
		}

		err = api.DeleteUnknownVolume(&instancehelpers.DeleteUnknownVolumeRequest{
			Zone:     zone,
			VolumeID: locality.ExpandID(volumeID),
		})
		if err != nil && !httperrors.Is404(err) {
			return diag.FromErr(err)
		}
	}

	return nil
}
