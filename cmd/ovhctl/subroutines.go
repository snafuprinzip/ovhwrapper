package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ovh/go-ovh/ovh"
	"github.com/snafuprinzip/ovhwrapper"
	"gopkg.in/yaml.v3"
)

type Inventory struct {
	Clustergroups []Clustergroup `yaml:"clustergroups"`
}

type Clustergroup struct {
	Name     string      `yaml:"name"`
	Projects []CGProject `yaml:"servicelines"`
}

type CGProject struct {
	Name         string   `yaml:"name"`
	Email        string   `yaml:"email"`
	TeamsWebhook string   `yaml:"teamsWebhook"`
	Clusters     []string `yaml:"clusters"`
}

var Flavors ovhwrapper.K8SFlavors

func GatherGlobalInventory(client *ovh.Client) {
	GlobalInventory = []ovhwrapper.ServiceLine{}
	projectIDs := ovhwrapper.GetServicelines(client)

	projectChannel := make(chan ovhwrapper.ServiceLine)
	for _, projectID := range projectIDs {
		go GatherServiceline(client, projectID, projectChannel)
	}
	for len(GlobalInventory) < len(projectIDs) {
		GlobalInventory = append(GlobalInventory, <-projectChannel)
	}
}

func GatherServiceline(client *ovh.Client, projectID string, projectChan chan<- ovhwrapper.ServiceLine) {
	detailChan := make(chan ovhwrapper.OVHServiceLine)
	clustersChan := make(chan []ovhwrapper.K8SCluster)

	go func(detailChan chan<- ovhwrapper.OVHServiceLine) {
		servicedetails, err := ovhwrapper.GetServicelineDetails(client, projectID)
		if err != nil {
			log.Printf("Failed to get serviceline details: %v", err)
			detailChan <- ovhwrapper.OVHServiceLine{}
		} else {
			detailChan <- servicedetails
		}
	}(detailChan)

	clusterids, err := ovhwrapper.GetK8SClusterIDs(client, projectID)
	if err != nil {
		log.Fatalf("Failed to get cluster IDs: %v", err)
	}

	go GatherClusters(client, projectID, clusterids, clustersChan)

	serviceline := ovhwrapper.ServiceLine{
		ID:        projectID,
		SLDetails: <-detailChan,
		Cluster:   <-clustersChan,
	}

	projectChan <- serviceline
}

func GatherClusters(client *ovh.Client, projectID string, clusterids []string, clustersChan chan<- []ovhwrapper.K8SCluster) {
	var clusters []ovhwrapper.K8SCluster
	clusterChan := make(chan ovhwrapper.K8SCluster)

	for _, clusterID := range clusterids {
		go func(projectID string, clusterID string, clusterChan chan<- ovhwrapper.K8SCluster) {
			GatherCluster(client, projectID, clusterID, clusterChan)
		}(projectID, clusterID, clusterChan)
	}

	for range len(clusterids) {
		cluster := <-clusterChan
		clusters = append(clusters, cluster)
	}
	close(clusterChan)
	clustersChan <- clusters
}

func GatherCluster(client *ovh.Client, projectID string, clusterID string, clusterChan chan<- ovhwrapper.K8SCluster) {
	etcdChan := make(chan ovhwrapper.K8SEtcd)
	nodesChan := make(chan []ovhwrapper.K8SNode)
	nodepoolsChan := make(chan []ovhwrapper.K8SNodepool)

	cluster := ovhwrapper.GetK8SCluster(client, projectID, clusterID)
	//fmt.Println(clusterID, cluster.ID, cluster.Name)

	go GatherEtcd(client, projectID, clusterID, etcdChan)
	go GatherNodes(client, projectID, clusterID, nodesChan)
	go GatherNodepools(client, projectID, clusterID, nodepoolsChan)

	semaphore := 0
	for {
		select {
		case cluster.EtcdUsage = <-etcdChan:
			semaphore++
		case cluster.Nodes = <-nodesChan:
			semaphore++
		case cluster.Nodepools = <-nodepoolsChan:
			semaphore++
		}
		if semaphore == 3 {
			break
		}
	}

	clusterChan <- *cluster
}

