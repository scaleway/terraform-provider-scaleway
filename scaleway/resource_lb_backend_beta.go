package scaleway

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
)

func resourceScalewayLbBackendBeta() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayLbBackendBetaCreate,
		Read:   resourceScalewayLbBackendBetaRead,
		Update: resourceScalewayLbBackendBetaUpdate,
		Delete: resourceScalewayLbBackendBetaDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"lb_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The load-balancer ID",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of the backend",
			},
			"forward_protocol": {
				Type: schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{
					lb.ProtocolTCP.String(),
					lb.ProtocolHTTP.String(),
				}, false),
				Required:    true,
				Description: "Backend protocol",
			},
			"forward_port": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "User sessions will be forwarded to this port of backend servers",
			},
			"forward_port_algorithm": {
				Type: schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{
					lb.ForwardPortAlgorithmRoundrobin.String(),
					lb.ForwardPortAlgorithmLeastconn.String(),
				}, false),
				Default:     lb.ForwardPortAlgorithmRoundrobin.String(),
				Optional:    true,
				Description: "Load balancing algorithm",
			},
			"sticky_sessions": {
				Type: schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{
					lb.StickySessionsTypeNone.String(),
					lb.StickySessionsTypeCookie.String(),
					lb.StickySessionsTypeTable.String(),
				}, false),
				Default:     lb.StickySessionsTypeNone.String(),
				Optional:    true,
				Description: "Load balancing algorithm",
			},
			"sticky_sessions_cookie_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Cookie name for for sticky sessions",
			},
			"server_ips": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.SingleIP(),
				},
				Optional:    true,
				Description: "Backend server IP addresses list (IPv4 or IPv6)",
			},
			"send_proxy_v2": {
				Type:        schema.TypeBool,
				Description: "Enables PROXY protocol version 2",
				Optional:    true,
				Default:     false,
			},

			// Timeouts
			"timeout_server": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: diffSuppressFuncDuration,
				ValidateFunc:     validateDuration(),
				Description:      "Maximum server connection inactivity time",
			},
			"timeout_connect": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: diffSuppressFuncDuration,
				ValidateFunc:     validateDuration(),
				Description:      "Maximum initial server connection establishment time",
			},
			"timeout_tunnel": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: diffSuppressFuncDuration,
				ValidateFunc:     validateDuration(),
				Description:      "Maximum tunnel inactivity time",
			},

			// Health Check
			"health_check_timeout": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateDuration(),
				Default:      "30s",
				Description:  "Timeout before we consider a HC request failed",
			},
			"health_check_delay": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateDuration(),
				Default:      "60s",
				Description:  "Interval between two HC requests",
			},
			"health_check_port": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Port the HC requests will be send to. Default to `forward_port`",
			},
			"health_check_max_retries": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     2,
				Description: "Number of allowed failed HC requests before the backend server is marked down",
			},
			"health_check_tcp": {
				Type:          schema.TypeList,
				MaxItems:      1,
				ConflictsWith: []string{"health_check_http", "health_check_https"},
				Optional:      true,
				Computed:      true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{},
				},
			},
			"health_check_http": {
				Type:          schema.TypeList,
				MaxItems:      1,
				ConflictsWith: []string{"health_check_tcp", "health_check_https"},
				Optional:      true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uri": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The HTTP endpoint URL to call for HC requests",
						},
						"method": {
							Type:        schema.TypeString,
							Default:     "GET",
							Optional:    true,
							Description: "The HTTP method to use for HC requests",
						},
						"code": {
							Type:        schema.TypeInt,
							Default:     200,
							Optional:    true,
							Description: "The expected HTTP status code",
						},
					},
				},
			},
			"health_check_https": {
				Type:          schema.TypeList,
				MaxItems:      1,
				ConflictsWith: []string{"health_check_tcp", "health_check_http"},
				Optional:      true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uri": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The HTTP endpoint URL to call for HC requests",
						},
						"method": {
							Type:        schema.TypeString,
							Default:     "GET",
							Optional:    true,
							Description: "The HTTP method to use for HC requests",
						},
						"code": {
							Type:        schema.TypeInt,
							Default:     200,
							Optional:    true,
							Description: "The expected HTTP status code",
						},
					},
				},
			},
			"on_marked_down_action": {
				Type: schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{
					"none",
					lb.OnMarkedDownActionShutdownSessions.String(),
				}, false),
				Default:     "none",
				Optional:    true,
				Description: "Modify what occurs when a backend server is marked down",
			},
		},
	}
}

