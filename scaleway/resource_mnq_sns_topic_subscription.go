package scaleway

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayMNQSNSTopicSubscription() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayMNQSNSTopicSubscriptionCreate,
		ReadContext:   resourceScalewayMNQSNSTopicSubscriptionRead,
		DeleteContext: resourceScalewayMNQSNSTopicSubscriptionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"protocol": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Protocol of the SNS Topic Subscription.", // TODO: add argument list
			},
			"endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Endpoint of the subscription",
			},
			"sns_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "https://sqs-sns.mnq.{region}.scw.cloud",
				Description: "SNS endpoint",
			},
			"topic_arn": {
				Type: schema.TypeString,
			},
			"topic_id": {
				Type: schema.TypeString,
			},
			"access_key": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "SNS access key",
			},
			"secret_key": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "SNS secret key",
			},
			"redrive_policy": {
				Type:        schema.TypeBool,
				Computed:    true,
				Optional:    true,
				Description: "JSON Redrive policy",
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region":     regionSchema(),
			"project_id": projectIDSchema(),
		},
		CustomizeDiff: resourceMNQSSNSTopicCustomizeDiff,
	}
}

func resourceScalewayMNQSNSTopicSubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := newMNQSNSAPI(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID, _, err := extractProjectID(d, meta.(*Meta))
	if err != nil {
		return diag.FromErr(err)
	}

	snsInfo, err := api.GetSnsInfo(&mnq.SnsAPIGetSnsInfoRequest{
		Region:    region,
		ProjectID: projectID,
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("expected sns to be enabled for given project, go %q", snsInfo.Status))
	}

	snsClient, _, err := SNSClientWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	attributes, err := awsResourceDataToAttributes(d, resourceScalewayMNQSNSTopic().Schema, SNSTopicSubscriptionAttributesToResourceMap)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get attributes from schema: %w", err))
	}

	input := &sns.SubscribeInput{
		Attributes:            attributes,
		Endpoint:              expandStringPtr(d.Get("endpoint")),
		Protocol:              expandStringPtr(d.Get("protocol")),
		ReturnSubscriptionArn: scw.BoolPtr(true),
		TopicArn:              nil,
	}

	output, err := snsClient.SubscribeWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create SNS Topic: %w", err))
	}

	if output.SubscriptionArn == nil {
		return diag.Errorf("subscription id is nil on creation")
	}

	d.SetId(newRegionalIDString(region, *output.SubscriptionArn))

	return resourceScalewayMNQSNSTopicSubscriptionRead(ctx, d, meta)
}

func resourceScalewayMNQSNSTopicSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	snsClient, region, id, err := SNSClientWithRegionAndID(d, meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	subAttributes, err := snsClient.GetSubscriptionAttributesWithContext(ctx, &sns.GetSubscriptionAttributesInput{
		SubscriptionArn: scw.StringPtr(id),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	schemaAttributes, err := awsAttributesToResourceData(subAttributes.Attributes, resourceScalewayMNQSNSTopic().Schema, SNSTopicSubscriptionAttributesToResourceMap)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("region", region)
	_ = d.Set("arn", id)

	for k, v := range schemaAttributes {
		_ = d.Set(k, v)
	}

	return nil
}

func resourceScalewayMNQSNSTopicSubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	snsClient, _, id, err := SNSClientWithRegionAndID(d, meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = snsClient.DeleteTopicWithContext(ctx, &sns.DeleteTopicInput{
		TopicArn: scw.StringPtr(id),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
