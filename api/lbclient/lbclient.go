package lbclient

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/labstack/echo/v4"
	"gitlab.cern.ch/lb-experts/goermis/bootstrap"
	"gitlab.cern.ch/lb-experts/goermis/db"
	"gitlab.cern.ch/lb-experts/goermis/api/ermis"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	log = bootstrap.GetLog()
)

type LBClient struct {
	NodeName string
	Status   []Status
	Aliases  []ermis.Alias
}
type Status struct {
	AliasName string
	Load      int
	Secret    string
}

func PostHandler(c echo.Context) error {
	var (
		lbclient LBClient
	)

	if err := c.Bind(&lbclient.Status); err != nil {
		fmt.Println("Failed in Bind for handler Update LBClient")
	}
	lbclient.NodeName = "node1.cern.ch"

	unreg, err := lbclient.findUnregistered()
	if err != nil {
		return err
	}
	log.Info("UNR")
	spew.Dump(unreg)

	log.Info("LBCLIENT")
	spew.Dump(lbclient)

	if len(unreg) != 0 {
		log.Info("Entered Registration")
		lbclient.registerNode(unreg)
	} else {
		log.Info("Entered Update")
		lbclient.updateNode()
	}

	return nil

}

func (lbclient *LBClient) findUnregistered() (unregistered []string, err error) {
	var (
		aliases []string
	)
	for _, v := range lbclient.Status {
		aliases = append(aliases, v.AliasName)
	}
	log.Info("ALIASES")
	spew.Dump(aliases)
	result := db.GetConn().Preload("Relations.Node").
		Where("alias_name IN ?", aliases).
		Find(&lbclient.Aliases)

	log.Info("ALIASESII")
	spew.Dump(lbclient.Aliases)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("there are no aliases for node %v.\nError: %v", lbclient.NodeName, result.Error)
	}

	for _, x := range lbclient.Aliases {
		if !containsNo(lbclient.NodeName, x.Relations) {
			unregistered = append(unregistered, x.AliasName)
		}
	}

	return unregistered, nil

}

func containsNo(s string, alias []ermis.Relation) bool {
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
