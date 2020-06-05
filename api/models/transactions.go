package models

import (
	"errors"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	cgorm "gitlab.cern.ch/lb-experts/goermis/db"
)

//CreateTransactions creates a new DB entry and its cname relations, with transactions
func CreateTransactions(a Alias, cnames []string) (err error) {
	return WithinTransaction(func(tx *gorm.DB) (err error) {

		// check new object's primary key
		if !cgorm.ManagerDB().NewRecord(&a) {
			return errors.New("Blank primary key for alias")
		}
		if err = tx.Create(&a).Error; err != nil {
			tx.Rollback() // rollback
			return errors.New(a.AliasName + " creation in DB failed with error: " +
				err.Error())
		}

		if len(cnames) > 0 {
			for _, cname := range cnames {
				if !cgorm.ManagerDB().NewRecord(&Cname{CName: cname}) {
					return errors.New("Blank primary key for cname")
				}

				if err = tx.Model(&a).Association("Cnames").Append(&Cname{CName: cname}).Error; err != nil {
					tx.Rollback()
					return errors.New(cname + " creation in DB failed with error: " +
						err.Error())
				}
			}
		}

		return nil
	})

}

//DeleteTransactions deletes an entry and its relations from DB, with transactions
func DeleteTransactions(name string, ID int) (err error) {
	var relation []AliasesNodes

	return WithinTransaction(func(tx *gorm.DB) (err error) {

		//Make sure alias exists
		if tx.Where("alias_name = ?", name).First(&Alias{}).RecordNotFound() {
			return errors.New("RecordNotFound Error while trying to delete alias from DB")

		}

		//Find and store all relations
		if err := tx.Where("alias_id=?", ID).Find(&relation).Error; err != nil {
			return errors.New("Failed to find node relations with error: " + err.Error())
		}

		for _, v := range relation {
			var node Node
			//Find node itself with reverse looking and load
			if err := tx.Where("id=?", v.NodeID).First(&node).Error; err != nil {
				return errors.New("Failed to reverse look node with ID " + strconv.Itoa(v.NodeID))
			}
			// Delete relation first
			err = tx.Where("node_id=? AND alias_id =? ", v.NodeID, ID).Delete(&AliasesNodes{}).Error
			if err != nil {
				return errors.New("Failed to delete the relation with nodeID " +
					strconv.Itoa(v.NodeID) +
					"Error: " + err.Error())
			}

			//Delete node with no other relations
			if tx.Model(&node).Association("Aliases").Count() == 0 {
				if err = con.Delete(&node).Error; err != nil {
					return errors.New("Failed to delete unrelated node " +
						node.NodeName +
						"Error: " + err.Error())

				}

			}

		}
		//Delete cnames
		err = tx.Where("alias_id= ?", ID).Delete(&Cname{}).Error
		if err != nil {
			return errors.New("Failed to delete cnames from DB with error: " + err.Error())
		}

		//Finally delete alias
		err = tx.Where("alias_name = ?", name).Delete(&Alias{}).Error
		if err != nil {
			return errors.New("Failed to delete alias from DB with error: " + err.Error())
		}

		return nil
	})

}

//DeleteNode deletes  a Node from the database
func DeleteNode(aliasID int, name string) (err error) {
	var node Node
	return WithinTransaction(func(tx *gorm.DB) (err error) {
		//find node
		if err := tx.First(&node, "node_name=?", name).Error; err != nil {
			tx.Rollback()
			return err
		}

		//Delete relation
		if err = tx.Set("gorm:association_autoupdate", false).
			Where("alias_id = ? AND node_id = ?", aliasID, node.ID).
			Delete(&AliasesNodes{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		//Delete node with no other relations
		if tx.Model(&node).Association("Aliases").Count() == 0 {
			if err = tx.Delete(&node).Error; err != nil {
				tx.Rollback()
				return err

			}

		}

		return nil

	})
}

//AddNode adds a node in the DB
func AddNode(aliasID int, name string, privilege bool) (err error) {
	var node Node

	return WithinTransaction(func(tx *gorm.DB) (err error) {
		err = tx.Where("node_name = ?", name).
			Assign(Node{NodeName: name,
				LastModification: time.Now()}).
			FirstOrCreate(&node).Error

		if err != nil {
			tx.Rollback()
			return err
		}
		if tx.Where("alias_id = ? AND node_id = ?", aliasID, node.ID).First(&AliasesNodes{}).RecordNotFound() {
			if err = tx.First(&Alias{}, "id=?", aliasID).Create(
				prepareRelation(node.ID, aliasID, privilege),
			).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		return nil
	})
}

//UpdateNodePrivilege updates the privilege of a node from allowed to forbidden and vice versa
func UpdateNodePrivilege(aliasID int, name string, p bool) (err error) {
	var node Node
	return WithinTransaction(func(tx *gorm.DB) (err error) {

		//find node
		if err := tx.First(&node, "node_name=?", name).Error; err != nil {
			tx.Rollback()
			return err
		}

		if err = tx.Model(&AliasesNodes{}).
			Where("alias_id=? AND node_id = ?", aliasID, node.ID).
			Update("blacklist", p).Error; err != nil {
			tx.Rollback()
			return err
		}

		return nil
	})

}

//AddCname appends a Cname
func AddCname(aliasID int, cname string) error {
	return WithinTransaction(func(tx *gorm.DB) (err error) {
		if !cgorm.ManagerDB().NewRecord(&Cname{CName: cname}) {
			return err
		}

		if err = tx.Set("gorm:association_autoupdate", false).First(&Alias{}, "id=?", aliasID).Association("Cnames").Append(&Cname{CName: cname}).Error; err != nil {
			tx.Rollback()
			return err
		}
		return nil
	})

}

//DeleteCname cname from db during modification
//AutoUpdate is false, because otherwise we will be adding what we just deleted
func DeleteCname(aliasID int, cname string) error {
	return WithinTransaction(func(tx *gorm.DB) (err error) {
		if err = tx.Set("gorm:association_autoupdate", false).Where("alias_id = ? AND c_name = ?", aliasID, cname).Delete(&Cname{}).Error; err != nil {
			tx.Rollback()
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
	}
	return err

}
