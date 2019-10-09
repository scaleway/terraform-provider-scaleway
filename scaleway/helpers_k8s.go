package scaleway

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	k8s "github.com/scaleway/scaleway-sdk-go/api/k8s/v1beta3"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type KubeconfigStruct struct {
	ApiVersion string `yaml:"apiVersion"`
	Clusters   []struct {
		Name    string `yaml:"name"`
		Cluster struct {
			CertificateAuthorityData string `yaml:"certificate-authority-data"`
			Server                   string `yaml:"server"`
		} `yaml:"cluster"`
	} `yaml:"clusters"`
	Contexts []struct {
		Name    string `yaml:"name"`
		Context struct {
			Cluster string `yaml:"cluster"`
			User    string `yaml:"user"`
		} `yaml:"context"`
	} `yaml:"contexts"`
	Kind  string `yaml:"kind"`
	Users []struct {
		Name string `yaml:"name"`
		User struct {
			Token string `yaml:"token"`
		} `yaml:"user"`
	} `yaml:"users"`
}

const (
	K8SClusterWaitForReadyTimeout   = 10 * time.Minute
	K8SClusterWaitForDeletedTimeout = 10 * time.Minute
)

func getK8SAPIWithRegion(d *schema.ResourceData, m interface{}) (*k8s.API, scw.Region, error) {
	meta := m.(*Meta)
	k8sAPI := k8s.NewAPI(meta.scwClient)

	region, err := getRegion(d, meta)

	return k8sAPI, region, err
}

func getK8SAPIWithRegionAndID(m interface{}, id string) (*k8s.API, scw.Region, string, error) {
	meta := m.(*Meta)
	k8sAPI := k8s.NewAPI(meta.scwClient)

	region, ID, err := parseRegionalID(id)
	return k8sAPI, region, ID, err
}

func waitK8SClusterReady(k8sAPI *k8s.API, region scw.Region, clusterID string) error {
	cluster, err := k8sAPI.WaitForCluster(&k8s.WaitForClusterRequest{
		ClusterID: clusterID,
		Region:    region,
		Timeout:   K8SClusterWaitForReadyTimeout,
	})
	if err != nil {
		return err
	}

	if cluster.Status == k8s.ClusterStatusReady {
		return nil
	}
	return fmt.Errorf("Cluster %s has state %s, wants %s", clusterID, cluster.Status.String(), k8s.ClusterStatusReady.String())
}

func waitK8SClusterDeleted(k8sAPI *k8s.API, region scw.Region, clusterID string) error {
	cluster, err := k8sAPI.WaitForCluster(&k8s.WaitForClusterRequest{
		ClusterID: clusterID,
		Region:    region,
		Timeout:   K8SClusterWaitForDeletedTimeout,
	})
	if err != nil {
		if is404Error(err) {
			return nil
		}
		return err
	}

	return fmt.Errorf("Cluster %s has state %s, wants %s", clusterID, cluster.Status.String(), k8s.ClusterStatusDeleted.String())
}

func clusterAutoscalerConfigFlatten(cluster *k8s.Cluster) map[string]interface{} {
	autoscalerConfig := map[string]interface{}{}
	autoscalerConfig["disable_scale_down"] = cluster.AutoscalerConfig.ScaleDownDisabled
	autoscalerConfig["scale_down_delay_after_add"] = cluster.AutoscalerConfig.ScaleDownDelayAfterAdd
	autoscalerConfig["estimator"] = cluster.AutoscalerConfig.Estimator
	autoscalerConfig["expander"] = cluster.AutoscalerConfig.Expander
	autoscalerConfig["ignore_daemonsets_utilization"] = cluster.AutoscalerConfig.IgnoreDaemonsetsUtilization
	autoscalerConfig["balance_similar_node_groups"] = cluster.AutoscalerConfig.BalanceSimilarNodeGroups
	autoscalerConfig["expendable_pods_priority_cutoff"] = cluster.AutoscalerConfig.ExpendablePodsPriorityCutoff

	return autoscalerConfig
}
