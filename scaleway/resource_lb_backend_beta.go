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
			Port:            80,
			CheckMaxRetries: 1,
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
	d.Set("on_marked_down_action", flattenLbBackendMarkdownAction(res.OnMarkedDownAction))

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
