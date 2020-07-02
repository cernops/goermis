package api

import (
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"gitlab.cern.ch/lb-experts/goermis/aiermis/orm"
	"gitlab.cern.ch/lb-experts/goermis/db"
)

func init() {
	govalidator.SetFieldsRequiredByDefault(true)
	customValidators()
}

//GetAliases returns a list of ALL aliases
func GetAliases(c echo.Context) error {

	var (
		list Objects
		e    error
	)
	username := c.Request().Header.Get("X-Forwarded-User")
	//If empty values provided,the MySQL query returns all aliases
	if list.Objects, e = GetObjects("", ""); e != nil {
		log.Error("[" + username + "] " + e.Error())
		return echo.NewHTTPError(http.StatusBadRequest, e.Error())
	}
	defer c.Request().Body.Close()
	log.Info("[" + username + "]" + " List of aliases retrieved successfully")
	return c.JSON(http.StatusOK, list)
}

//GetAlias queries for a specific alias
func GetAlias(c echo.Context) error {
	var (
		list     Objects
		e        error
		tablerow string
	)
	//Get the name/ID of alias
	username := c.Request().Header.Get("X-Forwarded-User")
	param := c.Param("alias")
	//Validate that the parameter is DNS-compatible
	if !govalidator.IsDNSName(param) {
		log.Error("[" + username + "] " + "Wrong type of query parameter.Expected alphanum, received " + param)
		return echo.NewHTTPError(http.StatusUnprocessableEntity)
	}

	//Switch between name and ID query(enables user to ask by name or id)
	if _, err := strconv.Atoi(c.Param("alias")); err == nil {
		tablerow = "id"
	} else {
		if !strings.HasSuffix(param, ".cern.ch") {
			param = param + ".cern.ch"
		}

		tablerow = "alias_name"
	}

	if list.Objects, e = GetObjects(string(param), tablerow); e != nil {
		log.Error("[" + username + "]" + "Unable to get alias" + param + " : " + e.Error())
		return echo.NewHTTPError(http.StatusBadRequest, e.Error())
	}
	defer c.Request().Body.Close()
	log.Info("[" + username + "] " + "Alias" + param + " retrieved successfully")
	return c.JSON(http.StatusOK, list)

}

//CreateAlias creates a new alias entry in the DB
func CreateAlias(c echo.Context) error {
	username := c.Request().Header.Get("X-Forwarded-User")
	//Struct r serves for getting request values and validate them
	var r Resource
	if err := c.Bind(&r); err != nil {
		log.Warn("[" + username + "] " + "Failed to bind params " + err.Error())
	}

	//User is provided in the HEADER
	r.User = c.Request().Header.Get("X-Forwarded-User")
	defer c.Request().Body.Close()

	//Validate structure
	if ok, err := govalidator.ValidateStruct(r); err != nil || ok == false {
		return MessageToUser(c, http.StatusUnprocessableEntity,
			"Validation error for "+r.AliasName+" : "+err.Error(), "home.html")
	}

	//Default values and hydrate(domain,visibility)
	r.DefaultAndHydrate()

	log.Info("[" + username + "] " + "Ready to create a new alias " + r.AliasName)
	//Create object
	if err := r.CreateObject(); err != nil {
		return MessageToUser(c, http.StatusBadRequest,
			"Creation error for "+r.AliasName+" : "+err.Error(), "home.html")
	}

	return MessageToUser(c, http.StatusCreated,
		r.AliasName+" created successfully ", "home.html")

}

//DeleteAlias deletes the requested alias from the DB
func DeleteAlias(c echo.Context) error {
	username := c.Request().Header.Get("X-Forwarded-User")
	aliasName := c.FormValue("alias_name")
	//Validate alias name only
	if !govalidator.IsDNSName(aliasName) {
		log.Warn("[" + username + "] " + "Wrong type of query parameter, expected Alias name")
		return echo.NewHTTPError(http.StatusUnprocessableEntity)
	}

	alias, err := GetObjects(aliasName, "alias_name")
	if err != nil {
		log.Error("[" + username + "] " + "Failed to retrieve alias " + aliasName + " : " + err.Error())
	}
	defer c.Request().Body.Close()
	if err := alias[0].DeleteObject(); err != nil {
		return MessageToUser(c, http.StatusBadRequest, err.Error(), "home.html")

	}

	return MessageToUser(c, http.StatusCreated,
		aliasName+" deleted successfully ", "home.html")

}

