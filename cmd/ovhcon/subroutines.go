package main

import (
	"encoding/base64"
	"fmt"
	"github.com/ovh/go-ovh/ovh"
	"github.com/snafuprinzip/ovhwrapper"
	"log"
	"os"
	"path"
	"time"
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

func DownloadKubeconfig(reader, writer *ovh.Client, all bool, serviceid, clusterid, output, outpath string) {
	var sls []ovhwrapper.ServiceLine
	var err error
	globalconfig := ovhwrapper.KubeConfig{
		APIVersion: "v1",
		Kind:       "Config",
	}

	if outpath == "" {
		outpath = "./"
	}

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
		for _, clid := range clusterids {
			cluster := ovhwrapper.GetK8SCluster(reader, slid, clid)
			if cluster != nil {
				clusterlist = append(clusterlist, *cluster)
			}
		}
		sl.Cluster = clusterlist
		sls = append(sls, sl)
	}

	if all {
		for _, sl := range sls {
			fmt.Println("Processing Serviceline: ", sl.SLDetails.Description)
			for _, cl := range sl.Cluster {
				kc, err := ovhwrapper.GetKubeconfig(writer, sl.ID, cl.ID)
				if err != nil {
					log.Printf("Failed to get kubeconfig: %v", err)
					continue
				}
				switch output {
				case "global":
					globalconfig.AddContext(kc)
				case "certs":
					certpath := path.Join(outpath, sl.SLDetails.Description, cl.Name)
					err := os.MkdirAll(certpath, 0700)
					if err != nil {
						log.Printf("Failed to create output directory: %v", err)
						continue
					}

					fmt.Printf("Extracting certificates for %s serviceline's %s cluster to %s...\n",
						sl.SLDetails.Description, cl.Name, certpath)

					ca, err := base64.StdEncoding.DecodeString(kc.Clusters[0].Cluster.CertificateAuthorityData)
					if err != nil {
						log.Fatal("error decoding ca certificate:", err)
					}
					err = os.WriteFile(path.Join(certpath, "ca.crt"), ca, 0600)
					if err != nil {
						fmt.Printf("Failed to write ca.crt output file: %v", err)
					}

					crt, err := base64.StdEncoding.DecodeString(kc.Users[0].User.ClientCertificateData)
					if err != nil {
						log.Fatal("error decoding user client certificate:", err)
					}
					err = os.WriteFile(path.Join(certpath, "client.crt"), crt, 0600)
					if err != nil {
						fmt.Printf("Failed to write client.crt output file: %v", err)
					}

					key, err := base64.StdEncoding.DecodeString(kc.Users[0].User.ClientKeyData)
					if err != nil {
						log.Fatal("error decoding user client private key:", err)
					}
					err = os.WriteFile(path.Join(certpath, "client.key"), key, 0600)
					if err != nil {
						fmt.Printf("Failed to write client.key output file: %v", err)
					}
				case "file":
					fallthrough
				default:
					err := os.MkdirAll(outpath, 0700)
					if err != nil {
						log.Printf("Failed to create output directory: %v", err)
						continue
					}
					fmt.Printf("Saving kubeconfig to %s...\n", path.Join(outpath, sl.SLDetails.Description+"_"+cl.Name+".yaml"))
					err = ovhwrapper.SaveYaml(kc, path.Join(outpath, sl.SLDetails.Description+"_"+cl.Name+".yaml"))
					if err != nil {
						log.Printf("Saving Kubeconfig failed: %v", err)
					}
				}
			}
		}
	} else if serviceid != "" && clusterid != "" {
		for _, sl := range sls {
			if MatchItem(sl, serviceid) {
				for _, cl := range sl.Cluster {
					if MatchItem(cl, clusterid) {
						kc, err := ovhwrapper.GetKubeconfig(writer, sl.ID, cl.ID)
						if err != nil {
							log.Printf("Failed to get kubeconfig: %v", err)
							return
						}
						switch output {
						case "global":
							globalconfig.AddContext(kc)
						case "certs":
							certpath := path.Join(outpath, sl.SLDetails.Description, cl.Name)
							err := os.MkdirAll(certpath, 0700)
							if err != nil {
								log.Printf("Failed to create output directory: %v", err)
								continue
							}

							fmt.Printf("Extracting certificates for %s serviceline's %s cluster to %s...\n",
								sl.SLDetails.Description, cl.Name, certpath)

							ca, err := base64.StdEncoding.DecodeString(kc.Clusters[0].Cluster.CertificateAuthorityData)
							if err != nil {
								log.Fatal("error decoding ca certificate:", err)
							}
							err = os.WriteFile(path.Join(certpath, "ca.crt"), ca, 0600)
							if err != nil {
								fmt.Printf("Failed to write ca.crt output file: %v", err)
							}

							crt, err := base64.StdEncoding.DecodeString(kc.Users[0].User.ClientCertificateData)
							if err != nil {
								log.Fatal("error decoding user client certificate:", err)
							}
							err = os.WriteFile(path.Join(certpath, "client.crt"), crt, 0600)
							if err != nil {
								fmt.Printf("Failed to write client.crt output file: %v", err)
							}

							key, err := base64.StdEncoding.DecodeString(kc.Users[0].User.ClientKeyData)
							if err != nil {
								log.Fatal("error decoding user client private key:", err)
							}
							err = os.WriteFile(path.Join(certpath, "client.key"), key, 0600)
							if err != nil {
								fmt.Printf("Failed to write client.key output file: %v", err)
							}
						case "file":
							fallthrough
						default:
							err = ovhwrapper.SaveYaml(kc, sl.SLDetails.Description+"_"+cl.Name+".yaml")
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

	if output == "global" {
		log.Printf("Saving global kubeconfig to %s...\n", path.Join(outpath, "global.yaml"))
		err = ovhwrapper.SaveYaml(globalconfig, path.Join(outpath, "global.yaml"))
		if err != nil {
			log.Printf("Saving global Kubeconfig failed: %v", err)
		}
	}
}

// status shows the current status of servicelines and their clusters (when -a is set),
// or a specific cluster (when -s and -c is set)
func status(client *ovh.Client, all bool, serviceline, cluster string) {
	var sls []ovhwrapper.ServiceLine
	var cl *ovhwrapper.K8SCluster
	var realslid, realclid string
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
				realslid = sl.ID
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
						cl = ovhwrapper.GetK8SCluster(client, sl.ID, clid)
						if MatchItem(*cl, cluster) {
							realclid = cl.ID
							cl, err = ovhwrapper.GetK8SClusterDetails(client, cl, realslid, realclid)
							if err == nil && cl != nil {
								clusterlist = append(clusterlist, *cl)
								break
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

		if len(sls[0].Cluster) == 0 {
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

func UpdateCluster(reader, writer *ovh.Client, config ovhwrapper.Configuration, serviceid, clusterid string, background, latest, force bool) {
	var realslid, realclid string

	//fmt.Printf("Serviceline: %s\n"+
	//	"Cluster ID: %s\n"+
	//	"Background: %v\n"+
	//	"Latest: %v\n"+
	//	"Force: %v\n", serviceid, clusterid, background, latest, force)

	slids := ovhwrapper.GetServicelines(reader)
	for _, slid := range slids {
		details := ovhwrapper.GetOVHServiceline(reader, slid)
		sl := ovhwrapper.ServiceLine{
			ID:        slid,
			SLDetails: *details,
		}
		if MatchItem(sl, serviceid) {
			realslid = sl.ID
			clids, err := ovhwrapper.GetK8SClusterIDs(reader, slid)
			if err != nil {
				fmt.Printf("Failed to get cluster IDs: %v\n", err)
				continue
			}

			for _, clid := range clids {
				cl := ovhwrapper.GetK8SCluster(reader, slid, clid)
				if MatchItem(*cl, clusterid) {
					realclid = cl.ID
				}
			}
		}
	}

	err := ovhwrapper.UpdateK8SCluster(writer, realslid, realclid, latest, force)
	if err != nil {
		log.Fatalf("Failed to initiate cluster update: %v", err)
	}

	if !background {
		time.Sleep(10 * time.Second) // give the update 10 seconds to get triggered

		for {
			client, err := ovhwrapper.CreateReader(config)
			if err != nil {
				log.Fatalf("Error creating OVH API Reader: %q\n", err)
			}

			cl := ovhwrapper.GetK8SCluster(client, realslid, realclid)
			if cl != nil {
				//fmt.Println("\033[2J")  // clear screen
				status(client, false, realslid, realclid)

				if cl.Status == "READY" {
					break
				}
				time.Sleep(60 * time.Second)
			}
		}
	}
}

func ResetKubeconfig(reader, writer *ovh.Client, config ovhwrapper.Configuration, serviceid, clusterid string, background bool) {
	var realslid, realclid string

	//fmt.Printf("Serviceline: %s\n"+
	//	"Cluster ID: %s\n"+
	//	"Background: %v\n", serviceid, clusterid, background)

	slids := ovhwrapper.GetServicelines(reader)
	for _, slid := range slids {
		details := ovhwrapper.GetOVHServiceline(reader, slid)
		sl := ovhwrapper.ServiceLine{
			ID:        slid,
			SLDetails: *details,
		}
		if MatchItem(sl, serviceid) {
			realslid = sl.ID
			clids, err := ovhwrapper.GetK8SClusterIDs(reader, slid)
			if err != nil {
				fmt.Printf("Failed to get cluster IDs: %v\n", err)
				continue
			}

			for _, clid := range clids {
				cl := ovhwrapper.GetK8SCluster(reader, slid, clid)
				if MatchItem(*cl, clusterid) {
					realclid = cl.ID
				}
			}
		}
	}

	if realslid == "" {
		log.Fatalf("Service line not found: %s\n", serviceid)
	}
	if realclid == "" {
		log.Fatalf("Cluster not found: %s\n", clusterid)
	}

	fmt.Printf("Resetting kubeconfig for serviceline %s (%s) cluster %s(%s)\n", serviceid, realslid, clusterid, realclid)
	kc, err := ovhwrapper.ResetKubeconfig(writer, realslid, realclid)
	if err != nil {
		log.Fatalf("Failed to initiate kubeconfig reset: %v", err)
	}
	fmt.Println(kc)

	if !background {
		time.Sleep(10 * time.Second) // give the reset 10 seconds to get triggered

		for {
			client, err := ovhwrapper.CreateReader(config)
			if err != nil {
				log.Fatalf("Error creating OVH API Reader: %q\n", err)
			}

			cl := ovhwrapper.GetK8SCluster(client, realslid, realclid)
			if cl != nil {
				//fmt.Println("\033[2J")  // clear screen
				status(client, false, realslid, realclid)

				if cl.Status == "READY" {
					break
				}
				time.Sleep(60 * time.Second)
			}
		}
	}
}

func Logout(writer *ovh.Client, config ovhwrapper.Configuration) {
	var result []byte
	if err := writer.Post("/auth/logout", nil, &result); err != nil {
		fmt.Printf("Error revoking consumer key: %q\n", err)
	}
	fmt.Println(string(result))
	config.Writer.ConsumerKey = ""

	err := ovhwrapper.SaveYaml(config, config.GetPath())
	if err != nil {
		log.Printf("Error saving configuration: %v", err)
	}
}
