package k8s

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var NetworkingDefaultValues = map[string]string{
	"pod_cidr":       "100.64.0.0/15",
	"service_cidr":   "10.32.0.0/20",
	"service_dns_ip": "10.32.0.10",
}

//go:embed descriptions/cluster.md
var clusterDescription string

func ResourceCluster() *schema.Resource {
	return &schema.Resource{
		Description:   clusterDescription,
		CreateContext: ResourceK8SClusterCreate,
		ReadContext:   ResourceK8SClusterRead,
		UpdateContext: ResourceK8SClusterUpdate,
		DeleteContext: ResourceK8SClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultK8SClusterTimeout),
			Read:    schema.DefaultTimeout(defaultK8SClusterTimeout),
			Update:  schema.DefaultTimeout(defaultK8SClusterTimeout),
			Delete:  schema.DefaultTimeout(defaultK8SClusterTimeout),
			Default: schema.DefaultTimeout(defaultK8SClusterTimeout),
		},
		SchemaVersion: 0,
		SchemaFunc:    clusterSchema,
		CustomizeDiff: customdiff.All(
			func(_ context.Context, diff *schema.ResourceDiff, _ any) error {
				autoUpgradeEnable, okAutoUpgradeEnable := diff.GetOkExists("auto_upgrade.0.enable")

				version := diff.Get("version").(string)
				versionIsOnlyMinor := len(strings.Split(version, ".")) == 2

				if okAutoUpgradeEnable && autoUpgradeEnable.(bool) && !versionIsOnlyMinor {
					return errors.New("only minor version x.y can be used with auto upgrade enabled")
				}

				if versionIsOnlyMinor && !autoUpgradeEnable.(bool) {
					return errors.New("minor version x.y must only be used with auto upgrade enabled")
				}

				return nil
			},
			func(_ context.Context, diff *schema.ResourceDiff, _ any) error {
				if diff.HasChange("private_network_id") {
					actual, planned := diff.GetChange("private_network_id")
					clusterType := diff.Get("type").(string)

					switch {
					// For Kosmos clusters
					case strings.HasPrefix(clusterType, "multicloud"):
						if planned != "" {
							return errors.New("only Kapsule clusters support private networks")
						}

					// For Kapsule clusters
					case clusterType == "" || strings.HasPrefix(clusterType, "kapsule"):
						if actual == "" {
							// If no private network has been set yet, migrate the cluster in the Update function
							return nil
						}

						if planned != "" {
							_, plannedPNID, err := locality.ParseLocalizedID(planned.(string))
							if err != nil {
								return err
							}

							if plannedPNID == actual {
								// If the private network ID is the same, do nothing
								return nil
							}
						}
						// Any other change will result in ForceNew
						err := diff.ForceNew("private_network_id")
						if err != nil {
							return err
						}
					}
				}

				return nil
			},
			func(ctx context.Context, diff *schema.ResourceDiff, i any) error {
				if diff.HasChange("type") && diff.Id() != "" {
					k8sAPI, region, clusterID, err := NewAPIWithRegionAndID(i, diff.Id())
					if err != nil {
						return err
					}

					possibleTypes, err := k8sAPI.ListClusterAvailableTypes(&k8s.ListClusterAvailableTypesRequest{
						Region:    region,
						ClusterID: clusterID,
					}, scw.WithContext(ctx))
					if err != nil {
						return err
					}

					planned := diff.Get("type")
					for _, possibleType := range possibleTypes.ClusterTypes {
						if possibleType.Name == planned {
							return nil
						}
					}

					err = diff.ForceNew("type")
					if err != nil {
						return err
					}
				}

				return nil
			},
		),
	}
}

func clusterSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the cluster",
		},
		"type": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "The type of cluster",
		},
		"description": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "The description of the cluster",
		},
		"version": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The version of the cluster",
		},
		"cni": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The CNI plugin of the cluster",
			ValidateDiagFunc: func(i any, p cty.Path) diag.Diagnostics {
				cniValues := k8s.CNI("").Values()

				cniStringValues := make([]string, 0, len(cniValues))
				for _, cniValue := range cniValues {
					cniStringValues = append(cniStringValues, cniValue.String())
				}

				return verify.ValidateStringInSliceWithWarning(cniStringValues, "cni")(i, p)
			},
		},
		"tags": {
			Type: schema.TypeList,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Optional:    true,
			Description: "The tags associated with the cluster",
		},
		"autoscaler_config": {
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Computed:    true,
			Description: "The autoscaler configuration for the cluster",
			Elem:        autoscalerConfigSchema(),
		},
		"auto_upgrade": {
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Computed:    true,
			Description: "The auto upgrade configuration for the cluster",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"enable": {
						Type:        schema.TypeBool,
						Required:    true,
						Description: "Enables the Kubernetes patch version auto upgrade",
					},
					"maintenance_window_start_hour": {
						Type:         schema.TypeInt,
						Required:     true,
						Description:  "Start hour of the 2-hour maintenance window",
						ValidateFunc: validation.IntBetween(0, 23),
					},
					"maintenance_window_day": {
						Type:             schema.TypeString,
						Required:         true,
						Description:      "Day of the maintenance window",
						ValidateDiagFunc: verify.ValidateEnum[k8s.MaintenanceWindowDayOfTheWeek](),
					},
				},
			},
		},
		"feature_gates": {
			Type: schema.TypeList,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Optional:    true,
			Description: "The list of feature gates to enable on the cluster",
		},
		"admission_plugins": {
			Type: schema.TypeList,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Optional:    true,
			Description: "The list of admission plugins to enable on the cluster",
		},
		"open_id_connect_config": {
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Computed:    true,
			Description: "The OpenID Connect configuration of the cluster",
			Elem:        openIDConnectConfigSchema(),
		},
		"apiserver_cert_sans": {
			Type: schema.TypeList,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Optional:    true,
			Description: "Additional Subject Alternative Names for the Kubernetes API server certificate",
		},
		"delete_additional_resources": {
			Type:        schema.TypeBool,
			Required:    true,
			Description: "Delete additional resources like block volumes, load-balancers and the private network (if empty) on cluster deletion",
		},
		"private_network_id": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "The ID of the cluster's private network",
			ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
			DiffSuppressFunc: dsf.Locality,
		},
		"pod_cidr": {
			Type:         schema.TypeString,
			Optional:     true,
			ForceNew:     true,
			Computed:     true,
			Description:  "The subnet used for the Pod CIDR.",
			ValidateFunc: validation.IsCIDR,
		},
		"service_cidr": {
			Type:         schema.TypeString,
			Optional:     true,
			ForceNew:     true,
			Computed:     true,
			Description:  "The subnet used for the Service CIDR.",
			ValidateFunc: validation.IsCIDR,
		},
		"service_dns_ip": {
			Type:         schema.TypeString,
			Optional:     true,
			ForceNew:     true,
			Computed:     true,
			Description:  "The IP used for the DNS Service.",
			ValidateFunc: validation.IsIPAddress,
		},
		"region":          regional.Schema(),
		"organization_id": account.OrganizationIDSchema(),
		"project_id":      account.ProjectIDSchema(),
		// Computed elements
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the creation of the Kubernetes cluster",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the last update of the Kubernetes cluster",
		},
		"apiserver_url": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Kubernetes API server URL",
		},
		"wildcard_dns": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Wildcard DNS pointing to all the ready nodes",
		},
		"kubeconfig": {
			Type:        schema.TypeList,
			Computed:    true,
			Sensitive:   true,
			Description: "The kubeconfig configuration file of the Kubernetes cluster",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"config_file": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The whole kubeconfig file",
					},
					"host": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The kubernetes master URL",
					},
					"cluster_ca_certificate": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The kubernetes cluster CA certificate",
					},
					"token": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The kubernetes cluster admin token",
					},
				},
			},
		},
		"upgrade_available": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "True if an upgrade is available",
		},
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The status of the cluster",
		},
	}
}

