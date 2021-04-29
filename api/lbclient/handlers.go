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
	Secret    string
	Load      int
}

func PostHandler(c echo.Context) error {
	var (
		lbclient LBClient
	)

	if err := c.Bind(&lbclient.Status); err != nil {
		log.Error("failed in Bind for handler Update LBClient")
	}
	lbclient.NodeName = c.Request().Header.Get("NameFromCert")
	if lbclient.NodeName == "" {
		log.Error("Nodename cannot be empty")
		return c.NoContent(http.StatusBadRequest)
	}
	log.Infof("node %v sent its status, first lets check if its registered on every alias", lbclient.NodeName)
	unreg, err := lbclient.findUnregistered()
	if err != nil {
		log.Errorf("error while looking for the aliases where node %v is unregistered: %v", lbclient.NodeName, err)
		return c.NoContent(http.StatusBadRequest)
	}
	if len(unreg) != 0 {
		log.Infof("preparing to register node %v in the following aliases:%v", lbclient.NodeName, unreg)
		status, err := lbclient.registerNode(unreg)
		if err != nil {
			log.Errorf("error while registering node %v error: %v", lbclient.NodeName, err)
			return c.NoContent(status)
		}
	} else {
		log.Infof("node %v is registered on every alias it reported, lets proceed with the load update", lbclient.NodeName)
		status, err := lbclient.updateNode()
		if err != nil {
			log.Errorf("error while updating load for node %v with error %v", lbclient.NodeName, err)
			return c.NoContent(status)
		}
	}

	return c.NoContent(http.StatusOK)

}