//LBWEB-SPECIFIC HANDLERS

//ModifyAlias modifes cnames, nodes, hostgroup and best_hosts parameters
func ModifyAlias(c echo.Context) error {
	var r Resource
	username := c.Request().Header.Get("X-Forwarded-User")

	if err := c.Bind(&r); err != nil {
		log.Warn("[" + username + "] " + "Failed to bind params " + err.Error())
	}
	//After that, we use the alias name for retrieving its profile from DB
	existingObj, err := GetObjects(r.AliasName, "alias_name")
	if err != nil {
		log.Error("[" + username + "] " + "Failed to retrieve alias " + r.AliasName + " : " + err.Error())
		return err
	}
	//Preserve the DB values for these fields, because they are needed for comparison
	stashC := existingObj[0].Cname          //Cnames
	stashA := existingObj[0].AllowedNodes   //AllowedNodes
	stashF := existingObj[0].ForbiddenNodes //ForbiddenNodes

	//Update fields if changed. Serves kermis PATCH and UI forms
	if r.External != "" {
		existingObj[0].External = r.External
	}
	if r.BestHosts != 0 {
		existingObj[0].BestHosts = r.BestHosts
	}
	if r.Metric != "" {
		existingObj[0].Metric = r.Metric
	}
	if r.PollingInterval != 0 {
		existingObj[0].PollingInterval = r.PollingInterval
	}
	if r.Hostgroup != "" {
		existingObj[0].Hostgroup = r.Hostgroup
	}
	if r.Tenant != "" {
		existingObj[0].Tenant = r.Tenant
	}

	if r.TTL != 0 {
		existingObj[0].TTL = r.TTL
	}
	//These three fields are updated even if value is empty
	//because empty values are part of the update in their case
	existingObj[0].ForbiddenNodes = r.ForbiddenNodes
	existingObj[0].AllowedNodes = r.AllowedNodes
	existingObj[0].Cname = r.Cname

	//Validate
	if ok, err := govalidator.ValidateStruct(existingObj[0]); err != nil || ok == false {
		return MessageToUser(c, http.StatusUnprocessableEntity,
			"Validation error for alias "+existingObj[0].AliasName+" : "+err.Error(), "home.html")
	}

	defer c.Request().Body.Close()

	// Call the modifier
	if err := existingObj[0].ModifyObject(stashC, stashA, stashF); err != nil {
		return MessageToUser(c, http.StatusBadRequest,
			"Update error for alias "+existingObj[0].AliasName+" : "+err.Error(), "home.html")
	}

	return MessageToUser(c, http.StatusOK,
		c.FormValue("alias_name")+" updated Successfully"+err.Error(), "home.html")

}

//CheckNameDNS checks if an alias or cname already exist in DB or DNS server
func CheckNameDNS(c echo.Context) error {
	var (
		result int
		con    = db.ManagerDB()
	)

	aliasToResolve := c.QueryParam("hostname")
	con.Model(&orm.Cname{}).Where("c_name=?", aliasToResolve).Count(&result)
	if result == 0 {
		con.Model(&orm.Alias{}).Where("alias_name=?", aliasToResolve+".cern.ch").Count(&result)
	}
	if result == 0 {
		if r, err := net.LookupHost(aliasToResolve); err != nil {
			result = 0
		} else {
			result = len(r)

			return MessageToUser(c, http.StatusConflict,
				"Duplicate for "+aliasToResolve+" in DNS "+err.Error(), "home.html")

		}
	}
	return c.JSON(http.StatusOK, result)
}
