package scaleway

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	sns "github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func TestAccScalewayMNQSNSTopic_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayMNQSNSTopicDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_account_project main {
						name = "tf_tests_mnq_sns_topic_basic"
					}

					resource scaleway_mnq_sns main {
						project_id = scaleway_account_project.main.id
					}

					resource scaleway_mnq_sns_credentials main {
						project_id = scaleway_mnq_sns.main.project_id
						permissions {
							can_manage = true
						}
					}

					resource scaleway_mnq_sns_topic main {
						project_id = scaleway_mnq_sns.main.project_id
						name = "test-mnq-sns-topic-basic"
						access_key = scaleway_mnq_sns_credentials.main.access_key
						secret_key = scaleway_mnq_sns_credentials.main.secret_key
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayMNQSNSTopicExists(tt, "scaleway_mnq_sns_topic.main"),
					testCheckResourceAttrUUID("scaleway_mnq_sns_topic.main", "id"),
					resource.TestCheckResourceAttr("scaleway_mnq_sns_topic.main", "name", "test-mnq-sns-topic-basic"),
				),
			},
			{
				Config: `
					resource scaleway_account_project main {
						name = "tf_tests_mnq_sns_topic_basic"
					}

					resource scaleway_mnq_sns main {
						project_id = scaleway_account_project.main.id
					}

					resource scaleway_mnq_sns_credentials main {
						project_id = scaleway_mnq_sns.main.project_id
						permissions {
							can_manage = true
						}
					}

					resource scaleway_mnq_sns_topic main {
						project_id = scaleway_mnq_sns.main.project_id
						name = "test-mnq-sns-topic-basic.fifo"
						access_key = scaleway_mnq_sns_credentials.main.access_key
						secret_key = scaleway_mnq_sns_credentials.main.secret_key
						fifo_topic = true
						content_based_deduplication = true
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayMNQSNSTopicExists(tt, "scaleway_mnq_sns_topic.main"),
					testCheckResourceAttrUUID("scaleway_mnq_sns_topic.main", "id"),
					resource.TestCheckResourceAttr("scaleway_mnq_sns_topic.main", "name", "test-mnq-sns-topic-basic.fifo"),
					resource.TestCheckResourceAttr("scaleway_mnq_sns_topic.main", "content_based_deduplication", "true"),
					resource.TestCheckResourceAttr("scaleway_mnq_sns_topic.main", "fifo_topic", "true"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_sns_topic.main", "arn"),
				),
			},
			{
				Config: `
						resource scaleway_account_project main {
							name = "tf_tests_mnq_sns_topic_basic"
						}

						resource scaleway_mnq_sns main {
							project_id = scaleway_account_project.main.id
						}

						resource scaleway_mnq_sns_credentials main {
							project_id = scaleway_mnq_sns.main.project_id
							permissions {
								can_manage = true
							}
						}

						resource scaleway_mnq_sns_topic main {
							project_id = scaleway_mnq_sns.main.project_id
							name_prefix = "test-mnq-sns-topic-basic"
							access_key = scaleway_mnq_sns_credentials.main.access_key
							secret_key = scaleway_mnq_sns_credentials.main.secret_key
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayMNQSNSTopicExists(tt, "scaleway_mnq_sns_topic.main"),
					testCheckResourceAttrUUID("scaleway_mnq_sns_topic.main", "id"),
					func(state *terraform.State) error {
						topic, exists := state.RootModule().Resources["scaleway_mnq_sns_topic.main"]
						if !exists {
							return errors.New("failed to find resource")
						}
						name, exists := topic.Primary.Attributes["name"]
						if !exists {
							return errors.New("failed to find atttribute")
						}

						if !strings.HasPrefix(name, "test-mnq-sns-topic-basic") {
							return fmt.Errorf("invalid name %q", name)
						}

						return nil
					},
					resource.TestCheckResourceAttrSet("scaleway_mnq_sns_topic.main", "arn"),
				),
			},
		},
	})
}

func testAccCheckScalewayMNQSNSTopicExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		region, projectID, topicName, err := decomposeMNQID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to parse id: %w", err)
		}

		snsClient, err := newSNSClient(tt.Meta.HttpClient(), region.String(), rs.Primary.Attributes["sns_endpoint"], rs.Primary.Attributes["access_key"], rs.Primary.Attributes["secret_key"])
		if err != nil {
			return err
		}

		_, err = snsClient.GetTopicAttributes(&sns.GetTopicAttributesInput{
			TopicArn: scw.StringPtr(composeSNSARN(region, projectID, topicName)),
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayMNQSNSTopicDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_mnq_sns_topic" {
				continue
			}

			region, projectID, topicName, err := decomposeMNQID(rs.Primary.ID)
			if err != nil {
				return fmt.Errorf("failed to parse id: %w", err)
			}

			snsClient, err := newSNSClient(tt.Meta.HttpClient(), region.String(), rs.Primary.Attributes["sns_endpoint"], rs.Primary.Attributes["access_key"], rs.Primary.Attributes["secret_key"])
			if err != nil {
				return err
			}

			_, err = snsClient.GetTopicAttributes(&sns.GetTopicAttributesInput{
				TopicArn: scw.StringPtr(composeSNSARN(region, projectID, topicName)),
			})
			if err != nil {
				if tfawserr.ErrCodeEquals(err, "AccessDeniedException") {
					return nil
				}
				return err
			}

			return fmt.Errorf("mnq sns topic (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}
