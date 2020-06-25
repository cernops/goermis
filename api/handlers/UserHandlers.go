package handlers

import (
	"net"
	"net/http"
	"strconv"

	"github.com/asaskevich/govalidator"
	schema "github.com/gorilla/Schema"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"gitlab.cern.ch/lb-experts/goermis/api/models"
	"gitlab.cern.ch/lb-experts/goermis/db"
)

//CRUD HANDLERS

var (
	con      = db.ManagerDB()
	all      models.Objects
	res      models.Resource
	tablerow string
	decoder  = schema.NewDecoder()
	e        error
)

func init() {
	govalidator.SetFieldsRequiredByDefault(true)
	decoder.IgnoreUnknownKeys(true)
	models.CustomValidators()
}

//Objects holds multiple result structs

//GetAliases handles requests of all aliases
func GetAliases(c echo.Context) error {

	log.Info("Ready do get all aliases")
	if all.Objects, e = models.GetObjects("", ""); e != nil {
		log.Error(e.Error())
		return echo.NewHTTPError(http.StatusBadRequest, e.Error())
	}
	defer c.Request().Body.Close()
	log.Info("List of aliases retrieved successfully")
	return c.JSON(http.StatusOK, all)
}

//GetAlias queries for a specific alias
func GetAlias(c echo.Context) error {
	param := c.Param("alias")
	if !govalidator.IsDNSName(param) {
		log.Error("Wrong type of query parameter.Expected alphanum, received " + param)
		return echo.NewHTTPError(http.StatusUnprocessableEntity)
	}

	//Swap between name and ID query(enables user to ask by name or id)
	if _, err := strconv.Atoi(c.Param("alias")); err == nil {
		tablerow = "id"
	} else {
		tablerow = "alias_name"
	}

	if all.Objects, e = models.GetObjects(string(param), tablerow); e != nil {
		log.Error("Unable to get the alias with error message: " + e.Error())

		return echo.NewHTTPError(http.StatusBadRequest, e.Error())
	}
	defer c.Request().Body.Close()
	log.Info("Alias retrieved successfully")
	return c.JSON(http.StatusOK, all)

}

//NewAlias creates a new alias entry in the DB
func NewAlias(c echo.Context) error {
	var r models.Resource
	//Get the params from the form
	params, err := c.FormParams()
	r.User = c.Request().Header.Get("X-Forwarded-User")
	if err != nil {
		log.Warn("Failed to get params from form with error : " + err.Error())
	}
	//Decode them into the resource model for validation
	log.Info("Preparing to create a new alias")
	if err := decoder.Decode(&r, params); err != nil {
		log.Warn("Failed to decode form parameters to structure with error: " +
			err.Error())
	}

	//Default values and domain
	r.DefaultAndHydrate()
	defer c.Request().Body.Close()
	//Validate structure
	if ok, err := govalidator.ValidateStruct(r); err != nil || ok == false {
		log.Warn("Failed to validate parameters with error: " + err.Error())

		return c.Render(http.StatusUnprocessableEntity, "home.html", map[string]interface{}{
			"Auth": true,
			"Message": "There was an error while validating the alias " +
				params.Get("alias_name") +
				"Error: " + err.Error(),
		})
	}
	//Create object in DB
	if err := r.CreateObject(); err != nil {
		log.Error(err.Error())
		return c.Render(http.StatusBadRequest, "home.html", map[string]interface{}{
			"Auth": true,
			"Message": "There was an error while creating the alias" +
				params.Get("alias_name") +
				"Error: " + err.Error(),
		})

	}

	log.Info("Alias created successfully")
	return c.Render(http.StatusCreated, "home.html", map[string]interface{}{
		"Auth":    true,
		"Message": params.Get("alias_name") + "  created Successfully",
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
	var newObj models.Resource
	//Retrieve the parameters from the form
	params, err := c.FormParams()
	if err != nil {
		log.Warn("Failed to get form parameters with error: " +
			err.Error())
		return err
	}
	//Pass the values from the form to our structure
	if err := decoder.Decode(&newObj, params); err != nil {
		log.Warn("Failed to decode form parameters into the structure with error: " +
			err.Error())
		return err
	}
	//Validate
	if ok, err := govalidator.ValidateStruct(newObj); err != nil || ok == false {
		log.Error("Validation failed with error: " + err.Error())
		return c.Render(http.StatusUnprocessableEntity, "home.html", map[string]interface{}{
			"Auth": true,
			"Message": "There was an error while validating the alias " +
				params.Get("alias_name") + "Error: " +
				err.Error(),
		})
	}

	existingObj, err := models.GetObjects(newObj.AliasName, "alias_name")
	if err != nil {
		log.Error("Failed to retrieve existing data for alias " +
			newObj.AliasName +
			"with error: " + err.Error())
		return err
	}
	defer c.Request().Body.Close()

	// Call the modifier
	if err := existingObj[0].ModifyObject(newObj); err != nil {

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
	aliasToResolve := c.QueryParam("hostname")
	var result int
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
