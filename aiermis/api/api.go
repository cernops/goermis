package api

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jinzhu/gorm"

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

	//Objects holds multiple result structs
	Objects struct {
		List []Alias `json:"objects"`
	}
)

////////////////////////METHODS////////////////////////////////

// A) GET object(s)

//GetObjects return list of aliases if no parameters are passed or a single alias if parameters are given
func GetObjects(param string) (query []Alias, err error) {

	//Preload bottom-to-top, starting with the Relations & Nodes first
	nodes := con.Preload("Nodes")       //Relations
	nodes = nodes.Preload("Nodes.Node") //From the relations, we find the node names then
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

//CreateObject creates an alias
func (r Alias) CreateObject() (err error) {

	//Create object in the DB with transactions, if smth goes wrong its rolledback
	if err := CreateTransactions(r); err != nil {
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

/*
// C) DELETE single object

//DeleteObject deletes an alias and its Relations
func (r Resource) DeleteObject() (err error) {
	//Delete from DB
	if err := orm.DeleteTransactions(r.AliasName, r.ID); err != nil {
		return err
	}

	//Now delete from DNS.
	if err := r.deleteFromDNS(); err != nil {
		//If deletion from DNS fails, we recreate the object.
		//It will be recreated in DB, but not DNS because it already exists there.
		r.CreateObject()
		return err
	}

	return nil

}

// D) MODIFY single object

//ModifyObject modifies aliases and its associations
func (r Resource) ModifyObject() (err error) {

	//First, lets get once more the old values.We need the cnames and nodes for comparison
	oldObject, _ := GetObjects(r.AliasName)

	//Let's update in DB the single-valued fields that do not require iteration/comparisson
	if err = con.Model(&orm.Alias{}).Where("id = ?", r.ID).UpdateColumns(
		map[string]interface{}{
			"external":          r.External,
			"hostgroup":         r.Hostgroup,
			"best_hosts":        r.BestHosts,
			"metric":            r.Metric,
			"polling_interval":  r.PollingInterval,
			"ttl":               r.TTL,
			"tenant":            r.Tenant,
			"last_modification": time.Now(),
		}).Error; err != nil {
		return errors.New("Failed to update the single-valued fields with error: " + err.Error())

	}

	//1.Update cnames in DB
	if err = r.updateCnames(oldObject[0]); err != nil {
		return err
	}
	/*2.Update nodes for r object with new nodes(nodesInMap converts string to map,
	  where value indicates privilege allowed/forbidden)*/ /*

	newNodesMap := nodesInMap(r.AllowedNodes, r.ForbiddenNodes)
	oldNodesMap := nodesInMap(oldObject[0].AllowedNodes, oldObject[0].ForbiddenNodes)

	if err = r.updateNodes(newNodesMap, oldNodesMap); err != nil {
		return err
	}

	//3.Update DNS
	if err = r.UpdateDNS(oldObject[0]); err != nil {
		//If something goes wrong while updating, then we use the object
		//we had in DB before the update to restore that state, before the error
		r.DeleteObject()            //Delete the DB updates we just made and existing DNS entries
		oldObject[0].CreateObject() //Recreate the alias as it was before the update
		return err
	}

	//4.Update alarms
	if err = r.updateAlarms(oldObject[0]); err != nil {
		return err
	}

	return nil
}

/////////// Logical sub-functions of UPDATE///////////

//UpdateNodes updates alias with new nodes in DB
func (r Resource) updateNodes(new map[string]bool, old map[string]bool) (err error) {
	for name := range old {
		if _, ok := new[name]; !ok {
			if err = orm.DeleteNodeTransactions(r.ID, name); err != nil {
				return errors.New("Failed to delete existing node " +
					name + " while updating, with error: " + err.Error())
			}
		}
	}
	for name, privilege := range new {
		if name == "" {
			continue
		}
		if _, ok := old[name]; !ok {
			if err = orm.AddNodeTransactions(r.ID, name, privilege); err != nil {
				return errors.New("Failed to add new node " +
					name + " while updating, with error: " + err.Error())
			}
		} else if value, ok := old[name]; ok && value != privilege {
			if err = orm.UpdatePrivilegeTransactions(r.ID, name, privilege); err != nil {
				return errors.New("Failed to update privilege for node " +
					name + " while updating, with error: " + err.Error())
			}
		}
	}

	return nil

}

//UpdateCnames updates cnames in DB
func (r Resource) updateCnames(oldObject Resource) (err error) {

	//Split string and delete any possible empty values

	newCnames := common.DeleteEmpty(strings.Split(r.Cname, ","))
	exCnames := common.DeleteEmpty(strings.Split(oldObject.Cname, ","))
	if len(newCnames) > 0 {
		for _, value := range exCnames {
			if !common.StringInSlice(value, newCnames) {
				if err = orm.DeleteCnameTransactions(r.ID, value); err != nil {
					return errors.New("Failed to delete existing cname " +
						value + " while updating, with error: " + err.Error())
				}
			}
		}

		for _, value := range newCnames {
			if value == "" {
				continue
			}
			if !common.StringInSlice(value, exCnames) {
				if err = orm.AddCnameTransactions(r.ID, value); err != nil {
					return errors.New("Failed to add new cname " +
						value + " while updating, with error: " + err.Error())
				}
			}

		}

	} else {
		for _, value := range exCnames {
			if err = orm.DeleteCnameTransactions(r.ID, value); err != nil {
				return errors.New("Failed to delete cname " +
					value + " while purging all, with error: " + err.Error())
			}
		}
	}
	return nil
}

func (r Resource) updateAlarms(oldObject Resource) (err error) {
	//Split string and delete any possible empty values

	newAlarms := common.DeleteEmpty(strings.Split(r.Alarms, ","))
	exAlarms := common.DeleteEmpty(strings.Split(oldObject.Alarms, ","))
	if len(newAlarms) > 0 {
		for _, value := range exAlarms {
			if !common.StringInSlice(value, newAlarms) {
				if err = orm.DeleteAlarmTransactions(r.ID, value); err != nil {
					return errors.New("Failed to delete existing alarm " +
						value + " while updating, with error: " + err.Error())
				}
			}
		}

		for _, value := range newAlarms {
			if value == "" {
				continue
			}
			if !common.StringInSlice(value, exAlarms) {
				if err = orm.AddAlarmTransactions(r.ID, r.AliasName, value); err != nil {
					return errors.New("Failed to add new alarm " +
						value + " while updating, with error: " + err.Error())
				}
			}

		}

	} else {
		for _, value := range exAlarms {
			if err = orm.DeleteAlarmTransactions(r.ID, value); err != nil {
				return errors.New("Failed to delete alarm " +
					value + " while purging all, with error: " + err.Error())
			}
		}
	}
	return nil
}

//nodesInMap puts the nodes in a map. The value is their privilege
func nodesInMap(AllowedNodes interface{}, ForbiddenNodes interface{}) map[string]bool {
	if AllowedNodes == nil {
		AllowedNodes = ""
	}
	if ForbiddenNodes == nil {
		ForbiddenNodes = ""
	}

	temp := make(map[string]bool)

	modes := map[interface{}]bool{
		AllowedNodes:   false,
		ForbiddenNodes: true,
	}
	for k, v := range modes {
		if k != "" {
			for _, val := range common.DeleteEmpty(strings.Split(fmt.Sprintf("%v", k), ",")) {
				temp[val] = v
			}
		}
	}

	return temp
}
*/
