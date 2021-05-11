package ermis

/* This file includes the ORM models and its methods*/

import (
	"database/sql"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"gitlab.cern.ch/lb-experts/goermis/bootstrap"
	"gitlab.cern.ch/lb-experts/goermis/db"
)

var (
	cfg = bootstrap.GetConf() //Getting an instance of config params
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
		Hostgroup        string       `  gorm:"type:varchar(40);not null"                     valid:"optional, hostgroup"`
		Aliases          []Relation   `                                                       valid:"optional"`
	}

	//dBFunc type which accept *gorm.DB and return error, used for transactions
	dBFunc func(tx *gorm.DB) error
)

//GetObjects return list of aliases if no parameters are passed or a single alias if parameters are given
func GetObjects(param string) (query []Alias, err error) {

	//Preload bottom-to-top, starting with the Relations & Nodes first
	nodes := db.GetConn().Preload("Relations.Node") //Relations
	if param == "all" {                             //get all aliases
		err = nodes.
			Preload("Cnames").
			Preload("Alarms").
			Order("alias_name").
			Find(&query).Error

	} else { //get only the specified one
		err = nodes.
			Preload("Cnames").
			Preload("Alarms").
			Where("id=?", param).Or("alias_name=?", param).
			Order("alias_name").
			Find(&query).Error

	}
	if err != nil {
		return nil, errors.New("Failed in query: " + err.Error())

	}
	return query, nil

}

////////////////////////ALIAS METHODS////////////////////////////////

//CreateObjectInDB creates an alias
func (alias Alias) createObjectInDB() (err error) {

	//Create object in the DB with transactions, if smth goes wrong its rolledback
	if err := CreateTransactions(alias); err != nil {
		return err
	}

	return nil

}

//deleteObject deletes an alias and its Relations
func (alias Alias) deleteObjectInDB() (err error) {
	//Delete from DB
	if err := deleteTransactions(alias); err != nil {
		return err
	}
	return nil

}

//UpdateAlias modifies aliases and its associations
func (alias Alias) updateAlias() (err error) {
	if err := aliasUpdateTransactions(alias); err != nil {
		return err
	}

	return nil
}

//updateNodes updates alias with new nodes
func (alias Alias) updateNodes() (err error) {
	var (
		relationsInDB []Relation
		intf          PrivilegeIntf
	)
	//Let's find the registered nodes for this alias
	db.GetConn().Preload("Node").Where("alias_id=?", alias.ID).Find(&relationsInDB)

	for _, r := range relationsInDB {
		intf = r
		if ok, _ := Compare(intf, alias.Relations); !ok {
			if err = deleteNodeTransactions(r); err != nil {
				return errors.New("Failed to delete existing node " +
					r.Node.NodeName + " while updating, with error: " + err.Error())
			}
		}
	}
	for _, r := range alias.Relations {
		intf = r
		if ok, _ := Compare(intf, relationsInDB); !ok {
			if err = AddNodeTransactions(r); err != nil {
				return errors.New("Failed to add new node " +
					r.Node.NodeName + " while updating, with error: " + err.Error())
			}
			//If relation exists we also check if user modified its privileges
		} else if ok, privilege := Compare(intf, relationsInDB); ok && !privilege {
			if err = updatePrivilegeTransactions(r); err != nil {
				return errors.New("Failed to update privilege for node " +
					r.Node.NodeName + " while updating, with error: " + err.Error())
			}
		}
	}

	return nil

}

//Update the cnames
//updateCnames updates cnames in DB
func (alias Alias) updateCnames() (err error) {
	var (
		cnamesInDB []Cname
		intf       ContainsIntf
	)
	//Let's see what cnames are already registered for this alias
	db.GetConn().Model(&alias).Association("Cnames").Find(&cnamesInDB)

	if len(alias.Cnames) > 0 { //there are cnames, delete and add accordingly
		for _, v := range cnamesInDB {
			intf = v
			if !Contains(intf, alias.Cnames) {
				if err = deleteCnameTransactions(v); err != nil {
					return errors.New("Failed to delete existing cname " +
						v.Cname + " while updating, with error: " + err.Error())
				}
			}
		}

		for _, v := range alias.Cnames {
			intf = v
			if !Contains(intf, cnamesInDB) {
				if err = addCnameTransactions(v); err != nil {
					return errors.New("Failed to add new cname " +
						v.Cname + " while updating, with error: " + err.Error())
				}
			}

		}

	} else { //user deleted everything, so do we
		for _, v := range cnamesInDB {
			if err = deleteCnameTransactions(v); err != nil {
				return errors.New("Failed to delete cname " +
					v.Cname + " while purging all, with error: " + err.Error())
			}
		}
	}
	return nil
}

//Update the alarms
func (alias Alias) updateAlarms() (err error) {
	var (
		alarmsInDB []Alarm
		intf       ContainsIntf
	)
	//Let's see what alarms are already registered for this alias
	db.GetConn().Model(&alias).Association("Alarms").Find(&alarmsInDB)
	if len(alias.Alarms) > 0 {
		for _, a := range alarmsInDB {
			intf = a
			if !Contains(intf, alias.Alarms) {
				if err = deleteAlarmTransactions(a); err != nil {
					return errors.New("Failed to delete existing alarm " +
						a.Name + " while updating, with error: " + err.Error())
				}
			}
		}

		for _, a := range alias.Alarms {
			intf = a
			if !Contains(intf, alarmsInDB) {
				if err = addAlarmTransactions(a); err != nil {
					return errors.New("Failed to add alarm " +
						a.Name + ":" +
						a.Recipient + ":" +
						fmt.Sprint(a.Parameter) +
						" while purging all, with error: " +
						err.Error())
				}
			}

		}

	} else {
		for _, a := range alarmsInDB {
			if err = deleteAlarmTransactions(a); err != nil {
				return errors.New("Failed to delete alarm " +
					a.Name + ":" +
					a.Recipient + ":" +
					fmt.Sprint(a.Parameter) +
					" while purging all, with error: " +
					err.Error())
			}
		}
	}
	return nil
}
/*
func (alias Alias) createSecret() error {
	newsecret := generateRandomSecret()
	err := auth.PostSecret(alias.AliasName, newsecret)
	if err != nil {
		return err
	}
	return alias.sendSecretToUser(newsecret)

}
func (alias Alias) deleteSecret() error {
	return auth.DeleteSecret(alias.AliasName)
}

//SendNotification sends an e-mail to the recipient when alarm is triggered
func (alias Alias) sendSecretToUser(secret string) error {
	recipient := alias.User + "@cern.ch"
	log.Infof("Sending the new secret of alias %v to %v", alias.AliasName, alias.User)
	msg := []byte("To: " + alias.User + "\r\n" +
		fmt.Sprintf("Subject: New secret created for alias %s: Please provide this to the nodes behind that alias. If not sure, check documentation(https://configdocs.web.cern.ch)\nSecret: %s ", alias.AliasName, secret))

	err := smtp.SendMail("localhost:25",
		nil,
		"lbd@cern.ch",
		[]string{recipient},
		msg)
	return err
}
*/