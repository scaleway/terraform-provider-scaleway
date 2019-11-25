package scaleway

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	k8s "github.com/scaleway/scaleway-sdk-go/api/k8s/v1beta3"
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
				ForceNew:    true,
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
				Computed:    true,
				Optional:    true,
				Default:     nil,
				Description: "The version of the cluster",
			},
			"cni": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The CNI plugin of the cluster",
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
				Default:     "no_ingress",
				Description: "The ingress to be deployed on the cluster",
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
							ForceNew:         true,
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
							ForceNew:    true,
							Default:     nil,
							Description: "ID of the placement group for the default pool",
						},
						"container_runtime": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Default:     "docker",
							Description: "Container runtime for the default pool",
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
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the cluster",
			},
		},
	}
}

func resourceScalewayK8SClusterBetaCreate(d *schema.ResourceData, m interface{}) error {
	k8sAPI, region, err := getK8SAPIWithRegion(d, m)
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
		Region:         region,
		OrganizationID: d.Get("organization_id").(string),
		Name:           expandOrGenerateString(d.Get("name"), "cluster"),
		Description:    description.(string),
		Version:        version.(string),
		Cni:            d.Get("cni").(string),
		Tags:           expandStrings(d.Get("tags")),
	}

	if dashboard, ok := d.GetOk("enable_dashboard"); ok {
		req.EnableDashboard = dashboard.(bool)
	}

	if ingress, ok := d.GetOk("ingress"); ok {
		req.Ingress = ingress.(string)
	}

	autoscalerReq := &k8s.CreateClusterRequestAutoscalerConfig{}

	if scaleDownDisabled, ok := d.GetOk("autoscaler_config.0.disable_scale_down"); ok {
		autoscalerReq.ScaleDownDisabled = scw.BoolPtr((scaleDownDisabled.(bool)))
	}

	if scaleDownDelayAfterAdd, ok := d.GetOk("autoscaler_config.0.scale_down_delay_after_add"); ok {
		autoscalerReq.ScaleDownDelayAfterAdd = scw.StringPtr(scaleDownDelayAfterAdd.(string))
	}

	if estimator, ok := d.GetOk("autoscaler_config.0.estimator"); ok {
		autoscalerReq.Estimator = scw.StringPtr(estimator.(string))
	}

	if expander, ok := d.GetOk("autoscaler_config.0.expander"); ok {
		autoscalerReq.Expander = scw.StringPtr(expander.(string))
	}

	if ignoreDaemonsetsUtilization, ok := d.GetOk("autoscaler_config.0.ignore_daemonsets_utilization"); ok {
		autoscalerReq.IgnoreDaemonsetsUtilization = scw.BoolPtr(ignoreDaemonsetsUtilization.(bool))
	}

	if balanceSimilarNodeGroups, ok := d.GetOk("autoscaler_config.0.balance_similar_node_groups"); ok {
		autoscalerReq.BalanceSimilarNodeGroups = scw.BoolPtr(balanceSimilarNodeGroups.(bool))
	}

	autoscalerReq.ExpendablePodsPriorityCutoff = scw.Int32Ptr(int32(d.Get("autoscaler_config.0.expendable_pods_priority_cutoff").(int)))

	req.AutoscalerConfig = autoscalerReq

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
		defaultPoolReq.ContainerRuntime = scw.StringPtr(containerRuntime.(string))
	}

	req.DefaultPoolConfig = defaultPoolReq

	res, err := k8sAPI.CreateCluster(req)
	if err != nil {
		return err
	}

	d.SetId(newRegionalId(region, res.ID))

	err = waitK8SClusterReady(k8sAPI, region, res.ID)
	if err != nil {
		return err
	}

	return resourceScalewayK8SClusterBetaRead(d, m)
}