func resourceScalewayLbBackendBetaCreate(d *schema.ResourceData, m interface{}) error {
	lbAPI := getLbAPI(m)

	region, LbID, err := parseRegionalID(d.Get("lb_id").(string))
	if err != nil {
		return err
	}

	name, ok := d.GetOk("name")
	if !ok {
		name = getRandomName("lb-bkd")
	}

	healthCheckPort := d.Get("health_check_port").(int)
	if healthCheckPort == 0 {
		healthCheckPort = d.Get("forward_port").(int)
	}

	createReq := &lb.CreateBackendRequest{
		Region:                   region,
		LbID:                     LbID,
		Name:                     name.(string),
		ForwardProtocol:          expandLbProtocol(d.Get("forward_protocol")),
		ForwardPort:              int32(d.Get("forward_port").(int)),
		ForwardPortAlgorithm:     expandLbForwardPortAlgorithm(d.Get("forward_port_algorithm")),
		StickySessions:           expandLbStickySessionsType(d.Get("sticky_sessions")),
		StickySessionsCookieName: d.Get("sticky_sessions_cookie_name").(string),
		HealthCheck: &lb.HealthCheck{
			Port:            int32(healthCheckPort),
			CheckMaxRetries: int32(d.Get("health_check_max_retries").(int)),
			CheckTimeout:    expandDuration(d.Get("health_check_timeout")),
			TCPConfig:       expandLbHCTCP(d.Get("health_check_tcp")),
			HTTPConfig:      expandLbHCHTTP(d.Get("health_check_http")),
			HTTPSConfig:     expandLbHCHTTPS(d.Get("health_check_https")),
		},
		ServerIP:           StringSliceFromState(d.Get("server_ips").([]interface{})),
		SendProxyV2:        d.Get("send_proxy_v2").(bool),
		TimeoutServer:      expandDuration(d.Get("timeout_server")),
		TimeoutConnect:     expandDuration(d.Get("timeout_connect")),
		TimeoutTunnel:      expandDuration(d.Get("timeout_tunnel")),
		OnMarkedDownAction: expandLbBackendMarkdownAction(d.Get("on_marked_down_action")),
	}

	res, err := lbAPI.CreateBackend(createReq)
	if err != nil {
		return err
	}

	d.SetId(newRegionalId(region, res.ID))

	return resourceScalewayLbBackendBetaRead(d, m)
}

