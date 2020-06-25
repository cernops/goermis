package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"gitlab.cern.ch/lb-experts/goermis/api/models"
	"gitlab.cern.ch/lb-experts/goermis/auth"
)

//CheckAuthorization checks if user is in the egroup and if he is allowed to create in the hostgroup
func CheckAuthorization(nextHandler echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := c.Request().Header.Get("X-Forwarded-User")
		conn := auth.GetConn()
		var d auth.Group
		if err := conn.InitConnection(); err != nil {
			return messageToUser(c, http.StatusBadRequest, "Failed to initiate teigi connection")

		}
		//spew.Dump(c.Request())
		//This username check most probably will change with OIDC
		if username != "" {
			log.Info("[" + username + "] Checking authorization")
			//If user is in ermis-lbaas-admins proceed
			// to the next handler, without pinging teigi
			if d.CheckCud(username) {
				log.Info("[" + username + "] Authorized as admin")
				return nextHandler(c)
			}
			//If user is not in the egroup but method is GET, proceed to the next handler
			if models.StringInSlice(c.Request().Method, []string{"GET"}) {
				log.Info("[" + username + "] Authorized as GET request")
				return nextHandler(c)
			}
			//If user is not in the egroup and method is {POST,PATCH,DELETE}, check with teigi
			//If hostgroup is in Request, we use that one. In case there is no hostgroup
			//,such as, when deleting with kermis, we find hostgroup in DB.
			if models.StringInSlice(c.Request().Method, []string{"POST", "DELETE", "PATCH"}) {
				log.Info("[" + username + "] Querying teigi for authorization")
				if conn.CheckWithForeman(username, findHostgroup(c)) {
					log.Info("[" + username + "] Authorized by teigi")
					return nextHandler(c)
				}

				return messageToUser(c, http.StatusUnauthorized,
					"Authorization from Teigi failed. User "+username+
						" is not allowed in hostgroup "+findHostgroup(c))

			}

			return messageToUser(c, http.StatusUnauthorized, "Authorization failed for "+username)
		}

		return messageToUser(c, http.StatusUnauthorized, "Authorization failed. No username provided")
	}
}

func messageToUser(c echo.Context, status int, message string) error {
	log.Info(message)
	return c.Render(status, "home.html", map[string]interface{}{
		"Auth":    true,
		"User":    c.Request().Header.Get("X-Forwarded-User"),
		"Message": message,
	})

}
func findHostgroup(c echo.Context) string {
	hostgroup := c.FormValue("hostgroup")
	if hostgroup == "" {
		alias, _ := models.GetObjects(c.FormValue("alias_name"), "alias_name")
		hostgroup = alias[0].Hostgroup
	}
	return hostgroup
}