func GatherEtcd(client *ovh.Client, projectID string, clusterID string, etcdChan chan<- ovhwrapper.K8SEtcd) {
	var etcd ovhwrapper.K8SEtcd
	etcd, err := ovhwrapper.GetK8SEtcd(client, projectID, clusterID)
	if err != nil {
		log.Printf("Error getting etcd usage: %q\n", err)
		etcdChan <- ovhwrapper.K8SEtcd{}
		return
	}
	etcdChan <- etcd
}

func GatherNodes(client *ovh.Client, projectID string, clusterID string, nodesChan chan<- []ovhwrapper.K8SNode) {
	var nodes []ovhwrapper.K8SNode
	nodes, err := ovhwrapper.GetK8SNodes(client, projectID, clusterID)
	if err != nil {
		log.Printf("Error getting nodes: %q\n", err)
		nodesChan <- []ovhwrapper.K8SNode{}
		return
	}
	nodesChan <- nodes
}

func GatherNodepools(client *ovh.Client, projectID string, clusterID string, nodepoolsChan chan<- []ovhwrapper.K8SNodepool) {
	var nodepools []ovhwrapper.K8SNodepool
	nodepools, err := ovhwrapper.GetK8SNodepools(client, projectID, clusterID)
	if err != nil {
		log.Printf("Error getting nodepools: %q\n", err)
		nodepoolsChan <- []ovhwrapper.K8SNodepool{}
		return
	}
	nodepoolsChan <- nodepools
}

