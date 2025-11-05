package main

import (
	"context"
	"log"
	"os"

	"github.com/ovh/go-ovh/ovh"
	"github.com/snafuprinzip/ovhwrapper"
	"github.com/urfave/cli/v3"
)

// global command options
var debug bool
var verbose bool

var GlobalInventory []ovhwrapper.ServiceLine

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
		Name:      "ovhdbctl",
		Version:   "v0.0.1",
		Copyright: "(c) 2025 IHK GfI",
		Authors:   []any{"Michael Leimenmeier"},
		Usage:     "cli tool for the ovh database service api ",
		UsageText: "ovhdbctl <command> [subcommand] [options]",
		Flags:     globalFlags,
		Commands: []*cli.Command{
			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "list servicelines and/or databases",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "all", Aliases: []string{"a"}, Usage: "list all servicelines and databases"},
					&cli.StringFlag{Name: "serviceline", Aliases: []string{"s"}, Usage: "list databases of given serviceline"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					List(reader, cmd.Bool("all"), cmd.String("serviceline"))
					return nil
				},
			},
			{
				Name:    "status",
				Aliases: []string{"s"},
				Usage:   "show status of a serviceline or databases",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "all", Aliases: []string{"a"}, Usage: "all servicelines and databases"},
					&cli.StringFlag{Name: "serviceline", Aliases: []string{"s"}, Usage: "databases of a given serviceline"},
					&cli.StringFlag{Name: "database", Aliases: []string{"d"}, Usage: "specific database of a given serviceline"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					Status(reader, cmd.Bool("all"), cmd.String("serviceline"), cmd.String("db"))
					return nil
				},
			},
			{
				Name:    "describe",
				Aliases: []string{"d"},
				Usage:   "show details of a serviceline and or database(s)",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "all", Aliases: []string{"a"}, Usage: "all servicelines and databases"},
					&cli.StringFlag{Name: "serviceline", Aliases: []string{"s"}, Usage: "describe a serviceline"},
					&cli.StringFlag{Name: "database", Aliases: []string{"d"}, Usage: "describe a database of a given serviceline"},
					&cli.StringFlag{Name: "output", Aliases: []string{"o"}, Usage: "set output format [yaml, json, text]"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					Describe(reader, cmd.Bool("all"), cmd.String("serviceline"), cmd.String("database"), cmd.String("output"))
					return nil
				},
			},
			{
				Name:    "update",
				Aliases: []string{"u"},
				Usage:   "update kubernetes version",
				Commands: []*cli.Command{
					{
						Name:    "database",
						Aliases: []string{"d"},
						Usage:   "update a single database",
						Flags: []cli.Flag{
							&cli.StringFlag{Name: "serviceline", Aliases: []string{"s"}, Usage: "clusters of a given serviceline"},
							&cli.StringFlag{Name: "database", Aliases: []string{"d"}, Usage: "specific database of a given serviceline"},
							//&cli.BoolFlag{Name: "force", Aliases: []string{"f"}, Usage: "force update"},
							//&cli.BoolFlag{Name: "latest", Aliases: []string{"l"},
							//	Usage: "set strategy to LATEST_PATCH (default is NEXT_MINOR)"},
							//&cli.BoolFlag{Name: "background", Aliases: []string{"b"},
							//	Usage: "if not set the update status will be printed in 1 minute intervals until the cluster is READY again, " +
							//		"if background is set the program will exit immediately after starting the upgrade"},
						},
						Action: func(ctx context.Context, cmd *cli.Command) error {
							UpdateDatabase(reader, writer, config, cmd.String("serviceline"), cmd.String("database"))
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
