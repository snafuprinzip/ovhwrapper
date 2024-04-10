package main

import (
	"fmt"
	"github.com/ovh/go-ovh/ovh"
	"github.com/snafuprinzip/ovhwrapper"
	"log"
	"os"
)

func CollectInformation(client *ovh.Client) {
	var servicelines []ovhwrapper.ServiceLine

	services := ovhwrapper.GetServicelines(client)
	for _, service := range services {
		clusterids, err := ovhwrapper.GetK8SClusterIDs(client, service)
		if err != nil {
			log.Printf("Error getting Cluster IDs for Serviceline %s: %q\n", service, err)
			continue
		}
		//fmt.Printf("k8s Cluster of service %s: %+v\n", service, clusterids)

		servicedetails, err := ovhwrapper.GetServicelineDetails(client, service)
		if err != nil {
			log.Printf("Error getting Service Line details for Serviceline %s: %q\n", service, err)
			continue
		}

		var clusterlist []ovhwrapper.K8SCluster
		for _, clusterid := range clusterids {
			cluster := ovhwrapper.GetK8SCluster(client, service, clusterid)
			//k := K8SCluster{
			//	ID: clustername,
			//}
			cluster.EtcdUsage, err = ovhwrapper.GetK8SEtcd(client, service, clusterid)
			if err != nil {
				log.Printf("Error getting etcd usage of cluster %s in SL %s: %v\n", service, clusterid, err)
				continue
			}

			cluster.Nodepools, err = ovhwrapper.GetK8SNodepools(client, service, clusterid)
			if err != nil {
				log.Printf("Error getting nodepools of cluster %s in SL %s: %v\n", service, clusterid, err)
				continue
			}

			cluster.Nodes, err = ovhwrapper.GetK8SNodes(client, service, clusterid)
			if err != nil {
				log.Printf("Error getting nodes of cluster %s in SL %s: %v\n", service, clusterid, err)
				continue
			}

			clusterlist = append(clusterlist, *cluster)
		}

		serviceline := ovhwrapper.ServiceLine{
			ID:        service,
			SLDetails: servicedetails,
			Cluster:   clusterlist,
		}
		servicelines = append(servicelines, serviceline)
	}

	f, err := os.OpenFile("data/clusterdetails.txt", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error opening file clusterdetails.txt: %q\n", err)
	}
	defer f.Close()

	for _, sl := range servicelines {
		fmt.Printf("SL ID: %s\n", sl.ID)
		fmt.Printf("SL Name: %s\n", sl.SLDetails.Description)
		_, err = f.WriteString(sl.SLDetails.Details())
		if err != nil {
			log.Printf("Error writing details of serviceline %s to file: %q\n", sl.SLDetails.Description, err)
		}

		for _, cluster := range sl.Cluster {
			fmt.Printf("   Cluster ID:     %s\n", cluster.ID)
			fmt.Printf("   Cluster Name:   %s\n", cluster.Name)
			fmt.Printf("   Cluster Region: %s\n", cluster.Region)
			fmt.Printf("   Cluster Status: %s\n", cluster.Status)
			fmt.Println("   ------------------------------")

			_, err = f.WriteString(cluster.Details() + "   -----\n")
			if err != nil {
				log.Printf("Error writing details of cluster %s to file: %q\n", cluster.Name, err)
			}

		}
		fmt.Println("\n\n")
		f.WriteString("------------------------------\n\n")
	}
	//y, err := os.OpenFile("clusterdetails.yaml", os.O_CREATE|os.O_WRONLY, 0644)
	//if err != nil {
	//	log.Fatalf("Error opening file clusterdetails.yaml: %q\n", err)
	//}
	//defer y.Close()
	//
	//y.WriteString(ovhwrapper.ToYaml(servicelines))
	ovhwrapper.SaveYaml(servicelines, "data/clusterdetails.yaml")
}

