package models

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	schema "github.com/gorilla/Schema"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"gitlab.cern.ch/lb-experts/goermis/db"
	cgorm "gitlab.cern.ch/lb-experts/goermis/db"
)

var (
	con     = db.ManagerDB()
	decoder = schema.NewDecoder()
	q       string
)

//GetObjects return list of aliases if no parameters are passed or a single alias if parameters are given
func (r Resource) GetObjects(param string, tablerow string) (b []Resource, err error) {
	if param == "" && tablerow == "" {
		q = "SELECT a.id, alias_name, behaviour, best_hosts, clusters, " +
			"COALESCE(GROUP_CONCAT(distinct case when r.blacklist = 1 then n.node_name else null end),'') AS ForbiddenNodes, " +
			"COALESCE(GROUP_CONCAT(distinct case when r.blacklist = 0 then n.node_name else null end),'') AS AllowedNodes, " +
			"COALESCE(GROUP_CONCAT(distinct c_name),'') AS cname, external,  a.hostgroup, a.last_modification, metric, polling_interval,tenant, ttl, user, statistics " +
			"FROM alias a " +
			"LEFT join cname c on ( a.id=c.alias_id) " +
			"LEFT JOIN aliases_nodes r on (a.id=r.alias_id) " +
			"LEFT JOIN node n on (n.id=r.node_id)" +
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
			"where a." + tablerow + " = " + "'" + param + "'" +
			"GROUP BY a.id, alias_name, behaviour, best_hosts, clusters,  external, a.hostgroup, " +
			"a.last_modification, metric, polling_interval, statistics, tenant, ttl, user ORDER BY alias_name"
	}

	rows, err := con.Raw(q).Rows()

	defer rows.Close()
	if err != nil {
		log.Error("Error while getting list of objects: " + err.Error())
		return b, echo.NewHTTPError(http.StatusBadRequest, err.Error())

	}
	for rows.Next() {
		var result Resource
		err := rows.Scan(&result.ID, &result.AliasName, &result.Behaviour, &result.BestHosts, &result.Clusters,
			&result.ForbiddenNodes, &result.AllowedNodes, &result.Cname, &result.External, &result.Hostgroup,
			&result.LastModification, &result.Metric, &result.PollingInterval,
			&result.Tenant, &result.TTL, &result.User, &result.Statistics)
		if err != nil {
			log.Error("Error when scanning in GetObjectsList: " + err.Error())

			return b, echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		b = append(b, result)
	}
	return b, nil
}

//CreateObject creates an alias
func (a Alias) CreateObject(params url.Values) (err error) {

	decoder.IgnoreUnknownKeys(true)
	err = decoder.Decode(&a, params)
	if err != nil {
		log.Error("Error while decoding parameters : " + err.Error())
		//panic(err)

	}
	cnames := DeleteEmpty(strings.Split(params.Get("cnames"), ","))

	return WithinTransaction(func(tx *gorm.DB) (err error) {

		// check new object
		if !cgorm.ManagerDB().NewRecord(&a) {
			log.Error("Alias with the same  name as " + a.AliasName + "already exists")
			return err
		}
		if err = tx.Create(&a).Error; err != nil {
			tx.Rollback() // rollback
			log.Error("Error in creating alias " + a.AliasName)
			return err
		}

		if len(cnames) > 0 {
			for _, cname := range cnames {
				if !cgorm.ManagerDB().NewRecord(&Cname{CName: cname}) {
					log.Error("Cname " + cname + "exists")
					return err
				}

				if err = tx.Model(&a).Association("Cnames").Append(&Cname{CName: cname}).Error; err != nil {
					log.Error("Failed to add cname " + cname + "for alias " + a.AliasName)
					tx.Rollback()
					return err
				}
			}
		}

		return nil
	})
}

//Prepare prepares alias before creation with def values
func (a *Alias) Prepare() {
	//Populate the struct with the default values
	log.Info("Preparing alias " + a.AliasName + "with default values")
	a.User = "kkouros"
	a.Behaviour = "mindless"
	a.Metric = "cmsfrontier"
	a.PollingInterval = 300
	a.Statistics = "long"
	a.Clusters = "none"
	a.Tenant = "golang"
	a.LastModification = time.Now()
}

//DeleteObject deletes an alias and its Relations
func (a Alias) DeleteObject() (err error) {

	return WithinTransaction(func(tx *gorm.DB) (err error) {

		if tx.Where("alias_name = ?", a.AliasName).First(&a).RecordNotFound() {
			log.Error("Alias " + a.AliasName + "doesn't exist ?! ")
			return err

		}
		err = tx.Where("alias_id= ?", a.ID).Delete(&Cname{}).Error
		if err != nil {
			log.Error("Cname deletion failed")
			return err
		}
		//con.Model(&Alias).Where("alias_name = ?", alias).Preload("Cnames").Delete(&Alias.Cnames)
		err = tx.Where("alias_name = ?", a.AliasName).Delete(&Alias{}).Error
		if err != nil {
			log.Error("Alias deletion failed.Aliasname :" + a.AliasName)
			return err
		}

		return nil
	})
}

//ModifyObject modifies aliases and its associations
func (a Alias) ModifyObject(ex Resource, new Resource) (err error) {
	//Prepare cnames separately
	newCnames := DeleteEmpty(strings.Split(new.Cname, ","))
	exCnames := DeleteEmpty(strings.Split(ex.Cname, ","))
	a.ID = ex.ID
	a.Hostgroup = ex.Hostgroup

	if err = con.Model(&Alias{}).UpdateColumns(
		map[string]interface{}{
			"external":   new.External,
			"hostgroup":  new.Hostgroup,
			"best_hosts": new.BestHosts,
		}).Error; err != nil {
		log.Error("Error while updating alias " + ex.AliasName)

	}
	//err = a.UpdateNodes()
	err = a.UpdateCnames(exCnames, newCnames)
	if err != nil {
		log.Error("Unable to update cnames ,Error : " + err.Error())
		return err
	}
	log.Info("Before node update")
	err = a.UpdateNodes(nodesToMap(ex), nodesToMap(new))
	if err != nil {

		log.Error("Unable to update nodes ,Error : " + err.Error())
		return err
	}

	return nil
}

//UpdateCnames updates cnames
func (a Alias) UpdateCnames(exCnames []string, newCnames []string) (err error) {
	// If there are no cnames from UI , delete them all, otherwise append them

	spew.Dump(exCnames)
	spew.Dump(newCnames)
	if len(newCnames) > 0 {
		for _, value := range exCnames {
			if !stringInSlice(value, newCnames) {
				log.Info("Deleting cname")
				if err = a.DeleteCname(value); err != nil {
					return err
				}
			}
		}

		for _, value := range newCnames {
			if value == "" {
				continue
			}
			if !stringInSlice(value, exCnames) {
				log.Info("Adding cname")
				if err = a.AddCname(value); err != nil {
					return err
				}
			}

		}

	} else {
		for _, value := range exCnames {
			log.Info("In cname deletion")

			if err = a.DeleteCname(value); err != nil {
				return err
			}
		}
	}
	return nil
}

//UpdateNodes updates cnames
func (a Alias) UpdateNodes(ex map[string]bool, new map[string]bool) (err error) {
	log.Info("Updating nodes")
	for k, v := range ex {
		log.Info("inside first for")
		if _, ok := new[k]; !ok {
			if err = a.DeleteNode(k, v); err != nil {
				return err
			}
		}
	}
	for k, v := range new {
		if k == "" {
			continue
		}
		if _, ok := ex[k]; !ok {
			log.Info("Addit")
			if err = a.AddNode(k, v); err != nil {
				log.Error("we aar fucked")
				return err
			}
		} else if value, ok := ex[k]; ok && value != v {
			if err = a.UpdateNodePrivilege(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

//DeleteNode deletes  a Node from the database
func (a Alias) DeleteNode(name string, p bool) (err error) {
	var node Node

	return WithinTransaction(func(tx *gorm.DB) (err error) {
		//find node
		if err := tx.First(&node, "node_name=?", name).Error; err != nil {
			log.Error("Coouldnt find the ID of the node for deletion")
			tx.Rollback()
			return err
		}

		//Delete relation
		if err = tx.Set("gorm:association_autoupdate", false).
			Where("alias_id = ? AND node_id = ?", a.ID, node.ID).
			Delete(&AliasesNodes{}).Error; err != nil {
			tx.Rollback()
			log.Error("Error while delete in transaction node" + name)
			return err
		}
		//Delete node with no other relations
		if tx.Model(&node).Association("Aliases").Count() == 0 {
			log.Info("Node doesnt have other relation, delete it")
			if err = tx.Delete(&node).Error; err != nil {
				log.Info("Error while deleting node with no relations")
				tx.Rollback()
				return err

			}

		}

		return nil

	})
}

//AddNode adds a node in the DB
func (a Alias) AddNode(name string, p bool) (err error) {
	var node Node

	return WithinTransaction(func(tx *gorm.DB) (err error) {
		err = tx.Where("node_name = ?", name).
			Assign(Node{NodeName: name,
				LastModification: time.Now()}).
			FirstOrCreate(&node).Error

		if err != nil {
			tx.Rollback()
			log.Error("Error while dealing with node ")
			return err
		}
		if tx.Where("alias_id = ? AND node_id = ?", a.ID, node.ID).First(&AliasesNodes{}).RecordNotFound() {
			log.Info("Righttt")
			if err = tx.Model(&AliasesNodes{}).Create(
				prepareRelation(node.ID, a.ID, p),
			).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		return nil
	})
}

//UpdateNodePrivilege updates the privilege of a node from allowed to forbidden and vice versa
func (a Alias) UpdateNodePrivilege(name string, p bool) (err error) {
	var node Node
	return WithinTransaction(func(tx *gorm.DB) (err error) {

		//find node
		if err := tx.First(&node, "node_name=?", name).Error; err != nil {
			log.Error("Coouldnt find the ID of the node for deletion")
			tx.Rollback()
			return err
		}

		if err = tx.Model(&AliasesNodes{}).
			Where("alias_id=? AND node_id = ?", a.ID, node.ID).
			Update("blacklist", p).Error; err != nil {
			log.Error("Error while updating privileges")
			tx.Rollback()
			return err
		}

		return nil
	})

}

//AddCname appends a Cname
func (a Alias) AddCname(cname string) error {
	return WithinTransaction(func(tx *gorm.DB) (err error) {
		if !cgorm.ManagerDB().NewRecord(&Cname{CName: cname}) {
			log.Error("Cname" + cname + " already exists")
			return err
		}

		if err = tx.Set("gorm:association_autoupdate", false).First(&Alias{}, "id=?", a.ID).Association("Cnames").Append(&Cname{CName: cname}).Error; err != nil {
			log.Info("There was an error while adding cname " + string(cname))
			tx.Rollback()
			return err
		}
		return nil
	})

}

//DeleteCname cname from db during modification
//AutoUpdate is false, because otherwise we will be adding what we just deleted
func (a Alias) DeleteCname(cname string) error {
	return WithinTransaction(func(tx *gorm.DB) (err error) {
		if err = tx.Set("gorm:association_autoupdate", false).Where("alias_id = ? AND c_name = ?", a.ID, cname).Delete(&Cname{}).Error; err != nil {
			tx.Rollback()
			log.Error("Error while delete in transaction cname" + cname)
			return err
		}
		return nil

	})

}

//GetExistingData retrieves all the data for a certain alias, for internal use
func GetExistingData(aliasName string) (a Alias, err error) {

	if con.Model(Alias{}).Preload("Cnames").Preload("Nodes").Where("alias_name = ?", aliasName).First(&a).RecordNotFound() {
		log.Error("There was an error while getting existing data for alias " + aliasName + "Error: " + err.Error())
		return a, err

	}
	return a, nil
}

// WithinTransaction  accept DBFunc as parameter call DBFunc function within transaction begin, and commit and return error from DBFunc
func WithinTransaction(fn DBFunc) (err error) {
	tx := cgorm.ManagerDB().Begin() // start db transaction
	defer tx.Commit()
	err = fn(tx)

	if err != nil {
		log.Error("Error within transaction: " + err.Error())
	}
	return err

}
