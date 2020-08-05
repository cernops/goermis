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
	//log.Info("[" + username + "]" + " List of aliases retrieved successfully")
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
		return echo.NewHTTPError(http.StatusBadRequest)
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
	//log.Info("[" + username + "] " + "Alias" + param + " retrieved successfully")
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
	r.User = username
	defer c.Request().Body.Close()

	//Validate structure
	if ok, err := govalidator.ValidateStruct(r); err != nil || ok == false {
		return MessageToUser(c, http.StatusBadRequest,
			"Validation error for "+r.AliasName+" : "+err.Error(), "home.html")
	}

	//Default values and hydrate(domain,visibility)
	r.DefaultAndHydrate()
	//Check for duplicates
	alias, _ := GetObjects(r.AliasName, "alias_name")
	if alias != nil {
		return MessageToUser(c, http.StatusConflict,
			"Alias "+r.AliasName+" already exists ", "home.html")

	}

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
	var r Resource

	//Bind request
	if err := c.Bind(&r); err != nil {
		log.Warn("[" + username + "] " + "Failed to bind params " + err.Error())
	}

	//User is provided in the HEADER
	r.User = username
	defer c.Request().Body.Close()
	//Validate alias name only, since the rest of the struct will be empty when DELETE
	if !govalidator.IsDNSName(r.AliasName) {
		log.Warn("[" + username + "] " + "Wrong type of query parameter, expected Alias name, got :" + r.AliasName)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	alias, err := GetObjects(r.AliasName, "alias_name")
	if err != nil {
		log.Error("[" + username + "] " + "Failed to retrieve alias " + r.AliasName + " : " + err.Error())
	}
	defer c.Request().Body.Close()

	if alias != nil {
		if err := alias[0].DeleteObject(); err != nil {
			return MessageToUser(c, http.StatusBadRequest, err.Error(), "home.html")

		}
		return MessageToUser(c, http.StatusOK,
			r.AliasName+" deleted successfully ", "home.html")

	}
	return MessageToUser(c, http.StatusNotFound, r.AliasName+" not found", "home.html")

}

//ModifyAlias modifes cnames, nodes, hostgroup and best_hosts parameters
func ModifyAlias(c echo.Context) error {
	//temp Resource is used to bind parameters from the request
	//Kermis allows us to change lots of fields.Since we don't know what
	//fields change each time, we get the existing object from DB and update
	//only the changed fields one-by-one.
	var temp Resource
	username := c.Request().Header.Get("X-Forwarded-User")
	//Bind request to the temp Resource
	if err := c.Bind(&temp); err != nil {
		log.Warn("[" + username + "] " + "Failed to bind params " + err.Error())
	}
	//After we bind request, we use the alias name for retrieving its profile from DB
	alias, err := GetObjects(temp.AliasName, "alias_name")
	if err != nil {
		log.Error("[" + username + "] " + "Failed to retrieve alias " + temp.AliasName + " : " + err.Error())
		return err
	}

	//UPDATE changed fields in the retrieved struct for that alias.
	//This helps in validation, since we don't know what fields are changing every time
	if temp.External != "" {
		alias[0].External = temp.External
	}
	if temp.BestHosts != 0 {
		alias[0].BestHosts = temp.BestHosts
	}
	if temp.Metric != "" {
		alias[0].Metric = temp.Metric
	}
	if temp.PollingInterval != 0 {
		alias[0].PollingInterval = temp.PollingInterval
	}
	if temp.Hostgroup != "" {
		alias[0].Hostgroup = temp.Hostgroup
	}
	if temp.Tenant != "" {
		alias[0].Tenant = temp.Tenant
	}

	if temp.TTL != 0 {
		alias[0].TTL = temp.TTL
	}
	//These three fields are updated even if value is empty
	//because empty values are part of the update in their case
	alias[0].ForbiddenNodes = temp.ForbiddenNodes
	alias[0].AllowedNodes = temp.AllowedNodes
	alias[0].Cname = temp.Cname

	//Validate the object alias , with the now-updated fields
	if ok, err := govalidator.ValidateStruct(alias[0]); err != nil || ok == false {
		return MessageToUser(c, http.StatusBadRequest,
			"Validation error for alias "+alias[0].AliasName+" : "+err.Error(), "home.html")
	}

	defer c.Request().Body.Close()

	// Call the modifier
	if err := alias[0].ModifyObject(); err != nil {
		return MessageToUser(c, http.StatusBadRequest,
			"Update error for alias "+alias[0].AliasName+" : "+err.Error(), "home.html")
	}

	return MessageToUser(c, http.StatusAccepted,
		alias[0].AliasName+" updated Successfully", "home.html")

}

//CheckNameDNS checks if an alias or cname already exist in DB or DNS server
func CheckNameDNS(c echo.Context) error {
	var (
		result int

		con = db.ManagerDB()
	)

	aliasToResolve := c.QueryParam("hostname")
	//Search cnames with the same name
	con.Model(&orm.Cname{}).Where("cname=?", aliasToResolve).Count(&result)
	if result == 0 {
		//Search aliases
		con.Model(&orm.Alias{}).Where("alias_name=?", aliasToResolve+".cern.ch").Count(&result)
	}
	if result == 0 {
		r, _ := net.LookupHost(aliasToResolve)
		result = len(r)
	}
	return c.JSON(http.StatusOK, result)
}
