package mnq

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceSNSTopicSubscription() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceMNQSNSTopicSubscriptionCreate,
		ReadContext:   ResourceMNQSNSTopicSubscriptionRead,
		DeleteContext: ResourceMNQSNSTopicSubscriptionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		SchemaFunc:    snsTopicSubscriptionSchema,
	}
}

func snsTopicSubscriptionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"protocol": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Protocol of the SNS Topic Subscription.", // TODO: add argument list
			ForceNew:    true,
		},
		"endpoint": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Endpoint of the subscription",
			ForceNew:    true,
		},
		"sns_endpoint": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "https://sns.mnq.{region}.scaleway.com",
			Description: "SNS endpoint",
			ForceNew:    true,
		},
		"topic_arn": {
			Type:         schema.TypeString,
			Description:  "ARN of the topic",
			Optional:     true,
			AtLeastOneOf: []string{"topic_id"},
			ForceNew:     true,
		},
		"topic_id": {
			Type:         schema.TypeString,
			Description:  "ID of the topic",
			Optional:     true,
			AtLeastOneOf: []string{"topic_arn"},
			ForceNew:     true,
		},
		"access_key": {
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
			Description: "SNS access key",
			ForceNew:    true,
		},
		"secret_key": {
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
			Description: "SNS secret key",
			ForceNew:    true,
		},
		"redrive_policy": {
			Type:        schema.TypeBool,
			Computed:    true,
			Optional:    true,
			Description: "JSON Redrive policy",
			ForceNew:    true,
		},
		"arn": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "ARN of the topic, should have format 'arn:scw:sns:project-${project_id}:${topic_name}:${subscription_id}'",
		},
		"region":     regional.Schema(),
		"project_id": account.ProjectIDSchema(),
	}
}

func ResourceMNQSNSTopicSubscriptionCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := newMNQSNSAPI(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID, _, err := meta.ExtractProjectID(d, m)
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

	snsClient, _, err := SNSClientWithRegion(ctx, m, d)
	if err != nil {
		return diag.FromErr(err)
	}

	attributes, err := awsResourceDataToAttributes(d, ResourceSNSTopic().SchemaFunc(), SNSTopicSubscriptionAttributesToResourceMap)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get attributes from schema: %w", err))
	}

	// Get topic ARN from either topic_arn or topic_id
	topicARN := ""
	if topicARNRaw, ok := d.GetOk("topic_arn"); ok {
		topicARN = topicARNRaw.(string)
	} else {
		topicRegion, topicProject, topicName, err := DecomposeMNQID(d.Get("topic_id").(string))
		if err != nil {
			return diag.Diagnostics{{
				Severity:      diag.Error,
				Summary:       "Failed to parse topic id",
				Detail:        err.Error(),
				AttributePath: cty.GetAttrPath("topic_id"),
			}}
		}

		topicARN = ComposeSNSARN(topicRegion, topicProject, topicName)
	}

	input := &sns.SubscribeInput{
		Attributes:            attributes,
		Endpoint:              types.ExpandStringPtr(d.Get("endpoint")),
		Protocol:              types.ExpandStringPtr(d.Get("protocol")),
		ReturnSubscriptionArn: true,
		TopicArn:              &topicARN,
	}

	output, err := snsClient.Subscribe(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create SNS Topic: %w", err))
	}

	if output.SubscriptionArn == nil {
		return diag.Errorf("subscription id is nil on creation")
	}

	arn, err := decomposeARN(*output.SubscriptionArn)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to parse arn: %w", err))
	}

	d.SetId(composeMNQSubscriptionID(arn.Region, arn.ProjectID, arn.ResourceName, arn.ExtraResourceID))

	return ResourceMNQSNSTopicSubscriptionRead(ctx, d, m)
}

func ResourceMNQSNSTopicSubscriptionRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	snsClient, region, err := SNSClientWithRegionFromID(ctx, d, m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	arn, err := DecomposeMNQSubscriptionID(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to parse id: %w", err))
	}

	subAttributes, err := snsClient.GetSubscriptionAttributes(ctx, &sns.GetSubscriptionAttributesInput{
		SubscriptionArn: new(arn.String()),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	schemaAttributes, err := awsAttributesToResourceData(subAttributes.Attributes, ResourceSNSTopic().SchemaFunc(), SNSTopicSubscriptionAttributesToResourceMap)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("region", region)
	_ = d.Set("arn", arn.String())

	for k, v := range schemaAttributes {
		_ = d.Set(k, v) // lintignore: R001
	}

	return nil
}

func ResourceMNQSNSTopicSubscriptionDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	snsClient, _, err := SNSClientWithRegionFromID(ctx, d, m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	arn, err := DecomposeMNQSubscriptionID(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to parse id: %w", err))
	}

	_, err = snsClient.Unsubscribe(ctx, &sns.UnsubscribeInput{
		SubscriptionArn: new(arn.String()),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
