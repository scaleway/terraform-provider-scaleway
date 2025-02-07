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
