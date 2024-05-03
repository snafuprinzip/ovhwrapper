package ovhwrapper

import (
	"fmt"
	"github.com/ovh/go-ovh/ovh"
	"strings"
	"time"
)

type OVHServiceLine struct {
	ProjectID    string    `json:"project_id"`
	ProjectName  string    `json:"projectName"`
	Description  string    `json:"description"`
	PlanCode     string    `json:"planCode"`
	Unleash      bool      `json:"unleash"`
	Expiration   any       `json:"expiration"`
	CreationDate time.Time `json:"creationDate"`
	OrderID      any       `json:"orderId"`
	Access       string    `json:"access"`
	Status       string    `json:"status"`
	ManualQuota  bool      `json:"manualQuota"`
	Iam          Iam       `json:"iam"`
}

type Iam struct {
	ID  string `json:"id"`
	Urn string `json:"urn"`
}

type ServiceLine struct {
	ID        string
	SLDetails OVHServiceLine
	Cluster   []K8SCluster
}

func (sl ServiceLine) Details() string {
	return fmt.Sprintf("Project ID: %s\n Project Name: %s\n Description: %s\n Plan Code: %s\n Unleash: %t\n "+
		"Expiration: %v\n Creation Date: %s\n Order ID: %v\n Access: %s\n Status: %s\n Manual Quota: %t\n IAM:\n%s\n",
		sl.SLDetails.ProjectID, sl.SLDetails.ProjectName, sl.SLDetails.Description, sl.SLDetails.PlanCode,
		sl.SLDetails.Unleash, sl.SLDetails.Expiration, sl.SLDetails.CreationDate, sl.SLDetails.OrderID,
		sl.SLDetails.Access, sl.SLDetails.Status, sl.SLDetails.ManualQuota, sl.SLDetails.Iam.Details())
}

func (sl ServiceLine) StatusMsg() string {
	return fmt.Sprintf("Serviceline: %-35s (%s)\t[%s]", sl.SLDetails.Description, sl.SLDetails.ProjectID,
		strings.ToUpper(sl.SLDetails.Status))
}

func (iam Iam) Details() string {
	return fmt.Sprintf("  ID: %s\n  URN: %s", iam.ID, iam.Urn)
}

func GetServicelineDetails(client *ovh.Client, service string) (OVHServiceLine, error) {
	var serviceline OVHServiceLine
	if err := client.Get(fmt.Sprintf("/cloud/project/%s", service), &serviceline); err != nil {
		return OVHServiceLine{}, err
	}
	return serviceline, nil
}

func GetServicelines(client *ovh.Client) []string {
	var servicelist []string

	if err := client.Get("/cloud/project/", &servicelist); err != nil {
		fmt.Printf("Error getting kube service list: %q\n", err)
		return servicelist
	}

	return servicelist
}

func GetOVHServiceline(client *ovh.Client, service string) *OVHServiceLine {
	serviceline := &OVHServiceLine{}

	if err := client.Get("/cloud/project/"+service, serviceline); err != nil {
		fmt.Printf("Error getting kube service list: %q\n", err)
		return nil
	}

	return serviceline
}
