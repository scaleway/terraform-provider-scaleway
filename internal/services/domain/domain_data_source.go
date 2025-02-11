package domain

//
//import (
//	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
//	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
//)
//
//func ResourceDomainsRegistration() *schema.Resource {
//	return &schema.Resource{
//		CreateContext: resourceDomainsRegistrationCreate,
//		ReadContext:   resourceDomainsRegistrationsRead,
//		UpdateContext: resourceDomainsRegistrationUpdate,
//		DeleteContext: resourceDomainsRegistrationDelete,
//		Timeouts: &schema.ResourceTimeout{
//			Create:  schema.DefaultTimeout(defaultDomainRecordTimeout),
//			Read:    schema.DefaultTimeout(defaultDomainRecordTimeout),
//			Update:  schema.DefaultTimeout(defaultDomainRecordTimeout),
//			Delete:  schema.DefaultTimeout(defaultDomainRecordTimeout),
//			Default: schema.DefaultTimeout(defaultDomainRecordTimeout),
//		},
//		Importer: &schema.ResourceImporter{
//			StateContext: schema.ImportStatePassthroughContext,
//		},
//		SchemaVersion: 0,
//		Schema: map[string]*schema.Schema{
//			"domain_names": {
//				Type:        schema.TypeList,
//				Required:    true,
//				Elem:        &schema.Schema{Type: schema.TypeString},
//				Description: "Liste des noms de domaines à gérer.",
//			},
//			"duration_in_years": {
//				Type:     schema.TypeInt,
//				Optional: true,
//				Default:  1,
//			},
//			"project_id": account.ProjectIDSchema(),
//			"owner_contact_id": {
//				Type:     schema.TypeString,
//				Optional: true,
//				Computed: true,
//				ExactlyOneOf: []string{
//					"owner_contact_id",
//					"owner_contact",
//				},
//				Description: "ID du contact propriétaire. Soit `owner_contact_id`, soit `owner_contact` doit être fourni.",
//			},
//			"owner_contact": {
//				Type:     schema.TypeList,
//				Optional: true,
//				Computed: true,
//				MaxItems: 1,
//				ExactlyOneOf: []string{
//					"owner_contact_id",
//					"owner_contact",
//				},
//				Elem: &schema.Resource{
//					Schema: contactSchema(),
//				},
//				Description: "Détails du contact propriétaire. Soit `owner_contact_id`, soit `owner_contact` doit être fourni.",
//			},
//			"administrative_contact_id": {
//				Type:     schema.TypeString,
//				Optional: true,
//			},
//			"administrative_contact": {
//				Type:     schema.TypeList,
//				Optional: true,
//				Computed: true,
//				MaxItems: 1,
//				Elem: &schema.Resource{
//					Schema: contactSchema(),
//				},
//				Description: "Détails du contact administratif.",
//			},
//			"technical_contact_id": {
//				Type:     schema.TypeString,
//				Optional: true,
//			},
//			"technical_contact": {
//				Type:     schema.TypeList,
//				Optional: true,
//				Computed: true,
//				MaxItems: 1,
//				Elem: &schema.Resource{
//					Schema: contactSchema(),
//				},
//				Description: "Détails du contact technique.",
//			},
//			"auto_renew": {
//				Type:        schema.TypeBool,
//				Optional:    true,
//				Default:     false,
//				Description: "Active ou désactive le renouvellement automatique du domaine.",
//			},
//			"dnssec": {
//				Type:        schema.TypeBool,
//				Optional:    true,
//				Default:     false,
//				Description: "Active ou désactive DNSSEC pour le domaine.",
//			},
//			"ds_record": {
//				Type:     schema.TypeList,
//				Optional: true,
//				Computed: true,
//				MaxItems: 1,
//				Elem: &schema.Resource{
//					Schema: map[string]*schema.Schema{
//						"key_id": {
//							Type:        schema.TypeInt,
//							Required:    true,
//							Description: "L’identifiant de la clé DNSSEC.",
//						},
//						"algorithm": {
//							Type:        schema.TypeString,
//							Required:    true,
//							Description: "L’algorithme utilisé pour DNSSEC (ex. rsasha256, ecdsap256sha256).",
//						},
//						"digest": {
//							Type:     schema.TypeList,
//							Optional: true,
//							MaxItems: 1,
//							Elem: &schema.Resource{
//								Schema: map[string]*schema.Schema{
//									"type": {
//										Type:        schema.TypeString,
//										Required:    true,
//										Description: "Le type de digest (ex. sha_1, sha_256).",
//									},
//									"digest": {
//										Type:        schema.TypeString,
//										Required:    true,
//										Description: "La valeur du digest.",
//									},
//									"public_key": {
//										Type:     schema.TypeList,
//										Optional: true,
//										MaxItems: 1,
//										Elem: &schema.Resource{
//											Schema: map[string]*schema.Schema{
//												"key": {
//													Type:        schema.TypeString,
//													Required:    true,
//													Description: "La valeur de la clé publique.",
//												},
//											},
//										},
//										Description: "La clé publique associée au digest.",
//									},
//								},
//							},
//							Description: "Détails sur le digest.",
//						},
//						"public_key": {
//							Type:     schema.TypeList,
//							Optional: true,
//							MaxItems: 1,
//							Elem: &schema.Resource{
//								Schema: map[string]*schema.Schema{
//									"key": {
//										Type:        schema.TypeString,
//										Required:    true,
//										Description: "La valeur de la clé publique.",
//									},
//								},
//							},
//							Description: "Clé publique associée au DS record DNSSEC.",
//						},
//					},
//				},
//				Description: "Configuration du DS record DNSSEC.",
//			},
//			"is_external": {
//				Type:        schema.TypeBool,
//				Optional:    true,
//				Computed:    true,
//				Description: "Indique si Scaleway est le registrar du domaine.",
//			},
//		},
//	}
//}
//

