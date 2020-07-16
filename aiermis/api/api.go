package api

import (
	"errors"

	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"github.com/labstack/gommon/log"
	"gitlab.cern.ch/lb-experts/goermis/aiermis/orm"
	"gitlab.cern.ch/lb-experts/goermis/bootstrap"
	"gitlab.cern.ch/lb-experts/goermis/db"
	landbsoap "gitlab.cern.ch/lb-experts/goermis/landb"
)

var (
	con = db.ManagerDB()
	q   string
	cfg = bootstrap.GetConf()
)

//Resource deals with the output from the queries
type (
	Resource struct {
		ID               int       `form:"alias_id" json:"alias_id" valid:"required,numeric"`
		AliasName        string    `form:"alias_name" json:"alias_name"  valid:"required,dns"`
		Behaviour        string    `form:"behaviour" json:"behaviour" valid:"-"`
		BestHosts        int       `form:"best_hosts" json:"best_hosts" valid:"required,int,best_hosts"`
		Clusters         string    `form:"clusters" json:"clusters"  valid:"alphanum"`
		ForbiddenNodes   string    `form:"ForbiddenNodes" json:"ForbiddenNodes" gorm:"not null" valid:"optional,nodes" `
		AllowedNodes     string    `form:"AllowedNodes" json:"AllowedNodes"  gorm:"not null" valid:"optional,nodes"`
		Cname            string    `form:"cnames" json:"cnames"   gorm:"not null" valid:"optional,cnames"`
		External         string    `form:"external" json:"external"  valid:"required,in(yes|no|internal|external)"`
		Hostgroup        string    `form:"hostgroup" json:"hostgroup"  valid:"required,hostgroup"`
		LastModification time.Time `form:"last_modification" json:"last_modification"  valid:"-"`
		Metric           string    `form:"metric" json:"metric"  valid:"in(cmsfrontier|minino|minimum|),optional"`
		PollingInterval  int       `form:"polling_interval" json:"polling_interval" valid:"numeric"`
		Tenant           string    `form:"tenant" json:"tenant"  valid:"optional,alphanum"`
		TTL              int       `form:"ttl" json:"ttl"  valid:"numeric"`
		User             string    `form:"user" json:"user"  valid:"optional,alphanum"`
		Statistics       string    `form:"statistics" json:"statistics"  valid:"alpha"`
		URI              string    `valid:"-"`
	}
	//Objects holds multiple result structs
	Objects struct {
		Objects []Resource `json:"objects"`
	}
)

/////METHODS/////

