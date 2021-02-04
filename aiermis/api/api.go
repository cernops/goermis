package api

import (
	"database/sql"
	"errors"
	"time"

	"github.com/davecgh/go-spew/spew"
	"gorm.io/gorm"

	"gitlab.cern.ch/lb-experts/goermis/bootstrap"
	"gitlab.cern.ch/lb-experts/goermis/db"
)

var (
	con = db.ManagerDB()
	q   string
	cfg = bootstrap.GetConf()
)

/*form tags --> for binding the form fields,
valid tag--> validation rules, extra funcs in the common.go file*/

type (
	//Alias structure is a model for describing the alias
	Alias struct {
		ID               int       `  gorm:"auto_increment;primaryKey" `
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
		Relations        []*Relation
		Alarms           []Alarm `  gorm:"foreignkey:AlarmAliasID" `
	}

	/*For future references, the many-to-many relation is not implemented
	  in the default way,as in the gorm docs. The reason for that is the need for an
	  extra column in the relations table*/

	//Relation describes the many-to-many relation between nodes/aliases
	Relation struct {
		ID        int `  gorm:"not null;auto_increment" `
		Node      *Node
		NodeID    int ` gorm:"not null"`
		Alias     *Alias
		AliasID   int  ` gorm:"not null"`
		Blacklist bool ` gorm:"not null"`
	}
	//Alarm describes the one to many relation between an alias and its alarms
	Alarm struct {
		ID           int          `  gorm:"auto_increment;primaryKey" `
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
		ID           int    `  gorm:"auto_increment;primaryKey" `
		CnameAliasID int    `  gorm:"not null" `
		Cname        string `  gorm:"type:varchar(40);not null;unique" `
	}

	//Node structure defines the model for the nodes params Node struct {
	Node struct {
		ID               int    `  gorm:"unique;not null;auto_increment;primaryKey"`
		NodeName         string `  gorm:"not null;type:varchar(40);unique" `
		LastModification time.Time
		Load             int
		State            string `  gorm:"type:varchar(15);not null" `
		Hostgroup        string `  gorm:"type:varchar(40);not null" `
		Aliases          []Relation
	}

	//dBFunc type which accept *gorm.DB and return error, used for transactions
	dBFunc func(tx *gorm.DB) error
)

////////////////////////METHODS////////////////////////////////

// A) GET object(s)

//GetObjects return list of aliases if no parameters are passed or a single alias if parameters are given
func GetObjects(param string) (query []Alias, err error) {

	//Preload bottom-to-top, starting with the Relations & Nodes first
	nodes := con.Preload("Relations")       //Relations
	nodes = nodes.Preload("Relations.Node") //From the relations, we find the node names then
	if param == "all" {
		err = nodes.
			Preload("Cnames").
			Preload("Alarms").
			Order("alias_name").
			Find(&query).Error

	} else {
		err = nodes.
			Preload("Cnames").
			Preload("Alarms").
			Where("id=?", param).Or("alias_name=?", param).
			Order("alias_name").
			First(&query).Error

	}
	if err != nil {
		return nil, errors.New("Failed in query: " + err.Error())

	}
	return query, nil

}

// B) Create single object

//CreateObjectInDB creates an alias
func (alias Alias) createObjectInDB() (err error) {

	//Create object in the DB with transactions, if smth goes wrong its rolledback
	if err := CreateTransactions(alias); err != nil {
		return err
	}

	//DNS
	//createInDNS will create the alias and cnames in DNS.//
	//if err := r.createInDNS(); err != nil {

	//If it fails to create alias in DNS, we delete from DB what we created in the previous step.
	//The r struct has ID=0, because ID is assigned after creation
	//For that reason, we retrieve the object from DB for deletion
	//alias, _ := GetObjects(r.AliasName)
	//alias[0].DeleteObject()
	//return err
	//}

	return nil

}

