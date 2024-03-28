package iot

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/iot/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/cdf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceRoute() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceIotRouteCreate,
		ReadContext:   ResourceIotRouteRead,
		DeleteContext: ResourceIotRouteDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultIoTHubTimeout),
			Default: schema.DefaultTimeout(defaultIoTHubTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the route",
			},
			"hub_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				Description:      "The ID of the route's hub",
				DiffSuppressFunc: dsf.Locality,
			},
			"topic": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The Topic the route subscribes to (wildcards allowed)",
			},
			"database": {
				Type:        schema.TypeList,
				MinItems:    1,
				MaxItems:    1,
				Optional:    true,
				ForceNew:    true,
				Description: "Database Route parameters",
				ExactlyOneOf: []string{
					iot.RouteRouteTypeDatabase.String(),
					iot.RouteRouteTypeRest.String(),
					iot.RouteRouteTypeS3.String(),
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"query": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "SQL query to be executed ($TOPIC and $PAYLOAD variables are available, see documentation)",
						},
						"host": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "The database hostname",
						},
						"port": {
							Type:        schema.TypeInt,
							Required:    true,
							ForceNew:    true,
							Description: "The database port",
						},
						"dbname": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "The database name",
						},
						"username": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "The database username",
						},
						"password": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "The database password",
							Sensitive:   true,
						},
					},
				},
			},
			"rest": {
				Type:        schema.TypeList,
				MinItems:    1,
				MaxItems:    1,
				Optional:    true,
				ForceNew:    true,
				Description: "Rest Route parameters",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"verb": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "The HTTP Verb used to call REST URI",
							ValidateFunc: validation.StringInSlice([]string{
								iot.RouteRestConfigHTTPVerbGet.String(),
								iot.RouteRestConfigHTTPVerbPost.String(),
								iot.RouteRestConfigHTTPVerbPut.String(),
								iot.RouteRestConfigHTTPVerbPatch.String(),
								iot.RouteRestConfigHTTPVerbDelete.String(),
							}, false),
						},
						"uri": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "The URI of the REST endpoint",
						},
						"headers": {
							Type:        schema.TypeMap,
							Required:    true,
							ForceNew:    true,
							Description: "The HTTP call extra headers",
							Elem: &schema.Schema{
								Type:     schema.TypeString,
								ForceNew: true,
							},
						},
					},
				},
			},
			"s3": {
				Type:        schema.TypeList,
				MinItems:    1,
				MaxItems:    1,
				Optional:    true,
				ForceNew:    true,
				Description: "S3 Route parameters",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket_region": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "The region of the S3 route's destination bucket",
						},
						"bucket_name": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "The name of the S3 route's destination bucket",
						},
						"object_prefix": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "The string to prefix object names with",
						},
						"strategy": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "How the S3 route's objects will be created: one per topic or one per message",
							ValidateFunc: validation.StringInSlice([]string{
								iot.RouteS3ConfigS3StrategyPerTopic.String(),
								iot.RouteS3ConfigS3StrategyPerMessage.String(),
							}, false),
						},
					},
				},
			},
			// Computed elements
			"region": regional.Schema(),
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the IoT Route",
			},
		},
		CustomizeDiff: cdf.LocalityCheck("hub_id"),
	}
}

func ResourceIotRouteCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	iotAPI, region, err := iotAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	hubID := zonal.ExpandID(d.Get("hub_id")).ID
	_, err = waitIotHub(ctx, iotAPI, region, hubID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	req := &iot.CreateRouteRequest{
		Region: region,
		Name:   types.ExpandOrGenerateString(d.Get("name"), "route"),
		HubID:  zonal.ExpandID(d.Get("hub_id")).ID,
		Topic:  d.Get("topic").(string),
	}

	if _, ok := d.GetOk(iot.RouteRouteTypeS3.String()); ok {
		prefixKey := iot.RouteRouteTypeS3.String() + ".0"
		req.S3Config = &iot.CreateRouteRequestS3Config{
			BucketRegion: d.Get(prefixKey + ".bucket_region").(string),
			BucketName:   d.Get(prefixKey + ".bucket_name").(string),
			ObjectPrefix: d.Get(prefixKey + ".object_prefix").(string),
			Strategy:     iot.RouteS3ConfigS3Strategy(d.Get(prefixKey + ".strategy").(string)),
		}
	} else if _, ok := d.GetOk(iot.RouteRouteTypeRest.String()); ok {
		prefixKey := iot.RouteRouteTypeRest.String() + ".0"
		req.RestConfig = &iot.CreateRouteRequestRestConfig{
			Verb:    iot.RouteRestConfigHTTPVerb(d.Get(prefixKey + ".verb").(string)),
			URI:     d.Get(prefixKey + ".uri").(string),
			Headers: extractRestHeaders(d, prefixKey+".headers"),
		}
	} else if _, ok := d.GetOk(iot.RouteRouteTypeDatabase.String()); ok {
		prefixKey := iot.RouteRouteTypeDatabase.String() + ".0"
		req.DbConfig = &iot.CreateRouteRequestDatabaseConfig{
			Host:     d.Get(prefixKey + ".host").(string),
			Port:     uint32(d.Get(prefixKey + ".port").(int)),
			Dbname:   d.Get(prefixKey + ".dbname").(string),
			Username: d.Get(prefixKey + ".username").(string),
			Password: d.Get(prefixKey + ".password").(string),
			Query:    d.Get(prefixKey + ".query").(string),
		}
	} else {
		return diag.FromErr(errors.New("no route type have been chosen"))
	}

	res, err := iotAPI.CreateRoute(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, res.ID))

	_, err = waitIotHub(ctx, iotAPI, region, hubID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceIotRouteRead(ctx, d, m)
}

func ResourceIotRouteRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	iotAPI, region, routeID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	response, err := iotAPI.GetRoute(&iot.GetRouteRequest{
		Region:  region,
		RouteID: routeID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("region", string(region))
	_ = d.Set("name", response.Name)
	_ = d.Set("hub_id", response.HubID)
	_ = d.Set("topic", response.Topic)
	_ = d.Set("created_at", response.CreatedAt.String())

	switch response.Type {
	case iot.RouteRouteTypeDatabase:
		conf := []map[string]interface{}{{
			"query":    response.DbConfig.Query,
			"host":     response.DbConfig.Host,
			"port":     int(response.DbConfig.Port),
			"dbname":   response.DbConfig.Dbname,
			"username": response.DbConfig.Username,
			// Password is never returned. To avoid password getting erased, take it back.
			"password": d.Get(iot.RouteRouteTypeDatabase.String() + ".0.password"),
		}}
		_ = d.Set("database", conf)
	case iot.RouteRouteTypeRest:
		conf := []map[string]interface{}{{
			"verb":    response.RestConfig.Verb,
			"uri":     response.RestConfig.URI,
			"headers": response.RestConfig.Headers,
		}}
		_ = d.Set("rest", conf)
	case iot.RouteRouteTypeS3:
		conf := []map[string]interface{}{{
			"bucket_region": response.S3Config.BucketRegion,
			"bucket_name":   response.S3Config.BucketName,
			"object_prefix": response.S3Config.ObjectPrefix,
			"strategy":      response.S3Config.Strategy,
		}}
		_ = d.Set("s3", conf)
	}

	return nil
}

func ResourceIotRouteDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	iotAPI, region, routeID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	hubID := zonal.ExpandID(d.Get("hub_id")).ID
	_, err = waitIotHub(ctx, iotAPI, region, hubID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	err = iotAPI.DeleteRoute(&iot.DeleteRouteRequest{
		Region:  region,
		RouteID: routeID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return nil
		}
		return diag.FromErr(err)
	}

	_, err = waitIotHub(ctx, iotAPI, region, hubID, d.Timeout(schema.TimeoutCreate))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
