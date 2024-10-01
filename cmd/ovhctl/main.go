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

// global command options
var debug bool
var verbose bool

var GlobalInventory []ovhwrapper.ServiceLine

/********************************************************************
 *** Main Program Functions                                       ***
 ********************************************************************/

func main() {
	var reader *ovh.Client
	var writer *ovh.Client
	var config ovhwrapper.Configuration

	config, err := ovhwrapper.ReadConfiguration()
	if err != nil {
		log.Fatalf("Error loading configuration, no valid config found: %v", err)
	}

	reader, err = ovhwrapper.CreateReader(config)
	if err != nil {
		log.Fatalf("Error creating OVH API Reader: %q\n", err)
	}

	writer, err = ovhwrapper.CreateWriter(config)
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

	GatherGlobalInventory(reader)
	Flavors, err = ovhwrapper.GetK8SFlavors(reader, GlobalInventory[0].ID, GlobalInventory[0].Cluster[0].ID)
	if err != nil {
		log.Fatalf("Error getting available flavors: %q\n", err)
	}

	for _, project := range GlobalInventory {
		fmt.Println(project.SLDetails.Description)
		for _, cluster := range project.Cluster {
			fmt.Printf("  %s: %s\n", cluster.ID, cluster.Name)
			for _, node := range cluster.Nodes {
				fmt.Printf("    %s\n", node.Name)
			}
			for _, nodepool := range cluster.Nodepools {
				fmt.Printf("    %s\n", nodepool.Name)
			}
		}
	}

	globalFlags := []cli.Flag{
		&cli.BoolFlag{
			Name:  "debug",
			Usage: "enable debug output",
			Action: func(_ context.Context, cmd *cli.Command, b bool) error {
				debug = true
				return nil
			},
		},
		&cli.BoolFlag{
			Name:  "verbose",
			Usage: "be more verbose",
			Action: func(_ context.Context, command *cli.Command, b bool) error {
				verbose = true
				return nil
			},
		},
	}

	cmd := &cli.Command{
		Name:      "ovhctl",
		Version:   "v0.1.2",
		Copyright: "(c) 2024 Michael Leimenmeier",
		Usage:     "cli tool for the ovh api",
		UsageText: "ovhctl <command> [subcommand] [options]",
		Flags:     globalFlags,
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
					List(reader, cmd.Bool("all"), cmd.String("serviceline"))
					return nil
				},
			},
			{
				Name:    "status",
				Aliases: []string{"s"},
				Usage:   "show status of a serviceline or cluster",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "all", Aliases: []string{"a"}, Usage: "all servicelines and clusters"},
					&cli.StringFlag{Name: "serviceline", Aliases: []string{"s"}, Usage: "clusters of a given serviceline"},
					&cli.StringFlag{Name: "cluster", Aliases: []string{"c"}, Usage: "specific cluster of a given serviceline"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					Status(reader, cmd.Bool("all"), cmd.String("serviceline"), cmd.String("cluster"))
					return nil
				},
			},
			{
				Name:    "describe",
				Aliases: []string{"d"},
				Usage:   "show details of a serviceline and or cluster(s)",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "all", Aliases: []string{"a"}, Usage: "all servicelines and clusters, including nodes and nodepools"},
					&cli.StringFlag{Name: "serviceline", Aliases: []string{"s"}, Usage: "describe a serviceline"},
					&cli.StringFlag{Name: "cluster", Aliases: []string{"c"}, Usage: "describe a cluster of a given serviceline"},
					&cli.StringFlag{Name: "output", Aliases: []string{"o"}, Usage: "set output format [yaml, json, text]"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					Describe(reader, cmd.Bool("all"), cmd.String("serviceline"), cmd.String("cluster"), cmd.String("output"))
					return nil
				},
			},
			{
				Name:    "update",
				Aliases: []string{"u"},
				Usage:   "update kubernetes version",
				Commands: []*cli.Command{
					{
						Name:    "cluster",
						Aliases: []string{"c"},
						Usage:   "update a single cluster",
						Flags: []cli.Flag{
							&cli.StringFlag{Name: "serviceline", Aliases: []string{"s"}, Usage: "clusters of a given serviceline"},
							&cli.StringFlag{Name: "cluster", Aliases: []string{"c"}, Usage: "specific cluster of a given serviceline"},
							&cli.BoolFlag{Name: "force", Aliases: []string{"f"}, Usage: "force update"},
							&cli.BoolFlag{Name: "latest", Aliases: []string{"l"},
								Usage: "set strategy to LATEST_PATCH (default is NEXT_MINOR)"},
							&cli.BoolFlag{Name: "background", Aliases: []string{"b"},
								Usage: "if not set the update status will be printed in 1 minute intervals until the cluster is READY again, " +
									"if background is set the program will exit immediately after starting the upgrade"},
						},
						Action: func(ctx context.Context, cmd *cli.Command) error {
							UpdateCluster(reader, writer, config, cmd.String("serviceline"), cmd.String("cluster"),
								cmd.Bool("background"), cmd.Bool("latest"), cmd.Bool("force"))
							return nil
						},
					},
					{
						Name:    "group",
						Aliases: []string{"g"},
						Usage:   "update a group of clusters",
						Flags: []cli.Flag{
							&cli.StringFlag{Name: "clustergroup", Aliases: []string{"s"}, Usage: "name of a group of clusters"},
							&cli.StringFlag{Name: "inventory", Aliases: []string{"i"}, Usage: "inventory file"},
							&cli.BoolFlag{Name: "force", Aliases: []string{"f"}, Usage: "force update"},
							&cli.BoolFlag{Name: "latest", Aliases: []string{"l"},
								Usage: "set strategy to LATEST_PATCH (default is NEXT_MINOR)"},
							&cli.BoolFlag{Name: "background", Aliases: []string{"b"},
								Usage: "if not set the cluster status will be printed in 1 minute intervals until the clusters all clusters are READY again, " +
									"if background is set the program will exit immediately after starting the upgrades"},
						},
						Action: func(ctx context.Context, cmd *cli.Command) error {
							UpdateClusterGroup(reader, writer, config, cmd.String("clustergroup"), cmd.String("inventory"),
								cmd.Bool("background"), cmd.Bool("latest"), cmd.Bool("force"))
							return nil
						},
					},
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
							&cli.StringFlag{Name: "path", Aliases: []string{"p"}, Usage: "output path"},
						},
						Action: func(ctx context.Context, cmd *cli.Command) error {

							DownloadKubeconfig(reader, writer, cmd.Bool("all"), cmd.String("serviceline"),
								cmd.String("cluster"), cmd.String("output"), cmd.String("path"))
							return nil
						},
					},
					{
						Name:  "reset",
						Usage: "reset kubeconfig of cluster in the ovh cloud, will redeploy the cluster and reinstall the nodes",
						Flags: []cli.Flag{
							&cli.StringFlag{Name: "serviceline", Aliases: []string{"s"}, Usage: "serviceline id or name"},
							&cli.StringFlag{Name: "cluster", Aliases: []string{"c"}, Usage: "cluster id or name"},
							&cli.BoolFlag{Name: "background", Aliases: []string{"b"},
								Usage: "if not set the cluster status will be printed in 1 minute intervals until the cluster is READY again, " +
									"if background is set the program will exit immediately after starting the reset"},
						},
						Action: func(ctx context.Context, cmd *cli.Command) error {
							ResetKubeconfig(reader, writer, config, cmd.String("serviceline"),
								cmd.String("cluster"), cmd.Bool("background"))
							//fmt.Println("reset kubeconfig: ", cmd.Args().First())
							return nil
						},
					},
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
					Credentials(reader, writer, cmd.String("output"))
					return nil
				},
			},
			{
				Name:    "logout",
				Aliases: []string{"o"},
				Usage:   "revoke consumer key, next time the command will be run it will create a new consumer key",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					Logout(writer, config)
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
