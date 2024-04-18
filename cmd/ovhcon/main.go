package main

import (
	"context"
	"fmt"
	"github.com/ovh/go-ovh/ovh"
	"github.com/snafuprinzip/ovhwrapper"
	"github.com/urfave/cli/v3"
	"log"
	"os"
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
	clusterlist := make([]ovhwrapper.K8SCluster, 1)
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

/********************************************************************
 *** Main Program Functions                                       ***
 ********************************************************************/

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
	sls := make([]ovhwrapper.ServiceLine, 5)
	if all {
		slids := ovhwrapper.GetServicelines(client)
		for _, slid := range slids {
			var err error
			sl := ovhwrapper.ServiceLine{ID: slid}
			sl.SLDetails, err = ovhwrapper.GetServicelineDetails(client, slid)
			if err != nil {
				log.Fatalf("Failed to get service lines: %v", err)
			}
			clusterids, err := ovhwrapper.GetK8SClusterIDs(client, slid)
			if err != nil {
				log.Fatalf("Failed to get cluster IDs: %v", err)
			}
			clusterlist := make([]ovhwrapper.K8SCluster, 1)
			for _, clusterid := range clusterids {
				cluster := GetCluster(client, serviceid, clusterid)
				if cluster != nil {
					clusterlist = append(clusterlist, *cluster)
				}
			}
			sl.Cluster = clusterlist
			sls = append(sls, sl)
		}
	}

	if serviceid != "" {
		var err error
		sl := ovhwrapper.ServiceLine{ID: serviceid}
		sl.SLDetails, err = ovhwrapper.GetServicelineDetails(client, serviceid)
		if err != nil {
			log.Fatalf("Failed to get service lines: %v", err)
		}
		clusterids, err := ovhwrapper.GetK8SClusterIDs(client, serviceid)
		if err != nil {
			log.Fatalf("Failed to get cluster IDs: %v", err)
		}
		clusterlist := make([]ovhwrapper.K8SCluster, 1)
		for _, clusterid := range clusterids {
			cluster := GetCluster(client, serviceid, clusterid)
			if cluster != nil {
				clusterlist = append(clusterlist, *cluster)
			}
		}
		sl.Cluster = clusterlist
		sls = append(sls, sl)
	}

	for _, sl := range sls {
		fmt.Printf("%-25s (%s) \t %s \n", ovhwrapper.ShortenName(sl.SLDetails.Description), sl.ID, sl.SLDetails.Description)
		for _, cluster := range sl.Cluster {
			fmt.Printf("  %-25s (%s) \t %s \n", ovhwrapper.ShortenName(cluster.Name), cluster.ID, cluster.Name)
		}
		fmt.Println()
	}
}

// status shows the current status of servicelines and their clusters (when -a is set),
// or a specific cluster (when -s and -c is set)
func status(sls []ovhwrapper.ServiceLine, flavors ovhwrapper.K8SFlavors, all bool, serviceline, cluster string) {
	if all {
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

// show cluster status
// show api credentials
// update cluster
// kubeconfig get
// kubeconfig update
// kubeconfig reset

// options
//   -a all clusters
//   -s serviceline
//   -c cluster
//   -o outputformat (yaml, certs)
//   -w wait for READY status

func main() {
	var reader *ovh.Client
	var writer *ovh.Client

	config, err := ovhwrapper.ReadConfiguration()
	if err != nil {
		log.Fatalf("Error loading configuration, no valid config found: %v", err)
	}

	//client, err := ovhwrapper.CreateClient()
	reader, err = ovhwrapper.CreateReader(*config)
	if err != nil {
		log.Fatalf("Error creating OVH API Reader: %q\n", err)
	}

	writer, err = ovhwrapper.CreateWriter(*config)
	if err != nil {
		log.Fatalf("Error creating OVH API Writer: %q\n", err)
	}

	// create Writer ConsumerKey if necessary
	if config.Writer.ConsumerKey == "" {
		// consumer key erzeugen
		consumerkey, err := ovhwrapper.CreateConsumerKey(reader, writer)
		if err != nil {
			log.Fatalf("Error: %q\n", err)
		}
		config.Writer.ConsumerKey = consumerkey

		err = ovhwrapper.SaveYaml(config, config.GetPath())
		if err != nil {
			log.Fatalf("Error saving config file %s: %q\n", config.GetPath(), err)
		}
		os.Exit(0)
	}

	cmd := &cli.Command{
		Name:      "ovhcon",
		Version:   "v0.0.1",
		Copyright: "(c) 2014 Michael Leimenmeier",
		Usage:     "cli tool for the ovh api",
		UsageText: "ovhcon - cli tool for the ovh api",
		Commands: []*cli.Command{
			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "list servicelines and/or clusters",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "all", Aliases: []string{"a"}, Usage: "list all servicelines and clusters"},
					&cli.StringFlag{Name: "serviceline", Aliases: []string{"s"}, Usage: "list clusters of given serviceline"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					list(reader, cmd.Bool("all"), cmd.String("serviceline"))
					return nil
				},
			},
			{
				Name:    "credentials",
				Aliases: []string{"cred"},
				Usage:   "shows the credentials used for api access",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "output", Aliases: []string{"o"}, Usage: "set output format [yaml, json, text]"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					credentials(reader, writer, cmd.String("output"))
					return nil
				},
			},
			{
				Name:    "status",
				Aliases: []string{"s"},
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "all", Aliases: []string{"a"}, Usage: "all servicelines and clusters"},
					&cli.StringFlag{Name: "serviceline", Aliases: []string{"s"}, Usage: "clusters of a given serviceline"},
					&cli.StringFlag{Name: "cluster", Aliases: []string{"c"}, Usage: "specific cluster of a given serviceline"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println("Collecting cluster information...")
					sls := CollectInformation(reader)
					fmt.Println("Collecting information about available flavors...")
					flavors, err := ovhwrapper.GetK8SFlavors(reader, sls[0].ID, sls[0].Cluster[0].ID)
					if err != nil {
						log.Printf("Error getting available flavors: %q\n", err)
					}

					status(sls, flavors, cmd.Bool("all"), cmd.String("serviceline"), cmd.String("cluster"))
					return nil
				},
			},
			{
				Name:    "kubeconfig",
				Aliases: []string{"kc"},
				Usage:   "kubernetes client configuration",
				Commands: []*cli.Command{
					{
						Name:    "get",
						Aliases: []string{"g"},
						Usage:   "get kubeconfig from ovh cloud",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							fmt.Println("get kubeconfig: ", cmd.Args().First())
							return nil
						},
					},
					{
						Name:    "update",
						Aliases: []string{"u"},
						Usage:   "update kubeconfig in global config from ovh cloud",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							fmt.Println("update kubeconfig in global config file: ", cmd.Args().First())
							return nil
						},
					},
					{
						Name:  "reset",
						Usage: "reset kubeconfig of cluster in the ovh cloud, will redeploy the cluster and reinstall the nodes",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							fmt.Println("reset kubeconfig: ", cmd.Args().First())
							return nil
						},
					},
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}

	//ShowStatusAll(reader)
	//ShowCredentials(reader, writer)

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

}
