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

/********************************************************************
 *** Main Program Functions                                       ***
 ********************************************************************/

// list servicelines and/or clusters
// show cluster status
// show api credentials
// update cluster
// kubeconfig get
// kubeconfig update
// kubeconfig reset
// logout (revoke consumer key)

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
		UsageText: "ovhcon - cli tool for the ovh managed k8s api",
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
					status(reader, cmd.Bool("all"), cmd.String("serviceline"), cmd.String("cluster"))
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
						Usage: "get kubeconfig from ovh cloud and save them to file, to certificate files or update " +
							"entries in a central kubeconfig file",
						Flags: []cli.Flag{
							&cli.BoolFlag{Name: "all", Aliases: []string{"a"}, Usage: "all servicelines and clusters"},
							&cli.StringFlag{Name: "serviceline", Aliases: []string{"s"}, Usage: "serviceline id or name"},
							&cli.StringFlag{Name: "cluster", Aliases: []string{"c"}, Usage: "cluster id or name"},
							&cli.StringFlag{Name: "output", Aliases: []string{"o"}, Usage: "file, central or certs"},
						},
						Action: func(ctx context.Context, cmd *cli.Command) error {

							DownloadKubeconfig(reader, writer, cmd.Bool("all"), cmd.String("serviceline"),
								cmd.String("cluster"), cmd.String("output"))
							return nil
						},
					},
					{
						Name:  "reset",
						Usage: "reset kubeconfig of cluster in the ovh cloud, will redeploy the cluster and reinstall the nodes",
						Flags: []cli.Flag{
							&cli.StringFlag{Name: "serviceline", Aliases: []string{"s"}, Usage: "serviceline id or name"},
							&cli.StringFlag{Name: "cluster", Aliases: []string{"c"}, Usage: "cluster id or name"},
							&cli.BoolFlag{Name: "force", Aliases: []string{"f"}, Usage: "force update"},
							&cli.BoolFlag{Name: "latest", Aliases: []string{"l"},
								Usage: "set strategy to LATEST_PATCH (default is NEXT_MINOR)"},
							&cli.BoolFlag{Name: "background", Aliases: []string{"b"},
								Usage: "if not set the update status will be printed in 1 minute intervals until the cluster is READY again, " +
									"if background is set the program will exit immediately after starting the upgrade"},
						},
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