//gocyclo:ignore
func ResourceK8SClusterCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	k8sAPI, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	////
	// Create cluster
	////

	var diags diag.Diagnostics

	description, ok := d.GetOk("description")
	if !ok {
		description = ""
	}

	clusterType, ok := d.GetOk("type")
	if !ok {
		clusterType = ""
	}

	if clusterType != "" && !strings.HasPrefix(clusterType.(string), "kapsule") && !strings.HasPrefix(clusterType.(string), "multicloud") {
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Warning,
			Summary:       "Unexpected cluster type",
			Detail:        fmt.Sprintf("The expected cluster type is one of %v, but got %s", []string{"kapsule", "multicloud", "kapsule-dedicated-*", "multicloud-dedicated-*"}, clusterType.(string)),
			AttributePath: cty.GetAttrPath("type"),
		})
	}

	req := &k8s.CreateClusterRequest{
		Region:            region,
		ProjectID:         types.ExpandStringPtr(d.Get("project_id")),
		Name:              types.ExpandOrGenerateString(d.Get("name"), "cluster"),
		Type:              clusterType.(string),
		Description:       description.(string),
		Cni:               k8s.CNI(d.Get("cni").(string)),
		Tags:              types.ExpandStrings(d.Get("tags")),
		FeatureGates:      types.ExpandStrings(d.Get("feature_gates")),
		AdmissionPlugins:  types.ExpandStrings(d.Get("admission_plugins")),
		ApiserverCertSans: types.ExpandStrings(d.Get("apiserver_cert_sans")),
	}

	// Autoscaler configuration

	autoscalerReq := &k8s.CreateClusterRequestAutoscalerConfig{}

	if scaleDownDisabled, ok := d.GetOk("autoscaler_config.0.disable_scale_down"); ok {
		autoscalerReq.ScaleDownDisabled = scw.BoolPtr(scaleDownDisabled.(bool))
	}

	if scaleDownDelayAfterAdd, ok := d.GetOk("autoscaler_config.0.scale_down_delay_after_add"); ok {
		autoscalerReq.ScaleDownDelayAfterAdd = types.ExpandStringPtr(scaleDownDelayAfterAdd)
	}

	if scaleDownUneededTime, ok := d.GetOk("autoscaler_config.0.scale_down_unneeded_time"); ok {
		autoscalerReq.ScaleDownUnneededTime = types.ExpandStringPtr(scaleDownUneededTime)
	}

	if estimator, ok := d.GetOk("autoscaler_config.0.estimator"); ok {
		autoscalerReq.Estimator = k8s.AutoscalerEstimator(estimator.(string))
	}

	if expander, ok := d.GetOk("autoscaler_config.0.expander"); ok {
		autoscalerReq.Expander = k8s.AutoscalerExpander(expander.(string))
	}

	if ignoreDaemonsetsUtilization, ok := d.GetOk("autoscaler_config.0.ignore_daemonsets_utilization"); ok {
		autoscalerReq.IgnoreDaemonsetsUtilization = scw.BoolPtr(ignoreDaemonsetsUtilization.(bool))
	}

	if balanceSimilarNodeGroups, ok := d.GetOk("autoscaler_config.0.balance_similar_node_groups"); ok {
		autoscalerReq.BalanceSimilarNodeGroups = scw.BoolPtr(balanceSimilarNodeGroups.(bool))
	}

	if balanceSimilarNodeGroups, ok := d.GetOk("autoscaler_config.0.balance_similar_node_groups"); ok {
		autoscalerReq.BalanceSimilarNodeGroups = scw.BoolPtr(balanceSimilarNodeGroups.(bool))
	}

	autoscalerReq.ExpendablePodsPriorityCutoff = scw.Int32Ptr(int32(d.Get("autoscaler_config.0.expendable_pods_priority_cutoff").(int)))

	if utilizationThreshold, ok := d.GetOk("autoscaler_config.0.scale_down_utilization_threshold"); ok {
		autoscalerReq.ScaleDownUtilizationThreshold = scw.Float32Ptr(float32(utilizationThreshold.(float64)))
	}

	autoscalerReq.MaxGracefulTerminationSec = scw.Uint32Ptr(uint32(d.Get("autoscaler_config.0.max_graceful_termination_sec").(int)))

	req.AutoscalerConfig = autoscalerReq

	// OpenIDConnect configuration

	createClusterRequestOpenIDConnectConfig := &k8s.CreateClusterRequestOpenIDConnectConfig{}

	if issuerURL, ok := d.GetOk("open_id_connect_config.0.issuer_url"); ok {
		req.OpenIDConnectConfig = createClusterRequestOpenIDConnectConfig
		createClusterRequestOpenIDConnectConfig.IssuerURL = issuerURL.(string)
	}

	if clientID, ok := d.GetOk("open_id_connect_config.0.client_id"); ok {
		req.OpenIDConnectConfig = createClusterRequestOpenIDConnectConfig
		createClusterRequestOpenIDConnectConfig.ClientID = clientID.(string)
	}

	// createClusterRequestOpenIDConnectConfig is always defined here

	if usernameClaim, ok := d.GetOk("open_id_connect_config.0.username_claim"); ok {
		createClusterRequestOpenIDConnectConfig.UsernameClaim = scw.StringPtr(usernameClaim.(string))
	}

	if usernamePrefix, ok := d.GetOk("open_id_connect_config.0.username_prefix"); ok {
		createClusterRequestOpenIDConnectConfig.UsernamePrefix = scw.StringPtr(usernamePrefix.(string))
	}

	if groupsClaim, ok := d.GetOk("open_id_connect_config.0.groups_claim"); ok {
		createClusterRequestOpenIDConnectConfig.GroupsClaim = scw.StringsPtr(types.ExpandStrings(groupsClaim))
	}

	if groupsPrefix, ok := d.GetOk("open_id_connect_config.0.groups_prefix"); ok {
		createClusterRequestOpenIDConnectConfig.GroupsPrefix = scw.StringPtr(groupsPrefix.(string))
	}

	if requiredClaim, ok := d.GetOk("open_id_connect_config.0.required_claim"); ok {
		createClusterRequestOpenIDConnectConfig.RequiredClaim = scw.StringsPtr(types.ExpandStrings(requiredClaim))
	}

	// Auto-upgrade configuration

	autoUpgradeEnable, okAutoUpgradeEnable := d.GetOkExists("auto_upgrade.0.enable")
	autoUpgradeStartHour, okAutoUpgradeStartHour := d.GetOkExists("auto_upgrade.0.maintenance_window_start_hour")
	autoUpgradeDay, okAutoUpgradeDay := d.GetOk("auto_upgrade.0.maintenance_window_day")

	if okAutoUpgradeEnable {
		// check if either all or none of the auto upgrade attribute are set.
		// if one auto upgrade attribute is set, they all must be set.
		// if none is set, auto upgrade attributes will be computed.
		if !okAutoUpgradeDay || !okAutoUpgradeStartHour {
			return append(diag.FromErr(errors.New("all field or zero field of auto_upgrade must be set")), diags...)
		}
	}

	var clusterAutoUpgradeEnabled bool

	if okAutoUpgradeDay && okAutoUpgradeEnable && okAutoUpgradeStartHour {
		clusterAutoUpgradeEnabled = autoUpgradeEnable.(bool)
		req.AutoUpgrade = &k8s.CreateClusterRequestAutoUpgrade{
			Enable: clusterAutoUpgradeEnabled,
			MaintenanceWindow: &k8s.MaintenanceWindow{
				StartHour: uint32(autoUpgradeStartHour.(int)),
				Day:       k8s.MaintenanceWindowDayOfTheWeek(autoUpgradeDay.(string)),
			},
		}
	}

	// K8S Version

	version := d.Get("version").(string)
	versionIsOnlyMinor := len(strings.Split(version, ".")) == 2

	if okAutoUpgradeEnable && autoUpgradeEnable.(bool) && !versionIsOnlyMinor {
		return append(diag.FromErr(errors.New("only minor version x.y can be used with auto upgrade enabled")), diags...)
	}

	if versionIsOnlyMinor && !autoUpgradeEnable.(bool) {
		return append(diag.FromErr(errors.New("minor version x.y must only be used with auto upgrade enabled")), diags...)
	}

	if versionIsOnlyMinor {
		version, err = k8sGetLatestVersionFromMinor(ctx, k8sAPI, region, version)
		if err != nil {
			return append(diag.FromErr(err), diags...)
		}
	}

	req.Version = version

	// Private network configuration

	if pnID, ok := d.GetOk("private_network_id"); ok {
		req.PrivateNetworkID = scw.StringPtr(regional.ExpandID(pnID.(string)).ID)
	}

	// Networking configuration

	if podCIDR, ok := d.GetOk("pod_cidr"); ok {
		podCIDRIPNet, _ := types.ExpandIPNet(podCIDR.(string))
		req.PodCidr = &podCIDRIPNet
	}

	if serviceCIDR, ok := d.GetOk("service_cidr"); ok {
		serviceCIDRIPNet, _ := types.ExpandIPNet(serviceCIDR.(string))
		req.ServiceCidr = &serviceCIDRIPNet
	}

	if serviceDNSIP, ok := d.GetOk("service_dns_ip"); ok {
		serviceDNSIPNetIP := net.ParseIP(serviceDNSIP.(string))
		req.ServiceDNSIP = &serviceDNSIPNetIP
	}

	// Cluster creation

	res, err := k8sAPI.CreateCluster(req, scw.WithContext(ctx))
	if err != nil {
		return append(diag.FromErr(err), diags...)
	}

	d.SetId(regional.NewIDString(region, res.ID))

	if strings.Contains(clusterType.(string), "multicloud") {
		// In case of multi-cloud, we do not have the guarantee that a pool will be created in Scaleway.
		_, err = waitCluster(ctx, k8sAPI, region, res.ID, d.Timeout(schema.TimeoutCreate))
	} else {
		// If we are not in multi-cloud, we can wait for the pool to be created.
		_, err = waitClusterPool(ctx, k8sAPI, region, res.ID, d.Timeout(schema.TimeoutCreate))
	}

	if err != nil {
		return append(diag.FromErr(err), diags...)
	}

	return append(ResourceK8SClusterRead(ctx, d, m), diags...)
}

