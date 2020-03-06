package models

import (
	"net/url"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	schema "github.com/gorilla/Schema"
	"github.com/labstack/gommon/log"
	"gitlab.cern.ch/lb-experts/goermis/api/common"
)

//Alias structure is a model for describing the alias
type Alias struct {
	ID               uint      `json:"alias_id" schema:"alias_id" gorm:"unique;not null;auto_increment;primary_key"`
	AliasName        string    `json:"alias_name" schema:"alias_name" gorm:"not null;size:40;unique"`
	Behaviour        string    `json:"behaviour" schema:"behaviour" gorm:"not null"`
	BestHosts        int       `json:"best_hosts" schema:"best_hosts" gorm:"not null"`
	External         string    `json:"external" schema:"external" gorm:"not null" `
	Metric           string    `json:"metric" schema:"metric" gorm:"not null"`
	PollingInterval  int       `json:"polling_interval" schema:"polling_interval" gorm:"not null" `
	Statistics       string    `json:"statistics" schema:"statistics" gorm:"not null"`
	Clusters         string    `json:"clusters" schema:"clusters" gorm:"not null" `
	Tenant           string    `json:"tenant" schema:"tenant"  gorm:"not null"`
	Hostgroup        string    `json:"hostgroup" schema:"hostgroup" gorm:"not null" `
	User             string    `json:"user" schema:"user" gorm:"not null" `
	TTL              int       `json:"ttl" schema:"ttl"`
	LastModification time.Time `json:"last_modification" schema:"last_modification" `
	Relations        []*Relation
	Cnames           []Cname `json:"cnames"  gorm:"foreignkey:AliasID"`
}

//Relation testing
type Relation struct {
	ID        uint
	Node      *Node
	NodeID    int
	Alias     *Alias
	AliasID   int
	Blacklist bool
}

//Cname structure is a model for the cname description
type Cname struct {
	ID      int    `json:"cname_id" gorm:"auto_increment;primary_key"`
	AliasID uint   `json:"alias_id" gorm:"not null"`
	CName   string `json:"cname" gorm:"not null;unique" `
}

//Node structure defines the model for the nodes params
type Node struct {
	ID               uint        `json:"node_id" schema:"node_id" gorm:"unique;not null;auto_increment;primary_key"`
	NodeName         string      `json:"node_name" schema:"node_name" gorm:"not null;size:60;unique"`
	LastModification time.Time   `json:"last_modification" schema:"last_modification"`
	Load             int         `json:"load" schema:"load"`
	State            string      `json:"state" schema:"state" gorm:"not null"`
	Hostgroup        string      `json:"hostgroup" schema:"hostgroup" gorm:"size:40;not null"`
	Relations        []*Relation `json:"relation" schema:"relation"`
}

var (
	decoder = schema.NewDecoder()
)

//CreateObject creates an alias
func (a Alias) CreateObject(params url.Values) (err error) {
	spew.Dump(params)
	decoder.IgnoreUnknownKeys(true)
	err = decoder.Decode(&a, params)
	if err != nil {
		log.Error(err)
		//panic(err)

	}

	con.Create(&a)

	cnames := common.DeleteEmpty(strings.Split(params.Get("cnames"), ","))
	err = a.UpdateCnames(cnames)
	if err != nil {
		return err
	}

	return err
}

//Prepare prepares alias before creation with def values
func (a Alias) Prepare() {
	//Populate the struct with the default values
	a.User = "kkouros"
	a.Behaviour = "mindless"
	a.Metric = "cmsfrontier"
	a.PollingInterval = 300
	a.Statistics = "long"
	a.Clusters = "none"
	a.Tenant = "golang"
	a.LastModification = time.Now()
}

//DeleteObject deletes an alias and its Relations
/////WIP
func (a Alias) DeleteObject(aliasName string) (err error) {
	a.clearCnames()
	if err != nil {
		log.Error("Something went wrong when deleting the Cnames")
		return err
	}

	//con.Model(&Alias).Where("alias_name = ?", alias).Preload("Cnames").Delete(&Alias.Cnames)
	err = con.Where("alias_name = ?", aliasName).Delete(&Alias{}).Error
	if err != nil {
		return err
	}

	return err
}

//ModifyObject modifies aliases and its associations
func (a Alias) ModifyObject(params url.Values) (err error) {
	//Get the values that we need
	aliasName := params.Get("alias_name")
	external := params.Get("external")
	hostgroup := params.Get("hostgroup")

	//Prepare cnames separately
	cnames := common.DeleteEmpty(strings.Split(params.Get("cnames"), ","))

	//Update external and hostgroup fields
	con.Model(&a).Where("alias_name = ?", aliasName).Take(&a).UpdateColumns(
		map[string]interface{}{
			"external":  external,
			"hostgroup": hostgroup,
		},
	)

	err = a.UpdateCnames(cnames)
	if err != nil {
		return err
	}
	return nil
}

//UpdateCnames updates cnames
func (a Alias) UpdateCnames(cnames []string) (err error) {

	// If there are no cnames from UI , delete them all, otherwise append them
	if len(cnames) > 0 {
		for _, cname := range cnames {
			spew.Dump(cname)
			con.Model(&a).Association("Cnames").Append(&Cname{CName: cname})
			return err
		}
	} else {
		con.Where("alias_id = ?", a.ID).Delete(&Cname{})
		return err

	}
	return nil
}

func (a Alias) clearCnames() (err error) {

	if con.Where("alias_name = ?", a.AliasName).First(&a).RecordNotFound() {
		log.Error("Alias not found")
		return err

	}
	err = con.Where("alias_id= ?", a.ID).Delete(&Cname{}).Error
	if err != nil {
		log.Error("Cname deletion failed")
		return err
	}
	return nil

}
