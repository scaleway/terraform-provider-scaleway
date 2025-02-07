package domain

import (
	"context"
	"fmt"
	"time"

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
			"administrative_contact_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"administrative_contact": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: contactSchema(),
				},
				Description: "Details of the administrative contact.",
			},
			"technical_contact_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"technical_contact": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
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

			// Computed fields populated by the Read function
			"auto_renew_status": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Map associating each domain with its auto-renewal status.",
			},
			"dnssec_status": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Map associating each domain with its dnssec status.",
			},
			"epp_code": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeList,
					Elem: &schema.Schema{Type: schema.TypeString},
				},
				Description: "Map associating each domain with the list of its EPP codes.",
			},
			"expired_at": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Map associating each domain with its expiration date (RFC3339).",
			},
			"updated_at": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Map associating each domain with its last updated date (RFC3339).",
			},
			"registrar": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Map associating each domain with the registrar managing it.",
			},
			"status": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Map associating each domain with its status.",
			},
			"organization_id": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Map associating each domain with the ID of its associated organization.",
			},
			"pending_trade": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeBool},
				Description: "Map associating each domain with the pending trade indicator.",
			},
			"external_domain_registration_status": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeList,
					Elem: &schema.Schema{Type: schema.TypeString},
				},
				Description: "Map associating each domain with its external registration status (if applicable).",
			},
			"transfer_registration_status": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeList,
					Elem: &schema.Schema{Type: schema.TypeString},
				},
				Description: "Map associating each domain with its transfer status (if applicable).",
			},
			"linked_products": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeList,
					Elem: &schema.Schema{Type: schema.TypeString},
				},
				Description: "Map associating each domain with the list of linked Scaleway products.",
			},
			"tld": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeList,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "TLD name.",
							},
							"dnssec_support": {
								Type:        schema.TypeBool,
								Computed:    true,
								Description: "Indicates whether the TLD supports dnssec.",
							},
							"duration_in_years_min": {
								Type:        schema.TypeInt,
								Computed:    true,
								Description: "Minimum registration duration (in years) for the TLD.",
							},
							"duration_in_years_max": {
								Type:        schema.TypeInt,
								Computed:    true,
								Description: "Maximum registration duration (in years) for the TLD.",
							},
							"idn_support": {
								Type:        schema.TypeBool,
								Computed:    true,
								Description: "Indicates whether the TLD supports internationalized domain names (IDN).",
							},
							"offers": {
								Type:     schema.TypeList,
								Computed: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"action": {
											Type:        schema.TypeString,
											Computed:    true,
											Description: "Offer action type (e.g., create, transfer).",
										},
										"operation_path": {
											Type:        schema.TypeString,
											Computed:    true,
											Description: "Path of the operation associated with the offer.",
										},
										"price": {
											Type:        schema.TypeMap,
											Computed:    true,
											Elem:        &schema.Schema{Type: schema.TypeString},
											Description: "Pricing information for the offer.",
										},
									},
								},
								Description: "List of offers available for the TLD.",
							},
							"specifications": {
								Type:        schema.TypeMap,
								Computed:    true,
								Elem:        &schema.Schema{Type: schema.TypeString},
								Description: "Additional TLD specifications (e.g., grace period).",
							},
						},
					},
				},
				Description: "Map associating each domain with its detailed TLD information.",
			},
			"dns_zones": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeList,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"domain": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "DNS zone domain name.",
							},
							"subdomain": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "DNS zone subdomain.",
							},
							"ns": {
								Type:        schema.TypeList,
								Computed:    true,
								Elem:        &schema.Schema{Type: schema.TypeString},
								Description: "List of name servers (NS) for the DNS zone.",
							},
							"ns_default": {
								Type:        schema.TypeList,
								Computed:    true,
								Elem:        &schema.Schema{Type: schema.TypeString},
								Description: "List of default name servers for the DNS zone.",
							},
							"ns_master": {
								Type:        schema.TypeList,
								Computed:    true,
								Elem:        &schema.Schema{Type: schema.TypeString},
								Description: "List of master name servers for the DNS zone.",
							},
							"status": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "DNS zone status.",
							},
							"message": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "Additional DNS zone message.",
							},
							"updated_at": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "Last updated date of the DNS zone.",
							},
							"project_id": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "Project ID associated with the DNS zone.",
							},
						},
					},
				},
				Description: "Map associating each domain with its corresponding DNS zones.",
			},
		},
	}
}

