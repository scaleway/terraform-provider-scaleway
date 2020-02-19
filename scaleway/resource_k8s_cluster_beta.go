package scaleway

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	k8s "github.com/scaleway/scaleway-sdk-go/api/k8s/v1beta4"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayK8SClusterBeta() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayK8SClusterBetaCreate,
		Read:   resourceScalewayK8SClusterBetaRead,
		Update: resourceScalewayK8SClusterBetaUpdate,
		Delete: resourceScalewayK8SClusterBetaDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
			"default_pool": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Description: "Default pool created for the cluster on creation",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"node_type": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "Server type of the default pool servers",
							DiffSuppressFunc: diffSuppressFuncIgnoreCaseAndHyphen,
						},
						"autoscaling": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Enable the autoscaling on the default pool",
						},
						"autohealing": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Enable the autohealing on the default pool",
						},
						"size": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Size of the default pool",
						},
						"min_size": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     1,
							Description: "Minimun size of the default pool",
						},
						"max_size": {
							Type:        schema.TypeInt,
							Optional:    true,
							Computed:    true,
							Default:     nil,
							Description: "Maximum size of the default pool",
						},
						"placement_group_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     nil,
							Description: "ID of the placement group for the default pool",
						},
						"container_runtime": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     k8s.RuntimeDocker.String(),
							Description: "Container runtime for the default pool",
							ValidateFunc: validation.StringInSlice([]string{
								k8s.RuntimeDocker.String(),
								k8s.RuntimeContainerd.String(),
								k8s.RuntimeCrio.String(),
							}, false),
						},
						"wait_for_pool_ready": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether to wait for the pool to be ready",
						},
						// Computed elements
						"pool_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of default pool",
						},
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The date and time of the creation of the default pool",
						},
						"updated_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The date and time of the last update of the default pool",
						},
						"nodes": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The name of the node",
									},
									"status": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The status of the node",
									},
									"public_ip": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The public IPv4 address of the node",
									},
									"public_ip_v6": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The public IPv6 address of the node",
									},
								},
							},
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The status of the default pool",
						},
					},
				},
			},
			"region":          regionSchema(),
			"organization_id": organizationIDSchema(),
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
				MaxItems:    1,
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

