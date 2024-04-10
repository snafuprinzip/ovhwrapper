package ovhwrapper

import (
	"fmt"
	"github.com/ovh/go-ovh/ovh"
	"time"
)

// K8SNode represents a node in a Kubernetes cluster.
// Fields:
// - Id: unique identifier of the node.
// - ProjectId: identifier of the project that the node belongs to.
// - InstanceId: identifier of the instance associated with the node.
// - NodePoolId: identifier of the node pool that the node belongs to.
// - Name: name of the node.
// - Flavor: flavor of the node.
// - Status: current status of the node.
// - IsUpToDate: indicates whether the node is up to date or not.
// - Version: version of Kubernetes running on the node.
// - CreatedAt: timestamp of when the node was created.
// - UpdatedAt: timestamp of when the node was last updated.
// - DeployedAt: timestamp of when the node was deployed.
type K8SNode struct {
	Id         string    `json:"id"`
	ProjectId  string    `json:"projectId"`
	InstanceId string    `json:"instanceId"`
	NodePoolId string    `json:"nodePoolId"`
	Name       string    `json:"name"`
	Flavor     string    `json:"flavor"`
	Status     string    `json:"status"`
	IsUpToDate bool      `json:"isUpToDate"`
	Version    string    `json:"version"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	DeployedAt time.Time `json:"deployedAt"`
}

type K8SNodepool struct {
	Id             string    `json:"id"`
	ProjectId      string    `json:"projectId"`
	Name           string    `json:"name"`
	Flavor         string    `json:"flavor"`
	Status         string    `json:"status"`
	SizeStatus     string    `json:"sizeStatus"`
	Autoscale      bool      `json:"autoscale"`
	MonthlyBilled  bool      `json:"monthlyBilled"`
	AntiAffinity   bool      `json:"antiAffinity"`
	DesiredNodes   int       `json:"desiredNodes"`
	MinNodes       int       `json:"minNodes"`
	MaxNodes       int       `json:"maxNodes"`
	CurrentNodes   int       `json:"currentNodes"`
	AvailableNodes int       `json:"availableNodes"`
	UpToDateNodes  int       `json:"upToDateNodes"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	Autoscaling    struct {
		ScaleDownUtilizationThreshold float64 `json:"scaleDownUtilizationThreshold"`
		ScaleDownUnneededTimeSeconds  int     `json:"scaleDownUnneededTimeSeconds"`
		ScaleDownUnreadyTimeSeconds   int     `json:"scaleDownUnreadyTimeSeconds"`
	} `json:"autoscaling"`
	Template struct {
		Metadata struct {
			Labels struct {
			} `json:"labels"`
			Annotations struct {
			} `json:"annotations"`
			Finalizers []interface{} `json:"finalizers"`
		} `json:"metadata"`
		Spec struct {
			Unschedulable bool          `json:"unschedulable"`
			Taints        []interface{} `json:"taints"`
		} `json:"spec"`
	} `json:"template"`
}

// K8sNodes is a type that represents a list of K8SNode objects. It is used to store information about Kubernetes nodes.
// K8SNode is a struct that represents a Kubernetes node. It contains various properties such as Id, ProjectId,
// InstanceId, NodePoolId, Name, Flavor, Status, IsUpToDate, Version, Created
type K8sNodes []K8SNode

// K8SNodepools represents a collection of K8SNodepool objects.
// Each K8SNodepool contains various details about a node pool in a Kubernetes cluster, such as its ID, project ID,
// name, flavor, status, size status, autoscaling configuration, and
type K8SNodepools []K8SNodepool

// GetK8SNodes retrieves the list of Kubernetes nodes in a given service and cluster ID.
// It takes in an OVH client, the service name, and the cluster ID as parameters.
// It returns a K8sNodes slice representing the list of nodes and an error if any occurred.
//
// The K8sNodes slice is a collection of K8SNode structs. Each K8SNode struct contains information
// about a node such as its ID, project ID, instance ID, node pool ID, name, flavor, status,
// update status, version, creation timestamp, update timestamp, and deployment timestamp.
func GetK8SNodes(client *ovh.Client, service, clusterid string) (K8sNodes, error) {
	var nodelist K8sNodes
	//	nodelist:=  make(K8sNodes, 3)
	if err := client.Get("/cloud/project/"+service+"/kube/"+clusterid+"/node", &nodelist); err != nil {
		fmt.Printf("Error getting k8s node list: %q\n", err)
		return nodelist, err
	}

	return nodelist, nil
}

// GetK8SNode retrieves the details of a specific Kubernetes node in a given service and cluster ID.
// It takes in an OVH client, the service name, the cluster ID, and the node ID as parameters.
// It returns a K8SNode struct representing the node and an error if any occurred.
// The K8SNode struct contains information about the node such as its ID, project ID, instance ID,
// node pool ID, name, flavor, status, update status, version, creation timestamp, update timestamp, and deployment timestamp.
func GetK8SNode(client *ovh.Client, service, clusterid, nodeid string) (K8SNode, error) {
	var node K8SNode
	//	nodelist:=  make(K8sNodes, 3)
	if err := client.Get("/cloud/project/"+service+"/kube/"+clusterid+"/node/"+nodeid, &node); err != nil {
		fmt.Printf("Error getting k8s node %s: %q\n", nodeid, err)
		return node, err
	}

	return node, nil
}

// GetK8SNodepools retrieves the list of Kubernetes node pools for a given service and cluster ID.
// It takes in an OVH client, the service name, and the cluster ID as parameters.
// It returns a K8SNodepools struct and an error.
func GetK8SNodepools(client *ovh.Client, service, clusterid string) (K8SNodepools, error) {
	var nodepoollist K8SNodepools
	//	nodelist:=  make(K8sNodes, 3)
	if err := client.Get("/cloud/project/"+service+"/kube/"+clusterid+"/nodepool", &nodepoollist); err != nil {
		fmt.Printf("Error getting k8s nodepool list: %q\n", err)
		return nodepoollist, err
	}

	return nodepoollist, nil
}

// GetK8SNodepool retrieves information about a specific Kubernetes node pool for a given service, cluster, and pool ID.
// It takes in an OVH client, the service name, the cluster ID, and the pool ID as parameters.
// It returns a K8SNodepool struct and an error. If there was an error retrieving the node pool information,
// the returned error will contain a description of the problem.
// Example usage:
//
//	nodepool, err := GetK8SNodepool(client, "my-service", "my-cluster", "my-pool")
//	if err != nil {
//	    fmt.Printf("Error getting k8s node pool: %s\n", err.Error())
//	    return
//	}
//	fmt.Printf("Node Pool ID: %s\n", nodepool.Id)
func GetK8SNodepool(client *ovh.Client, service, clusterid, poolid string) (K8SNodepool, error) {
	var nodepool K8SNodepool
	if err := client.Get("/cloud/project/"+service+"/kube/"+clusterid+"/nodepool/"+poolid, &nodepool); err != nil {
		fmt.Printf("Error getting k8s nodepool %s: %q\n", poolid, err)
		return nodepool, err
	}

	return nodepool, nil
}
