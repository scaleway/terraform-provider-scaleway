package lb

import (
	"context"
	"fmt"

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
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of backends.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Computed:    true,
							Description: "Backend ID",
							Type:        schema.TypeString,
						},
						"name": {
							Computed:    true,
							Description: "Name of the backend",
							Type:        schema.TypeString,
						},
						"lb_id": {
							Computed:    true,
							Description: "ID of the load balancer the backend is associated with",
							Type:        schema.TypeString,
						},
						"forward_protocol": {
							Computed:    true,
							Description: "Protocol to be used by the backend when forwarding traffic to backend servers",
							Type:        schema.TypeString,
						},
						"forward_port": {
							Computed:    true,
							Description: "Port to be used by the backend when forwarding traffic to backend servers",
							Type:        schema.TypeInt,
						},
						"forward_port_algorithm": {
							Computed: true,
							Description: func() string {
								var t lb.ForwardPortAlgorithm
								values := t.Values()
								return fmt.Sprintf("Load balancing algorithm to be used when determining which backend server to forward new traffic to. Possible values are: %s", values)
							}(),
							Type: schema.TypeString,
						},
						"sticky_sessions": {
							Computed:    true,
							Description: "Defines whether sticky sessions (binding a particular session to a particular backend server) are activated and the method to use if so. None disables sticky sessions. Cookie-based uses an HTTP cookie to stick a session to a backend server. Table-based uses the source (client) IP address to stick a session to a backend server.",
							Type:        schema.TypeString,
						},
						"sticky_sessions_cookie_name": {
							Computed:    true,
							Description: "cookie name for cookie-based sticky sessions.",
							Type:        schema.TypeString,
						},
						"server_ips": {
							Computed:    true,
							Description: "List of IP addresses for the server attached to the backend.",
							Type:        schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"timeout_server": {
							Computed:    true,
							Description: "Maximum allowed time for a backend server to process a request",
							Type:        schema.TypeString,
						},
						"timeout_connect": {
							Computed:    true,
							Description: "Maximum allowed time for establishing a connection to a backend server",
							Type:        schema.TypeString,
						},
						"timeout_tunnel": {
							Computed:    true,
							Description: "Maximum allowed tunnel inactivity time after Websocket is established (takes precedence over client and server timeout)",
							Type:        schema.TypeString,
						},
						"on_marked_down_action": {
							Computed:    true,
							Description: "Action to take when a backend server is marked as down",
							Type:        schema.TypeString,
						},
						"proxy_protocol": {
							Computed:    true,
							Description: "protocol to use between the Load Balancer and backend servers. Allows the backend servers to be informed of the client's real IP address. The PROXY protocol must be supported by the backend servers' software.",
							Type:        schema.TypeString,
						},
						"failover_host": {
							Computed:    true,
							Description: "Scaleway Object Storage bucket website to be served as failover if all backend servers are down, e.g. failover-website.s3-website.fr-par.scw.cloud",
							Type:        schema.TypeString,
						},
						"ssl_bridging": {
							Computed:    true,
							Description: "Defines whether to enable SSL bridging between the Load Balancer and backend servers",
							Type:        schema.TypeBool,
						},
						"ignore_ssl_server_verify": {
							Computed:    true,
							Description: "Defines whether the server certificate verification should be ignored",
							Type:        schema.TypeBool,
						},
						"health_check_port": {
							Computed:    true,
							Description: "port to use for the backend server health check.",
							Type:        schema.TypeInt,
						},
						"health_check_max_retries": {
							Computed:    true,
							Description: "Number of retries when a backend server connection failed.",
							Type:        schema.TypeInt,
						},
						"health_check_timeout": {
							Computed:    true,
							Description: "Maximum time a backend server has to reply to the health check.",
							Type:        schema.TypeString,
						},
						"health_check_delay": {
							Computed:    true,
							Description: "Time to wait between two consecutive health checks.",
							Type:        schema.TypeString,
						},
						"health_check_tcp": {
							Type:        schema.TypeList,
							Description: "TCP Health check configuration.",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{},
							},
						},
						"health_check_http": {
							Type:        schema.TypeList,
							Description: "HTTP Health check configuration.",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"uri": {
										Type:        schema.TypeString,
										Description: "the HTTP path to use when performing a health check on backend servers.",
										Computed:    true,
									},
									"method": {
										Type:        schema.TypeString,
										Description: "the HTTP method used when performing a health check on backend servers.",
										Computed:    true,
									},
									"code": {
										Type:        schema.TypeInt,
										Description: "the HTTP response code that should be returned for a health check to be considered successful.",
										Computed:    true,
									},
									"host_header": {
										Computed:    true,
										Description: "the HTTP host header used when performing a health check on backend servers.",
										Type:        schema.TypeString,
									},
								},
							},
						},
						"health_check_https": {
							Type:        schema.TypeList,
							Description: "Health checks configuration in HTTPS",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"uri": {
										Type:        schema.TypeString,
										Description: "the HTTP path to use when performing a health check on backend servers.",
										Computed:    true,
									},
									"method": {
										Type:        schema.TypeString,
										Description: "the HTTP method used when performing a health check on backend servers.",
										Computed:    true,
									},
									"code": {
										Type:        schema.TypeInt,
										Description: "the HTTP response code that should be returned for a health check to be considered successful.",
										Computed:    true,
									},
									"host_header": {
										Computed:    true,
										Description: "the HTTP host header used when performing a health check on backend servers.",
										Type:        schema.TypeString,
									},
									"sni": {
										Computed:    true,
										Description: "the SNI value used when performing a health check on backend servers over SSL.",
										Type:        schema.TypeString,
									},
								},
							},
						},
						"created_at": {
							Computed:    true,
							Description: "Timestamp when the backend server was created (RFC3339)",
							Type:        schema.TypeString,
						},
						"update_at": {
							Computed:    true,
							Description: "Time at which the backend server was last updated (RFC3339).",
							Type:        schema.TypeString,
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

func DataSourceLbBackendsRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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

	backends := []any(nil)

	for _, backend := range res.Backends {
		rawBackend := make(map[string]any)
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
