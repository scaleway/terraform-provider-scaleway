package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
	"github.com/scaleway/terraform-provider-scaleway/v2/scaleway"
)

var terraformBetaEnabled = os.Getenv(scw.ScwEnableBeta) != ""

// Config can be used to provide additional config when creating provider.
type Config struct {
	// Meta can be used to override Meta that will be used by the provider.
	// This is useful for tests.
	Meta *meta.Meta
}

// DefaultConfig return default Config struct
func DefaultConfig() *Config {
	return &Config{}
}

func addBetaResources(provider *schema.Provider) {
	if !terraformBetaEnabled {
		return
	}
	betaResources := map[string]*schema.Resource{}
	betaDataSources := map[string]*schema.Resource{}
	for resourceName, resource := range betaResources {
		provider.ResourcesMap[resourceName] = resource
	}
	for resourceName, resource := range betaDataSources {
		provider.DataSourcesMap[resourceName] = resource
	}
}

// Provider returns a terraform.ResourceProvider.
func Provider(config *Config) plugin.ProviderFunc {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"access_key": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The Scaleway access key.",
				},
				"secret_key": {
					Type:         schema.TypeString,
					Optional:     true, // To allow user to use deprecated `token`.
					Description:  "The Scaleway secret Key.",
					ValidateFunc: verify.IsUUID(),
				},
				"profile": {
					Type:        schema.TypeString,
					Optional:    true, // To allow user to use `access_key`, `secret_key`, `project_id`...
					Description: "The Scaleway profile to use.",
				},
				"project_id": {
					Type:         schema.TypeString,
					Optional:     true, // To allow user to use organization instead of project
					Description:  "The Scaleway project ID.",
					ValidateFunc: verify.IsUUID(),
				},
				"organization_id": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "The Scaleway organization ID.",
					ValidateFunc: verify.IsUUID(),
				},
				"region": regional.Schema(),
				"zone":   zonal.Schema(),
				"api_url": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The Scaleway API URL to use.",
				},
			},

			ResourcesMap: map[string]*schema.Resource{
				"scaleway_account_project":                     scaleway.ResourceScalewayAccountProject(),
				"scaleway_account_ssh_key":                     scaleway.ResourceScalewayAccountSSKKey(),
				"scaleway_apple_silicon_server":                scaleway.ResourceScalewayAppleSiliconServer(),
				"scaleway_baremetal_server":                    scaleway.ResourceScalewayBaremetalServer(),
				"scaleway_block_volume":                        scaleway.ResourceScalewayBlockVolume(),
				"scaleway_block_snapshot":                      scaleway.ResourceScalewayBlockSnapshot(),
				"scaleway_cockpit":                             scaleway.ResourceScalewayCockpit(),
				"scaleway_cockpit_token":                       scaleway.ResourceScalewayCockpitToken(),
				"scaleway_cockpit_grafana_user":                scaleway.ResourceScalewayCockpitGrafanaUser(),
				"scaleway_container_namespace":                 scaleway.ResourceScalewayContainerNamespace(),
				"scaleway_container_cron":                      scaleway.ResourceScalewayContainerCron(),
				"scaleway_container_domain":                    scaleway.ResourceScalewayContainerDomain(),
				"scaleway_container_trigger":                   scaleway.ResourceScalewayContainerTrigger(),
				"scaleway_documentdb_instance":                 scaleway.ResourceScalewayDocumentDBInstance(),
				"scaleway_documentdb_database":                 scaleway.ResourceScalewayDocumentDBDatabase(),
				"scaleway_documentdb_private_network_endpoint": scaleway.ResourceScalewayDocumentDBInstancePrivateNetworkEndpoint(),
				"scaleway_documentdb_user":                     scaleway.ResourceScalewayDocumentDBUser(),
				"scaleway_documentdb_privilege":                scaleway.ResourceScalewayDocumentDBPrivilege(),
				"scaleway_documentdb_read_replica":             scaleway.ResourceScalewayDocumentDBReadReplica(),
				"scaleway_domain_record":                       scaleway.ResourceScalewayDomainRecord(),
				"scaleway_domain_zone":                         scaleway.ResourceScalewayDomainZone(),
				"scaleway_flexible_ip":                         scaleway.ResourceScalewayFlexibleIP(),
				"scaleway_flexible_ip_mac_address":             scaleway.ResourceScalewayFlexibleIPMACAddress(),
				"scaleway_function":                            scaleway.ResourceScalewayFunction(),
				"scaleway_function_cron":                       scaleway.ResourceScalewayFunctionCron(),
				"scaleway_function_domain":                     scaleway.ResourceScalewayFunctionDomain(),
				"scaleway_function_namespace":                  scaleway.ResourceScalewayFunctionNamespace(),
				"scaleway_function_token":                      scaleway.ResourceScalewayFunctionToken(),
				"scaleway_function_trigger":                    scaleway.ResourceScalewayFunctionTrigger(),
				"scaleway_iam_api_key":                         scaleway.ResourceScalewayIamAPIKey(),
				"scaleway_iam_application":                     scaleway.ResourceScalewayIamApplication(),
				"scaleway_iam_group":                           scaleway.ResourceScalewayIamGroup(),
				"scaleway_iam_group_membership":                scaleway.ResourceScalewayIamGroupMembership(),
				"scaleway_iam_policy":                          scaleway.ResourceScalewayIamPolicy(),
				"scaleway_iam_user":                            scaleway.ResourceScalewayIamUser(),
				"scaleway_instance_user_data":                  scaleway.ResourceScalewayInstanceUserData(),
				"scaleway_instance_image":                      scaleway.ResourceScalewayInstanceImage(),
				"scaleway_instance_ip":                         scaleway.ResourceScalewayInstanceIP(),
				"scaleway_instance_ip_reverse_dns":             scaleway.ResourceScalewayInstanceIPReverseDNS(),
				"scaleway_instance_volume":                     scaleway.ResourceScalewayInstanceVolume(),
				"scaleway_instance_security_group":             scaleway.ResourceScalewayInstanceSecurityGroup(),
				"scaleway_instance_security_group_rules":       scaleway.ResourceScalewayInstanceSecurityGroupRules(),
				"scaleway_instance_server":                     scaleway.ResourceScalewayInstanceServer(),
				"scaleway_instance_snapshot":                   scaleway.ResourceScalewayInstanceSnapshot(),
				"scaleway_iam_ssh_key":                         scaleway.ResourceScalewayIamSSKKey(),
				"scaleway_instance_placement_group":            scaleway.ResourceScalewayInstancePlacementGroup(),
				"scaleway_instance_private_nic":                scaleway.ResourceScalewayInstancePrivateNIC(),
				"scaleway_iot_hub":                             scaleway.ResourceScalewayIotHub(),
				"scaleway_iot_device":                          scaleway.ResourceScalewayIotDevice(),
				"scaleway_iot_route":                           scaleway.ResourceScalewayIotRoute(),
				"scaleway_iot_network":                         scaleway.ResourceScalewayIotNetwork(),
				"scaleway_ipam_ip":                             scaleway.ResourceScalewayIPAMIP(),
				"scaleway_ipam_ip_reverse_dns":                 scaleway.ResourceScalewayIPAMIPReverseDNS(),
				"scaleway_job_definition":                      scaleway.ResourceScalewayJobDefinition(),
				"scaleway_k8s_cluster":                         scaleway.ResourceScalewayK8SCluster(),
				"scaleway_k8s_pool":                            scaleway.ResourceScalewayK8SPool(),
				"scaleway_lb":                                  scaleway.ResourceScalewayLb(),
				"scaleway_lb_acl":                              scaleway.ResourceScalewayLbACL(),
				"scaleway_lb_ip":                               scaleway.ResourceScalewayLbIP(),
				"scaleway_lb_backend":                          scaleway.ResourceScalewayLbBackend(),
				"scaleway_lb_certificate":                      scaleway.ResourceScalewayLbCertificate(),
				"scaleway_lb_frontend":                         scaleway.ResourceScalewayLbFrontend(),
				"scaleway_lb_route":                            scaleway.ResourceScalewayLbRoute(),
				"scaleway_registry_namespace":                  scaleway.ResourceScalewayRegistryNamespace(),
				"scaleway_tem_domain":                          scaleway.ResourceScalewayTemDomain(),
				"scaleway_container":                           scaleway.ResourceScalewayContainer(),
				"scaleway_container_token":                     scaleway.ResourceScalewayContainerToken(),
				"scaleway_rdb_acl":                             scaleway.ResourceScalewayRdbACL(),
				"scaleway_rdb_database":                        scaleway.ResourceScalewayRdbDatabase(),
				"scaleway_rdb_database_backup":                 scaleway.ResourceScalewayRdbDatabaseBackup(),
				"scaleway_rdb_instance":                        scaleway.ResourceScalewayRdbInstance(),
				"scaleway_rdb_privilege":                       scaleway.ResourceScalewayRdbPrivilege(),
				"scaleway_rdb_user":                            scaleway.ResourceScalewayRdbUser(),
				"scaleway_rdb_read_replica":                    scaleway.ResourceScalewayRdbReadReplica(),
				"scaleway_redis_cluster":                       scaleway.ResourceScalewayRedisCluster(),
				"scaleway_sdb_sql_database":                    scaleway.ResourceScalewaySDBSQLDatabase(),
				"scaleway_object":                              scaleway.ResourceScalewayObject(),
				"scaleway_object_bucket":                       scaleway.ResourceScalewayObjectBucket(),
				"scaleway_object_bucket_acl":                   scaleway.ResourceScalewayObjectBucketACL(),
				"scaleway_object_bucket_lock_configuration":    scaleway.ResourceObjectLockConfiguration(),
				"scaleway_object_bucket_policy":                scaleway.ResourceScalewayObjectBucketPolicy(),
				"scaleway_object_bucket_website_configuration": scaleway.ResourceBucketWebsiteConfiguration(),
				"scaleway_mnq_nats_account":                    scaleway.ResourceScalewayMNQNatsAccount(),
				"scaleway_mnq_nats_credentials":                scaleway.ResourceScalewayMNQNatsCredentials(),
				"scaleway_mnq_sns":                             scaleway.ResourceScalewayMNQSNS(),
				"scaleway_mnq_sns_credentials":                 scaleway.ResourceScalewayMNQSNSCredentials(),
				"scaleway_mnq_sns_topic":                       scaleway.ResourceScalewayMNQSNSTopic(),
				"scaleway_mnq_sns_topic_subscription":          scaleway.ResourceScalewayMNQSNSTopicSubscription(),
				"scaleway_mnq_sqs":                             scaleway.ResourceScalewayMNQSQS(),
				"scaleway_mnq_sqs_queue":                       scaleway.ResourceScalewayMNQSQSQueue(),
				"scaleway_mnq_sqs_credentials":                 scaleway.ResourceScalewayMNQSQSCredentials(),
				"scaleway_secret":                              scaleway.ResourceScalewaySecret(),
				"scaleway_secret_version":                      scaleway.ResourceScalewaySecretVersion(),
				"scaleway_vpc":                                 scaleway.ResourceScalewayVPC(),
				"scaleway_vpc_public_gateway":                  scaleway.ResourceScalewayVPCPublicGateway(),
				"scaleway_vpc_gateway_network":                 scaleway.ResourceScalewayVPCGatewayNetwork(),
				"scaleway_vpc_public_gateway_dhcp":             scaleway.ResourceScalewayVPCPublicGatewayDHCP(),
				"scaleway_vpc_public_gateway_dhcp_reservation": scaleway.ResourceScalewayVPCPublicGatewayDHCPReservation(),
				"scaleway_vpc_public_gateway_ip":               scaleway.ResourceScalewayVPCPublicGatewayIP(),
				"scaleway_vpc_public_gateway_ip_reverse_dns":   scaleway.ResourceScalewayVPCPublicGatewayIPReverseDNS(),
				"scaleway_vpc_public_gateway_pat_rule":         scaleway.ResourceScalewayVPCPublicGatewayPATRule(),
				"scaleway_vpc_private_network":                 scaleway.ResourceScalewayVPCPrivateNetwork(),
				"scaleway_webhosting":                          scaleway.ResourceScalewayWebhosting(),
			},

			DataSourcesMap: map[string]*schema.Resource{
				"scaleway_account_project":                     scaleway.DataSourceScalewayAccountProject(),
				"scaleway_account_ssh_key":                     scaleway.DataSourceScalewayAccountSSHKey(),
				"scaleway_availability_zones":                  scaleway.DataSourceAvailabilityZones(),
				"scaleway_baremetal_offer":                     scaleway.DataSourceScalewayBaremetalOffer(),
				"scaleway_baremetal_option":                    scaleway.DataSourceScalewayBaremetalOption(),
				"scaleway_baremetal_os":                        scaleway.DataSourceScalewayBaremetalOs(),
				"scaleway_baremetal_server":                    scaleway.DataSourceScalewayBaremetalServer(),
				"scaleway_billing_invoices":                    scaleway.DataSourceScalewayBillingInvoices(),
				"scaleway_billing_consumptions":                scaleway.DataSourceScalewayBillingConsumptions(),
				"scaleway_block_volume":                        scaleway.DataSourceScalewayBlockVolume(),
				"scaleway_block_snapshot":                      scaleway.DataSourceScalewayBlockSnapshot(),
				"scaleway_cockpit":                             scaleway.DataSourceScalewayCockpit(),
				"scaleway_cockpit_plan":                        scaleway.DataSourceScalewayCockpitPlan(),
				"scaleway_domain_record":                       scaleway.DataSourceScalewayDomainRecord(),
				"scaleway_domain_zone":                         scaleway.DataSourceScalewayDomainZone(),
				"scaleway_documentdb_instance":                 scaleway.DataSourceScalewayDocumentDBInstance(),
				"scaleway_documentdb_database":                 scaleway.DataSourceScalewayDocumentDBDatabase(),
				"scaleway_documentdb_load_balancer_endpoint":   scaleway.DataSourceScalewayDocumentDBEndpointLoadBalancer(),
				"scaleway_container_namespace":                 scaleway.DataSourceScalewayContainerNamespace(),
				"scaleway_container":                           scaleway.DataSourceScalewayContainer(),
				"scaleway_function":                            scaleway.DataSourceScalewayFunction(),
				"scaleway_function_namespace":                  scaleway.DataSourceScalewayFunctionNamespace(),
				"scaleway_iam_application":                     scaleway.DataSourceScalewayIamApplication(),
				"scaleway_flexible_ip":                         scaleway.DataSourceScalewayFlexibleIP(),
				"scaleway_flexible_ips":                        scaleway.DataSourceScalewayFlexibleIPs(),
				"scaleway_iam_group":                           scaleway.DataSourceScalewayIamGroup(),
				"scaleway_iam_ssh_key":                         scaleway.DataSourceScalewayIamSSHKey(),
				"scaleway_iam_user":                            scaleway.DataSourceScalewayIamUser(),
				"scaleway_instance_ip":                         scaleway.DataSourceScalewayInstanceIP(),
				"scaleway_instance_placement_group":            scaleway.DataSourceScalewayInstancePlacementGroup(),
				"scaleway_instance_private_nic":                scaleway.DataSourceScalewayInstancePrivateNIC(),
				"scaleway_instance_security_group":             scaleway.DataSourceScalewayInstanceSecurityGroup(),
				"scaleway_instance_server":                     scaleway.DataSourceScalewayInstanceServer(),
				"scaleway_instance_servers":                    scaleway.DataSourceScalewayInstanceServers(),
				"scaleway_instance_image":                      scaleway.DataSourceScalewayInstanceImage(),
				"scaleway_instance_volume":                     scaleway.DataSourceScalewayInstanceVolume(),
				"scaleway_instance_snapshot":                   scaleway.DataSourceScalewayInstanceSnapshot(),
				"scaleway_iot_hub":                             scaleway.DataSourceScalewayIotHub(),
				"scaleway_iot_device":                          scaleway.DataSourceScalewayIotDevice(),
				"scaleway_ipam_ip":                             scaleway.DataSourceScalewayIPAMIP(),
				"scaleway_ipam_ips":                            scaleway.DataSourceScalewayIPAMIPs(),
				"scaleway_k8s_cluster":                         scaleway.DataSourceScalewayK8SCluster(),
				"scaleway_k8s_pool":                            scaleway.DataSourceScalewayK8SPool(),
				"scaleway_k8s_version":                         scaleway.DataSourceScalewayK8SVersion(),
				"scaleway_lb":                                  scaleway.DataSourceScalewayLb(),
				"scaleway_lbs":                                 scaleway.DataSourceScalewayLbs(),
				"scaleway_lb_acls":                             scaleway.DataSourceScalewayLbACLs(),
				"scaleway_lb_backend":                          scaleway.DataSourceScalewayLbBackend(),
				"scaleway_lb_backends":                         scaleway.DataSourceScalewayLbBackends(),
				"scaleway_lb_certificate":                      scaleway.DataSourceScalewayLbCertificate(),
				"scaleway_lb_frontend":                         scaleway.DataSourceScalewayLbFrontend(),
				"scaleway_lb_frontends":                        scaleway.DataSourceScalewayLbFrontends(),
				"scaleway_lb_ip":                               scaleway.DataSourceScalewayLbIP(),
				"scaleway_lb_ips":                              scaleway.DataSourceScalewayLbIPs(),
				"scaleway_lb_route":                            scaleway.DataSourceScalewayLbRoute(),
				"scaleway_lb_routes":                           scaleway.DataSourceScalewayLbRoutes(),
				"scaleway_marketplace_image":                   scaleway.DataSourceScalewayMarketplaceImage(),
				"scaleway_mnq_sqs":                             scaleway.DataSourceScalewayMNQSQS(),
				"scaleway_object_bucket":                       scaleway.DataSourceScalewayObjectBucket(),
				"scaleway_object_bucket_policy":                scaleway.DataSourceScalewayObjectBucketPolicy(),
				"scaleway_rdb_acl":                             scaleway.DataSourceScalewayRDBACL(),
				"scaleway_rdb_instance":                        scaleway.DataSourceScalewayRDBInstance(),
				"scaleway_rdb_database":                        scaleway.DataSourceScalewayRDBDatabase(),
				"scaleway_rdb_database_backup":                 scaleway.DataSourceScalewayRDBDatabaseBackup(),
				"scaleway_rdb_privilege":                       scaleway.DataSourceScalewayRDBPrivilege(),
				"scaleway_redis_cluster":                       scaleway.DataSourceScalewayRedisCluster(),
				"scaleway_registry_namespace":                  scaleway.DataSourceScalewayRegistryNamespace(),
				"scaleway_tem_domain":                          scaleway.DataSourceScalewayTemDomain(),
				"scaleway_secret":                              scaleway.DataSourceScalewaySecret(),
				"scaleway_secret_version":                      scaleway.DataSourceScalewaySecretVersion(),
				"scaleway_registry_image":                      scaleway.DataSourceScalewayRegistryImage(),
				"scaleway_vpc":                                 scaleway.DataSourceScalewayVPC(),
				"scaleway_vpcs":                                scaleway.DataSourceScalewayVPCs(),
				"scaleway_vpc_public_gateway":                  scaleway.DataSourceScalewayVPCPublicGateway(),
				"scaleway_vpc_gateway_network":                 scaleway.DataSourceScalewayVPCGatewayNetwork(),
				"scaleway_vpc_public_gateway_dhcp":             scaleway.DataSourceScalewayVPCPublicGatewayDHCP(),
				"scaleway_vpc_public_gateway_dhcp_reservation": scaleway.DataSourceScalewayVPCPublicGatewayDHCPReservation(),
				"scaleway_vpc_public_gateway_ip":               scaleway.DataSourceScalewayVPCPublicGatewayIP(),
				"scaleway_vpc_private_network":                 scaleway.DataSourceScalewayVPCPrivateNetwork(),
				"scaleway_vpc_public_gateway_pat_rule":         scaleway.DataSourceScalewayVPCPublicGatewayPATRule(),
				"scaleway_webhosting":                          scaleway.DataSourceScalewayWebhosting(),
				"scaleway_webhosting_offer":                    scaleway.DataSourceScalewayWebhostingOffer(),
			},
		}

		addBetaResources(p)

		p.ConfigureContextFunc = func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
			terraformVersion := p.TerraformVersion

			// If we provide meta in config use it. This is useful for tests
			if config.Meta != nil {
				return config.Meta, nil
			}

			m, err := meta.NewMeta(ctx, &meta.Config{
				ProviderSchema:   data,
				TerraformVersion: terraformVersion,
			})
			if err != nil {
				return nil, diag.FromErr(err)
			}
			return m, nil
		}

		return p
	}
}

//gocyclo:ignore
