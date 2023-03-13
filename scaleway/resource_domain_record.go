package scaleway

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayDomainRecord() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayDomainRecordCreate,
		ReadContext:   resourceScalewayDomainRecordRead,
		UpdateContext: resourceScalewayDomainRecordUpdate,
		DeleteContext: resourceScalewayDomainRecordDelete,
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultDomainRecordTimeout),
			Read:    schema.DefaultTimeout(defaultDomainRecordTimeout),
			Update:  schema.DefaultTimeout(defaultDomainRecordTimeout),
			Delete:  schema.DefaultTimeout(defaultDomainRecordTimeout),
			Default: schema.DefaultTimeout(defaultDomainRecordTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"dns_zone": {
				Type:        schema.TypeString,
				Description: "The zone you want to add the record in",
				Required:    true,
				ForceNew:    true,
			},
			"keep_empty_zone": {
				Type:        schema.TypeBool,
				Description: "When destroy a resource record, if a zone have only NS, delete the zone",
				Optional:    true,
				Default:     false,
			},
			"root_zone": {
				Type:        schema.TypeBool,
				Description: "Does the DNS zone is the root zone or not",
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the record",
				ForceNew:    true,
				Optional:    true,
			},
			"type": {
				Type:        schema.TypeString,
				Description: "The type of the record",
				ValidateFunc: validation.StringInSlice([]string{
					domain.RecordTypeA.String(),
					domain.RecordTypeAAAA.String(),
					domain.RecordTypeALIAS.String(),
					domain.RecordTypeCNAME.String(),
					domain.RecordTypeDNAME.String(),
					domain.RecordTypeMX.String(),
					domain.RecordTypeNS.String(),
					domain.RecordTypePTR.String(),
					domain.RecordTypeSRV.String(),
					domain.RecordTypeTXT.String(),
					domain.RecordTypeTLSA.String(),
					domain.RecordTypeCAA.String(),
				}, false),
				ForceNew: true,
				Required: true,
			},
			"data": {
				Type:        schema.TypeString,
				Description: "The data of the record",
				Required:    true,
			},
			"ttl": {
				Type:         schema.TypeInt,
				Description:  "The ttl of the record",
				Optional:     true,
				Default:      3600,
				ValidateFunc: validation.IntBetween(60, 2592000),
			},
			"priority": {
				Type:         schema.TypeInt,
				Description:  "The priority of the record",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"geo_ip": {
				Type:          schema.TypeList,
				Description:   "Return record based on client localisation",
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"view", "http_service", "weighted"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"matches": {
							Type:        schema.TypeList,
							Description: "The list of matches",
							MinItems:    1,
							Required:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"countries": {
										Type:        schema.TypeList,
										Optional:    true,
										MinItems:    1,
										Description: "List of countries (eg: FR for France, US for the United States, GB for Great Britain...). List of all countries code: https://api.scaleway.com/domain-private/v2beta1/countries",
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(2, 2),
										},
									},
									"continents": {
										Type:        schema.TypeList,
										Optional:    true,
										MinItems:    1,
										Description: "List of continents (eg: EU for Europe, NA for North America, AS for Asia...). List of all continents code: https://api.scaleway.com/domain-private/v2beta1/continents",
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(2, 2),
										},
									},
									"data": {
										Type:        schema.TypeString,
										Description: "The data of the match result",
										Required:    true,
									},
								},
							},
						},
					},
				},
			},
			"http_service": {
				Type:          schema.TypeList,
				Description:   "Return record based on client localisation",
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"geo_ip", "view", "weighted"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ips": {
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.IsIPAddress,
							},
							Required:    true,
							MinItems:    1,
							Description: "IPs to check",
						},
						"must_contain": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Text to search",
						},
						"url": {
							Type:         schema.TypeString,
							ValidateFunc: validation.IsURLWithHTTPorHTTPS,
							Required:     true,
							Description:  "URL to match the must_contain text to validate an IP",
						},
						"user_agent": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "User-agent used when checking the URL",
						},
						"strategy": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Strategy to return an IP from the IPs list",
							ValidateFunc: validation.StringInSlice([]string{
								domain.RecordHTTPServiceConfigStrategyRandom.String(),
								domain.RecordHTTPServiceConfigStrategyHashed.String(),
							}, false),
						},
					},
				},
			},
			"view": {
				Type:          schema.TypeList,
				Description:   "Return record based on client subnet",
				Optional:      true,
				ConflictsWith: []string{"geo_ip", "http_service", "weighted"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet": {
							Type:         schema.TypeString,
							Description:  "The subnet of the view",
							Required:     true,
							ValidateFunc: validation.IsCIDR,
						},
						"data": {
							Type:        schema.TypeString,
							Description: "The data of the view record",
							Required:    true,
						},
					},
				},
			},
			"weighted": {
				Type:          schema.TypeList,
				Description:   "Return record based on weight",
				Optional:      true,
				ConflictsWith: []string{"geo_ip", "http_service", "view"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:         schema.TypeString,
							Description:  "The weighted IP",
							Required:     true,
							ValidateFunc: validation.IsIPAddress,
						},
						"weight": {
							Type:         schema.TypeInt,
							Description:  "The weight of the IP",
							Required:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
					},
				},
			},
			"project_id": projectIDSchema(),
		},
	}
}

func resourceScalewayDomainRecordCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	domainAPI := newDomainAPI(meta)

	dnsZone := d.Get("dns_zone").(string)
	geoIP, okGeoIP := d.GetOk("geo_ip")
	recordType := domain.RecordType(d.Get("type").(string))
	recordData := d.Get("data").(string)
	record := &domain.Record{
		Data:              recordData,
		Name:              d.Get("name").(string),
		TTL:               uint32(d.Get("ttl").(int)),
		Type:              recordType,
		Priority:          uint32(d.Get("priority").(int)),
		GeoIPConfig:       expandDomainGeoIPConfig(d.Get("data").(string), geoIP, okGeoIP),
		HTTPServiceConfig: expandDomainHTTPService(d.GetOk("http_service")),
		WeightedConfig:    expandDomainWeighted(d.GetOk("weighted")),
		ViewConfig:        expandDomainView(d.GetOk("view")),
		Comment:           nil,
	}
	_, err := domainAPI.UpdateDNSZoneRecords(&domain.UpdateDNSZoneRecordsRequest{
		DNSZone: dnsZone,
		Changes: []*domain.RecordChange{
			{
				Add: &domain.RecordChangeAdd{
					Records: []*domain.Record{record},
				},
			},
		},
		ReturnAllRecords: scw.BoolPtr(false),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	dnsZoneRes, err := waitForDNSZone(ctx, domainAPI, dnsZone, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Debug(ctx, fmt.Sprintf("DNS ZONE domain: %s subdomain: %s, status: %s",
		dnsZoneRes.Domain,
		dnsZoneRes.Subdomain,
		dnsZoneRes.Status))

	dnsZoneData, err := domainAPI.ListDNSZoneRecords(&domain.ListDNSZoneRecordsRequest{
		DNSZone: dnsZone,
		Name:    d.Get("name").(string),
		Type:    recordType,
	}, scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	currentRecord, err := getRecordFromTypeAndData(recordType, recordData, dnsZoneData.Records)
	if err != nil {
		return diag.FromErr(err)
	}

	recordID := fmt.Sprintf("%s/%s", dnsZone, currentRecord.ID)

	d.SetId(recordID)
	tflog.Debug(ctx, fmt.Sprintf("record ID[%s]", recordID))

	return resourceScalewayDomainRecordRead(ctx, d, meta)
}

func resourceScalewayDomainRecordRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	domainAPI := newDomainAPI(meta)
	var record *domain.Record
	var dnsZone string
	var projectID string

	currentData := d.Get("data")
	// check if this is an inline import. Like: "terraform import scaleway_domain_record.www subdomain.domain.tld/11111111-1111-1111-1111-111111111111"
	if strings.Contains(d.Id(), "/") {
		tab := strings.Split(d.Id(), "/")
		if len(tab) != 2 {
			return diag.FromErr(fmt.Errorf("cant parse record id: %s", d.Id()))
		}

		dnsZone = tab[0]
		recordID := tab[1]

		res, err := domainAPI.ListDNSZoneRecords(&domain.ListDNSZoneRecordsRequest{
			DNSZone: dnsZone,
			ID:      &recordID,
		}, scw.WithAllPages(), scw.WithContext(ctx))
		if err != nil {
			if is404Error(err) || is403Error(err) {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}

		if len(res.Records) > 0 {
			record = res.Records[0]
		}
	} else {
		dnsZone = d.Get("dns_zone").(string)

		recordTypeRaw, recordtTypeExist := d.GetOk("type")
		if !recordtTypeExist {
			return diag.FromErr(fmt.Errorf("record type not found"))
		}
		recordType := domain.RecordType(recordTypeRaw.(string))
		if recordType == domain.RecordTypeUnknown {
			return diag.FromErr(fmt.Errorf("record type unknow"))
		}

		idRecord := expandID(d.Id())
		res, err := domainAPI.ListDNSZoneRecords(&domain.ListDNSZoneRecordsRequest{
			DNSZone: dnsZone,
			Name:    d.Get("name").(string),
			Type:    recordType,
			ID:      &idRecord,
		}, scw.WithAllPages(), scw.WithContext(ctx))
		if err != nil {
			if is404Error(err) || is403Error(err) {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}

		if len(res.Records) > 0 {
			record = res.Records[0]
		}
	}

	if record == nil {
		d.SetId("")
		return nil
	}

	res, err := waitForDNSZone(ctx, domainAPI, dnsZone, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if is404Error(err) || is403Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	projectID = res.ProjectID
	_ = d.Set("root_zone", res.Subdomain == "")

	// retrieve data from record
	if len(currentData.(string)) == 0 {
		currentData = flattenDomainData(record.Data, record.Type).(string)
	}

	d.SetId(record.ID)
	_ = d.Set("dns_zone", dnsZone)
	_ = d.Set("name", record.Name)
	_ = d.Set("type", record.Type.String())
	_ = d.Set("data", currentData.(string))
	_ = d.Set("ttl", int(record.TTL))
	_ = d.Set("priority", int(record.Priority))
	_ = d.Set("geo_ip", flattenDomainGeoIP(record.GeoIPConfig))
	_ = d.Set("http_service", flattenDomainHTTPService(record.HTTPServiceConfig))
	_ = d.Set("weighted", flattenDomainWeighted(record.WeightedConfig))
	_ = d.Set("view", flattenDomainView(record.ViewConfig))
	_ = d.Set("project_id", projectID)

	return nil
}

func resourceScalewayDomainRecordUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	domainAPI := newDomainAPI(meta)

	req := &domain.UpdateDNSZoneRecordsRequest{
		DNSZone:          d.Get("dns_zone").(string),
		ReturnAllRecords: scw.BoolPtr(false),
	}

	record := &domain.Record{Name: d.Get("name").(string)}
	hasChange := false
	if d.HasChanges("geo_ip") {
		if geoIP, ok := d.GetOk("geo_ip"); ok {
			record.GeoIPConfig = expandDomainGeoIPConfig(d.Get("data").(string), geoIP, ok)
		}
		hasChange = true
	}

	if d.HasChange("name") {
		record.Name = d.Get("name").(string)
		hasChange = true
	}

	if d.HasChange("data") {
		record.Data = d.Get("data").(string)
		hasChange = true
	}

	if d.HasChange("priority") {
		record.Priority = uint32(d.Get("priority").(int))
		hasChange = true
	}

	if d.HasChange("ttl") {
		record.TTL = uint32(d.Get("ttl").(int))
		hasChange = true
	}

	if d.HasChange("type") {
		record.Type = domain.RecordType(d.Get("type").(string))
		hasChange = true
	}

	if d.HasChanges("http_service") {
		record.HTTPServiceConfig = expandDomainHTTPService(d.GetOk("http_service"))
		hasChange = true
	}

	if d.HasChanges("weighted") {
		record.WeightedConfig = expandDomainWeighted(d.GetOk("weighted"))
		hasChange = true
	}

	if d.HasChanges("view") {
		record.ViewConfig = expandDomainView(d.GetOk("view"))
		hasChange = true
	}

	req.Changes = []*domain.RecordChange{
		{
			Set: &domain.RecordChangeSet{
				ID:      scw.StringPtr(expandID(d.Id())),
				Records: []*domain.Record{record},
			},
		},
	}

	if hasChange || d.HasChanges("dns_zone", "keep_empty_zone") {
		_, err := domainAPI.UpdateDNSZoneRecords(req)
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = waitForDNSZone(ctx, domainAPI, d.Get("dns_zone").(string), d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceScalewayDomainRecordRead(ctx, d, meta)
}

func resourceScalewayDomainRecordDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	domainAPI := newDomainAPI(meta)

	recordID := expandID(d.Id())
	_, err := domainAPI.UpdateDNSZoneRecords(&domain.UpdateDNSZoneRecordsRequest{
		DNSZone: d.Get("dns_zone").(string),
		Changes: []*domain.RecordChange{
			{
				Delete: &domain.RecordChangeDelete{
					ID: &recordID,
				},
			},
		},
		ReturnAllRecords: scw.BoolPtr(false),
	})
	if err != nil {
		d.SetId("")
		return diag.FromErr(err)
	}
	d.SetId("")

	_, err = waitForDNSZone(ctx, domainAPI, d.Get("dns_zone").(string), d.Timeout(schema.TimeoutDelete))
	if err != nil && !errorCheck(err, domain.ErrCodeNoSuchDNSZone) {
		if is404Error(err) || is403Error(err) {
			return nil
		}
		return diag.FromErr(err)
	}

	// for non-root zone, if the zone have only NS records, then delete the zone
	if d.Get("keep_empty_zone").(bool) || d.Get("root_zone").(bool) {
		return nil
	}

	res, err := domainAPI.ListDNSZoneRecords(&domain.ListDNSZoneRecordsRequest{
		DNSZone: d.Get("dns_zone").(string),
	})
	if err != nil {
		if is404Error(err) || is403Error(err) {
			return nil
		}
		return diag.FromErr(err)
	}

	for _, r := range res.Records {
		if r.Type != domain.RecordTypeNS {
			// The zone isn't empty, keep it
			return nil
		}
		tflog.Debug(ctx, fmt.Sprintf("record [%s], type [%s]", r.Name, r.Type))
	}

	_, err = waitForDNSZone(ctx, domainAPI, d.Get("dns_zone").(string), d.Timeout(schema.TimeoutDelete))
	if err != nil {
		if errorCheck(err, domain.ErrCodeNoSuchDNSZone) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("failed to wait for dns zone before deleting: %w", err))
	}

	_, err = domainAPI.DeleteDNSZone(&domain.DeleteDNSZoneRequest{
		DNSZone:   d.Get("dns_zone").(string),
		ProjectID: d.Get("project_id").(string),
	})

	if err != nil {
		if is404Error(err) || is403Error(err) {
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}