func resourceScalewayK8SClusterBetaCreate(d *schema.ResourceData, m interface{}) error {
	k8sAPI, region, err := k8sAPIWithRegion(d, m)
	if err != nil {
		return err
	}

	////
	// Create cluster
	////

	description, ok := d.GetOk("description")
	if !ok {
		description = ""
	}

	version, ok := d.GetOk("version")
	if !ok {
		version = ""
	}

	req := &k8s.CreateClusterRequest{
		Region:           region,
		OrganizationID:   d.Get("organization_id").(string),
		Name:             expandOrGenerateString(d.Get("name"), "cluster"),
		Description:      description.(string),
		Version:          version.(string),
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
		autoscalerReq.ScaleDownDisabled = scw.BoolPtr((scaleDownDisabled.(bool)))
	}

	if scaleDownDelayAfterAdd, ok := d.GetOk("autoscaler_config.0.scale_down_delay_after_add"); ok {
		autoscalerReq.ScaleDownDelayAfterAdd = scw.StringPtr(scaleDownDelayAfterAdd.(string))
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

	// check if either all or none of the auto upgrade attribute are set.
	// if one auto upgrade attribute is set, they all must be set.
	// if none is set, auto upgrade attributes will be computed.
	if okAutoUpgradeEnable != okAutoUpgradeDay || okAutoUpgradeEnable != okAutoUpgradeStartHour {
		return fmt.Errorf("all field or zero field of auto_upgrade must be set")
	}

	if okAutoUpgradeDay && okAutoUpgradeEnable && okAutoUpgradeStartHour {
		req.AutoUpgrade = &k8s.CreateClusterRequestAutoUpgrade{
			Enable: autoUpgradeEnable.(bool),
			MaintenanceWindow: &k8s.MaintenanceWindow{
				StartHour: uint32(autoUpgradeStartHour.(int)),
				Day:       k8s.MaintenanceWindowDayOfTheWeek(autoUpgradeDay.(string)),
			},
		}
	}

	defaultPoolReq := &k8s.CreateClusterRequestDefaultPoolConfig{
		NodeType:    d.Get("default_pool.0.node_type").(string),
		Autoscaling: d.Get("default_pool.0.autoscaling").(bool),
		Autohealing: d.Get("default_pool.0.autohealing").(bool),
		Size:        uint32(d.Get("default_pool.0.size").(int)),
	}

	if placementGroupID, ok := d.GetOk("default_pool.0.placement_group_id"); ok {
		defaultPoolReq.PlacementGroupID = scw.StringPtr(expandID(placementGroupID.(string)))
	}

	defaultPoolReq.MinSize = scw.Uint32Ptr(uint32(d.Get("default_pool.0.min_size").(int)))

	if maxSize, ok := d.GetOk("default_pool.0.max_size"); ok {
		defaultPoolReq.MaxSize = scw.Uint32Ptr(uint32(maxSize.(int)))
	} else {
		defaultPoolReq.MaxSize = scw.Uint32Ptr(defaultPoolReq.Size)
	}

	if containerRuntime, ok := d.GetOk("default_pool.0.container_runtime"); ok {
		defaultPoolReq.ContainerRuntime = k8s.Runtime(containerRuntime.(string))
	}

	req.DefaultPoolConfig = defaultPoolReq

	res, err := k8sAPI.CreateCluster(req)
	if err != nil {
		return err
	}

	d.SetId(newRegionalId(region, res.ID))

	err = waitK8SClusterReady(k8sAPI, region, res.ID) // wait for the cluster status to be ready
	if err != nil {
		return err
	}

	if d.Get("default_pool.0.wait_for_pool_ready").(bool) { // wait for the pool status to be ready (if specified)
		pool, err := readDefaultPool(d, m) // ensure that 'default_pool.0.pool_id' is set
		if err != nil {
			return err
		}

		err = waitK8SPoolReady(k8sAPI, region, expandID(pool.ID))
		if err != nil {
			return err
		}
	}

	return resourceScalewayK8SClusterBetaRead(d, m)
}

// resourceScalewayK8SClusterBetaDefaultPoolRead is only called after a resourceScalewayK8SClusterBetaCreate
// thus ensuring the uniqueness of the only pool listed
func resourceScalewayK8SClusterBetaDefaultPoolRead(d *schema.ResourceData, m interface{}) error {
	k8sAPI, region, _, err := k8sAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	pool, err := readDefaultPool(d, m)
	if err != nil {
		return err
	}

	nodes, err := getNodes(k8sAPI, pool)
	if err != nil {
		return err
	}

	defaultPool := map[string]interface{}{}
	defaultPool["pool_id"] = newRegionalId(region, pool.ID)
	defaultPool["node_type"] = pool.NodeType
	defaultPool["autoscaling"] = pool.Autoscaling
	defaultPool["autohealing"] = pool.Autohealing
	defaultPool["size"] = pool.Size
	defaultPool["min_size"] = pool.MinSize
	defaultPool["max_size"] = pool.MaxSize
	defaultPool["container_runtime"] = pool.ContainerRuntime
	defaultPool["created_at"] = pool.CreatedAt.String()
	defaultPool["updated_at"] = pool.UpdatedAt.String()
	defaultPool["nodes"] = nodes
	defaultPool["wait_for_pool_ready"] = d.Get("default_pool.0.wait_for_pool_ready")
	defaultPool["status"] = pool.Status.String()

	if pool.PlacementGroupID != nil {
		defaultPool["placement_group_id"] = newZonedIdFromRegion(region, *pool.PlacementGroupID) // TODO fix this ZonedIdFromRegion
	}

	err = d.Set("default_pool", []map[string]interface{}{defaultPool})
	if err != nil {
		return err
	}
	return nil
}

func readDefaultPool(d *schema.ResourceData, m interface{}) (*k8s.Pool, error) {
	k8sAPI, region, clusterID, err := k8sAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return nil, err
	}

	var pool *k8s.Pool

	if defaultPoolID, ok := d.GetOk("default_pool.0.pool_id"); ok {
		poolResp, err := k8sAPI.GetPool(&k8s.GetPoolRequest{
			Region: region,
			PoolID: expandID(defaultPoolID.(string)),
		})
		if err != nil {
			return nil, err
		}
		pool = poolResp
	} else {
		response, err := k8sAPI.ListPools(&k8s.ListPoolsRequest{
			Region:    region,
			ClusterID: clusterID,
		})
		if err != nil {
			return nil, err
		}

		if len(response.Pools) != 1 {
			return nil, fmt.Errorf("Newly created pool on cluster %s has %d pools instead of 1", clusterID, len(response.Pools))
		}

		pool = response.Pools[0]
	}
	return pool, nil
}

