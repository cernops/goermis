package models

import (
	"net/http"
	"strings"

	schema "github.com/gorilla/Schema"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"gitlab.cern.ch/lb-experts/goermis/db"
)

type (
	result struct {
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
	//Objects holds multiple result structs
	Objects struct {
		Objects []result `json:"objects"`
	}
)

var decoder = schema.NewDecoder()
var con = db.ManagerDB()

//GetObjectsList return list of aliases
func GetObjectsList() (Objects, error) {

	var results Objects
	con := db.ManagerDB()

	rows, err := con.Raw("SELECT a.id, alias_name, behaviour, best_hosts, clusters, COALESCE(GROUP_CONCAT(distinct case when r.blacklist = 1 " +
		"then n.node_name else null end),'') AS ForbiddenNodes, COALESCE(GROUP_CONCAT(distinct case when r.blacklist = 0 then n.node_name else null end),'') AS AllowedNodes, " +
		"COALESCE(GROUP_CONCAT(distinct c_name),'') AS cname, external,  a.hostgroup, a.last_modification, metric, polling_interval,tenant, " +
		"ttl, user, statistics FROM alias a left join cname c on ( a.id=c.alias_id) LEFT JOIN relation r on (a.id=r.alias_id) " +
		"LEFT JOIN node n on (n.id=r.node_id)group by a.id, alias_name, behaviour, best_hosts, clusters,  external, a.hostgroup, " +
		"a.last_modification, metric, polling_interval, statistics, tenant, ttl, user ORDER BY alias_name").Rows()

	defer rows.Close()
	if err != nil {
		log.Error(err)
		return results, echo.NewHTTPError(http.StatusBadRequest, err.Error())

	}
	for rows.Next() {
		var result result

		err := rows.Scan(&result.ID, &result.AliasName, &result.Behaviour, &result.BestHosts, &result.Clusters,
			&result.ForbiddenNodes, &result.AllowedNodes, &result.Cname, &result.External, &result.Hostgroup,
			&result.LastModification, &result.Metric, &result.PollingInterval,
			&result.Tenant, &result.TTL, &result.User, &result.Statistics)
		if err != nil {

			return results, echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		results.Objects = append(results.Objects, result)
	}
	return results, err
}

//GetObject returns a single alias
func GetObject(param string, tablerow string) (Objects, error) {
	var results Objects
	rows, err := con.Raw("SELECT a.id, alias_name, behaviour, best_hosts, clusters, COALESCE(GROUP_CONCAT(distinct case when r.blacklist = 1 " +
		"then n.node_name else null end),'') AS ForbiddenNodes, COALESCE(GROUP_CONCAT(distinct case when r.blacklist = 0 then n.node_name else null end),'') AS AllowedNodes, " +
		"COALESCE(GROUP_CONCAT(distinct c_name),'') AS cname, external,  a.hostgroup, a.last_modification, metric, polling_interval,tenant, " +
		"ttl, user, statistics FROM alias a left join cname c on ( a.id=c.alias_id) LEFT JOIN relation r on (a.id=r.alias_id) " +
		"LEFT JOIN node n on (n.id=r.node_id) " +
		"where a." + tablerow + " = " + "'" + param + "'" +
		" GROUP BY a.id, alias_name, behaviour, best_hosts, clusters,  external, a.hostgroup, " +
		"a.last_modification, metric, polling_interval, statistics, tenant, ttl, user ORDER BY alias_name ").Rows()

	defer rows.Close()
	if err != nil {
		return results, err
	}
	for rows.Next() {
		var result result
		err := rows.Scan(&result.ID, &result.AliasName, &result.Behaviour, &result.BestHosts, &result.Clusters,
			&result.ForbiddenNodes, &result.AllowedNodes, &result.Cname, &result.External, &result.Hostgroup,
			&result.LastModification, &result.Metric, &result.PollingInterval,
			&result.Tenant, &result.TTL, &result.User, &result.Statistics)

		if err != nil {
			return results, err
		}
		results.Objects = append(results.Objects, result)
	}
	return results, err
}

//NewObject creates an alias
func NewObject(alias *Alias, cnames string) (err error) {

	con.Create(&alias)
	if cnames != "" {
		cnames := strings.Split(cnames, ",")
		for _, cname := range cnames {
			con.Model(alias).Association("Cnames").Append(&Cname{CName: string(cname)})
		}
	}
	return err
}

//DeleteObject deletes an alias and its Relations
/////WIP
func DeleteObject(alias string) (err error) {

	con.Model(&alias).Association("Cnames").Clear()
	con.Where("alias_name = ?", alias).Delete(&Alias{})

	return err
}
