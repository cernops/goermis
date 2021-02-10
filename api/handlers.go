package api

/*This file contains the route handlers */
import (
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func init() {
	govalidator.SetFieldsRequiredByDefault(true)
	customValidators() //enable the custom validators, see helpers.go
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

//CreateAlias creates a new alias entry
func CreateAlias(c echo.Context) error {
	var temp Resource
	username := GetUsername()
	if err := c.Bind(&temp); err != nil {
		log.Warn("[" + username + "] " + "Failed to bind params " + err.Error())
	}
	defer c.Request().Body.Close()

	//Validate structure
	if ok, err := govalidator.ValidateStruct(temp); err != nil || ok == false {
		return MessageToUser(c, http.StatusBadRequest,
			"Validation error for "+temp.AliasName+" : "+err.Error(), "home.html")
	}

	//Check for duplicates
	retrieved, _ := GetObjects(temp.AliasName)
	if len(retrieved) != 0 {
		return MessageToUser(c, http.StatusConflict,
			"Alias "+retrieved[0].AliasName+" already exists ", "home.html")

	}

	alias := sanitazeInCreation(temp)

	log.Info("[" + username + "] " + "Ready to create a new alias " + temp.AliasName)

	//Create object in DB
	if err := alias.createObjectInDB(); err != nil {
		return MessageToUser(c, http.StatusBadRequest,
			"Creation error for "+temp.AliasName+" : "+err.Error(), "home.html")
	}

	//Create in DNS
	if err := alias.createInDNS(); err != nil {
		//If it fails to create alias in DNS, we delete from DB what we created in the previous step.
		if err := alias.deleteObjectInDB(); err != nil {
			//Failed to rollback the newly created alias
			return MessageToUser(c, http.StatusBadRequest,
				"Failed to delete stray alias "+alias.AliasName+"from DB after failing to create in DNS, with error"+": "+err.Error(), "home.html")
		}
		//Failed to create in DNS, but managed to delete the newly created alias in DB
		return MessageToUser(c, http.StatusBadRequest,
			"Failed to create "+alias.AliasName+"in DNS with error"+": "+err.Error(), "home.html")
	}
	//Success message
	return MessageToUser(c, http.StatusCreated,
		temp.AliasName+" created successfully ", "home.html")

}

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
		//if alias actually exists, delete from DB
		if err := alias[0].deleteObjectInDB(); err != nil {
			return MessageToUser(c, http.StatusBadRequest, err.Error(), "home.html")

		}
		//Now delete from DNS.
		if err := alias[0].deleteFromDNS(); err != nil {

			//If deletion from DNS fails, we recreate the object in DB.
			if err := alias[0].createObjectInDB(); err != nil {
				//Failed to rollback deletion
				return MessageToUser(c, http.StatusBadRequest,
					"Failed to recreate "+alias[0].AliasName+"in DB, after failing to delete it from DNS with error"+": "+err.Error(), "home.html")
			}
			//Rollback message when deletion from DNS fails
			return MessageToUser(c, http.StatusBadRequest,
				"Failed to delete "+alias[0].AliasName+"in DNS with error"+": "+err.Error(), "home.html")
		}
		return MessageToUser(c, http.StatusOK,
			aliasToDelete+" deleted successfully ", "home.html")

	}
	return MessageToUser(c, http.StatusNotFound, aliasToDelete+" not found", "home.html")

}

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

	//Validate the object alias , with the now-updated fields
	if ok, err := govalidator.ValidateStruct(temp); err != nil || ok == false {
		return MessageToUser(c, http.StatusBadRequest,
			"Validation error for alias "+temp.AliasName+" : "+err.Error(), "home.html")
	}
	//Here we distignuish between kermis PATCH and UI form binding
	if c.Request().Method == "PATCH" {
		param = c.Param("id")
	} else {
		param = temp.AliasName
	}

	//We use the alias name for retrieving its profile from DB
	retrieved, err := GetObjects(param)
	if err != nil {
		log.Error("[" + username + "] " + "Failed to retrieve alias " + temp.AliasName + " : " + err.Error())
		return err
	}
	alias := sanitazeInUpdate(retrieved[0], temp)
	defer c.Request().Body.Close()

	// Update alias
	if err := alias.updateAlias(); err != nil {
		return MessageToUser(c, http.StatusBadRequest,
			"Update error for alias "+alias.AliasName+" : "+err.Error(), "home.html")
	}

	// Update his cnames
	if err := alias.updateCnames(); err != nil {
		return MessageToUser(c, http.StatusBadRequest,
			"Update error for alias "+alias.AliasName+" : "+err.Error(), "home.html")
	}

	// Update his nodes
	if err := alias.updateNodes(); err != nil {
		return MessageToUser(c, http.StatusBadRequest,
			"Update error for alias "+alias.AliasName+" : "+err.Error(), "home.html")
	}

	// Update his alarms
	if err := alias.updateAlarms(); err != nil {
		return MessageToUser(c, http.StatusBadRequest,
			"Update error for alias "+alias.AliasName+" : "+err.Error(), "home.html")
	}

	//Update in DNS
	//3.Update DNS
	if err = alias.updateDNS(retrieved[0]); err != nil {
		//If something goes wrong while updating, then we use the object
		//we had in DB before the update to restore that state, before the error

		//Delete the DB updates we just made and existing DNS entries
		if err = alias.deleteObjectInDB(); err != nil {
			return MessageToUser(c, http.StatusAccepted,
				"Could not delete updates while rolling back a failed DNS update for alias "+alias.AliasName, "home.html")
		}
		//Recreate the alias as it was before the update
		if err = retrieved[0].createObjectInDB(); err != nil {
			return MessageToUser(c, http.StatusAccepted,
				"Could not restore previous state while rolling back a failed DNS update for alias "+alias.AliasName, "home.html")
		}
		//Successful rollback message
		return MessageToUser(c, http.StatusAccepted,
			"Failed to update DNS for alias "+alias.AliasName+". Rolling back to previous state", "home.html")
	}

	//Success message
	return MessageToUser(c, http.StatusAccepted,
		alias.AliasName+" updated Successfully", "home.html")

}

/*CheckNameDNS checks if an alias or cname already exist in DB or DNS server
This function serves the immediate check, which is performed while writing
the alias name or cnames*/
func CheckNameDNS(c echo.Context) error {
	var (
		result int64
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
