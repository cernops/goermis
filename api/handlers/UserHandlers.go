package handlers

import (
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
	schema "github.com/gorilla/Schema"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"gitlab.cern.ch/lb-experts/goermis/api/models"
	"gitlab.cern.ch/lb-experts/goermis/db"
)

var (
	decoder = schema.NewDecoder()
)

func init() {
	govalidator.SetFieldsRequiredByDefault(true)
	decoder.IgnoreUnknownKeys(true)
	models.CustomValidators()
}

//GetAliases returns a list of ALL aliases
func GetAliases(c echo.Context) error {
	var (
		list models.Objects
		e    error
	)

	log.Info("Ready do get all aliases")
	//If empty values provided,the MySQL query returns all aliases
	if list.Objects, e = models.GetObjects("", ""); e != nil {
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
		list     models.Objects
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

	if list.Objects, e = models.GetObjects(string(param), tablerow); e != nil {
		log.Error("Unable to get the alias with error message: " + e.Error())

		return echo.NewHTTPError(http.StatusBadRequest, e.Error())
	}
	defer c.Request().Body.Close()
	log.Info("Alias retrieved successfully")
	return c.JSON(http.StatusOK, list)

}

//NewAlias creates a new alias entry in the DB
func NewAlias(c echo.Context) error {
	r := populateStruct(c)
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

	alias, err := models.GetObjects(aliasName, "alias_name")
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

//ModifyAlias modifes cnames, nodes, hostgroup and best_hosts parameters
func ModifyAlias(c echo.Context) error {
	newObject := populateStruct(c)
	//Validate
	if ok, err := govalidator.ValidateStruct(newObject); err != nil || ok == false {
		log.Error("Validation failed with error: " + err.Error())
		return c.Render(http.StatusUnprocessableEntity, "home.html", map[string]interface{}{
			"Auth": true,
			"Message": "There was an error while validating the alias " +
				newObject.AliasName + "Error: " +
				err.Error(),
		})
	}
	//Retrieve from DB the info for the alias that we will modify
	existingObj, err := models.GetObjects(newObject.AliasName, "alias_name")
	if err != nil {
		log.Error("Failed to retrieve existing data for alias " +
			newObject.AliasName +
			"with error: " + err.Error())
		return err
	}
	// Modify existingObj with newObject as parameter
	if err := existingObj[0].ModifyObject(newObject); err != nil {

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
	con.Model(&models.Cname{}).Where("c_name=?", aliasToResolve).Count(&result)
	if result == 0 {
		con.Model(&models.Alias{}).Where("alias_name=?", aliasToResolve+".cern.ch").Count(&result)
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

func populateStruct(c echo.Context) (r models.Resource) {
	//UI parameters are passed inside a form
	params, err := c.FormParams()
	if err != nil {
		log.Warn("Failed to get params from form with error : " + err.Error())
	}
	//If form has values, decode them in structure
	if len(params) != 0 {
		//Decode parameters in structure
		if err := decoder.Decode(&r, params); err != nil {
			log.Warn("Failed to decode form parameters to structure with error: " +
				err.Error())
		}
		//Else handle json params
	} else {
		//Kermis is using this one
		log.Info("Kermis parameters")
		if err := c.Bind(&r); err != nil {
			log.Warn("Failed to bind params " + err.Error())
		}

	}
	//User is provided in the HEADER
	r.User = c.Request().Header.Get("X-Forwarded-User")
	defer c.Request().Body.Close()
	return r
}