func resourceScalewayK8SClusterBetaRead(d *schema.ResourceData, m interface{}) error {
	k8sAPI, region, clusterID, err := k8sAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	////
	// Read Cluster
	////
	response, err := k8sAPI.GetCluster(&k8s.GetClusterRequest{
		Region:    region,
		ClusterID: clusterID,
	})
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return err
	}

	_ = d.Set("region", string(region))
	_ = d.Set("name", response.Name)
	_ = d.Set("description", response.Description)
	_ = d.Set("version", response.Version)
	_ = d.Set("cni", response.Cni)
	_ = d.Set("tags", response.Tags)
	_ = d.Set("created_at", response.CreatedAt)
	_ = d.Set("updated_at", response.UpdatedAt)
	_ = d.Set("apiserver_url", response.ClusterURL)
	_ = d.Set("wildcard_dns", response.DNSWildcard)
	_ = d.Set("status", response.Status.String())
	_ = d.Set("upgrade_available", response.UpgradeAvailable)

	// autoscaler_config
	_ = d.Set("autoscaler_config", clusterAutoscalerConfigFlatten(response))
	_ = d.Set("auto_upgrade", clusterAutoUpgradeFlatten(response))

	// default_pool_config
	err = resourceScalewayK8SClusterBetaDefaultPoolRead(d, m)
	if err != nil {
		return err
	}

	////
	// Read kubeconfig
	////
	kubeconfig, err := k8sAPI.GetClusterKubeConfig(&k8s.GetClusterKubeConfigRequest{
		Region:    region,
		ClusterID: clusterID,
	})
	if err != nil {
		return err
	}

	kubeconfigServer, err := kubeconfig.GetServer()
	if err != nil {
		return err
	}

	kubeconfigCa, err := kubeconfig.GetCertificateAuthorityData()
	if err != nil {
		return err
	}

	kubeconfigToken, err := kubeconfig.GetToken()
	if err != nil {
		return err
	}

	kubeconf := map[string]interface{}{}
	kubeconf["config_file"] = string(kubeconfig.GetRaw())
	kubeconf["host"] = kubeconfigServer
	kubeconf["cluster_ca_certificate"] = kubeconfigCa
	kubeconf["token"] = kubeconfigToken

	_ = d.Set("kubeconfig", []map[string]interface{}{kubeconf})

	return nil
}