// Credentials returns information about the reader and writer accounts in different formats (yaml, json or text)
func Credentials(reader, writer *ovh.Client, format string) {
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

// List lists servicelines and their clusters (when -a is set),
// servicelines (when -s is not set) or clusters (when -s is set)
func List(client *ovh.Client, all bool, serviceid string) {
	if all { // list all servicelines and their clusters
		for _, sl := range GlobalInventory {
			fmt.Printf("%-25s (%s) \t %s \n", ovhwrapper.ShortenName(sl.SLDetails.Description), sl.ID, sl.SLDetails.Description)
			for _, cluster := range sl.Cluster {
				fmt.Printf("  %-25s (%s) \t %s \n", ovhwrapper.ShortenName(cluster.Name), cluster.ID, cluster.Name)
			}
			fmt.Println()
		}
	} else if serviceid != "" { // show clusters for a specific serviceline
		var sl ovhwrapper.ServiceLine

		for _, s := range GlobalInventory {
			if MatchItem(s, serviceid) {
				sl = s
				break
			}
		}

		if sl.ID == "" {
			log.Printf("Service ID not found: %s", serviceid)
			return
		}

		fmt.Printf("%-25s (%s) \t %s \n", ovhwrapper.ShortenName(sl.SLDetails.Description), sl.ID, sl.SLDetails.Description)
		for _, cluster := range sl.Cluster {
			fmt.Printf("  %-25s (%s) \t %s \n", ovhwrapper.ShortenName(cluster.Name), cluster.ID, cluster.Name)
		}
		fmt.Println()
	} else { // serviceid == ""
		for _, sl := range GlobalInventory {
			fmt.Printf("%-25s (%s) \t %s \n", ovhwrapper.ShortenName(sl.SLDetails.Description), sl.ID, sl.SLDetails.Description)
		}
	}
}

func GetKubeConfig(reader, writer *ovh.Client, projectID string, clusterID, output, outpath string,
	globalconfig *ovhwrapper.KubeConfig) {

	kc, err := ovhwrapper.GetKubeconfig(writer, projectID, clusterID)
	if err != nil {
		log.Printf("Failed to get kubeconfig: %s", err)
		return
	}

	for _, project := range GlobalInventory {
		if project.ID == projectID {
			for _, cluster := range project.Cluster {
				if cluster.ID == clusterID {
					switch output {
					case "global":
						globalconfig.AddContext(kc)
					case "certs":
						certpath := path.Join(outpath, project.SLDetails.Description, cluster.Name)
						err := os.MkdirAll(certpath, 0700)
						if err != nil {
							log.Printf("Failed to create output directory: %v", err)
							continue
						}

						fmt.Printf("Extracting certificates for %s serviceline's %s cluster to %s...\n",
							project.SLDetails.Description, cluster.Name, certpath)

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
						fmt.Printf("Saving kubeconfig to %s...\n", path.Join(outpath, project.SLDetails.Description+"_"+cluster.Name+".yaml"))
						err = ovhwrapper.SaveYaml(kc, path.Join(outpath, project.SLDetails.Description+"_"+cluster.Name+".yaml"))
						if err != nil {
							log.Printf("Saving Kubeconfig failed: %v", err)
						}
					}
				}
			}
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

	if all {
		for _, sl := range GlobalInventory {
			fmt.Println("Processing Serviceline: ", sl.SLDetails.Description)
			for _, cl := range sl.Cluster {
				GetKubeConfig(reader, writer, sl.ID, cl.ID, output, outpath, &globalconfig)
			}
		}
	} else if serviceid != "" && clusterid != "" {
		for _, sl := range sls {
			if MatchItem(sl, serviceid) {
				for _, cl := range sl.Cluster {
					if MatchItem(cl, clusterid) {
						GetKubeConfig(reader, writer, sl.ID, cl.ID, output, outpath, &globalconfig)
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

// Status shows the current status of servicelines and their clusters (when -a is set),
// or a specific cluster (when -s and -c is set)
func Status(client *ovh.Client, all bool, serviceline, cluster string) {
	flavors, err := ovhwrapper.GetK8SFlavors(client, GlobalInventory[0].ID, GlobalInventory[0].Cluster[0].ID)

	if all {
		if err != nil {
			log.Printf("Error getting available flavors: %q\n", err)
		}

		for _, sl := range GlobalInventory {
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
		for _, service := range GlobalInventory {
			if MatchItem(service, serviceline) {
				fmt.Println(service.StatusMsg())
				for _, cl := range service.Cluster {
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

// statusString returns the current status of a specific cluster
func statusString(client *ovh.Client, serviceline, cluster string) string {
	var sls []ovhwrapper.ServiceLine
	var cl *ovhwrapper.K8SCluster
	var s string
	var realslid, realclid string
	var err error

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
			return ""
		}

		if len(sls[0].Cluster) == 0 {
			log.Printf("No clusters found for the given identifier: %s\n", cluster)
			return ""
		}

		flavors, err := ovhwrapper.GetK8SFlavors(client, sls[0].ID, sls[0].Cluster[0].ID)
		if err != nil {
			log.Printf("Error getting available flavors: %q\n", err)
			return ""
		}

		for _, sl := range sls {
			if MatchItem(sl, serviceline) {
				s += sl.StatusMsg() + "\n"
				for _, cl := range sl.Cluster {
					if cluster == "" || MatchItem(cl, cluster) {
						s += cl.StatusMsg() + "\n"
						for _, n := range cl.Nodes {
							f := flavors[n.Flavor]
							s += n.StatusMsg(f) + "\n"
						}
						s += "\n"
						return s
					}
				}
			}
		}
		return ""
	}
	return s
}

// Describe shows the details of servicelines and their clusters (when only -a is set),
// a serviceline (when -c is not set), a serviceline and all it's clusters (when -a is set as well)
// or a specific cluster (when -s and -c is set, including serviceline if -a is set as well)
func Describe(client *ovh.Client, all bool, serviceid, clusterid, output string) {
	if serviceid == "" && clusterid == "" { // all servicelines and their clusters

		// output
		switch output {
		case "yaml":
			fmt.Println(ovhwrapper.ToYaml(GlobalInventory))
		case "json":
			fmt.Println(ovhwrapper.ToJSON(GlobalInventory))
		case "text":
			fallthrough
		default:
			for _, sl := range GlobalInventory {
				fmt.Println(sl.Details())
				for _, cluster := range sl.Cluster {
					fmt.Println(cluster.Details())
					fmt.Println(cluster.EtcdUsage.Details())
					fmt.Println()
					for _, n := range cluster.Nodes {
						fmt.Println(n.Details())
						f := Flavors[n.Flavor]
						fmt.Println(f.Details())
						fmt.Println()
					}
					for _, p := range cluster.Nodepools {
						fmt.Println(p.Details())
						fmt.Println()
					}
					fmt.Println("\n-----")
				}
				fmt.Println("\n\n")
			}
		}
	} else if serviceid != "" && clusterid == "" { // show all clusters for a specific serviceline
		var sl ovhwrapper.ServiceLine

		for _, s := range GlobalInventory {
			if MatchItem(s, serviceid) {
				sl = s
				break
			}
		}

		if sl.ID == "" {
			log.Printf("Service ID not found: %s", serviceid)
			return
		}

		// output
		switch output {
		case "yaml":
			fmt.Println(ovhwrapper.ToYaml(sl))
		case "json":
			fmt.Println(ovhwrapper.ToJSON(sl))
		case "text":
			fallthrough
		default:
			fmt.Println(sl.Details())
			for _, cluster := range sl.Cluster {
				fmt.Println(cluster.Details())
				fmt.Println(cluster.EtcdUsage.Details())
				fmt.Println()
				for _, n := range cluster.Nodes {
					fmt.Println(n.Details())
					f := Flavors[n.Flavor]
					fmt.Println(f.Details())
					fmt.Println()
				}
				for _, p := range cluster.Nodepools {
					fmt.Println(p.Details())
					fmt.Println()
				}
				fmt.Println("\n-----")
			}
			//fmt.Println()
		}
	} else if serviceid != "" && clusterid != "" {
		var sl ovhwrapper.ServiceLine

		for _, s := range GlobalInventory {
			if MatchItem(s, serviceid) {
				sl = s
				break
			}
		}

		if sl.ID == "" {
			log.Printf("Service ID not found: %s", serviceid)
			return
		}

		var cluster *ovhwrapper.K8SCluster
		for _, cl := range sl.Cluster {
			if MatchItem(cl, clusterid) {
				cluster = &cl
				break
			}
		}

		// output
		switch output {
		case "yaml":
			if all {
				fmt.Println(ovhwrapper.ToYaml(sl))
			} else {
				fmt.Println(ovhwrapper.ToYaml(cluster))
			}
		case "json":
			if all {
				fmt.Println(ovhwrapper.ToJSON(sl))
			} else {
				fmt.Println(ovhwrapper.ToJSON(cluster))
			}
		case "text":
			fallthrough
		default:
			if cluster == nil {
				log.Printf("Cluster ID not found for service %s", serviceid)
				return
			}

			if all {
				fmt.Println(sl.Details())
				fmt.Println()
			}

			fmt.Println(cluster.Details())
			fmt.Println(cluster.EtcdUsage.Details())
			if all {
				fmt.Println("Nodes:")
				fmt.Println()
				for _, n := range cluster.Nodes {
					fmt.Println(n.Details())
					f := Flavors[n.Flavor]
					fmt.Println(f.Details())
					fmt.Println()
				}
				for _, p := range cluster.Nodepools {
					fmt.Println(p.Details())
					fmt.Println()
				}
			}
			fmt.Println("\n-----")
		}
		//fmt.Println()
	}
}

func readInventory(reader *ovh.Client, config ovhwrapper.Configuration, inventory string) Inventory {
	if inventory == "" {
		// check if local inventory file "./clustergroups.yaml" exists
		if fileExists("./clustergroups.yaml") {
			inventory = "./clustergroups.yaml"
		} else if fileExists("/etc/k8s/clustergroups.yaml") {
			inventory = "/etc/k8s/clustergroups.yaml"
		} else {
			log.Fatalf("No inventory file found. Please specify one with the -i flag.")
		}
	} else {
		if !fileExists(inventory) {
			log.Fatalf("Inventory file %s not found", inventory)
		}
	}

	// open inventory file
	inventoryString, err := os.ReadFile(inventory)
	if err != nil {
		log.Fatalf("Failed to open inventory file %s: %v", inventory, err)
	}

	// read file and convert yaml to Inventory struct
	var inv Inventory
	err = yaml.Unmarshal(inventoryString, &inv)
	if err != nil {
		log.Fatalf("Failed to parse inventory file %s: %v", inventory, err)
	}
	return inv
}

func UpdateClusterGroup(reader, writer *ovh.Client, config ovhwrapper.Configuration, clustergroup, inventory string, latest, force bool) {
	var wg sync.WaitGroup
	var status chan string

	// read inventory file
	inv := readInventory(reader, config, inventory)
	fmt.Printf("%s\n", inv)

	// count number of clusters to update
	var count int
	for _, cg := range inv.Clustergroups {
		if cg.Name == clustergroup {
			for _, project := range cg.Projects {
				count += len(project.Clusters)
			}
			break
		}
	}

	// create buffered channel
	status = make(chan string, count)

	fmt.Printf("Updating %d clusters in group %s\n", count, clustergroup)

	for _, cg := range inv.Clustergroups {
		// find selected cluster group
		if cg.Name == clustergroup {
			// every service line in cluster group
			for _, project := range cg.Projects {
				// every cluster in service line
				for _, clustername := range project.Clusters {
					realslid := ""
					realclid := ""
					slEmail := project.Email
					slTeamsHook := project.TeamsWebhook

					// determine project and cluster IDs
					slids := ovhwrapper.GetServicelines(reader)
					for _, slid := range slids {
						details := ovhwrapper.GetOVHServiceline(reader, slid)
						sl := ovhwrapper.ServiceLine{
							ID:        slid,
							SLDetails: *details,
						}
						if MatchItem(sl, project.Name) {
							realslid = sl.ID
							clids, err := ovhwrapper.GetK8SClusterIDs(reader, slid)
							if err != nil {
								fmt.Printf("Failed to get cluster IDs: %v\n", err)
								continue
							}

							for _, clid := range clids {
								cl := ovhwrapper.GetK8SCluster(reader, slid, clid)
								if MatchItem(*cl, clustername) {
									realclid = cl.ID
								}
							}
						}
					}

					wg.Add(1)
					go func(wg *sync.WaitGroup, sl, slid, cl, clid string) {
						defer wg.Done()
						fmt.Printf("Updating cluster %25s (%s) in serviceline %25s (%s)\n", cl, clid, sl, slid)
						err := ovhwrapper.UpdateK8SCluster(writer, slid, clid, latest, force)
						if err != nil {
							log.Fatalf("Failed to initiate cluster update: %v", err)
						}
						res := CheckCronClusterUpdate(reader, writer, config, sl, slid, cl, clid, slEmail, slTeamsHook)
						status <- res
					}(&wg, project.Name, realslid, clustername, realclid)
				} // cluster
			} // project
			break
		} // cluster group
	}
	wg.Wait()
	close(status)
	fmt.Printf("\n%d clusters in group %s updated:\n\n", count, clustergroup)
	for result := range status {
		fmt.Println(result)
	}
}

func CheckCronClusterUpdate(reader, writer *ovh.Client, config ovhwrapper.Configuration, sl, realslid, cl, realclid, email, teamshook string) string {
	var curStatus, prevStatus string

	logfile, err := os.OpenFile(path.Join("/var/log/k8s/updates", sl+"-"+cl+".log"), os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
	}
	defer logfile.Close()

	fmt.Fprintf(logfile, "Update for %s started at %s...\n\n", cl, time.Now().Format(time.RFC1123Z))

	// mock sleep to simulate a little bit of update time
	time.Sleep(time.Second * time.Duration(rand.Intn(30)))

	for {
		client, err := ovhwrapper.CreateReader(config)
		if err != nil {
			log.Fatalf("Error creating OVH API Reader: %q\n", err)
		}

		cl := ovhwrapper.GetK8SCluster(client, realslid, realclid)
		if cl != nil {
			//fmt.Println("\033[2J")  // clear screen
			curStatus = statusString(client, realslid, realclid)
			// show status if status has changed since last check
			if curStatus != prevStatus {
				fmt.Fprintln(logfile, curStatus)
				prevStatus = curStatus
			}

			// end update loop if cluster is in ready state
			if cl.Status == "READY" {
				break
			}
			time.Sleep(60 * time.Second)
		}
	}
	fmt.Fprintf(logfile, "Update for %s finished at %s...\n", cl, time.Now().Format(time.RFC1123Z))
	err = logfile.Close()
	if err != nil {
		log.Printf("Failed to close log file: %v", err)
	}

	recipients := []string{"michael.leimenmeier@gfi.ihk.de", "SLAP.Application-Hosting-Operations@gfi.ihk.de"}
	if email != "" {
		recipients = append(recipients, email)
	}

	teamshooks := []string{}
	if teamshook != "" {
		teamshooks = append(teamshooks, teamshook)
	}

	logtext, err := os.ReadFile(path.Join("/var/log/k8s/updates", sl+"-"+cl+".log"))
	if err != nil {
		log.Fatalf("Failed to read log file: %v", err)
	} else {
		log.Printf("Sending mail for %s to %s...\n", cl, strings.Join(recipients, ", "))
		err := SendMail("k8s Update: "+cl, string(logtext), recipients)
		if err != nil {
			log.Printf("Failed to send mail: %v", err)
		}
		err = TeamsNotify("k8s Update: "+cl, string(logtext), teamshooks)
		if err != nil {
			log.Printf("Failed to send teams notification: %v", err)
		}
	}
	return curStatus
}

func MockCheckClusterUpdate(reader, writer *ovh.Client, config ovhwrapper.Configuration, realslid, realclid string) string {
	time.Sleep(time.Second * time.Duration(rand.Intn(60)))
	return statusString(reader, realslid, realclid)
}

func UpdateCluster(reader, writer *ovh.Client, config ovhwrapper.Configuration, serviceid, clusterid string, background, latest, force bool) {
	var realslid, realclid string
	var curStatus, prevStatus string

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
				curStatus = statusString(client, realslid, realclid)
				// show status if status has changed since last check
				if curStatus != prevStatus {
					fmt.Println(curStatus)
					prevStatus = curStatus
				}

				// end update loop if cluster is in ready state
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
				Status(client, false, realslid, realclid)

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

func (i Inventory) String() string {
	var s string
	for _, cg := range i.Clustergroups {
		s += fmt.Sprintf("%s\n", cg.Name)
		for _, sl := range cg.Projects {
			s += fmt.Sprintf("  %s\n", sl.Name)
			for _, cl := range sl.Clusters {
				s += fmt.Sprintf("    %s\n", cl)
			}
		}
	}
	return s
}

// ListFlavors lists available nodepool flavors
func ListFlavors(client *ovh.Client, flavors ovhwrapper.K8SFlavors) {

	var flavorList []ovhwrapper.K8SFlavor
	for _, flavor := range flavors {
		flavorList = append(flavorList, flavor)
	}

	sort.Slice(flavorList, func(i, j int) bool {
		return flavorList[i].Name < flavorList[j].Name
	})

	for _, flavor := range flavorList {
		fmt.Printf("%-12s %3d cpu, %4d gb ram, %2d gpu\n", flavor.Name, flavor.VCPUs, flavor.RAM, flavor.GPUs)
	}
}

func ListOVHVolumes(reader, writer *ovh.Client, all bool, serviceid, clusterid string, unattachedOnly bool) {
	volumes := ovhwrapper.ReadVolumesFromFile()
	var filteredVolumes []ovhwrapper.OVHVolume

	if unattachedOnly {
		for _, volume := range volumes {
			if volume.Status != "in-use" {
				filteredVolumes = append(filteredVolumes, volume)
			}
		}
		volumes = filteredVolumes
	}
	ovhwrapper.ListOVHVolumes(volumes)
}

func DescribeOVHVolumes(reader, writer *ovh.Client, all bool, serviceid, clusterid string, unattachedOnly bool, output string) {
	volumes := ovhwrapper.ReadVolumesFromFile()
	var filteredVolumes []ovhwrapper.OVHVolume

	if unattachedOnly {
		for _, volume := range volumes {
			if volume.Status != "in-use" {
				filteredVolumes = append(filteredVolumes, volume)
			}
		}
		volumes = filteredVolumes
	}

	for _, volume := range volumes {
		ovhwrapper.DescribeOVHVolume(volume, output)
	}
}

func DeleteOVHVolume(reader, writer *ovh.Client, serviceid, clusterid string, force bool) {

}
