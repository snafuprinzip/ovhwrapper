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

func PrintCredential(cred *OVHCredential) {
	str := fmt.Sprintf("Application ID: %d\n"+
		"Credential ID: %d\n"+
		"Status: %s\n"+
		"Creation: %v\n"+
		"Expiration: %v\n"+
		"Last Use: %v\n"+
		"OVH Support: %v\nRules:\n",
		cred.ApplicationId, cred.CredentialId, cred.Status, cred.Creation, cred.Expiration, cred.LastUse, cred.OvhSupport)

	for _, rule := range cred.Rules {
		str += fmt.Sprintf("  %-7s %s", rule.Method, rule.Path)
	}
	str += "\nAllowed IPs: \n"
	for _, ip := range cred.AllowedIPs {
		str += fmt.Sprintf("  - %s\n", ip)
	}

	fmt.Println(str)
}

func GetCredential(client *ovh.Client) (OVHCredential, error) {
	cred := OVHCredential{}
	if err := client.Get("/auth/currentCredential", &cred); err != nil {
		return cred, err
	}
	return cred, nil
}
