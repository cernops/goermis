package lbclient

import (
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
		status   int = http.StatusOK
	)

	if err := c.Bind(&lbclient.Status); err != nil {
		log.Error("failed in Bind for handler Update LBClient")
	}
	lbclient.NodeName = "node1.cern.ch"
	log.Infof("node %v sent its status, first lets check if its registered on every alias", lbclient.NodeName)
	unreg, err := lbclient.findUnregistered()
	if err != nil {
		log.Errorf("error while looking for the aliases where node %v is unregistered: %v", lbclient.NodeName, err)
		status = http.StatusBadRequest
	}
	if len(unreg) != 0 {
		log.Infof("preparing to register node %v in the following aliases:%v", lbclient.NodeName, unreg)
		regstatus, err := lbclient.registerNode(unreg)
		if err != nil {
			log.Errorf("error while registering node %v error: %v", lbclient.NodeName, err)
			status = regstatus
		}
	} else {
		log.Infof("node %v is registered on every alias it reported, lets proceed with the load update", lbclient.NodeName)
		updstatus, err := lbclient.updateNode()
		if err != nil {
			log.Errorf("error while updating load for node %v with error %v", lbclient.NodeName, err)
			status = updstatus
		}
	}

	return c.NoContent(status)

}
