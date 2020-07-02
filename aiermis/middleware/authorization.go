package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"gitlab.cern.ch/lb-experts/goermis/aiermis/api"
	"gitlab.cern.ch/lb-experts/goermis/auth"
)

//CheckAuthorization checks if user is in the egroup and if he is allowed to create in the hostgroup
func CheckAuthorization(nextHandler echo.HandlerFunc) echo.HandlerFunc {
	home := "home.html"
	return func(c echo.Context) error {
		username := c.Request().Header.Get("X-Forwarded-User")
		conn := auth.GetConn()
		var d auth.Group
		if err := conn.InitConnection(); err != nil {
			return api.MessageToUser(c, http.StatusBadRequest, "Failed to initiate teigi connection", home)

		}
		//spew.Dump(c.Request())
		//This username check most probably will change with OIDC
		if username != "" {
			log.Debug("[" + username + "] Checking authorization")
			//If user is in ermis-lbaas-admins proceed
			// to the next handler, without pinging teigi
			if d.CheckCud(username) {
				log.Debug("[" + username + "] Authorized as admin")
				return nextHandler(c)
			}
			//If user is not in the egroup but method is GET, proceed to the next handler
			if api.StringInSlice(c.Request().Method, []string{"GET"}) {
				log.Debug("[" + username + "] Authorized as GET request")
				return nextHandler(c)
			}
			//If user is not in the egroup and method is {POST,PATCH,DELETE}, check with teigi
			//If hostgroup is in Request, we use that one. In case there is no hostgroup
			//,such as, when deleting with kermis, we find hostgroup in DB.
			if api.StringInSlice(c.Request().Method, []string{"POST", "DELETE", "PATCH"}) {
				log.Debug("[" + username + "] Querying teigi for authorization")
				if conn.CheckWithForeman(username, findHostgroup(c)) {
					log.Debug("[" + username + "] Authorized by teigi")
					return nextHandler(c)
				}

				return api.MessageToUser(c, http.StatusUnauthorized,
					"Authorization from Teigi failed. User "+username+
						" is not allowed in hostgroup "+findHostgroup(c), home)

			}

			return api.MessageToUser(c, http.StatusUnauthorized, "Authorization failed for "+username, home)
		}

		return api.MessageToUser(c, http.StatusUnauthorized,
			"Authorization failed. No username provided", home)
	}
}

func findHostgroup(c echo.Context) string {
	hostgroup := c.FormValue("hostgroup")
	if hostgroup == "" {
		alias, _ := api.GetObjects(c.FormValue("alias_name"), "alias_name")
		hostgroup = alias[0].Hostgroup
	}
	return hostgroup
}
