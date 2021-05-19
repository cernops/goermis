package ermis

/*This file contains the router handlers */
import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"
	"gitlab.cern.ch/lb-experts/goermis/auth"
	"gitlab.cern.ch/lb-experts/goermis/bootstrap"
	"gitlab.cern.ch/lb-experts/goermis/db"
	landbsoap "gitlab.cern.ch/lb-experts/goermis/landb"
)

func init() {
	govalidator.SetFieldsRequiredByDefault(true)
	customValidators() //enable the custom validators, see helpers.go

}

var (
	log = bootstrap.GetLog()
)

//GetAlias returns aliases objects, where cnames/alarms/nodes are condensed to a list of names
func GetAlias(c echo.Context) error {
	queryResults, e := get(c)
	if e != nil {
		return echo.NewHTTPError(http.StatusBadRequest, e.Error())
	}

	return c.JSON(http.StatusOK, parse(queryResults))
}

//GetAliasRaw returns aliases objects, where alarms/cnames/nodes objects are fully represented
func GetAliasRaw(c echo.Context) error {
	queryResults, e := get(c)
	if e != nil {
		return echo.NewHTTPError(http.StatusBadRequest, e.Error())
	}

	return c.JSON(http.StatusOK, queryResults)

}

/*Used from GetAlias & GetRaw to actually get the data,
before deciding their representation format*/
func get(c echo.Context) ([]Alias, error) {

	var (
		queryResults = []Alias{}
		e            error
	)
	username := GetUsername()
	param := c.QueryParam("alias_name")
	if param == "" {
		log.Infof("[%v] is querying for all aliases", username)
		//If empty values provided,the MySQL query returns all aliases
		if queryResults, e = GetObjects("all"); e != nil {
			log.Errorf("[%v] %v", username, e.Error())
			return queryResults, e
		}
	} else {
		log.Infof("[%v] is querying for alias with name or ID =%v ", username, param)
		//Validate that the parameter is DNS-compatible
		if !govalidator.IsDNSName(param) {
			e := fmt.Errorf("[%v] Wrong type of query parameter.Expected alphanum, received: %v\n ",
				username, param)
			log.Error(e)
			return queryResults, e
		}

		if _, err := strconv.Atoi(param); err != nil {
			if !strings.HasSuffix(param, ".cern.ch") {
				param = param + ".cern.ch"
			}
		}

		if queryResults, e = GetObjects(string(param)); e != nil {
			log.Errorf("[%v] unable to get alias %v with error %v ",
				username, param, e.Error())
			return queryResults, e
		}

	}

	defer c.Request().Body.Close()
	return queryResults, nil

}

//CreateAlias creates a new alias entry
func CreateAlias(c echo.Context) error {

	var temp Resource
	username := GetUsername()

	/******bind request data*******/
	if err := c.Bind(&temp); err != nil {
		log.Warnf("[%v] failed to bind parameters: %v",
			username, err.Error())
	}
	defer c.Request().Body.Close()
	log.Infof("[%v] ready to create alias %v",
		username, temp.AliasName)

	/******check existance in all distributed systems******/
	_, checks, err := checkexistance(temp.AliasName)
	if err != nil {
		return MessageToUser(c, http.StatusBadRequest,
			fmt.Sprint(err), "home.html")
	}
	for k, v := range checks {
		if !v {
			return MessageToUser(c, http.StatusBadRequest,
				fmt.Sprintf("cannot continue with creation, because there is an existing entry for alias %v in %v", temp.AliasName, k), "home.html")
		}
	}
	log.Infof("[%v] duplicate check passed for alias %v",
		username, temp.AliasName)

	/******sanitaze the binded data into ORM******/
	alias := sanitazeInCreation(c, temp)
	log.Infof("[" + username + "] " + "Sanitazed succesfully " + temp.AliasName)

	/******Validate structure******/
	if ok, err := govalidator.ValidateStruct(alias); err != nil || !ok {
		return MessageToUser(c, http.StatusBadRequest,
			fmt.Sprintf("validation error for %v: %v", temp.AliasName, err), "home.html")
	}
	log.Infof("[%v] validation passed for alias %v",
		username, temp.AliasName)

	/******Create object in DB******/
	if err := alias.createObjectInDB(); err != nil {
		return MessageToUser(c, http.StatusBadRequest,
			fmt.Sprintf("creation error for %v: %v", temp.AliasName, err), "home.html")
	}
	log.Infof("[%v] created %v in database, now creating in DNS  ",
		username, alias.AliasName)

	/******Create in DNS******/
	if err := alias.createInDNS(); err != nil {

		log.Errorf("[%v] failed to create entry in DNS, initiating rollback for alias %v\nError:%v",
			alias.User, alias.AliasName, err)

		//rollback only DB , after failed DNS creation
		err := alias.RollbackInCreate(true, false)
		if err != nil {
			return MessageToUser(c, http.StatusBadRequest,
				fmt.Sprint(err), "home.html")

		}

		//on successful rollback
		return MessageToUser(c, http.StatusBadRequest,
			fmt.Sprintf("failed to create alias %v in DNS, database rolled back", alias.AliasName), "home.html")

	}

	/******Create secret in tbag******/
	if err := alias.createSecret(); err != nil {
		log.Errorf("[%v] failed to create secret in tbag for alias %v, initiating rollback\nerror:%v",
			alias.User, alias.AliasName, err)

		//rollback db and dns after failed secret creation
		err := alias.RollbackInCreate(true, true)
		if err != nil {
			return MessageToUser(c, http.StatusBadRequest,
				fmt.Sprint(err), "home.html")
		}

	}

	/******Success message******/
	return MessageToUser(c, http.StatusCreated,
		temp.AliasName+" created successfully ", "home.html")

}

