package api

import (
	"strconv"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"gitlab.cern.ch/lb-experts/goermis/aiermis/common"
)

type (
	//Resource deals with the output from the queries
	Resource struct {
		ID               int       `form:"alias_id"          json:"alias_id"          valid:"-"`
		AliasName        string    `form:"alias_name"        json:"alias_name"        valid:"required,dns"`
		Behaviour        string    `form:"behaviour"         json:"behaviour"         valid:"-"`
		BestHosts        int       `form:"best_hosts"        json:"best_hosts"        valid:"required,int,best_hosts"`
		Clusters         string    `form:"clusters"          json:"clusters"          valid:"alphanum"`
		ForbiddenNodes   []string  `form:"ForbiddenNodes"    json:"ForbiddenNodes"    valid:"optional,nodes"   gorm:"not null"`
		AllowedNodes     []string  `form:"AllowedNodes"      json:"AllowedNodes"      valid:"optional,nodes"   gorm:"not null"`
		Cname            []string  `form:"cnames"            json:"cnames"            valid:"optional,cnames"  gorm:"not null"`
		External         string    `form:"external"          json:"external"          valid:"required,in(yes|no)"`
		Hostgroup        string    `form:"hostgroup"         json:"hostgroup"         valid:"required,hostgroup"`
		LastModification time.Time `form:"last_modification" json:"last_modification" valid:"-"`
		Metric           string    `form:"metric"            json:"metric"            valid:"in(cmsfrontier|minino|minimum|),optional"`
		PollingInterval  int       `form:"polling_interval"  json:"polling_interval"  valid:"numeric"`
		Tenant           string    `form:"tenant"            json:"tenant"            valid:"optional,alphanum"`
		TTL              int       `form:"ttl"               json:"ttl,omitempty"     valid:"numeric,optional"`
		User             string    `form:"user"              json:"user"              valid:"optional,alphanum"`
		Statistics       string    `form:"statistics"        json:"statistics"        valid:"alpha"`
		ResourceURI      string    `                         json:"resource_uri"      valid:"-"`
		Pwned            bool      `                         json:"pwned"             valid:"-"`
		Alarms           []string  `form:"alarms"              json:"alarms"             valid:"-"`
	}
)

func translate(resource Resource) (object Alias) {
	var (
		cnames []Cname
	)
	//Cnames
	spew.Dump(resource)
	if len(resource.Cname) != 0 {

		split := common.DeleteEmpty(strings.Split(resource.Cname[0], ","))
		for _, cname := range split {
			cnames = append(cnames, Cname{Cname: cname})
		}
	}

	object.AliasName = resource.AliasName
	object.Behaviour = resource.Behaviour
	object.BestHosts = resource.BestHosts
	object.Clusters = resource.Clusters
	object.Hostgroup = resource.Hostgroup
	object.External = resource.External
	object.LastModification = resource.LastModification
	object.Metric = resource.Metric
	object.PollingInterval = resource.PollingInterval
	object.TTL = resource.TTL
	object.Tenant = resource.Tenant
	object.User = resource.User
	object.Statistics = resource.Statistics
	object.Cnames = cnames

	return
}

//defaultAndHydrate prepares the object with default values and domain before CREATE
func (r *Resource) defaultAndHydrate() {
	//Populate the struct with the default values
	r.Behaviour = "mindless"
	r.Metric = "cmsfrontier"
	r.PollingInterval = 300
	r.Statistics = "long"
	r.Clusters = "none"
	r.Tenant = "golang"
	r.TTL = 60
	r.LastModification = time.Now()
	//Hydrate
	if !strings.HasSuffix(r.AliasName, ".cern.ch") {
		r.AliasName = r.AliasName + ".cern.ch"
	}
	if common.StringInSlice(strings.ToLower(r.External), []string{"yes", "external"}) {
		r.External = "yes"
	}
	if common.StringInSlice(strings.ToLower(r.External), []string{"no", "internal"}) {
		r.External = "no"
	}
}

func parse(queryResults []Alias) (parsed Objects) {
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
		if len(element.Cnames) != 0 {
			for _, v := range element.Cnames {
				temp.Cname = append(temp.Cname, v.Cname)
			}

		}

		//The alarms
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

		parsed = append(Objects.L , temp)
	}
	return Objects
}
