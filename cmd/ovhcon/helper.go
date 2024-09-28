package main

import (
	goteamsnotify "github.com/atc0005/go-teams-notify/v2"
	"github.com/atc0005/go-teams-notify/v2/adaptivecard"
	"github.com/ovh/go-ovh/ovh"
	"github.com/snafuprinzip/ovhwrapper"
	"log"
	"net/smtp"
	"os"
	"strings"
)

// CollectInformation collects the information of all service lines, including their clusters down to the nodes.
func CollectInformation(client *ovh.Client) []ovhwrapper.ServiceLine {
	var servicelines []ovhwrapper.ServiceLine

	services := ovhwrapper.GetServicelines(client)
	for _, service := range services {
		serviceline := CollectServiceline(client, service)
		servicelines = append(servicelines, *serviceline)
	}

	return servicelines
}

// GetServiceline asks the API for 'shallow' information about a specific serviceline, excluding the clusters.
func GetServiceline(client *ovh.Client, serviceid string) *ovhwrapper.ServiceLine {
	servicedetails, err := ovhwrapper.GetServicelineDetails(client, serviceid)
	if err != nil {
		log.Fatalf("Failed to get serviceline details: %v", err)
	}
	serviceline := ovhwrapper.ServiceLine{
		ID:        serviceid,
		SLDetails: servicedetails,
		Cluster:   []ovhwrapper.K8SCluster{},
	}
	return &serviceline
}

// CollectServiceline collects information about a serviceline, including its clusters.
func CollectServiceline(client *ovh.Client, serviceid string) *ovhwrapper.ServiceLine {
	servicedetails, err := ovhwrapper.GetServicelineDetails(client, serviceid)
	if err != nil {
		log.Fatalf("Failed to get serviceline details: %v", err)
	}
	clusterids, err := ovhwrapper.GetK8SClusterIDs(client, serviceid)
	if err != nil {
		log.Fatalf("Failed to get cluster IDs: %v", err)
	}
	var clusterlist []ovhwrapper.K8SCluster
	for _, clusterid := range clusterids {
		cluster := CollectCluster(client, serviceid, clusterid)
		if cluster != nil {
			clusterlist = append(clusterlist, *cluster)
		}
	}
	serviceline := ovhwrapper.ServiceLine{
		ID:        serviceid,
		SLDetails: servicedetails,
		Cluster:   clusterlist,
	}

	return &serviceline
}

// GetCluster asks the API for 'shallow' information about a specific cluster, excluding nested information
// like etcd usage, nodes or nodepools
func GetCluster(client *ovh.Client, serviceid, clusterid string) *ovhwrapper.K8SCluster {
	return ovhwrapper.GetK8SCluster(client, serviceid, clusterid)
}

// CollectCluster returns information about a Cluster, including its etcd usage, nodepools and nodes.
func CollectCluster(client *ovh.Client, serviceid, clusterid string) *ovhwrapper.K8SCluster {
	cluster := ovhwrapper.GetK8SCluster(client, serviceid, clusterid)
	var err error

	cluster, err = ovhwrapper.GetK8SClusterDetails(client, cluster, serviceid, clusterid)
	if err != nil {
		log.Printf("Failed to get cluster details: %v", err)
	}
	return cluster
}

// MatchItem will check if the id or the (abbreviated) name matches with the identifier and returns true or false
func MatchItem[T ovhwrapper.ServiceLine | ovhwrapper.K8SCluster](object T, identifier string) bool {
	match := false
	switch object := any(object).(type) { // lazy hack, as generic functions implement specific types and not interfaces, so we cast to any to check its type
	case ovhwrapper.ServiceLine:
		if object.ID == identifier || object.SLDetails.Description == identifier || ovhwrapper.ShortenName(object.SLDetails.Description) == identifier {
			match = true
		}
	case ovhwrapper.K8SCluster:
		if object.ID == identifier || object.Name == identifier || ovhwrapper.ShortenName(object.Name) == identifier {
			match = true
		}
	}
	return match
}

// fileExists returns true if a file exists, false if not
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func SendMail(subject, body string, to []string) error {
	r := strings.NewReplacer("\r\n", "", "\r", "", "\n", "", "%0a", "", "%0d", "")

	addr := "127.0.0.1:25"
	from := "k8s updater <ovhcon@example.com>"

	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer c.Close()

	if err = c.Mail(r.Replace(from)); err != nil {
		return err
	}
	for i := range to {
		to[i] = r.Replace(to[i])
		if err = c.Rcpt(to[i]); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	msg := "To: " + strings.Join(to, ",") + "\r\n" +
		"From: " + from + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"\n\n" +
		`
<html>
 <body>
   <pre>
` + body +
		`
  </pre>
 </body>
</html>
`
	//"Content-Transfer-Encoding: base64\r\n" +
	//"\r\n" + base64.StdEncoding.EncodeToString([]byte(body))

	_, err = w.Write([]byte(msg))
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}

func TeamsNotify(subject, body string, webhooks []string) error {
	mstClient := goteamsnotify.NewTeamsClient()
	webhookDefault := ""
	webhooks = append(webhooks, webhookDefault)

	// create simple message
	//msgText := "```" + body + "```"
	//msg, err := adaptivecard.NewSimpleMessage(msgText, subject, true)
	//if err != nil {
	//	log.Printf("failed to create message: %v\n", err)
	//	return err
	//}

	// or embed logfile as code snippet
	card, err := adaptivecard.NewTextBlockCard("Update Log:", subject, false)
	if err != nil {
		log.Printf("failed to create card: %v", err)
		return err
	}

	// Create CodeBlock using our snippet.
	codeBlock := adaptivecard.NewCodeBlock(body, "Bash", 0)

	// Add CodeBlock to our Card.
	if err := card.AddElement(false, codeBlock); err != nil {
		log.Printf("failed to add codeblock to card: %v", err)
		return err
	}

	// Create Message from Card
	msg, err := adaptivecard.NewMessageFromCard(card)
	if err != nil {
		log.Printf("failed to create message from card: %v", err)
		return err
	}

	// send message
	for _, webhookUrl := range webhooks {
		if err := mstClient.Send(webhookUrl, msg); err != nil {
			log.Printf("failed to send message: %v\n", err)
			return err
		}
	}

	return nil
}