//GetObjects return list of aliases if no parameters are passed or a single alias if parameters are given
func GetObjects(param string, tablerow string) (b []Resource, err error) {

	if param == "" && tablerow == "" {
		q = "SELECT a.id, alias_name, behaviour, best_hosts, clusters, " +
			"COALESCE(GROUP_CONCAT(distinct case when r.blacklist = 1 then n.node_name else null end),'') AS ForbiddenNodes, " +
			"COALESCE(GROUP_CONCAT(distinct case when r.blacklist = 0 then n.node_name else null end),'') AS AllowedNodes, " +
			"COALESCE(GROUP_CONCAT(distinct c_name),'') AS cname, external,  a.hostgroup, a.last_modification, metric, polling_interval,tenant, ttl, user, statistics " +
			"FROM alias a " +
			"LEFT join cname c on ( a.id=c.alias_id) " +
			"LEFT JOIN aliases_nodes r on (a.id=r.alias_id) " +
			"LEFT JOIN node n on (n.id=r.node_id) " +
			"GROUP BY a.id, alias_name, behaviour, best_hosts, clusters,  external, a.hostgroup, " +
			"a.last_modification, metric, polling_interval, statistics, tenant, ttl, user ORDER BY alias_name"

	} else {
		q = "SELECT a.id, alias_name, behaviour, best_hosts, clusters, " +
			"COALESCE(GROUP_CONCAT(distinct case when r.blacklist = 1 then n.node_name else null end),'') AS ForbiddenNodes," +
			"COALESCE(GROUP_CONCAT(distinct case when r.blacklist = 0 then n.node_name else null end),'') AS AllowedNodes, " +
			"COALESCE(GROUP_CONCAT(distinct c_name),'') AS cname, external,  a.hostgroup, a.last_modification, metric, polling_interval,tenant, ttl, user, statistics " +
			"FROM alias a " +
			"LEFT JOIN cname c on ( a.id=c.alias_id) " +
			"LEFT JOIN aliases_nodes r on (a.id=r.alias_id) " +
			"LEFT JOIN node n on (n.id=r.node_id) " +
			"where a." + tablerow + " = " + "'" + param + "' " +
			"GROUP BY a.id, alias_name, behaviour, best_hosts, clusters,  external, a.hostgroup, " +
			"a.last_modification, metric, polling_interval, statistics, tenant, ttl, user ORDER BY alias_name"
	}

	rows, err := con.Raw(q).Rows()

	if err != nil {
		return b, errors.New("Failed in query: " + err.Error())

	}
	defer rows.Close()
	for rows.Next() {
		var result Resource
		//Fill the struct with the results of the DB query
		err := rows.Scan(&result.ID, &result.AliasName, &result.Behaviour, &result.BestHosts, &result.Clusters,
			&result.ForbiddenNodes, &result.AllowedNodes, &result.Cname, &result.External, &result.Hostgroup,
			&result.LastModification, &result.Metric, &result.PollingInterval,
			&result.Tenant, &result.TTL, &result.User, &result.Statistics)

		//Infer the URI value
		result.URI = "api/v1/" + strconv.Itoa(result.ID)

		if err != nil {
			return b, errors.New("Failed in scanning query results with err: " +
				err.Error())
		}
		b = append(b, result)
	}
	return b, nil
}

//CreateObject creates an alias
func (r Resource) CreateObject() (err error) {
	//DB//

	//We use ORM struct here, so that we are able to create the relations
	var a orm.Alias

	//Copier fills Alias struct with the values from Resource struct
	copier.Copy(&a, &r)
	//Cnames are treated seperately, because they will be created using their struct
	cnames := deleteEmpty(strings.Split(r.Cname, ","))

	//Create object in the DB with transactions
	if err := orm.CreateTransactions(a, cnames); err != nil {
		return err
	}

	//createInDNS will create the alias and cnames in DNS.//
	if err := r.createInDNS(); err != nil {

		//If it fails to create alias in DNS, we rollback the creation in DB.
		r.rollBack("add")
		return err
	}

	return nil

}

//DefaultAndHydrate prepares the object with default values and domain
func (r *Resource) DefaultAndHydrate() {
	//Populate the struct with the default values
	r.Behaviour = "mindless"
	r.Metric = "cmsfrontier"
	r.PollingInterval = 300
	r.Statistics = "long"
	r.Clusters = "none"
	r.Tenant = "golang"
	r.LastModification = time.Now()
	//Hydrate
	if !strings.HasSuffix(r.AliasName, ".cern.ch") {
		r.AliasName = r.AliasName + ".cern.ch"
	}
	if StringInSlice(strings.ToLower(r.External), []string{"yes", "external"}) {
		r.External = "yes"
	}
	if StringInSlice(strings.ToLower(r.External), []string{"no", "internal"}) {
		r.External = "no"
	}
}

//DeleteObject deletes an alias and its Relations
func (r Resource) DeleteObject() (err error) {
	//Delete from DB
	if err := orm.DeleteTransactions(r.AliasName, r.ID); err != nil {
		return err
	}

	//Now delete from DNS.If something goes wrong then we
	//rollback the freshly created DB entry
	if err := r.deleteFromDNS(); err != nil {
		r.rollBack("delete")
		return err
	}

	return nil

}

