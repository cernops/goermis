package models

import (
	"strconv"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/labstack/gommon/log"
)

//DeleteEmpty filters an array for empty string values
func DeleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

//getExistingCnames extracts the names of cnames for a certain alias
func getExistingCnames(a Alias) (s []string) {

	for _, value := range a.Cnames {
		s = append(s, value.CName)
	}
	return s
}

func getExistingNodesToMap(a Alias) (temp map[string]bool) {
	temp = make(map[string]bool)
	relation := []Relation{}
	con.Where("alias_id=?", a.ID).Find(&relation)
	if len(relation) > 0 {
		for _, v := range relation {
			if v.Blacklist {
				temp[nodename(v.NodeID)] = true
			} else {
				temp[nodename(v.NodeID)] = false
			}

		}
	}
	spew.Dump(temp)
	return temp
}

func nodename(id int) (name string) {
	n := new(Node)
	con.Where("id=?", id).First(n)
	return n.NodeName

}

func stringToInt(s string) (i int) {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Error("Error while converting string to int")
	}
	return i
}

func nodesToMap(p Resource) map[string]bool {

	temp := make(map[string]bool)

	modes := map[string]bool{
		p.AllowedNodes:   false,
		p.ForbiddenNodes: true,
	}
	for k, v := range modes {
		if k != "" {
			for _, val := range DeleteEmpty(strings.Split(k, ",")) {
				temp[val] = v
			}
		}
	}

	return temp
}

//PrepareNode creates a Node object
func PrepareNode(a Alias, name string) (n Node) {

	return Node{
		NodeName:         name,
		LastModification: time.Now(),
		Hostgroup:        a.Hostgroup,
	}

}
func prepareRelation(nodeID int, aliasID int, p bool) (r *Relation) {
	r = &Relation{AliasID: aliasID, NodeID: nodeID, Blacklist: p}
	return r
}
