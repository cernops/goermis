package models

import (
	"strings"
	"time"

	schema "github.com/gorilla/Schema"
	"github.com/jinzhu/copier"
	"github.com/jinzhu/gorm"
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
		log.Error("Error while getting list of objects: " + err.Error())
		return b, err

	}
	defer rows.Close()
	for rows.Next() {
		var result Resource
		err := rows.Scan(&result.ID, &result.AliasName, &result.Behaviour, &result.BestHosts, &result.Clusters,
			&result.ForbiddenNodes, &result.AllowedNodes, &result.Cname, &result.External, &result.Hostgroup,
			&result.LastModification, &result.Metric, &result.PollingInterval,
			&result.Tenant, &result.TTL, &result.User, &result.Statistics)
		if err != nil {
			log.Error("Error when scanning in GetObjectsList: " + err.Error())

			return b, err
		}
		b = append(b, result)
	}
	return b, nil
}

//CreateObject creates an alias
func (r Resource) CreateObject() (err error) {
	var a Alias
	copier.Copy(&a, &r)
	cnames := DeleteEmpty(strings.Split(r.Cname, ","))
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

//AddDefaultValues prepares alias before creation with def values
func (r *Resource) AddDefaultValues() {
	//Populate the struct with the default values
	log.Info("Preparing alias " + r.AliasName + " with the default values")

	//Passing default values
	r.User = "kkouros"
	r.Behaviour = "mindless"
	r.Metric = "cmsfrontier"
	r.PollingInterval = 300
	r.Statistics = "long"
	r.Clusters = "none"
	r.Tenant = "golang"
	r.LastModification = time.Now()
}

//Hydrate prepares a few input data before creation of new alias
func (r *Resource) Hydrate() {
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
	log.Info("Inside delete")
	var relation []AliasesNodes

	return WithinTransaction(func(tx *gorm.DB) (err error) {

		//Make sure alias exists
		if tx.Where("alias_name = ?", r.AliasName).First(&Alias{}).RecordNotFound() {
			log.Error("Alias " + r.AliasName + "doesn't exist ?! ")
			return err

		}
		//Find and store all relations
		if err := tx.Where("alias_id=?", r.ID).Find(&relation).Error; err != nil {
			log.Error("Error while getting all relations for the alias")
			return err
		}

		for _, v := range relation {
			var node Node
			//Find node itself and load
			if err := tx.Where("id=?", v.NodeID).First(&node).Error; err != nil {
				log.Error("Error while getting the node of a certain relation")
				return err
			}
			// Delete relation first
			err = tx.Where("node_id=? AND alias_id =? ", v.NodeID, r.ID).Delete(&AliasesNodes{}).Error
			if err != nil {
				log.Error("Failed to clear nodes during alias deletion")
				return err
			}

			//Delete node with no other relations
			if tx.Model(&node).Association("Aliases").Count() == 0 {
				log.Info("Node doesnt have other relation, delete it")
				if err = con.Delete(&node).Error; err != nil {
					log.Info("Error while deleting node with no relations")
					//con.Rollback()
					return err

				}

			}

		}
		//Delete cnames
		err = tx.Where("alias_id= ?", r.ID).Delete(&Cname{}).Error
		if err != nil {
			log.Error("Cname deletion failed")
			return err
		}

		//Finally delete alias
		err = tx.Where("alias_name = ?", r.AliasName).Delete(&Alias{}).Error
		if err != nil {
			log.Error("Alias deletion failed.Aliasname :" + r.AliasName)
			return err
		}

		return nil
	})
}

//ModifyObject modifies aliases and its associations
func (r Resource) ModifyObject(new Resource) (err error) {
	//Prepare cnames separately
	newCnames := DeleteEmpty(strings.Split(new.Cname, ","))
	exCnames := DeleteEmpty(strings.Split(r.Cname, ","))

	if err = con.Model(&Alias{}).UpdateColumns(
		map[string]interface{}{
			"external":   new.External,
			"hostgroup":  new.Hostgroup,
			"best_hosts": new.BestHosts,
		}).Error; err != nil {
		log.Error("Error while updating alias " + r.AliasName)

	}
	//err = a.UpdateNodes()
	err = r.UpdateCnames(exCnames, newCnames)
	if err != nil {
		log.Error("Unable to update cnames ,Error : " + err.Error())
		return err
	}
	log.Info("Before node update")
	err = r.UpdateNodes(nodesToMap(r), nodesToMap(new))
	if err != nil {

		log.Error("Unable to update nodes ,Error : " + err.Error())
		return err
	}

	return nil
}

//UpdateCnames updates cnames
func (r Resource) UpdateCnames(exCnames []string, newCnames []string) (err error) {

	if len(newCnames) > 0 {
		for _, value := range exCnames {
			if !StringInSlice(value, newCnames) {
				log.Info("Deleting cname")
				if err = r.DeleteCname(value); err != nil {
					return err
				}
			}
		}

		for _, value := range newCnames {
			if value == "" {
				continue
			}
			if !StringInSlice(value, exCnames) {
				log.Info("Adding cname")
				if err = r.AddCname(value); err != nil {
					return err
				}
			}

		}

	} else {
		for _, value := range exCnames {
			log.Info("In cname deletion")

			if err = r.DeleteCname(value); err != nil {
				return err
			}
		}
	}
	return nil
}

//UpdateNodes updates cnames
func (r Resource) UpdateNodes(ex map[string]bool, new map[string]bool) (err error) {
	log.Info("Updating nodes")
	for k, v := range ex {
		if _, ok := new[k]; !ok {
			if err = r.DeleteNode(k, v); err != nil {
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
			if err = r.AddNode(k, v); err != nil {
				log.Error("Error in node addition ")
				return err
			}
		} else if value, ok := ex[k]; ok && value != v {
			if err = r.UpdateNodePrivilege(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

//DeleteNode deletes  a Node from the database
func (r Resource) DeleteNode(name string, p bool) (err error) {
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
			Where("alias_id = ? AND node_id = ?", r.ID, node.ID).
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
func (r Resource) AddNode(name string, p bool) (err error) {
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
		if tx.Where("alias_id = ? AND node_id = ?", r.ID, node.ID).First(&AliasesNodes{}).RecordNotFound() {
			if err = tx.First(&Alias{}, "id=?", r.ID).Create(
				prepareRelation(node.ID, r.ID, p),
			).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		return nil
	})
}

//UpdateNodePrivilege updates the privilege of a node from allowed to forbidden and vice versa
func (r Resource) UpdateNodePrivilege(name string, p bool) (err error) {
	var node Node
	return WithinTransaction(func(tx *gorm.DB) (err error) {

		//find node
		if err := tx.First(&node, "node_name=?", name).Error; err != nil {
			log.Error("Coouldnt find the ID of the node for deletion")
			tx.Rollback()
			return err
		}

		if err = tx.Model(&AliasesNodes{}).
			Where("alias_id=? AND node_id = ?", r.ID, node.ID).
			Update("blacklist", p).Error; err != nil {
			log.Error("Error while updating privileges")
			tx.Rollback()
			return err
		}

		return nil
	})

}

//AddCname appends a Cname
func (r Resource) AddCname(cname string) error {
	return WithinTransaction(func(tx *gorm.DB) (err error) {
		if !cgorm.ManagerDB().NewRecord(&Cname{CName: cname}) {
			log.Error("Cname" + cname + " already exists")
			return err
		}

		if err = tx.Set("gorm:association_autoupdate", false).First(&Alias{}, "id=?", r.ID).Association("Cnames").Append(&Cname{CName: cname}).Error; err != nil {
			log.Info("There was an error while adding cname " + string(cname))
			tx.Rollback()
			return err
		}
		return nil
	})

}

//DeleteCname cname from db during modification
//AutoUpdate is false, because otherwise we will be adding what we just deleted
func (r Resource) DeleteCname(cname string) error {
	return WithinTransaction(func(tx *gorm.DB) (err error) {
		if err = tx.Set("gorm:association_autoupdate", false).Where("alias_id = ? AND c_name = ?", r.ID, cname).Delete(&Cname{}).Error; err != nil {
			tx.Rollback()
			log.Error("Error while delete in transaction cname" + cname)
			return err
		}
		return nil

	})

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