func GetCluster(client *ovh.Client, serviceid, clusterid string) *ovhwrapper.K8SCluster {
	cluster := ovhwrapper.GetK8SCluster(client, serviceid, clusterid)
	var err error

	cluster.EtcdUsage, err = ovhwrapper.GetK8SEtcd(client, serviceid, clusterid)
	if err != nil {
		log.Printf("Error getting etcd usage of cluster %s in SL %s: %v\n", serviceid, clusterid, err)
		return nil
	}

	cluster.Nodepools, err = ovhwrapper.GetK8SNodepools(client, serviceid, clusterid)
	if err != nil {
		log.Printf("Error getting nodepools of cluster %s in SL %s: %v\n", serviceid, clusterid, err)
		return nil
	}

	cluster.Nodes, err = ovhwrapper.GetK8SNodes(client, serviceid, clusterid)
	if err != nil {
		log.Printf("Error getting nodes of cluster %s in SL %s: %v\n", serviceid, clusterid, err)
		return nil
	}

	return cluster
}

func main() {
	config := ovhwrapper.ReadConfig("ovh.conf")
	config.Default.Endpoint = "ovh-eu"

	client, err := ovhwrapper.CreateClient()
	if err != nil {
		log.Fatalf("Error creating OVH API Client: %q\n", err)
	}

	if config.EU.ConsumerKey == "" {
		// consumer key erzeugen
		consumerkey, err := ovhwrapper.CreateConsumerKey(client)
		if err != nil {
			log.Fatalf("Error: %q\n", err)
		}
		config.EU.ConsumerKey = consumerkey

		config.Save("ovh.conf")

		os.Exit(0)
	}

	//CollectInformation(client)

	//serviceid := "xxx"
	//clusterid := "xxx"

	// Update Cluster
	//var cl *ovhwrapper.K8SCluster = ovhwrapper.GetK8SCluster(client, serviceid, clusterid)

	//if cl.Status == "READY" {
	//	err = ovhwrapper.UpdateK8SCluster(client, serviceid, clusterid, false, false)
	//	if err != nil {
	//		log.Fatalf("Error updating K8S cluster: %q\n", err)
	//	}
	//}

	// reset kubeconfig
	//var kc ovhwrapper.KubeConfig
	//if cl.Status == "READY" {
	//	kc, err = ovhwrapper.ResetKubeconfig(client, serviceid, clusterid)
	//	if err != nil {
	//		log.Fatalf("Error resetting kubeconfig of K8S cluster: %q\n", err)
	//	}
	//}
	//
	//time.Sleep(10 * time.Second)

	// Update Cluster Status
	//for {
	//	client, err := ovhwrapper.CreateClient()
	//	if err != nil {
	//		log.Fatalf("Error creating OVH API Client: %q\n", err)
	//	}
	//	cl = GetCluster(client, serviceid, clusterid)
	//	if cl != nil {
	//		fmt.Printf("Cluster: %s (%s)\nVersion: %s (available: %v)\n\n", cl.Name, cl.Status, cl.Version, cl.NextUpgradeVersions)
	//		for _, n := range cl.Nodes {
	//			fmt.Printf("  Nodename: %s (%-10s)\t-\t%s (up2date: %v)\n", n.Name, n.Status, n.Version, n.IsUpToDate)
	//		}
	//		fmt.Printf("\n\n\n")
	//
	//		if cl.Status == "READY" {
	//			break
	//		}
	//		time.Sleep(60 * time.Second)
	//	}
	//}

	// Get Kubeconfig
	//kc, err := ovhwrapper.GetKubeconfig(client, serviceid, clusterid)
	//if err != nil {
	//	log.Fatalf("Error getting kubeconfig: %q\n", err)
	//}
	//fmt.Printf("Kubeconfig: %+v\n", kc)
	//ovhwrapper.SaveYaml(kc, "data/kubeconfig."+serviceid+"_"+clusterid)

	//fmt.Println(kc)

	cred, err := ovhwrapper.GetCredential(client)
	fmt.Println(ovhwrapper.ToYaml(cred))
}