func ResourceK8SClusterRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	k8sAPI, region, clusterID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	////
	// Read Cluster
	////
	cluster, err := waitCluster(ctx, k8sAPI, region, clusterID, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("region", string(region))
	_ = d.Set("name", cluster.Name)
	_ = d.Set("type", cluster.Type)
	_ = d.Set("organization_id", cluster.OrganizationID)
	_ = d.Set("project_id", cluster.ProjectID)
	_ = d.Set("description", cluster.Description)
	_ = d.Set("cni", cluster.Cni)
	_ = d.Set("tags", cluster.Tags)
	_ = d.Set("apiserver_cert_sans", cluster.ApiserverCertSans)
	_ = d.Set("created_at", cluster.CreatedAt.Format(time.RFC3339))
	_ = d.Set("updated_at", cluster.UpdatedAt.Format(time.RFC3339))
	_ = d.Set("apiserver_url", cluster.ClusterURL)
	_ = d.Set("wildcard_dns", cluster.DNSWildcard)
	_ = d.Set("status", cluster.Status.String())
	_ = d.Set("upgrade_available", cluster.UpgradeAvailable)
	_ = d.Set("feature_gates", cluster.FeatureGates)
	_ = d.Set("admission_plugins", cluster.AdmissionPlugins)

	// if autoupgrade is enabled, we only set the minor k8s version (x.y)
	version := cluster.Version
	if cluster.AutoUpgrade != nil && cluster.AutoUpgrade.Enabled {
		version, err = GetMinorVersionFromFull(version)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	_ = d.Set("version", version)

	// autoscaler_config
	_ = d.Set("autoscaler_config", clusterAutoscalerConfigFlatten(cluster))
	_ = d.Set("open_id_connect_config", clusterOpenIDConnectConfigFlatten(cluster))
	_ = d.Set("auto_upgrade", clusterAutoUpgradeFlatten(cluster))

	// private_network
	_ = d.Set("private_network_id", types.FlattenStringPtr(cluster.PrivateNetworkID))

	// networking
	_ = d.Set("pod_cidr", cluster.PodCidr.String())
	_ = d.Set("service_cidr", cluster.ServiceCidr.String())
	_ = d.Set("service_dns_ip", cluster.ServiceDNSIP.String())

	////
	// Read kubeconfig
	////
	kubeconfig, err := flattenKubeconfig(ctx, k8sAPI, region, clusterID)
	if err != nil {
		if httperrors.Is403(err) {
			return diag.Diagnostics{diag.Diagnostic{
				Severity:      diag.Warning,
				Summary:       "Cannot read kubeconfig: unauthorized",
				Detail:        "Got 403 while reading kubeconfig, please check your permissions",
				AttributePath: cty.GetAttrPath("kubeconfig"),
			}}
		}

		return diag.FromErr(err)
	}

	_ = d.Set("kubeconfig", []map[string]any{kubeconfig})

	return nil
}

//gocyclo:ignore
func ResourceK8SClusterUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	k8sAPI, region, clusterID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("type") {
		_, err = waitCluster(ctx, k8sAPI, region, clusterID, defaultK8SClusterTimeout)
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = k8sAPI.SetClusterType(&k8s.SetClusterTypeRequest{
			Region:    region,
			ClusterID: clusterID,
			Type:      d.Get("type").(string),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		// We have to wait for the pools to reach a stable state too (e.g. being detached from the private network)
		_, err = waitClusterPool(ctx, k8sAPI, region, clusterID, defaultK8SClusterTimeout)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	var diags diag.Diagnostics

	if !strings.HasPrefix(d.Get("type").(string), "kapsule") && !strings.HasPrefix(d.Get("type").(string), "multicloud") {
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Warning,
			Summary:       "Unexpected cluster type",
			Detail:        fmt.Sprintf("The expected cluster type is one of %v, but got %s", []string{"kapsule", "multicloud", "kapsule-dedicated-*", "multicloud-dedicated-*"}, d.Get("type").(string)),
			AttributePath: cty.GetAttrPath("type"),
		})
	}

	canUpgrade := false

	////
	// Construct UpdateClusterRequest
	////
	updateRequest := &k8s.UpdateClusterRequest{
		Region:    region,
		ClusterID: clusterID,
	}

	if d.HasChange("name") {
		updateRequest.Name = types.ExpandStringPtr(d.Get("name"))
	}

	if d.HasChange("description") {
		updateRequest.Description = types.ExpandUpdatedStringPtr(d.Get("description"))
	}

	if d.HasChange("tags") {
		updateRequest.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
	}

	if d.HasChange("apiserver_cert_sans") {
		updateRequest.ApiserverCertSans = types.ExpandUpdatedStringsPtr(d.Get("apiserver_cert_sans"))
	}

	if d.HasChange("feature_gates") {
		updateRequest.FeatureGates = types.ExpandUpdatedStringsPtr(d.Get("feature_gates"))
	}

	if d.HasChange("admission_plugins") {
		updateRequest.AdmissionPlugins = types.ExpandUpdatedStringsPtr(d.Get("admission_plugins"))
	}

	////
	// AutoUpgrade changes
	////
	updateRequest.AutoUpgrade = &k8s.UpdateClusterRequestAutoUpgrade{}
	autoupgradeEnabled := d.Get("auto_upgrade.0.enable").(bool)

	if d.HasChange("auto_upgrade.0.enable") {
		updateRequest.AutoUpgrade.Enable = scw.BoolPtr(d.Get("auto_upgrade.0.enable").(bool))
	}

	if d.HasChanges("auto_upgrade.0.maintenance_window_start_hour", "auto_upgrade.0.maintenance_window_day") {
		updateRequest.AutoUpgrade.MaintenanceWindow = &k8s.MaintenanceWindow{}
		updateRequest.AutoUpgrade.MaintenanceWindow.StartHour = uint32(d.Get("auto_upgrade.0.maintenance_window_start_hour").(int))
		updateRequest.AutoUpgrade.MaintenanceWindow.Day = k8s.MaintenanceWindowDayOfTheWeek(d.Get("auto_upgrade.0.maintenance_window_day").(string))
	}

	////
	// Version changes
	////
	version := d.Get("version").(string)
	versionIsOnlyMinor := len(strings.Split(version, ".")) == 2

	if autoupgradeEnabled && !versionIsOnlyMinor {
		return append(diag.FromErr(errors.New("only minor version x.y can be used with auto upgrade enabled")), diags...)
	}

	if versionIsOnlyMinor && !autoupgradeEnabled {
		return append(diag.FromErr(errors.New("minor version x.y must only be used with auto upgrade enabled")), diags...)
	}

	if versionIsOnlyMinor {
		version, err = k8sGetLatestVersionFromMinor(ctx, k8sAPI, region, version)
		if err != nil {
			return append(diag.FromErr(err), diags...)
		}
	}

	if d.HasChange("version") {
		// maybe it's a change from minor to patch or patch to minor
		// we need to check the current version
		clusterResp, err := k8sAPI.GetCluster(&k8s.GetClusterRequest{
			ClusterID: clusterID,
			Region:    region,
		}, scw.WithContext(ctx))
		if err != nil {
			return append(diag.FromErr(err), diags...)
		}

		if clusterResp.Version == version {
			// no upgrades if same version
			canUpgrade = false
		} else {
			// we let the API decide if we can upgrade
			canUpgrade = true
		}
	}

	////
	// Autoscaler changes
	////
	autoscalerReq := &k8s.UpdateClusterRequestAutoscalerConfig{}

	if d.HasChange("autoscaler_config.0.disable_scale_down") {
		autoscalerReq.ScaleDownDisabled = scw.BoolPtr(d.Get("autoscaler_config.0.disable_scale_down").(bool))
	}

	if d.HasChange("autoscaler_config.0.scale_down_delay_after_add") {
		autoscalerReq.ScaleDownDelayAfterAdd = types.ExpandStringPtr(d.Get("autoscaler_config.0.scale_down_delay_after_add"))
	}

	if d.HasChange("autoscaler_config.0.scale_down_unneeded_time") {
		autoscalerReq.ScaleDownUnneededTime = types.ExpandStringPtr(d.Get("autoscaler_config.0.scale_down_unneeded_time"))
	}

	if d.HasChange("autoscaler_config.0.estimator") {
		autoscalerReq.Estimator = k8s.AutoscalerEstimator(d.Get("autoscaler_config.0.estimator").(string))
	}

	if d.HasChange("autoscaler_config.0.expander") {
		autoscalerReq.Expander = k8s.AutoscalerExpander(d.Get("autoscaler_config.0.expander").(string))
	}

	if d.HasChange("autoscaler_config.0.ignore_daemonsets_utilization") {
		autoscalerReq.IgnoreDaemonsetsUtilization = scw.BoolPtr(d.Get("autoscaler_config.0.ignore_daemonsets_utilization").(bool))
	}

	if d.HasChange("autoscaler_config.0.balance_similar_node_groups") {
		autoscalerReq.BalanceSimilarNodeGroups = scw.BoolPtr(d.Get("autoscaler_config.0.balance_similar_node_groups").(bool))
	}

	if d.HasChange("autoscaler_config.0.expendable_pods_priority_cutoff") {
		autoscalerReq.ExpendablePodsPriorityCutoff = scw.Int32Ptr(int32(d.Get("autoscaler_config.0.expendable_pods_priority_cutoff").(int)))
	}

	if d.HasChange("autoscaler_config.0.scale_down_utilization_threshold") {
		autoscalerReq.ScaleDownUtilizationThreshold = scw.Float32Ptr(float32(d.Get("autoscaler_config.0.scale_down_utilization_threshold").(float64)))
	}

	if d.HasChange("autoscaler_config.0.max_graceful_termination_sec") {
		autoscalerReq.MaxGracefulTerminationSec = scw.Uint32Ptr(uint32(d.Get("autoscaler_config.0.max_graceful_termination_sec").(int)))
	}

	updateRequest.AutoscalerConfig = autoscalerReq

	////
	// OpenIDConnect Config changes
	////
	updateClusterRequestOpenIDConnectConfig := &k8s.UpdateClusterRequestOpenIDConnectConfig{}

	if d.HasChange("open_id_connect_config.0.issuer_url") {
		updateClusterRequestOpenIDConnectConfig.IssuerURL = scw.StringPtr(d.Get("open_id_connect_config.0.issuer_url").(string))
	}

	if d.HasChange("open_id_connect_config.0.client_id") {
		updateClusterRequestOpenIDConnectConfig.ClientID = scw.StringPtr(d.Get("open_id_connect_config.0.client_id").(string))
	}

	if d.HasChange("open_id_connect_config.0.username_claim") {
		updateClusterRequestOpenIDConnectConfig.UsernameClaim = scw.StringPtr(d.Get("open_id_connect_config.0.username_claim").(string))
	}

	if d.HasChange("open_id_connect_config.0.username_prefix") {
		updateClusterRequestOpenIDConnectConfig.UsernamePrefix = scw.StringPtr(d.Get("open_id_connect_config.0.username_prefix").(string))
	}

	if d.HasChange("open_id_connect_config.0.groups_claim") {
		updateClusterRequestOpenIDConnectConfig.GroupsClaim = types.ExpandUpdatedStringsPtr(d.Get("open_id_connect_config.0.groups_claim"))
	}

	if d.HasChange("open_id_connect_config.0.groups_prefix") {
		updateClusterRequestOpenIDConnectConfig.GroupsPrefix = scw.StringPtr(d.Get("open_id_connect_config.0.groups_prefix").(string))
	}

	if d.HasChange("open_id_connect_config.0.required_claim") {
		updateClusterRequestOpenIDConnectConfig.RequiredClaim = types.ExpandUpdatedStringsPtr(d.Get("open_id_connect_config.0.required_claim"))
	}

	updateRequest.OpenIDConnectConfig = updateClusterRequestOpenIDConnectConfig

	////
	// Private Network changes
	////
	if d.HasChange("private_network_id") {
		actual, planned := d.GetChange("private_network_id")
		if planned == "" && actual != "" {
			// It's not possible to remove the private network anymore
			return append(diag.FromErr(errors.New("it is only possible to change the private network attached to the cluster, but not to remove it")), diags...)
		}
	}

	////
	// Apply Update
	////
	_, err = k8sAPI.UpdateCluster(updateRequest, scw.WithContext(ctx))
	if err != nil {
		return append(diag.FromErr(err), diags...)
	}

	_, err = waitCluster(ctx, k8sAPI, region, clusterID, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return append(diag.FromErr(err), diags...)
	}

	////
	// Upgrade if needed
	////
	if canUpgrade {
		upgradeRequest := &k8s.UpgradeClusterRequest{
			Region:       region,
			ClusterID:    clusterID,
			Version:      version,
			UpgradePools: true,
		}

		_, err = k8sAPI.UpgradeCluster(upgradeRequest)
		if err != nil {
			return append(diag.FromErr(err), diags...)
		}

		_, err = waitCluster(ctx, k8sAPI, region, clusterID, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return append(diag.FromErr(err), diags...)
		}

		if !strings.Contains(d.Get("type").(string), "multicloud") {
			// In case of multi-cloud, we do not have the guarantee that a pool will be created in Scaleway.
			// But if we are not, we can wait for the pool to be upgraded.
			_, err = waitClusterPool(ctx, k8sAPI, region, clusterID, d.Timeout(schema.TimeoutUpdate))
			if err != nil {
				return append(diag.FromErr(err), diags...)
			}
		}
	}

	// Display warning if an update of networking fields is requested
	// Setting the value can be done with ForceNew, but unsetting is not possible at the time because Terraform doesn't
	// detect changes on fields that are both optional and computed, and those fields are not updatable with the API.
	for _, key := range []string{"pod_cidr", "service_cidr", "service_dns_ip"} {
		raw, ok := meta.GetRawConfigForKey(d, key, cty.String)
		if !ok || raw == "" && d.Get(key) != NetworkingDefaultValues[key] {
			diags = append(diags, diag.Diagnostic{
				Severity:      diag.Warning,
				Summary:       fmt.Sprintf("It is not possible to unset %q at the time", key),
				Detail:        "Once it has been set to a custom value, unsetting it in order to go back to the default value will not have any effect.",
				AttributePath: cty.GetAttrPath(key),
			})
		}
	}

	return append(ResourceK8SClusterRead(ctx, d, m), diags...)
}

func ResourceK8SClusterDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	k8sAPI, region, clusterID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	deleteAdditionalResources := d.Get("delete_additional_resources").(bool)

	////
	// Delete Cluster
	////
	cluster, err := k8sAPI.DeleteCluster(&k8s.DeleteClusterRequest{
		Region:                  region,
		ClusterID:               clusterID,
		WithAdditionalResources: deleteAdditionalResources,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return nil
		}

		return diag.FromErr(err)
	}

	_, err = waitClusterStatus(ctx, k8sAPI, cluster, k8s.ClusterStatusDeleted, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func autoscalerConfigSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"disable_scale_down": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Disable the scale down feature of the autoscaler",
			},
			"scale_down_delay_after_add": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "10m",
				Description: "How long after scale up that scale down evaluation resumes",
			},
			"scale_down_unneeded_time": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "10m",
				Description: "How long a node should be unneeded before it is eligible for scale down",
			},
			"estimator": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          k8s.AutoscalerEstimatorBinpacking.String(),
				Description:      "Type of resource estimator to be used in scale up",
				ValidateDiagFunc: verify.ValidateEnum[k8s.AutoscalerEstimator](),
			},
			"expander": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          k8s.AutoscalerExpanderRandom.String(),
				Description:      "Type of node group expander to be used in scale up",
				ValidateDiagFunc: verify.ValidateEnum[k8s.AutoscalerExpander](),
			},
			"ignore_daemonsets_utilization": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Ignore DaemonSet pods when calculating resource utilization for scaling down",
			},
			"balance_similar_node_groups": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Detect similar node groups and balance the number of nodes between them",
			},
			"expendable_pods_priority_cutoff": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     -10,
				Description: "Pods with priority below cutoff will be expendable. They can be killed without any consideration during scale down and they don't cause scale up. Pods with null priority (PodPriority disabled) are non expendable",
			},
			"scale_down_utilization_threshold": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Default:     0.5,
				Description: "Node utilization level, defined as sum of requested resources divided by capacity, below which a node can be considered for scale down",
			},
			"max_graceful_termination_sec": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     600,
				Description: "Maximum number of seconds the cluster autoscaler waits for pod termination when trying to scale down a node",
			},
		},
	}
}

func openIDConnectConfigSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"issuer_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "URL of the provider which allows the API server to discover public signing keys",
			},
			"client_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A client id that all tokens must be issued for",
			},
			"username_claim": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "JWT claim to use as the user name",
			},
			"username_prefix": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Prefix prepended to username",
			},
			"groups_claim": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "JWT claim to use as the user's group",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"groups_prefix": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Prefix prepended to group claims",
			},
			"required_claim": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Multiple key=value pairs that describes a required claim in the ID Token",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}
