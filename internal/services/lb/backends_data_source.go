package lb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func DataSourceBackends() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourceLbBackendsRead,
		Schema: map[string]*schema.Schema{
			"lb_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "backends with a lb id like it are listed.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Backends with a name like it are listed.",
			},
			"backends": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"name": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"lb_id": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"forward_protocol": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"forward_port": {
							Computed: true,
							Type:     schema.TypeInt,
						},
						"forward_port_algorithm": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"sticky_sessions": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"sticky_sessions_cookie_name": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"server_ips": {
							Computed: true,
							Type:     schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"timeout_server": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"timeout_connect": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"timeout_tunnel": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"on_marked_down_action": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"proxy_protocol": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"failover_host": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"ssl_bridging": {
							Computed: true,
							Type:     schema.TypeBool,
						},
						"ignore_ssl_server_verify": {
							Computed: true,
							Type:     schema.TypeBool,
						},
						"health_check_port": {
							Computed: true,
							Type:     schema.TypeInt,
						},
						"health_check_max_retries": {
							Computed: true,
							Type:     schema.TypeInt,
						},
						"health_check_timeout": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"health_check_delay": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"health_check_tcp": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{},
							},
						},
						"health_check_http": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"uri": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"method": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"code": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"host_header": {
										Computed: true,
										Type:     schema.TypeString,
									},
								},
							},
						},
						"health_check_https": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"uri": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"method": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"code": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"host_header": {
										Computed: true,
										Type:     schema.TypeString,
									},
									"sni": {
										Computed: true,
										Type:     schema.TypeString,
									},
								},
							},
						},
						"created_at": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"update_at": {
							Computed: true,
							Type:     schema.TypeString,
						},
					},
				},
			},
			"zone":            zonal.Schema(),
			"organization_id": account.OrganizationIDSchema(),
			"project_id":      account.ProjectIDSchema(),
		},
	}
}

func DataSourceLbBackendsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	lbAPI, zone, err := lbAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	_, lbID, err := zonal.ParseID(d.Get("lb_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := lbAPI.ListBackends(&lb.ZonedAPIListBackendsRequest{
		Zone: zone,
		LBID: lbID,
		Name: types.ExpandStringPtr(d.Get("name")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	backends := []interface{}(nil)

	for _, backend := range res.Backends {
		rawBackend := make(map[string]interface{})
		rawBackend["id"] = zonal.NewID(zone, backend.ID).String()
		rawBackend["name"] = backend.Name
		rawBackend["lb_id"] = zonal.NewIDString(zone, backend.LB.ID)
		rawBackend["created_at"] = types.FlattenTime(backend.CreatedAt)
		rawBackend["update_at"] = types.FlattenTime(backend.UpdatedAt)
		rawBackend["forward_protocol"] = backend.ForwardProtocol
		rawBackend["forward_port"] = backend.ForwardPort
		rawBackend["forward_port_algorithm"] = flattenLbForwardPortAlgorithm(backend.ForwardPortAlgorithm)
		rawBackend["sticky_sessions"] = flattenLbStickySessionsType(backend.StickySessions)
		rawBackend["sticky_sessions_cookie_name"] = backend.StickySessionsCookieName
		rawBackend["server_ips"] = backend.Pool
		rawBackend["timeout_server"] = types.FlattenDuration(backend.TimeoutServer)
		rawBackend["timeout_connect"] = types.FlattenDuration(backend.TimeoutConnect)
		rawBackend["timeout_tunnel"] = types.FlattenDuration(backend.TimeoutTunnel)
		rawBackend["on_marked_down_action"] = flattenLbBackendMarkdownAction(backend.OnMarkedDownAction)
		rawBackend["proxy_protocol"] = flattenLbProxyProtocol(backend.ProxyProtocol)
		rawBackend["failover_host"] = types.FlattenStringPtr(backend.FailoverHost)
		rawBackend["ssl_bridging"] = types.FlattenBoolPtr(backend.SslBridging)
		rawBackend["ignore_ssl_server_verify"] = types.FlattenBoolPtr(backend.IgnoreSslServerVerify)
		rawBackend["health_check_port"] = backend.HealthCheck.Port
		rawBackend["health_check_max_retries"] = backend.HealthCheck.CheckMaxRetries
		rawBackend["health_check_timeout"] = types.FlattenDuration(backend.HealthCheck.CheckTimeout)
		rawBackend["health_check_delay"] = types.FlattenDuration(backend.HealthCheck.CheckDelay)
		rawBackend["health_check_tcp"] = flattenLbHCTCP(backend.HealthCheck.TCPConfig)
		rawBackend["health_check_http"] = flattenLbHCHTTP(backend.HealthCheck.HTTPConfig)
		rawBackend["health_check_https"] = flattenLbHCHTTPS(backend.HealthCheck.HTTPSConfig)

		backends = append(backends, rawBackend)
	}

	d.SetId(zone.String())
	_ = d.Set("backends", backends)

	return nil
}
