package ovhwrapper

import (
	"fmt"
	"github.com/ovh/go-ovh/ovh"
	"gopkg.in/ini.v1"
	"log"
	"os"
	"path"
	"time"
)

type Endpoint struct {
	AppKey      string `ini:"application_key" yaml:"application_key"`
	AppSecret   string `ini:"application_secret" yaml:"application_secret"`
	ConsumerKey string `ini:"consumer_key" yaml:"consumer_key"`
}

type OVHConfig struct {
	Default struct {
		Endpoint string `ini:"endpoint"`
	} `ini:"default"`
	EU Endpoint `ini:"ovh-eu"`
}

type Configuration struct {
	fpath  string
	Reader Endpoint `yaml:"reader" json:"reader"`
	Writer Endpoint `yaml:"writer" json:"writer"`
}

// GetPath returns the path of the config file used previously.
func (c *Configuration) GetPath() string {
	// just return fpath, so we don't need to export it, avoiding showing up in the yaml itself
	return c.fpath
}

func ReadConfiguration() (*Configuration, error) {
	var config Configuration
	homedir := os.Getenv("HOME")
	config.fpath = path.Join(homedir, ".ovhcredentials.conf")

	var locations []string = []string{
		"./ovhcredentials.conf",
		config.fpath, // home directory
		"/etc/k8s/ovhcredentials.conf",
	}

	for _, location := range locations {
		if _, err := os.Stat(location); err == nil {
			if !os.IsNotExist(err) {
				err := LoadYaml(&config, location)
				if err != nil {
					log.Printf("Error loading configuration file %s: %v", location, err)
					return nil, err
				}
				config.fpath = location
				break
			}
		}
	}
	return &config, nil
}

func CreateClient() (*ovh.Client, error) {
	client, err := ovh.NewEndpointClient("ovh-eu")
	if err != nil {
		fmt.Printf("Error creating new endpoint client: %q\n", err)
		return nil, err
	}

	return client, nil
}

func CreateReader(config Configuration) (*ovh.Client, error) {
	client, err := ovh.NewClient("ovh-eu", config.Reader.AppKey, config.Reader.AppSecret,
		config.Reader.ConsumerKey)
	if err != nil {
		fmt.Printf("Error creating new endpoint reader client: %q\n", err)
		return nil, err
	}

	return client, nil
}

func CreateWriter(config Configuration) (*ovh.Client, error) {
	client, err := ovh.NewClient("ovh-eu", config.Writer.AppKey, config.Writer.AppSecret,
		config.Writer.ConsumerKey)
	if err != nil {
		fmt.Printf("Error creating new endpoint writer client: %q\n", err)
		return nil, err
	}

	return client, nil
}

// CreateConsumerKey generates a consumer key for the writer by making multiple API requests to set
// up appropriate rules using the ovh.Client objects for reading and writing operations.
// It uses the GetServicelines function to fetch a list of available services,
// then calls the GetK8SClusterIDs function to retrieve a list of cluster IDs for each service.
// For each cluster, it adds the necessary read and write rules to the ckReq object.
// After running the request, it prints the validation URL and the generated consumer key for the writer.
// It returns the writer's consumer key and any errors encountered during the process.
func CreateConsumerKey(reader, writer *ovh.Client) (string, error) {
	ckReq := writer.NewCkRequest()

	// Allow GET method on /cloud and all its sub routes
	//ckReq.AddRecursiveRules(ovh.ReadOnly, "/cloud")

	for _, service := range GetServicelines(reader) {
		clusterList, err := GetK8SClusterIDs(reader, service)
		if err != nil {
			fmt.Printf("Error getting cluster list: %q\n", err)
			continue
		}

		for _, cluster := range clusterList {
			ckReq.AddRecursiveRules(ovh.ReadWriteSafe, fmt.Sprintf("/cloud/project/%s/kube/%s/kubeconfig", service, cluster))
			ckReq.AddRules(ovh.ReadWriteSafe, fmt.Sprintf("/cloud/project/%s/kube/%s/update", service, cluster))
		}
	}

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
