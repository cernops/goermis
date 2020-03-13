package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

//Alias structure is a model for describing the alias
type (
	Alias struct {
		ID               int       `json:"alias_id" schema:"alias_id" gorm:"unique;not null;auto_increment;primary_key"`
		AliasName        string    `json:"alias_name" schema:"alias_name" gorm:"not null;size:40;unique"`
		Behaviour        string    `json:"behaviour" schema:"behaviour" gorm:"not null"`
		BestHosts        int       `json:"best_hosts" schema:"best_hosts" gorm:"not null"`
		External         string    `json:"external" schema:"external" gorm:"not null" `
		Metric           string    `json:"metric" schema:"metric" gorm:"not null"`
		PollingInterval  int       `json:"polling_interval" schema:"polling_interval" gorm:"not null" `
		Statistics       string    `json:"statistics" schema:"statistics" gorm:"not null"`
		Clusters         string    `json:"clusters" schema:"clusters" gorm:"not null" `
		Tenant           string    `json:"tenant" schema:"tenant"  gorm:"not null"`
		Hostgroup        string    `json:"hostgroup" schema:"hostgroup" gorm:"not null" `
		User             string    `json:"user" schema:"user" gorm:"not null" `
		TTL              int       `json:"ttl" schema:"ttl"`
		LastModification time.Time `json:"last_modification" schema:"last_modification" `
		Nodes            []Node    `json:"nodes" gorm:"many2many:aliases_nodes;"`
		Cnames           []Cname   `json:"cnames"  gorm:"foreignkey:AliasID"`
	}

	//Relation testing
	Relation struct {
		ID        int
		Node      *Node
		NodeID    int
		Alias     *Alias
		AliasID   int
		Blacklist bool
	}

	//Cname structure is a model for the cname description
	Cname struct {
		ID      int    `json:"cname_id" gorm:"auto_increment;primary_key"`
		AliasID int    `json:"alias_id" gorm:"not null"`
		CName   string `json:"cname" gorm:"not null;unique" `
	}

	//Node structure defines the model for the nodes params Node struct {
	Node struct {
		ID               int       `json:"node_id"  gorm:"unique;not null;auto_increment;primary_key"`
		NodeName         string    `json:"node_name"  gorm:"not null;size:60;unique"`
		LastModification time.Time `json:"last_modification" `
		Load             int       `json:"load" `
		State            string    `json:"state"  gorm:"not null"`
		Hostgroup        string    `json:"hostgroup"  gorm:"size:40;not null"`
		Aliases          []Alias   `json:"aliases"  gorm:"many2many:aliases_nodes;" `
	}
	//DBFunc type which accept *gorm.DB and return error
	DBFunc func(tx *gorm.DB) error
	//Resource deals with the output from the queries
	Resource struct {
		ID               int       `json:"alias_id"`
		AliasName        string    `json:"alias_name" schema:"alias_name"`
		Behaviour        string    `json:"behaviour" schema:"behaviour" `
		BestHosts        int       `json:"best_hosts" schema:"best_hosts"`
		Clusters         string    `json:"clusters" schema:"clusters"`
		ForbiddenNodes   string    `json:"ForbiddenNodes"  schema:"ForbiddenNodes"  gorm:"not null"`
		AllowedNodes     string    `json:"AllowedNodes" schema:"AllowedNodes" gorm:"not null"`
		Cname            string    `json:"cnames"  schema:"cnames" gorm:"not null"`
		External         string    `json:"external" schema:"external"`
		Hostgroup        string    `json:"hostgroup" schema:"hostgroup"`
		LastModification time.Time `json:"last_modification" schema:"last_modification"`
		Metric           string    `json:"metric" schema:"metric"`
		PollingInterval  int       `json:"polling_interval" schema:"polling_interval"`
		Tenant           string    `json:"tenant" schema:"tenant"`
		TTL              int       `json:"ttl" schema:"ttl"`
		User             string    `json:"user" schema:"user"`
		Statistics       string    `json:"statistics" schema:"statistics"`
	}
	//Objects holds multiple result structs
	Objects struct {
		Objects []Resource `json:"objects"`
	}
)