//ModifyObject modifies aliases and its associations
func (r Resource) ModifyObject() (err error) {

	//First, lets get once more the old values.We need the cnames and nodes for comparison
	oldObject, _ := GetObjects(r.AliasName, "alias_name")

	//Let's update in DB the single-valued fields that do not require iterations and comparison
	if err = con.Model(&orm.Alias{}).Where("id = ?", r.ID).UpdateColumns(
		map[string]interface{}{
			"external":         r.External,
			"hostgroup":        r.Hostgroup,
			"best_hosts":       r.BestHosts,
			"metric":           r.Metric,
			"polling_interval": r.PollingInterval,
			"ttl":              r.TTL,
			"tenant":           r.Tenant,
		}).Error; err != nil {
		return errors.New("Failed to update the single-valued fields with error: " + err.Error())

	}

	//1.Update cnames
	if err = r.UpdateCnames(oldObject[0]); err != nil {
		return err
	}

	//2.Update DNS only if any of view or cnames has changed
	if r.External != oldObject[0].External || !Equal(r.Cname, oldObject[0].Cname) {
		if err = r.UpdateDNS(oldObject[0]); err != nil {
			return err
		}
	}

	/*3.Update nodes for r object with new nodes(nodesToMap converts string to map,
	  where value indicates privilege allowed/forbidden)*/
	newNodesMap := nodesInMap(r.AllowedNodes, r.ForbiddenNodes)
	oldNodesMap := nodesInMap(oldObject[0].AllowedNodes, oldObject[0].ForbiddenNodes)

	if err = r.UpdateNodes(newNodesMap, oldNodesMap); err != nil {
		return err
	}

	return nil
}

