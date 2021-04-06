package lbclient

import (
	"fmt"
	"net/http"

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
		status   int
	)

	if err := c.Bind(&lbclient.Status); err != nil {
		fmt.Println("Failed in Bind for handler Update LBClient")
	}
	lbclient.NodeName = "node1.cern.ch"

	unreg, err := lbclient.findUnregistered()
	if err != nil {
		log.Errorf("error while looking for the aliases where node %v is unregistered: %v", lbclient.NodeName, err)
		status = http.StatusBadRequest
	}
	if len(unreg) != 0 {
		log.Info("Entered Registration")
		st, err := lbclient.registerNode(unreg)
		if err != nil {
			log.Errorf("error while registering node %v with the aliases %v\n%v", lbclient.NodeName, unreg, err)
			status = st
		}
	} else {
		log.Info("Entered Update")
		st, err := lbclient.updateNode()
		if err != nil {
			log.Errorf("error while updating load for node %v with error %v", lbclient.NodeName, err)
			status = st
		}
	}

	return c.JSON(status, err)

}
