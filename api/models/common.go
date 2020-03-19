package models

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
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

func prepareRelation(nodeID int, aliasID int, p bool) (r *AliasesNodes) {
	r = &AliasesNodes{AliasID: aliasID, NodeID: nodeID, Blacklist: p}
	return r
}

//CustomValidators adds our new tags in the govalidator
func CustomValidators() (err error) {
	//Alias validation with domain allowance
	govalidator.TagMap["alias"] = govalidator.Validator(func(str string) bool {
		if len(str) > 255 {
			return false
		}
		part := strings.Split(str, ".")
		var allowed = regexp.MustCompile(`(?!-)[A-Z\\d-]{1,63}(?<!-)$`)
		for _, p := range part {
			if !allowed.MatchString(p) || !govalidator.InRange(len(p), 2, 40) {
				return false
			}
		}

		return true

	})

	// Metric validation
	govalidator.TagMap["metric"] = govalidator.Validator(func(str string) bool {

		allowed := []string{"minino", "minimum", "cmsfrontier"}

		return stringInSlice(str, allowed)
	})

	govalidator.TagMap["best_hosts"] = govalidator.Validator(func(str string) bool {
		return true
	})

	return nil
}