//DeleteAlias deletes the requested alias from the DB
func DeleteAlias(c echo.Context) error {
	var (
		aliasToDelete string
	)
	username := GetUsername()

	/******switch between kermis and lbwebUI******/
	switch c.Request().Header.Get("Content-Type") {
	case "application/json":
		aliasToDelete = c.QueryParam("alias_name")
	case "application/x-www-form-urlencoded":
		aliasToDelete = c.FormValue("alias_name")

	}
	defer c.Request().Body.Close()

	/******Validate alias name ******/
	if !govalidator.IsDNSName(aliasToDelete) {
		log.Warnf("[%v] wrong type of query parameter, expected DNS name, received:%v", username, aliasToDelete)
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	log.Infof("[%v] validation passed for %v",
		username, aliasToDelete)

	/******check existance in all systems and retrieve alias object******/
	alias, checks, err := checkexistance(aliasToDelete)
	if err != nil {
		return MessageToUser(c, http.StatusBadRequest,
			fmt.Sprint(err), "home.html")
	}
	for k, v := range checks {
		if v {
			return MessageToUser(c, http.StatusBadRequest,
				fmt.Sprintf("cannot continue with deletion, because there is a missing entry for alias %v in %v", aliasToDelete, k), "home.html")
		}
	}
	log.Infof("[%v] retrieved alias %v from database, ready to delete it",
		username, aliasToDelete)

	/******delete from db******/
	if err := alias[0].deleteObjectInDB(); err != nil {
		return MessageToUser(c, http.StatusBadRequest, err.Error(), "home.html")

	}
	log.Infof("[%v] deleted from the database, now deleting from the DNS %v",
		username, aliasToDelete)

	/******Now delete from DNS******/
	if err := alias[0].deleteFromDNS(); err != nil {
		log.Errorf("[%v] something went wrong while deleting %v from DNS, initiating the rollback",
			username, aliasToDelete)

		//rollback db deletion
		if err := alias[0].RollbackInDelete(true, false); err != nil {
			return MessageToUser(c, http.StatusBadRequest, err.Error(), "home.html")

		}
	}

	/******Delete secret from tbag******/
	if len(auth.GetSecret(alias[0].AliasName)) != 0 {

		if err := alias[0].deleteSecret(); err != nil {
			log.Errorf("[%v] failed to delete the secret from tbag for alias %v, initiating the rollback",
				username, aliasToDelete)

			//rollback db and dns deletions
			if err := alias[0].RollbackInDelete(true, true); err != nil {
				return MessageToUser(c, http.StatusBadRequest, err.Error(), "home.html")

			}
		}
	}

	/******tres bien******/
	return MessageToUser(c, http.StatusOK,
		fmt.Sprintf("%v deleted successfully", aliasToDelete), "home.html")
}

//ModifyAlias modifes cnames, nodes, hostgroup and best_hosts parameters
func ModifyAlias(c echo.Context) error {
	var (
		param string
		temp  Resource
	)
	username := GetUsername()

	/******Bind request to the temp Resource******/
	if err := c.Bind(&temp); err != nil {
		log.Warnf("[%v] failed to bind parameters with error %v",
			username, err.Error())
	}
	log.Infof("[%v] ready to modify alias %v",
		username, temp.AliasName)

	/******Here we switch between kermis PATCH and UI form binding******/
	switch c.Request().Method {
	case "PATCH":
		param = c.Param("id")
	default:
		param = temp.AliasName
	}

	/******check its existance is all systems and retrieve alias profile******/
	retrieved, checks, err := checkexistance(param)
	if err != nil {
		return MessageToUser(c, http.StatusBadRequest,
			fmt.Sprint(err), "home.html")
	}
	for k, v := range checks {
		if v && k != "tbag" { //secret is not modifiable, so no need to consider it
			return MessageToUser(c, http.StatusBadRequest,
				fmt.Sprintf("cannot continue with modification, because there is a missing entry for alias %v in %v", param, k), "home.html")
		}
	}
	log.Infof("[%v] existance check passed and retrieved existing data for %v",
		username, temp.AliasName)

	/******sanitaze incoming data into ORM before updating******/
	alias, err := sanitazeInUpdate(c, retrieved[0], temp)
	if err != nil {
		return MessageToUser(c, http.StatusBadRequest,
			fmt.Sprintf("failed to sanitize %v: %v ", temp.AliasName, err), "home.html")

	}
	defer c.Request().Body.Close()
	log.Infof("[%v] sanitized successfully %v",
		username, temp.AliasName)

	/******Validate object ******/
	if ok, err := govalidator.ValidateStruct(alias); err != nil || !ok {
		return MessageToUser(c, http.StatusBadRequest,
			fmt.Sprintf("validation error for alias %v: %v", temp.AliasName, err), "home.html")
	}
	log.Infof("[%v] validation check passed for %v",
		username, temp.AliasName)

	/****** Update alias fields(hg, external, best hosts etc.) ******/
	if err := alias.updateAlias(); err != nil {
		return MessageToUser(c, http.StatusBadRequest,
			fmt.Sprintf("update error for alias %v: %v ", alias.AliasName, err), "home.html")
	}
	log.Infof("[%v] updated alias %v, now will check his associations", username, alias.AliasName)

	/****** Update the cnames ******/
	if err := alias.updateCnames(); err != nil {
		return MessageToUser(c, http.StatusBadRequest,
			fmt.Sprintf("update error for alias %v: %v ", alias.AliasName, err), "home.html")
	}
	log.Infof("[%v] finished the cnames update for %v",
		username, temp.AliasName)

	/****** Update the nodes ******/
	if err := alias.updateNodes(); err != nil {
		return MessageToUser(c, http.StatusBadRequest,
			fmt.Sprintf("update error for alias %v: %v ", alias.AliasName, err), "home.html")
	}
	log.Infof("[%v] finished the nodes update for %v ",
		username, temp.AliasName)

	/****** Update the alarms ******/
	if err := alias.updateAlarms(); err != nil {
		return MessageToUser(c, http.StatusBadRequest,
			fmt.Sprintf("update error for alias %v: %v ", alias.AliasName, err), "home.html")
	}
	log.Infof("[%v] the database was updated successfully, now we can update the DNS", username)

	/****** Update in DNS ******/
	if err = alias.updateDNS(retrieved[0]); err != nil {
		//If something goes wrong while updating, then we use the object
		//we had in DB before the update to restore that state, before the error

		log.Errorf("[%v] could not update %v  in DNS, starting the rollback procedure",
			username, alias.AliasName)
		//Delete the DB updates we just made
		if err = alias.deleteObjectInDB(); err != nil {
			return MessageToUser(c, http.StatusAccepted,
				fmt.Sprintf("failed to delete update for alias %v during rollback", alias.AliasName), "home.html")
		}
		//Recreate the alias as it was before the update
		if err = retrieved[0].createObjectInDB(); err != nil {
			return MessageToUser(c, http.StatusAccepted,
				fmt.Sprintf("failed to restore previous state for alias %v, during rollback", alias.AliasName), "home.html")
		}
		//Successful rollback message
		return MessageToUser(c, http.StatusAccepted,
			fmt.Sprintf("rolled back to previous state completed for alias %v	, please try again later or contact admin", alias.AliasName), "home.html")
	}

	/****** Success message ******/
	return MessageToUser(c, http.StatusAccepted,
		fmt.Sprintf("%v updated Successfully", alias.AliasName), "home.html")

}

/*CheckNameDNS checks if an alias or cname already exist in DB or DNS server
This function serves the immediate check, which is performed while writing
the alias name or cnames*/
func CheckNameDNS(c echo.Context) error {
	var (
		result int64
	)

	aliasToResolve := c.QueryParam("hostname")
	/****** Search cnames with the same name ******/
	db.GetConn().Model(&Cname{}).Where("cname=?", aliasToResolve).Count(&result)
	if result == 0 {
		/****** Search aliases ******/
		db.GetConn().Model(&Alias{}).Where("alias_name=?", aliasToResolve+".cern.ch").Count(&result)
	}
	if result == 0 {
		r, _ := net.LookupHost(aliasToResolve)
		result = int64(len(r))
	}
	return c.JSON(http.StatusOK, result)
}

func checkexistance(alias string) (result []Alias, checks map[string]bool, err error) {
	checks = make(map[string]bool)
	/****** retrieve from db ******/
	result, err = GetObjects(alias)
	if err != nil {
		return
	}
	checks["database"] = len(result) == 0

	/****** check in landb ******/
	entries := landbsoap.Conn().DNSDelegatedSearch(strings.Split(alias, ".")[0] + "*")
	checks["landb"] = len(entries) == 0

	/****** check if secret is in tbag ******/
	secret := auth.GetSecret(alias)
	checks["tbag"] = len(secret) == 0
	return
}