// doc = https://developer.hashicorp.com/terraform/language/expressions/dynamic-blocks
// add description
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
		},
		"lastname": {
			Type:             schema.TypeString,
			Required:         true,
			DiffSuppressFunc: dsf.IgnoreCase,
		},
		"company_name": {
			Type:             schema.TypeString,
			Optional:         true,
			DiffSuppressFunc: dsf.IgnoreCase,
		},
		"email": {
			Type:     schema.TypeString,
			Required: true,
		},
		"email_alt": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"phone_number": {
			Type:     schema.TypeString,
			Required: true,
		},
		"fax_number": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"address_line_1": {
			Type:             schema.TypeString,
			Required:         true,
			DiffSuppressFunc: dsf.IgnoreCase,
		},
		"address_line_2": {
			Type:             schema.TypeString,
			Optional:         true,
			DiffSuppressFunc: dsf.IgnoreCase,
		},
		"zip": {
			Type:     schema.TypeString,
			Required: true,
		},
		"city": {
			Type:             schema.TypeString,
			Required:         true,
			DiffSuppressFunc: dsf.IgnoreCase,
		},
		"country": {
			Type:             schema.TypeString,
			Required:         true,
			DiffSuppressFunc: dsf.IgnoreCase,
		},
		"vat_identification_code": {
			Type:     schema.TypeString,
			Required: true,
		},
		"company_identification_code": {
			Type:     schema.TypeString,
			Required: true,
		},
		"lang": {
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			DiffSuppressFunc: dsf.IgnoreCase,
		},
		"resale": {
			Type:     schema.TypeBool,
			Optional: true,
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
						Description: "Mode of the French extension.",
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
									Description: "Whether WHOIS opt-in is enabled.",
								},
							},
						},
						Description: "Information about the individual owning the domain.",
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
									Description: "DUNS ID associated with the domain owner.",
								},
								"local_id": {
									Type:        schema.TypeString,
									Optional:    true,
									Description: "Local ID of the domain owner.",
								},
							},
						},
						Description: "DUNS information for the domain owner.",
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
									Description: "Publication date in the Official Journal (RFC3339 format).",
								},
								"publication_jo_page": {
									Type:        schema.TypeInt,
									Optional:    true,
									Description: "Page number of the publication in the Official Journal.",
								},
							},
						},
						Description: "Association-specific information for the domain.",
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
									Description: "Trademark information from INPI.",
								},
							},
						},
						Description: "Trademark-related information for the domain.",
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
									Description: "Authorization code from AFNIC.",
								},
							},
						},
						Description: "Information regarding AFNIC authorization.",
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
			Type:     schema.TypeBool,
			Optional: true,
		},
		"state": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"extension_nl": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
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

	// Extract the list of domain names from the TaskID.
	domainNames, err := ExtractDomainsFromTaskID(ctx, id, registrarAPI)
	if err != nil {
		return diag.FromErr(err)
	}

	// Prepare maps for each computed field.
	autoRenewStatus := make(map[string]string)
	dnssecStatus := make(map[string]string)
	eppCode := make(map[string][]string)
	expiredAt := make(map[string]string)
	updatedAt := make(map[string]string)
	registrar := make(map[string]string)
	statusMap := make(map[string]string)
	organizationID := make(map[string]string)
	pendingTrade := make(map[string]bool)
	externalDomainRegistrationStatus := make(map[string][]string)
	transferRegistrationStatus := make(map[string][]string)
	linkedProducts := make(map[string][]string)
	tld := make(map[string][]map[string]interface{})
	dnsZones := make(map[string][]map[string]interface{})

	// Iterate through each domain and populate the maps.
	for _, domainName := range domainNames {
		domainResp, err := registrarAPI.GetDomain(&domain.RegistrarAPIGetDomainRequest{
			Domain: domainName,
		}, scw.WithContext(ctx))
		if err != nil {
			if httperrors.Is404(err) {
				// If the domain no longer exists, remove it from the state.
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}

		// Auto-renew status.
		autoRenewStatus[domainName] = domainResp.AutoRenewStatus.String()

		// Dnssec status.
		if domainResp.Dnssec != nil {
			dnssecStatus[domainName] = domainResp.Dnssec.Status.String()
		} else {
			dnssecStatus[domainName] = ""
		}

		// EPP codes (assuming domainResp.EppCode is []string).
		eppCode[domainName] = domainResp.EppCode

		// Expiration date.
		if domainResp.ExpiredAt != nil {
			expiredAt[domainName] = domainResp.ExpiredAt.Format(time.RFC3339)
		} else {
			expiredAt[domainName] = ""
		}

		// Updated date.
		if domainResp.UpdatedAt != nil {
			updatedAt[domainName] = domainResp.UpdatedAt.Format(time.RFC3339)
		} else {
			updatedAt[domainName] = ""
		}

		registrar[domainName] = domainResp.Registrar
		statusMap[domainName] = string(domainResp.Status)
		organizationID[domainName] = domainResp.OrganizationID
		pendingTrade[domainName] = domainResp.PendingTrade

		// External registration status.
		if domainResp.ExternalDomainRegistrationStatus != nil {
			externalDomainRegistrationStatus[domainName] = flattenExternalDomainRegistrationStatus(domainResp.ExternalDomainRegistrationStatus)
		} else {
			externalDomainRegistrationStatus[domainName] = []string{}
		}

		// Transfer registration status.
		if domainResp.TransferRegistrationStatus != nil {
			transferRegistrationStatus[domainName] = flattenDomainRegistrationStatusTransfer(domainResp.TransferRegistrationStatus)
		} else {
			transferRegistrationStatus[domainName] = []string{}
		}

		// Linked products.
		var linkedProductsStr []string
		for _, lp := range domainResp.LinkedProducts {
			linkedProductsStr = append(linkedProductsStr, lp.String())
		}
		linkedProducts[domainName] = linkedProductsStr

		// TLD details (assuming flattenTLD returns []map[string]interface{}).
		if domainResp.Tld != nil {
			tld[domainName] = flattenTLD(domainResp.Tld)
		} else {
			tld[domainName] = []map[string]interface{}{}
		}

		// DNS zones (assuming flattenDNSZones returns []map[string]interface{}).
		if len(domainResp.DNSZones) > 0 {
			dnsZones[domainName] = flattenDNSZones(domainResp.DNSZones)
		} else {
			dnsZones[domainName] = []map[string]interface{}{}
		}
	}

	// Set computed fields (without checking the return values).
	_ = d.Set("auto_renew_status", autoRenewStatus)
	_ = d.Set("dnssec_status", dnssecStatus)
	_ = d.Set("epp_code", eppCode)
	_ = d.Set("expired_at", expiredAt)
	_ = d.Set("updated_at", updatedAt)
	_ = d.Set("registrar", registrar)
	_ = d.Set("status", statusMap)
	_ = d.Set("organization_id", organizationID)
	_ = d.Set("pending_trade", pendingTrade)
	_ = d.Set("external_domain_registration_status", externalDomainRegistrationStatus)
	_ = d.Set("transfer_registration_status", transferRegistrationStatus)
	_ = d.Set("linked_products", linkedProducts)
	_ = d.Set("tld", tld)
	_ = d.Set("dns_zones", dnsZones)

	return nil
}

func resourceDomainsRegistrationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	registrarAPI := NewRegistrarDomainAPI(m)
	id := d.Id()
	domainName, err := ExtractDomainFromID(id)
	if err != nil {
		return diag.FromErr(err)
	}
	if d.HasChange("owner_contact_id") || d.HasChange("owner_contact") {
		return diag.FromErr(fmt.Errorf("the domain ownership transfer feature is not implemented in this provider because it requires manual validation through email notifications. This action can only be performed via the Scaleway Console"))
	}

	hasChanges := false
	updateRequest := &domain.RegistrarAPIUpdateDomainRequest{
		Domain: domainName,
	}

	if d.HasChange("administrative_contact_id") {
		administrativeContactID := d.Get("administrative_contact_id").(string)
		updateRequest.AdministrativeContactID = &administrativeContactID
		hasChanges = true
	}

	if d.HasChange("technical_contact_id") {
		technicalContactID := d.Get("technical_contact_id").(string)
		updateRequest.TechnicalContactID = &technicalContactID
		hasChanges = true
	}

	if d.HasChange("administrative_contact") {
		if adminContacts, ok := d.GetOk("administrative_contact"); ok {
			contacts := adminContacts.([]interface{})
			if len(contacts) > 0 {
				updateRequest.AdministrativeContact = ExpandNewContact(contacts[0].(map[string]interface{}))
				hasChanges = true
			}
		}
	}

	if d.HasChange("technical_contact") {
		if techContacts, ok := d.GetOk("technical_contact"); ok {
			contacts := techContacts.([]interface{})
			if len(contacts) > 0 {
				updateRequest.TechnicalContact = ExpandNewContact(contacts[0].(map[string]interface{}))
				hasChanges = true
			}
		}
	}

	if hasChanges {
		_, err = registrarAPI.UpdateDomain(updateRequest, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		tasks, err := waitForUpdateDomainTaskCompletion(ctx, registrarAPI, domainName, 3600)
		if err != nil {
			return diag.FromErr(err)
		}
		fmt.Printf("coucou")
		_ = tasks
	}

	return resourceDomainsRegistrationsRead(ctx, d, m)
}

func resourceDomainsRegistrationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	registrarAPI := NewRegistrarDomainAPI(m)
	id := d.Id()

	// Récupère la liste de noms de domaines à partir du TaskID
	domainNames, err := ExtractDomainsFromTaskID(ctx, id, registrarAPI)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, domainName := range domainNames {
		// Récupère les détails du domaine
		domainResp, err := registrarAPI.GetDomain(&domain.RegistrarAPIGetDomainRequest{
			Domain: domainName,
		}, scw.WithContext(ctx))
		if err != nil {
			// Si le domaine n'existe plus, on passe au suivant
			if httperrors.Is404(err) {
				continue
			}
			return diag.FromErr(fmt.Errorf("failed to get domain details for %s: %v", domainName, err))
		}

		// Désactive l’auto-renew si nécessaire
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

	// On supprime la ressource (vider l'ID Terraform)
	d.SetId("")

	return nil
}
