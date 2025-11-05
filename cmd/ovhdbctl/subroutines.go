package main

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/ovh/go-ovh/ovh"
	"github.com/snafuprinzip/ovhwrapper"
)

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
	dbsChan := make(chan []ovhwrapper.OVHDatabase)

	go func(detailChan chan<- ovhwrapper.OVHServiceLine) {
		servicedetails, err := ovhwrapper.GetServicelineDetails(client, projectID)
		if err != nil {
			log.Printf("Failed to get serviceline details: %v", err)
			detailChan <- ovhwrapper.OVHServiceLine{}
		} else {
			detailChan <- servicedetails
		}
	}(detailChan)

	dbIDs, err := ovhwrapper.GetDatabaseIDs(client, projectID)
	if err != nil {
		log.Fatalf("Failed to get database IDs: %v", err)
	}

	go GatherDatabases(client, projectID, dbIDs, dbsChan)

	serviceline := ovhwrapper.ServiceLine{
		ID:        projectID,
		SLDetails: <-detailChan,
		Databases: <-dbsChan,
	}

	projectChan <- serviceline
}

func GatherDatabases(client *ovh.Client, projectID string, dbIDs []uuid.UUID, dbsChan chan<- []ovhwrapper.OVHDatabase) {
	var databases []ovhwrapper.OVHDatabase
	dbChan := make(chan ovhwrapper.OVHDatabase)

	for _, databaseID := range dbIDs {
		go func(projectID string, databaseID uuid.UUID, dbsChan chan<- ovhwrapper.OVHDatabase) {
			GatherDatabase(client, projectID, databaseID, dbsChan)
		}(projectID, databaseID, dbChan)
	}

	for range len(dbIDs) {
		database := <-dbChan
		databases = append(databases, database)
	}
	close(dbChan)
	dbsChan <- databases
}

func GatherDatabase(client *ovh.Client, projectID string, databaseID uuid.UUID, dbsChan chan<- ovhwrapper.OVHDatabase) {
	database := ovhwrapper.GetDatabase(client, projectID, databaseID)

	if database != nil {
		dbsChan <- *database
	}
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

// List lists servicelines and their databases (when -a is set),
// servicelines (when -s is not set) or databases (when -s is set)
func List(client *ovh.Client, all bool, serviceid string) {
	if all { // list all servicelines and their databases
		for _, sl := range GlobalInventory {
			fmt.Printf("%-25s (%s)\n", sl.SLDetails.Description, sl.ID)
			for _, database := range sl.Databases {
				fmt.Printf("  %-40s (%s) \t %10s:%5s [%s]\n", database.Description,
					database.Id, database.Engine, database.Version, database.Status)
			}
			fmt.Println()
		}
	} else if serviceid != "" { // show databases for a specific serviceline
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

		fmt.Printf("%-25s (%s)\n", sl.SLDetails.Description, sl.ID)
		for _, database := range sl.Databases {
			fmt.Printf("  %-40s (%s) \t %10s:%5s [%s]\n", database.Description,
				database.Id, database.Engine, database.Version, database.Status)
		}
		fmt.Println()
	} else { // serviceid == ""
		for _, sl := range GlobalInventory {
			if len(sl.Databases) == 0 {
				continue
			}
			fmt.Printf("%-25s (%s)\n", sl.SLDetails.Description, sl.ID)
			for _, database := range sl.Databases {
				fmt.Printf("  %-40s (%s) \t %10s:%5s [%s]\n", database.Description,
					database.Id, database.Engine, database.Version, database.Status)
			}
			fmt.Println()
		}
	}
}

// Status shows the current status of servicelines and their databases (when -a is set),
// or a specific database (when -s and -d is set)
func Status(client *ovh.Client, all bool, serviceline, db string) {
}

// Describe shows the details of a specific databases (when -a is not set),
// a serviceline and all it's databases (when -s and -a are set)
// or all servicelines and their databases (when only -a is set)
func Describe(client *ovh.Client, all bool, serviceid, databaseid, output string) {
	if !all { // describe one specific database
		var sl ovhwrapper.ServiceLine
		var db ovhwrapper.OVHDatabase

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

		for _, d := range sl.Databases {
			if MatchItem(d, databaseid) {
				db = d
				break
			}

			if db.Id.String() == "00000000-0000-0000-0000-000000000000" {
				log.Printf("Database ID not found: %s", db.Id.String())
				return
			}

			fmt.Printf("%s\n---\n", ovhwrapper.ToYaml(db))
		}
		fmt.Println()

	} else { // all is true
		if serviceid == "" { // describe all servicelines and their databases
			for _, sl := range GlobalInventory {
				fmt.Printf("%s\n---\n", sl.Details())
				for _, database := range sl.Databases {
					fmt.Printf("%s\n---\n", ovhwrapper.ToYaml(database))
				}
				fmt.Println()
			}
		} else { // describe all databases for a specific serviceline
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

			for _, database := range sl.Databases {
				fmt.Printf("%s\n---\n", ovhwrapper.ToYaml(database))
			}
			fmt.Println()
		}
	}
}

func UpdateDatabase(reader, writer *ovh.Client, config ovhwrapper.Configuration, serviceid, db string) {
}
