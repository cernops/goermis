package ermis

/*This file contains the functions that transform the binded
data from the Resource struct to the ORM one(and vice versa)
This is done to allow a more Object oriented experience later on */
import (
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

/* Resource struct facilitates the binding of data from requests

form tags --> for binding the form fields from UI,
json tags -- > binding data sent by kermis CLI
valid tag--> validation rules, extra funcs in the common.go file*/

type (
	//Resource deals with the output from the queries
	Resource struct {
		ID               int       `json:"alias_id"`
		AliasName        string    `json:"alias_name"        form:"alias_name"    `
		Behaviour        string    `json:"behaviour" `
		BestHosts        int       `json:"best_hosts"        form:"best_hosts" `
		Clusters         string    `json:"clusters"  `
		ForbiddenNodes   []string  `json:"ForbiddenNodes"    form:"ForbiddenNodes" `
		AllowedNodes     []string  `json:"AllowedNodes"      form:"AllowedNodes" `
		Cnames           []string  `json:"cnames"            form:"cnames"`
		External         string    `json:"external"          form:"external"`
		Hostgroup        string    `json:"hostgroup"         form:"hostgroup"`
		LastModification time.Time `json:"last_modification" `
		Metric           string    `json:"metric"            `
		PollingInterval  int       `json:"polling_interval"  `
		Tenant           string    `json:"tenant"            `
		TTL              int       `json:"ttl,omitempty"     `
		User             string    `json:"user"              `
		Statistics       string    `json:"statistics"        `
		ResourceURI      string    `json:"resource_uri"      `
		Pwned            bool      `json:"pwned"             `
		Alarms           []string  `json:"alarms"            form:"alarms"`
	}
	//Objects holds multiple result structs
	Objects struct {
		Objects []Resource `json:"objects"`
	}
)

/*sanitazeInCreation stretches the freshly binded data into
the ORM models, giving us a more appropriate object to work on*/
func sanitazeInCreation(c echo.Context, resource Resource) (object Alias) {

	contentType := c.Request().Header.Get("Content-Type")
	//Cnames
	object.Cnames = []Cname{}
	if len(resource.Cnames) != 0 {
		split := Explode(contentType, resource.Cnames)
		for _, cname := range split {
			object.Cnames = append(object.Cnames, Cname{Cname: cname})
		}

	}

	//Alias name hydration
	if !strings.HasSuffix(resource.AliasName, ".cern.ch") {
		object.AliasName = resource.AliasName + ".cern.ch"
	} else {
		object.AliasName = resource.AliasName
	}

	if resource.Hostgroup != "" {
		object.Hostgroup = resource.Hostgroup
	}
	object.User = GetUsername()

	if resource.BestHosts != 0 {
		object.BestHosts = resource.BestHosts
	}
	//View
	if StringInSlice(strings.ToLower(resource.External), []string{"yes", "external"}) {
		object.External = "yes"
	} else if StringInSlice(strings.ToLower(resource.External), []string{"no", "internal"}) {
		object.External = "no"
	}

	//Default values while creating
	object.Statistics = "long"
	object.Tenant = "golang"
	//Time
	object.LastModification.Time = time.Now()
	object.LastModification.Valid = true
	object.Metric = "cmsfrontier"
	object.PollingInterval = 300
	object.TTL = 60
	object.Clusters = "none"
	object.Behaviour = "mindless"

	return
}

/*parse accomplishes the reverse action, it receives
the ORM model(s) of alias(s) and "packages" the into a more
readable format
EXTRA FUNCTIONALITIES:
1. The construction of the URI
2. The assignment of Pwned value, which determines the
   alias filtration based on ownership in UI
   NOTE: Alias filtration is done with an extra field
   because we still need to show the full list to the user
   and prevent modification on not owned aliases*/
func parse(queryResults []Alias) Objects {
	var (
		parsed Objects
	)
	if queryResults == nil {
		return parsed
	}
	for _, element := range queryResults {
		var temp Resource
		//The ones that are the same
		temp.ID = element.ID
		temp.AliasName = element.AliasName
		temp.Behaviour = element.Behaviour
		temp.BestHosts = element.BestHosts
		temp.Clusters = element.Clusters
		temp.Hostgroup = element.Hostgroup
		temp.External = element.External
		//Time
		temp.LastModification = element.LastModification.Time
		temp.Metric = element.Metric
		temp.PollingInterval = element.PollingInterval
		temp.TTL = element.TTL
		temp.Tenant = element.Tenant
		temp.ResourceURI = "/p/api/v1/alias/" + strconv.Itoa(element.ID)
		temp.User = element.User
		temp.Statistics = element.Statistics

		//The cnames
		temp.Cnames = []string{}
		if len(element.Cnames) != 0 {
			for _, v := range element.Cnames {
				temp.Cnames = append(temp.Cnames, v.Cname)
			}

		}

		//The alarms
		temp.Alarms = []string{}
		if len(element.Alarms) != 0 {
			for _, v := range element.Alarms {
				alarm := v.Name + ":" +
					v.Recipient + ":" +
					strconv.Itoa(v.Parameter) + ":" +
					strconv.FormatBool(v.Active)
				if v.LastActive.Valid {
					alarm += ":" + v.LastActive.Time.String()
				}
				temp.Alarms = append(temp.Alarms, alarm)
			}

		}

		//The nodes
		temp.ForbiddenNodes = []string{}
		temp.AllowedNodes = []string{}
		if len(element.Relations) != 0 {
			for _, v := range element.Relations {
				nameload := v.Node.NodeName + ":" + strconv.Itoa(v.Load)
				if v.Blacklist {
					temp.ForbiddenNodes = append(temp.ForbiddenNodes, nameload)
				} else {
					temp.AllowedNodes = append(temp.AllowedNodes, nameload)
				}

			}

		}

		//Set the pwn value(true/false)
		//Sole purpose of pwned field is to be used in the UI for alias filtering
		//Ermis-lbaas-admins are superusers
		if IsSuperuser() {
			temp.Pwned = true
		} else {
			temp.Pwned = StringInSlice(temp.Hostgroup, GetUsersHostgroups())
		}

		parsed.Objects = append(parsed.Objects, temp)
	}
	return parsed
}

/*sanitazeInUpdate generates the ORM model of an alias from
the binded request data*/
func sanitazeInUpdate(c echo.Context, current Alias, new Resource) (Alias, error) {
	contentType := c.Request().Header.Get("Content-Type")
	//Cnames
	current.Cnames = []Cname{}
	if len(new.Cnames) != 0 {
		split := Explode(contentType, new.Cnames)
		for _, cname := range split {
			current.Cnames = append(current.Cnames,
				Cname{
					Cname:        cname,
					CnameAliasID: current.ID})
		}

	}

	//Alarms
	current.Alarms = []Alarm{}
	if len(new.Alarms) != 0 {
		split := Explode(contentType, new.Alarms)
		for _, alarm := range split {
			element := DeleteEmpty(strings.Split(alarm, ":"))
			//Convert param from string to int
			param, err := strconv.Atoi(element[2])
			if err != nil {
				return Alias{}, err
			}
			current.Alarms = append(current.Alarms, Alarm{
				Name:         element[0],
				Recipient:    element[1],
				Parameter:    param,
				AlarmAliasID: current.ID,
				Alias:        current.AliasName})

		}
	}

	//Nodes
	//Keep a copy of relations, that we need to determine the ID of existing relations, with assigned ID
	//If we didnt do this, what would happen is that a new ID=0 would be assigned.
	copyOfRelations := current.Relations
	current.Relations = []Relation{}
	fields := map[bool][]string{false: new.AllowedNodes, true: new.ForbiddenNodes}
	for k, field := range fields {
		if len(field) != 0 {
			split := Explode(contentType, field)
			for _, node := range split {
				current.Relations = append(current.Relations, Relation{
					AliasID:   current.ID,
					NodeID:    FindNodeID(node, copyOfRelations),
					Blacklist: k,
					Node: &Node{
						ID:       FindNodeID(node, copyOfRelations),
						NodeName: node,
						LastModification: sql.NullTime{
							Time:  time.Now(),
							Valid: true,
						},
					}})
			}
		}
	}

	if new.External != "" {
		if StringInSlice(new.External, []string{"yes", "external"}) {
			current.External = "yes"
		} else if StringInSlice(new.External, []string{"no", "internal"}) {
			current.External = "no"
		}
	}

	current.User = GetUsername()

	if new.BestHosts != 0 {
		current.BestHosts = new.BestHosts
	}
	if new.Metric != "" {
		current.Metric = new.Metric
	}
	if new.PollingInterval != 0 {
		current.PollingInterval = new.PollingInterval
	}
	if new.Hostgroup != "" {
		current.Hostgroup = new.Hostgroup
	}
	if new.Tenant != "" {
		current.Tenant = new.Tenant
	}

	if new.TTL != 0 {
		current.TTL = new.TTL
	}

	return current, nil
}
