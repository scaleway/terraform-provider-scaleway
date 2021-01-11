package scaleway

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayK8SCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayK8SClusterCreate,
		ReadContext:   resourceScalewayK8SClusterRead,
		UpdateContext: resourceScalewayK8SClusterUpdate,
		DeleteContext: resourceScalewayK8SClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultK8SClusterTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the cluster",
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
				ValidateFunc: validation.StringInSlice([]string{
					k8s.CNICilium.String(),
					k8s.CNICalico.String(),
					k8s.CNIFlannel.String(),
					k8s.CNIWeave.String(),
				}, false),
			},
			"enable_dashboard": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable the dashboard on the cluster",
			},
			"ingress": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     k8s.IngressNone.String(),
				Description: "The ingress to be deployed on the cluster",
				ValidateFunc: validation.StringInSlice([]string{
					k8s.IngressNone.String(),
					k8s.IngressTraefik.String(),
					k8s.IngressTraefik2.String(),
					k8s.IngressNginx.String(),
				}, false),
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
							ValidateFunc: validateHour(),
						},
						"maintenance_window_day": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Day of the maintenance window",
							ValidateFunc: validation.StringInSlice([]string{
								k8s.MaintenanceWindowDayOfTheWeekAny.String(),
								k8s.MaintenanceWindowDayOfTheWeekMonday.String(),
								k8s.MaintenanceWindowDayOfTheWeekTuesday.String(),
								k8s.MaintenanceWindowDayOfTheWeekWednesday.String(),
								k8s.MaintenanceWindowDayOfTheWeekThursday.String(),
								k8s.MaintenanceWindowDayOfTheWeekFriday.String(),
								k8s.MaintenanceWindowDayOfTheWeekSaturday.String(),
								k8s.MaintenanceWindowDayOfTheWeekSunday.String(),
							}, false),
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
			"delete_additional_resources": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Delete additional resources like block volumes and loadbalancers on cluster deletion",
			},
			"region":          regionSchema(),
			"organization_id": organizationIDSchema(),
			"project_id":      projectIDSchema(),
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
		},
	}
}

