package scaleway

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
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
	"github.com/scaleway/terraform-provider-scaleway/v2/version"
)

var terraformBetaEnabled = os.Getenv(scw.ScwEnableBeta) != ""

// ProviderConfig config can be used to provide additional config when creating provider.
type ProviderConfig struct {
	// Meta can be used to override Meta that will be used by the provider.
	// This is useful for tests.
	Meta *meta.Meta
}

// DefaultProviderConfig return default ProviderConfig struct
func DefaultProviderConfig() *ProviderConfig {
	return &ProviderConfig{}
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
func Provider(config *ProviderConfig) plugin.ProviderFunc {
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
					ValidateFunc: validationUUID(),
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
					ValidateFunc: validationUUID(),
				},
				"organization_id": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "The Scaleway organization ID.",
					ValidateFunc: validationUUID(),
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
				"scaleway_account_project":                     resourceScalewayAccountProject(),
				"scaleway_account_ssh_key":                     resourceScalewayAccountSSKKey(),
				"scaleway_apple_silicon_server":                resourceScalewayAppleSiliconServer(),
				"scaleway_baremetal_server":                    resourceScalewayBaremetalServer(),
				"scaleway_block_volume":                        resourceScalewayBlockVolume(),
				"scaleway_block_snapshot":                      resourceScalewayBlockSnapshot(),
				"scaleway_cockpit":                             resourceScalewayCockpit(),
				"scaleway_cockpit_token":                       resourceScalewayCockpitToken(),
				"scaleway_cockpit_grafana_user":                resourceScalewayCockpitGrafanaUser(),
				"scaleway_container_namespace":                 resourceScalewayContainerNamespace(),
				"scaleway_container_cron":                      resourceScalewayContainerCron(),
				"scaleway_container_domain":                    resourceScalewayContainerDomain(),
				"scaleway_container_trigger":                   resourceScalewayContainerTrigger(),
				"scaleway_documentdb_instance":                 resourceScalewayDocumentDBInstance(),
				"scaleway_documentdb_database":                 resourceScalewayDocumentDBDatabase(),
				"scaleway_documentdb_private_network_endpoint": resourceScalewayDocumentDBInstancePrivateNetworkEndpoint(),
				"scaleway_documentdb_user":                     resourceScalewayDocumentDBUser(),
				"scaleway_documentdb_privilege":                resourceScalewayDocumentDBPrivilege(),
				"scaleway_documentdb_read_replica":             resourceScalewayDocumentDBReadReplica(),
				"scaleway_domain_record":                       resourceScalewayDomainRecord(),
				"scaleway_domain_zone":                         resourceScalewayDomainZone(),
				"scaleway_flexible_ip":                         resourceScalewayFlexibleIP(),
				"scaleway_flexible_ip_mac_address":             resourceScalewayFlexibleIPMACAddress(),
				"scaleway_function":                            resourceScalewayFunction(),
				"scaleway_function_cron":                       resourceScalewayFunctionCron(),
				"scaleway_function_domain":                     resourceScalewayFunctionDomain(),
				"scaleway_function_namespace":                  resourceScalewayFunctionNamespace(),
				"scaleway_function_token":                      resourceScalewayFunctionToken(),
				"scaleway_function_trigger":                    resourceScalewayFunctionTrigger(),
				"scaleway_iam_api_key":                         resourceScalewayIamAPIKey(),
				"scaleway_iam_application":                     resourceScalewayIamApplication(),
				"scaleway_iam_group":                           resourceScalewayIamGroup(),
				"scaleway_iam_group_membership":                resourceScalewayIamGroupMembership(),
				"scaleway_iam_policy":                          resourceScalewayIamPolicy(),
				"scaleway_iam_user":                            resourceScalewayIamUser(),
				"scaleway_instance_user_data":                  resourceScalewayInstanceUserData(),
				"scaleway_instance_image":                      resourceScalewayInstanceImage(),
				"scaleway_instance_ip":                         resourceScalewayInstanceIP(),
				"scaleway_instance_ip_reverse_dns":             resourceScalewayInstanceIPReverseDNS(),
				"scaleway_instance_volume":                     resourceScalewayInstanceVolume(),
				"scaleway_instance_security_group":             resourceScalewayInstanceSecurityGroup(),
				"scaleway_instance_security_group_rules":       resourceScalewayInstanceSecurityGroupRules(),
				"scaleway_instance_server":                     resourceScalewayInstanceServer(),
				"scaleway_instance_snapshot":                   resourceScalewayInstanceSnapshot(),
				"scaleway_iam_ssh_key":                         resourceScalewayIamSSKKey(),
				"scaleway_instance_placement_group":            resourceScalewayInstancePlacementGroup(),
				"scaleway_instance_private_nic":                resourceScalewayInstancePrivateNIC(),
				"scaleway_iot_hub":                             resourceScalewayIotHub(),
				"scaleway_iot_device":                          resourceScalewayIotDevice(),
				"scaleway_iot_route":                           resourceScalewayIotRoute(),
				"scaleway_iot_network":                         resourceScalewayIotNetwork(),
				"scaleway_ipam_ip":                             resourceScalewayIPAMIP(),
				"scaleway_ipam_ip_reverse_dns":                 resourceScalewayIPAMIPReverseDNS(),
				"scaleway_job_definition":                      resourceScalewayJobDefinition(),
				"scaleway_k8s_cluster":                         resourceScalewayK8SCluster(),
				"scaleway_k8s_pool":                            resourceScalewayK8SPool(),
				"scaleway_lb":                                  resourceScalewayLb(),
				"scaleway_lb_acl":                              resourceScalewayLbACL(),
				"scaleway_lb_ip":                               resourceScalewayLbIP(),
				"scaleway_lb_backend":                          resourceScalewayLbBackend(),
				"scaleway_lb_certificate":                      resourceScalewayLbCertificate(),
				"scaleway_lb_frontend":                         resourceScalewayLbFrontend(),
				"scaleway_lb_route":                            resourceScalewayLbRoute(),
				"scaleway_registry_namespace":                  resourceScalewayRegistryNamespace(),
				"scaleway_tem_domain":                          resourceScalewayTemDomain(),
				"scaleway_container":                           resourceScalewayContainer(),
				"scaleway_container_token":                     resourceScalewayContainerToken(),
				"scaleway_rdb_acl":                             resourceScalewayRdbACL(),
				"scaleway_rdb_database":                        resourceScalewayRdbDatabase(),
				"scaleway_rdb_database_backup":                 resourceScalewayRdbDatabaseBackup(),
				"scaleway_rdb_instance":                        resourceScalewayRdbInstance(),
				"scaleway_rdb_privilege":                       resourceScalewayRdbPrivilege(),
				"scaleway_rdb_user":                            resourceScalewayRdbUser(),
				"scaleway_rdb_read_replica":                    resourceScalewayRdbReadReplica(),
				"scaleway_redis_cluster":                       resourceScalewayRedisCluster(),
				"scaleway_sdb_sql_database":                    resourceScalewaySDBSQLDatabase(),
				"scaleway_object":                              resourceScalewayObject(),
				"scaleway_object_bucket":                       resourceScalewayObjectBucket(),
				"scaleway_object_bucket_acl":                   resourceScalewayObjectBucketACL(),
				"scaleway_object_bucket_lock_configuration":    resourceObjectLockConfiguration(),
				"scaleway_object_bucket_policy":                resourceScalewayObjectBucketPolicy(),
				"scaleway_object_bucket_website_configuration": ResourceBucketWebsiteConfiguration(),
				"scaleway_mnq_nats_account":                    resourceScalewayMNQNatsAccount(),
				"scaleway_mnq_nats_credentials":                resourceScalewayMNQNatsCredentials(),
				"scaleway_mnq_sns":                             resourceScalewayMNQSNS(),
				"scaleway_mnq_sns_credentials":                 resourceScalewayMNQSNSCredentials(),
				"scaleway_mnq_sns_topic":                       resourceScalewayMNQSNSTopic(),
				"scaleway_mnq_sns_topic_subscription":          resourceScalewayMNQSNSTopicSubscription(),
				"scaleway_mnq_sqs":                             resourceScalewayMNQSQS(),
				"scaleway_mnq_sqs_queue":                       resourceScalewayMNQSQSQueue(),
				"scaleway_mnq_sqs_credentials":                 resourceScalewayMNQSQSCredentials(),
				"scaleway_secret":                              resourceScalewaySecret(),
				"scaleway_secret_version":                      resourceScalewaySecretVersion(),
				"scaleway_vpc":                                 resourceScalewayVPC(),
				"scaleway_vpc_public_gateway":                  resourceScalewayVPCPublicGateway(),
				"scaleway_vpc_gateway_network":                 resourceScalewayVPCGatewayNetwork(),
				"scaleway_vpc_public_gateway_dhcp":             resourceScalewayVPCPublicGatewayDHCP(),
				"scaleway_vpc_public_gateway_dhcp_reservation": resourceScalewayVPCPublicGatewayDHCPReservation(),
				"scaleway_vpc_public_gateway_ip":               resourceScalewayVPCPublicGatewayIP(),
				"scaleway_vpc_public_gateway_ip_reverse_dns":   resourceScalewayVPCPublicGatewayIPReverseDNS(),
				"scaleway_vpc_public_gateway_pat_rule":         resourceScalewayVPCPublicGatewayPATRule(),
				"scaleway_vpc_private_network":                 resourceScalewayVPCPrivateNetwork(),
				"scaleway_webhosting":                          resourceScalewayWebhosting(),
			},

			DataSourcesMap: map[string]*schema.Resource{
				"scaleway_account_project":                     dataSourceScalewayAccountProject(),
				"scaleway_account_ssh_key":                     dataSourceScalewayAccountSSHKey(),
				"scaleway_availability_zones":                  DataSourceAvailabilityZones(),
				"scaleway_baremetal_offer":                     dataSourceScalewayBaremetalOffer(),
				"scaleway_baremetal_option":                    dataSourceScalewayBaremetalOption(),
				"scaleway_baremetal_os":                        dataSourceScalewayBaremetalOs(),
				"scaleway_baremetal_server":                    dataSourceScalewayBaremetalServer(),
				"scaleway_billing_invoices":                    dataSourceScalewayBillingInvoices(),
				"scaleway_billing_consumptions":                dataSourceScalewayBillingConsumptions(),
				"scaleway_block_volume":                        dataSourceScalewayBlockVolume(),
				"scaleway_block_snapshot":                      dataSourceScalewayBlockSnapshot(),
				"scaleway_cockpit":                             dataSourceScalewayCockpit(),
				"scaleway_cockpit_plan":                        dataSourceScalewayCockpitPlan(),
				"scaleway_domain_record":                       dataSourceScalewayDomainRecord(),
				"scaleway_domain_zone":                         dataSourceScalewayDomainZone(),
				"scaleway_documentdb_instance":                 dataSourceScalewayDocumentDBInstance(),
				"scaleway_documentdb_database":                 dataSourceScalewayDocumentDBDatabase(),
				"scaleway_documentdb_load_balancer_endpoint":   dataSourceScalewayDocumentDBEndpointLoadBalancer(),
				"scaleway_container_namespace":                 dataSourceScalewayContainerNamespace(),
				"scaleway_container":                           dataSourceScalewayContainer(),
				"scaleway_function":                            dataSourceScalewayFunction(),
				"scaleway_function_namespace":                  dataSourceScalewayFunctionNamespace(),
				"scaleway_iam_application":                     dataSourceScalewayIamApplication(),
				"scaleway_flexible_ip":                         dataSourceScalewayFlexibleIP(),
				"scaleway_flexible_ips":                        dataSourceScalewayFlexibleIPs(),
				"scaleway_iam_group":                           dataSourceScalewayIamGroup(),
				"scaleway_iam_ssh_key":                         dataSourceScalewayIamSSHKey(),
				"scaleway_iam_user":                            dataSourceScalewayIamUser(),
				"scaleway_instance_ip":                         dataSourceScalewayInstanceIP(),
				"scaleway_instance_placement_group":            dataSourceScalewayInstancePlacementGroup(),
				"scaleway_instance_private_nic":                dataSourceScalewayInstancePrivateNIC(),
				"scaleway_instance_security_group":             dataSourceScalewayInstanceSecurityGroup(),
				"scaleway_instance_server":                     dataSourceScalewayInstanceServer(),
				"scaleway_instance_servers":                    dataSourceScalewayInstanceServers(),
				"scaleway_instance_image":                      dataSourceScalewayInstanceImage(),
				"scaleway_instance_volume":                     dataSourceScalewayInstanceVolume(),
				"scaleway_instance_snapshot":                   dataSourceScalewayInstanceSnapshot(),
				"scaleway_iot_hub":                             dataSourceScalewayIotHub(),
				"scaleway_iot_device":                          dataSourceScalewayIotDevice(),
				"scaleway_ipam_ip":                             dataSourceScalewayIPAMIP(),
				"scaleway_ipam_ips":                            dataSourceScalewayIPAMIPs(),
				"scaleway_k8s_cluster":                         dataSourceScalewayK8SCluster(),
				"scaleway_k8s_pool":                            dataSourceScalewayK8SPool(),
				"scaleway_k8s_version":                         dataSourceScalewayK8SVersion(),
				"scaleway_lb":                                  dataSourceScalewayLb(),
				"scaleway_lbs":                                 dataSourceScalewayLbs(),
				"scaleway_lb_acls":                             dataSourceScalewayLbACLs(),
				"scaleway_lb_backend":                          dataSourceScalewayLbBackend(),
				"scaleway_lb_backends":                         dataSourceScalewayLbBackends(),
				"scaleway_lb_certificate":                      dataSourceScalewayLbCertificate(),
				"scaleway_lb_frontend":                         dataSourceScalewayLbFrontend(),
				"scaleway_lb_frontends":                        dataSourceScalewayLbFrontends(),
				"scaleway_lb_ip":                               dataSourceScalewayLbIP(),
				"scaleway_lb_ips":                              dataSourceScalewayLbIPs(),
				"scaleway_lb_route":                            dataSourceScalewayLbRoute(),
				"scaleway_lb_routes":                           dataSourceScalewayLbRoutes(),
				"scaleway_marketplace_image":                   dataSourceScalewayMarketplaceImage(),
				"scaleway_mnq_sqs":                             dataSourceScalewayMNQSQS(),
				"scaleway_object_bucket":                       dataSourceScalewayObjectBucket(),
				"scaleway_object_bucket_policy":                dataSourceScalewayObjectBucketPolicy(),
				"scaleway_rdb_acl":                             dataSourceScalewayRDBACL(),
				"scaleway_rdb_instance":                        dataSourceScalewayRDBInstance(),
				"scaleway_rdb_database":                        dataSourceScalewayRDBDatabase(),
				"scaleway_rdb_database_backup":                 dataSourceScalewayRDBDatabaseBackup(),
				"scaleway_rdb_privilege":                       dataSourceScalewayRDBPrivilege(),
				"scaleway_redis_cluster":                       dataSourceScalewayRedisCluster(),
				"scaleway_registry_namespace":                  dataSourceScalewayRegistryNamespace(),
				"scaleway_tem_domain":                          dataSourceScalewayTemDomain(),
				"scaleway_secret":                              dataSourceScalewaySecret(),
				"scaleway_secret_version":                      dataSourceScalewaySecretVersion(),
				"scaleway_registry_image":                      dataSourceScalewayRegistryImage(),
				"scaleway_vpc":                                 dataSourceScalewayVPC(),
				"scaleway_vpcs":                                dataSourceScalewayVPCs(),
				"scaleway_vpc_public_gateway":                  dataSourceScalewayVPCPublicGateway(),
				"scaleway_vpc_gateway_network":                 dataSourceScalewayVPCGatewayNetwork(),
				"scaleway_vpc_public_gateway_dhcp":             dataSourceScalewayVPCPublicGatewayDHCP(),
				"scaleway_vpc_public_gateway_dhcp_reservation": dataSourceScalewayVPCPublicGatewayDHCPReservation(),
				"scaleway_vpc_public_gateway_ip":               dataSourceScalewayVPCPublicGatewayIP(),
				"scaleway_vpc_private_network":                 dataSourceScalewayVPCPrivateNetwork(),
				"scaleway_vpc_public_gateway_pat_rule":         dataSourceScalewayVPCPublicGatewayPATRule(),
				"scaleway_webhosting":                          dataSourceScalewayWebhosting(),
				"scaleway_webhosting_offer":                    dataSourceScalewayWebhostingOffer(),
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
