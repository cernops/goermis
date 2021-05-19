package lbclient

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"gitlab.cern.ch/lb-experts/goermis/api/ermis"
	"gitlab.cern.ch/lb-experts/goermis/db"
)

func (lbclient *LBClient) findAliases() (err error) {
	var (
		reportedAliases []string
	)
	for _, v := range lbclient.Status {
		reportedAliases = append(reportedAliases, v.AliasName)
	}
	//store the found aliases in lbclient struct
	err = db.GetConn().Preload("Relations.Node").
		Where("alias_name IN ?", reportedAliases).
		Find(&lbclient.Aliases).Error
	if err != nil {
		return fmt.Errorf("error while retrieving the claimed aliases %v from node %v, with error %v", reportedAliases, lbclient.NodeName, err)
	}
	return nil
}

func (lbclient *LBClient) missingfromdb() []string {
	var (
		missingfromdb []string
	)

	for _, v1 := range lbclient.Status {
		found := false
		for _, v2 := range lbclient.Aliases {
			if v1.AliasName == v2.AliasName {
				found = true
				break
			}
		}
		// String not found. We add it to return slice
		if !found {
			missingfromdb = append(missingfromdb, v1.AliasName)
		}
	}
	return missingfromdb
}

func (lbclient *LBClient) missingfromstatus() []string {
	var (
		missingfromstatus []string
	)

	for _, v1 := range lbclient.Aliases {
		found := false
		for _, v2 := range lbclient.Status {
			if v1.AliasName == v2.AliasName {
				found = true
				break
			}
		}
		// String not found. We add it to return slice
		if !found {
			missingfromstatus = append(missingfromstatus, v1.AliasName)
		}
	}
	return missingfromstatus
}

func (lbclient *LBClient) findUnregistered() (unregistered []string, err error) {
	var (
		intf ermis.PrivilegeIntf
	)
	//compare the relations of every alias we retrieved with the node name
	for _, x := range lbclient.Aliases {
		intf = ermis.Relation{
			Node: &ermis.Node{
				NodeName: lbclient.NodeName}}

		if ok, _ := ermis.Compare(intf, x.Relations); !ok {
			unregistered = append(unregistered, x.AliasName)
		}
	}
	return unregistered, nil

}

func (lbclient *LBClient) findStatus(s string) (status Status) {
	for _, v := range lbclient.Status {
		if v.AliasName == s {
			return v
		}
	}
	return Status{}
}

func (lbclient *LBClient) registerNode(unreg []string) (int, error) {
	log.Infof("started registration procedure for node %v in aliases %v", lbclient.NodeName, unreg)
	for _, alias := range lbclient.Aliases {
		if ermis.StringInSlice(alias.AliasName, unreg) {
			status := lbclient.findStatus(alias.AliasName)

			log.Infof("checking if node %v is authorized to register on alias %v", lbclient.NodeName, alias.AliasName)
			/*if !checkLbclientAuth(status.AliasName, status.Secret) {
				err := fmt.Errorf("unauthorized to register the load for node %v and alias %v, secret missmatch", lbclient.NodeName, status.AliasName)
				return http.StatusUnauthorized, err
			}*/

			log.Infof("preparing the relation between node %v and alias %v", lbclient.NodeName, alias.AliasName)
			relation := ermis.Relation{
				AliasID:   alias.ID,
				NodeID:    0,
				Blacklist: false,
				Node: &ermis.Node{
					ID:       0,
					NodeName: lbclient.NodeName,
					LastModification: sql.NullTime{
						Time:  time.Now(),
						Valid: true,
					},
				},
				Load: status.Load,
				LastCheck: sql.NullTime{
					Time:  time.Now(),
					Valid: true,
				},
			}

			log.Infof("ready to create the relation for node %v and alias %v", lbclient.NodeName, alias.AliasName)
			err := ermis.AddNodeTransactions(relation)
			if err != nil {
				return http.StatusBadRequest, err
			} else {
				log.Infof("successful registration for node %v on alias %v", lbclient.NodeName, alias.AliasName)
			}

		}

	}
	log.Infof("successful registration for node %v with the latest load value. exiting node registration...", lbclient.NodeName)
	return http.StatusCreated, nil
}

func (lbclient LBClient) updateNode() (int, error) {
	log.Infof("started update procedure for node %v in every alias", lbclient.NodeName)
	for _, alias := range lbclient.Aliases {

		status := lbclient.findStatus(alias.AliasName)

		log.Infof("checking if node %v is authorized to update alias %v", lbclient.NodeName, alias.AliasName)
		/*if !checkLbclientAuth(status.AliasName, status.Secret) {
			err := fmt.Errorf("unauthorized to update the load for node %v and alias %v, secret missmatch", lbclient.NodeName, status.AliasName)
			return http.StatusUnauthorized, err
		}*/
		log.Infof("preparing to update the load for node %v and alias %v", lbclient.NodeName, alias.AliasName)
		for _, rel := range alias.Relations {
			if rel.Node.NodeName == lbclient.NodeName {
				err := db.GetConn().Select("load", "last_check").
					Where("alias_id=? AND node_id=?", alias.ID, rel.NodeID).Updates(
					ermis.Relation{
						Load: status.Load,
						LastCheck: sql.NullTime{
							Time:  time.Now(),
							Valid: true,
						}}).Error
				if err != nil {
					return http.StatusBadRequest, err
				}
				log.Infof("updated node %v on alias %v with new load value %v", lbclient.NodeName, status.AliasName, status.Load)
			}

		}

	}
	log.Infof("successful load update for node %v. exiting load update...", lbclient.NodeName)
	return http.StatusOK, nil
}

/*
func checkLbclientAuth(aliasname, secret string) bool {
	return ermis.StringInSlice(secret, auth.GetSecret(aliasname))
}*/
