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

	//DNS//
	entries := landbsoap.Soap.DNSDelegatedSearch(strings.Split(r.AliasName, ".")[0])
	//Double-check that DNS doesn't contain such an alias
	if len(entries) == 0 {
		log.Info("Preparing to add " + r.AliasName + " in DNS")
		view := "internal"
		keyname := cfg.Soap.SoapKeynameI
		if StringInSlice(r.External, []string{"external", "yes"}) {
			view = "external"
			keyname = cfg.Soap.SoapKeynameE
		}
		//Create the alias first
		if landbsoap.Soap.DNSDelegatedAdd(r.AliasName, view, keyname, "Created by: gouser", "testing go") {
			log.Info(r.AliasName + "/" + view + "has been created")

			//If alias is created successfully and there are also cnames...
			if len(cnames) > 0 {
				for _, cname := range cnames {
					log.Info("Adding in DNS the cname " + cname)

					//...Create cnames,too
					if landbsoap.Soap.DNSDelegatedAliasAdd(r.AliasName, view, cname) {
						log.Info("Alias " + cname + " has been created for " +
							r.AliasName + "/" + view)
					} else {
						//If cname creation fails, clear the mess by deleting alias
						//First from DNS
						if landbsoap.Soap.DNSDelegatedRemove(r.AliasName, view) {
							log.Info("Cleared DNS from the failed addition")
						}
						//Then from DB
						orm.DeleteTransactions(r.AliasName, r.ID)
						return errors.New("Failed to add cname " +
							cname + "for alias " + r.AliasName + " in DNS. Rolling back ")
					}
				}

			}
			return nil
		}
		//Failed to add in DNS, so lets clean DB
		orm.DeleteTransactions(r.AliasName, r.ID)
		return errors.New("Failed to add alias " + r.AliasName + " in DNS")
	}
	return errors.New("Alias entry with the same name exist in DNS, skipping creation")
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
	//DB
	if err := orm.DeleteTransactions(r.AliasName, r.ID); err != nil {
		return err
	}

	//DNS
	entries := landbsoap.Soap.DNSDelegatedSearch(strings.Split(r.AliasName, ".")[0])
	if len(entries) != 0 {
		log.Info("Preparing to delete " + r.AliasName + " from DNS")
		view := "internal"
		if StringInSlice(r.External, []string{"external", "yes"}) {
			view = "external"
		}
		if landbsoap.Soap.DNSDelegatedRemove(r.AliasName, view) {
			log.Info(r.AliasName + "/" + view + "and all its cnames have been deleted ")
			return nil
		}
		log.Info("Failed to delete " + r.AliasName + " from DNS. Rolling Back")
		//Recreate it in DB
		r.CreateObject()

		return errors.New("Failed to delete alias " + r.AliasName + "from DNS")
	}
	return errors.New("The requested alias for deletion doesn't exist in DNS.Skipping deletion there")
}

//ModifyObject modifies aliases and its associations
func (r Resource) ModifyObject(oldValues map[string]string) (err error) {
	//Prepare cnames separately
	newCnames := deleteEmpty(strings.Split(r.Cname, ","))
	exCnames := deleteEmpty(strings.Split(oldValues["Cnames"], ","))

	//Let's update the single-valued fields first
	if err = con.Model(&orm.Alias{}).Where("id = ?", r.ID).UpdateColumns(
		map[string]interface{}{
			"external":   r.External,
			"hostgroup":  r.Hostgroup,
			"best_hosts": r.BestHosts,
		}).Error; err != nil {
		return errors.New("Failed to update the single-valued fields with error: " + err.Error())

	}

	//Update cnames for object r with new cnames
	if err = UpdateCnames(r.ID, exCnames, newCnames); err != nil {
		return err
	}
	/*Update nodes for r object with new nodes(nodesToMap converts string to map,
	  where value indicates privilege allowed/forbidden)*/
	newNodesMap := nodesInMap(r.AllowedNodes, r.ForbiddenNodes)
	oldNodesMap := nodesInMap(oldValues["Allowed"], oldValues["Forbidden"])
	if err = UpdateNodes(r.ID, newNodesMap, oldNodesMap); err != nil {
		return err
	}

	if err = UpdateDNS(r.AliasName, oldValues["View"], r.External, newCnames); err != nil {
		return err
	}
	return nil
}

