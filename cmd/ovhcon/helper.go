package main

import (
	"github.com/ovh/go-ovh/ovh"
	"github.com/snafuprinzip/ovhwrapper"
	"log"
)

// CollectInformation collects the information of all service lines, including their clusters down to the nodes.
func CollectInformation(client *ovh.Client) []ovhwrapper.ServiceLine {
	var servicelines []ovhwrapper.ServiceLine

	services := ovhwrapper.GetServicelines(client)
	for _, service := range services {
		serviceline := CollectServiceline(client, service)
		servicelines = append(servicelines, *serviceline)
	}

	return servicelines
}

// GetServiceline asks the API for 'shallow' information about a specific serviceline, excluding the clusters.
func GetServiceline(client *ovh.Client, serviceid string) *ovhwrapper.ServiceLine {
	servicedetails, err := ovhwrapper.GetServicelineDetails(client, serviceid)
	if err != nil {
		log.Fatalf("Failed to get serviceline details: %v", err)
	}
	serviceline := ovhwrapper.ServiceLine{
		ID:        serviceid,
		SLDetails: servicedetails,
		Cluster:   []ovhwrapper.K8SCluster{},
	}
	return &serviceline
}

// CollectServiceline collects information about a serviceline, including its clusters.
func CollectServiceline(client *ovh.Client, serviceid string) *ovhwrapper.ServiceLine {
	servicedetails, err := ovhwrapper.GetServicelineDetails(client, serviceid)
	if err != nil {
		log.Fatalf("Failed to get serviceline details: %v", err)
	}
	clusterids, err := ovhwrapper.GetK8SClusterIDs(client, serviceid)
	if err != nil {
		log.Fatalf("Failed to get cluster IDs: %v", err)
	}
	var clusterlist []ovhwrapper.K8SCluster
	for _, clusterid := range clusterids {
		cluster := CollectCluster(client, serviceid, clusterid)
		if cluster != nil {
			clusterlist = append(clusterlist, *cluster)
		}
	}
	serviceline := ovhwrapper.ServiceLine{
		ID:        serviceid,
		SLDetails: servicedetails,
		Cluster:   clusterlist,
	}

	return &serviceline
}

// GetCluster asks the API for 'shallow' information about a specific cluster, excluding nested information
// like etcd usage, nodes or nodepools
func GetCluster(client *ovh.Client, serviceid, clusterid string) *ovhwrapper.K8SCluster {
	return ovhwrapper.GetK8SCluster(client, serviceid, clusterid)
}

// CollectCluster returns information about a Cluster, including its etcd usage, nodepools and nodes.
func CollectCluster(client *ovh.Client, serviceid, clusterid string) *ovhwrapper.K8SCluster {
	cluster := ovhwrapper.GetK8SCluster(client, serviceid, clusterid)
	var err error

	cluster, err = ovhwrapper.GetK8SClusterDetails(client, cluster, serviceid, clusterid)
	if err != nil {
		log.Printf("Failed to get cluster details: %v", err)
	}
	return cluster
}

// MatchItem will check if the id or the (abbreviated) name matches with the identifier and returns true or false
func MatchItem[T ovhwrapper.ServiceLine | ovhwrapper.K8SCluster](object T, identifier string) bool {
	match := false
	switch object := any(object).(type) { // lazy hack, as generic functions implement specific types and not interfaces, so we cast to any to check its type
	case ovhwrapper.ServiceLine:
		if object.ID == identifier || object.SLDetails.Description == identifier || ovhwrapper.ShortenName(object.SLDetails.Description) == identifier {
			match = true
		}
	case ovhwrapper.K8SCluster:
		if object.ID == identifier || object.Name == identifier || ovhwrapper.ShortenName(object.Name) == identifier {
			match = true
		}
	}
	return match
}
