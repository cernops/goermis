package api

/*This file contains the route handlers */
import (
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"
	"gitlab.cern.ch/lb-experts/goermis/bootstrap"
)

func init() {
	govalidator.SetFieldsRequiredByDefault(true)
	customValidators() //enable the custom validators, see helpers.go

}

var (
	log = bootstrap.GetLog()
)

func get(c echo.Context) ([]Alias, error) {

	var (
		queryResults = []Alias{}
		e            error
	)
	username := GetUsername()
	param := c.QueryParam("alias_name")
	if param == "" {
		log.Info("[" + username + "] " + " is querying for all aliases")
		//If empty values provided,the MySQL query returns all aliases
		if queryResults, e = GetObjects("all"); e != nil {
			log.Error("[" + username + "] " + e.Error())
			return queryResults, e
		}
	} else {
		log.Info("[" + username + "] " + " is querying for alias with name/ID = " + param)
		//Validate that the parameter is DNS-compatible
		if !govalidator.IsDNSName(param) {
			e := errors.New("[" + username + "] " + "Wrong type of query parameter.Expected alphanum, received " + param)
			log.Error(e)
			return queryResults, e
		}

		if _, err := strconv.Atoi(param); err != nil {
			if !strings.HasSuffix(param, ".cern.ch") {
				param = param + ".cern.ch"
			}
		}

		if queryResults, e = GetObjects(string(param)); e != nil {
			log.Error("[" + username + "]" + "Unable to get alias" + param + " : " + e.Error())
			return queryResults, e
		}

	}

	defer c.Request().Body.Close()
	return queryResults, nil

}

//GetAlias returns a list of ALL aliases
func GetAlias(c echo.Context) error {
	queryResults, e := get(c)
	if e != nil {
		return echo.NewHTTPError(http.StatusBadRequest, e.Error())
	}

	return c.JSON(http.StatusOK, parse(queryResults))
}

//GetAliasRaw returns row data
func GetAliasRaw(c echo.Context) error {
	queryResults, e := get(c)
	if e != nil {
		return echo.NewHTTPError(http.StatusBadRequest, e.Error())
	}

	return c.JSON(http.StatusOK, queryResults)

}

