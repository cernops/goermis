package lbclient

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"gitlab.cern.ch/lb-experts/goermis/api/ermis"
	"gitlab.cern.ch/lb-experts/goermis/db"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func (lbclient LBClient) findUnregistered() (unregistered []string, err error) {
	var (
		aliases []string
		intf    ermis.PrivilegeIntf
	)
	for _, v := range lbclient.Status {
		aliases = append(aliases, v.AliasName)
	}
	result := db.GetConn().Preload("Relations.Node").
		Where("alias_name IN ?", aliases).
		Find(&lbclient.Aliases)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("there are no aliases for node %v.\nError: %v", lbclient.NodeName, result.Error)
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

func (lbclient LBClient) findStatus(s string) (status Status) {
	for _, v := range lbclient.Status {
		if v.AliasName == s {
			return v
		}
	}
	return Status{}
}

func (lbclient LBClient) registerNode(unreg []string) (int, error) {
	for _, alias := range lbclient.Aliases {
		if ermis.StringInSlice(alias.AliasName, unreg) {

			status := lbclient.findStatus(alias.AliasName)
			if compareSecret(status, alias.Secret, lbclient.NodeName) != nil {
				_, err := fmt.Printf("unauthorized to update the load for node %v and alias %v, secret missmatch", lbclient.NodeName, status.AliasName)
				return http.StatusUnauthorized, err
			}
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
			err := ermis.AddNodeTransactions(relation)
			if err != nil {
				return http.StatusBadRequest, err
			}

		}

	}

	return http.StatusCreated, nil
}

func (lbclient LBClient) updateNode() (int, error) {
	for _, alias := range lbclient.Aliases {
		status := lbclient.findStatus(alias.AliasName)
		if compareSecret(status, alias.Secret, lbclient.NodeName) != nil {
			_, err := fmt.Printf("unauthorized to update the load for node %v and alias %v, secret missmatch", lbclient.NodeName, status.AliasName)
			return http.StatusUnauthorized, err
		}
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
				log.Errorf("Could not find the relation between alias %v and node %v, while updating load", alias, lbclient.NodeName)
			}

		}

	}
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

func compareSecret(status Status, regSecret, node string) error {
	return bcrypt.CompareHashAndPassword([]byte(regSecret), []byte(status.Secret))

}
