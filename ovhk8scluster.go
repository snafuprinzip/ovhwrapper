package ovhwrapper

import (
	"fmt"
	"github.com/ovh/go-ovh/ovh"
	"log"
	"time"
)

// K8SCluster represents a Kubernetes cluster.
//
// Fields:
// - ID: the unique identifier of the cluster.
// - Region: the region where the cluster is located.
// - Name: the name of the cluster.
// - URL: the URL of the cluster.
// - NodesURL: the URL of the cluster's nodes.
// - Version: the version of Kubernetes used by the cluster.
// - NextUpgradeVersions: the list of versions available for upgrade.
// - KubeProxyMode: the mode of the kube-proxy component.
// - Customization: the customization settings of the cluster.
// - Status: the status of the cluster.
// - UpdatePolicy: the policy for cluster updates.
// - IsUpToDate: indicates if the cluster is up to date.
// - ControlPlaneIsUpToDate: indicates if the control plane is up to date.
// - PrivateNetworkID: the ID of the private network used by the cluster.
// - NodesSubnetID: the ID of the subnet used by the cluster's nodes.
// - PrivateNetworkConfiguration: the configuration of the private network.
// - CreatedAt: the timestamp of when the cluster was created.
// - UpdatedAt: the timestamp of when the cluster was last updated.
// - AuditLogsSubscribed: indicates if the cluster is subscribed to audit logs.
// - Nodes: the list of nodes in the cluster (optional).
// - Nodepools: the list of nodepools in the cluster (optional).
// - EtcdUsage: the usage statistics of the etcd component (optional).
type K8SCluster struct {
	ID                          string                      `json:"id"`
	Region                      string                      `json:"region"`
	Name                        string                      `json:"name"`
	URL                         string                      `json:"url"`
	NodesURL                    string                      `json:"nodesUrl"`
	Version                     string                      `json:"version"`
	NextUpgradeVersions         []string                    `json:"nextUpgradeVersions"`
	KubeProxyMode               string                      `json:"kubeProxyMode"`
	Customization               Customization               `json:"customization"`
	Status                      string                      `json:"status"`
	UpdatePolicy                string                      `json:"updatePolicy"`
	IsUpToDate                  bool                        `json:"isUpToDate"`
	ControlPlaneIsUpToDate      bool                        `json:"controlPlaneIsUpToDate"`
	PrivateNetworkID            string                      `json:"privateNetworkId"`
	NodesSubnetID               string                      `json:"nodesSubnetId"`
	PrivateNetworkConfiguration PrivateNetworkConfiguration `json:"privateNetworkConfiguration"`
	CreatedAt                   time.Time                   `json:"createdAt"`
	UpdatedAt                   time.Time                   `json:"updatedAt"`
	AuditLogsSubscribed         bool                        `json:"auditLogsSubscribed"`
	Nodes                       K8sNodes                    `json:"nodes,omitempty"`
	Nodepools                   K8SNodepools                `json:"nodepools,omitempty"`
	EtcdUsage                   K8SEtcd                     `json:",omitempty"`
}

type AdmissionPlugins struct {
	Enabled  []string `json:"enabled"`
	Disabled []any    `json:"disabled"`
}

type APIServer struct {
	AdmissionPlugins AdmissionPlugins `json:"admissionPlugins"`
}

type Customization struct {
	APIServer APIServer `json:"apiServer"`
}

type PrivateNetworkConfiguration struct {
	PrivateNetworkRoutingAsDefault bool   `json:"privateNetworkRoutingAsDefault"`
	DefaultVrackGateway            string `json:"defaultVrackGateway"`
}

func (cluster K8SCluster) Details() string {
	return fmt.Sprintf("Cluster ID: %s\n Region: %s\n Name: %s\n URL: %s\n Nodes URL: %s\n Version: %s\n "+
		"Next Upgrade Versions: %v\n Kube Proxy Mode: %s\n Customization: %s\n Status: %s\n Update Policy: %s\n "+
		"Is Up To Date: %v\n Control Plane Is Up To Date: %v\n Private Network ID: %s\n Nodes Subnet ID: %s\n "+
		"Private Network Configuration: %s\n Created At: %s\n Updated At: %s\n Audit Logs Subscribed: %v\n",
		cluster.ID, cluster.Region, cluster.Name, cluster.URL, cluster.NodesURL, cluster.Version,
		cluster.NextUpgradeVersions, cluster.KubeProxyMode, cluster.Customization.Details(), cluster.Status,
		cluster.UpdatePolicy, cluster.IsUpToDate, cluster.ControlPlaneIsUpToDate, cluster.PrivateNetworkID,
		cluster.NodesSubnetID, cluster.PrivateNetworkConfiguration.Details(), cluster.CreatedAt, cluster.UpdatedAt,
		cluster.AuditLogsSubscribed)
}