func resourceScalewayK8SClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	k8sAPI, region, err := k8sAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	////
	// Create cluster
	////

	description, ok := d.GetOk("description")
	if !ok {
		description = ""
	}

	req := &k8s.CreateClusterRequest{
		Region:           region,
		ProjectID:        expandStringPtr(d.Get("project_id")),
		Name:             expandOrGenerateString(d.Get("name"), "cluster"),
		Description:      description.(string),
		Cni:              k8s.CNI(d.Get("cni").(string)),
		Tags:             expandStrings(d.Get("tags")),
		FeatureGates:     expandStrings(d.Get("feature_gates")),
		AdmissionPlugins: expandStrings(d.Get("admission_plugins")),
	}

	if dashboard, ok := d.GetOk("enable_dashboard"); ok {
		req.EnableDashboard = dashboard.(bool)
	}

	if ingress, ok := d.GetOk("ingress"); ok {
		req.Ingress = k8s.Ingress(ingress.(string))
	}

	autoscalerReq := &k8s.CreateClusterRequestAutoscalerConfig{}

	if scaleDownDisabled, ok := d.GetOk("autoscaler_config.0.disable_scale_down"); ok {
		autoscalerReq.ScaleDownDisabled = scw.BoolPtr(scaleDownDisabled.(bool))
	}

	if scaleDownDelayAfterAdd, ok := d.GetOk("autoscaler_config.0.scale_down_delay_after_add"); ok {
		autoscalerReq.ScaleDownDelayAfterAdd = expandStringPtr(scaleDownDelayAfterAdd)
	}

	if scaleDownUneededTime, ok := d.GetOk("autoscaler_config.0.scale_down_unneeded_time"); ok {
		autoscalerReq.ScaleDownUnneededTime = expandStringPtr(scaleDownUneededTime)
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

	autoscalerReq.ExpendablePodsPriorityCutoff = scw.Int32Ptr(int32(d.Get("autoscaler_config.0.expendable_pods_priority_cutoff").(int)))

	req.AutoscalerConfig = autoscalerReq

	autoUpgradeEnable, okAutoUpgradeEnable := d.GetOk("auto_upgrade.0.enable")
	autoUpgradeStartHour, okAutoUpgradeStartHour := d.GetOk("auto_upgrade.0.maintenance_window_start_hour")
	autoUpgradeDay, okAutoUpgradeDay := d.GetOk("auto_upgrade.0.maintenance_window_day")

	if okAutoUpgradeEnable {
		// check if either all or none of the auto upgrade attribute are set.
		// if one auto upgrade attribute is set, they all must be set.
		// if none is set, auto upgrade attributes will be computed.
		if !(okAutoUpgradeDay && okAutoUpgradeStartHour) {
			return diag.FromErr(fmt.Errorf("all field or zero field of auto_upgrade must be set"))
		}
	}

	clusterAutoUpgradeEnabled := false

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

	version := d.Get("version").(string)
	versionIsOnlyMinor := len(strings.Split(version, ".")) == 2

	if versionIsOnlyMinor != clusterAutoUpgradeEnabled {
		return diag.FromErr(fmt.Errorf("minor version x.y must be used with auto upgrade enabled"))
	}

	if versionIsOnlyMinor {
		version, err = k8sGetLatestVersionFromMinor(ctx, k8sAPI, region, version)
		if err != nil {
			return diag.FromErr(fmt.Errorf("minor version x.y must be used with auto upgrade enabled"))
		}
	}

	req.Version = version

	if _, ok := d.GetOk("default_pool"); ok {
		defaultPoolReq := &k8s.CreateClusterRequestPoolConfig{
			Name:        "default",
			NodeType:    d.Get("default_pool.0.node_type").(string),
			Autoscaling: d.Get("default_pool.0.autoscaling").(bool),
			Autohealing: d.Get("default_pool.0.autohealing").(bool),
			Size:        uint32(d.Get("default_pool.0.size").(int)),
			Tags:        expandStrings(d.Get("default_pool.0.tags")),
		}

		if placementGroupID, ok := d.GetOk("default_pool.0.placement_group_id"); ok {
			defaultPoolReq.PlacementGroupID = expandStringPtr(expandID(placementGroupID))
		}

		if minSize, ok := d.GetOk("default_pool.0.min_size"); ok {
			defaultPoolReq.MinSize = scw.Uint32Ptr(uint32(minSize.(int)))
		}

		if maxSize, ok := d.GetOk("default_pool.0.max_size"); ok {
			defaultPoolReq.MaxSize = scw.Uint32Ptr(uint32(maxSize.(int)))
		}

		if containerRuntime, ok := d.GetOk("default_pool.0.container_runtime"); ok {
			defaultPoolReq.ContainerRuntime = k8s.Runtime(containerRuntime.(string))
		}

		req.Pools = []*k8s.CreateClusterRequestPoolConfig{defaultPoolReq}
	}

	res, err := k8sAPI.CreateCluster(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if _, ok := d.GetOk("default_pool"); ok {
		err = waitK8SCluster(ctx, k8sAPI, region, res.ID, k8s.ClusterStatusReady)
		if err != nil {
			return diag.FromErr(err)
		}
		if d.Get("default_pool.0.wait_for_pool_ready").(bool) { // wait for the pool status to be ready (if specified)
			pool, err := readDefaultPool(ctx, d, m) // ensure that 'default_pool.0.pool_id' is set
			if err != nil {
				return diag.FromErr(err)
			}

			err = waitK8SPoolReady(ctx, k8sAPI, region, expandID(pool.ID))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	} else {
		err = waitK8SCluster(ctx, k8sAPI, region, res.ID, k8s.ClusterStatusPoolRequired)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(newRegionalIDString(region, res.ID))

	return resourceScalewayK8SClusterRead(ctx, d, m)
}

// resourceScalewayK8SClusterDefaultPoolRead is only called after a resourceScalewayK8SClusterCreate
// thus ensuring the uniqueness of the only pool listed
func resourceScalewayK8SClusterDefaultPoolRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	k8sAPI, region, _, err := k8sAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	pool, err := readDefaultPool(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	nodes, err := getNodes(ctx, k8sAPI, pool)
	if err != nil {
		return diag.FromErr(err)
	}

	defaultPool := map[string]interface{}{}
	defaultPool["pool_id"] = newRegionalIDString(region, pool.ID)
	defaultPool["node_type"] = pool.NodeType
	defaultPool["autoscaling"] = pool.Autoscaling
	defaultPool["autohealing"] = pool.Autohealing
	defaultPool["size"] = pool.Size
	defaultPool["min_size"] = pool.MinSize
	defaultPool["max_size"] = pool.MaxSize
	defaultPool["tags"] = pool.Tags
	defaultPool["container_runtime"] = pool.ContainerRuntime
	defaultPool["created_at"] = pool.CreatedAt.String()
	defaultPool["updated_at"] = pool.UpdatedAt.String()
	defaultPool["nodes"] = nodes
	defaultPool["wait_for_pool_ready"] = d.Get("default_pool.0.wait_for_pool_ready")
	defaultPool["status"] = pool.Status.String()

	if pool.PlacementGroupID != nil {
		zone := scw.Zone(region + "-1") // Placement groups are zoned resources.
		defaultPool["placement_group_id"] = newZonedID(zone, *pool.PlacementGroupID)
	}

	err = d.Set("default_pool", []map[string]interface{}{defaultPool})
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func readDefaultPool(ctx context.Context, d *schema.ResourceData, m interface{}) (*k8s.Pool, error) {
	k8sAPI, region, clusterID, err := k8sAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return nil, err
	}

	var pool *k8s.Pool

	if defaultPoolID, ok := d.GetOk("default_pool.0.pool_id"); ok {
		poolResp, err := k8sAPI.GetPool(&k8s.GetPoolRequest{
			Region: region,
			PoolID: expandID(defaultPoolID.(string)),
		}, scw.WithContext(ctx))
		if err != nil {
			return nil, err
		}
		pool = poolResp
	} else {
		response, err := k8sAPI.ListPools(&k8s.ListPoolsRequest{
			Region:    region,
			ClusterID: clusterID,
		}, scw.WithContext(ctx))
		if err != nil {
			return nil, err
		}

		if len(response.Pools) != 1 {
			return nil, fmt.Errorf("newly created pool on cluster %s has %d pools instead of 1", clusterID, len(response.Pools))
		}

		pool = response.Pools[0]
	}
	return pool, nil
}

func resourceScalewayK8SClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	k8sAPI, region, clusterID, err := k8sAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	////
	// Read Cluster
	////
	response, err := k8sAPI.GetCluster(&k8s.GetClusterRequest{
		Region:    region,
		ClusterID: clusterID,
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("region", string(region))
	_ = d.Set("name", response.Name)
	_ = d.Set("organization_id", response.OrganizationID)
	_ = d.Set("project_id", response.ProjectID)
	_ = d.Set("description", response.Description)
	_ = d.Set("cni", response.Cni)
	_ = d.Set("tags", response.Tags)
	_ = d.Set("created_at", response.CreatedAt.Format(time.RFC3339))
	_ = d.Set("updated_at", response.UpdatedAt.Format(time.RFC3339))
	_ = d.Set("apiserver_url", response.ClusterURL)
	_ = d.Set("wildcard_dns", response.DNSWildcard)
	_ = d.Set("status", response.Status.String())
	_ = d.Set("upgrade_available", response.UpgradeAvailable)

	// if autoupgrade is enabled, we only set the minor k8s version (x.y)
	version := response.Version
	if response.AutoUpgrade != nil && response.AutoUpgrade.Enabled {
		version, err = k8sGetMinorVersionFromFull(version)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	_ = d.Set("version", version)

	// autoscaler_config
	_ = d.Set("autoscaler_config", clusterAutoscalerConfigFlatten(response))
	_ = d.Set("auto_upgrade", clusterAutoUpgradeFlatten(response))

	// default_pool_config
	if _, ok := d.GetOk("default_pool"); ok {
		diagnostics := resourceScalewayK8SClusterDefaultPoolRead(ctx, d, m)
		if diagnostics != nil {
			return diagnostics
		}
	}

	////
	// Read kubeconfig
	////
	kubeconfig, err := k8sAPI.GetClusterKubeConfig(&k8s.GetClusterKubeConfigRequest{
		Region:    region,
		ClusterID: clusterID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	kubeconfigServer, err := kubeconfig.GetServer()
	if err != nil {
		return diag.FromErr(err)
	}

	kubeconfigCa, err := kubeconfig.GetCertificateAuthorityData()
	if err != nil {
		return diag.FromErr(err)
	}

	kubeconfigToken, err := kubeconfig.GetToken()
	if err != nil {
		return diag.FromErr(err)
	}

	kubeconf := map[string]interface{}{}
	kubeconf["config_file"] = string(kubeconfig.GetRaw())
	kubeconf["host"] = kubeconfigServer
	kubeconf["cluster_ca_certificate"] = kubeconfigCa
	kubeconf["token"] = kubeconfigToken

	_ = d.Set("kubeconfig", []map[string]interface{}{kubeconf})

	return nil
}

// resourceScalewayK8SClusterDefaultPoolUpdate is only called after a resourceScalewayK8SClusterUpdate
// thus guarating that "default_pool.id" is set
func resourceScalewayK8SClusterDefaultPoolUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	k8sAPI, region, clusterID, err := k8sAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	////
	// Update default Pool
	////
	if d.HasChange("default_pool") {
		defaultPoolID := d.Get("default_pool.0.pool_id").(string)

		forceNew := false
		oldPoolID := ""
		if d.HasChanges("default_pool.0.container_runtime", "default_pool.0.node_type", "default_pool.0.placement_group_id") {
			forceNew = true
			oldPoolID = defaultPoolID
		} else {
			updateRequest := &k8s.UpdatePoolRequest{
				Region: region,
				PoolID: expandID(defaultPoolID),
				Tags:   scw.StringsPtr(expandStrings(d.Get("default_pool.0.tags"))),
			}

			if autohealing, ok := d.GetOk("default_pool.0.autohealing"); ok {
				updateRequest.Autohealing = scw.BoolPtr(autohealing.(bool))
			}

			if d.HasChange("default_pool.0.min_size") {
				updateRequest.MinSize = scw.Uint32Ptr(uint32(d.Get("default_pool.0.min_size").(int)))
			}

			if d.HasChange("default_pool.0.max_size") {
				updateRequest.MaxSize = scw.Uint32Ptr(uint32(d.Get("default_pool.0.max_size").(int)))
			}

			if autoscaling, ok := d.GetOk("default_pool.0.autoscaling"); ok {
				updateRequest.Autoscaling = scw.BoolPtr(autoscaling.(bool))
			}

			if !d.Get("default_pool.0.autoscaling").(bool) {
				if size, ok := d.GetOk("default_pool.0.size"); ok {
					updateRequest.Size = scw.Uint32Ptr(uint32(size.(int)))
				}
			}

			_, err := k8sAPI.UpdatePool(updateRequest, scw.WithContext(ctx))
			if err != nil {
				if !is404Error(err) {
					return diag.FromErr(err)
				}
				l.Warningf("default node pool %s is not found, recreating a new one", defaultPoolID)
				forceNew = true
			}
		}

		if forceNew {
			defaultPoolRequest := &k8s.CreatePoolRequest{
				Region:      region,
				ClusterID:   clusterID,
				Name:        "default",
				NodeType:    d.Get("default_pool.0.node_type").(string),
				Autoscaling: d.Get("default_pool.0.autoscaling").(bool),
				Autohealing: d.Get("default_pool.0.autohealing").(bool),
				Size:        uint32(d.Get("default_pool.0.size").(int)),
			}
			if placementGroupID, ok := d.GetOk("default_pool.0.placement_group_id"); ok {
				defaultPoolRequest.PlacementGroupID = expandStringPtr(expandID(placementGroupID))
			}

			if d.HasChange("default_pool.0.min_size") {
				defaultPoolRequest.MinSize = scw.Uint32Ptr(uint32(d.Get("default_pool.0.min_size").(int)))
			}

			if d.HasChange("default_pool.0.max_size") {
				defaultPoolRequest.MaxSize = scw.Uint32Ptr(uint32(d.Get("default_pool.0.max_size").(int)))
			}

			if containerRuntime, ok := d.GetOk("default_pool.0.container_runtime"); ok {
				defaultPoolRequest.ContainerRuntime = k8s.Runtime(containerRuntime.(string))
			}

			defaultPoolRes, err := k8sAPI.CreatePool(defaultPoolRequest, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}
			defaultPoolID = newRegionalIDString(region, defaultPoolRes.ID)
			defaultPool := map[string]interface{}{}
			defaultPool["pool_id"] = defaultPoolID

			_ = d.Set("default_pool", []map[string]interface{}{defaultPool})

			if oldPoolID != "" {
				// wait for new pool to be ready before deleting old one
				err = waitK8SPoolReady(ctx, k8sAPI, region, expandID(defaultPoolID))
				if err != nil {
					return diag.FromErr(err)
				}

				_, err = k8sAPI.DeletePool(&k8s.DeletePoolRequest{
					Region: region,
					PoolID: expandID(oldPoolID),
				}, scw.WithContext(ctx))
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}

		if d.Get("default_pool.0.wait_for_pool_ready").(bool) { // wait for the pool to be ready if specified
			err = waitK8SPoolReady(ctx, k8sAPI, region, expandID(defaultPoolID))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return resourceScalewayK8SClusterDefaultPoolRead(ctx, d, m)
}

func resourceScalewayK8SClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	k8sAPI, region, clusterID, err := k8sAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
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
		updateRequest.Name = expandStringPtr(d.Get("name"))
	}

	if d.HasChange("description") {
		updateRequest.Description = expandStringPtr(d.Get("description"))
	}

	if d.HasChange("tags") {
		tags := expandStrings(d.Get("tags"))
		updateRequest.Tags = scw.StringsPtr(tags)
	}

	if d.HasChange("feature_gates") {
		updateRequest.FeatureGates = scw.StringsPtr(expandStrings(d.Get("feature_gates")))
	}

	if d.HasChange("admission_plugins") {
		updateRequest.AdmissionPlugins = scw.StringsPtr(expandStrings(d.Get("admission_plugins")))
	}

	if d.HasChange("ingress") {
		updateRequest.Ingress = k8s.Ingress(d.Get("ingress").(string))
	}

	if d.HasChange("enable_dashboard") {
		updateRequest.EnableDashboard = scw.BoolPtr(d.Get("enable_dashboard").(bool))
	}

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

	version := d.Get("version").(string)
	versionIsOnlyMinor := len(strings.Split(version, ".")) == 2

	if versionIsOnlyMinor != autoupgradeEnabled {
		return diag.FromErr(fmt.Errorf("minor version x.y must be used with auto upgrades enabled"))
	}

	if versionIsOnlyMinor {
		version, err = k8sGetLatestVersionFromMinor(ctx, k8sAPI, region, version)
		if err != nil {
			return diag.FromErr(err)
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
			return diag.FromErr(err)
		}

		if clusterResp.Version == version {
			// no upgrades if same version
			canUpgrade = false
		} else {
			// we let the API decide if we can upgrade
			canUpgrade = true
		}
	}

	autoscalerReq := &k8s.UpdateClusterRequestAutoscalerConfig{}

	if d.HasChange("autoscaler_config.0.disable_scale_down") {
		autoscalerReq.ScaleDownDisabled = scw.BoolPtr(d.Get("autoscaler_config.0.disable_scale_down").(bool))
	}

	if d.HasChange("autoscaler_config.0.scale_down_delay_after_add") {
		autoscalerReq.ScaleDownDelayAfterAdd = expandStringPtr(d.Get("autoscaler_config.0.scale_down_delay_after_add"))
	}

	if d.HasChange("autoscaler_config.0.scale_down_unneeded_time") {
		autoscalerReq.ScaleDownUnneededTime = expandStringPtr(d.Get("autoscaler_config.0.scale_down_unneeded_time"))
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

	updateRequest.AutoscalerConfig = autoscalerReq

	////
	// Apply Update
	////
	_, err = k8sAPI.UpdateCluster(updateRequest, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	err = waitK8SCluster(ctx, k8sAPI, region, clusterID, k8s.ClusterStatusReady, k8s.ClusterStatusPoolRequired)
	if err != nil {
		return diag.FromErr(err)
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
			return diag.FromErr(err)
		}

		err = waitK8SCluster(ctx, k8sAPI, region, clusterID, k8s.ClusterStatusReady, k8s.ClusterStatusPoolRequired)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if _, ok := d.GetOk("default_pool"); ok {
		diagnostics := resourceScalewayK8SClusterDefaultPoolUpdate(ctx, d, m)
		if diagnostics != nil {
			return diagnostics
		}
	}

	return resourceScalewayK8SClusterRead(ctx, d, m)
}

func resourceScalewayK8SClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	k8sAPI, region, clusterID, err := k8sAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	deleteAdditionalResources := d.Get("delete_additional_resources").(bool)

	////
	// Delete Cluster
	////
	_, err = k8sAPI.DeleteCluster(&k8s.DeleteClusterRequest{
		Region:                  region,
		ClusterID:               clusterID,
		WithAdditionalResources: deleteAdditionalResources,
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			return nil
		}
		return diag.FromErr(err)
	}

	err = waitK8SClusterDeleted(ctx, k8sAPI, region, clusterID)
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
				Type:        schema.TypeString,
				Optional:    true,
				Default:     k8s.AutoscalerEstimatorBinpacking.String(),
				Description: "Type of resource estimator to be used in scale up",
				ValidateFunc: validation.StringInSlice([]string{
					k8s.AutoscalerEstimatorBinpacking.String(),
				}, false),
			},
			"expander": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     k8s.AutoscalerExpanderRandom.String(),
				Description: "Type of node group expander to be used in scale up",
				ValidateFunc: validation.StringInSlice([]string{
					k8s.AutoscalerExpanderRandom.String(),
					k8s.AutoscalerExpanderLeastWaste.String(),
					k8s.AutoscalerExpanderMostPods.String(),
					k8s.AutoscalerExpanderPriority.String(),
				}, false),
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
		},
	}
}