// resourceScalewayK8SClusterBetaDefaultPoolUpdate is only called after a resourceScalewayK8SClusterBetaUpdate
// thus guarating that "default_pool.id" is set
func resourceScalewayK8SClusterBetaDefaultPoolUpdate(d *schema.ResourceData, m interface{}) error {
	k8sAPI, region, clusterID, err := k8sAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	////
	// Update default Pool
	////
	if d.HasChange("default_pool") {
		defaultPoolID := d.Get("default_pool.0.pool_id").(string)

		forceNew := false
		oldPoolID := ""
		if d.HasChange("default_pool.0.container_runtime") || d.HasChange("default_pool.0.node_type") || d.HasChange("default_pool.0.placement_group_id") {
			forceNew = true
			oldPoolID = defaultPoolID
		} else {
			updateRequest := &k8s.UpdatePoolRequest{
				Region: region,
				PoolID: expandID(defaultPoolID),
			}

			if autohealing, ok := d.GetOk("default_pool.0.autohealing"); ok {
				updateRequest.Autohealing = scw.BoolPtr(autohealing.(bool))
			}

			if minSize, ok := d.GetOk("default_pool.0.min_size"); ok {
				updateRequest.MinSize = scw.Uint32Ptr(uint32(minSize.(int)))
			}

			if maxSize, ok := d.GetOk("default_pool.0.max_size"); ok {
				updateRequest.MaxSize = scw.Uint32Ptr(uint32(maxSize.(int)))
			}

			if autoscaling, ok := d.GetOk("default_pool.0.autoscaling"); ok {
				updateRequest.Autoscaling = scw.BoolPtr(autoscaling.(bool))
			}

			if !d.Get("default_pool.0.autoscaling").(bool) {
				if size, ok := d.GetOk("default_pool.0.size"); ok {
					updateRequest.Size = scw.Uint32Ptr(uint32(size.(int)))
				}
			}

			_, err := k8sAPI.UpdatePool(updateRequest)
			if err != nil {
				if !is404Error(err) {
					return err
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
				defaultPoolRequest.PlacementGroupID = scw.StringPtr(expandID(placementGroupID.(string)))
			}

			if minSize, ok := d.GetOk("default_pool.0.min_size"); ok {
				defaultPoolRequest.MinSize = scw.Uint32Ptr(uint32(minSize.(int)))
			}

			if maxSize, ok := d.GetOk("default_pool.0.max_size"); ok {
				defaultPoolRequest.MaxSize = scw.Uint32Ptr(uint32(maxSize.(int)))
			}

			if containerRuntime, ok := d.GetOk("default_pool.0.container_runtime"); ok {
				defaultPoolRequest.ContainerRuntime = k8s.Runtime(containerRuntime.(string))
			}

			defaultPoolRes, err := k8sAPI.CreatePool(defaultPoolRequest)
			if err != nil {
				return err
			}
			defaultPool := map[string]interface{}{}
			defaultPool["pool_id"] = newRegionalId(region, defaultPoolRes.ID)

			_ = d.Set("default_pool", []map[string]interface{}{defaultPool})

			if oldPoolID != "" {
				_, err = k8sAPI.DeletePool(&k8s.DeletePoolRequest{
					Region: region,
					PoolID: expandID(oldPoolID),
				})
				if err != nil {
					return err
				}
			}
		}

		if d.Get("default_pool.0.wait_for_pool_ready").(bool) { // wait for the pool to be ready if specified
			err = waitK8SPoolReady(k8sAPI, region, expandID(defaultPoolID))
			if err != nil {
				return err
			}
		}
	}

	return resourceScalewayK8SClusterBetaDefaultPoolRead(d, m)
}

func resourceScalewayK8SClusterBetaUpdate(d *schema.ResourceData, m interface{}) error {
	k8sAPI, region, clusterID, err := k8sAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
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
		updateRequest.Name = scw.StringPtr(d.Get("name").(string))
	}

	if d.HasChange("description") {
		updateRequest.Description = scw.StringPtr(d.Get("description").(string))
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

	if d.HasChange("version") {
		versions, err := k8sAPI.ListClusterAvailableVersions(&k8s.ListClusterAvailableVersionsRequest{
			Region:    region,
			ClusterID: clusterID,
		})
		if err != nil {
			return err
		}

		for _, version := range versions.Versions {
			if version.Name == d.Get("version").(string) {
				canUpgrade = true
				break
			}
		}
		if !canUpgrade {
			return fmt.Errorf("cluster %s can not be upgraded to version %s", clusterID, d.Get("version").(string))
		}
	}

	if d.HasChange("ingress") {
		updateRequest.Ingress = k8s.Ingress(d.Get("ingress").(string))
	}

	if d.HasChange("enable_dashboard") {
		updateRequest.EnableDashboard = scw.BoolPtr(d.Get("enable_dashboard").(bool))
	}

	updateRequest.AutoUpgrade = &k8s.UpdateClusterRequestAutoUpgrade{}

	if d.HasChange("auto_upgrade.0.enable") {
		updateRequest.AutoUpgrade.Enable = scw.BoolPtr(d.Get("auto_upgrade.0.enable").(bool))
	}
	updateRequest.AutoUpgrade.MaintenanceWindow = &k8s.MaintenanceWindow{}
	if d.HasChange("auto_upgrade.0.maintenance_window_start_hour") {
		updateRequest.AutoUpgrade.MaintenanceWindow.StartHour = uint32(d.Get("auto_upgrade.0.maintenance_window_start_hour").(int))
	}
	if d.HasChange("auto_upgrade.0.maintenance_window_day") {
		updateRequest.AutoUpgrade.MaintenanceWindow.Day = k8s.MaintenanceWindowDayOfTheWeek(d.Get("auto_upgrade.0.maintenance_window_day").(string))
	}

	////
	// Apply Update
	////
	_, err = k8sAPI.UpdateCluster(updateRequest)
	if err != nil {
		return err
	}

	err = waitK8SClusterReady(k8sAPI, region, clusterID)
	if err != nil {
		return err
	}

	////
	// Upgrade if needed
	////
	if canUpgrade {
		upgradeRequest := &k8s.UpgradeClusterRequest{
			Region:       region,
			ClusterID:    clusterID,
			Version:      d.Get("version").(string),
			UpgradePools: true,
		}
		_, err = k8sAPI.UpgradeCluster(upgradeRequest)
		if err != nil {
			return err
		}

		err = waitK8SClusterReady(k8sAPI, region, clusterID)
		if err != nil {
			return err
		}
	}

	err = resourceScalewayK8SClusterBetaDefaultPoolUpdate(d, m)
	if err != nil {
		return err
	}

	return resourceScalewayK8SClusterBetaRead(d, m)
}

func resourceScalewayK8SClusterBetaDelete(d *schema.ResourceData, m interface{}) error {
	k8sAPI, region, clusterID, err := k8sAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	////
	// Delete Cluster
	////
	_, err = k8sAPI.DeleteCluster(&k8s.DeleteClusterRequest{
		Region:    region,
		ClusterID: clusterID,
	})
	if err != nil {
		if is404Error(err) {
			return nil
		}
		return err
	}

	err = waitK8SClusterDeleted(k8sAPI, region, clusterID)
	if err != nil {
		return err
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
			"estimator": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     k8s.AutoscalerEstimatorBinpacking.String(),
				Description: "Type of resource estimator to be used in scale up",
				ValidateFunc: validation.StringInSlice([]string{
					k8s.AutoscalerEstimatorBinpacking.String(),
					k8s.AutoscalerEstimatorOldbinpacking.String(),
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
