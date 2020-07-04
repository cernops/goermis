package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"gitlab.cern.ch/lb-experts/goermis/aiermis/api"
	"gitlab.cern.ch/lb-experts/goermis/auth"
)

var u user

type user struct {
	username string
	isAdmin  bool
	isGET    bool
	isTeigi  bool
}

func (u *user) set(username string, admin bool, get bool, teigi bool) {
	u.username = username
	u.isAdmin = admin
	u.isGET = get
	u.isTeigi = teigi
}
func currentuser() *user {
	return &u
}
func (u *user) reset() {
	u.username = ""
	u.isAdmin = false
	u.isGET = false
	u.isTeigi = false
}

//CheckAuthorization checks if user is in the egroup and if he is allowed to create in the hostgroup
func CheckAuthorization(nextHandler echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		//Prevent continuous checks for every request for already authorized users
		u := currentuser()
		username := c.Request().Header.Get("X-Forwarded-User")
		if username == u.username {
			if u.isAdmin == true || u.isGET == true || u.isTeigi == true {
				return nextHandler(c)
			}
		}

		if username != "" {
			log.Info("Checking authorization for " + username)
			var d auth.Group
			//If user is in ermis-lbaas-admins proceed
			// to the next handler, without pinging teigi
			if d.CheckCud(username) {
				u.set(username, true, false, false)
				log.Info("[" + username + "] Authorized as admin")
				return nextHandler(c)
			}

			//If user is not in the egroup but method is GET, proceed to the next handler
			if api.StringInSlice(c.Request().Method, []string{"GET"}) {
				u.set(username, false, true, false)
				log.Info("[" + username + "] Authorized as GET request")
				return nextHandler(c)
			}

			//If user is not in the egroup and method is {POST,PATCH,DELETE}, check with teigi
			//If hostgroup is in Request, we use that one. In case there is no hostgroup
			//,such as, when deleting with kermis, we find hostgroup in DB.
			if api.StringInSlice(c.Request().Method, []string{"POST", "DELETE", "PATCH"}) {
				log.Info("[" + username + "] Querying teigi for authorization")

				conn := auth.GetConn()
				if err := conn.InitConnection(); err != nil {
					return api.MessageToUser(c, http.StatusBadRequest, "Failed to initiate teigi connection", "home.html")

				}
				if conn.CheckWithForeman(username, findHostgroup(c)) {
					u.set(username, false, false, true)
					log.Info("[" + username + "] Authorized by teigi")
					return nextHandler(c)
				}

				u.reset()
				return api.MessageToUser(c, http.StatusUnauthorized,
					"Authorization from Teigi failed. User "+username+
						" is not allowed in hostgroup "+findHostgroup(c), "home.html")

			}

			u.reset()
			return api.MessageToUser(c, http.StatusUnauthorized, "Authorization failed for "+username, "home.html")
		}

		u.reset()
		return api.MessageToUser(c, http.StatusUnauthorized,
			"Authorization failed. No username provided", "home.html")
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
