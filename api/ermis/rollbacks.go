package ermis

import (
	"fmt"
)

//DNSCreateRollback deletes new alias from DB, if DNS creation failed
func (alias Alias) RollbackInCreate(db, dns bool) error {
	//We dont know the newly assigned ID for our alias
	//We need the ID for clearing its associations
	/*****************REVISIT HOW WE FIND ID***********/
	if db {
		newlycreated, err := GetObjects(alias.AliasName)
		if err != nil {
			return fmt.Errorf("could not find orphan alias %v in DB after failing to create it in DNS, with error %v", alias.AliasName, err)
		}

		//discard new alias entry from DB
		err = newlycreated[0].deleteObjectInDB()
		if err != nil {
			return fmt.Errorf("failed to delete orphan alias %v from DB after DNS creation failure, with error: %v ", alias.AliasName, err)

		}
	}
	if dns {
		err := alias.deleteFromDNS()
		if err != nil {
			return fmt.Errorf("failed to delete alias %v from DNS after secret creation failure, error: %v", alias.AliasName, err)
		}

	}

	return nil

}
func (alias Alias) RollbackInDelete(db, dns bool) error {
	if db {
		//If deletion from DNS fails, we recreate the object in DB.
		err := alias.createObjectInDB()
		if err != nil {
			return fmt.Errorf("[%v] failed to recreate alias %v in database, as part of the creation rollback, error %v", alias.User, alias.AliasName, err)
		}
	}
	if dns {
		err := alias.createInDNS()
		if err != nil {
			return fmt.Errorf("[%v] failed to recreate alias %v in DNS, as part of the deletion rollback, error %v", alias.User, alias.AliasName, err)
		}
	}

	return nil
}

func (alias Alias) RollbackInModify(oldstate Alias) error {
	//Delete the DB updates we just made
	if err := alias.deleteObjectInDB(); err != nil {
		return fmt.Errorf("[%v] failed to clean the new updates while rolling back alias %v", alias.User, alias.AliasName)
	}
	//Recreate the alias as it was before the update
	if err := oldstate.createObjectInDB(); err != nil {

		return fmt.Errorf("[%v] failed to restore previous state for alias %v, during rollback", alias.User, alias.AliasName)
	}
	return nil
}
