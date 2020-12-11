package orm

import (
	"database/sql"
	"time"

	"github.com/jinzhu/gorm"
)

/*This is a the model of the DB relations. It is used
  exclusevily by GORM
*/

//Alias structure is a model for describing the alias
type (
	Alias struct {
		ID               int
		AliasName        string    `  gorm:"not null;type:varchar(40);unique" `
		Behaviour        string    `  gorm:"type:varchar(15);not null" `
		BestHosts        int       `  gorm:"type:smallint(6);not null" `
		External         string    `  gorm:"type:varchar(15);not null" `
		Metric           string    `  gorm:"type:varchar(15);not null" `
		PollingInterval  int       `  gorm:"type:smallint(6);not null" `
		Statistics       string    `  gorm:"type:varchar(15);not null" valid:"-"`
		Clusters         string    `  gorm:"type:longtext;not null" `
		Tenant           string    `  gorm:"type:longtext;not null" `
		Hostgroup        string    `  gorm:"type:longtext;not null" `
		User             string    `  gorm:"type:varchar(40);not null" `
		TTL              int       `  gorm:"type:smallint(6);default:60;not null"`
		LastModification time.Time `  gorm:"type:date"`
		Cnames           []Cname   `  gorm:"foreignkey:CnameAliasID" `
		Nodes            []*Relation
		Alarms           []Alarm `  gorm:"foreignkey:AlarmAliasID" `
	}

	/*For future references, the many-to-many relation is not implemented
	  in the default way,as in the gorm docs. The reason for that is the need for an
	  extra column in the relations table*/

	//Relation describes the many-to-many relation between nodes/aliases
	Relation struct {
		ID        int
		Node      *Node
		NodeID    int ` gorm:"not null"`
		Alias     *Alias
		AliasID   int  ` gorm:"not null"`
		Blacklist bool ` gorm:"not null"`
	}
	//Alarm describes the one to many relation between an alias and its alarms
	Alarm struct {
		ID           int          `  gorm:"auto_increment;primary_key" `
		AlarmAliasID int          `  gorm:"not null" `
		Alias        string       `  gorm:"type:varchar(40);not null" `
		Name         string       `  gorm:"type:varchar(20);not null" `
		Recipient    string       `  gorm:"type:varchar(40);not null" `
		Parameter    int          `  gorm:"type:smallint(6);not null" `
		Active       bool         `  gorm:"not null" `
		LastCheck    sql.NullTime `  gorm:"type:date"`
		LastActive   sql.NullTime `  gorm:"type:date"`
	}

	//Cname structure is a model for the cname description
	Cname struct {
		ID           int    `  gorm:"auto_increment;primary_key" `
		CnameAliasID int    `  gorm:"not null" `
		Cname        string `  gorm:"type:varchar(40);not null;unique" `
	}

	//Node structure defines the model for the nodes params Node struct {
	Node struct {
		ID               int       `  gorm:"unique;not null;auto_increment;primary_key"`
		NodeName         string    `  gorm:"not null;type:varchar(40);unique" `
		LastModification time.Time `  gorm:"DEFAULT:current_timestamp"`
		Load             int
		State            string `  gorm:"type:varchar(15);not null" `
		Hostgroup        string `  gorm:"type:varchar(40);not null" `
		Aliases          []*Relation
	}

	//dBFunc type which accept *gorm.DB and return error, used for transactions
	dBFunc func(tx *gorm.DB) error
)
