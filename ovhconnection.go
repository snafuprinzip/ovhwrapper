package ovhwrapper

import (
	"fmt"
	"github.com/ovh/go-ovh/ovh"
	"gopkg.in/ini.v1"
	"log"
	"os"
	"time"
)

type Endpoint struct {
	AppKey      string `ini:"application_key"`
	AppSecret   string `ini:"application_secret"`
	ConsumerKey string `ini:"consumer_key"`
}

type OVHConfig struct {
	Default struct {
		Endpoint string `ini:"endpoint"`
	} `ini:"default"`
	EU Endpoint `ini:"ovh-eu"`
}

func CreateClient() (*ovh.Client, error) {
	client, err := ovh.NewEndpointClient("ovh-eu")
	if err != nil {
		fmt.Printf("Error creating new endpoint client: %q\n", err)
		return nil, err
	}

	return client, nil
}

func CreateConsumerKey(client *ovh.Client) (string, error) {
	ckReq := client.NewCkRequest()

	// Allow GET method on /cloud and all its sub routes
	ckReq.AddRecursiveRules(ovh.ReadOnly, "/cloud")

	// Run the request
	response, err := ckReq.Do()
	if err != nil {
		fmt.Printf("Error: %q\n", err)
		return "", err
	}

	// Print the validation URL and the Consumer key
	fmt.Printf("Generated consumer key: %s\n", response.ConsumerKey)
	fmt.Printf("Please visit %s to validate it\n", response.ValidationURL)

	return response.ConsumerKey, nil
}

func ReadConfig(path string) *OVHConfig {
	var conf OVHConfig

	inidata, err := ini.Load("ovh.conf")
	if err != nil {
		log.Fatalf("Fail to read file: %v", err)
	}

	err = inidata.MapTo(&conf)
	if err != nil {
		log.Fatalf("Fail to map file: %v", err)
	}

	return &conf
}

func (config *OVHConfig) Save(path string) {
	//  alte config Datei sichern
	currenttime, _ := time.Now().MarshalText()
	if err := os.Rename(path, "data/"+path+"_"+string(currenttime)); err != nil {
		log.Printf("Error renaming config file: %q\n", err)
	}

	// neue config speichern
	cfg := ini.Empty()
	err := ini.ReflectFrom(cfg, config)
	if err != nil {
		log.Printf("Error: %q\n", err)
	}
	err = cfg.SaveTo("ovh.conf")
	if err != nil {
		log.Printf("Error: %q\n", err)
	}
}
