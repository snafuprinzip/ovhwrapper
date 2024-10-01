// apiserver is an OVH API emulation for client testing purposes
// it reads a static cluster inventory from file and returns this information in the same way
// the OVH API does at the moment of writing

package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"net/http"
	"os"
	"ovhwrapper"
)

const debug = false

var projects []ovhwrapper.ServiceLine

func main() {
	fmt.Println("OVH API emulation server for client testing")

	f, err := os.Open("data/clusterdetails.yaml")
	if err != nil {
		log.Fatalf("cannot open cluster details: %s\n", err)
	}

	// marshall yaml from f
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&projects); err != nil {
		log.Fatalf("cannot decode cluster details: %s\n", err)
	}

	if debug {
		for _, project := range projects {
			fmt.Printf("%s\n", project.SLDetails.Description)
			for _, cluster := range project.Cluster {
				fmt.Printf("  %s\n", cluster.Name)
			}
		}
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /auth/time", GetCurrentTimeHandler)
	mux.HandleFunc("GET /cloud/project/", GetProjectsHandler)
	mux.HandleFunc("GET /cloud/project/{projectid}", GetProjectDetailsHandler)
	mux.HandleFunc("GET /cloud/project/{projectid}/kube/", GetClustersHandler)
	mux.HandleFunc("GET /cloud/project/{projectid}/kube/{clusterid}", GetClusterDetailsHandler)
	mux.HandleFunc("GET /cloud/project/{projectid}/kube/{clusterid}/node", GetNodesHandler)
	mux.HandleFunc("GET /cloud/project/{projectid}/kube/{clusterid}/node/{nodeid}", GetNodeDetailsHandler)
	mux.HandleFunc("GET /cloud/project/{projectid}/kube/{clusterid}/nodepool", GetNodepoolsHandler)
	mux.HandleFunc("GET /cloud/project/{projectid}/kube/{clusterid}/nodepool/{poolid}", GetNodepoolDetailsHandler)
	mux.HandleFunc("GET /cloud/project/{projectid}/kube/{clusterid}/metrics/etcdUsage", GetEtcdUsageHandler)
	mux.HandleFunc("GET /cloud/project/{projectid}/kube/{clusterid}/flavors", GetFlavorsHandler)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err.Error())
	}
}
