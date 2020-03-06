package handlers

import (
	"net/http"
	"os"
	"strconv"

	"github.com/asaskevich/govalidator"

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
	tablerow string
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
	response, err := obj.GetObjectsList()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, response)
}

//GetAlias queries for a specific alias
func GetAlias(c echo.Context) error {

	param := c.Param("alias")

	if !govalidator.IsAlphanumeric(param) {
		log.Error("Wrong type of query parameter, expected Alphanumeric")
		return echo.NewHTTPError(http.StatusUnprocessableEntity)
	}

	//Swap between name and ID query
	if _, err := strconv.Atoi(c.Param("alias")); err == nil {
		tablerow = "id"
	} else {
		tablerow = "alias_name"
	}

	defer c.Request().Body.Close()

	response, err := obj.GetObject(string(param), tablerow)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, response)

}

//NewAlias creates a new alias entry in the DB
func NewAlias(c echo.Context) error {

	//Get the params from the form
	params, err := c.FormParams()
	alias.Prepare()
	err = alias.CreateObject(params)
	if err != nil {
		return c.Render(http.StatusCreated, "home.html", map[string]interface{}{
			"Auth":    true,
			"Message": "There was an error while creating the alias " + params.Get("alias_name"),
		})

	}

	return c.Render(http.StatusCreated, "home.html", map[string]interface{}{
		"Auth":    true,
		"Message": params.Get("alias_name") + "  created Successfully",
	})

}

//DeleteAlias is a prototype
func DeleteAlias(c echo.Context) error {
	//Get the params from the form
	aliasName := c.FormValue("alias_name")
	defer c.Request().Body.Close()
	err := alias.DeleteObject(aliasName)
	if err != nil {

		return c.Render(http.StatusBadRequest, "home.html", map[string]interface{}{
			"Auth":    true,
			"Message": err,
		})

	}

	return c.Render(http.StatusOK, "home.html", map[string]interface{}{
		"Auth":    true,
		"Message": aliasName + "  deleted Successfully",
	})

}

//ModifyAlias is a prototype
func ModifyAlias(c echo.Context) error {

	//Retrieve the parameters from the form
	params, err := c.FormParams()
	if err != nil {
		log.Error("There was an error reading the parameters:" + err.Error())
		return err
	}
	// Call the modifier
	err = alias.ModifyObject(params)
	if err != nil {
		log.Error("There was an error updating the alias:" + err.Error())

		return c.Render(http.StatusOK, "home.html", map[string]interface{}{
			"Auth":    true,
			"Message": "There was an error while updating the alias" + err.Error(),
		})
	}

	return c.Render(http.StatusOK, "home.html", map[string]interface{}{
		"Auth":    true,
		"Message": c.FormValue("alias_name") + "  updated Successfully",
	})
}

//CheckNameDNS for now is a prototype function that enables frontend to work
func CheckNameDNS(c echo.Context) error {
	return c.JSON(http.StatusOK, 0)
}