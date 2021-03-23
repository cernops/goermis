package lbclient

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/labstack/echo/v4"
	"gitlab.cern.ch/lb-experts/goermis/api/ermis"
	"gitlab.cern.ch/lb-experts/goermis/bootstrap"
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