func resourceScalewayLbBackendBetaRead(d *schema.ResourceData, m interface{}) error {
	lbAPI, region, ID, err := getLbAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	res, err := lbAPI.GetBackend(&lb.GetBackendRequest{
		Region:    region,
		BackendID: ID,
	})

	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("lb_id", newRegionalId(region, res.Lb.ID))
	d.Set("name", res.Name)
	d.Set("forward_protocol", flattenLbProtocol(res.ForwardProtocol))
	d.Set("forward_port", res.ForwardPort)
	d.Set("forward_port_algorithm", flattenLbForwardPortAlgorithm(res.ForwardPortAlgorithm))
	d.Set("sticky_sessions", flattenLbStickySessionsType(res.StickySessions))
	d.Set("sticky_sessions_cookie_name", res.StickySessionsCookieName)
	d.Set("server_ips", res.Pool)
	d.Set("send_proxy_v2", res.SendProxyV2)
	d.Set("timeout_server", flattenDuration(res.TimeoutServer))
	d.Set("timeout_connect", flattenDuration(res.TimeoutConnect))
	d.Set("timeout_tunnel", flattenDuration(res.TimeoutTunnel))
	d.Set("health_check_port", res.HealthCheck.Port)
	d.Set("health_check_max_retries", res.HealthCheck.CheckMaxRetries)
	d.Set("health_check_timeout", flattenDuration(res.HealthCheck.CheckTimeout))
	d.Set("on_marked_down_action", flattenLbBackendMarkdownAction(res.OnMarkedDownAction))
	d.Set("health_check_tcp", flattenLbHCTCP(res.HealthCheck.TCPConfig))
	d.Set("health_check_http", flattenLbHCHTTP(res.HealthCheck.HTTPConfig))
	d.Set("health_check_https", flattenLbHCHTTPS(res.HealthCheck.HTTPSConfig))

	return nil
}

func resourceScalewayLbBackendBetaUpdate(d *schema.ResourceData, m interface{}) error {
	lbAPI, region, ID, err := getLbAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	req := &lb.UpdateBackendRequest{
		Region:                   region,
		BackendID:                ID,
		Name:                     d.Get("name").(string),
		ForwardProtocol:          expandLbProtocol(d.Get("forward_protocol")),
		ForwardPort:              int32(d.Get("forward_port").(int)),
		ForwardPortAlgorithm:     expandLbForwardPortAlgorithm(d.Get("forward_port_algorithm")),
		StickySessions:           expandLbStickySessionsType(d.Get("sticky_sessions")),
		StickySessionsCookieName: d.Get("sticky_sessions_cookie_name").(string),
		SendProxyV2:              d.Get("send_proxy_v2").(bool),
		TimeoutServer:            expandDuration(d.Get("timeout_server")),
		TimeoutConnect:           expandDuration(d.Get("timeout_connect")),
		TimeoutTunnel:            expandDuration(d.Get("timeout_tunnel")),
		OnMarkedDownAction:       expandLbBackendMarkdownAction(d.Get("on_marked_down_action")),
	}

	_, err = lbAPI.UpdateBackend(req)
	if err != nil {
		return err
	}

	// Update Health Check
	updateHCRequest := &lb.UpdateHealthCheckRequest{
		Region:          region,
		BackendID:       ID,
		Port:            int32(d.Get("health_check_port").(int)),
		CheckMaxRetries: int32(d.Get("health_check_max_retries").(int)),
		CheckTimeout:    expandDuration(d.Get("health_check_timeout")),
		HTTPConfig:      expandLbHCHTTP(d.Get("health_check_http")),
		HTTPSConfig:     expandLbHCHTTPS(d.Get("health_check_https")),
	}

	// As this is the default behaviour If no other HC type are present we enable TCP
	if updateHCRequest.HTTPConfig == nil && updateHCRequest.HTTPSConfig == nil {
		updateHCRequest.TCPConfig = expandLbHCTCP(d.Get("health_check_tcp"))
	}

	_, err = lbAPI.UpdateHealthCheck(updateHCRequest)
	if err != nil {
		return err
	}

	// Update Backend servers
	_, err = lbAPI.SetBackendServers(&lb.SetBackendServersRequest{
		Region:    region,
		BackendID: ID,
		ServerIP:  StringSliceFromState(d.Get("server_ips").([]interface{})),
	})
	if err != nil {
		return err
	}

	return resourceScalewayLbBackendBetaRead(d, m)
}

func resourceScalewayLbBackendBetaDelete(d *schema.ResourceData, m interface{}) error {
	lbAPI, region, ID, err := getLbAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	err = lbAPI.DeleteBackend(&lb.DeleteBackendRequest{
		Region:    region,
		BackendID: ID,
	})

	if err != nil && !is404Error(err) {
		return err
	}

	return nil
}
