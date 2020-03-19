package handlers

import (
	"net/http"
	"os"
	"strconv"

	schema "github.com/gorilla/Schema"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"gitlab.cern.ch/lb-experts/goermis/api/models"
	"gitlab.cern.ch/lb-experts/goermis/db"
)

//CRUD HANDLERS

var (
	con      = db.ManagerDB()
	alias    models.Alias
	obj      models.Objects
	res      models.Resource
	tablerow string
	err      error
	decoder  = schema.NewDecoder()
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
}

//GetAliases handles requests of all aliases
func GetAliases(c echo.Context) error {

	obj.Objects, err = res.GetObjects("", "")

	if err != nil {

		log.Errorf("Error while getting list of aliases with error : " + err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	log.Info("Success while retrieving aliases")
	return c.JSON(http.StatusOK, obj)
}

//GetAlias queries for a specific alias
func GetAlias(c echo.Context) error {
	param := c.Param("alias")

	/*if !c.Validate(IsAlphanumeric(param)) {
		log.Error("Wrong type of query parameter, expected Alphanumeric")
		return echo.NewHTTPError(http.StatusUnprocessableEntity)
	}*/

	//Swap between name and ID query
	if _, err := strconv.Atoi(c.Param("alias")); err == nil {
		tablerow = "id"
	} else {
		tablerow = "alias_name"
	}

	defer c.Request().Body.Close()

	obj.Objects, err = res.GetObjects(string(param), tablerow)
	if err != nil {
		log.Error("Unable to get the alias " + param + "with error : " + err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	log.Info("Alias retrieved successfully")
	return c.JSON(http.StatusOK, obj)

}

//NewAlias creates a new alias entry in the DB
func NewAlias(c echo.Context) error {

	//Get the params from the form
	params, err := c.FormParams()
	alias.Prepare()
	err = alias.CreateObject(params)
	if err != nil {
		log.Error("Error while creating alias " + params.Get("alias_name") + "with error : " + err.Error())
		return c.Render(http.StatusCreated, "home.html", map[string]interface{}{
			"Auth":    true,
			"Message": "There was an error while creating the alias " + params.Get("alias_name") + "Error: " + err.Error(),
		})

	}
	log.Info("Alias created successfully")
	return c.Render(http.StatusCreated, "home.html", map[string]interface{}{
		"Auth":    true,
		"Message": params.Get("alias_name") + "  created Successfully",
	})

}

//DeleteAlias is a prototype
func DeleteAlias(c echo.Context) error {
	var r models.Resource
	//Get the params from the form
	aliasName := c.FormValue("alias_name")
	defer c.Request().Body.Close()
	alias, err := r.GetObjects(aliasName, "alias_name")
	if err != nil {
		log.Error("Error while getting existing data for alias " + aliasName + "with error : " + err.Error())
	}
	err = alias[0].DeleteObject()
	if err != nil {
		log.Error("Failed to delete object")
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

//ModifyAlias is a prototype
func ModifyAlias(c echo.Context) error {
	var new models.Resource
	//Retrieve the parameters from the form
	params, err := c.FormParams()
	decoder.Decode(&new, params)

	if err != nil {
		log.Error("There was an error reading the parameters:" + err.Error())
		return err
	}

	r, err := res.GetObjects(new.AliasName, "alias_name")
	if err != nil {
		log.Error("There was an error retrieving existing data for alias " + new.AliasName + "ERROR:" + err.Error())
		return err
	}
	// Call the modifier
	err = r[0].ModifyObject(new)
	if err != nil {
		log.Error("There was an error updating the alias: " + new.AliasName + "Error: " + err.Error())

		return c.Render(http.StatusOK, "home.html", map[string]interface{}{
			"Auth":    true,
			"Message": "There was an error while updating the alias" + err.Error(),
		})
	}
	log.Info("Alias updated successfully")
	return c.Render(http.StatusOK, "home.html", map[string]interface{}{
		"Auth":    true,
		"Message": c.FormValue("alias_name") + "  updated Successfully",
	})
}

//CheckNameDNS for now is a prototype function that enables frontend to work
func CheckNameDNS(c echo.Context) error {
	return c.JSON(http.StatusOK, 0)
}