//CreateAlias creates a new alias entry
func CreateAlias(c echo.Context) error {

	var temp Resource
	username := GetUsername()

	if err := c.Bind(&temp); err != nil {
		log.Warn("[" + username + "] " + "Failed to bind params " + err.Error())
	}
	defer c.Request().Body.Close()
	log.Info("[" + username + "] " + "Ready to create alias " + temp.AliasName)

	//Check for duplicates
	retrieved, _ := GetObjects(temp.AliasName)
	if len(retrieved) != 0 {
		return MessageToUser(c, http.StatusConflict,
			"Alias "+retrieved[0].AliasName+" already exists ", "home.html")

	}
	log.Info("[" + username + "] " + "Duplicate check passed for alias " + temp.AliasName)
	alias := sanitazeInCreation(c, temp)

	log.Info("[" + username + "] " + "Sanitazed succesfully " + temp.AliasName)

	//Validate structure
	if ok, err := govalidator.ValidateStruct(alias); err != nil || ok == false {
		return MessageToUser(c, http.StatusBadRequest,
			"Validation error for "+temp.AliasName+" : "+err.Error(), "home.html")
	}

	log.Info("[" + username + "] " + "Validation passed for alias " + temp.AliasName)

	//Create object in DB
	if err := alias.createObjectInDB(); err != nil {
		return MessageToUser(c, http.StatusBadRequest,
			"Creation error for "+temp.AliasName+" : "+err.Error(), "home.html")
	}

	log.Info("[" + username + "] " + "Created in DB, now creating in DNS alias: " + alias.AliasName)

	//Create in DNS
	if err := alias.createInDNS(); err != nil {

		log.Error("[" + username + "] " + "Failed to create entry in DNS, initiating rollback for alias: " + alias.AliasName)

		//We dont know the newly assigned ID for our alias
		//We need the ID for clearing its associations
		alias.ID = findAliasID(alias.AliasName)

		//If it fails to create alias in DNS, we delete from DB what we created in the previous step.
		if err := alias.deleteObjectInDB(); err != nil {

			//Failed to rollback the newly created alias
			return MessageToUser(c, http.StatusBadRequest,
				"Failed to delete stray alias "+alias.AliasName+"from DB after failing to create in DNS, with error"+": "+err.Error(), "home.html")
		}
		//Failed to create in DNS, but managed to delete the newly created alias in DB
		return MessageToUser(c, http.StatusBadRequest,
			"Failed to create "+alias.AliasName+" in DNS with error"+": "+err.Error(), "home.html")
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
	log.Info("[" + username + "] " + "Ready to delete alias " + aliasToDelete)

	//Validate alias name only, since the rest of the struct will be empty when DELETE
	if !govalidator.IsDNSName(aliasToDelete) {
		log.Warn("[" + username + "] " + "Wrong type of query parameter, expected Alias name, got :" + aliasToDelete)
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	log.Info("[" + username + "] " + "Validation passed for " + aliasToDelete)

	alias, err := GetObjects(aliasToDelete)
	if err != nil {
		log.Error("[" + username + "] " + "Failed to retrieve alias " + aliasToDelete + " : " + err.Error())
	}
	defer c.Request().Body.Close()
	log.Info("[" + username + "] " + "Retrieved existing alias from DB " + aliasToDelete)
	if len(alias) != 0 {
		log.Info("[" + username + "] " + "Now deleting from the DB the alias " + aliasToDelete)
		//if alias actually exists, delete from DB
		if err := alias[0].deleteObjectInDB(); err != nil {
			return MessageToUser(c, http.StatusBadRequest, err.Error(), "home.html")

		}
		log.Info("[" + username + "] " + "Deleted from DB, now deleting from DNS " + aliasToDelete)
		//Now delete from DNS.
		if err := alias[0].deleteFromDNS(); err != nil {
			log.Error("[" + username + "] " + "Something went wrong while deleting " + aliasToDelete + " from DNS, trying to rollback")
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

	log.Info("[" + username + "] " + "Ready to modify alias" + temp.AliasName)

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
	log.Info("[" + username + "] " + "Retrieved existing data for " + temp.AliasName)

	alias := sanitazeInUpdate(c, retrieved[0], temp)

	defer c.Request().Body.Close()

	log.Info("[" + username + "] " + "Sanitized successfully" + temp.AliasName)

	//Validate the object alias , with the now-updated fields
	if ok, err := govalidator.ValidateStruct(alias); err != nil || ok == false {
		return MessageToUser(c, http.StatusBadRequest,
			"Validation error for alias "+temp.AliasName+" : "+err.Error(), "home.html")
	}
	log.Info("[" + username + "] " + "Validation check passed" + temp.AliasName)

	// Update alias
	if err := alias.updateAlias(); err != nil {
		return MessageToUser(c, http.StatusBadRequest,
			"Update error for alias "+alias.AliasName+" : "+err.Error(), "home.html")
	}
	log.Info("[" + username + "] " + "Updated alias" + alias.AliasName + ", now checking his associations")

	// Update his cnames
	if err := alias.updateCnames(); err != nil {
		return MessageToUser(c, http.StatusBadRequest,
			"Update error for alias "+alias.AliasName+" : "+err.Error(), "home.html")
	}
	log.Info("[" + username + "] " + "Finished the cnames update for " + temp.AliasName)

	// Update his nodes
	if err := alias.updateNodes(); err != nil {
		return MessageToUser(c, http.StatusBadRequest,
			"Update error for alias "+alias.AliasName+" : "+err.Error(), "home.html")
	}
	log.Info("[" + username + "] " + "Finished the nodes update for " + temp.AliasName)

	// Update his alarms
	if err := alias.updateAlarms(); err != nil {
		return MessageToUser(c, http.StatusBadRequest,
			"Update error for alias "+alias.AliasName+" : "+err.Error(), "home.html")
	}
	log.Info("[" + username + "] " + "The DB was updated successfully, now we can update DNS")

	//Update in DNS
	if err = alias.updateDNS(retrieved[0]); err != nil {
		//If something goes wrong while updating, then we use the object
		//we had in DB before the update to restore that state, before the error
		log.Error("[" + username + "] " + "Could not update " + alias.AliasName + " in DNS, starting rollback")
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
