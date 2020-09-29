package scaleway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	scwLogger "github.com/scaleway/scaleway-sdk-go/logger"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// Meta contains config and SDK clients used by resources.
//
// This meta value is passed into all resources.
type Meta struct {
	AccessKey             string
	SecretKey             string
	DefaultOrganizationID string
	DefaultRegion         scw.Region
	DefaultZone           scw.Zone
	APIURL                string
	TerraformVersion      string

	// scwClient is the Scaleway SDK client.
	scwClient *scw.Client

	// s3Client is the S3 client
	s3Client *s3.S3
}

// bootstrap initializes all the clients for this meta config object.
func (m *Meta) bootstrap() error {
	err := m.bootstrapScwClient()
	if err != nil {
		return err
	}

	err = m.bootstrapS3Client()
	if err != nil {
		return err
	}

	return nil
}

// bootstrapScwClient initializes a new scw.Client from the configuration.
func (m *Meta) bootstrapScwClient() error {
	// Init SDK logger
	scwLogger.SetLogger(l)

	options := []scw.ClientOption{
		scw.WithHTTPClient(createRetryableHTTPClient(false)),
		scw.WithUserAgent(fmt.Sprintf("terraform-provider/%s terraform/%s", version, m.TerraformVersion)),
	}

	// The access key is not used for API authentications.
	if m.SecretKey != "" {
		options = append(options, scw.WithAuth(m.AccessKey, m.SecretKey))
	}

	if m.DefaultOrganizationID != "" {
		options = append(options, scw.WithDefaultOrganizationID(m.DefaultOrganizationID))
	}

	if m.DefaultRegion != "" {
		options = append(options, scw.WithDefaultRegion(m.DefaultRegion))
	}

	if m.DefaultZone != "" {
		options = append(options, scw.WithDefaultZone(m.DefaultZone))
	}

	if m.APIURL != "" {
		options = append(options, scw.WithAPIURL(m.APIURL))
	}

	client, err := scw.NewClient(options...)
	if err != nil {
		return fmt.Errorf("cannot create SDK client: %s", err)
	}

	m.scwClient = client
	return nil
}

// createRetryableHTTPClient creates a retryablehttp.Client.
func createRetryableHTTPClient(shouldLog bool) *client {
	c := retryablehttp.NewClient()

	if shouldLog {
		c.HTTPClient.Transport = logging.NewTransport("Scaleway", c.HTTPClient.Transport)
	}
	c.RetryMax = 3
	c.RetryWaitMax = 2 * time.Minute
	c.Logger = l
	c.RetryWaitMin = time.Second * 2
	c.CheckRetry = func(_ context.Context, resp *http.Response, err error) (bool, error) {
		if resp == nil || resp.StatusCode == http.StatusTooManyRequests {
			return true, err
		}
		return retryablehttp.DefaultRetryPolicy(context.TODO(), resp, err)
	}

	return &client{c}
}

// client is a bridge between scw.httpClient interface and retryablehttp.Client
type client struct {
	*retryablehttp.Client
}

// Do wraps calling an HTTP method with retries.
func (c *client) Do(r *http.Request) (*http.Response, error) {
	var body io.ReadSeeker
	if r.Body != nil {
		bs, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(bs)
	}
	req, err := retryablehttp.NewRequest(r.Method, r.URL.String(), body)
	if err != nil {
		return nil, err
	}
	for key, val := range r.Header {
		req.Header.Set(key, val[0])
	}
	return c.Client.Do(req)
}

// bootstrapS3Client initializes a new s3 client from the configuration.
func (m *Meta) bootstrapS3Client() error {
	var err error

	config := &aws.Config{}
	config.WithRegion(string(m.DefaultRegion))
	config.WithCredentials(credentials.NewStaticCredentials(m.AccessKey, m.SecretKey, ""))
	config.WithEndpoint(m.s3Endpoint(m.DefaultRegion))

	s, err := session.NewSession(config)
	if err != nil {
		return err
	}

	m.s3Client = s3.New(s)
	return nil
}

// s3Endpoint returns the correct S3 endpoint for object storage based on the current region
func (m *Meta) s3Endpoint(region scw.Region) string {
	return "https://s3." + string(region) + ".scw.cloud"
}

// deprecatedScalewayConfig is the structure of the deprecated Scaleway config file.
type deprecatedScalewayConfig struct {
	Organization string `json:"organization"`
	Token        string `json:"token"`
	Version      string `json:"version"`
}

// readDeprecatedScalewayConfig parses the deprecated Scaleway config file.
func readDeprecatedScalewayConfig(path string) (string, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	var data deprecatedScalewayConfig
	if err := json.NewDecoder(f).Decode(&data); err != nil {
		return "", "", err
	}
	return data.Token, data.Organization, nil
}
