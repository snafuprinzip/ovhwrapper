package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"ovhwrapper"
	"time"
)

func GetProjectsHandler(w http.ResponseWriter, r *http.Request) {
	if debug {
		log.Println("GetProjectsHandler")
	}
	//if f, err := os.Open("data/apiserver/projects.json"); err != nil {
	//	log.Printf("error opening projects.json file: %s\n", err)
	//} else {
	//	defer f.Close()
	//	_, err = io.Copy(w, f)
	//	if err != nil {
	//		log.Printf("error copying projects.json to http writer: %s\n", err)
	//	}
	//}

	p := []string{}
	for _, project := range projects {
		p = append(p, project.ID)
	}

	jsonData, err := json.Marshal(p)
	if err != nil {
		log.Printf("error marshalling projects list to JSON: %s\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	_, err = w.Write(jsonData)
	if err != nil {
		log.Printf("error writing projects list to http writer: %s\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func GetProjectDetailsHandler(w http.ResponseWriter, r *http.Request) {
	if debug {
		log.Println("GetProjectDetailsHandler")
	}
	projectID := r.PathValue("projectid")

	//if f, err := os.Open(path.Join("data/apiserver/projects/", projectID)); err != nil {
	//	log.Printf("error opening project %s.json file: %s\n", projectID, err)
	//} else {
	//	defer f.Close()
	//	_, err = io.Copy(w, f)
	//	if err != nil {
	//		log.Printf("error copying project %s.json to http writer: %s\n", projectID, err)
	//	}
	//}
	for _, project := range projects {
		if project.ID == projectID {
			jsonData, err := json.Marshal(project.SLDetails)
			if err != nil {
				log.Printf("error marshalling project %s details to JSON: %s\n", projectID, err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")

			_, err = w.Write(jsonData)
			if err != nil {
				log.Printf("error writing project %s details to http writer: %s\n", projectID, err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			return
		}
	}
}

func GetClustersHandler(w http.ResponseWriter, r *http.Request) {
	if debug {
		log.Println("GetClustersHandler")
	}
	projectID := r.PathValue("projectid")

	cl := []string{}
	for _, project := range projects {
		if project.ID == projectID {
			for _, cluster := range project.Cluster {
				cl = append(cl, cluster.ID)

			}
		}
	}

	jsonData, err := json.Marshal(cl)
	if err != nil {
		log.Printf("error marshalling cluster list to JSON: %s\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	_, err = w.Write(jsonData)
	if err != nil {
		log.Printf("error writing cluster list to http writer: %s\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func GetClusterDetailsHandler(w http.ResponseWriter, r *http.Request) {
	if debug {
		log.Println("GetClusterDetailsHandler")
	}
	projectID := r.PathValue("projectid")
	clusterID := r.PathValue("clusterid")

	for _, project := range projects {
		if project.ID == projectID {
			for _, cluster := range project.Cluster {
				if cluster.ID == clusterID {
					jsonData, err := json.Marshal(cluster)
					if err != nil {
						log.Printf("error marshalling cluster %s details to JSON: %s\n", clusterID, err)
						http.Error(w, "Internal server error", http.StatusInternalServerError)
						return
					}
					w.Header().Set("Content-Type", "application/json")

					_, err = w.Write(jsonData)
					if err != nil {
						log.Printf("error writing cluster %s details to http writer: %s\n", clusterID, err)
						http.Error(w, "Internal server error", http.StatusInternalServerError)
						return
					}
					return
				}
			}
		}
	}
}

func GetNodesHandler(w http.ResponseWriter, r *http.Request) {
	if debug {
		log.Println("GetNodesHandler")
	}
	projectID := r.PathValue("projectid")
	clusterID := r.PathValue("clusterid")

	nd := []ovhwrapper.K8SNode{}
	for _, project := range projects {
		if project.ID == projectID {
			for _, cluster := range project.Cluster {
				if cluster.ID == clusterID {
					for _, node := range cluster.Nodes {
						nd = append(nd, node)
					}
				}
			}
		}
	}

	jsonData, err := json.Marshal(nd)
	if err != nil {
		log.Printf("error marshalling node list to JSON: %s\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	_, err = w.Write(jsonData)
	if err != nil {
		log.Printf("error writing node list to http writer: %s\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func GetNodeDetailsHandler(w http.ResponseWriter, r *http.Request) {
	if debug {
		log.Println("GetNodeDetailsHandler")
	}
	projectID := r.PathValue("projectid")
	clusterID := r.PathValue("clusterid")
	nodeID := r.PathValue("nodeid")

	for _, project := range projects {
		if project.ID == projectID {
			for _, cluster := range project.Cluster {
				if cluster.ID == clusterID {
					for _, node := range cluster.Nodes {
						if node.Id == nodeID {
							jsonData, err := json.Marshal(node)
							if err != nil {
								log.Printf("error marshalling node %s details to JSON: %s\n", nodeID, err)
								http.Error(w, "Internal server error", http.StatusInternalServerError)
								return
							}
							w.Header().Set("Content-Type", "application/json")

							_, err = w.Write(jsonData)
							if err != nil {
								log.Printf("error writing node %s details to http writer: %s\n", nodeID, err)
								http.Error(w, "Internal server error", http.StatusInternalServerError)
								return
							}
							return
						}
					}
				}
			}
		}
	}
	log.Printf("warning: node %s not found\n", nodeID)
	http.Error(w, "Node "+nodeID+" not found", http.StatusNotFound)
}

func GetNodepoolsHandler(w http.ResponseWriter, r *http.Request) {
	if debug {
		log.Println("GetNodepoolsHandler")
	}
	projectID := r.PathValue("projectid")
	clusterID := r.PathValue("clusterid")

	np := []ovhwrapper.K8SNodepool{}
	for _, project := range projects {
		if project.ID == projectID {
			for _, cluster := range project.Cluster {
				if cluster.ID == clusterID {
					for _, nodepool := range cluster.Nodepools {
						np = append(np, nodepool)
					}
				}
			}
		}
	}

	jsonData, err := json.Marshal(np)
	if err != nil {
		log.Printf("error marshalling nodepool list to JSON: %s\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	_, err = w.Write(jsonData)
	if err != nil {
		log.Printf("error writing nodepool list to http writer: %s\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func GetNodepoolDetailsHandler(w http.ResponseWriter, r *http.Request) {
	if debug {
		log.Println("GetNodepoolDetailsHandler")
	}
	projectID := r.PathValue("projectid")
	clusterID := r.PathValue("clusterid")
	nodepoolID := r.PathValue("poolid")

	for _, project := range projects {
		if project.ID == projectID {
			for _, cluster := range project.Cluster {
				if cluster.ID == clusterID {
					for _, nodepool := range cluster.Nodepools {
						if nodepool.Id == nodepoolID {
							jsonData, err := json.Marshal(nodepool)
							if err != nil {
								log.Printf("error marshalling nodepool %s details to JSON: %s\n", nodepoolID, err)
								http.Error(w, "Internal server error", http.StatusInternalServerError)
								return
							}
							w.Header().Set("Content-Type", "application/json")

							_, err = w.Write(jsonData)
							if err != nil {
								log.Printf("error writing nodepool %s details to http writer: %s\n", nodepoolID, err)
								http.Error(w, "Internal server error", http.StatusInternalServerError)
								return
							}
							return
						}
					}
				}
			}
		}
	}
	log.Printf("warning: nodepool %s not found\n", nodepoolID)
	http.Error(w, "Nodepool "+nodepoolID+" not found", http.StatusNotFound)
}

func GetEtcdUsageHandler(w http.ResponseWriter, r *http.Request) {
	if debug {
		log.Println("GetEtcdUsageHandler")
	}
	projectID := r.PathValue("projectid")
	clusterID := r.PathValue("clusterid")

	etcdUsage := ovhwrapper.K8SEtcd{}

	for _, project := range projects {
		if project.ID == projectID {
			for _, cluster := range project.Cluster {
				if cluster.ID == clusterID {
					etcdUsage = cluster.EtcdUsage
				}
			}
		}
	}

	jsonData, err := json.Marshal(etcdUsage)
	if err != nil {
		log.Printf("error marshalling etcd usage of cluster %s to JSON: %s\n", clusterID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	_, err = w.Write(jsonData)
	if err != nil {
		log.Printf("error writing etcd usage of cluster %s to http writer: %s\n", clusterID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func GetFlavorsHandler(w http.ResponseWriter, r *http.Request) {
	if debug {
		log.Println("GetFlavorsHandler")
	}

	if f, err := os.Open("data/apiserver/flavors.json"); err != nil {
		log.Printf("error opening flavors.json file: %s\n", err)
	} else {
		defer f.Close()

		w.Header().Set("Content-Type", "application/json")
		_, err = io.Copy(w, f)
		if err != nil {
			log.Printf("error copying flavors.json to http writer: %s\n", err)
		}
	}
}

func GetCurrentTimeHandler(w http.ResponseWriter, r *http.Request) {
	if debug {
		log.Println("GetCurrentTimeHandler")
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, err := fmt.Fprintf(w, "%d\n", time.Now().Unix())
	if err != nil {
		log.Printf("error writing current time to http writer: %s\n", err)
	}
}
