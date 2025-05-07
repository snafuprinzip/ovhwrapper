package ovhwrapper

import (
	"encoding/json"
	"fmt"
	"github.com/ovh/go-ovh/ovh"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type OVHVolume struct {
	Id           string    `json:"id"`
	AttachedTo   []string  `json:"attachedTo"`
	CreationDate time.Time `json:"creationDate"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Size         int       `json:"size"`
	Status       string    `json:"status"`
	Region       string    `json:"region"`
	Bootable     bool      `json:"bootable"`
	PlanCode     string    `json:"planCode"`
	Type         string    `json:"type"`
}

// ListOVHVolumes prints the list of OVHVolume with the fields Name, Size, Status, and Type.
func ListOVHVolumes(volumes []OVHVolume) {
	fmt.Printf("%-72s  %8s  %-10s  %s\n", "NAME", "SIZE", "STATUS", "TYPE")
	for _, volume := range volumes {
		fmt.Printf("%-72s  %5d GB  %-10s  %s\n", volume.Name, volume.Size, volume.Status, volume.Type)
	}
}

// DescribeOVHVolume prints volume details either as text (default), yaml or json
func DescribeOVHVolume(volume OVHVolume, output string) {
	switch output {
	case "yaml":
		fmt.Println(ToYaml(volume))
	case "json":
		fmt.Println(ToJSON(volume))
	default:
		fmt.Printf("Name:\t\t %s\n", volume.Name)
		fmt.Printf("ID:\t\t %s\n", volume.Id)
		fmt.Printf("Attached To:\t %v\n", volume.AttachedTo)
		fmt.Printf("Creation Date:\t %s\n", volume.CreationDate)
		fmt.Printf("Description:\t %s\n", volume.Description)
		fmt.Printf("Size:\t\t %d GB\n", volume.Size)
		fmt.Printf("Status:\t\t %s\n", volume.Status)
		fmt.Printf("Region:\t\t %s\n", volume.Region)
		fmt.Printf("Bootable:\t %t\n", volume.Bootable)
		fmt.Printf("Plan Code:\t %s\n", volume.PlanCode)
	}
	fmt.Println()
}

// GetOVHVolumes retrieves the list of Kubernetes volumes in a given service and cluster ID.
// It takes in an OVH client, the service name, and the cluster ID as parameters.
// It returns a []OVHVolume slice representing the list of volumes and an error if any occurred.
//
// The []OVHVolume slice is a collection of OVHVolume structs. Each OVHVolume struct contains information
// about a volume such as its ID, project ID, instance ID, volume pool ID, name, flavor, status,
// update status, version, creation timestamp, update timestamp, and deployment timestamp.
func GetOVHVolumes(client *ovh.Client, service, clusterid string) ([]OVHVolume, error) {
	var volumelist []OVHVolume
	//	volumelist:=  make([]OVHVolume, 3)
	if err := client.Get("/cloud/project/"+service+"/kube/"+clusterid+"/volume", &volumelist); err != nil {
		fmt.Printf("Error getting k8s volume list: %q\n", err)
		return volumelist, err
	}

	return volumelist, nil
}

// GetOVHVolume retrieves the details of a specific Kubernetes volume in a given service and cluster ID.
// It takes in an OVH client, the service name, the cluster ID, and the volume ID as parameters.
// It returns a OVHVolume struct representing the volume and an error if any occurred.
// The OVHVolume struct contains information about the volume such as its ID, project ID, instance ID,
// volume pool ID, name, flavor, status, update status, version, creation timestamp, update timestamp, and deployment timestamp.
func GetOVHVolume(client *ovh.Client, service, clusterid, volumeid string) (OVHVolume, error) {
	var volume OVHVolume
	//	volumelist:=  make([]OVHVolume, 3)
	if err := client.Get("/cloud/project/"+service+"/kube/"+clusterid+"/volume/"+volumeid, &volume); err != nil {
		fmt.Printf("Error getting k8s volume %s: %q\n", volumeid, err)
		return volume, err
	}

	return volume, nil
}

// ReadVolumesFromFile is a test function to read the disks from one cluster from file
func ReadVolumesFromFile() []OVHVolume {
	var volumes []OVHVolume

	file, err := os.Open("ovh_volumes.json")
	if err != nil {
		log.Fatalf("Failed to open file: %s", err)
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("Failed to read file: %s", err)
	}

	if err := json.Unmarshal(byteValue, &volumes); err != nil {
		log.Fatalf("Failed to unmarshal JSON: %s", err)
	}

	return volumes
}
