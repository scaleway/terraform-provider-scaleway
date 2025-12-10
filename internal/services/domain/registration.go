package domain

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
)

func ResourceRegistration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRegistrationCreate,
		ReadContext:   resourceRegistrationsRead,
		UpdateContext: resourceRegistrationUpdate,
		DeleteContext: resourceRegistrationDelete,
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
		SchemaFunc:    registrationSchema,
	}
}

func registrationSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"domain_names": {
			Type:        schema.TypeList,
			Required:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "List of domain names to be managed.",
		},
		"duration_in_years": {
			Type:        schema.TypeInt,
			Description: "Duration of the registration period in years.",
			Optional:    true,
			Default:     1,
		},
		"owner_contact_id": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ExactlyOneOf: []string{
				"owner_contact_id",
				"owner_contact",
			},
			ValidateFunc: validation.IsUUID,
			Description:  "ID of the owner contact. Either `owner_contact_id` or `owner_contact` must be provided.",
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
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"key_id": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "The identifier for the dnssec key.",
					},
					"algorithm": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The algorithm used for dnssec (e.g., rsasha256, ecdsap256sha256).",
					},
					"digest": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"type": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "The digest type for the DS record (e.g., sha_1, sha_256, gost_r_34_11_94, sha_384).",
								},
								"digest": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "The digest value.",
								},
								"public_key": {
									Type:     schema.TypeList,
									Computed: true,
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
						Computed: true,
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
		"project_id": account.ProjectIDSchema(),

		"task_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "ID of the task that created the domain.",
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
			MaxItems: 1,
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

func resourceRegistrationCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	registrarAPI := NewRegistrarDomainAPI(m)

	projectID := d.Get("project_id").(string)

	domainNames := make([]string, 0)
	for _, v := range d.Get("domain_names").([]any) {
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
		contacts := ownerContacts.([]any)
		if len(contacts) > 0 {
			buyDomainsRequest.OwnerContact = ExpandNewContact(contacts[0].(map[string]any))
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

	newDnssec := d.Get("dnssec").(bool)

	if newDnssec {
		for _, domainName := range domainNames {
			_, err = registrarAPI.EnableDomainDNSSEC(&domain.RegistrarAPIEnableDomainDNSSECRequest{
				Domain: domainName,
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}

			_, err = waitForDNSSECStatus(ctx, registrarAPI, domainName, d.Timeout(schema.TimeoutCreate))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	d.SetId(projectID + "/" + resp.TaskID)

	return resourceRegistrationsRead(ctx, d, m)
}

func resourceRegistrationsRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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

	computedAutoRenew := firstResp.AutoRenewStatus == domain.DomainFeatureStatusEnabled

	computedDnssec := firstResp.Dnssec.Status == domain.DomainFeatureStatusEnabled

	var computedDSRecord []any
	if firstResp.Dnssec != nil {
		computedDSRecord = FlattenDSRecord(firstResp.Dnssec.DsRecords)
	} else {
		computedDSRecord = []any{}
	}

	_ = d.Set("domain_names", domainNames)
	_ = d.Set("owner_contact_id", computedOwnerContactID)
	_ = d.Set("owner_contact", computedOwnerContact)
	_ = d.Set("administrative_contact", computedAdministrativeContact)
	_ = d.Set("technical_contact", computedTechnicalContact)
	_ = d.Set("auto_renew", computedAutoRenew)
	_ = d.Set("dnssec", computedDnssec)
	_ = d.Set("ds_record", computedDSRecord)
	parts := strings.Split(id, "/")

	if len(parts) != 2 {
		return diag.FromErr(fmt.Errorf("invalid ID format, expected 'projectID/domainName', got: %s", id))
	}

	_ = d.Set("task_id", parts[1])

	return nil
}

func resourceRegistrationUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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

	if d.HasChange("auto_renew") {
		newAutoRenew := d.Get("auto_renew").(bool)

		for _, domainName := range domainNames {
			domainResp, err := registrarAPI.GetDomain(&domain.RegistrarAPIGetDomainRequest{
				Domain: domainName,
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}

			if newAutoRenew {
				if domainResp.AutoRenewStatus != domain.DomainFeatureStatusEnabled &&
					domainResp.AutoRenewStatus != domain.DomainFeatureStatusEnabling {
					_, err = registrarAPI.EnableDomainAutoRenew(&domain.RegistrarAPIEnableDomainAutoRenewRequest{
						Domain: domainName,
					}, scw.WithContext(ctx))
					if err != nil {
						return diag.FromErr(err)
					}
				}
			} else {
				if domainResp.AutoRenewStatus == domain.DomainFeatureStatusEnabled ||
					domainResp.AutoRenewStatus == domain.DomainFeatureStatusEnabling {
					_, err = registrarAPI.DisableDomainAutoRenew(&domain.RegistrarAPIDisableDomainAutoRenewRequest{
						Domain: domainName,
					}, scw.WithContext(ctx))
					if err != nil {
						return diag.FromErr(err)
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
				return diag.FromErr(err)
			}

			if newDnssec {
				_, err = registrarAPI.EnableDomainDNSSEC(&domain.RegistrarAPIEnableDomainDNSSECRequest{
					Domain: domainName,
				}, scw.WithContext(ctx))
				if err != nil {
					return diag.FromErr(err)
				}
			} else if domainResp.Dnssec != nil &&
				domainResp.Dnssec.Status == domain.DomainFeatureStatusEnabled {
				_, err = registrarAPI.DisableDomainDNSSEC(&domain.RegistrarAPIDisableDomainDNSSECRequest{
					Domain: domainName,
				}, scw.WithContext(ctx))
				if err != nil {
					return diag.FromErr(err)
				}
			}

			_, err = waitForDNSSECStatus(ctx, registrarAPI, domainName, d.Timeout(schema.TimeoutUpdate))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return resourceRegistrationsRead(ctx, d, m)
}

func resourceRegistrationDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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

			return diag.FromErr(err)
		}

		if domainResp.AutoRenewStatus == domain.DomainFeatureStatusEnabled ||
			domainResp.AutoRenewStatus == domain.DomainFeatureStatusEnabling {
			_, err = registrarAPI.DisableDomainAutoRenew(&domain.RegistrarAPIDisableDomainAutoRenewRequest{
				Domain: domainName,
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	d.SetId("")

	return nil
}
