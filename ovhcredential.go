package ovhwrapper

import (
	"fmt"
	"github.com/ovh/go-ovh/ovh"
	"time"
)

type OVHCredential struct {
	AllowedIPs    []string  `json:"allowedIPs"`
	ApplicationId int       `json:"applicationId"`
	Creation      time.Time `json:"creation"`
	CredentialId  int       `json:"credentialId"`
	Expiration    time.Time `json:"expiration"`
	LastUse       time.Time `json:"lastUse"`
	OvhSupport    bool      `json:"ovhSupport"`
	Rules         []struct {
		Method string `json:"method"`
		Path   string `json:"path"`
	} `json:"rules"`
	Status string `json:"status"`
}

func GetCredential(client *ovh.Client) (OVHCredential, error) {
	cred := OVHCredential{}
	if err := client.Get("/auth/currentCredential", &cred); err != nil {
		fmt.Printf("Error getting k8s cluster details: %q\n", err)
		return cred, err
	}
	return cred, nil
}
