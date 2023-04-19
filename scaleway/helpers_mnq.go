package scaleway

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceMNQQueueCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	fifoQueue := d.Get("fifo_queue").(bool)

	if d.Id() == "" {
		var name string
		if v, ok := d.GetOk("name"); ok {
			name = v.(string)
		} else if v, ok := d.GetOk("name_prefix"); ok {
			name = id.PrefixedUniqueId(v.(string))

			if _, ok := d.GetOk("sqs"); ok && fifoQueue {
				name += SQSFIFOQueueNameSuffix
			}
		}

		var re *regexp.Regexp

		if fifoQueue {
			re = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,75}\.fifo$`)
		} else {
			re = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,80}$`)
		}

		if !re.MatchString(name) {
			return fmt.Errorf("invalid queue name: %s (format is %s)", name, re.String())
		}
	}

	if _, ok := d.GetOk("sqs"); ok {
		contentBasedDeduplication := d.Get("sqs.0.content_based_deduplication").(bool)

		if !fifoQueue && contentBasedDeduplication {
			return fmt.Errorf("content-based deduplication can only be set for FIFO queue")
		}
	}

	return nil
}

func composeMNQID(region scw.Region, namespaceID string, queueName string) string {
	return fmt.Sprintf("%s/%s/%s", region, namespaceID, queueName)
}

func decomposeMNQID(id string) (region scw.Region, namespaceID string, name string, err error) {
	parts := strings.Split(id, "/")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid ID format: %q", id)
	}

	region, err = scw.ParseRegion(parts[0])
	if err != nil {
		return "", "", "", err
	}

	return region, parts[1], parts[2], nil
}

func getMNQNamespaceFromComposedID(ctx context.Context, d *schema.ResourceData, meta interface{}, composedID string) (*mnq.Namespace, error) {
	api, region, err := newMNQAPI(d, meta)
	if err != nil {
		return nil, err
	}

	namespaceRegion, namespaceID, _, err := decomposeMNQID(composedID)
	if err != nil {
		return nil, err
	}
	if namespaceRegion != region {
		return nil, fmt.Errorf("namespace region (%s) and queue region (%s) must be the same", namespaceRegion, region)
	}

	return api.GetNamespace(&mnq.GetNamespaceRequest{
		Region:      namespaceRegion,
		NamespaceID: namespaceID,
	}, scw.WithContext(ctx))
}

func newMNQAPI(d *schema.ResourceData, m interface{}) (*mnq.API, scw.Region, error) {
	meta := m.(*Meta)
	api := mnq.NewAPI(meta.scwClient)
	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

func SQSClientWithRegion(d *schema.ResourceData, m interface{}) (*sqs.SQS, scw.Region, error) {
	meta := m.(*Meta)
	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	if _, ok := d.GetOk("sqs"); !ok {
		return nil, "", fmt.Errorf("sqs access_key and secret_key are required")
	}

	accessKey := d.Get("sqs.0.access_key").(string)
	secretKey := d.Get("sqs.0.secret_key").(string)

	sqsClient, err := newSQSClient(meta.httpClient, region.String(), accessKey, secretKey)
	if err != nil {
		return nil, "", err
	}

	return sqsClient, region, err
}

func newSQSClient(httpClient *http.Client, region, accessKey, secretKey string) (*sqs.SQS, error) {
	config := &aws.Config{}
	config.WithRegion(region)
	config.WithCredentials(credentials.NewStaticCredentials(accessKey, secretKey, ""))
	config.WithEndpoint("http://sqs-sns.mnq." + region + ".scw.cloud")
	config.WithHTTPClient(httpClient)
	if strings.ToLower(os.Getenv("TF_LOG")) == "debug" {
		config.WithLogLevel(aws.LogDebugWithHTTPBody)
	}

	s, err := session.NewSession(config)
	if err != nil {
		return nil, err
	}
	return sqs.New(s), nil
}

func mnqAPIWithRegionAndID(m interface{}, regionalID string) (*mnq.API, scw.Region, string, error) {
	meta := m.(*Meta)
	mnqAPI := mnq.NewAPI(meta.scwClient)

	region, ID, err := parseRegionalID(regionalID)
	if err != nil {
		return nil, "", "", err
	}
	return mnqAPI, region, ID, nil
}
