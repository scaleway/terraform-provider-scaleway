package provider_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	iamchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/iam/testfuncs"
	instancechecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/testfuncs"
	"github.com/scaleway/terraform-provider-scaleway/v2/provider"
	"github.com/stretchr/testify/require"
)

func TestSDKProvider_SchemaFuncIsUsed(t *testing.T) {
	p := provider.SDKProvider(nil)()
	for name, d := range p.ResourcesMap {
		if d.SchemaFunc == nil {
			t.Errorf("SchemaFunc for resource %s is nil", name)
		}

		if d.Schema != nil {
			t.Errorf("Schema for resource %s is %v, want nil", name, d.Schema)
		}
	}
}

func TestAccProvider_InstanceIPZones(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	ctx := t.Context()

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: func() map[string]func() (*schema.Provider, error) {
			metaProd, err := meta.NewMeta(ctx, &meta.Config{
				TerraformVersion: "terraform-tests",
				ForceZone:        scw.ZoneFrPar2,
				HTTPClient:       tt.Meta.HTTPClient(),
			})
			require.NoError(t, err)

			metaDev, err := meta.NewMeta(ctx, &meta.Config{
				TerraformVersion: "terraform-tests",
				ForceZone:        scw.ZoneFrPar1,
				HTTPClient:       tt.Meta.HTTPClient(),
			})
			require.NoError(t, err)

			return map[string]func() (*schema.Provider, error){
				"prod": func() (*schema.Provider, error) {
					return provider.SDKProvider(&provider.Config{Meta: metaProd})(), nil
				},
				"dev": func() (*schema.Provider, error) {
					return provider.SDKProvider(&provider.Config{Meta: metaDev})(), nil
				},
			}
		}(),
		CheckDestroy: iamchecks.CheckSSHKeyDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_instance_ip dev {
					  provider = "dev"
					}

					resource scaleway_instance_ip prod {
					  provider = "prod"
					}
`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.CheckIPExists(tt, "scaleway_instance_ip.prod"),
					instancechecks.CheckIPExists(tt, "scaleway_instance_ip.dev"),
					resource.TestCheckResourceAttr("scaleway_instance_ip.prod", "zone", "fr-par-2"),
					resource.TestCheckResourceAttr("scaleway_instance_ip.dev", "zone", "fr-par-1"),
				),
			},
		},
	})
}

func TestAccProvider_SSHKeys(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	SSHKeyName := "TestAccScalewayProvider_SSHKeys"
	SSHKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEEYrzDOZmhItdKaDAEqJQ4ORS2GyBMtBozYsK5kiXXX opensource@scaleway.com"

	ctx := t.Context()

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: func() map[string]func() (*schema.Provider, error) {
			metaProd, err := meta.NewMeta(ctx, &meta.Config{
				TerraformVersion: "terraform-tests",
				HTTPClient:       tt.Meta.HTTPClient(),
			})
			require.NoError(t, err)

			metaDev, err := meta.NewMeta(ctx, &meta.Config{
				TerraformVersion: "terraform-tests",
				HTTPClient:       tt.Meta.HTTPClient(),
			})
			require.NoError(t, err)

			return map[string]func() (*schema.Provider, error){
				"prod": func() (*schema.Provider, error) {
					return provider.SDKProvider(&provider.Config{Meta: metaProd})(), nil
				},
				"dev": func() (*schema.Provider, error) {
					return provider.SDKProvider(&provider.Config{Meta: metaDev})(), nil
				},
			}
		}(),
		CheckDestroy: iamchecks.CheckSSHKeyDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_account_ssh_key" "prod" {
						provider   = "prod"
						name 	   = "%[1]s"
						public_key = "%[2]s"
					}

					resource "scaleway_account_ssh_key" "dev" {
						provider   = "dev"
						name 	   = "%[1]s"
						public_key = "%[2]s"
					}
				`, SSHKeyName, SSHKey),
				Check: resource.ComposeTestCheckFunc(
					iamchecks.CheckSSHKeyExists(tt, "scaleway_account_ssh_key.prod"),
					iamchecks.CheckSSHKeyExists(tt, "scaleway_account_ssh_key.dev"),
				),
			},
		},
	})
}

