package mnq

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
)

func ResourceSNSTopic() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceMNQSNSTopicCreate,
		ReadContext:   ResourceMNQSNSTopicRead,
		UpdateContext: ResourceMNQSNSTopicUpdate,
		DeleteContext: ResourceMNQSNSTopicDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ForceNew:      true,
				Description:   "Name of the SNS Topic.",
				ConflictsWith: []string{"name_prefix"},
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				Description:   "Creates a unique name beginning with the specified prefix.",
				ConflictsWith: []string{"name"},
			},
			"sns_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "https://sns.mnq.{region}.scaleway.com",
				Description: "SNS endpoint",
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
			"content_based_deduplication": {
				Type:        schema.TypeBool,
				Computed:    true,
				Optional:    true,
				Description: "Specifies whether to enable content-based deduplication.",
			},
			"fifo_topic": {
				Type:        schema.TypeBool,
				Computed:    true,
				Optional:    true,
				Description: "Whether the topic is a FIFO topic. If true, the topic name must end with .fifo",
			},
			"owner": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Owner of the SNS topic, should have format 'project-${project_id}'",
			},
			"arn": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ARN of the topic, should have format 'arn:scw:sns:project-${project_id}:${topic_name}'",
			},
			"region":     regional.Schema(),
			"project_id": account.ProjectIDSchema(),
		},
		CustomizeDiff: resourceMNQSSNSTopicCustomizeDiff,
	}
}

func ResourceMNQSNSTopicCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	snsClient, _, err := SNSClientWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	attributes, err := awsResourceDataToAttributes(d, ResourceSNSTopic().Schema, SNSTopicAttributesToResourceMap)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get attributes from schema: %w", err))
	}

	isFifo := d.Get("fifo_topic").(bool)
	topicName := resourceMNQSNSTopicName(d.Get("name"), d.Get("name_prefix"), true, isFifo)

	input := &sns.CreateTopicInput{
		Name:       scw.StringPtr(topicName),
		Attributes: attributes,
	}

	output, err := snsClient.CreateTopic(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create SNS Topic: %w", err))
	}

	if output.TopicArn == nil {
		return diag.Errorf("topic id is nil on creation")
	}

	d.SetId(composeMNQID(region, projectID, topicName))

	return ResourceMNQSNSTopicRead(ctx, d, m)
}

func ResourceMNQSNSTopicRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	snsClient, _, err := SNSClientWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	region, projectID, topicName, err := DecomposeMNQID(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to parse id: %w", err))
	}

	topicAttributes, err := snsClient.GetTopicAttributes(ctx, &sns.GetTopicAttributesInput{
		TopicArn: scw.StringPtr(ComposeSNSARN(region, projectID, topicName)),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	schemaAttributes, err := awsAttributesToResourceData(topicAttributes.Attributes, ResourceSNSTopic().Schema, SNSTopicAttributesToResourceMap)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("name", topicName)
	_ = d.Set("region", region)

	for k, v := range schemaAttributes {
		_ = d.Set(k, v) // lintignore: R001
	}

	return nil
}

func ResourceMNQSNSTopicUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	snsClient, _, err := SNSClientWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	region, projectID, topicName, err := DecomposeMNQID(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to parse id: %w", err))
	}

	topicARN := ComposeSNSARN(region, projectID, topicName)

	changedAttributes := []string(nil)
	for attributeName, schemaName := range SNSTopicAttributesToResourceMap {
		if d.HasChange(schemaName) {
			changedAttributes = append(changedAttributes, attributeName)
		}
	}

	attributes, err := awsResourceDataToAttributes(d, ResourceSNSTopic().Schema, SNSTopicAttributesToResourceMap)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get attributes from schema: %w", err))
	}

	updatedAttributes := map[string]string{}

	for _, changedAttribute := range changedAttributes {
		updatedAttributes[changedAttribute] = attributes[changedAttribute]
	}

	if len(updatedAttributes) > 0 {
		for attributeName, attributeValue := range updatedAttributes {
			_, err := snsClient.SetTopicAttributes(ctx, &sns.SetTopicAttributesInput{
				AttributeName:  scw.StringPtr(attributeName),
				AttributeValue: &attributeValue,
				TopicArn:       &topicARN,
			})
			if err != nil {
				return diag.FromErr(fmt.Errorf("failed to set attribute %q: %w", attributeName, err))
			}
		}
	}

	return ResourceMNQSNSTopicRead(ctx, d, m)
}

func ResourceMNQSNSTopicDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	snsClient, _, err := SNSClientWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	region, projectID, topicName, err := DecomposeMNQID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = snsClient.DeleteTopic(ctx, &sns.DeleteTopicInput{
		TopicArn: scw.StringPtr(ComposeSNSARN(region, projectID, topicName)),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
