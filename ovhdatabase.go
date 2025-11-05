package ovhwrapper

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/ovh/go-ovh/ovh"

	"time"
)

type OVHDatabases []OVHDatabase

type OVHDatabase struct {
	CreatedAt time.Time `json:"createdAt"`
	Plan      string    `json:"plan"`
	Disk      struct {
		Type string `json:"type"`
		Size int    `json:"size"`
	} `json:"disk"`
	Storage struct {
		Type string `json:"type"`
		Size struct {
			Unit  string `json:"unit"`
			Value int    `json:"value"`
		} `json:"size"`
	} `json:"storage"`
	Id        uuid.UUID `json:"id"`
	Engine    string    `json:"engine"`
	Category  string    `json:"category"`
	Endpoints []struct {
		Component string      `json:"component"`
		Domain    string      `json:"domain"`
		Port      int         `json:"port"`
		Path      interface{} `json:"path"`
		Scheme    string      `json:"scheme"`
		Ssl       bool        `json:"ssl"`
		SslMode   string      `json:"sslMode"`
		Uri       string      `json:"uri"`
	} `json:"endpoints"`
	IpRestrictions []struct {
		Ip          string `json:"ip"`
		Description string `json:"description"`
		Status      string `json:"status"`
	} `json:"ipRestrictions"`
	Status string `json:"status"`
	Nodes  []struct {
		Id        string    `json:"id"`
		CreatedAt time.Time `json:"createdAt"`
		Flavor    string    `json:"flavor"`
		Name      string    `json:"name"`
		Port      int       `json:"port"`
		Region    string    `json:"region"`
		Status    string    `json:"status"`
	} `json:"nodes"`
	NodeNumber      int    `json:"nodeNumber"`
	Description     string `json:"description"`
	Version         string `json:"version"`
	NetworkType     string `json:"networkType"`
	NetworkId       string `json:"networkId"`
	SubnetId        string `json:"subnetId"`
	Flavor          string `json:"flavor"`
	MaintenanceTime string `json:"maintenanceTime"`
	BackupTime      string `json:"backupTime"`
	Backups         struct {
		Time          string    `json:"time"`
		Regions       []string  `json:"regions"`
		RetentionDays int       `json:"retentionDays"`
		Pitr          time.Time `json:"pitr"`
	} `json:"backups"`
	Capabilities struct {
		AdvancedConfiguration struct {
			Read   string `json:"read"`
			Update string `json:"update"`
		} `json:"advancedConfiguration"`
		BackupTime struct {
			Read   string `json:"read"`
			Update string `json:"update"`
		} `json:"backupTime"`
		Backups struct {
			Read string `json:"read"`
		} `json:"backups"`
		Certificates struct {
			Read string `json:"read"`
		} `json:"certificates"`
		ConnectionPools struct {
			Create string `json:"create"`
			Read   string `json:"read"`
			Update string `json:"update"`
			Delete string `json:"delete"`
		} `json:"connectionPools"`
		CurrentQueries struct {
			Read string `json:"read"`
		} `json:"currentQueries"`
		CurrentQueriesCancel struct {
			Create string `json:"create"`
		} `json:"currentQueriesCancel"`
		Databases struct {
			Create string `json:"create"`
			Read   string `json:"read"`
			Delete string `json:"delete"`
		} `json:"databases"`
		DeletionProtection struct {
			Read   string `json:"read"`
			Update string `json:"update"`
		} `json:"deletionProtection"`
		EnableWrites struct {
			Create string `json:"create"`
		} `json:"enableWrites"`
		Fork struct {
			Create string `json:"create"`
		} `json:"fork"`
		Integrations struct {
			Create string `json:"create"`
			Read   string `json:"read"`
			Delete string `json:"delete"`
		} `json:"integrations"`
		IpRestrictions struct {
			Create string `json:"create"`
			Read   string `json:"read"`
			Update string `json:"update"`
			Delete string `json:"delete"`
		} `json:"ipRestrictions"`
		MaintenanceApply struct {
			Create string `json:"create"`
		} `json:"maintenanceApply"`
		MaintenanceTime struct {
			Read   string `json:"read"`
			Update string `json:"update"`
		} `json:"maintenanceTime"`
		Maintenances struct {
			Read string `json:"read"`
		} `json:"maintenances"`
		Nodes struct {
			Read string `json:"read"`
		} `json:"nodes"`
		Prometheus struct {
			Read string `json:"read"`
		} `json:"prometheus"`
		PrometheusCredentialsReset struct {
			Create string `json:"create"`
		} `json:"prometheusCredentialsReset"`
		QueryStatistics struct {
			Read string `json:"read"`
		} `json:"queryStatistics"`
		QueryStatisticsReset struct {
			Create string `json:"create"`
		} `json:"queryStatisticsReset"`
		Service struct {
			Read   string `json:"read"`
			Update string `json:"update"`
			Delete string `json:"delete"`
		} `json:"service"`
		ServiceDisk struct {
			Read   string `json:"read"`
			Update string `json:"update"`
		} `json:"serviceDisk"`
		ServiceFlavor struct {
			Read   string `json:"read"`
			Update string `json:"update"`
		} `json:"serviceFlavor"`
		ServiceIpRestriction struct {
			Read   string `json:"read"`
			Update string `json:"update"`
		} `json:"serviceIpRestriction"`
		UserCredentialsReset struct {
			Create string `json:"create"`
		} `json:"userCredentialsReset"`
		Users struct {
			Create string `json:"create"`
			Read   string `json:"read"`
			Update string `json:"update"`
			Delete string `json:"delete"`
		} `json:"users"`
	} `json:"capabilities"`
	EnablePrometheus   bool `json:"enablePrometheus"`
	DeletionProtection bool `json:"deletionProtection"`
}

func GetDatabaseIDs(client *ovh.Client, service string) ([]uuid.UUID, error) {
	var dblist []uuid.UUID

	if err := client.Get("/cloud/project/"+service+"/database/service", &dblist); err != nil {
		fmt.Printf("Error getting database list: %q\n", err)
		return dblist, err
	}

	return dblist, nil
}

func GetDatabase(client *ovh.Client, service string, databaseID uuid.UUID) *OVHDatabase {
	var db OVHDatabase
	id := databaseID.String()
	if err := client.Get("/cloud/project/"+service+"/database/"+id, &db); err != nil {
		fmt.Printf("Error getting database for %s in sl %s: %q\n", databaseID, service, err)
		return nil
	}
	return &db
}
