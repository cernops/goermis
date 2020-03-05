package models

import (
	"time"
)

//Node structure defines the model for the nodes params
type Node struct {
	ID               uint        `json:"node_id" schema:"node_id" gorm:"unique;not null;auto_increment;primary_key"`
	NodeName         string      `json:"node_name" schema:"node_name" gorm:"not null;size:60;unique"`
	LastModification time.Time   `json:"last_modification" schema:"last_modification"`
	Load             int         `json:"load" schema:"load"`
	State            string      `json:"state" schema:"state" gorm:"not null"`
	Hostgroup        string      `json:"hostgroup" schema:"hostgroup" gorm:"size:40;not null"`
	Relations        []*Relation `json:"relation" schema:"relation"`
}

//Alias structure is a model for describing the alias
type Alias struct {
	ID               uint      `json:"alias_id" schema:"alias_id" gorm:"unique;not null;auto_increment;primary_key"`
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
	Relations        []*Relation
	Cnames           []Cname `json:"cnames"  gorm:"foreignkey:AliasID"`
}

//Relation testing
type Relation struct {
	ID        uint
	Node      *Node
	NodeID    int
	Alias     *Alias
	AliasID   int
	Blacklist bool
}

//Cname structure is a model for the cname description
type Cname struct {
	ID      int    `json:"cname_id" gorm:"auto_increment;primary_key"`
	AliasID uint   `json:"alias_id" gorm:"not null"`
	CName   string `json:"cname" schema:"cnames" gorm:"not null;unique" `
}
