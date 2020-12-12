package api

import (
	"strconv"
	"strings"
	"time"

	"gitlab.cern.ch/lb-experts/goermis/aiermis/common"
)

type (
	//Resource deals with the output from the queries
	Resource struct {
		ID               int       `                         json:"alias_id"          valid:"-"`
		AliasName        string    `form:"alias_name"        json:"alias_name"        valid:"required,dns"`
		Behaviour        string    `                         json:"behaviour"         valid:"-"`
		BestHosts        int       `form:"best_hosts"        json:"best_hosts"        valid:"required,int,best_hosts"`
		Clusters         string    `                         json:"clusters"          valid:"alphanum"`
		ForbiddenNodes   []string  `form:"ForbiddenNodes"    json:"ForbiddenNodes"    valid:"optional,nodes"                    gorm:"not null"`
		AllowedNodes     []string  `form:"AllowedNodes"      json:"AllowedNodes"      valid:"optional,nodes"                    gorm:"not null"`
		Cname            []string  `form:"cnames"            json:"cnames"            valid:"optional,cnames"                   gorm:"not null"`
		External         string    `form:"external"          json:"external"          valid:"required,in(yes|no)"`
		Hostgroup        string    `form:"hostgroup"         json:"hostgroup"         valid:"required,hostgroup"`
		LastModification time.Time `                         json:"last_modification" valid:"-"`
		Metric           string    `                         json:"metric"            valid:"in(cmsfrontier|minino|minimum|),optional"`
		PollingInterval  int       `                         json:"polling_interval"  valid:"numeric"`
		Tenant           string    `                         json:"tenant"            valid:"optional,alphanum"`
		TTL              int       `                         json:"ttl,omitempty"     valid:"numeric,optional"`
		User             string    `                         json:"user"              valid:"optional,alphanum"`
		Statistics       string    `                         json:"statistics"        valid:"alpha"`
		ResourceURI      string    `                         json:"resource_uri"      valid:"-"`
		Pwned            bool      `                         json:"pwned"             valid:"-"`
		Alarms           []string  `form:"alarms"            json:"alarms"             valid:"-"`
	}
)

func sanitazeInCreation(resource Resource) (object Alias) {
	var (
		cnames []Cname
	)
	//Cnames
	if len(resource.Cname) != 0 {
		split := common.DeleteEmpty(strings.Split(resource.Cname[0], ","))
		for _, cname := range split {
			cnames = append(cnames, Cname{Cname: cname})
		}
	}
	//Alias name hydration
	if !strings.HasSuffix(resource.AliasName, ".cern.ch") {
		object.AliasName = resource.AliasName + ".cern.ch"
	} else {
		object.AliasName = resource.AliasName
	}

	//View
	if common.StringInSlice(strings.ToLower(resource.External), []string{"yes", "external"}) {
		object.External = "yes"
	} else if common.StringInSlice(strings.ToLower(resource.External), []string{"no", "internal"}) {
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

	//The rest

	return
}

func parse(queryResults []Alias) List {
	var (
		parsed List
	)
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
		temp.Cname = []string{}
		if len(element.Cnames) != 0 {
			for _, v := range element.Cnames {
				temp.Cname = append(temp.Cname, v.Cname)
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
		if len(element.Nodes) != 0 {

			for _, v := range element.Nodes {
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
			temp.Pwned = common.StringInSlice(temp.Hostgroup, GetUsersHostgroups())
		}

		parsed.Objects = append(parsed.Objects, temp)
	}
	return parsed
}

func sanitazeInUpdate(current Alias, new Resource) Alias {

	//Cnames
	current.Cnames = []Cname{}
	if len(new.Cname) != 0 {
		split := common.DeleteEmpty(strings.Split(new.Cname[0], ","))
		for _, cname := range split {
			current.Cnames = append(current.Cnames, Cname{Cname: cname})
		}

	}

	//Alarms
	current.Alarms = []Alarm{}
	if len(new.Alarms) != 0 {
		split := common.DeleteEmpty(strings.Split(new.Alarms[0], ","))
		for _, alarm := range split {
			element := common.DeleteEmpty(strings.Split(alarm, ":"))
			current.Alarms = append(current.Alarms, Alarm{
				Name:      element[0],
				Recipient: element[1],
				Parameter: common.StringToInt(element[2])})

		}
	}

	//Nodes
	current.Nodes = []*Relation{}
	fields := [][]string{new.AllowedNodes, new.ForbiddenNodes}
	for _, field := range fields {
		if len(field) != 0 {
			nodes := common.DeleteEmpty(strings.Split(field[0], ","))
			for _, node := range nodes {
				current.Nodes = append(current.Nodes, &Relation{
					Blacklist: false,
					Node: &Node{
						NodeName:         node,
						LastModification: time.Now(),
						Hostgroup:        current.Hostgroup}})
			}
		}
	}

	if new.External != "" {
		if common.StringInSlice(new.External, []string{"yes", "external"}) {
			current.External = "yes"
		} else if common.StringInSlice(new.External, []string{"no", "internal"}) {
			current.External = "no"
		}
	}

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
