package domain

import (
	"context"
	"fmt"
	// "time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
)

func ResourceDomainsRegistration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDomainsRegistrationCreate,
		ReadContext:   resourceDomainsRegistrationsRead,
		UpdateContext: resourceDomainsRegistrationUpdate,
		DeleteContext: resourceDomainsRegistrationDelete,
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultDomainRegistrationTimeout),
			Read:    schema.DefaultTimeout(defaultDomainRegistrationTimeout),
			Update:  schema.DefaultTimeout(defaultDomainRegistrationTimeout),
			Delete:  schema.DefaultTimeout(defaultDomainRegistrationTimeout),
			Default: schema.DefaultTimeout(defaultDomainRegistrationTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"domain_names": {
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of domain names to be managed.",
			},
			"duration_in_years": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"project_id": account.ProjectIDSchema(),

			"owner_contact_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ExactlyOneOf: []string{
					"owner_contact_id",
					"owner_contact",
				},
				Description: "ID of the owner contact. Either `owner_contact_id` or `owner_contact` must be provided.",
			},
			"owner_contact": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				ExactlyOneOf: []string{
					"owner_contact_id",
					"owner_contact",
				},
				Elem: &schema.Resource{
					Schema: contactSchema(),
				},
				Description: "Details of the owner contact. Either `owner_contact_id` or `owner_contact` must be provided.",
			},
			"administrative_contact": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: contactSchema(),
				},
				Description: "Details of the administrative contact.",
			},
			"technical_contact": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: contactSchema(),
				},
				Description: "Details of the technical contact.",
			},
			"auto_renew": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable or disable auto-renewal of the domain.",
			},
			"dnssec": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable or disable dnssec for the domain.",
			},
			"ds_record": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key_id": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The identifier for the dnssec key.",
						},
						"algorithm": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The algorithm used for dnssec (e.g., rsasha256, ecdsap256sha256).",
						},
						"digest": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The type of digest (e.g., sha_1, sha_256).",
									},
									"digest": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The digest value.",
									},
									"public_key": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "The public key value.",
												},
											},
										},
										Description: "The public key associated with the digest.",
									},
								},
							},
							Description: "Details about the digest.",
						},
						"public_key": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The public key value.",
									},
								},
							},
							Description: "Public key associated with the dnssec record.",
						},
					},
				},
				Description: "dnssec DS record configuration.",
			},
			"is_external": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Indicates whether Scaleway is the domain's registrar.",
			},
		},
	}
}

func contactSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"legal_form": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Legal form of the contact (e.g., 'individual' or 'organization').",
		},
		"firstname": {
			Type:             schema.TypeString,
			Required:         true,
			DiffSuppressFunc: dsf.IgnoreCase,
			Description:      "First name of the contact.",
		},
		"lastname": {
			Type:             schema.TypeString,
			Required:         true,
			DiffSuppressFunc: dsf.IgnoreCase,
			Description:      "Last name of the contact.",
		},
		"company_name": {
			Type:             schema.TypeString,
			Optional:         true,
			DiffSuppressFunc: dsf.IgnoreCase,
			Description:      "Name of the company associated with the contact (if applicable).",
		},
		"email": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Primary email address of the contact.",
		},
		"email_alt": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Alternative email address for the contact.",
		},
		"phone_number": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Primary phone number of the contact.",
		},
		"fax_number": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Fax number for the contact (if available).",
		},
		"address_line_1": {
			Type:             schema.TypeString,
			Required:         true,
			DiffSuppressFunc: dsf.IgnoreCase,
			Description:      "Primary address line for the contact.",
		},
		"address_line_2": {
			Type:             schema.TypeString,
			Optional:         true,
			DiffSuppressFunc: dsf.IgnoreCase,
			Description:      "Secondary address line for the contact (optional).",
		},
		"zip": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Postal code of the contact's address.",
		},
		"city": {
			Type:             schema.TypeString,
			Required:         true,
			DiffSuppressFunc: dsf.IgnoreCase,
			Description:      "City of the contact's address.",
		},
		"country": {
			Type:             schema.TypeString,
			Required:         true,
			DiffSuppressFunc: dsf.IgnoreCase,
			Description:      "Country code of the contact's address (ISO format).",
		},
		"vat_identification_code": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "VAT identification code of the contact, if applicable.",
		},
		"company_identification_code": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Company identification code (e.g., SIREN/SIRET in France) for the contact.",
		},
		"lang": {
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			DiffSuppressFunc: dsf.IgnoreCase,
			Description:      "Preferred language of the contact (e.g., 'en_US', 'fr_FR').",
		},
		"resale": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Indicates if the contact is used for resale purposes.",
		},
		"extension_fr": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			MaxItems: 1, // Ensure it's a single-item list if needed
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"mode": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Mode of the French extension (e.g., 'individual', 'duns', 'association', etc.).",
					},
					"individual_info": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"whois_opt_in": {
									Type:        schema.TypeBool,
									Optional:    true,
									Description: "Whether the individual contact has opted into WHOIS publishing.",
								},
							},
						},
						Description: "Information about the individual registration for French domains.",
					},
					"duns_info": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"duns_id": {
									Type:        schema.TypeString,
									Optional:    true,
									Description: "DUNS ID associated with the domain owner (for French domains).",
								},
								"local_id": {
									Type:        schema.TypeString,
									Optional:    true,
									Description: "Local identifier of the domain owner (for French domains).",
								},
							},
						},
						Description: "DUNS information for the domain owner (specific to French domains).",
					},
					"association_info": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"publication_jo": {
									Type:        schema.TypeString,
									Optional:    true,
									Description: "Publication date in the Official Journal (RFC3339 format) for association information.",
								},
								"publication_jo_page": {
									Type:        schema.TypeInt,
									Optional:    true,
									Description: "Page number of the publication in the Official Journal for association information.",
								},
							},
						},
						Description: "Association-specific information for the domain (French extension).",
					},
					"trademark_info": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"trademark_inpi": {
									Type:        schema.TypeString,
									Optional:    true,
									Description: "Trademark information from INPI (French extension).",
								},
							},
						},
						Description: "Trademark-related information for the domain (French extension).",
					},
					"code_auth_afnic_info": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"code_auth_afnic": {
									Type:        schema.TypeString,
									Optional:    true,
									Description: "AFNIC authorization code for the contact (specific to French domains).",
								},
							},
						},
						Description: "AFNIC authorization information for the contact (French extension).",
					},
				},
			},
			Description: "Details specific to French domain extensions.",
		},
		"extension_eu": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"european_citizenship": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Indicates the European citizenship of the contact.",
					},
				},
			},
			Description: "Details specific to European domain extensions.",
		},
		"whois_opt_in": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Indicates whether the contact has opted into WHOIS publishing.",
		},
		"state": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "State or region of the contact.",
		},
		"extension_nl": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			Elem: &schema.Schema{
				Type:        schema.TypeString,
				Description: "Additional extension field specific to Dutch domains.",
			},
			Description: "Extension details specific to Dutch domain registrations.",
		},
	}
}

func resourceDomainsRegistrationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	registrarAPI := NewRegistrarDomainAPI(m)

	projectID := d.Get("project_id").(string)
	domainNames := make([]string, 0)
	for _, v := range d.Get("domain_names").([]interface{}) {
		domainNames = append(domainNames, v.(string))
	}
	durationInYears := uint32(d.Get("duration_in_years").(int))

	buyDomainsRequest := &domain.RegistrarAPIBuyDomainsRequest{
		Domains:         domainNames,
		DurationInYears: durationInYears,
		ProjectID:       projectID,
	}

	ownerContactID := d.Get("owner_contact_id").(string)
	if ownerContactID != "" {
		buyDomainsRequest.OwnerContactID = &ownerContactID
	} else if ownerContacts, ok := d.GetOk("owner_contact"); ok {
		contacts := ownerContacts.([]interface{})
		if len(contacts) > 0 {
			buyDomainsRequest.OwnerContact = ExpandNewContact(contacts[0].(map[string]interface{}))
		}
	}

	resp, err := registrarAPI.BuyDomains(buyDomainsRequest, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	err = waitForTaskCompletion(ctx, registrarAPI, resp.TaskID, 3600)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, domainName := range domainNames {
		_, err = waitForDomainsRegistration(ctx, registrarAPI, domainName, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(projectID + "/" + resp.TaskID)

	return resourceDomainsRegistrationsRead(ctx, d, m)
}

func resourceDomainsRegistrationsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	registrarAPI := NewRegistrarDomainAPI(m)
	id := d.Id()

	domainNames, err := ExtractDomainsFromTaskID(ctx, id, registrarAPI)
	if err != nil {
		return diag.FromErr(err)
	}
	if len(domainNames) == 0 {
		d.SetId("")
		return nil
	}

	firstDomain := domainNames[0]
	firstResp, err := registrarAPI.GetDomain(&domain.RegistrarAPIGetDomainRequest{
		Domain: firstDomain,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	var computedOwnerContactID string
	if firstResp.OwnerContact != nil {
		computedOwnerContactID = firstResp.OwnerContact.ID
	}

	computedOwnerContact := flattenContact(firstResp.OwnerContact)
	computedAdministrativeContact := flattenContact(firstResp.AdministrativeContact)
	computedTechnicalContact := flattenContact(firstResp.TechnicalContact)

	computedAutoRenew := false
	if firstResp.AutoRenewStatus.String() == "enabled" {
		computedAutoRenew = true
	}

	computedDnssec := false

	if firstResp.Dnssec.Status == "enabled" {
		computedDnssec = true
	}

	var computedDSRecord []interface{}
	if firstResp.Dnssec != nil {
		computedDSRecord = FlattenDSRecord(firstResp.Dnssec.DsRecords)
	} else {
		computedDSRecord = []interface{}{}
	}

	computedIsExternal := firstResp.IsExternal

	_ = d.Set("domain_names", domainNames)
	_ = d.Set("owner_contact_id", computedOwnerContactID)
	_ = d.Set("owner_contact", computedOwnerContact)
	_ = d.Set("administrative_contact", computedAdministrativeContact)
	_ = d.Set("technical_contact", computedTechnicalContact)
	_ = d.Set("auto_renew", computedAutoRenew)
	_ = d.Set("dnssec", computedDnssec)
	_ = d.Set("ds_record", computedDSRecord)
	_ = d.Set("is_external", computedIsExternal)

	return nil
}

func resourceDomainsRegistrationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	registrarAPI := NewRegistrarDomainAPI(m)
	id := d.Id()

	domainNames, err := ExtractDomainsFromTaskID(ctx, id, registrarAPI)
	if err != nil {
		return diag.FromErr(err)
	}
	if len(domainNames) == 0 {
		d.SetId("")
		return nil
	}

	if d.HasChange("owner_contact_id") || d.HasChange("owner_contact") {
		return diag.FromErr(fmt.Errorf("the domain ownership transfer feature is not implemented in this provider because it requires manual validation through email notifications. This action can only be performed via the Scaleway Console"))
	}

	if d.HasChange("auto_renew") {
		newAutoRenew := d.Get("auto_renew").(bool)
		for _, domainName := range domainNames {
			domainResp, err := registrarAPI.GetDomain(&domain.RegistrarAPIGetDomainRequest{
				Domain: domainName,
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(fmt.Errorf("failed to get domain details for %s: %v", domainName, err))
			}

			if newAutoRenew {
				if domainResp.AutoRenewStatus != domain.DomainFeatureStatusEnabled && domainResp.AutoRenewStatus != domain.DomainFeatureStatusEnabling {
					_, err = registrarAPI.EnableDomainAutoRenew(&domain.RegistrarAPIEnableDomainAutoRenewRequest{
						Domain: domainName,
					}, scw.WithContext(ctx))
					if err != nil {
						return diag.FromErr(fmt.Errorf("failed to enable auto-renew for %s: %v", domainName, err))
					}
				}
			} else {
				if domainResp.AutoRenewStatus == domain.DomainFeatureStatusEnabled || domainResp.AutoRenewStatus == domain.DomainFeatureStatusEnabling {
					_, err = registrarAPI.DisableDomainAutoRenew(&domain.RegistrarAPIDisableDomainAutoRenewRequest{
						Domain: domainName,
					}, scw.WithContext(ctx))
					if err != nil {
						return diag.FromErr(fmt.Errorf("failed to disable auto-renew for %s: %v", domainName, err))
					}
				}
			}
			_, err = waitForAutoRenewStatus(ctx, registrarAPI, domainName, d.Timeout(schema.TimeoutUpdate))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("dnssec") {
		newDnssec := d.Get("dnssec").(bool)
		for _, domainName := range domainNames {
			domainResp, err := registrarAPI.GetDomain(&domain.RegistrarAPIGetDomainRequest{
				Domain: domainName,
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(fmt.Errorf("failed to get domain details for %s: %v", domainName, err))
			}

			if newDnssec {
				var dsRecord *domain.DSRecord
				if v, ok := d.GetOk("ds_record"); ok {
					dsRecordList := v.([]interface{})
					if len(dsRecordList) > 0 && dsRecordList[0] != nil {
						dsRecord = ExpandDSRecord(dsRecordList)
					}
				}
				_, err = registrarAPI.EnableDomainDNSSEC(&domain.RegistrarAPIEnableDomainDNSSECRequest{
					Domain:   domainName,
					DsRecord: dsRecord,
				}, scw.WithContext(ctx))
				if err != nil {
					return diag.FromErr(fmt.Errorf("failed to enable dnssec for %s: %v", domainName, err))
				}
			} else {
				if domainResp.Dnssec != nil && domainResp.Dnssec.Status == "enabled" {
					_, err = registrarAPI.DisableDomainDNSSEC(&domain.RegistrarAPIDisableDomainDNSSECRequest{
						Domain: domainName,
					}, scw.WithContext(ctx))
					if err != nil {
						return diag.FromErr(fmt.Errorf("failed to disable dnssec for %s: %v", domainName, err))
					}
				}
			}
			_, err = waitForDNSSECStatus(ctx, registrarAPI, domainName, d.Timeout(schema.TimeoutUpdate))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return resourceDomainsRegistrationsRead(ctx, d, m)
}

func resourceDomainsRegistrationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	registrarAPI := NewRegistrarDomainAPI(m)
	id := d.Id()

	domainNames, err := ExtractDomainsFromTaskID(ctx, id, registrarAPI)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, domainName := range domainNames {
		domainResp, err := registrarAPI.GetDomain(&domain.RegistrarAPIGetDomainRequest{
			Domain: domainName,
		}, scw.WithContext(ctx))
		if err != nil {
			if httperrors.Is404(err) {
				continue
			}
			return diag.FromErr(fmt.Errorf("failed to get domain details for %s: %v", domainName, err))
		}

		if domainResp.AutoRenewStatus == domain.DomainFeatureStatusEnabled ||
			domainResp.AutoRenewStatus == domain.DomainFeatureStatusEnabling {
			_, err = registrarAPI.DisableDomainAutoRenew(&domain.RegistrarAPIDisableDomainAutoRenewRequest{
				Domain: domainName,
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(fmt.Errorf("failed to disable auto-renew for %s: %v", domainName, err))
			}
		}
	}

	d.SetId("")

	return nil
}
