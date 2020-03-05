package handlers

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/davecgh/go-spew/spew"
	schema "github.com/gorilla/Schema"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"gitlab.cern.ch/lb-experts/goermis/api/common"
	"gitlab.cern.ch/lb-experts/goermis/api/models"
	"gitlab.cern.ch/lb-experts/goermis/db"
)

//CRUD HANDLERS

type result struct {
	ID               uint
	AliasName        string `json:"alias_name"`
	Behaviour        string `json:"behaviour"`
	BestHosts        int    `json:"best_hosts"`
	Clusters         string `json:"clusters"`
	ForbiddenNodes   string `json:"ForbiddenNodes"`
	AllowedNodes     string `json:"AllowedNodes"`
	Cname            string `json:"cnames"`
	External         string `json:"external"`
	Hostgroup        string `json:"hostgroup"`
	LastModification string `json:"last_modification"`
	Metric           string `json:"metric"`
	PollingInterval  int    `json:"polling_interval"`
	Tenant           string `json:"tenant"`
	TTL              int    `json:"ttl"`
	User             string `json:"user"`
	Statistics       string `json:"statistics"`
}

type objects struct {
	Objects []result `json:"objects"`
}

var decoder = schema.NewDecoder()
var con = db.ManagerDB()

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
	response, err := models.GetObjectsList()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, response)
}

//GetAlias queries for a specific alias
func GetAlias(c echo.Context) error {

	var tablerow string
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

	response, err := models.GetObject(string(param), tablerow)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, response)

}

//NewAlias creates a new alias entry in the DB
func NewAlias(c echo.Context) error {

	alias := new(models.Alias)
	defer c.Request().Body.Close()

	//Get the params from the form
	params, err := c.FormParams()
	cnames := params.Get("cnames")
	if err != nil {
		panic(err)
	}

	//Populate the struct with the user-defined data
	decoder.Decode(alias, params)

	//Populate the struct with the default values
	alias.User = "kkouros"
	alias.Behaviour = "mindless"
	alias.Metric = "cmsfrontier"
	alias.PollingInterval = 300
	alias.Statistics = "long"
	alias.Clusters = "none"
	alias.Tenant = "golang"
	alias.LastModification = time.Now()

	err = models.NewObject(alias, cnames)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error)

	}

	return c.Render(http.StatusCreated, "home.html", map[string]interface{}{
		"Auth":    true,
		"Message": alias.AliasName + "  created Successfully",
	})

}

//DeleteAlias is a prototype
func DeleteAlias(c echo.Context) error {
	var cnames bool
	//Get the params from the form
	aliasName := c.FormValue("alias_name")
	defer c.Request().Body.Close()

	if c.FormValue("cnames") != "" {
		cnames = true
	}

	err := models.DeleteObject(aliasName, cnames)

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
	//Prepare cnames separately
	var cnames []string
	cnames = common.DeleteEmpty(strings.Split(c.FormValue("cnames"), ","))

	//Declare the model that will be used
	alias := new(models.Alias)

	//Update external and hostgroup fields
	con.Model(&alias).Where("alias_name = ?", c.FormValue("alias_name")).Take(&alias).UpdateColumns(
		map[string]interface{}{
			"external":  c.FormValue("external"),
			"hostgroup": c.FormValue("hostgroup"),
		},
	)

	//WIP: If there are no cnames from UI , delete them all, otherwise append them
	if len(cnames) > 0 {
		for _, cname := range cnames {
			spew.Dump(cname)
			con.Model(alias).Association("Cnames").Append(&models.Cname{CName: cname})
		}
	} else {
		con.Where("alias_id = ?", alias.ID).Delete(&models.Cname{})

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
