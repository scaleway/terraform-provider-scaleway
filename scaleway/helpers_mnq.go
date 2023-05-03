package scaleway

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nats-io/nats.go"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

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

func NATSClientWithRegion(d *schema.ResourceData, m interface{}) (nats.JetStreamContext, scw.Region, error) { //nolint:ireturn
	meta := m.(*Meta)
	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	if _, ok := d.GetOk("nats"); !ok {
		return nil, "", fmt.Errorf("nats credentials are required")
	}

	credentials := d.Get("nats.0.credentials").(string)
	js, err := newNATSJetStreamClient(region.String(), credentials)
	if err != nil {
		return nil, "", err
	}

	return js, region, err
}

func newNATSJetStreamClient(region, credentials string) (nats.JetStreamContext, error) { //nolint:ireturn
	jwt, seed, err := splitNATSJWTAndSeed(credentials)
	if err != nil {
		return nil, err
	}

	nc, err := nats.Connect("nats://nats.mnq."+region+".scw.cloud:4222", nats.UserJWTAndSeed(jwt, seed))
	if err != nil {
		return nil, err
	}

	js, err := nc.JetStream()
	if err != nil {
		return nil, err
	}

	return js, nil
}

func splitNATSJWTAndSeed(credentials string) (string, string, error) {
	lines := strings.Split(credentials, "\n")
	if len(lines) < 6 {
		return "", "", fmt.Errorf("invalid credentials format")
	}

	jwt := lines[1]
	seed := lines[4]
	return jwt, seed, nil
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
