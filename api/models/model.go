package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

//Alias structure is a model for describing the alias
type (
	Alias struct {
		ID               int       `json:"alias_id" schema:"alias_id" `
		AliasName        string    `json:"alias_name" schema:"alias_name" gorm:"not null;size:40;unique" `
		Behaviour        string    `json:"behaviour" schema:"behaviour" gorm:"not null" `
		BestHosts        int       `json:"best_hosts" schema:"best_hosts" gorm:"not null" `
		External         string    `json:"external" schema:"external" gorm:"not null" `
		Metric           string    `json:"metric" schema:"metric" gorm:"not null" `
		PollingInterval  int       `json:"polling_interval" schema:"polling_interval" gorm:"not null" `
		Statistics       string    `json:"statistics" schema:"statistics" gorm:"not null" valid:"-"`
		Clusters         string    `json:"clusters" schema:"clusters" gorm:"not null" `
		Tenant           string    `json:"tenant" schema:"tenant"  gorm:"not null" `
		Hostgroup        string    `json:"hostgroup" schema:"hostgroup" gorm:"not null" `
		User             string    `json:"user" schema:"user" gorm:"not null" `
		TTL              int       `json:"ttl" schema:"ttl" `
		LastModification time.Time `json:"last_modification" schema:"last_modification"`
		Cnames           []Cname   `json:"cnames"  gorm:"foreignkey:AliasID" `
		Nodes            []*AliasesNodes
	}

	//AliasesNodes testing
	AliasesNodes struct {
		ID        int
		Node      *Node
		NodeID    int `gorm:"not null"`
		Alias     *Alias
		AliasID   int `gorm:"not null"`
		Blacklist bool
	}

	//Cname structure is a model for the cname description
	Cname struct {
		ID      int    `json:"cname_id" gorm:"auto_increment;primary_key" `
		AliasID int    `json:"alias_id" gorm:"not null" `
		CName   string `json:"cname" gorm:"not null;unique" `
	}

	//Node structure defines the model for the nodes params Node struct {
	Node struct {
		ID               int       `json:"node_id"  gorm:"unique;not null;auto_increment;primary_key"`
		NodeName         string    `json:"node_name"  gorm:"not null;size:60;unique" `
		LastModification time.Time `json:"last_modification" `
		Load             int       `json:"load" `
		State            string    `json:"state"  gorm:"not null" `
		Hostgroup        string    `json:"hostgroup"  gorm:"size:40;not null" `
		Aliases          []*AliasesNodes
	}
	//DBFunc type which accept *gorm.DB and return error
	DBFunc func(tx *gorm.DB) error
	//Resource deals with the output from the queries
	Resource struct {
		ID               int       `json:"alias_id" valid:"required,numeric"`
		AliasName        string    `json:"alias_name" schema:"alias_name" valid:"required,dns"`
		Behaviour        string    `json:"behaviour" schema:"behaviour" valid:"-"`
		BestHosts        int       `json:"best_hosts" schema:"best_hosts" valid:"required,int,best_hosts"`
		Clusters         string    `json:"clusters" schema:"clusters" valid:"alphanum"`
		ForbiddenNodes   string    `json:"ForbiddenNodes"  schema:"ForbiddenNodes"  gorm:"not null" valid:"optional,nodes" `
		AllowedNodes     string    `json:"AllowedNodes" schema:"AllowedNodes" gorm:"not null" valid:"optional,nodes"`
		Cname            string    `json:"cnames"  schema:"cnames" gorm:"not null" valid:"optional,cnames"`
		External         string    `json:"external" schema:"external" valid:"required,in(yes|no)"`
		Hostgroup        string    `json:"hostgroup" schema:"hostgroup" valid:"required,alphanum"`
		LastModification time.Time `json:"last_modification" schema:"last_modification" valid:"-"`
		Metric           string    `json:"metric" schema:"metric" valid:"metric,optional"`
		PollingInterval  int       `json:"polling_interval" schema:"polling_interval" valid:"numeric"`
		Tenant           string    `json:"tenant" schema:"tenant" valid:"optional,alphanum"`
		TTL              int       `json:"ttl" schema:"ttl" valid:"numeric"`
		User             string    `json:"user" schema:"user" valid:"optional,alphanum"`
		Statistics       string    `json:"statistics" schema:"statistics" valid:"alpha"`
	}
	//Objects holds multiple result structs
	Objects struct {
		Objects []Resource `json:"objects"`
	}
)
