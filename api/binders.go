package api

/*This file contains the functions that transform the binded
data from the Resource struct to the ORM one(and vice versa)
This is done to allow a more Object oriented experience later on */
import (
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
		ID               int       `json:"alias_id"          valid:"-"`
		AliasName        string    `json:"alias_name"        valid:"required,dns"                           form:"alias_name"    `
		Behaviour        string    `json:"behaviour"         valid:"-"`
		BestHosts        int       `json:"best_hosts"        valid:"required,int,best_hosts"                form:"best_hosts" `
		Clusters         string    `json:"clusters"          valid:"optional,alphanum"`
		ForbiddenNodes   []string  `json:"ForbiddenNodes"    valid:"optional,nodes"                         form:"ForbiddenNodes" `
		AllowedNodes     []string  `json:"AllowedNodes"      valid:"optional,nodes"                         form:"AllowedNodes" `
		Cnames           []string  `json:"cnames"            valid:"optional,cnames"                        form:"cnames"`
		External         string    `json:"external"          valid:"required,in(yes|no|external|internal)"                    form:"external"`
		Hostgroup        string    `json:"hostgroup"         valid:"required,hostgroup"                     form:"hostgroup"`
		LastModification time.Time `json:"last_modification" valid:"-"`
		Metric           string    `json:"metric"            valid:"in(cmsfrontier),optional"`
		PollingInterval  int       `json:"polling_interval"  valid:"optional,numeric"`
		Tenant           string    `json:"tenant"            valid:"optional,alphanum"`
		TTL              int       `json:"ttl,omitempty"     valid:"optional,numeric"`
		User             string    `json:"user"              valid:"optional,alphanum"`
		Statistics       string    `json:"statistics"        valid:"optional,alpha"`
		ResourceURI      string    `json:"resource_uri"      valid:"-"`
		Pwned            bool      `json:"pwned"             valid:"-"`
		Alarms           []string  `json:"alarms"            valid:"optional, alarms"                                      form:"alarms"`
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
		split := explode(contentType, resource.Cnames)
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
	if stringInSlice(strings.ToLower(resource.External), []string{"yes", "external"}) {
		object.External = "yes"
	} else if stringInSlice(strings.ToLower(resource.External), []string{"no", "internal"}) {
		object.External = "no"
	}

	//Default values while creating
	object.Statistics = "long"
	object.Tenant = "golang"
	object.LastModification = time.Now()
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
		temp.LastModification = element.LastModification
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
				if v.Blacklist == true {
					temp.ForbiddenNodes = append(temp.ForbiddenNodes, v.Node.NodeName)
				} else {
					temp.AllowedNodes = append(temp.AllowedNodes, v.Node.NodeName)
				}

			}

		}

		//Set the pwn value(true/false)
		//Sole purpose of pwned field is to be used in the UI for alias filtering
		//Ermis-lbaas-admins are superusers
		if IsSuperuser() {
			temp.Pwned = true
		} else {
			temp.Pwned = stringInSlice(temp.Hostgroup, GetUsersHostgroups())
		}

		parsed.Objects = append(parsed.Objects, temp)
	}
	return parsed
}

/*sanitazeInUpdate generates the ORM model of an alias from
the binded request data*/
func sanitazeInUpdate(c echo.Context, current Alias, new Resource) Alias {
	contentType := c.Request().Header.Get("Content-Type")
	//Cnames
	current.Cnames = []Cname{}
	if len(new.Cnames) != 0 {
		split := explode(contentType, new.Cnames)
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
		split := explode(contentType, new.Alarms)
		for _, alarm := range split {
			element := deleteEmpty(strings.Split(alarm, ":"))
			current.Alarms = append(current.Alarms, Alarm{
				Name:         element[0],
				Recipient:    element[1],
				Parameter:    stringToInt(element[2]),
				AlarmAliasID: current.ID,
				Alias:        current.AliasName})

		}
	}

	//Nodes
	current.Relations = []*Relation{}
	fields := map[bool][]string{false: new.AllowedNodes, true: new.ForbiddenNodes}
	for k, field := range fields {
		if len(field) != 0 {
			split := explode(contentType, field)
			for _, node := range split {
				current.Relations = append(current.Relations, &Relation{
					AliasID:   current.ID,
					NodeID:    findNodeID(node),
					Blacklist: k,
					Node: &Node{
						ID:               findNodeID(node),
						NodeName:         node,
						LastModification: time.Now(),
						Hostgroup:        current.Hostgroup}})
			}
		}
	}

	if new.External != "" {
		if stringInSlice(new.External, []string{"yes", "external"}) {
			current.External = "yes"
		} else if stringInSlice(new.External, []string{"no", "internal"}) {
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

	return current
}
