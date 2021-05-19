package ermis

import (
	"database/sql"

	"gorm.io/gorm"
)

type (
	//Alias structure is a model for describing the alias
	Alias struct { //DB definitions    --Validation-->               //Validation
		ID               int          `  gorm:"auto_increment;primaryKey"             valid:"optional, int"`
		AliasName        string       `  gorm:"not null;type:varchar(40);unique"      valid:"required,dns" `
		BestHosts        int          `  gorm:"type:smallint(6);not null"             valid:"required,best_hosts"`
		External         string       `  gorm:"type:varchar(15);not null"             valid:"required,in(yes|no|external|internal)"`
		Metric           string       `  gorm:"type:varchar(15);not null"             valid:"in(cmsfrontier),optional"`
		PollingInterval  int          `  gorm:"type:smallint(6);not null"             valid:"optional,int"`
		Statistics       string       `  gorm:"type:varchar(15);not null"             valid:"optional,alpha"`
		Clusters         string       `  gorm:"type:longtext;not null"                valid:"optional,alphanum"`
		Tenant           string       `  gorm:"type:longtext;not null"                valid:"optional,alphanum" `
		Hostgroup        string       `  gorm:"type:longtext;not null"                valid:"required,hostgroup"`
		User             string       `  gorm:"type:varchar(40);not null"             valid:"optional,alphanum" `
		TTL              int          `  gorm:"type:smallint(6);default:60;not null"  valid:"optional,int"`
		LastModification sql.NullTime `  gorm:"type:date"                             valid:"-"`
		Cnames           []Cname      `  gorm:"foreignkey:CnameAliasID"               valid:"optional"`
		Relations        []Relation   `                                               valid:"optional"`
		Alarms           []Alarm      `  gorm:"foreignkey:AlarmAliasID"               valid:"optional" `
	}

	/*For future reference:
	1.The many-to-many relation is not implemented
	  in the default way,as in the gorm docs. The reason for that is the need for an
	  extra column in the relations table
	2.ID fields are validated as OPTIONAL int, because
	marking them as required fields will cause validation failures for ID=0.
	New entries have initially ID=0 and after created in DB they are assigned a proper value*/

	//Relation describes the many-to-many relation between nodes/aliases
	Relation struct {
		ID        int          `  gorm:"not null;auto_increment"  valid:"optional, int" `
		Node      *Node        `                                  valid:"required"`
		NodeID    int          ` gorm:"not null"                  valid:"optional, int"`
		Alias     *Alias       `                                  valid:"optional"`
		AliasID   int          ` gorm:"not null"                  valid:"optional,int"`
		Blacklist bool         ` gorm:"not null"                  valid:"-"`
		Load      int          `                                  valid:"optional, int"`
		LastCheck sql.NullTime `valid:"-"`
	}
	//Alarm describes the one to many relation between an alias and its alarms
	Alarm struct {
		ID           int          `  gorm:"auto_increment;primaryKey"   valid:"optional, int"`
		AlarmAliasID int          `  gorm:"not null"                    valid:"optional,int"`
		Alias        string       `  gorm:"type:varchar(40);not null"   valid:"required, dns" `
		Name         string       `  gorm:"type:varchar(20);not null"   valid:"required, in(minimum)"`
		Recipient    string       `  gorm:"type:varchar(40);not null"   valid:"required, email"`
		Parameter    int          `  gorm:"type:smallint(6);not null"   valid:"required, range(0|1000)"`
		Active       bool         `  gorm:"not null"                    valid:"-"`
		LastCheck    sql.NullTime `  gorm:"type:date"                   valid:"-"`
		LastActive   sql.NullTime `  gorm:"type:date"                   valid:"-"`
	}

	//Cname structure is a model for the cname description
	Cname struct {
		ID           int    `  gorm:"auto_increment;primaryKey"         valid:"optional,int"`
		CnameAliasID int    `  gorm:"not null"                          valid:"optional,int"`
		Cname        string `  gorm:"type:varchar(40);not null;unique"  valid:"required, cnames" `
	}

	//Node structure defines the model for the nodes params Node struct {
	Node struct {
		ID               int          `  gorm:"unique;not null;auto_increment;primaryKey"     valid:"optional,int" `
		NodeName         string       `  gorm:"not null;type:varchar(40);unique"              valid:"required, nodes"`
		LastModification sql.NullTime `                                                       valid:"-"`
		Aliases          []Relation   `                                                       valid:"optional"`
	}

	//dBFunc type which accept *gorm.DB and return error, used for transactions
	dBFunc func(tx *gorm.DB) error
)