// C) DELETE single object
/*
//DeleteObject deletes an alias and its Relations
func (r Resource) DeleteObject() (err error) {
	//Delete from DB
	if err := orm.DeleteTransactions(r.AliasName, r.ID); err != nil {
		return err
	}
	/*
		//Now delete from DNS.
		if err := r.deleteFromDNS(); err != nil {
			//If deletion from DNS fails, we recreate the object.
			//It will be recreated in DB, but not DNS because it already exists there.
			r.CreateObject()
			return err
		}
	/*
	//return nil

}
*/
// D) MODIFY single object

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
		relationsInDB []*Relation
	)
	//Let's find the registered nodes for this alias

	con.Preload("Node").Where("alias_id=?", alias.ID).Find(&relationsInDB)

	for _, r := range relationsInDB {
		if ok, _ := containsNode(alias.Relations, r); !ok {
			if err = deleteNodeTransactions(r); err != nil {
				return errors.New("Failed to delete existing node " +
					r.Node.NodeName + " while updating, with error: " + err.Error())
			}
		}
	}
	for _, r := range alias.Relations {
		spew.Dump(r)
		if ok, _ := containsNode(relationsInDB, r); !ok {
			if err = addNodeTransactions(r); err != nil {
				return errors.New("Failed to add new node " +
					r.Node.NodeName + " while updating, with error: " + err.Error())
			}
		} else if ok, privilege := containsNode(relationsInDB, r); ok && !privilege {
			if err = updatePrivilegeTransactions(r); err != nil {
				return errors.New("Failed to update privilege for node " +
					r.Node.NodeName + " while updating, with error: " + err.Error())
			}
		}
	}

	return nil

}

//updateCnames updates cnames in DB
func (alias Alias) updateCnames() (err error) {
	var (
		cnamesInDB []Cname
	)
	//Let's see what cnames are already registered for this alias
	con.Model(&alias).Association("Cnames").Find(&cnamesInDB)

	if len(alias.Cnames) > 0 {
		for _, v := range cnamesInDB {
			if !containsCname(alias.Cnames, v.Cname) {
				if err = deleteCnameTransactions(v); err != nil {
					return errors.New("Failed to delete existing cname " +
						v.Cname + " while updating, with error: " + err.Error())
				}
			}
		}

		for _, v := range alias.Cnames {
			if !containsCname(cnamesInDB, v.Cname) {
				if err = addCnameTransactions(v); err != nil {
					return errors.New("Failed to add new cname " +
						v.Cname + " while updating, with error: " + err.Error())
				}
			}

		}

	} else {
		for _, v := range cnamesInDB {
			if err = deleteCnameTransactions(v); err != nil {
				return errors.New("Failed to delete cname " +
					v.Cname + " while purging all, with error: " + err.Error())
			}
		}
	}
	return nil
}

func (alias Alias) updateAlarms() (err error) {
	var (
		alarmsInDB []Alarm
	)
	//Let's see what alarms are already registered for this alias
	con.Model(&alias).Association("Alarms").Find(&alarmsInDB)
	//Split string and delete any possible empty values
	if len(alias.Alarms) > 0 {
		for _, a := range alarmsInDB {
			if !containsAlarm(alias.Alarms, a) {
				if err = deleteAlarmTransactions(a); err != nil {
					return errors.New("Failed to delete existing alarm " +
						a.Name + " while updating, with error: " + err.Error())
				}
			}
		}

		for _, a := range alias.Alarms {
			if !containsAlarm(alarmsInDB, a) {
				if err = addAlarmTransactions(a); err != nil {
					return errors.New("Failed to add alarm " +
						a.Name + ":" +
						a.Recipient + ":" +
						string(a.Parameter) +
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
					string(a.Parameter) +
					" while purging all, with error: " +
					err.Error())
			}
		}
	}
	return nil
}

/*
	//3.Update DNS
	if err = r.UpdateDNS(oldObject[0]); err != nil {
		//If something goes wrong while updating, then we use the object
		//we had in DB before the update to restore that state, before the error
		r.DeleteObject()            //Delete the DB updates we just made and existing DNS entries
		oldObject[0].CreateObject() //Recreate the alias as it was before the update
		return err
	}
*/

func containsCname(s []Cname, e string) bool {
	for _, a := range s {
		if a.Cname == e {
			return true
		}
	}
	return false

}
func containsAlarm(s []Alarm, a Alarm) bool {
	for _, alarm := range s {
		if alarm.Name == a.Name &&
			alarm.Recipient == a.Recipient &&
			alarm.Parameter == a.Parameter {
			return true
		}
	}
	return false

}

//containsNode checks if a node has a relation with an alias and the status of that relation(allowed or forbidden)
func containsNode(a []*Relation, b *Relation) (bool, bool) {
	for _, v := range a {
		if v.Node.NodeName == b.Node.NodeName {
			if v.Blacklist == b.Blacklist {
				return true, true
			}
			return true, false
		}

	}
	return false, false

}
