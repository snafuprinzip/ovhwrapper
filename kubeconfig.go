package ovhwrapper

import (
	"fmt"
	"github.com/ovh/go-ovh/ovh"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"slices"
	"sort"
	"strings"
)

type KubeConfig struct {
	APIVersion     string      `yaml:"apiVersion"`
	Clusters       []Clusters  `yaml:"clusters"`
	Contexts       []Contexts  `yaml:"contexts"`
	CurrentContext string      `yaml:"current-context"`
	Kind           string      `yaml:"kind"`
	Preferences    Preferences `yaml:"preferences"`
	Users          []Users     `yaml:"users"`
}
type Cluster struct {
	CertificateAuthorityData string `yaml:"certificate-authority-data"`
	Server                   string `yaml:"server"`
}
type Clusters struct {
	Cluster Cluster `yaml:"cluster"`
	Name    string  `yaml:"name"`
}
type Context struct {
	Cluster   string `yaml:"cluster"`
	User      string `yaml:"user"`
	Namespace string `yaml:"namespace,omitempty"`
}
type Contexts struct {
	Context Context `yaml:"context"`
	Name    string  `yaml:"name"`
}
type Preferences struct {
}
type User struct {
	ClientCertificateData string `yaml:"client-certificate-data"`
	ClientKeyData         string `yaml:"client-key-data"`
}
type Users struct {
	Name string `yaml:"name"`
	User User   `yaml:"user"`
}

// ShortenName returns a shortened version of the given string
func ShortenName(name string) string {
	shortname := name
	shortname = strings.TrimPrefix(shortname, "kubernetes-admin@")
	shortname = strings.TrimPrefix(shortname, "sl_")
	shortname = strings.TrimPrefix(shortname, "ovh-k8s-")
	shortname = strings.TrimPrefix(shortname, "sl-")
	shortname = strings.TrimPrefix(shortname, "app-plat-")
	shortname = strings.Replace(shortname, "-00", "", 1)
	return shortname
}

// ListContexts listet alle Kontexte der globalen kubeconfig unter /etc/k8s/config auf
func (c *KubeConfig) ListContexts(localConfigPath string) {
	var oldConfig KubeConfig
	var currentContext string
	var currentNamespace string

	activeSign := " "
	sortedContexts := c.Contexts
	sort.Slice(sortedContexts, func(i, j int) bool {
		return sortedContexts[i].Name < sortedContexts[j].Name
	})

	// lokale config einlesen, falls diese existiert
	if _, err := os.Stat(localConfigPath); err == nil {
		readerr := LoadYaml(oldConfig, localConfigPath)
		if readerr != nil {
			log.Printf("Warnung: Kann lokale Konfiguration nicht lesen: %v\n", readerr)
		} else {
			currentContext = oldConfig.CurrentContext
			for _, cntCon := range oldConfig.Contexts {
				if cntCon.Name == currentContext {
					currentNamespace = cntCon.Context.Namespace
					break
				}
			}
		}
	}

	var wide bool
	if term.IsTerminal(0) {
		w, _, _ := term.GetSize(0)
		if w > 175 {
			wide = true
		}
	}

	if wide {
		fmt.Printf("%-3s %-35s %-50s %-60s %-30s\n", "CUR", "NAME", "CLUSTER", "AUTHINFO", "NAMESPACE")
	} else {
		fmt.Printf("%-3s %-35s %-50s\n", "CUR", "NAME", "CLUSTER")
	}

	var ns string
	for _, con := range sortedContexts {
		{
			if con.Name == currentContext {
				activeSign = " * "
				ns = currentNamespace
			} else {
				activeSign = "   "
				ns = ""
			}
			if wide {
				fmt.Printf("%-3s %-35s %-50s %-60s %-30s\n", activeSign, con.Name, con.Context.Cluster, con.Context.User, ns)
			} else {
				fmt.Printf("%-3s %-35s %-50s\n", activeSign, con.Name, con.Context.Cluster)
			}
		}
	}
}

func (c *KubeConfig) AddContext(newConfig KubeConfig) {
	if len(newConfig.Contexts) == 0 {
		return
	}

	newConfig.Contexts[0].Name = ShortenName(newConfig.Contexts[0].Name)
	if _, conptr := c.GetContext(newConfig.Contexts[0].Name); conptr != nil {
		log.Printf("Kontext %s existiert bereits in der globalen config und wurde daher nicht hinzugefuegt.",
			newConfig.Contexts[0].Name)
	} else {
		c.Clusters = append(c.Clusters, newConfig.Clusters[0])
		c.Contexts = append(c.Contexts, newConfig.Contexts[0])
		c.Users = append(c.Users, newConfig.Users[0])
	}
}

func (c *KubeConfig) GetContext(contextname string) (int, *Contexts) {
	for conidx, con := range c.Contexts {
		if con.Name == contextname {
			return conidx, &con
		}
	}
	return -1, nil
}

func (c *KubeConfig) RemoveContext(contextname string) {
	var confound, clfound, userfound bool
	for conidx, con := range c.Contexts {
		if con.Name == contextname {
			confound = true
			for clidx, cl := range c.Clusters {
				if cl.Name == con.Context.Cluster {
					clfound = true
					c.Clusters = slices.Delete(c.Clusters, clidx, clidx+1)
					break
				}
			}
			if !clfound {
				log.Printf("Kann Cluster %s nicht in der Config finden\n", con.Context.Cluster)
			}

			for useridx, user := range c.Users {
				if user.Name == con.Context.User {
					userfound = true
					c.Users = slices.Delete(c.Users, useridx, useridx+1)
					break
				}
			}
			if !userfound {
				log.Printf("Kann User %s nicht in der Config finden\n", con.Context.User)
			}

			c.Contexts = slices.Delete(c.Contexts, conidx, conidx+1)
			break
		}
	}
	if !confound {
		log.Fatalf("Kann Kontext %s nicht in Config finden.", contextname)
	}
}

func GetKubeconfig(client *ovh.Client, service, clusterid string) (KubeConfig, error) {
	type kcresponse struct {
		Content string `json:"content"`
	}

	var response kcresponse
	var kubeconfig KubeConfig

	url := fmt.Sprintf("/cloud/project/%s/kube/%s/kubeconfig", service, clusterid)
	if err := client.Post(url, nil, &response); err != nil {
		fmt.Printf("Error recieving kubeconfig (url: %s): %q\n", url, err)
		return kubeconfig, err
	}

	err := yaml.Unmarshal([]byte(response.Content), &kubeconfig)
	if err != nil {
		fmt.Printf("Error unmarshaling kubeconfig: %q\n", err)
		return kubeconfig, err
	}
	return kubeconfig, nil
}

func ResetKubeconfig(client *ovh.Client, service, clusterid string) (KubeConfig, error) {
	var kubeconfig KubeConfig

	if err := client.Post("/cloud/project/"+service+"/kube/"+clusterid+"/kubeconfig/reset", nil, &kubeconfig); err != nil {
		fmt.Printf("Error resetting kubeconfig: %q\n", err)
		return kubeconfig, err
	}

	//fmt.Println("Description updated")
	return kubeconfig, nil
}
