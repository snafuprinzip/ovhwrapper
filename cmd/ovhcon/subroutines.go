package main

import (
	"fmt"
	"github.com/ovh/go-ovh/ovh"
	"log"
	"ovhwrapper"
)

// credentials returns information about the reader and writer accounts in different formats (yaml, json or text)
func credentials(reader, writer *ovh.Client, format string) {
	rcred, err := ovhwrapper.GetCredential(reader)
	if err != nil {
		log.Printf("Error getting reader credentials: %q\n", err)
	}
	wcred, err := ovhwrapper.GetCredential(writer)
	if err != nil {
		log.Printf("Error getting writer credentials: %q\n", err)
	}

	switch format {
	case "yaml":
		fmt.Println("Reader Token:\n-------------")
		fmt.Println(ovhwrapper.ToYaml(rcred))

		fmt.Println("Writer Token:\n-------------")
		fmt.Println(ovhwrapper.ToYaml(wcred))
	case "json":
		fmt.Println("Reader Token:\n-------------")
		fmt.Println(ovhwrapper.ToJSON(rcred))

		fmt.Println("Writer Token:\n-------------")
		fmt.Println(ovhwrapper.ToJSON(wcred))
	case "text":
		fallthrough
	default:
		fmt.Println("Reader Token:\n-------------")
		ovhwrapper.PrintCredential(&rcred)

		fmt.Println("Writer Token:\n-------------")
		ovhwrapper.PrintCredential(&wcred)
	}
}

// list lists servicelines and their clusters (when -a is set),
// servicelines (when -s is not set) or clusters (when -s is set)
func list(client *ovh.Client, all bool, serviceid string) {
	var err error
	var sls []ovhwrapper.ServiceLine

	//fmt.Println(all, serviceid)

	// Get flat Serviceline info
	slids := ovhwrapper.GetServicelines(client) // list of sl ids
	for _, slid := range slids {
		sl := ovhwrapper.ServiceLine{ID: slid}
		sl.SLDetails, err = ovhwrapper.GetServicelineDetails(client, slid)
		if err != nil {
			log.Fatalf("Failed to get service lines: %v", err)
		}
		sls = append(sls, sl)
	}

	if all { // list all servicelines and their clusters
		for idx := range sls {
			clusterids, err := ovhwrapper.GetK8SClusterIDs(client, sls[idx].ID)
			if err != nil {
				log.Fatalf("Failed to get cluster IDs: %v", err)
			}
			var clusterlist []ovhwrapper.K8SCluster
			for _, clusterid := range clusterids {
				cluster := ovhwrapper.GetK8SCluster(client, sls[idx].ID, clusterid)
				//fmt.Printf("sl %-2d: %s\t%v\n", idx, clusterid, cluster)
				if cluster != nil {
					clusterlist = append(clusterlist, *cluster)
				}
			}
			sls[idx].Cluster = clusterlist
		}

		// output
		for _, sl := range sls {
			fmt.Printf("%-25s (%s) \t %s \n", ovhwrapper.ShortenName(sl.SLDetails.Description), sl.ID, sl.SLDetails.Description)
			for _, cluster := range sl.Cluster {
				fmt.Printf("  %-25s (%s) \t %s \n", ovhwrapper.ShortenName(cluster.Name), cluster.ID, cluster.Name)
			}
			fmt.Println()
		}
	} else if serviceid != "" { // show clusters for a specific serviceline
		var sl ovhwrapper.ServiceLine

		for _, s := range sls {
			if MatchItem(s, serviceid) {
				sl = s
				break
			}
		}

		if sl.ID == "" {
			log.Printf("Service ID not found: %s", serviceid)
			return
		}

		clusterids, err := ovhwrapper.GetK8SClusterIDs(client, sl.ID)
		if err != nil {
			log.Fatalf("Failed to get cluster IDs: %v", err)
		}
		var clusterlist []ovhwrapper.K8SCluster
		for _, clusterid := range clusterids {
			cluster := ovhwrapper.GetK8SCluster(client, sl.ID, clusterid)
			//fmt.Printf("sl %-2d: %s\t%v\n", idx, clusterid, cluster)
			if cluster != nil {
				clusterlist = append(clusterlist, *cluster)
			}
		}
		sl.Cluster = clusterlist

		fmt.Printf("%-25s (%s) \t %s \n", ovhwrapper.ShortenName(sl.SLDetails.Description), sl.ID, sl.SLDetails.Description)
		for _, cluster := range sl.Cluster {
			fmt.Printf("  %-25s (%s) \t %s \n", ovhwrapper.ShortenName(cluster.Name), cluster.ID, cluster.Name)
		}
		fmt.Println()
	} else { // serviceid == ""
		for _, sl := range sls {
			fmt.Printf("%-25s (%s) \t %s \n", ovhwrapper.ShortenName(sl.SLDetails.Description), sl.ID, sl.SLDetails.Description)
		}
	}
}

