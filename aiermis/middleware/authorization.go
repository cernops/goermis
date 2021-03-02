package middleware

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"gitlab.cern.ch/lb-experts/goermis/aiermis/api"
	"gitlab.cern.ch/lb-experts/goermis/aiermis/common"
)

//CheckAuthorization checks if user is in the egroup and if he is allowed to create in the hostgroup
func CheckAuthorization(nextHandler echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := c.Request().Header.Get("X-Forwarded-User")
		if username != api.GetUsername() {
			api.SetUser(username)
		}
		if api.GetUsername() != "" {
			//Ermis-lbaas-admins are superusers
			if api.IsSuperuser() {
				return nextHandler(c)
				//If user is not in the egroup but method is GET, proceed to the next handler
			} else if c.Request().Method == "GET" {
				return nextHandler(c)
			} else {
				return askTeigi(c, nextHandler, api.GetUsername())
			}
		}
		return common.MessageToUser(c, http.StatusUnauthorized,
			"Authorization failed. No username provided", "home.html")
	}
}

func askTeigi(c echo.Context, nextHandler echo.HandlerFunc, username string) error {
	var (
		authInNewHg bool
		authInOldHg bool
	)

	//We extract the hostgroup values from the Req Body and the one in DB, for the same alias.
	newHg, oldHg, err := findHostgroup(c)
	if err != nil {
		return common.MessageToUser(c, http.StatusBadRequest,
			"Failed to process hostgroups "+err.Error(), "home.html")

	}

	//In this step we check username , against both hostgroups.
	//We need this step to prevent unauthorized alias movements.
	if newHg != "" {
		authInNewHg = common.StringInSlice(newHg, api.GetUsersHostgroups())
	}
	if oldHg != "" {
		authInOldHg = common.StringInSlice(oldHg, api.GetUsersHostgroups())
	}

	switch c.Request().Method {
	//1.In case method is PATCH...
	case "PATCH":
		//...and there is no hostgroup field in the Request,allow to PATCH other fields
		//if the user is authorized in the old hostgroup
		if newHg == "" && authInOldHg {
			log.Info("[" + api.GetUsername() + "] Authorized by teigi for PATCH, using existing hostgroup")
			return nextHandler(c)
			//When PATCH-ing hostgroup value itself, verify user in both hostgroups
		} else if authInNewHg && authInOldHg {
			log.Info("[" + api.GetUsername() + "] Authorized by teigi for PATCH, using both hg")
			return nextHandler(c)
		}
		return common.MessageToUser(c, http.StatusUnauthorized,
			api.GetUsername()+" is unauthorized to PATCH in hostgroup "+oldHg, "home.html")
		//2.In case method is POST...
	case "POST":
		//Here we authorize the creation of new aliases(no hostgroup value in DB),
		// if teigi gives the OK for the new hostgroup value.
		if authInNewHg && oldHg == "" {
			log.Info("[" + api.GetUsername() + "] Authorized by teigi to POST new alias")
			return nextHandler(c)

			//When modifying , check both hostgroups
		} else if authInNewHg && authInOldHg {
			log.Info("[" + api.GetUsername() + "] Authorized by teigi for POST")
			return nextHandler(c)
		}
		return common.MessageToUser(c, http.StatusUnauthorized,
			api.GetUsername()+" is unauthorized to POST in hostgroup "+oldHg, "home.html")
		// 3.In case method is DELETE...
	case "DELETE":
		//We make sure user is auth in the existing hg
		if authInOldHg {
			log.Info("[" + api.GetUsername() + "] Authorized by teigi for DELETE")
			return nextHandler(c)
		}
		return common.MessageToUser(c, http.StatusUnauthorized,
			api.GetUsername()+" is unauthorized to DELETE from hostgroup "+oldHg, "home.html")
	default:
		return common.MessageToUser(c, http.StatusMethodNotAllowed,
			"Method "+c.Request().Method, "home.html")

	}
}

func findHostgroup(c echo.Context) (newHg string, oldHg string, err error) {
	type body struct {
		Alias     string `json:"alias_name"`
		Hostgroup string `json:"hostgroup,omitempty"` //if there is no hostgroup provided, don't panic
	}
	var (
		aliasToquery string
	)
	//Kermis and Behave tests send Content-Type = application/json
	if c.Request().Header.Get("Content-Type") == "application/json" {
		if c.Request().Method == "DELETE" {
			aliasToquery = c.QueryParam("alias_name")
		} else {
			// Read body
			raw, err := ioutil.ReadAll(c.Request().Body)
			if err != nil {
				return "", "", err
			}

			//Restore body, so we can bind again in the handlers
			c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(raw))

			// Unmarshal body of request
			var b body
			err = json.Unmarshal(raw, &b)
			if err != nil {
				return "", "", err
			}
			//Set values we need
			aliasToquery = b.Alias
			newHg = b.Hostgroup
		}
		//UI sends Content-Type x-www-form-urlencoded
	} else if c.Request().Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
		newHg = c.FormValue("hostgroup")
		aliasToquery = c.FormValue("alias_name")
	}

	//Get the hostgroup that is registered for the same alias.
	alias, _ := api.GetObjects(aliasToquery)
	if alias != nil {
		oldHg = alias[0].Hostgroup
	}

	//In case the hostgroup fields are empty in the request and the DB
	//we throw a bad request error, because this is a scenario we don't want.
	if newHg == "" && oldHg == "" {
		return "", "", common.MessageToUser(c, http.StatusBadRequest,
			"Not allowed to modify/create/delete without hostgroup", "home.html")

	}

	return newHg, oldHg, nil
}