func (ap AdmissionPlugins) Details() string {
	return fmt.Sprintf("Enabled: %v, Disabled: %v", ap.Enabled, ap.Disabled)
}

func (as APIServer) Details() string {
	return fmt.Sprintf("Admission Plugins: %s\n",
		as.AdmissionPlugins.Details())
}

func (c Customization) Details() string {
	return fmt.Sprintf("API Server Admission Plugins Enabled: %v, Disabled: %v",
		c.APIServer.AdmissionPlugins.Enabled, c.APIServer.AdmissionPlugins.Disabled)
}

func (pnc PrivateNetworkConfiguration) Details() string {
	return fmt.Sprintf("Private Network Routing as Default: %v, Default Vrack Gateway: %s",
		pnc.PrivateNetworkRoutingAsDefault, pnc.DefaultVrackGateway)
}

func GetK8SClusterIDs(client *ovh.Client, service string) ([]string, error) {
	var clusterlist []string

	if err := client.Get("/cloud/project/"+service+"/kube", &clusterlist); err != nil {
		fmt.Printf("Error getting k8s cluster list: %q\n", err)
		return clusterlist, err
	}

	return clusterlist, nil
}

func GetK8SCluster(client *ovh.Client, service, clusterid string) *K8SCluster {
	var cluster K8SCluster
	if err := client.Get("/cloud/project/"+service+"/kube/"+clusterid, &cluster); err != nil {
		fmt.Printf("Error getting k8s cluster details for %s in sl %s: %q\n", clusterid, service, err)
		return nil
	}
	return &cluster
}

func GetK8SClusterDetails(client *ovh.Client, cluster *K8SCluster, serviceid, clusterid string) (*K8SCluster, error) {
	var err error

	cluster.EtcdUsage, err = GetK8SEtcd(client, serviceid, clusterid)
	if err != nil {
		log.Printf("Error getting etcd usage of cluster %s in SL %s: %v\n", serviceid, clusterid, err)
		return nil, err
	}

	cluster.Nodepools, err = GetK8SNodepools(client, serviceid, clusterid)
	if err != nil {
		log.Printf("Error getting nodepools of cluster %s in SL %s: %v\n", serviceid, clusterid, err)
		return nil, err
	}

	cluster.Nodes, err = GetK8SNodes(client, serviceid, clusterid)
	if err != nil {
		log.Printf("Error getting nodes of cluster %s in SL %s: %v\n", serviceid, clusterid, err)
		return nil, err
	}

	return cluster, nil
}

func UpdateK8SCluster(client *ovh.Client, service, clusterid string, latest, force bool) error {
	type UpdatePostParams struct {
		Strategy string `json:"strategy"`
		Force    bool   `json:"force"`
	}
	var params *UpdatePostParams

	if latest {
		params = &UpdatePostParams{Strategy: "LATEST_PATCH"}
	} else {
		params = &UpdatePostParams{Strategy: "NEXT_MINOR"}
	}

	if params != nil {
		params.Force = force
	}

	if err := client.Post("/cloud/project/"+service+"/kube/"+clusterid+"/update", &params, nil); err != nil {
		fmt.Printf("Error updating cluster %s in SL %s: %q\n", service, clusterid, err)
		return err
	}

	return nil
}

func (cluster K8SCluster) StatusMsg() string {
	return fmt.Sprintf("  Cluster: %s\t[%s]\n  Version: %s (available: %v)\n  etcd: %d%% (%d of %d)",
		cluster.Name, cluster.Status, cluster.Version, cluster.NextUpgradeVersions,
		cluster.EtcdUsage.Usage*100/cluster.EtcdUsage.Quota, cluster.EtcdUsage.Usage, cluster.EtcdUsage.Quota)
}