//
//func dataSourceDomainRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
//	registrarAPI := NewRegistrarDomainAPI(m)
//	domainName := d.Get("domain").(string)
//
//	resp, err := registrarAPI.GetDomain(&domain.RegistrarAPIGetDomainRequest{
//		Domain: domainName,
//	}, scw.WithContext(ctx))
//	if err != nil {
//		return diag.FromErr(err)
//	}
//
//	if err := d.Set("auto_renew_status", resp.AutoRenewStatus.String()); err != nil {
//		return diag.FromErr(err)
//	}
//	if resp.Dnssec != nil {
//		if err := d.Set("dnssec_status", resp.Dnssec.Status.String()); err != nil {
//			return diag.FromErr(err)
//		}
//	}
//	if err := d.Set("epp_code", resp.EppCode); err != nil {
//		return diag.FromErr(err)
//	}
//	if resp.ExpiredAt != nil {
//		if err := d.Set("expired_at", resp.ExpiredAt.Format(time.RFC3339)); err != nil {
//			return diag.FromErr(err)
//		}
//	}
//	if resp.UpdatedAt != nil {
//		if err := d.Set("updated_at", resp.UpdatedAt.Format(time.RFC3339)); err != nil {
//			return diag.FromErr(err)
//		}
//	}
//	if err := d.Set("registrar", resp.Registrar); err != nil {
//		return diag.FromErr(err)
//	}
//	if err := d.Set("status", string(resp.Status)); err != nil {
//		return diag.FromErr(err)
//	}
//	if err := d.Set("organization_id", resp.OrganizationID); err != nil {
//		return diag.FromErr(err)
//	}
//	if err := d.Set("pending_trade", resp.PendingTrade); err != nil {
//		return diag.FromErr(err)
//	}
//	if resp.ExternalDomainRegistrationStatus != nil {
//		if err := d.Set("external_domain_registration_status", flattenExternalDomainRegistrationStatus(resp.ExternalDomainRegistrationStatus)); err != nil {
//			return diag.FromErr(err)
//		}
//	}
//	if resp.TransferRegistrationStatus != nil {
//		if err := d.Set("transfer_registration_status", flattenDomainRegistrationStatusTransfer(resp.TransferRegistrationStatus)); err != nil {
//			return diag.FromErr(err)
//		}
//	}
//	var linkedProductsStr []string
//	for _, lp := range resp.LinkedProducts {
//		linkedProductsStr = append(linkedProductsStr, lp.String())
//	}
//	if err := d.Set("linked_products", linkedProductsStr); err != nil {
//		return diag.FromErr(err)
//	}
//	if resp.Tld != nil {
//		if err := d.Set("tld", flattenTLD(resp.Tld)); err != nil {
//			return diag.FromErr(err)
//		}
//	}
//	if len(resp.DNSZones) > 0 {
//		if err := d.Set("dns_zones", flattenDNSZones(resp.DNSZones)); err != nil {
//			return diag.FromErr(err)
//		}
//	}
//
//	d.SetId(domainName)
//	return nil
//}
