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

	domainNames, err := ExtractDomainsFromTaskID(ctx, id, registrarAPI)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, domainName := range domainNames {
		_, err := registrarAPI.GetDomain(&domain.RegistrarAPIGetDomainRequest{
			Domain: domainName,
		}, scw.WithContext(ctx))
		if err != nil {
			if httperrors.Is404(err) {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
	}

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
