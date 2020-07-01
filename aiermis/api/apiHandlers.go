package api

import (
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/davecgh/go-spew/spew"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"gitlab.cern.ch/lb-experts/goermis/aiermis/orm"
	"gitlab.cern.ch/lb-experts/goermis/db"
)

func init() {
	govalidator.SetFieldsRequiredByDefault(true)
	customValidators()
}

//COMMON HANDLERS

//GetAliases returns a list of ALL aliases
func GetAliases(c echo.Context) error {
	var (
		list Objects
		e    error
	)

	log.Info("Ready do get all aliases")
	//If empty values provided,the MySQL query returns all aliases
	if list.Objects, e = GetObjects("", ""); e != nil {
		log.Error(e.Error())
		return echo.NewHTTPError(http.StatusBadRequest, e.Error())
	}
	defer c.Request().Body.Close()
	log.Info("List of aliases retrieved successfully")
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
	param := c.Param("alias")
	//Validate that the parameter is DNS-compatible
	if !govalidator.IsDNSName(param) {
		log.Error("Wrong type of query parameter.Expected alphanum, received " + param)
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
		log.Error("Unable to get the alias with error message: " + e.Error())

		return echo.NewHTTPError(http.StatusBadRequest, e.Error())
	}
	defer c.Request().Body.Close()
	log.Info("Alias retrieved successfully")
	return c.JSON(http.StatusOK, list)

}

//CreateAlias creates a new alias entry in the DB
func CreateAlias(c echo.Context) error {

	//Struct r serves for getting request values and validate them
	var r Resource
	if err := c.Bind(&r); err != nil {
		log.Warn("Failed to bind params " + err.Error())
	}

	//User is provided in the HEADER
	r.User = c.Request().Header.Get("X-Forwarded-User")
	defer c.Request().Body.Close()

	//Validate structure
	if ok, err := govalidator.ValidateStruct(r); err != nil || ok == false {
		log.Warn("Failed to validate parameters with error: " + err.Error())
		return c.Render(http.StatusUnprocessableEntity, "home.html", map[string]interface{}{
			"Auth": true,
			"Message": "There was an error while validating the alias " +
				r.AliasName +
				"Error: " + err.Error(),
		})
	}

	//Default values and hydrate(domain,visibility)
	r.DefaultAndHydrate()

	log.Info("Ready to create a new alias")
	//Create object
	if err := r.CreateObject(); err != nil {
		log.Error(err.Error())
		return c.Render(http.StatusBadRequest, "home.html", map[string]interface{}{
			"Auth": true,
			"Message": "There was an error while creating the alias" +
				r.AliasName +
				"Error: " + err.Error(),
		})

	}

	log.Info("Alias created successfully")
	return c.Render(http.StatusCreated, "home.html", map[string]interface{}{
		"Auth":    true,
		"Message": r.AliasName + "  created Successfully",
	})

}

//DeleteAlias deletes the requested alias from the DB
func DeleteAlias(c echo.Context) error {
	aliasName := c.FormValue("alias_name")
	//Validate alias name only
	if !govalidator.IsDNSName(aliasName) {

		log.Warn("Wrong type of query parameter, expected Alias name")

		return echo.NewHTTPError(http.StatusUnprocessableEntity)
	}

	alias, err := GetObjects(aliasName, "alias_name")
	if err != nil {
		log.Error("Failed to retrieve the existing data of alias " +
			aliasName +
			"with error: " + err.Error())
	}

	defer c.Request().Body.Close()

	if err := alias[0].DeleteObject(); err != nil {
		log.Error("Failed to delete alias" + alias[0].AliasName +
			"with error: " + err.Error())

		return c.Render(http.StatusBadRequest, "home.html", map[string]interface{}{
			"Auth":    true,
			"Message": err,
		})

	}
	log.Info("Alias deleted successfully")
	return c.Render(http.StatusOK, "home.html", map[string]interface{}{
		"Auth":    true,
		"Message": alias[0].AliasName + "  deleted Successfully",
	})

}

//LBWEB-SPECIFIC HANDLERS

//ModifyAlias modifes cnames, nodes, hostgroup and best_hosts parameters
func ModifyAlias(c echo.Context) error {
	var r Resource
	if err := c.Bind(&r); err != nil {
		log.Warn("Failed to bind params " + err.Error())
	}
	spew.Dump(r)

	//After that, we use the alias name for retrieving its profile from DB
	existingObj, err := GetObjects(r.AliasName, "alias_name")
	if err != nil {
		log.Error("Failed to retrieve existing data for alias " +
			r.AliasName +
			"with error: " + err.Error())
		return err
	}

	//Preserve the DB values for these fields, because they are needed for comparison
	stashC := existingObj[0].Cname          //Cnames
	stashA := existingObj[0].AllowedNodes   //AllowedNodes
	stashF := existingObj[0].ForbiddenNodes //ForbiddenNodes

	//Update fields if changed, PATCH
	if r.AllowedNodes != "" {
		existingObj[0].AllowedNodes = r.AllowedNodes
	}
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
	if r.ForbiddenNodes != "" {
		existingObj[0].ForbiddenNodes = r.ForbiddenNodes
	}
	if r.Cname != "" {
		existingObj[0].Cname = r.Cname
	}
	if r.TTL != 0 {
		existingObj[0].TTL = r.TTL
	}

	/*
		//Now that we have old and new information, we update existingObj with changed fields
		if err := c.Bind(&existingObj[0]); err != nil {
			log.Warn("Failed to decode form parameters into the structure with error: " +
				err.Error())
			return err
		}*/

	//Validate
	if ok, err := govalidator.ValidateStruct(existingObj[0]); err != nil || ok == false {
		log.Error("Validation failed with error: " + err.Error())
		return c.Render(http.StatusUnprocessableEntity, "home.html", map[string]interface{}{
			"Auth": true,
			"Message": "There was an error while validating the alias " +
				existingObj[0].AliasName + "Error: " +
				err.Error(),
		})
	}

	defer c.Request().Body.Close()

	// Call the modifier
	if err := existingObj[0].ModifyObject(stashC, stashA, stashF); err != nil {

		log.Error("Failed to update alias " + existingObj[0].AliasName +
			"with error: " + err.Error())

		return c.Render(http.StatusOK, "home.html", map[string]interface{}{
			"Auth":    true,
			"Message": "There was an error while updating the alias in DB: " + err.Error(),
		})
	}

	log.Info("Alias updated successfully")
	return c.Render(http.StatusOK, "home.html", map[string]interface{}{
		"Auth":    true,
		"Message": c.FormValue("alias_name") + "  updated Successfully",
	})
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
			log.Warn("Alias with the same name exists in the DNS")
			result = len(r)
			return c.Render(http.StatusConflict, "home.html", map[string]interface{}{
				"Auth":    true,
				"Message": "Alias with the same name exists in the DNS",
			})

		}
	}
	return c.JSON(http.StatusOK, result)
}

//KERMIS-SPECIFIC HANDLERS

//PatchAlias patches an existing alias. It serves kermis update operation
func PatchAlias(c echo.Context) error {
	var r Resource
	if err := c.Bind(&r); err != nil {
		log.Info("Failed to bind request to structure")
		return c.JSON(http.StatusUnprocessableEntity, "Failed to bind Request")
	}
	a := c.ParamValues()
	b := c.Param("alias_name")
	c1 := c.Get("alias_name")
	d := c.Get("alias_name")
	e1 := c.QueryParam("alias_name")

	log.Info(a)
	log.Info(b)
	log.Info(c1)
	log.Info(d)
	log.Info(e1)
	return c.JSON(http.StatusOK, "Update succedeed")

}
