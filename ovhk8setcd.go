package ovhwrapper

import (
	"fmt"
	"github.com/ovh/go-ovh/ovh"
)

type K8SEtcd struct {
	Quota int `json:"quota"`
	Usage int `json:"usage"`
}

func GetK8SEtcd(client *ovh.Client, service, clusterid string) (K8SEtcd, error) {
	etcd := K8SEtcd{}
	if err := client.Get("/cloud/project/"+service+"/kube/"+clusterid+"/metrics/etcdUsage", &etcd); err != nil {
		fmt.Printf("Error getting k8s etcd usage: %q\n", err)
		return etcd, err
	}
	return etcd, nil
}
