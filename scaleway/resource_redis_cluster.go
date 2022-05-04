package scaleway

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	redis "github.com/scaleway/scaleway-sdk-go/api/redis/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayRedisCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayRedisClusterCreate,
		ReadContext:   resourceScalewayRedisClusterRead,
		UpdateContext: resourceScalewayRedisClusterUpdate,
		DeleteContext: resourceScalewayRedisClusterDelete,
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultRedisClusterTimeout),
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
				Description: "Name of the redis cluster",
			},
			"version": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Redis version of the cluster",
			},
			"node_type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Type of node to use for the cluster",
			},
			"user_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the user created when the cluster is created",
			},
			"password": {
				Type:        schema.TypeString,
				Sensitive:   true,
				Required:    true,
				Description: "Password of the user",
			},
			"tags": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of tags [\"tag1\", \"tag2\", ...] attached to a redis cluster",
			},
			"cluster_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Number of nodes for the cluster.",
			},
			"tls_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether or not TLS is enabled.",
				ForceNew:    true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the Redis cluster",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last update of the Redis cluster",
			},
			"acl": {
				Type:        schema.TypeList,
				Description: "List of acl rules.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Description: "ID of the rule (UUID format).",
							Computed:    true,
						},
						"ip": {
							Type:        schema.TypeString,
							Description: "IPv4 network address of the rule (IP network).",
							Required:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "Description of the rule.",
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
			// Common
			"zone":       zoneSchema(),
			"project_id": projectIDSchema(),
		},
	}
}

func resourceScalewayRedisClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	redisAPI, zone, err := redisAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	createReq := &redis.CreateClusterRequest{
		Zone:      zone,
		ProjectID: d.Get("project_id").(string),
		Name:      expandOrGenerateString(d.Get("name"), "redis"),
		Version:   d.Get("version").(string),
		NodeType:  d.Get("node_type").(string),
		UserName:  d.Get("user_name").(string),
		Password:  d.Get("password").(string),
	}

	tags, tagsExist := d.GetOk("tags")
	if tagsExist {
		createReq.Tags = expandStrings(tags)
	}
	clusterSize, clusterSizeExist := d.GetOk("cluster_size")
	if clusterSizeExist {
		createReq.ClusterSize = scw.Int32Ptr(int32(clusterSize.(int)))
	}
	tlsEnabled, tlsEnabledExist := d.GetOk("tls_enabled")
	if tlsEnabledExist {
		createReq.TLSEnabled = tlsEnabled.(bool)
	}
	aclRules, aclRulesExist := d.GetOk("acl")
	if aclRulesExist {
		rules, err := expandRedisACLSpecs(aclRules)
		if err != nil {
			return diag.FromErr(err)
		}
		createReq.ACLRules = rules
	}

	res, err := redisAPI.CreateCluster(createReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newZonedIDString(zone, res.ID))

	_, err = waitForRedisCluster(ctx, redisAPI, zone, res.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayRedisClusterRead(ctx, d, meta)
}

func resourceScalewayRedisClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	redisAPI, zone, ID, err := redisAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	getReq := &redis.GetClusterRequest{
		Zone:      zone,
		ClusterID: ID,
	}
	cluster, err := redisAPI.GetCluster(getReq, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("name", cluster.Name)
	_ = d.Set("node_type", cluster.NodeType)
	_ = d.Set("user_name", d.Get("user_name").(string))
	_ = d.Set("password", d.Get("password").(string))
	_ = d.Set("zone", cluster.Zone.String())
	_ = d.Set("project_id", cluster.ProjectID)
	_ = d.Set("version", cluster.Version)
	_ = d.Set("cluster_size", cluster.ClusterSize)
	_ = d.Set("created_at", cluster.CreatedAt.Format(time.RFC3339))
	_ = d.Set("updated_at", cluster.UpdatedAt.Format(time.RFC3339))
	_ = d.Set("acl", flattenRedisACLs(cluster.ACLRules))

	if len(cluster.Tags) > 0 {
		_ = d.Set("tags", cluster.Tags)
	}

	return nil
}

func resourceScalewayRedisClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	redisAPI, zone, ID, err := redisAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	req := &redis.UpdateClusterRequest{
		Zone:      zone,
		ClusterID: ID,
	}

	if d.HasChange("name") {
		req.Name = expandStringPtr(d.Get("name"))
	}
	if d.HasChange("user_name") {
		req.UserName = expandStringPtr(d.Get("user_name"))
	}
	if d.HasChange("password") {
		req.Password = expandStringPtr(d.Get("password"))
	}
	if d.HasChange("tags") {
		req.Tags = expandStrings(d.Get("tags"))
	}
	if d.HasChange("acl") {
		diagnostics := resourceScalewayRedisClusterUpdateACL(ctx, d, redisAPI, zone, ID)
		if diagnostics != nil {
			return diagnostics
		}
	}

	_, err = waitForRedisCluster(ctx, redisAPI, zone, ID, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = redisAPI.UpdateCluster(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	migrateClusterRequests := []redis.MigrateClusterRequest(nil)
	if d.HasChange("cluster_size") {
		migrateClusterRequests = append(migrateClusterRequests, redis.MigrateClusterRequest{
			Zone:        zone,
			ClusterID:   ID,
			ClusterSize: scw.Uint32Ptr(uint32(d.Get("cluster_size").(int))),
		})
	}
	if d.HasChange("version") {
		migrateClusterRequests = append(migrateClusterRequests, redis.MigrateClusterRequest{
			Zone:      zone,
			ClusterID: ID,
			Version:   expandStringPtr(d.Get("version")),
		})
	}
	if d.HasChange("node_type") {
		migrateClusterRequests = append(migrateClusterRequests, redis.MigrateClusterRequest{
			Zone:      zone,
			ClusterID: ID,
			NodeType:  expandStringPtr(d.Get("node_type")),
		})
	}
	for _, request := range migrateClusterRequests {
		_, err = waitForRedisCluster(ctx, redisAPI, zone, ID, d.Timeout(schema.TimeoutUpdate))
		if err != nil && !is404Error(err) {
			return diag.FromErr(err)
		}
		_, err = redisAPI.MigrateCluster(&request, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = waitForRedisCluster(ctx, redisAPI, zone, ID, d.Timeout(schema.TimeoutUpdate))
		if err != nil && !is404Error(err) {
			return diag.FromErr(err)
		}
	}

	_, err = waitForRedisCluster(ctx, redisAPI, zone, ID, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayRedisClusterRead(ctx, d, meta)
}

func resourceScalewayRedisClusterUpdateACL(ctx context.Context, d *schema.ResourceData, redisAPI *redis.API, zone scw.Zone, clusterID string) diag.Diagnostics {
	rules, err := expandRedisACLSpecs(d.Get("acl"))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = redisAPI.SetACLRules(&redis.SetACLRulesRequest{
		Zone:      zone,
		ClusterID: clusterID,
		ACLRules:  rules,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceScalewayRedisClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	redisAPI, zone, ID, err := redisAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForRedisCluster(ctx, redisAPI, zone, ID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = redisAPI.DeleteCluster(&redis.DeleteClusterRequest{
		Zone:      zone,
		ClusterID: ID,
	}, scw.WithContext(ctx))

	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForRedisCluster(ctx, redisAPI, zone, ID, d.Timeout(schema.TimeoutDelete))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
