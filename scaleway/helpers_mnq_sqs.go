package scaleway

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func waitQueueAttributesPropagated(ctx context.Context, conn *sqs.SQS, url string, expected map[string]string) error {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{queueAttributeStateNotEqual},
		Target:                    []string{queueAttributeStateEqual},
		Refresh:                   statusQueueAttributeState(ctx, conn, url, expected),
		Timeout:                   queueAttributePropagationTimeout,
		ContinuousTargetOccurence: 6,               // set to accommodate GovCloud, commercial, China, etc. - avoid lowering
		MinTimeout:                5 * time.Second, // set to accommodate GovCloud, commercial, China, etc. - avoid lowering
		NotFoundChecks:            10,              // set to accommodate GovCloud, commercial, China, etc. - avoid lowering
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func statusQueueAttributeState(ctx context.Context, conn *sqs.SQS, url string, expected map[string]string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		attributesMatch := func(got map[string]string) string {
			for k, e := range expected {
				g, ok := got[k]

				if !ok {
					// Missing attribute equivalent to empty expected value.
					if e == "" {
						continue
					}

					// Backwards compatibility: https://github.com/hashicorp/terraform-provider-aws/issues/19786.
					if k == sqs.QueueAttributeNameKmsDataKeyReusePeriodSeconds && e == strconv.Itoa(DefaultQueueKMSDataKeyReusePeriodSeconds) {
						continue
					}

					return queueAttributeStateNotEqual
				}

				switch k {
				case sqs.QueueAttributeNamePolicy:
					equivalent, err := awspolicy.PoliciesAreEquivalent(g, e)
					if err != nil {
						return queueAttributeStateNotEqual
					}

					if !equivalent {
						return queueAttributeStateNotEqual
					}
				case sqs.QueueAttributeNameRedriveAllowPolicy, sqs.QueueAttributeNameRedrivePolicy:
					if !StringsEquivalent(g, e) {
						return queueAttributeStateNotEqual
					}
				default:
					if g != e {
						return queueAttributeStateNotEqual
					}
				}
			}

			return queueAttributeStateEqual
		}

		got, err := FindQueueAttributesByURL(ctx, conn, url)

		if NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		status := attributesMatch(got)

		return got, status, nil
	}
}

// retryWhenNotFound retries the specified function when it returns a retry.NotFoundError.
func retryWhenNotFound(ctx context.Context, timeout time.Duration, f func() (interface{}, error)) (interface{}, error) {
	return RetryWhen(ctx, timeout, f, func(err error) (bool, error) {
		if NotFound(err) {
			return true, err
		}

		return false, err
	})
}

func FindQueueAttributesByURL(ctx context.Context, conn *sqs.SQS, url string) (map[string]string, error) {
	input := &sqs.GetQueueAttributesInput{
		AttributeNames: aws.StringSlice([]string{sqs.QueueAttributeNameAll}),
		QueueUrl:       aws.String(url),
	}

	output, err := conn.GetQueueAttributesWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, sqs.ErrCodeQueueDoesNotExist) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Attributes == nil {
		return nil, NewEmptyResultError(input)
	}

	return aws.StringValueMap(output.Attributes), nil
}

// QueueNameFromURL returns the SQS queue name from the specified URL.
func QueueNameFromURL(u string) (string, error) {
	v, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	// http://sqs-sns.mnq.fr-par.scw.com/123456789012/queueName
	parts := strings.Split(v.Path, "/")

	if len(parts) != 3 {
		return "", fmt.Errorf("SQS Queue URL (%s) is in the incorrect format", u)
	}

	return parts[2], nil
}

// APIAttributesToResourceData sets Terraform ResourceData from a map of AWS API attributes.
func (m AttributeMap) APIAttributesToResourceData(apiAttributes map[string]string, d *schema.ResourceData) error {
	for tfAttributeName, attributeInfo := range m {
		if v, ok := apiAttributes[attributeInfo.apiAttributeName]; ok {
			var err error
			var tfAttributeValue interface{}

			switch t := attributeInfo.tfType; t {
			case schema.TypeBool:
				tfAttributeValue, err = strconv.ParseBool(v)

				if err != nil {
					return fmt.Errorf("parsing %s value (%s) into boolean: %w", tfAttributeName, v, err)
				}
			case schema.TypeInt:
				tfAttributeValue, err = strconv.Atoi(v)

				if err != nil {
					return fmt.Errorf("parsing %s value (%s) into integer: %w", tfAttributeName, v, err)
				}
			case schema.TypeString:
				tfAttributeValue = v

				if attributeInfo.isIAMPolicy {
					policy, err := PolicyToSet(d.Get(tfAttributeName).(string), tfAttributeValue.(string))
					if err != nil {
						return err
					}

					tfAttributeValue = policy
				}
			default:
				return fmt.Errorf("attribute %s is of unsupported type: %d", tfAttributeName, t)
			}

			if err := d.Set(tfAttributeName, tfAttributeValue); err != nil {
				return fmt.Errorf("setting %s: %w", tfAttributeName, err)
			}
		} else if attributeInfo.missingSetToNil {
			_ = d.Set(tfAttributeName, nil)
		}
	}

	return nil
}

// NamePrefixFromName returns a name prefix if the string matches prefix criteria
//
// The input to this function must be strictly the "name" and not any
// additional information such as a full Amazon Resource Name (ARN).
//
// An expected usage might be:
//
//	d.Set("name_prefix", create.NamePrefixFromName(d.Id()))
func NamePrefixFromName(name string) *string {
	return NamePrefixFromNameWithSuffix(name, "")
}

// hasResourceUniqueIDPlusAdditionalSuffix returns true if the string has the built-in unique ID suffix plus an additional suffix
func hasResourceUniqueIDPlusAdditionalSuffix(s string, additionalSuffix string) bool {
	re := regexp.MustCompile(fmt.Sprintf("[[:xdigit:]]{%d}%s$", id.UniqueIDSuffixLength, additionalSuffix))
	return re.MatchString(s)
}

func NamePrefixFromNameWithSuffix(name, nameSuffix string) *string {
	if !hasResourceUniqueIDPlusAdditionalSuffix(name, nameSuffix) {
		return nil
	}

	namePrefixIndex := len(name) - id.UniqueIDSuffixLength - len(nameSuffix)

	if namePrefixIndex <= 0 {
		return nil
	}

	namePrefix := name[:namePrefixIndex]

	return &namePrefix
}

func resourceQueueCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	fifoQueue := diff.Get("fifo_queue").(bool)
	contentBasedDeduplication := diff.Get("content_based_deduplication").(bool)

	if diff.Id() == "" {
		var name string

		if fifoQueue {
			name = NameWithSuffix(diff.Get("name").(string), diff.Get("name_prefix").(string), FIFOQueueNameSuffix)
		} else {
			name = Name(diff.Get("name").(string), diff.Get("name_prefix").(string))
		}

		var re *regexp.Regexp

		if fifoQueue {
			re = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,75}\.fifo$`)
		} else {
			re = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,80}$`)
		}

		if !re.MatchString(name) {
			return fmt.Errorf("invalid queue name: %s", name)
		}
	}

	if !fifoQueue && contentBasedDeduplication {
		return fmt.Errorf("content-based deduplication can only be set for FIFO queue")
	}

	return nil
}
