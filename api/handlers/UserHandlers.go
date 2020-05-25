package handlers

import (
	"net"
	"net/http"
	"strconv"

	"github.com/asaskevich/govalidator"
	schema "github.com/gorilla/Schema"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"gitlab.cern.ch/lb-experts/goermis/api/models"
	"gitlab.cern.ch/lb-experts/goermis/bootstrap"
	"gitlab.cern.ch/lb-experts/goermis/db"
)

//CRUD HANDLERS

var (
	con      = db.ManagerDB()
	obj      models.Objects
	res      models.Resource
	tablerow string
	err      error
	decoder  = schema.NewDecoder()
	log      = bootstrap.Log
)

func init() {
	govalidator.SetFieldsRequiredByDefault(true)
	decoder.IgnoreUnknownKeys(true)
	models.CustomValidators()
}

//GetAliases handles requests of all aliases
func GetAliases(c echo.Context) error {
	log.WithFields(logrus.Fields{
		"package":  "handlers",
		"function": "GetAliases",
	}).Info("Ready do get all aliases")

	obj.Objects, err = models.GetObjects("", "")

	if err != nil {
		log.WithFields(logrus.Fields{
			"package":  "models",
			"function": "GetObjects",
			"error":    err,
		}).Error("Error while getting list of aliases")

		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	defer c.Request().Body.Close()
	log.WithFields(logrus.Fields{
		"package":  "models",
		"function": "GetObjects",
	}).Info("List of aliases retrieved successfully")
	return c.JSON(http.StatusOK, obj)
}

//GetAlias queries for a specific alias
func GetAlias(c echo.Context) error {
	param := c.Param("alias")

	log.WithFields(logrus.Fields{
		"package":  "handlers",
		"function": "GetAlias",
		"alias":    param,
	}).Info("Ready to retrieve alias")

	if !govalidator.IsDNSName(param) {

		log.WithFields(logrus.Fields{
			"package":  "handlers",
			"function": "GetAlias",
			"error":    err,
			"data":     param,
		}).Error("Wrong type of query parameter, expected Alphanumeric")

		return echo.NewHTTPError(http.StatusUnprocessableEntity)
	}

	//Swap between name and ID query(enables user to ask by name or id)
	if _, err := strconv.Atoi(c.Param("alias")); err == nil {
		tablerow = "id"
	} else {
		tablerow = "alias_name"
	}

	obj.Objects, err = models.GetObjects(string(param), tablerow)
	if err != nil {
		log.WithFields(logrus.Fields{
			"package":  "models",
			"function": "GetObjects",
			"error":    err,
			"data":     param,
		}).Error("Unable to get the alias")

		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	defer c.Request().Body.Close()

	log.WithFields(logrus.Fields{
		"package":  "handlers",
		"function": "GetAlias",
		"alias":    param,
	}).Info("Alias retrieved successfully")
	return c.JSON(http.StatusOK, obj)

}

//NewAlias creates a new alias entry in the DB
func NewAlias(c echo.Context) error {
	var r models.Resource
	//Get the params from the form
	params, err := c.FormParams()
	//Decode them into the resource model for validation
	log.WithFields(logrus.Fields{
		"package":  "handlers",
		"function": "NewAlias",
		"alias":    params,
	}).Info("Preparing to create a new alias")

	err = decoder.Decode(&r, params)
	if err != nil {
		log.WithFields(logrus.Fields{
			"package":  "handlers",
			"function": "NewAlias",
			"error":    err,
			"data":     params,
		}).Warn("Error while decoding parameters")

	}
	//Default values and domain
	r.AddDefaultValues()
	r.Hydrate()

	defer c.Request().Body.Close()
	//Validate structure
	ok, errs := govalidator.ValidateStruct(r)
	if errs != nil || ok == false {
		log.WithFields(logrus.Fields{
			"package":  "handlers",
			"function": "NewAlias",
			"error":    err,
			"data":     r,
		}).Warn("Error while validating parameters")

		return c.Render(http.StatusUnprocessableEntity, "home.html", map[string]interface{}{
			"Auth":    true,
			"Message": "There was an error while validating the alias " + params.Get("alias_name") + "Error: " + errs.Error(),
		})
	}
	//Create object
	err = r.CreateObject()

	if err != nil {
		log.WithFields(logrus.Fields{
			"package":  "models",
			"function": "NewAlias",
			"error":    err,
			"data":     r,
		}).Error("Error while creating the new alias")

		return c.Render(http.StatusBadRequest, "home.html", map[string]interface{}{
			"Auth":    true,
			"Message": "There was an error while creating the alias " + params.Get("alias_name") + "Error: " + err.Error(),
		})

	}
	log.WithFields(logrus.Fields{
		"package":  "handlers",
		"function": "NewAlias",
		"alias":    r.AliasName,
	}).Info("Alias created successfully")

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

		log.WithFields(logrus.Fields{
			"package":  "handlers",
			"function": "DeleteAlias",
			"alias":    aliasName,
		}).Warn("Wrong type of query parameter, expected Alias name")

		return echo.NewHTTPError(http.StatusUnprocessableEntity)
	}

	alias, err := models.GetObjects(aliasName, "alias_name")
	if err != nil {
		log.WithFields(logrus.Fields{
			"package":  "models",
			"function": "GetObjects",
			"error":    err.Error(),
			"alias":    aliasName,
		}).Error("There was an error while retrieving existing data of alias")
	}

	defer c.Request().Body.Close()

	err = alias[0].DeleteObject()
	if err != nil {
		log.WithFields(logrus.Fields{
			"package":  "models",
			"function": "DeleteObject",
			"error":    err.Error(),
			"alias":    alias[0].AliasName,
		}).Error("There was an error while deleting alias")

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
	//Pass the values from the form to our structure
	decoder.Decode(&newObj, params)
	if err != nil {
		log.WithFields(logrus.Fields{
			"package":  "handlers",
			"function": "ModifyObject",
			"error":    err.Error(),
			"data":     params,
		}).Warn("There was an error while decoding FormParams")
		return err
	}
	//Validate
	ok, errs := govalidator.ValidateStruct(newObj)
	if errs != nil || ok == false {
		return c.Render(http.StatusUnprocessableEntity, "home.html", map[string]interface{}{
			"Auth":    true,
			"Message": "There was an error while validating the alias " + params.Get("alias_name") + "Error: " + errs.Error(),
		})
	}

	existingObj, err := models.GetObjects(newObj.AliasName, "alias_name")
	if err != nil {
		log.WithFields(logrus.Fields{
			"package":  "models",
			"function": "GetObjects",
			"error":    err.Error(),
			"alias":    newObj.AliasName,
		}).Error("There was an error while retrieving existing data of alias")
		return err
	}
	defer c.Request().Body.Close()
	// Call the modifier
	err = existingObj[0].ModifyObject(newObj)
	if err != nil {

		log.WithFields(logrus.Fields{
			"package":  "models",
			"function": "ModifyObject",
			"error":    err.Error(),
			"alias":    newObj.AliasName,
		}).Error("Error while updating alias")

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

//CheckNameDNS checks if an alias or cname already exist in DB or DNS server
func CheckNameDNS(c echo.Context) error {
	aliasToResolve := c.QueryParam("hostname")
	var result int
	log.WithFields(logrus.Fields{
		"package":  "handlers",
		"function": "CheckNameDNS",
		"alias":    aliasToResolve,
	}).Info("Checking if the object exists")

	con.Model(&models.Cname{}).Where("c_name=?", aliasToResolve).Count(&result)
	if result == 0 {
		con.Model(&models.Alias{}).Where("alias_name=?", aliasToResolve+".cern.ch").Count(&result)
	}
	if result == 0 {
		if r, err := net.LookupHost(aliasToResolve); err != nil {
			log.WithFields(logrus.Fields{
				"package":  "handlers",
				"function": "CheckNameDNS",
				"reply":    err.Error(),
				"alias":    aliasToResolve,
			}).Info("Reply of LookupHost")
			result = 0
		} else {
			log.WithFields(logrus.Fields{
				"package":  "handlers",
				"function": "CheckNameDNS",
				"alias":    aliasToResolve,
			}).Warn("Alias with the same name exists in the DNS")
			result = len(r)
		}
	}
	return c.JSON(http.StatusOK, result)
}
