package scaleway

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayMNQCredential() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayMNQCredentialCreate,
		ReadContext:   resourceScalewayMNQCredentialRead,
		UpdateContext: resourceScalewayMNQCredentialUpdate,
		DeleteContext: resourceScalewayMNQCredentialDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The name of the Credential",
			},
			"namespace_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the Namespace associated to",
			},
			// computed
			"region": regionSchema(),
			"protocol": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Namespace protocol",
			},
			"nats_credentials": {
				Type:          schema.TypeList,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"sqs_sns_credentials"},
				Description:   "credential for NATS protocol",
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"content": {
							Type:        schema.TypeString,
							Computed:    true,
							Sensitive:   true,
							Description: "Raw content of the NATS credentials file",
						},
					},
				},
			},
			"sqs_sns_credentials": {
				Type:          schema.TypeList,
				Optional:      true,
				Description:   "The credential used to connect to the SQS/SNS service",
				MaxItems:      1,
				ConflictsWith: []string{"nats_credentials"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"permissions": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "The permission associated to this credential.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"can_publish": {
										Type:        schema.TypeBool,
										Default:     false,
										Optional:    true,
										Description: "Allow publish messages to the service",
									},
									"can_receive": {
										Type:        schema.TypeBool,
										Default:     false,
										Optional:    true,
										Description: "Allow receive messages from the service",
									},
									"can_manage": {
										Type:        schema.TypeBool,
										Default:     false,
										Optional:    true,
										Description: "Allow manage the associated resource",
									},
								},
							},
						},
						"secret_key": {
							Type:        schema.TypeString,
							Computed:    true,
							Sensitive:   true,
							Description: "The secret value of the key",
						},
						"access_key": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The key of the credential",
						},
					},
				},
			},
		},
	}
}

func resourceScalewayMNQCredentialCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := newMNQAPI(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	request := &mnq.CreateCredentialRequest{
		Name:        expandOrGenerateString(d.Get("name").(string), "cred"),
		NamespaceID: expandID(d.Get("namespace_id")),
		Region:      region,
	}

	if _, ok := d.GetOk("sqs_sns_credentials"); ok {
		perm := mnq.Permissions{}
		perm.CanPublish = expandBoolPtr(d.Get("sqs_sns_credentials.0.permissions.0.can_publish"))
		perm.CanManage = expandBoolPtr(d.Get("sqs_sns_credentials.0.permissions.0.can_manage"))
		perm.CanReceive = expandBoolPtr(d.Get("sqs_sns_credentials.0.permissions.0.can_receive"))
		request.Permissions = &perm
	}
	credential, err := api.CreateCredential(request, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	_ = d.Set(
		"nats_credentials",
		setCreedsNATS(credential.NatsCredentials),
	)
	_ = d.Set(
		"sqs_sns_credentials",
		setPermissionsSQS(credential.SqsSnsCredentials),
	)

	d.SetId(newRegionalIDString(region, credential.ID))

	return resourceScalewayMNQCredentialRead(ctx, d, meta)
}

func resourceScalewayMNQCredentialRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := mnqAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	request := &mnq.GetCredentialRequest{
		CredentialID: id,
		Region:       region,
	}

	credential, err := api.GetCredential(request, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("name", credential.Name)
	_ = d.Set("protocol", credential.Protocol.String())
	_ = d.Set("region", region)

	return nil
}

func setCreedsNATS(credentials *mnq.CredentialNATSCredsFile) interface{} {
	var flattened []map[string]interface{}

	if credentials == nil {
		return flattened
	}
	return []map[string]interface{}{{"content": credentials.Content}}
}

func setPermissionsSQS(credentials *mnq.CredentialSQSSNSCreds) interface{} {
	var flattened []map[string]interface{}

	if credentials == nil {
		return flattened
	}

	flattened = []map[string]interface{}{
		{
			"access_key": credentials.AccessKey,
			"secret_key": credentials.SecretKey,
		},
	}

	if credentials.Permissions != nil {
		flattened[0]["permissions"] = []map[string]interface{}{{
			"can_publish": credentials.Permissions.CanPublish,
			"can_receive": credentials.Permissions.CanReceive,
			"can_manage":  credentials.Permissions.CanManage,
		}}
	}

	return flattened
}

func resourceScalewayMNQCredentialUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := mnqAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	request := &mnq.UpdateCredentialRequest{
		CredentialID: id,
		Region:       region,
	}

	if d.HasChange("name") {
		request.Name = scw.StringPtr(d.Get("name").(string))
	}

	if _, exist := d.GetOk("sqs_sns_credentials"); exist && d.HasChange("sqs_sns_credentials") {
		perm := mnq.Permissions{}
		perm.CanPublish = expandBoolPtr(d.Get("sqs_sns_credentials.0.permissions.0.can_publish"))
		perm.CanManage = expandBoolPtr(d.Get("sqs_sns_credentials.0.permissions.0.can_manage"))
		perm.CanReceive = expandBoolPtr(d.Get("sqs_sns_credentials.0.permissions.0.can_receive"))
		request.Permissions = &perm
	}

	_, err = api.UpdateCredential(request, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayMNQCredentialRead(ctx, d, meta)
}

func resourceScalewayMNQCredentialDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := mnqAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	request := &mnq.DeleteCredentialRequest{
		CredentialID: id,
		Region:       region,
	}
	err = api.DeleteCredential(request, scw.WithContext(ctx))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
