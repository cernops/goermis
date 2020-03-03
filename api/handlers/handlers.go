package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"

	schema "github.com/gorilla/Schema"
	"github.com/labstack/echo/v4"
	"gitlab.cern.ch/lb-experts/goermis/api/models"
	"gitlab.cern.ch/lb-experts/goermis/db"
)

// TEMPLATE HANDLERS

//CreateHandler handles home page
func CreateHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "create.html", map[string]interface{}{
		"Auth": true,
	})
}

//DeleteHandler handles home page
func DeleteHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "delete.html", map[string]interface{}{
		"Auth": true,
	})
}

//DisplayHandler handles home page
func DisplayHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "display.html", map[string]interface{}{
		"Auth": true,
	})
}

//HomeHandler handles home page
func HomeHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "home.html", map[string]interface{}{
		"Auth": true,
	})
}

//LogsHandler handles home page
func LogsHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "logs.html", map[string]interface{}{
		"Auth": true,
	})
}

//ModifyHandler handles modify page
func ModifyHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "modify.html", map[string]interface{}{
		"Auth": true,
	})

}

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

//AliasesList returns all the aliases
func AliasesList(c echo.Context) error {
	var results objects
	con := db.ManagerDB()
	rows, err := con.Raw("SELECT a.id, alias_name, behaviour, best_hosts, clusters, COALESCE(GROUP_CONCAT(distinct case when r.blacklist = 1 " +
		"then n.node_name else null end),'') AS ForbiddenNodes, COALESCE(GROUP_CONCAT(distinct case when r.blacklist = 0 then n.node_name else null end),'') AS AllowedNodes, " +
		"COALESCE(GROUP_CONCAT(distinct c_name),'') AS cname, external,  a.hostgroup, a.last_modification, metric, polling_interval,tenant, " +
		"ttl, user, statistics FROM alias a left join cname c on ( a.id=c.alias_id) LEFT JOIN relation r on (a.id=r.alias_id) " +
		"LEFT JOIN node n on (n.id=r.node_id)group by a.id, alias_name, behaviour, best_hosts, clusters,  external, a.hostgroup, " +
		"a.last_modification, metric, polling_interval, statistics, tenant, ttl, user ORDER BY alias_name").Rows()

	defer rows.Close()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	for rows.Next() {
		var result result

		err := rows.Scan(&result.ID, &result.AliasName, &result.Behaviour, &result.BestHosts, &result.Clusters,
			&result.ForbiddenNodes, &result.AllowedNodes, &result.Cname, &result.External, &result.Hostgroup,
			&result.LastModification, &result.Metric, &result.PollingInterval,
			&result.Tenant, &result.TTL, &result.User, &result.Statistics)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		results.Objects = append(results.Objects, result)
	}
	return c.JSON(http.StatusOK, results)
}

//GetAlias queries for a specific alias
func GetAlias(c echo.Context) error {
	var results objects
	var tablerow string
	param := c.Param("alias")
	if _, err := strconv.Atoi(c.Param("alias")); err == nil {
		tablerow = "id"
	} else {
		tablerow = "alias_name"
	}
	defer c.Request().Body.Close()
	rows, err := con.Raw("SELECT a.id, alias_name, behaviour, best_hosts, clusters, COALESCE(GROUP_CONCAT(distinct case when r.blacklist = 1 " +
		"then n.node_name else null end),'') AS ForbiddenNodes, COALESCE(GROUP_CONCAT(distinct case when r.blacklist = 0 then n.node_name else null end),'') AS AllowedNodes, " +
		"COALESCE(GROUP_CONCAT(distinct c_name),'') AS cname, external,  a.hostgroup, a.last_modification, metric, polling_interval,tenant, " +
		"ttl, user, statistics FROM alias a left join cname c on ( a.id=c.alias_id) LEFT JOIN relation r on (a.id=r.alias_id) " +
		"LEFT JOIN node n on (n.id=r.node_id) " +
		"where a." + tablerow + " = " + "'" + string(param) + "'" +
		" GROUP BY a.id, alias_name, behaviour, best_hosts, clusters,  external, a.hostgroup, " +
		"a.last_modification, metric, polling_interval, statistics, tenant, ttl, user ORDER BY alias_name ").Rows()

	defer rows.Close()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	for rows.Next() {
		var result result
		err := rows.Scan(&result.ID, &result.AliasName, &result.Behaviour, &result.BestHosts, &result.Clusters,
			&result.ForbiddenNodes, &result.AllowedNodes, &result.Cname, &result.External, &result.Hostgroup,
			&result.LastModification, &result.Metric, &result.PollingInterval,
			&result.Tenant, &result.TTL, &result.User, &result.Statistics)

		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		results.Objects = append(results.Objects, result)
	}

	return c.JSON(http.StatusOK, results)

}

//NewAlias creates a new alias entry in the DB
func NewAlias(c echo.Context) error {

	alias := new(models.Alias)
	params, err := c.FormParams()
	if err != nil {
		panic(err)
	}

	decoder.Decode(alias, params)

	alias.User = "kkouros"
	alias.Behaviour = "mindless"
	alias.Metric = "cmsfrontier"
	alias.PollingInterval = 300
	alias.Statistics = "long"
	alias.Clusters = "none"
	alias.Tenant = "golang"
	alias.LastModification = time.Now()
	spew.Dump(params)

	con.Create(&alias).Association("Cnames")
	if params.Get("cnames") != "" {
		con.Model(alias).Association("Cnames").Append(&models.Cname{CName: params.Get("cnames")})

	}

	return c.Render(http.StatusCreated, "home.html", map[string]interface{}{
		"Auth":    true,
		"Message": alias.AliasName + "  created Successfully",
	})

}

//DeleteAlias is a prototype
func DeleteAlias(c echo.Context) error {
	params, err := c.FormParams()
	if err != nil {
		panic(err)
	}

	con.Where("alias_name = ?", params.Get("alias_name")).Delete(&models.Alias{})

	return c.Render(http.StatusOK, "home.html", map[string]interface{}{
		"Auth":    true,
		"Message": params.Get("alias_name") + "  deleted Successfully",
	})

}

//ModifyAlias is a prototype
func ModifyAlias(c echo.Context) error {
	alias := new(models.Alias)
	params, err := c.FormParams()

	if err != nil {
		panic(err)
	}

	decoder.Decode(alias, params)
	spew.Dump(alias)
	err = con.Model(&models.Alias{}).Where("alias_name = ?", params.Get("alias_name")).Updates(alias).Error
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error)
	}

	return c.Render(http.StatusOK, "home.html", map[string]interface{}{
		"Auth":    true,
		"Message": params.Get("alias_name") + "  updated Successfully",
	})
}

//CheckNameDNS for now is a prototype function that enables frontend to work
func CheckNameDNS(c echo.Context) error {
	return c.JSON(http.StatusOK, 0)
}
