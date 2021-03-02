package api

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"gitlab.cern.ch/lb-experts/goermis/aiermis/common"
	"gitlab.cern.ch/lb-experts/goermis/aiermis/orm"
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

//Resource deals with the output from the queries
type (
	Resource struct {
		ID               int       `form:"alias_id"          json:"alias_id"          valid:"-"`
		AliasName        string    `form:"alias_name"        json:"alias_name"        valid:"required,dns"`
		Behaviour        string    `form:"behaviour"         json:"behaviour"         valid:"-"`
		BestHosts        int       `form:"best_hosts"        json:"best_hosts"        valid:"required,int,best_hosts"`
		Clusters         string    `form:"clusters"          json:"clusters"          valid:"alphanum"`
		ForbiddenNodes   string    `form:"ForbiddenNodes"    json:"ForbiddenNodes"    valid:"optional,nodes"   gorm:"not null"`
		AllowedNodes     string    `form:"AllowedNodes"      json:"AllowedNodes"      valid:"optional,nodes"   gorm:"not null"`
		Cname            string    `form:"cnames"            json:"cnames"            valid:"optional,cnames"  gorm:"not null"`
		External         string    `form:"external"          json:"external"          valid:"required,in(yes|no)"`
		Hostgroup        string    `form:"hostgroup"         json:"hostgroup"         valid:"required,hostgroup"`
		LastModification time.Time `form:"last_modification" json:"last_modification" valid:"-"`
		Metric           string    `form:"metric"            json:"metric"            valid:"in(cmsfrontier|minino|minimum|),optional"`
		PollingInterval  int       `form:"polling_interval"  json:"polling_interval"  valid:"numeric"`
		Tenant           string    `form:"tenant"            json:"tenant"            valid:"optional,alphanum"`
		TTL              int       `form:"ttl"               json:"ttl,omitempty"     valid:"numeric,optional"`
		User             string    `form:"user"              json:"user"              valid:"optional,alphanum"`
		Statistics       string    `form:"statistics"        json:"statistics"        valid:"alpha"`
		ResourceURI      string    `                         json:"resource_uri"      valid:"-"`
		Pwned            bool      `                         json:"pwned"             valid:"-"`
		Alarms           string    `form:"alarms"              json:"alarms"             valid:"-"`
	}
	//Objects holds multiple result structs
	Objects struct {
		Objects []Resource `json:"objects"`
	}
)

////////////////////////METHODS////////////////////////////////

// A) GET object(s)

//GetObjects return list of aliases if no parameters are passed or a single alias if parameters are given
func GetObjects(param string) (temp []Resource, err error) {
	var (
		query []orm.Alias
	)
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
	return parse(query), nil

}

func parse(queryResults []orm.Alias) (parsed []Resource) {
	for _, element := range queryResults {
		var temp Resource
		//The ones that are the same
		temp.ID = element.ID
		temp.AliasName = element.AliasName
		temp.Behaviour = element.Behaviour
		temp.BestHosts = element.BestHosts
		temp.Clusters = element.Clusters
		temp.Hostgroup = element.Hostgroup
		temp.External = element.External
		temp.LastModification = element.LastModification
		temp.Metric = element.Metric
		temp.PollingInterval = element.PollingInterval
		temp.TTL = element.TTL
		temp.Tenant = element.Tenant
		temp.ResourceURI = "/p/api/v1/alias/" + strconv.Itoa(element.ID)
		temp.User = element.User
		temp.Statistics = element.Statistics

		//The cnames
		if len(element.Cnames) != 0 {
			var tmpslice []string
			for _, v := range element.Cnames {
				tmpslice = append(tmpslice, v.Cname)
			}
			temp.Cname = strings.Join(tmpslice, ",")
		}

		//The alarms
		if len(element.Alarms) != 0 {
			var tmpslice []string
			for _, v := range element.Alarms {
				alarm := v.Name + ":" +
					v.Recipient + ":" +
					strconv.Itoa(v.Parameter) + ":" +
					strconv.FormatBool(v.Active)
				if v.LastActive.Valid {
					alarm += ":" + v.LastActive.Time.String()
				}
				tmpslice = append(tmpslice, alarm)
			}
			temp.Alarms = strings.Join(tmpslice, ",")
		}

		//The nodes
		if len(element.Nodes) != 0 {
			var tmpallowed []string
			var tmpforbidden []string
			for _, v := range element.Nodes {
				if v.Blacklist == true {
					tmpforbidden = append(tmpforbidden, v.Node.NodeName)
				} else {
					tmpallowed = append(tmpallowed, v.Node.NodeName)
				}

			}
			temp.AllowedNodes = strings.Join(tmpallowed, ",")
			temp.ForbiddenNodes = strings.Join(tmpforbidden, ",")

		}

		//Set the pwn value(true/false)
		//Sole purpose of pwned field is to be used in the UI for alias filtering
		//Ermis-lbaas-admins are superusers
		if IsSuperuser() {
			temp.Pwned = true
		} else {
			temp.Pwned = common.StringInSlice(temp.Hostgroup, GetUsersHostgroups())
		}

		parsed = append(parsed, temp)
	}
	return parsed
}

// B) Create single object

//CreateObject creates an alias
func (r Resource) CreateObject() (err error) {
	//DB//

	//We use ORM struct here, so that we are able to create the relations
	var a orm.Alias
	/*Copier fills Alias struct with the values from Resource struct
	This is possible because during creation there are no complex relations present
	(alarms/nodes are in the modification handler) */
	copier.Copy(&a, &r)

	//Cnames are treated seperately, because they will be created using their struct
	cnames := common.DeleteEmpty(strings.Split(r.Cname, ","))

	//Create object in the DB with transactions, if smth goes wrong its rolledback
	if err := orm.CreateTransactions(a, cnames); err != nil {
		return err
	}

	//DNS
	//createInDNS will create the alias and cnames in DNS.//
	if err := r.createInDNS(); err != nil {

		//If it fails to create alias in DNS, we delete from DB what we created in the previous step.
		//The r struct has ID=0, because ID is assigned after creation
		//For that reason, we retrieve the object from DB for deletion
		alias, _ := GetObjects(r.AliasName)
		alias[0].DeleteObject()
		return err
	}

	return nil

}

//DefaultAndHydrate prepares the object with default values and domain before CREATE
func (r *Resource) defaultAndHydrate() {
	//Populate the struct with the default values
	r.Behaviour = "mindless"
	r.Metric = "cmsfrontier"
	r.PollingInterval = 300
	r.Statistics = "long"
	r.Clusters = "none"
	r.Tenant = "golang"
	r.TTL = 60
	r.LastModification = time.Now()
	//Hydrate
	if !strings.HasSuffix(r.AliasName, ".cern.ch") {
		r.AliasName = r.AliasName + ".cern.ch"
	}
	if common.StringInSlice(strings.ToLower(r.External), []string{"yes", "external"}) {
		r.External = "yes"
	}
	if common.StringInSlice(strings.ToLower(r.External), []string{"no", "internal"}) {
		r.External = "no"
	}
}

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
	  where value indicates privilege allowed/forbidden)*/

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