//UpdateNodes updates alias with new nodes in DB
func (r Resource) UpdateNodes(new map[string]bool, old map[string]bool) (err error) {
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
func (r Resource) UpdateCnames(oldObject Resource) (err error) {

	//Split string and delete any possible empty values
	newCnames := deleteEmpty(strings.Split(r.Cname, ","))
	exCnames := deleteEmpty(strings.Split(oldObject.Cname, ","))

	if len(newCnames) > 0 {
		for _, value := range exCnames {
			if !StringInSlice(value, newCnames) {
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
			if !StringInSlice(value, exCnames) {
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

/////////////////////////DNS ACTIONS/////////////////////////

func (r Resource) createInDNS() error {
	entries := landbsoap.Conn().DNSDelegatedSearch(strings.Split(r.AliasName, ".")[0]) //check for existing aliases in DNS with the same name

	//Double-check that DNS doesn't contain such an alias
	if len(entries) == 0 {
		log.Info("Preparing to add " + r.AliasName + " in DNS")
		//If view is external, we need to create two entries in DNS
		views := make(map[string]string)
		views["internal"] = cfg.Soap.SoapKeynameI
		if StringInSlice(r.External, []string{"external", "yes"}) {
			views["external"] = cfg.Soap.SoapKeynameE
		}

		//Create the alias first
		for view, keyname := range views {
			//If alias creation succeeds, then we proceed with the rest
			if landbsoap.Conn().DNSDelegatedAdd(r.AliasName, view, keyname, "Created by:"+r.User, "goermis") {
				log.Info(r.AliasName + "/" + view + "has been created")
				//If alias is created successfully and there are also cnames...
				if r.Cname != "" {
					if r.createCnamesDNS(view) {
						log.Info("Cnames added in DNS for alias " + r.AliasName + "/" + view)
					} else {
						return errors.New("Failed to create cnames")
					}

				}

			} else {
				return errors.New("Failed to create domain " + r.AliasName + "/" + view + " in DNS")
			}
		}
		return nil

	}
	return errors.New("Alias entry with the same name exist in DNS, skipping creation")

}

func (r Resource) createCnamesDNS(view string) bool {

	cnames := deleteEmpty(strings.Split(r.Cname, ","))
	for _, cname := range cnames {
		log.Info("Adding in DNS the cname " + cname)
		if !landbsoap.Conn().DNSDelegatedAliasAdd(r.AliasName, view, cname) {
			return false
		}
	}

	return true
}

//UpdateDNS updates the cname or visibility changes in DNS
func (r Resource) UpdateDNS(oldObject Resource) (err error) {
	//1.If view has changed delete and recreate alias and cnames
	log.Info("Visibility has changed from " + oldObject.External + " to " + r.External)
	//a.Delete alias and Cnames from DNS
	if err := oldObject.deleteFromDNS(); err != nil {
		return errors.New("Failed to clean DNS from alias while updating: " + oldObject.AliasName + "/" + oldObject.External)
	}
	//b.Recreate alias with the new view and cnames
	if err := r.createInDNS(); err != nil {
		return errors.New("Failed to recreate alias while updating: " + r.AliasName + "/" + r.External)
	}
	return nil
	//2.Else if there is a change in cnames, update DNS
	/*} else if !Equal(r.Cname, oldObject.Cname) {
		if err := r.updateCnamesInDNS(oldObject); err != nil {
			return errors.New("Failed to update DNS with the new cnames for " + r.AliasName + "/" + r.External)
		}
	}*/

}

/*
func (r Resource) updateCnamesInDNS(oldObject Resource) error {
	//2.If only cnames have changed, update only them
	newCnames := deleteEmpty(strings.Split(r.Cname, ","))
	existingCnames := deleteEmpty(strings.Split(oldObject.Cname, ","))
	var views []string
	views = append(views, "internal")
	if StringInSlice(r.External, []string{"external", "yes"}) {
		views = append(views, "external")
	}
	for _, view := range views {
		if r.Cname != "" {
			for _, existingCname := range existingCnames {
				if !StringInSlice(existingCname, newCnames) {
					if !landbsoap.Conn().DNSDelegatedAliasRemove(name, oview, existingCname) {
						return errors.New("Failed to delete existing cname " +
							existingCname + " while updating DNS")
					}
				}
			}

			for _, newCname := range newCnames {
				if newCname == "" {
					continue
				}
				if !StringInSlice(newCname, existingCnames) {
					if !landbsoap.Conn().DNSDelegatedAliasAdd(name, oview, newCname) {
						return errors.New("Failed to add new cname in DNS " +
							newCname + " while updating alias " + name)
					}
				}

			}

		} else {
			for _, cname := range existingCnames {
				if !landbsoap.Conn().DNSDelegatedAliasRemove(name, oview, cname) {
					return errors.New("Failed to delete cname from DNS" +
						cname + " while purging all")
				}
			}
		}
	}
	return nil

}
*/
func (r Resource) deleteFromDNS() error {

	entries := landbsoap.Conn().DNSDelegatedSearch(strings.Split(r.AliasName, ".")[0])
	if len(entries) != 0 {
		log.Info("Preparing to delete " + r.AliasName + " from DNS")
		var views []string
		views = append(views, "internal")
		if StringInSlice(r.External, []string{"external", "yes"}) {
			views = append(views, "external")
		}
		for _, view := range views {
			if landbsoap.Conn().DNSDelegatedRemove(r.AliasName, view) {
				log.Info(r.AliasName + "/" + view + " has been deleted ")
			} else {
				return errors.New("Failed to delete " + r.AliasName + "/" + view + " from DNS. Rolling Back")

			}

		}

		return nil
	}
	return errors.New("The requested alias for deletion doesn't exist in DNS.Skipping deletion there")
}

//Clean the mess when something goes wrong
func (r Resource) rollBack(process string) {

	if process == "add" {
		//The r struct has ID=0, because ID is assigned after creation
		//For that reason, we retrieve the object from DB for deletion
		alias, _ := GetObjects(r.AliasName, "alias_name")
		alias[0].DeleteObject()

	}
	if process == "delete" {
		r.CreateObject()

	}

}