//UpdateDNS updates the cname or visibility changes in DNS
func UpdateDNS(name string, oldView string, newView string, newCnames []string) (err error) {
	oview := "internal"
	nview := "internal"
	keyname := cfg.Soap.SoapKeynameI
	existingCnames := landbsoap.Soap.GimeCnamesOf(strings.Split(name, ".")[0])

	if StringInSlice(oldView, []string{"yes", "external"}) {
		oview = "external"
	}
	if StringInSlice(newView, []string{"yes", "external"}) {
		nview = "external"
		keyname = cfg.Soap.SoapKeynameE
	}

	//View has changed so we delete and recreate alias with the new visibility
	if oview != nview {
		log.Info("Visibility has changed from " + oview + " to " + nview)
		//Delete alias
		if landbsoap.Soap.DNSDelegatedRemove(name, oview) {
			log.Info("Deleting existing entry for " + name)
			//Add again with the new view
			if landbsoap.Soap.DNSDelegatedAdd(name, nview, keyname, "createdby:GO", "gotest") {
				log.Info("The new entry with updated visibility created successfully for " + name + "/" + nview)
				for _, cname := range newCnames {
					if !landbsoap.Soap.DNSDelegatedAliasAdd(name, nview, cname) {
						return errors.New("Failed to update DNS ,couldn't recreate cnames for alias " + name)
					}
				}
				log.Info("Successful DNS update for alias " + name)
				return nil
			}
			return errors.New("Failed to update DNS, couldn't recreate alias with new visibility")

		}
		return errors.New("Failed to update DNS, couldn't delete existing alias")

	}

	if len(newCnames) > 0 {
		for _, existingCname := range existingCnames {
			if !StringInSlice(existingCname, newCnames) {
				if !landbsoap.Soap.DNSDelegatedAliasRemove(name, oview, existingCname) {
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
				if !landbsoap.Soap.DNSDelegatedAliasAdd(name, oview, newCname) {
					return errors.New("Failed to add new cname in DNS " +
						newCname + " while updating alias " + name)
				}
			}

		}

	} else {
		for _, cname := range existingCnames {
			if !landbsoap.Soap.DNSDelegatedAliasRemove(name, oview, cname) {
				return errors.New("Failed to delete cname from DNS" +
					cname + " while purging all")
			}
		}
	}
	return nil
}

//UpdateNodes updates alias with new nodes
func UpdateNodes(aliasID int, new map[string]bool, old map[string]bool) (err error) {
	for name := range old {
		if _, ok := new[name]; !ok {
			if err = orm.DeleteNodeTransactions(aliasID, name); err != nil {
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
			if err = orm.AddNodeTransactions(aliasID, name, privilege); err != nil {
				return errors.New("Failed to add new node " +
					name + " while updating, with error: " + err.Error())
			}
		} else if value, ok := old[name]; ok && value != privilege {
			if err = orm.UpdatePrivilegeTransactions(aliasID, name, privilege); err != nil {
				return errors.New("Failed to update privilege for node " +
					name + " while updating, with error: " + err.Error())
			}
		}
	}

	return nil

}

//UpdateCnames updates cnames
func UpdateCnames(aliasID int, exCnames []string, newCnames []string) (err error) {

	if len(newCnames) > 0 {
		for _, value := range exCnames {
			if !StringInSlice(value, newCnames) {
				if err = orm.DeleteCnameTransactions(aliasID, value); err != nil {
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
				if err = orm.AddCnameTransactions(aliasID, value); err != nil {
					return errors.New("Failed to add new cname " +
						value + " while updating, with error: " + err.Error())
				}
			}

		}

	} else {
		for _, value := range exCnames {
			if err = orm.DeleteCnameTransactions(aliasID, value); err != nil {
				return errors.New("Failed to delete cname " +
					value + " while purging all, with error: " + err.Error())
			}
		}
	}
	return nil
}