package models

import (
	"net/http"

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

var con = db.ManagerDB()

//GetObjectsList return list of aliases
func (r Objects) GetObjectsList() (Objects, error) {

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
		return r, echo.NewHTTPError(http.StatusBadRequest, err.Error())

	}
	for rows.Next() {
		var result result

		err := rows.Scan(&result.ID, &result.AliasName, &result.Behaviour, &result.BestHosts, &result.Clusters,
			&result.ForbiddenNodes, &result.AllowedNodes, &result.Cname, &result.External, &result.Hostgroup,
			&result.LastModification, &result.Metric, &result.PollingInterval,
			&result.Tenant, &result.TTL, &result.User, &result.Statistics)
		if err != nil {

			return r, echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		r.Objects = append(r.Objects, result)
	}
	return r, nil
}

//GetObject returns a single alias
func (r Objects) GetObject(param string, tablerow string) (Objects, error) {

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
		return r, err
	}
	for rows.Next() {
		var result result
		err := rows.Scan(&result.ID, &result.AliasName, &result.Behaviour, &result.BestHosts, &result.Clusters,
			&result.ForbiddenNodes, &result.AllowedNodes, &result.Cname, &result.External, &result.Hostgroup,
			&result.LastModification, &result.Metric, &result.PollingInterval,
			&result.Tenant, &result.TTL, &result.User, &result.Statistics)

		if err != nil {
			return r, err
		}
		r.Objects = append(r.Objects, result)
	}
	return r, err
}
