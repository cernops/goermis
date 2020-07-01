package orm

import (
	"time"

	"github.com/jinzhu/gorm"
)

//This is a the model of the DB relations. It is used
//exclusevily by GORM

//Alias structure is a model for describing the alias
type (
	Alias struct {
		ID               int
		AliasName        string ` gorm:"not null;size:40;unique" `
		Behaviour        string ` gorm:"not null" `
		BestHosts        int    ` gorm:"not null" `
		External         string ` gorm:"not null" `
		Metric           string ` gorm:"not null" `
		PollingInterval  int    ` gorm:"not null" `
		Statistics       string ` gorm:"not null" valid:"-"`
		Clusters         string ` gorm:"not null" `
		Tenant           string `  gorm:"not null" `
		Hostgroup        string ` gorm:"not null" `
		User             string ` gorm:"not null" `
		TTL              int
		LastModification time.Time
		Cnames           []Cname ` gorm:"foreignkey:AliasID" `
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
		ID      int    ` gorm:"auto_increment;primary_key" `
		AliasID int    ` gorm:"not null" `
		CName   string ` gorm:"not null;unique" `
	}

	//Node structure defines the model for the nodes params Node struct {
	Node struct {
		ID               int    ` gorm:"unique;not null;auto_increment;primary_key"`
		NodeName         string `  gorm:"not null;size:60;unique" `
		LastModification time.Time
		Load             int
		State            string `  gorm:"not null" `
		Hostgroup        string `gorm:"size:40;not null" `
		Aliases          []*AliasesNodes
	}

	//dBFunc type which accept *gorm.DB and return error, used for transactions
	dBFunc func(tx *gorm.DB) error
)
