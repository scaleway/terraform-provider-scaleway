package scaleway

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func newMNQNatsAPI(d *schema.ResourceData, m interface{}) (*mnq.NatsAPI, scw.Region, error) {
	meta := m.(*Meta)
	api := mnq.NewNatsAPI(meta.scwClient)
	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

func mnqNatsAPIWithRegionAndID(m interface{}, regionalID string) (*mnq.NatsAPI, scw.Region, string, error) {
	meta := m.(*Meta)
	api := mnq.NewNatsAPI(meta.scwClient)

	region, ID, err := parseRegionalID(regionalID)
	if err != nil {
		return nil, "", "", err
	}

	return api, region, ID, nil
}

func newMNQSQSAPI(d *schema.ResourceData, m any) (*mnq.SqsAPI, scw.Region, error) {
	meta := m.(*Meta)
	api := mnq.NewSqsAPI(meta.scwClient)

	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

func mnqSQSAPIWithRegionAndID(m interface{}, regionalID string) (*mnq.SqsAPI, scw.Region, string, error) {
	meta := m.(*Meta)
	api := mnq.NewSqsAPI(meta.scwClient)

	region, ID, err := parseRegionalID(regionalID)
	if err != nil {
		return nil, "", "", err
	}

	return api, region, ID, nil
}

func newMNQSNSAPI(d *schema.ResourceData, m any) (*mnq.SnsAPI, scw.Region, error) {
	meta := m.(*Meta)
	api := mnq.NewSnsAPI(meta.scwClient)

	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

func mnqSNSAPIWithRegionAndID(m interface{}, regionalID string) (*mnq.SnsAPI, scw.Region, string, error) {
	meta := m.(*Meta)
	api := mnq.NewSnsAPI(meta.scwClient)

	region, ID, err := parseRegionalID(regionalID)
	if err != nil {
		return nil, "", "", err
	}

	return api, region, ID, nil
}

func composeMNQID(region scw.Region, projectID string, queueName string) string {
	return fmt.Sprintf("%s/%s/%s", region, projectID, queueName)
}

func decomposeMNQID(id string) (region scw.Region, projectID string, name string, err error) {
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

func composeARN(region string, subject string, projectID string, resourceName string) string {
	return fmt.Sprintf("arn:scw:%s:%s:project-%s:%s", region, subject, projectID, resourceName)
}

func composeSNSARN(region string, projectID string, resourceName string) string {
	return composeARN(region, "sns", projectID, resourceName)
}

// Set the value inside values at the resource path (e.g. a.0.b sets b's value)
func setResourceValue(values map[string]interface{}, resourcePath string, value interface{}, resourceSchemas map[string]*schema.Schema) {
	parts := strings.Split(resourcePath, ".")
	if len(parts) > 1 {
		// Terraform's nested objects are represented as slices of maps
		if _, ok := values[parts[0]]; !ok {
			values[parts[0]] = []interface{}{make(map[string]interface{})}
		}

		setResourceValue(values[parts[0]].([]interface{})[0].(map[string]interface{}), strings.Join(parts[2:], "."), value, resourceSchemas[parts[0]].Elem.(*schema.Resource).Schema)
		return
	}

	values[resourcePath] = value
}

// Get the schema for the resource path (e.g. a.0.b gives b's schema)
func resolveSchemaPath(resourcePath string, resourceSchemas map[string]*schema.Schema) *schema.Schema {
	if resourceSchema, ok := resourceSchemas[resourcePath]; ok {
		return resourceSchema
	}

	parts := strings.Split(resourcePath, ".")
	if len(parts) > 1 {
		return resolveSchemaPath(strings.Join(parts[2:], "."), resourceSchemas[parts[0]].Elem.(*schema.Resource).Schema)
	}

	return nil
}

// Sets a specific SNS attribute from the resource data
func awsResourceDataToAttribute(awsAttributes map[string]*string, awsAttribute string, resourceValue interface{}, resourcePath string, resourceSchemas map[string]*schema.Schema) error {
	resourceSchema := resolveSchemaPath(resourcePath, resourceSchemas)
	if resourceSchema == nil {
		return fmt.Errorf("unable to resolve schema for %s", resourcePath)
	}

	// Only set writable attributes
	if !resourceSchema.Optional && !resourceSchema.Required {
		return nil
	}

	var s string
	switch resourceSchema.Type {
	case schema.TypeBool:
		s = strconv.FormatBool(resourceValue.(bool))
	case schema.TypeInt:
		s = strconv.Itoa(resourceValue.(int))
	case schema.TypeString:
		s = resourceValue.(string)
	default:
		return fmt.Errorf("unsupported type %s for %s", resourceSchema.Type, resourcePath)
	}

	awsAttributes[awsAttribute] = &s
	return nil
}

// awsResourceDataToAttributes returns a map of attributes from a terraform schema and a conversion map
func awsResourceDataToAttributes(d *schema.ResourceData, resourceSchemas map[string]*schema.Schema, attributesToResourceMap map[string]string) (map[string]*string, error) {
	attributes := make(map[string]*string)

	for attribute, resourcePath := range attributesToResourceMap {
		if v, ok := d.GetOk(resourcePath); ok {
			err := awsResourceDataToAttribute(attributes, attribute, v, resourcePath, resourceSchemas)
			if err != nil {
				return nil, err
			}
		}
	}

	return attributes, nil
}

// awsAttributeToResourceData sets a specific resource data from the given attribute
func awsAttributeToResourceData(values map[string]interface{}, value string, resourcePath string, resourceSchemas map[string]*schema.Schema) error {
	resourceSchema := resolveSchemaPath(resourcePath, resourceSchemas)
	if resourceSchema == nil {
		return fmt.Errorf("unable to resolve schema for %s", resourcePath)
	}

	switch resourceSchema.Type {
	case schema.TypeBool:
		b, _ := strconv.ParseBool(value)
		setResourceValue(values, resourcePath, b, resourceSchemas)
	case schema.TypeInt:
		i, _ := strconv.Atoi(value)
		setResourceValue(values, resourcePath, i, resourceSchemas)
	case schema.TypeString:
		setResourceValue(values, resourcePath, value, resourceSchemas)
	default:
		return fmt.Errorf("unsupported type %s for %s", resourceSchema.Type, resourcePath)
	}

	return nil
}

// awsAttributesToResourceData returns a map of valid values for a terraform schema from an attributes map and a conversion map
func awsAttributesToResourceData(attributes map[string]*string, resourceSchemas map[string]*schema.Schema, attributesToResourceMap map[string]string) (map[string]interface{}, error) {
	values := make(map[string]interface{})

	for attribute, resourcePath := range attributesToResourceMap {
		if value, ok := attributes[attribute]; ok && value != nil {
			err := awsAttributeToResourceData(values, *value, resourcePath, resourceSchemas)
			if err != nil {
				return nil, err
			}
		}
	}

	return values, nil
}
