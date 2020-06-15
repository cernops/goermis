package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.cern.ch/lb-experts/goermis/auth"
)

//Unless the method is POST, teigi is assumed as true.
var (
	teigi = true
	ldap  = false
)

//CheckAuthorization checks if user is in the egroup and if he is allowed to create in the hostgroup
func CheckAuthorization(nextHandler echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := "kkouros"
		hostgroup := c.FormValue("hostgroup")
		conn := auth.GetConn()
		var d auth.Group
		if err := conn.InitConnection(); err != nil {

			return messageToUser(c, http.StatusBadRequest, "Failed to initiate teigi connection")

		}
		if username != "" {
			if d.CheckCrud(username) {
				if hostgroup != "" && c.Request().Method == "POST" {
					if conn.CheckWithForeman(username, hostgroup) {

						return nextHandler(c)
					}
					return messageToUser(c, http.StatusUnauthorized,
						"Authorization from Teigi failed. User "+username+
							" is not allowed in hostgroup "+hostgroup)

				}
				return nextHandler(c)
			}

			return messageToUser(c, http.StatusUnauthorized,
				"Authorization from LDAP failed. User "+username+
					" is not part of the e-group")

		}

		return messageToUser(c, http.StatusUnauthorized, "Authorization failed. No username provided")

	}
}
func messageToUser(c echo.Context, status int, message string) error {
	return c.Render(status, "home.html", map[string]interface{}{
		"Auth":    true,
		"User":    username,
		"Message": message,
	})

}
