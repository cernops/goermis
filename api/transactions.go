package api

/*This file contains the DB transactions*/
import (
	"errors"
	"time"

	"github.com/davecgh/go-spew/spew"
	cgorm "gitlab.cern.ch/lb-experts/goermis/db"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//CreateTransactions creates a new DB entry and its cname relations, with transactions
func CreateTransactions(alias Alias) (err error) {
	return WithinTransaction(func(tx *gorm.DB) (err error) {

		if err = tx.Where("alias_name=?", alias.AliasName).
			Create(&alias).
			Error; err != nil {
			tx.Rollback() // rollback
			return errors.New(alias.AliasName + " creation in DB failed with error: " +
				err.Error())
		}
		return nil
	})
}

//aliasUpdateTransactions updates non-associative alias parameters
//(best hosts, behaviour, hostgroup, metric, tenant etc.)
func aliasUpdateTransactions(a Alias) (err error) {
	return WithinTransaction(func(tx *gorm.DB) (err error) {
		if err = tx.Model(&a).Omit(clause.Associations).Updates(
			map[string]interface{}{
				"external":          a.External,
				"hostgroup":         a.Hostgroup,
				"best_hosts":        a.BestHosts,
				"metric":            a.Metric,
				"polling_interval":  a.PollingInterval,
				"ttl":               a.TTL,
				"tenant":            a.Tenant,
				"last_modification": time.Now(),
			}).Error; err != nil {
			return errors.New("Failed to update the single-valued fields with error: " + err.Error())

		}
		return nil
	})
}

//deleteTransactions deletes an entry and its relations from DB, with transactions
func deleteTransactions(alias Alias) (err error) {
	return WithinTransaction(func(tx *gorm.DB) (err error) {
		if tx.Select(clause.Associations).
			Delete(&alias); err != nil {
			return errors.New("Failed to delete alias from DB with error: " + err.Error())

		}
		//Delete node with no other relations
		for _, relation := range alias.Relations {
			if tx.Model(&relation.Node).Association("Aliases").Count() == 0 {
				if err = tx.Delete(&relation.Node).
					Error; err != nil {
					return errors.New("Failed to delete unrelated node " +
						relation.Node.NodeName +
						"Error: " + err.Error())

				}

			}
		}

		return nil
	})

}

//deleteNodeTransactions deletes  a Node from the database
func deleteNodeTransactions(v *Relation) (err error) {
	return WithinTransaction(func(tx *gorm.DB) (err error) {
		//Delete relation
		if err = tx.Set("gorm:association_autoupdate", false).
			Where("alias_id = ? AND node_id = ?", v.AliasID, v.NodeID).
			Delete(&Relation{}).
			Error; err != nil {
			tx.Rollback()
			return err
		}
		//Delete node with no other relations
		if tx.Model(&v.Node).Association("Aliases").Count() == 0 {
			if err = tx.Delete(&v.Node).
				Error; err != nil {
				tx.Rollback()
				return err

			}

		}

		return nil

	})
}

//addNodeTransactions adds a node in the DB
func addNodeTransactions(v *Relation) (err error) {
	return WithinTransaction(func(tx *gorm.DB) (err error) {
		//Either create a new node or find an existing one
		//Remember that its a many-2-many relationship, so nodes
		//can exist already, assigned to another alias
		spew.Dump(v)
		if err = tx.Where("node_name = ?", v.Node.NodeName).
			FirstOrCreate(&v.Node).
			Error; err != nil {
			tx.Rollback()
			return err
		}
		//Create the relationship for that alias and the
		//existing or newly created node
		if err = tx.Create(
			&Relation{
				AliasID:   v.AliasID,
				NodeID:    v.Node.ID,
				Blacklist: v.Blacklist},
		).Error; err != nil {
			tx.Rollback()
			return err
		}

		return nil
	})
}

//updatePrivilegeTransactions updates the privilege of a node from allowed to forbidden and vice versa
func updatePrivilegeTransactions(v *Relation) (err error) {
	return WithinTransaction(func(tx *gorm.DB) (err error) {
		//Update single column blacklist in Relations table
		if err = tx.Model(&v).
			Where("alias_id=? AND node_id = ?", v.AliasID, v.NodeID).
			Update("blacklist", v.Blacklist).
			Error; err != nil {
			tx.Rollback()
			return err
		}

		return nil
	})

}

//AddCnameTransactions appends a Cname
func addCnameTransactions(cname Cname) error {
	return WithinTransaction(func(tx *gorm.DB) (err error) {
		if err = tx.Create(&cname).Error; err != nil {
			tx.Rollback()
			return err
		}
		return nil
	})

}

//DeleteCnameTransactions cname from db during modification
//AutoUpdate is false, because otherwise we will be adding what we just deleted
func deleteCnameTransactions(cname Cname) error {
	return WithinTransaction(func(tx *gorm.DB) (err error) {
		if err = tx.Delete(&cname).
			Error; err != nil {
			tx.Rollback()
			return err
		}
		return nil

	})

}

//addAlarmTransactions appends an alarm
func addAlarmTransactions(alarm Alarm) error {
	return WithinTransaction(func(tx *gorm.DB) (err error) {
		if err = tx.Create(&alarm).
			Error; err != nil {
			tx.Rollback()
			return err
		}
		return nil
	})

}

//deleteAlarmTransactions deletes an alarm from db during modification
//AutoUpdate is false, because otherwise we will be adding what we just deleted
func deleteAlarmTransactions(alarm Alarm) error {
	return WithinTransaction(func(tx *gorm.DB) (err error) {
		if err = tx.Delete(&alarm).
			Error; err != nil {
			tx.Rollback()
			return err
		}
		return nil

	})

}

// WithinTransaction  accept dBFunc as parameter call dBFunc function within transaction begin, and commit and return error from dBFunc
func WithinTransaction(fn dBFunc) (err error) {
	tx := cgorm.ManagerDB().Begin() // start db transaction
	defer tx.Commit()
	err = fn(tx)

	if err != nil {
	}
	return err

}
