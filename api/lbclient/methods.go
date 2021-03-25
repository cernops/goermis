package lbclient

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/davecgh/go-spew/spew"
	"gitlab.cern.ch/lb-experts/goermis/api/ermis"
	"gitlab.cern.ch/lb-experts/goermis/db"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func (lbclient *LBClient) findUnregistered() (unregistered []string, err error) {
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
				
		if ok, _ := ermis.CompareRelations(intf, x.Relations); !ok {
			unregistered = append(unregistered, x.AliasName)
		}
	}

	return unregistered, nil

}

func containsNode(s string, alias []ermis.Relation) bool {
	for _, v := range alias {
		if v.Node.NodeName == s {
			return true
		}
	}
	return false
}
func (lbclient *LBClient) findStatus(s string) (status Status) {
	for _, v := range lbclient.Status {
		if v.AliasName == s {
			return v
		}
	}
	return Status{}
}

func (lbclient *LBClient) registerNode(unreg []string) error {
	for _, alias := range lbclient.Aliases {
		if ermis.StringInSlice(alias.AliasName, unreg) {
			status := lbclient.findStatus(alias.AliasName)
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
				Secret: string(lbclient.saltAndHash(status)),
			}
			err := ermis.AddNodeTransactions(relation)
			if err != nil {
				return err
			}

		}

	}

	return nil
}

func (lbclient *LBClient) updateNode() {
	log.Info("This one")
	spew.Dump(lbclient)
	for _, alias := range lbclient.Aliases {
		for _, rel := range alias.Relations {
			status := lbclient.findStatus(alias.AliasName)
			if rel.Node.NodeName == lbclient.NodeName {

				db.GetConn().Select("load", "last_check", "secret").
					Where("alias_id=? AND node_id=?", alias.ID, rel.NodeID).Updates(
					ermis.Relation{
						Load:   status.Load,
						Secret: string(lbclient.saltAndHash(status)),
						LastCheck: sql.NullTime{
							Time:  time.Now(),
							Valid: true,
						}})

			}

		}
	}

}

func (lbclient *LBClient) saltAndHash(status Status) []byte {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(status.Secret), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Failed to salt & hash the secret for node %v", lbclient.NodeName)
	}
	return hashedPassword
}
