package lbclient

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"gitlab.cern.ch/lb-experts/goermis/api/ermis"
	"gitlab.cern.ch/lb-experts/goermis/auth"
	"gitlab.cern.ch/lb-experts/goermis/db"
)

func (lbclient *LBClient) findUnregistered() (unregistered []string, err error) {
	var (
		aliases []string
		intf    ermis.PrivilegeIntf
	)
	log.Infof("checking if node %v is registered on the aliases it reported %v ", lbclient.NodeName, lbclient.Status)
	for _, v := range lbclient.Status {
		aliases = append(aliases, v.AliasName)
	}
	db.GetConn().Preload("Relations.Node").
		Where("alias_name IN ?", aliases).
		Find(&lbclient.Aliases)
	if len(lbclient.Aliases) == 0 {
		return nil, fmt.Errorf("there are no aliases for node %v", lbclient.NodeName)
	}
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
			if auth.CheckLbclientAuth(lbclient.NodeName, status.Secret){
				err := fmt.Errorf("unauthorized to register the load for node %v and alias %v, secret missmatch", lbclient.NodeName, status.AliasName)
				return http.StatusUnauthorized, err
			}

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
	log.Infof("done with the registration for node %v. exiting...", lbclient.NodeName)
	return http.StatusCreated, nil
}

func (lbclient LBClient) updateNode() (int, error) {
	log.Infof("started registration procedure for node %v in every alias", lbclient.NodeName)
	for _, alias := range lbclient.Aliases {
		status := lbclient.findStatus(alias.AliasName)

		log.Infof("checking if node %v is authorized to update alias %v", lbclient.NodeName, alias.AliasName)
		if auth.CheckLbclientAuth(lbclient.NodeName, status.Secret){
			err := fmt.Errorf("unauthorized to update the load for node %v and alias %v, secret missmatch", lbclient.NodeName, status.AliasName)
			return http.StatusUnauthorized, err
		}
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
			} else {
				log.Errorf("could not find the relation between alias %v and node %v, while updating load", alias, lbclient.NodeName)
			}

		}

	}
	log.Infof("successful load update for node %v. exiting...", lbclient.NodeName)
	return http.StatusOK, nil
}

/*
func (lbclient LBClient) saltAndHash(status Status) []byte {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(status.Secret), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Failed to salt & hash the secret for node %v", lbclient.NodeName)
	}
	return hashedPassword
}*/
