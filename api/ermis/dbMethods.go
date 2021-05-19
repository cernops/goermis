package ermis

/* This file includes the ORM models and its methods*/

import (
	"errors"
	"fmt"
	"net/smtp"

	"gitlab.cern.ch/lb-experts/goermis/auth"
	"gitlab.cern.ch/lb-experts/goermis/bootstrap"
	"gitlab.cern.ch/lb-experts/goermis/db"
)

var (
	cfg = bootstrap.GetConf() //Getting an instance of config params
)



//GetObjects return list of aliases if no parameters are passed or a single alias if parameters are given
func GetObjects(param string) (query []Alias, err error) {

	//Preload bottom-to-top, starting with the Relations & Nodes first
	nodes := db.GetConn().Preload("Relations.Node") //Relations
	if param == "all" {                             //get all aliases
		err = nodes.
			Preload("Cnames").
			Preload("Alarms").
			Order("alias_name").
			Find(&query).Error

	} else { //get only the specified one
		err = nodes.
			Preload("Cnames").
			Preload("Alarms").
			Where("id=?", param).Or("alias_name=?", param).
			Order("alias_name").
			Find(&query).Error

	}
	if err != nil {
		return nil, errors.New("Failed in query: " + err.Error())

	}
	return query, nil

}

////////////////////////ALIAS METHODS////////////////////////////////

//CreateObjectInDB creates an alias
func (alias Alias) createObjectInDB() (err error) {

	//Create object in the DB with transactions, if smth goes wrong its rolledback
	if err := CreateTransactions(alias); err != nil {
		return err
	}

	return nil

}

//deleteObject deletes an alias and its Relations
func (alias Alias) deleteObjectInDB() (err error) {
	//Delete from DB
	if err := deleteTransactions(alias); err != nil {
		return err
	}
	return nil

}

//UpdateAlias modifies aliases and its associations
func (alias Alias) updateAlias() (err error) {
	if err := aliasUpdateTransactions(alias); err != nil {
		return err
	}

	return nil
}

//updateNodes updates alias with new nodes
func (alias Alias) updateNodes() (err error) {
	var (
		relationsInDB []Relation
		intf          PrivilegeIntf
	)
	//Let's find the registered nodes for this alias
	db.GetConn().Preload("Node").Where("alias_id=?", alias.ID).Find(&relationsInDB)

	for _, r := range relationsInDB {
		intf = r
		if ok, _ := Compare(intf, alias.Relations); !ok {
			if err = deleteNodeTransactions(r); err != nil {
				return errors.New("Failed to delete existing node " +
					r.Node.NodeName + " while updating, with error: " + err.Error())
			}
		}
	}
	for _, r := range alias.Relations {
		intf = r
		if ok, _ := Compare(intf, relationsInDB); !ok {
			if err = AddNodeTransactions(r); err != nil {
				return errors.New("Failed to add new node " +
					r.Node.NodeName + " while updating, with error: " + err.Error())
			}
			//If relation exists we also check if user modified its privileges
		} else if ok, privilege := Compare(intf, relationsInDB); ok && !privilege {
			if err = updatePrivilegeTransactions(r); err != nil {
				return errors.New("Failed to update privilege for node " +
					r.Node.NodeName + " while updating, with error: " + err.Error())
			}
		}
	}

	return nil

}

//Update the cnames
//updateCnames updates cnames in DB
func (alias Alias) updateCnames() (err error) {
	var (
		cnamesInDB []Cname
		intf       ContainsIntf
	)
	//Let's see what cnames are already registered for this alias
	db.GetConn().Model(&alias).Association("Cnames").Find(&cnamesInDB)

	if len(alias.Cnames) > 0 { //there are cnames, delete and add accordingly
		for _, v := range cnamesInDB {
			intf = v
			if !Contains(intf, alias.Cnames) {
				if err = deleteCnameTransactions(v); err != nil {
					return errors.New("Failed to delete existing cname " +
						v.Cname + " while updating, with error: " + err.Error())
				}
			}
		}

		for _, v := range alias.Cnames {
			intf = v
			if !Contains(intf, cnamesInDB) {
				if err = addCnameTransactions(v); err != nil {
					return errors.New("Failed to add new cname " +
						v.Cname + " while updating, with error: " + err.Error())
				}
			}

		}

	} else { //user deleted everything, so do we
		for _, v := range cnamesInDB {
			if err = deleteCnameTransactions(v); err != nil {
				return errors.New("Failed to delete cname " +
					v.Cname + " while purging all, with error: " + err.Error())
			}
		}
	}
	return nil
}

//Update the alarms
func (alias Alias) updateAlarms() (err error) {
	var (
		alarmsInDB []Alarm
		intf       ContainsIntf
	)
	//Let's see what alarms are already registered for this alias
	db.GetConn().Model(&alias).Association("Alarms").Find(&alarmsInDB)
	if len(alias.Alarms) > 0 {
		for _, a := range alarmsInDB {
			intf = a
			if !Contains(intf, alias.Alarms) {
				if err = deleteAlarmTransactions(a); err != nil {
					return errors.New("Failed to delete existing alarm " +
						a.Name + " while updating, with error: " + err.Error())
				}
			}
		}

		for _, a := range alias.Alarms {
			intf = a
			if !Contains(intf, alarmsInDB) {
				if err = addAlarmTransactions(a); err != nil {
					return errors.New("Failed to add alarm " +
						a.Name + ":" +
						a.Recipient + ":" +
						fmt.Sprint(a.Parameter) +
						" while purging all, with error: " +
						err.Error())
				}
			}

		}

	} else {
		for _, a := range alarmsInDB {
			if err = deleteAlarmTransactions(a); err != nil {
				return errors.New("Failed to delete alarm " +
					a.Name + ":" +
					a.Recipient + ":" +
					fmt.Sprint(a.Parameter) +
					" while purging all, with error: " +
					err.Error())
			}
		}
	}
	return nil
}
//createSecret in tbag
func (alias Alias) createSecret() error {
	newsecret := generateRandomSecret()
	err := auth.PostSecret(alias.AliasName, newsecret)
	if err != nil {
		return err
	}
	return alias.sendSecretToUser(newsecret)

}

//deleteSecret from tbag
func (alias Alias) deleteSecret() error {
	return auth.DeleteSecret(alias.AliasName)
}

//sendNotification sends an e-mail to the recipient when alarm is triggered
func (alias Alias) sendSecretToUser(secret string) error {
	recipient := alias.User + "@cern.ch"
	log.Infof("Sending the new secret of alias %v to %v", alias.AliasName, alias.User)
	msg := []byte("To: " + alias.User + "\r\n" +
		fmt.Sprintf("Subject: New secret created for alias %s: Please provide this to the nodes behind that alias. If not sure, check documentation(https://configdocs.web.cern.ch)\nSecret: %s ", alias.AliasName, secret))

	err := smtp.SendMail("localhost:25",
		nil,
		"lbd@cern.ch",
		[]string{recipient},
		msg)
	return err
}