// resourceScalewayK8SClusterBetaDefaultPoolRead is only called after a resourceScalewayK8SClusterBetaCreate
// thus ensuring the uniqueness of the only pool listed
func resourceScalewayK8SClusterBetaDefaultPoolRead(d *schema.ResourceData, m interface{}) error {
	k8sAPI, region, clusterID, err := getK8SAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	////
	// Read default Pool
	////

	var pool *k8s.Pool

	if defaultPoolID, ok := d.GetOk("default_pool.0.pool_id"); ok {
		poolResp, err := k8sAPI.GetPool(&k8s.GetPoolRequest{
			Region: region,
			PoolID: expandID(defaultPoolID.(string)),
		})
		if err != nil {
			return err
		}
		pool = poolResp
	} else {
		response, err := k8sAPI.ListPools(&k8s.ListPoolsRequest{
			Region:    region,
			ClusterID: clusterID,
		})
		if err != nil {
			return err
		}

		if len(response.Pools) != 1 {
			return fmt.Errorf("Newly created pool on cluster %s has %d pools instead of 1", clusterID, len(response.Pools))
		}

		pool = response.Pools[0]
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

func resourceScalewayK8SClusterBetaRead(d *schema.ResourceData, m interface{}) error {
	k8sAPI, region, clusterID, err := getK8SAPIWithRegionAndID(m, d.Id())
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

	d.Set("region", string(region))
	d.Set("name", response.Name)
	d.Set("description", response.Description)
	d.Set("version", response.Version)
	d.Set("cni", response.Cni)
	d.Set("tags", response.Tags)
	d.Set("created_at", response.CreatedAt)
	d.Set("updated_at", response.UpdatedAt)
	d.Set("apiserver_url", response.ClusterURL)
	d.Set("wildcard_dns", response.DNSWildcard)
	d.Set("status", response.Status.String())

	// autoscaler_config
	d.Set("autoscaler_config", []map[string]interface{}{clusterAutoscalerConfigFlatten(response)})

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

	d.Set("kubeconfig", []map[string]interface{}{kubeconf})

	return nil
}

// resourceScalewayK8SClusterBetaDefaultPoolUpdate is only called after a resourceScalewayK8SClusterBetaUpdate
// thus guarating that "default_pool.id" is set
func resourceScalewayK8SClusterBetaDefaultPoolUpdate(d *schema.ResourceData, m interface{}) error {
	k8sAPI, region, clusterID, err := getK8SAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	////
	// Update default Pool
	////
	if d.HasChange("default_pool") {
		defaultPoolID := d.Get("default_pool.0.pool_id").(string)

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

		if d.Get("default_pool.0.autoscaling").(bool) == false {
			if size, ok := d.GetOk("default_pool.0.size"); ok {
				updateRequest.Size = scw.Uint32Ptr(uint32(size.(int)))
			}
		}

		_, err := k8sAPI.UpdatePool(updateRequest)
		if err != nil {
			if !is404Error(err) {
				return err
			}
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
				defaultPoolRequest.ContainerRuntime = scw.StringPtr(containerRuntime.(string))
			}

			defaultPoolRes, err := k8sAPI.CreatePool(defaultPoolRequest)
			if err != nil {
				return err
			}
			defaultPool := map[string]interface{}{}
			defaultPool["pool_id"] = newRegionalId(region, defaultPoolRes.ID)

			d.Set("default_pool", []map[string]interface{}{defaultPool})

		}
	}

	return resourceScalewayK8SClusterBetaDefaultPoolRead(d, m)
}

func resourceScalewayK8SClusterBetaUpdate(d *schema.ResourceData, m interface{}) error {
	k8sAPI, region, clusterID, err := getK8SAPIWithRegionAndID(m, d.Id())
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

	if d.HasChange("description") {
		updateRequest.Description = scw.StringPtr(d.Get("description").(string))
	}

	if d.HasChange("tags") {
		tags := expandStrings(d.Get("tags"))
		updateRequest.Tags = scw.StringsPtr(tags)
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
		updateRequest.Ingress = scw.StringPtr(d.Get("ingress").(string))
	}

	if d.HasChange("enable_dashboard") {
		updateRequest.EnableDashboard = scw.BoolPtr(d.Get("enable_dashboard").(bool))
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
	k8sAPI, region, clusterID, err := getK8SAPIWithRegionAndID(m, d.Id())
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
				Default:     "binpacking",
				Description: "Type of resource estimator to be used in scale up",
			},
			"expander": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "random",
				Description: "Type of node group expander to be used in scale up",
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
