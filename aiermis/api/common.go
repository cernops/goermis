package api

import (
	"bytes"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"gitlab.cern.ch/lb-experts/goermis/aiermis/orm"
)

func deleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

//StringInSlice checks if a string is in a slice
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

//GetExistingCnames extracts the names of cnames for a certain alias
func getExistingCnames(a orm.Alias) (s []string) {

	for _, value := range a.Cnames {
		s = append(s, value.CName)
	}
	return s
}

//StringToInt converts a string to a int. It is used to hide error checks
func stringToInt(s string) (i int) {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Error("Error while converting string to int")
	}
	return i
}

//StoreBodyOfRequest is used to store the body of a request in case its needed
func storeBodyOfRequest(c echo.Context) []byte {
	// Read the content since we need it more than once
	var bodyOfRequest []byte
	if c.Request().Body != nil {
		bodyOfRequest, _ = ioutil.ReadAll(c.Request().Body)
	}
	// Restore the io.ReadCloser to its original state
	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(bodyOfRequest)) // Use the content
	return bodyOfRequest
}

//NodesInMap puts the nodes in a map. The value is their privilege
func nodesInMap(AllowedNodes string, ForbiddenNodes string) map[string]bool {

	temp := make(map[string]bool)

	modes := map[string]bool{
		AllowedNodes:   false,
		ForbiddenNodes: true,
	}
	for k, v := range modes {
		if k != "" {
			for _, val := range deleteEmpty(strings.Split(k, ",")) {
				temp[val] = v
			}
		}
	}

	return temp
}

//CustomValidators adds our new tags in the govalidator
func customValidators() {
	govalidator.TagMap["nodes"] = govalidator.Validator(func(str string) bool {
		if len(str) > 0 {
			split := strings.Split(str, ",")
			var allowed = regexp.MustCompile(`^[a-z][a-z0-9\-]*[a-z0-9]$`)

			for _, s := range split {
				part := strings.Split(s, ".")
				for _, p := range part {
					if !allowed.MatchString(p) || !govalidator.InRange(len(p), 2, 40) {
						log.Error("Not valid node name: " + s)
						return false
					}
				}
			}
		}
		return true
	})

	govalidator.TagMap["cnames"] = govalidator.Validator(func(str string) bool {

		if len(str) > 0 {
			split := strings.Split(str, ",")
			var allowed = regexp.MustCompile(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`)

			for _, s := range split {
				if !allowed.MatchString(s) || !govalidator.InRange(len(s), 2, 511) {
					log.Error("Not valid cname: " + s)
					return false
				}
			}
		}
		return true
	})

	govalidator.TagMap["best_hosts"] = govalidator.Validator(func(str string) bool {
		return stringToInt(str) >= -1

	})

	govalidator.TagMap["hostgroup"] = govalidator.Validator(func(str string) bool {

		if len(str) > 0 {
			var allowed = regexp.MustCompile(`^[a-z][a-z0-9\_\/]*[a-z0-9]$`)
			if !allowed.MatchString(str) || !govalidator.InRange(len(str), 2, 50) {
				log.Error("Not valid hostgroup: " + str)
				return false
			}
			return true
		}
		return false
	})

}