func TestSDKProvider_ResourceIdentityNotEmpty(t *testing.T) {
	exceptions := []string{
		"scaleway_account_project",
		"scaleway_account_ssh_key",
		"scaleway_autoscaling_instance_group",
		"scaleway_autoscaling_instance_policy",
		"scaleway_autoscaling_instance_template",
		"scaleway_baremetal_server",
		"scaleway_cockpit",
		"scaleway_cockpit_source",
		"scaleway_cockpit_grafana_user",
		"scaleway_cockpit_token",
		"scaleway_cockpit_alert_manager",
		"scaleway_container",
		"scaleway_container_cron",
		"scaleway_container_domain",
		"scaleway_container_namespace",
		"scaleway_container_token",
		"scaleway_container_trigger",
		"scaleway_datawarehouse_deployment",
		"scaleway_datawarehouse_user",
		"scaleway_datawarehouse_database",
		"scaleway_domain_record",
		"scaleway_domain_registration",
		"scaleway_domain_zone",
		"scaleway_edge_services_backend_stage",
		"scaleway_edge_services_cache_stage",
		"scaleway_edge_services_dns_stage",
		"scaleway_edge_services_head_stage",
		"scaleway_edge_services_pipeline",
		"scaleway_edge_services_plan",
		"scaleway_edge_services_route_stage",
		"scaleway_edge_services_tls_stage",
		"scaleway_edge_services_waf_stage",
		"scaleway_file_filesystem",
		"scaleway_flexible_ip",
		"scaleway_flexible_ip_mac_address",
		"scaleway_function",
		"scaleway_function_cron",
		"scaleway_function_domain",
		"scaleway_function_namespace",
		"scaleway_function_token",
		"scaleway_function_trigger",
		"scaleway_iam_api_key",
		"scaleway_iam_application",
		"scaleway_iam_group",
		"scaleway_iam_group_membership",
		"scaleway_iam_policy",
		"scaleway_iam_ssh_key",
		"scaleway_iam_user",
		"scaleway_inference_deployment",
		"scaleway_inference_model",
		"scaleway_instance_image",
		"scaleway_instance_ip",
		"scaleway_instance_ip_reverse_dns",
		"scaleway_instance_placement_group",
		"scaleway_instance_private_nic",
		"scaleway_instance_security_group",
		"scaleway_instance_security_group_rules",
		"scaleway_instance_server",
		"scaleway_instance_snapshot",
		"scaleway_instance_user_data",
		"scaleway_instance_volume",
		"scaleway_iot_device",
		"scaleway_iot_hub",
		"scaleway_iot_network",
		"scaleway_iot_route",
		"scaleway_ipam_ip",
		"scaleway_ipam_ip_reverse_dns",
		"scaleway_job_definition",
		"scaleway_k8s_acl",
		"scaleway_k8s_cluster",
		"scaleway_k8s_pool",
		"scaleway_key_manager_key",
		"scaleway_lb",
		"scaleway_lb_acl",
		"scaleway_lb_backend",
		"scaleway_lb_certificate",
		"scaleway_lb_frontend",
		"scaleway_lb_ip",
		"scaleway_lb_private_network",
		"scaleway_lb_route",
		"scaleway_mnq_nats_account",
		"scaleway_mnq_nats_credentials",
		"scaleway_mnq_sns",
		"scaleway_mnq_sns_credentials",
		"scaleway_mnq_sns_topic",
		"scaleway_mnq_sns_topic_subscription",
		"scaleway_mnq_sqs",
		"scaleway_mnq_sqs_credentials",
		"scaleway_mnq_sqs_queue",
		"scaleway_mongodb_instance",
		"scaleway_mongodb_snapshot",
		"scaleway_mongodb_user",
		"scaleway_object",
		"scaleway_object_bucket",
		"scaleway_object_bucket_acl",
		"scaleway_object_bucket_lock_configuration",
		"scaleway_object_bucket_policy",
		"scaleway_object_bucket_website_configuration",
		"scaleway_rdb_acl",
		"scaleway_rdb_database",
		"scaleway_rdb_database_backup",
		"scaleway_rdb_instance",
		"scaleway_rdb_privilege",
		"scaleway_rdb_read_replica",
		"scaleway_rdb_user",
		"scaleway_rdb_snapshot",
		"scaleway_redis_cluster",
		"scaleway_registry_namespace",
		"scaleway_s2s_vpn_gateway",
		"scaleway_s2s_vpn_customer_gateway",
		"scaleway_s2s_vpn_connection",
		"scaleway_s2s_vpn_routing_policy",
		"scaleway_sdb_sql_database",
		"scaleway_secret",
		"scaleway_secret_version",
		"scaleway_tem_domain",
		"scaleway_tem_domain_validation",
		"scaleway_tem_blocked_list",
		"scaleway_tem_webhook",
		"scaleway_vpc_gateway_network",
		"scaleway_vpc_public_gateway",
		"scaleway_vpc_public_gateway_dhcp",
		"scaleway_vpc_public_gateway_dhcp_reservation",
		"scaleway_vpc_public_gateway_ip",
		"scaleway_vpc_public_gateway_ip_reverse_dns",
		"scaleway_vpc_public_gateway_pat_rule",
		"scaleway_webhosting",
	}

	p := provider.SDKProvider(nil)()
	for name, d := range p.ResourcesMap {
		if d.Identity == nil && !slices.Contains(exceptions, name) {
			t.Errorf("Identity for resource %s is nil", name)
		}
	}
}

func TestSDKProvider_ResourceImporterNotEmpty(t *testing.T) {
	p := provider.SDKProvider(nil)()
	for name, d := range p.ResourcesMap {
		if d.Importer == nil {
			t.Errorf("Importer for resource %s is nil", name)
		}
	}
}
