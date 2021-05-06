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
	Secret    string
	Load      int
}

func PostHandler(c echo.Context) error {
	var (
		lbclient LBClient
	)

	if err := c.Bind(&lbclient.Status); err != nil {
		return messageToNode(http.StatusBadRequest, fmt.Sprintf("failed in Bind for handler Update LBClient, error:%v", err))

	}
	//set nodename
	lbclient.NodeName = c.Request().Header.Get("NameFromCert")
	//make sure there is a nodename
	if lbclient.NodeName == "" {
		return messageToNode(http.StatusBadRequest, fmt.Sprint("Nodename cannot be empty"))
	}
	log.Infof("node %v sent its status, first lets retrieve the aliases from db", lbclient.NodeName)
	//retrieve reported aliases from database
	if err := lbclient.findAliases(); err != nil {
		return messageToNode(http.StatusBadRequest, fmt.Sprintf("failed to find reported aliases in database; %v", err))
	}
	//make sure the reported aliases exist
	if len(lbclient.Aliases) == 0 {
		return messageToNode(http.StatusNotFound, fmt.Sprintf("there are no aliases for node %v", lbclient.NodeName))
	}

	log.Infof(" checking if node %v its registered on every alias", lbclient.NodeName)
	unreg, err := lbclient.findUnregistered()
	if err != nil {
		return messageToNode(http.StatusBadRequest, fmt.Sprintf("error while looking for the aliases where node %v is unregistered, error: %v", lbclient.NodeName, err))
	}

	if len(unreg) != 0 {
		log.Infof("preparing to register node %v in the following aliases:%v", lbclient.NodeName, unreg)
		if status, err := lbclient.registerNode(unreg); err != nil {
			return messageToNode(status, fmt.Sprintf("error while registering node %v error: %v", lbclient.NodeName, err))
		}
	} else {
		log.Infof("node %v is registered on every alias it reported, lets proceed with the load update", lbclient.NodeName)
		if status, err := lbclient.updateNode(); err != nil {
			return messageToNode(status, fmt.Sprintf("error while updating load for node %v with error %v", lbclient.NodeName, err))
		}
	}
	/*prepare to report for missing aliases in both sides*/

	//check for reported aliases, missing from db
	missingfromdb := lbclient.missingfromdb()

	//check for db aliases, missing from status report
	missingfromreport := lbclient.missingfromstatus()

	return messageToNode(http.StatusOK, fmt.Sprintf("process completed for node %v.\n Aliases not found in database: %v\nAliases which did not receive an update:%v", lbclient.NodeName, missingfromdb, missingfromreport))

}
func messageToNode(status int, message string) error {

	if 200 <= status && status < 300 {
		log.Info(message)
	} else {
		log.Error(message)
	}
	return echo.NewHTTPError(status, message)
}
