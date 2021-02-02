package api

/*
//CreateTransactions creates a new DB entry and its cname relations, with transactions
func CreateTransactions(a Alias) (err error) {
	return WithinTransaction(func(tx *gorm.DB) (err error) {

		// check new object's doesnt already exist
		if cgorm.ManagerDB().First(&a) == nil {
			return errors.New("An alias exists with the same name")
		}
		if err = tx.Create(&a).
			Error; err != nil {
			tx.Rollback() // rollback
			return errors.New(a.AliasName + " creation in DB failed with error: " +
				err.Error())
		}
		/*
			if len(cnames) > 0 {
				for _, cname := range cnames {
					if !cgorm.ManagerDB().
						NewRecord(&Cname{
							Cname: cname}) {
						return errors.New("Blank primary key for cname")
					}

					if err = tx.Model(&a).
						Association("Cnames").
						Append(&Cname{
							Cname: cname}).
						Error; err != nil {
						tx.Rollback()
						return errors.New(cname + " creation in DB failed with error: " +
							err.Error())
					}
				}
			}

		//return nil
	})

}

//DeleteTransactions deletes an entry and its relations from DB, with transactions
func DeleteTransactions(name string, ID int) (err error) {
	var relation []Relation

	return WithinTransaction(func(tx *gorm.DB) (err error) {

		//Make sure alias exists
		if tx.Where("alias_name = ?", name).
			First(&Alias{}).RecordNotFound() {
			return errors.New("RecordNotFound Error while trying to delete alias from DB")

		}

		//Find and store all relations
		if err := tx.Where("alias_id=?", ID).
			Find(&relation).
			Error; err != nil {
			return errors.New("Failed to find node relations with error: " + err.Error())
		}

		for _, v := range relation {
			var node Node
			//Find node itself with reverse looking and load
			if err := tx.Where("id=?", v.NodeID).
				First(&node).
				Error; err != nil {
				return errors.New("Failed to reverse look node with ID " + strconv.Itoa(v.NodeID))
			}
			// Delete relation first
			if err = tx.Where("node_id=? AND alias_id =? ", v.NodeID, ID).
				Delete(&Relation{}).
				Error; err != nil {
				return errors.New("Failed to delete the relation with nodeID " +
					strconv.Itoa(v.NodeID) +
					"Error: " + err.Error())
			}

			//Delete node with no other relations
			if tx.Model(&node).Association("Aliases").Count() == 0 {
				if err = tx.Delete(&node).
					Error; err != nil {
					return errors.New("Failed to delete unrelated node " +
						node.NodeName +
						"Error: " + err.Error())

				}

			}

		}
		//Delete cnames
		if err = tx.Where("cname_alias_id= ?", ID).
			Delete(&Cname{}).
			Error; err != nil {
			return errors.New("Failed to delete cnames from DB with error: " + err.Error())
		}

		//Delete alarms
		if err = tx.Where("alarm_alias_id= ?", ID).
			Delete(&Alarm{}).
			Error; err != nil {
			return errors.New("Failed to delete alarms from DB with error: " + err.Error())
		}

		//Finally delete alias
		if err = tx.Where("alias_name = ?", name).
			Delete(&Alias{}).
			Error; err != nil {
			return errors.New("Failed to delete alias from DB with error: " + err.Error())
		}

		return nil
	})

}
/*
//DeleteNodeTransactions deletes  a Node from the database
func DeleteNodeTransactions(aliasID int, name string) (err error) {
	var node Node
	return WithinTransaction(func(tx *gorm.DB) (err error) {
		//find node
		if err := tx.First(&node, "node_name=?", name).
			Error; err != nil {
			tx.Rollback()
			return err
		}

		//Delete relation
		if err = tx.Set("gorm:association_autoupdate", false).
			Where("alias_id = ? AND node_id = ?", aliasID, node.ID).
			Delete(&Relation{}).
			Error; err != nil {
			tx.Rollback()
			return err
		}
		//Delete node with no other relations
		if tx.Model(&node).Association("Aliases").Count() == 0 {
			if err = tx.Delete(&node).
				Error; err != nil {
				tx.Rollback()
				return err

			}

		}

		return nil

	})
}/*

//AddNodeTransactions adds a node in the DB
func AddNodeTransactions(aliasID int, name string, privilege bool) (err error) {
	var node Node

	return WithinTransaction(func(tx *gorm.DB) (err error) {
		if err = tx.Where("node_name = ?", name).
			Assign(Node{
				NodeName:         name,
				LastModification: time.Now()}).
			FirstOrCreate(&node).
			Error; err != nil {
			tx.Rollback()
			return err
		}
		if tx.Where("alias_id = ? AND node_id = ?", aliasID, node.ID).First(&Relation{}).RecordNotFound() {
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
/*
//UpdatePrivilegeTransactions updates the privilege of a node from allowed to forbidden and vice versa
func UpdatePrivilegeTransactions(aliasID int, name string, p bool) (err error) {
	var node Node
	return WithinTransaction(func(tx *gorm.DB) (err error) {

		//find node
		if err := tx.First(&node, "node_name=?", name).
			Error; err != nil {
			tx.Rollback()
			return err
		}

		if err = tx.Model(&Relation{}).
			Where("alias_id=? AND node_id = ?", aliasID, node.ID).
			Update("blacklist", p).
			Error; err != nil {
			tx.Rollback()
			return err
		}

		return nil
	})

}
/*
//AddCnameTransactions appends a Cname
func AddCnameTransactions(aliasID int, cname string) error {
	return WithinTransaction(func(tx *gorm.DB) (err error) {
		if cgorm.ManagerDB().First(&Cname{Cname: cname}) == nil {
			return err
		}

		if err = tx.Set("gorm:association_autoupdate", false).
			First(&Alias{}, "id=?", aliasID).
			Association("Cnames").
			Append(&Cname{
				Cname: cname}).
			Error; err != nil {
			tx.Rollback()
			return err
		}
		return nil
	})

}/*

//DeleteCnameTransactions cname from db during modification
//AutoUpdate is false, because otherwise we will be adding what we just deleted
func DeleteCnameTransactions(aliasID int, cname string) error {
	return WithinTransaction(func(tx *gorm.DB) (err error) {
		if err = tx.Set("gorm:association_autoupdate", false).
			Where("cname_alias_id = ? AND cname = ?", aliasID, cname).
			Delete(&Cname{}).
			Error; err != nil {
			tx.Rollback()
			return err
		}
		return nil

	})

}/*

//AddAlarmTransactions appends an alarm
func AddAlarmTransactions(aliasID int, aliasName string, alarm string) error {
	return WithinTransaction(func(tx *gorm.DB) (err error) {
		//parametes[0] --> alarm name ; parameters[1] --> recipient;
		//parameters[1] --> threshold parameter
		//Checking if we can actually create it by making sure there is no duplicate
		parameters := strings.Split(alarm, ":")
		if !cgorm.ManagerDB().
			NewRecord(&Alarm{
				Name: parameters[0]}) {
			return err
		}

		if err = tx.Set("gorm:association_autoupdate", false).
			First(&Alias{}, "id=?", aliasID).
			Association("Alarms").
			Append(&Alarm{
				Name:      parameters[0],
				Alias:     aliasName,
				Recipient: parameters[1],
				Parameter: common.StringToInt(parameters[2])}).
			Error; err != nil {
			tx.Rollback()
			return err
		}
		return nil
	})

}/*

//DeleteAlarmTransactions deletes an alarm from db during modification
//AutoUpdate is false, because otherwise we will be adding what we just deleted
func DeleteAlarmTransactions(aliasID int, alarm string) error {
	return WithinTransaction(func(tx *gorm.DB) (err error) {
		//parametes[0] --> alarm name ; parameters[1] --> recipient;
		//parameters[1] --> threshold parameter
		parameters := strings.Split(alarm, ":")
		if err = tx.Set("gorm:association_autoupdate", false).
			Where("alarm_alias_id = ? AND name = ?", aliasID, parameters[0]).
			Delete(&Alarm{}).
			Error; err != nil {
			tx.Rollback()
			return err
		}
		return nil

	})

}/*

// WithinTransaction  accept dBFunc as parameter call dBFunc function within transaction begin, and commit and return error from dBFunc
func WithinTransaction(fn dBFunc) (err error) {
	tx := cgorm.ManagerDB().Begin() // start db transaction
	defer tx.Commit()
	err = fn(tx)

	if err != nil {
	}
	return err

}*/
/*
//PrepareRelation prepares an entry for the node-alias table
func prepareRelation(nodeID int, aliasID int, p bool) (r *Relation) {
	r = &Relation{AliasID: aliasID, NodeID: nodeID, Blacklist: p}
	return r
}
*/