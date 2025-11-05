package main

import (
	"fmt"
	"github.com/ovh/go-ovh/ovh"
	"github.com/snafuprinzip/ovhwrapper"
	"log"
)

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
}

// Status shows the current status of servicelines and their databases (when -a is set),
// or a specific database (when -s and -d is set)
func Status(client *ovh.Client, all bool, serviceline, db string) {
}

// Describe shows the details of servicelines and their databases (when only -a is set),
// a serviceline (when -d is not set), a serviceline and all it's databases (when -a is set as well)
// or a specific cluster (when -s and -d is set, including serviceline if -a is set as well)
func Describe(client *ovh.Client, all bool, serviceid, db, output string) {
}
