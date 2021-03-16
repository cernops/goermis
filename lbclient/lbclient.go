package lbclient

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"gitlab.cern.ch/lb-experts/goermis/db"
	"gitlab.cern.ch/lb-experts/goermis/ermis"
)

type LBClient struct {
	Host   ermis.Node
	Status map[string]int
}

func UpdateLBClient(c echo.Context) error {
	var lbc LBClient
	lbc.Host.NodeName = "node1.cern.ch"

	//var a []ermis.Alias
	//nodes := ermis.Node{NodeName: node}
	if err := c.Bind(&lbc.Status); err != nil {
		fmt.Println("Failed in Bind for handler Update LBClient")
	}

	if !nodeExists(lbc.Host.NodeName) {
		registerNode(lbc)
	}

	

	//db.GetConn().
	//	Preload("Aliases.Alias").Where("node_name=?", node).Find(&nodes)
	//db.GetConn().Model(&nodes).Association("Aliases").Find(&a)
	//spew.Dump(nodes)
	//populateAliases()
	//updateValues()
	spew.Dump(lbc.Status)
	return nil

}

func nodeExists(node string) bool {
	var result ermis.Node
	err := db.GetConn().Where("node_name=?", node).Find(&result).Error
	if err != nil {
		log.Error("Failed to check node existance in lbclient handlers with error: ", err)
		return false
	}
	if result.NodeName == "" {
		return false
	}

	return true

}
func registerNode(lbc LBClient) bool {
	db.GetConn().
	return true
}

func GetAll(c echo.Context) error {

	return nil
}
