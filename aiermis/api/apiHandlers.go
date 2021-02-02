package api

import (
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"

	"gitlab.cern.ch/lb-experts/goermis/aiermis/common"
	"gitlab.cern.ch/lb-experts/goermis/db"
)

func init() {
	govalidator.SetFieldsRequiredByDefault(true)
	common.CustomValidators()
}

//GetAlias returns a list of ALL aliases
func GetAlias(c echo.Context) error {

	var (
		queryResults []Alias
		e            error
	)
	username := GetUsername()
	param := c.QueryParam("alias_name")
	if param == "" {
		//If empty values provided,the MySQL query returns all aliases
		if queryResults, e = GetObjects("all"); e != nil {
			log.Error("[" + username + "] " + e.Error())
			return echo.NewHTTPError(http.StatusBadRequest, e.Error())
		}
	} else {
		//Validate that the parameter is DNS-compatible
		if !govalidator.IsDNSName(param) {
			log.Error("[" + username + "] " + "Wrong type of query parameter.Expected alphanum, received " + param)
			return echo.NewHTTPError(http.StatusBadRequest)
		}

		if _, err := strconv.Atoi(param); err != nil {
			if !strings.HasSuffix(param, ".cern.ch") {
				param = param + ".cern.ch"
			}
		}

		if queryResults, e = GetObjects(string(param)); e != nil {
			log.Error("[" + username + "]" + "Unable to get alias" + param + " : " + e.Error())
			return echo.NewHTTPError(http.StatusBadRequest, e.Error())
		}

	}
	defer c.Request().Body.Close()
	return c.JSON(http.StatusOK, parse(queryResults))
}

//CreateAlias creates a new alias entry in the DB
func CreateAlias(c echo.Context) error {
	var temp Resource
	username := GetUsername()
	if err := c.Bind(&temp); err != nil {
		log.Warn("[" + username + "] " + "Failed to bind params " + err.Error())
	}
	temp.User = username
	defer c.Request().Body.Close()
	//Validate structure
	if ok, err := govalidator.ValidateStruct(temp); err != nil || ok == false {
		return common.MessageToUser(c, http.StatusBadRequest,
			"Validation error for "+temp.AliasName+" : "+err.Error(), "home.html")
	}

	//Check for duplicates
	alias, _ := GetObjects(temp.AliasName)
	if len(alias) != 0 {
		return common.MessageToUser(c, http.StatusConflict,
			"Alias "+temp.AliasName+" already exists ", "home.html")

	}

	object := sanitazeInCreation(temp)

	log.Info("[" + username + "] " + "Ready to create a new alias " + temp.AliasName)
	//Create object
	if err := object.CreateObject(); err != nil {
		return common.MessageToUser(c, http.StatusBadRequest,
			"Creation error for "+temp.AliasName+" : "+err.Error(), "home.html")
	}

	return common.MessageToUser(c, http.StatusCreated,
		temp.AliasName+" created successfully ", "home.html")

}

/*
//DeleteAlias deletes the requested alias from the DB
func DeleteAlias(c echo.Context) error {
	var (
		aliasToDelete string
	)
	username := GetUsername()

	switch c.Request().Header.Get("Content-Type") {
	case "application/json":
		aliasToDelete = c.QueryParam("alias_name")
	case "application/x-www-form-urlencoded":
		aliasToDelete = c.FormValue("alias_name")

	}
	//Validate alias name only, since the rest of the struct will be empty when DELETE
	if !govalidator.IsDNSName(aliasToDelete) {
		log.Warn("[" + username + "] " + "Wrong type of query parameter, expected Alias name, got :" + aliasToDelete)
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	alias, err := GetObjects(aliasToDelete)
	if err != nil {
		log.Error("[" + username + "] " + "Failed to retrieve alias " + aliasToDelete + " : " + err.Error())
	}
	defer c.Request().Body.Close()

	if alias != nil {
		if err := alias[0].DeleteObject(); err != nil {
			return common.MessageToUser(c, http.StatusBadRequest, err.Error(), "home.html")

		}
		return common.MessageToUser(c, http.StatusOK,
			aliasToDelete+" deleted successfully ", "home.html")

	}
	return common.MessageToUser(c, http.StatusNotFound, aliasToDelete+" not found", "home.html")

}
*/

//ModifyAlias modifes cnames, nodes, hostgroup and best_hosts parameters
func ModifyAlias(c echo.Context) error {
	//temp Resource is used to bind parameters from the request
	//Kermis allows us to change lots of fields.Since we don't know what
	//fields change each time, we get the existing object from DB and update
	//only the changed fields one-by-one.
	var (
		param string
		temp  Resource
	)
	username := GetUsername()
	//Bind request to the temp Resource
	if err := c.Bind(&temp); err != nil {
		log.Warn("[" + username + "] " + "Failed to bind params " + err.Error())
	}
	temp.User = username
	//Here we distignuish between kermis PATCH and UI form binding
	if c.Request().Method == "PATCH" {
		param = c.Param("id")
	} else {
		param = temp.AliasName
	}

	//We use the alias name for retrieving its profile from DB
	alias, err := GetObjects(param)
	if err != nil {
		log.Error("[" + username + "] " + "Failed to retrieve alias " + temp.AliasName + " : " + err.Error())
		return err
	}
	sanitazed := sanitazeInUpdate(alias[0], temp)

	//Validate the object alias , with the now-updated fields
	//if ok, err := govalidator.ValidateStruct(alias[0]); err != nil || ok == false {
	//	return common.MessageToUser(c, http.StatusBadRequest,
	//		"Validation error for alias "+alias[0].AliasName+" : "+err.Error(), "home.html")
	//}

	defer c.Request().Body.Close()

	// Call the modifier
	if err := sanitazed.ModifyObject(); err != nil {
		return common.MessageToUser(c, http.StatusBadRequest,
			"Update error for alias "+sanitazed.AliasName+" : "+err.Error(), "home.html")
	}

	return common.MessageToUser(c, http.StatusAccepted,
		sanitazed.AliasName+" updated Successfully", "home.html")

}

//CheckNameDNS checks if an alias or cname already exist in DB or DNS server
func CheckNameDNS(c echo.Context) error {
	var (
		result int64

		con = db.ManagerDB()
	)

	aliasToResolve := c.QueryParam("hostname")
	//Search cnames with the same name
	con.Model(&Cname{}).Where("cname=?", aliasToResolve).Count(&result)
	if result == 0 {
		//Search aliases
		con.Model(&Alias{}).Where("alias_name=?", aliasToResolve+".cern.ch").Count(&result)
	}
	if result == 0 {
		r, _ := net.LookupHost(aliasToResolve)
		result = int64(len(r))
	}
	return c.JSON(http.StatusOK, result)
}