func DownloadKubeconfig(reader, writer *ovh.Client, all bool, serviceid, clusterid, output string) {
	var sls []ovhwrapper.ServiceLine
	var err error

	// Get flat Serviceline and cluster info
	slids := ovhwrapper.GetServicelines(reader) // list of sl ids
	for _, slid := range slids {
		sl := ovhwrapper.ServiceLine{ID: slid}
		sl.SLDetails, err = ovhwrapper.GetServicelineDetails(reader, slid)
		if err != nil {
			log.Fatalf("Failed to get service lines: %v", err)
		}
		clusterids, err := ovhwrapper.GetK8SClusterIDs(reader, sl.ID)
		if err != nil {
			log.Fatalf("Failed to get cluster IDs: %v", err)
		}
		var clusterlist []ovhwrapper.K8SCluster
		for _, clusterid := range clusterids {
			cluster := ovhwrapper.GetK8SCluster(reader, sl.ID, clusterid)
			if cluster != nil {
				clusterlist = append(clusterlist, *cluster)
			}
		}
		sl.Cluster = clusterlist
		sls = append(sls, sl)
	}

	if all {
		for _, sl := range sls {
			for _, cluster := range sl.Cluster {
				kc, err := ovhwrapper.GetKubeconfig(writer, sl.ID, cluster.ID)
				if err != nil {
					log.Printf("Failed to get kubeconfig: %v", err)
					continue
				}
				switch output {
				case "global":
				case "certs":
				case "file":
					fallthrough
				default:
					err = ovhwrapper.SaveYaml(kc, sl.SLDetails.Description+"_"+cluster.Name+".yaml")
					if err != nil {
						log.Printf("Saving Kubeconfig failed: %v", err)
					}
				}
			}
			return
		}
	} else if serviceid != "" && clusterid != "" {
		kc, err := ovhwrapper.GetKubeconfig(writer, serviceid, clusterid)
		if err != nil {
			log.Printf("Failed to get kubeconfig: %v", err)
			return
		}
		for _, sl := range sls {
			if MatchItem(sl, serviceid) {
				for _, cluster := range sl.Cluster {
					if MatchItem(cluster, clusterid) {
						switch output {
						case "global":
						case "certs":
						case "file":
							fallthrough
						default:
							err = ovhwrapper.SaveYaml(kc, sl.SLDetails.Description+"_"+cluster.Name+".yaml")
							if err != nil {
								log.Printf("Saving Kubeconfig failed: %v", err)
							}
						}
					}
				}

			}
		}
	} else {
		log.Printf("no service id/name or cluster id/name given\n")
	}
}

// status shows the current status of servicelines and their clusters (when -a is set),
// or a specific cluster (when -s and -c is set)
func status(client *ovh.Client, all bool, serviceline, cluster string) {
	var sls []ovhwrapper.ServiceLine
	var err error

	if all {
		services := ovhwrapper.GetServicelines(client)
		for _, service := range services {
			sl := CollectServiceline(client, service)
			sls = append(sls, *sl)
		}
		flavors, err := ovhwrapper.GetK8SFlavors(client, sls[0].ID, sls[0].Cluster[0].ID)
		if err != nil {
			log.Printf("Error getting available flavors: %q\n", err)
		}

		for _, sl := range sls {
			fmt.Println(sl.StatusMsg())
			for _, cl := range sl.Cluster {
				fmt.Println(cl.StatusMsg())
				for _, n := range cl.Nodes {
					f := flavors[n.Flavor]
					fmt.Println(n.StatusMsg(f))
				}
				fmt.Println()
			}
			fmt.Println("-------------------\n")
		}
		return
	}

	if serviceline != "" { // list serviceline and it's clusters
		services := ovhwrapper.GetServicelines(client)
		for _, service := range services {
			sl := GetServiceline(client, service)
			if MatchItem(*sl, serviceline) {
				sl.SLDetails, err = ovhwrapper.GetServicelineDetails(client, sl.ID)
				if err != nil {
					break
				}

				clusterids, err := ovhwrapper.GetK8SClusterIDs(client, sl.ID)
				if err != nil {
					log.Fatalf("Failed to get cluster IDs: %v", err)
				}

				var clusterlist []ovhwrapper.K8SCluster
				for _, clid := range clusterids {
					if cluster != "" { // cluster id is given on the command line
						cl := ovhwrapper.GetK8SCluster(client, sl.ID, clid)
						if MatchItem(*cl, cluster) {
							cl, err = ovhwrapper.GetK8SClusterDetails(client, cl, sl.ID, cl.ID)
							if err != nil && cl != nil {
								clusterlist = append(clusterlist, *cl)
							}
						}
					} else { // all clusters
						cl := CollectCluster(client, sl.ID, clid)
						if cl != nil {
							clusterlist = append(clusterlist, *cl)
						}
					}
				}
				sl.Cluster = clusterlist

				sls = append(sls, *sl)
			}
		}

		if len(sls) == 0 {
			log.Printf("No servicelines found for the given identifier: %s\n", serviceline)
			return
		}

		if len(sls[0.].Cluster) == 0 {
			log.Printf("No clusters found for the given identifier: %s\n", cluster)
			return
		}

		flavors, err := ovhwrapper.GetK8SFlavors(client, sls[0].ID, sls[0].Cluster[0].ID)
		if err != nil {
			log.Printf("Error getting available flavors: %q\n", err)
			return
		}

		for _, sl := range sls {
			if MatchItem(sl, serviceline) {
				fmt.Println(sl.StatusMsg())
				for _, cl := range sl.Cluster {
					if cluster == "" || MatchItem(cl, cluster) {
						fmt.Println(cl.StatusMsg())
						for _, n := range cl.Nodes {
							f := flavors[n.Flavor]
							fmt.Println(n.StatusMsg(f))
						}
						fmt.Println()
					}
				}
			}
		}
		return
	}
}
