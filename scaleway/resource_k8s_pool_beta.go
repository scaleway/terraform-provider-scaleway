package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	k8s "github.com/scaleway/scaleway-sdk-go/api/k8s/v1beta3"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayK8SPoolBeta() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayK8SPoolBetaCreate,
		Read:   resourceScalewayK8SPoolBetaRead,
		Update: resourceScalewayK8SPoolBetaUpdate,
		Delete: resourceScalewayK8SPoolBetaDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the cluster on which this pool will be created",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the cluster",
			},
			"node_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				Description:      "Server type of the pool servers",
				DiffSuppressFunc: diffSuppressFuncIgnoreCaseAndHyphen,
			},
			"autoscaling": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable the autoscaling on the pool",
			},
			"autohealing": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable the autohealing on the pool",
			},
			"size": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Size of the pool",
			},
			"min_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "Minimun size of the pool",
			},
			"max_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Maximum size of the pool",
			},
			"container_runtime": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "docker",
				ForceNew:    true,
				Description: "Container runtime for the pool",
			},
			"placement_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     nil,
				Description: "ID of the placement group",
			},
			"region": regionSchema(),
			// Computed elements
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the pool",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last update of the pool",
			},
			"version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Kubernetes version of the pool",
			},
		},
	}
}

func resourceScalewayK8SPoolBetaCreate(d *schema.ResourceData, m interface{}) error {
	k8sAPI, region, err := getK8SAPIWithRegion(d, m)
	if err != nil {
		return err
	}

	////
	// Create pool
	////

	req := &k8s.CreatePoolRequest{
		Region:      region,
		ClusterID:   expandID(d.Get("cluster_id")),
		Name:        expandOrGenerateString(d.Get("name"), "pool"),
		NodeType:    d.Get("node_type").(string),
		Autoscaling: d.Get("autoscaling").(bool),
		Autohealing: d.Get("autohealing").(bool),
		Size:        uint32(d.Get("size").(int)),
	}

	if placementGroupID, ok := d.GetOk("placement_group_id"); ok {
		req.PlacementGroupID = scw.StringPtr(expandID(placementGroupID.(string)))
	}

	if minSize, ok := d.GetOk("min_size"); ok {
		req.MinSize = scw.Uint32Ptr(uint32(minSize.(int)))
	}

	if maxSize, ok := d.GetOk("max_size"); ok {
		req.MaxSize = scw.Uint32Ptr(uint32(maxSize.(int)))
	} else {
		req.MaxSize = scw.Uint32Ptr(req.Size)
	}

	if containerRuntime, ok := d.GetOk("container_runtime"); ok {
		req.ContainerRuntime = scw.StringPtr(containerRuntime.(string))
	}

	res, err := k8sAPI.CreatePool(req)
	if err != nil {
		return err
	}

	d.SetId(newRegionalId(region, res.ID))

	return resourceScalewayK8SPoolBetaRead(d, m)
}

func resourceScalewayK8SPoolBetaRead(d *schema.ResourceData, m interface{}) error {
	k8sAPI, region, poolID, err := getK8SAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	////
	// Read Pool
	////
	pool, err := k8sAPI.GetPool(&k8s.GetPoolRequest{
		Region: region,
		PoolID: poolID,
	})
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("cluster_id", newRegionalId(region, pool.ClusterID))
	d.Set("name", pool.Name)
	d.Set("node_type", pool.NodeType)
	d.Set("autoscaling", pool.Autoscaling)
	d.Set("autohealing", pool.Autohealing)
	d.Set("size", pool.Size)
	d.Set("version", pool.Version)
	d.Set("min_size", pool.MinSize)
	d.Set("max_size", pool.MaxSize)
	d.Set("container_runtime", pool.ContainerRuntime)
	d.Set("created_at", pool.CreatedAt)
	d.Set("updated_at", pool.UpdatedAt)

	if pool.PlacementGroupID != nil {
		d.Set("placement_group_id", newZonedIdFromRegion(region, *pool.PlacementGroupID)) // TODO fix this ZonedIdFromRegion
	}

	return nil
}

func resourceScalewayK8SPoolBetaUpdate(d *schema.ResourceData, m interface{}) error {
	k8sAPI, region, poolID, err := getK8SAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	////
	// Update Pool
	////
	updateRequest := &k8s.UpdatePoolRequest{
		Region: region,
		PoolID: poolID,
	}

	if d.HasChange("autoscaling") {
		updateRequest.Autoscaling = scw.BoolPtr(d.Get("autoscaling").(bool))
	}

	if d.HasChange("autohealing") {
		updateRequest.Autohealing = scw.BoolPtr(d.Get("autohealing").(bool))
	}

	if d.HasChange("min_size") {
		updateRequest.MinSize = scw.Uint32Ptr(uint32(d.Get("min_size").(int)))
	}

	if d.HasChange("max_size") {
		updateRequest.MaxSize = scw.Uint32Ptr(uint32(d.Get("max_size").(int)))
	}

	if d.Get("autoscaling").(bool) == false && d.HasChange("size") {
		updateRequest.Size = scw.Uint32Ptr(uint32(d.Get("size").(int)))
	}

	_, err = k8sAPI.UpdatePool(updateRequest)
	if err != nil {
		return err
	}

	return resourceScalewayK8SPoolBetaRead(d, m)
}

func resourceScalewayK8SPoolBetaDelete(d *schema.ResourceData, m interface{}) error {
	k8sAPI, region, poolID, err := getK8SAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	////
	// Delete Pool
	////
	_, err = k8sAPI.DeletePool(&k8s.DeletePoolRequest{
		Region: region,
		PoolID: poolID,
	})
	if err != nil {
		if !is404Error(err) {
			return err
		}
	}

	return nil
}
