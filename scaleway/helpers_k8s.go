package scaleway

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	k8s "github.com/scaleway/scaleway-sdk-go/api/k8s/v1beta4"
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
	K8SPoolWaitForReadyTimeout      = 10 * time.Minute
)

func k8sAPIWithRegion(d *schema.ResourceData, m interface{}) (*k8s.API, scw.Region, error) {
	meta := m.(*Meta)
	k8sAPI := k8s.NewAPI(meta.scwClient)

	region, err := extractRegion(d, meta)

	return k8sAPI, region, err
}

func k8sAPIWithRegionAndID(m interface{}, id string) (*k8s.API, scw.Region, string, error) {
	meta := m.(*Meta)
	k8sAPI := k8s.NewAPI(meta.scwClient)

	region, ID, err := parseRegionalID(id)
	return k8sAPI, region, ID, err
}

func waitK8SClusterReady(k8sAPI *k8s.API, region scw.Region, clusterID string) error {
	cluster, err := k8sAPI.WaitForCluster(&k8s.WaitForClusterRequest{
		ClusterID: clusterID,
		Region:    region,
		Timeout:   scw.TimeDurationPtr(K8SClusterWaitForReadyTimeout),
	})
	if err != nil {
		return err
	}

	if cluster.Status == k8s.ClusterStatusReady {
		return nil
	}
	return fmt.Errorf("cluster %s has state %s, wants %s", clusterID, cluster.Status, k8s.ClusterStatusReady)
}

func waitK8SClusterDeleted(k8sAPI *k8s.API, region scw.Region, clusterID string) error {
	cluster, err := k8sAPI.WaitForCluster(&k8s.WaitForClusterRequest{
		ClusterID: clusterID,
		Region:    region,
		Timeout:   scw.TimeDurationPtr(K8SClusterWaitForDeletedTimeout),
	})
	if err != nil {
		if is404Error(err) {
			return nil
		}
		return err
	}

	return fmt.Errorf("cluster %s has state %s, wants %s", clusterID, cluster.Status, k8s.ClusterStatusDeleted)
}

func waitK8SPoolReady(k8sAPI *k8s.API, region scw.Region, poolID string) error {
	pool, err := k8sAPI.WaitForPool(&k8s.WaitForPoolRequest{
		PoolID:  poolID,
		Region:  region,
		Timeout: scw.TimeDurationPtr(K8SPoolWaitForReadyTimeout),
	})

	if err != nil {
		return err
	}

	if pool.Status == k8s.PoolStatusReady {
		return nil
	}
	return fmt.Errorf("pool %s has state %s, wants %s", poolID, pool.Status, k8s.PoolStatusReady)
}

// convert a list of nodes to a list of map
func convertNodes(res *k8s.ListNodesResponse) []map[string]interface{} {
	var result []map[string]interface{}
	for _, node := range res.Nodes {
		n := make(map[string]interface{})
		n["name"] = node.Name
		n["status"] = node.Status.String()
		if node.PublicIPV4 != nil && node.PublicIPV4.String() != "<nil>" {
			n["public_ip"] = node.PublicIPV4.String()
		}
		if node.PublicIPV6 != nil && node.PublicIPV6.String() != "<nil>" {
			n["public_ip_v6"] = node.PublicIPV6.String()
		}
		result = append(result, n)
	}
	return result
}

func getNodes(k8sAPI *k8s.API, pool *k8s.Pool) ([]map[string]interface{}, error) {
	req := &k8s.ListNodesRequest{
		Region:    pool.Region,
		ClusterID: pool.ClusterID,
		PoolID:    &pool.ID,
	}

	nodes, err := k8sAPI.ListNodes(req, scw.WithAllPages())

	if err != nil {
		return nil, err
	}

	return convertNodes(nodes), nil
}

func clusterAutoscalerConfigFlatten(cluster *k8s.Cluster) []map[string]interface{} {
	autoscalerConfig := map[string]interface{}{}
	autoscalerConfig["disable_scale_down"] = cluster.AutoscalerConfig.ScaleDownDisabled
	autoscalerConfig["scale_down_delay_after_add"] = cluster.AutoscalerConfig.ScaleDownDelayAfterAdd
	autoscalerConfig["estimator"] = cluster.AutoscalerConfig.Estimator
	autoscalerConfig["expander"] = cluster.AutoscalerConfig.Expander
	autoscalerConfig["ignore_daemonsets_utilization"] = cluster.AutoscalerConfig.IgnoreDaemonsetsUtilization
	autoscalerConfig["balance_similar_node_groups"] = cluster.AutoscalerConfig.BalanceSimilarNodeGroups
	autoscalerConfig["expendable_pods_priority_cutoff"] = cluster.AutoscalerConfig.ExpendablePodsPriorityCutoff

	return []map[string]interface{}{autoscalerConfig}
}

func clusterAutoUpgradeFlatten(cluster *k8s.Cluster) []map[string]interface{} {
	autoUpgrade := map[string]interface{}{}
	autoUpgrade["enable"] = cluster.AutoUpgrade.Enabled
	autoUpgrade["maintenance_window_start_hour"] = cluster.AutoUpgrade.MaintenanceWindow.StartHour
	autoUpgrade["maintenance_window_day"] = cluster.AutoUpgrade.MaintenanceWindow.Day

	return []map[string]interface{}{autoUpgrade}
}
