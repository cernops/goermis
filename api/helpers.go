package api

/* This file contains helper functions and custom validator tags*/
import (
	"regexp"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"
)

/*////////////Helper Functions///////////////////*/
func containsCname(s []Cname, e string) bool {
	for _, a := range s {
		if a.Cname == e {
			return true
		}
	}
	return false

}

/*When it comes to nodes/cnames/alarms, i strived for standardization
So instead of having some as string type and some as array, i decided to
keep all of them as []string type. The problem arises that the current UI
sends form-urlencoded content type, which is a string. As a result, the default
echo binder we are using binds the whole string of elements in the [0] of our []string
Since, the revamp of the UI has not been part of my project scope(where we could change how data is sent),
explode solves that issue, by splitting the element [0] when content type is form.
*/
func explode(contentType string, slice []string) []string {
	if contentType == "application/json" {
		return slice

	} else if contentType == "application/x-www-form-urlencoded" {
		exploded := deleteEmpty(strings.Split(slice[0], ","))
		return exploded
	} else {
		log.Error("Received an unpredictable content type, not sure how to bind array fields")
		return []string{}
	}

}

//containsAlarm checks if an alarm object is in a slice of objects
func containsAlarm(s []Alarm, a Alarm) bool {
	for _, alarm := range s {
		if alarm.Name == a.Name &&
			alarm.Recipient == a.Recipient &&
			alarm.Parameter == a.Parameter {
			return true
		}
	}
	return false

}

//containsNode checks if a node has a relation with an alias
// and the status of that relation(allowed or forbidden)
func containsNode(a []*Relation, b *Relation) (bool, bool) {
	for _, v := range a {
		if v.Node.NodeName == b.Node.NodeName {
			if v.Blacklist == b.Blacklist {
				return true, true
			}
			return true, false
		}

	}
	return false, false

}

//find returns the ID of a node. If it doesnt exists, returns 0
func findNodeID(name string) int {
	var node Node
	con.Select("id").Where("node_name=?", name).Find(&node)
	return node.ID
}

//find returns the ID of a node. If it doesnt exists, returns 0
func findAliasID(name string) int {
	var alias Alias
	con.Select("id").Where("alias_name=?", name).Find(&alias)
	return alias.ID
}

//deleteEmpty makes sure we do not have empty values in our slices
func deleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

//stringInSlice checks if a string is in a slice
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

//stringToInt converts a string to a int. It is used to hide error checks
func stringToInt(s string) (i int) {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Error("Error while converting string to int")
	}
	return i
}

func equal(cname1, cname2 []Cname) bool {
	if len(cname1) != len(cname2) {
		return false
	}
	for _, v := range cname1 {
		if !containsCname(cname2, v.Cname) {
			return false
		}
	}
	return true
}

/*//////////////Custom Validator Tags/////////////////////////*/

//customValidators adds our new tags in the govalidator
func customValidators() {
	govalidator.TagMap["nodes"] = govalidator.Validator(func(str string) bool {
		if len(str) > 0 {
			var allowed = regexp.MustCompile(`^[a-z][a-z0-9\-]*[a-z0-9]$`)
			part := strings.Split(str, ".")
			for _, p := range part {
				if !allowed.MatchString(p) || !govalidator.InRange(len(p), 2, 40) {
					log.Error("Not valid node name: " + str)
					return false
				}
			}
		}

		return true
	})

	//govalidator.TagMap["cnames"] = govalidator.Iterator(func(str interface{},index int)  {
	govalidator.TagMap["cnames"] = govalidator.Validator(func(str string) bool {

		if len(str) > 0 {
			var allowed = regexp.MustCompile(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`)
			if !allowed.MatchString(str) || !govalidator.InRange(len(str), 2, 511) {
				log.Error("Not valid cname: " + str)
				return false
			}
		}

		return true
	})

	govalidator.TagMap["alarms"] = govalidator.Validator(func(str string) bool {

		if len(str) > 0 {
			alarm := strings.Split(str, ":")
			if !stringInSlice(alarm[0], []string{"minimum"}) {
				log.Error("No valid type of alarm")
				return false
			}
			if !govalidator.IsEmail(alarm[1]) {
				log.Error("No valid e-mail address " + alarm[1] + " in alarm " + str)
				return false

			}
			if !govalidator.IsInt(alarm[2]) {
				log.Error("No valid parameter value " + alarm[2] + " in alarm " + str)
				return false

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

/*/////////////Unified way to return responses to user/////////////////////////////*/

//MessageToUser renders the reply for the user
func MessageToUser(c echo.Context, status int, message string, page string) error {
	username := c.Request().Header.Get("X-Forwarded-User")
	httphost := c.Request().Header.Get("X-Forwarded-Host")
	if message != "" {
		if 200 <= status && status < 300 {
			log.Info("[" + username + "]" + message)
		} else {
			log.Error("[" + username + "]" + message)
		}
	}

	return c.Render(status, page, map[string]interface{}{
		"Auth":    true,
		"csrf":    c.Get("csrf"),
		"User":    username,
		"Message": message,
		"Host":    httphost,
	})
}
