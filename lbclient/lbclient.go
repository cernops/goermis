package lbclient

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/labstack/echo/v4"
	"gitlab.cern.ch/lb-experts/goermis/bootstrap"
	"gitlab.cern.ch/lb-experts/goermis/db"
	"gitlab.cern.ch/lb-experts/goermis/ermis"
	"golang.org/x/crypto/bcrypt"
)

var (
	log = bootstrap.GetLog()
)

type Status struct {
	NodeName  string
	AliasName string
	Load      int
	Secret    string
	Alias     ermis.Alias
}

func PostHandler(c echo.Context) error {
	var lbc []Status
	if err := c.Bind(&lbc); err != nil {
		fmt.Println("Failed in Bind for handler Update LBClient")
	}
	for _, v := range lbc {
		if !v.isRegistered() {
			v.registerNode()
		} else {
			v.updateNode()
		}
	}

	spew.Dump(lbc)

	return nil

}
func (status Status) isRegistered() bool {

	result := db.GetConn().Preload("Relations.Node").Where("alias_name=?", status.Alias).Find(&status.Alias)

	if result.RowsAffected == 0 {
		log.Error("Alias %v sent by node %v does not exist", status.Alias, status.NodeName)
		return false
	}
	for _, q := range status.Alias.Relations {
		if q.Node.NodeName == status.NodeName {
			return true
		}
	}
	log.Info("Node %v is not registered in alias %v. We will proceed with the creation of the relation and a new node will be created if needed", status.NodeName, status.Alias)
	return false

}
func (status Status) registerNode() error {
	relation := ermis.Relation{
		AliasID:   status.Alias.ID,
		NodeID:    0,
		Blacklist: false,
		Node: &ermis.Node{
			ID:       0,
			NodeName: status.NodeName,
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
		Secret: string(status.saltAndHash()),
	}
	err := ermis.AddNodeTransactions(relation)
	if err != nil {
		return err
	}
	return nil
}

func (status Status) updateNode() {
	var r ermis.Relation
	for _, v := range status.Alias.Relations {
		if status.NodeName == v.Node.NodeName {
			r = ermis.Relation{
				AliasID: status.Alias.ID,
				NodeID:  v.NodeID,
				Load:    status.Load,
				LastCheck: sql.NullTime{
					Time:  time.Now(),
					Valid: true,
				},
			}

		}
	}

	db.GetConn().Save(&r)
}

func (status Status) saltAndHash() []byte {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(status.Secret), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Failed to salt & hash the secret for node %v", status.NodeName)
	}
	return hashedPassword
}
