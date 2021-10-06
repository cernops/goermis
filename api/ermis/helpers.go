package ermis

/* This file contains helper functions and custom validator tags*/
import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"
)

var (
	// These variable will come from the top project
	Version = ""
	Release = ""
)

// Sets the version and release number
func SetVersion(my_version, my_release string) {
	Version = my_version
	Release = my_release
}

/*////////////Helper Functions///////////////////*/

/*When it comes to nodes/cnames/alarms, i strived for standardization
So instead of having some as string type and some as array, i decided to
keep all of them as []string type. The problem arises that the current UI
sends form-urlencoded content type, which is a string. As a result, the default
echo binder we are using binds the whole string of elements in the [0] of our []string
Since, the revamp of the UI has not been part of my project scope(where we could change how data is sent),
explode solves that issue, by splitting the element [0] when content type is form.
*/
//Turn this slice []string{"a,b,c"} to this one ==> []string{"a","b","c"}
func Explode(contentType string, slice []string) []string {
	if contentType == "application/json" {
		return slice

	} else if contentType == "application/x-www-form-urlencoded" {
		exploded := DeleteEmpty(strings.Split(slice[0], ","))
		return exploded
	} else {
		log.Error("Received an unpredictable content type, not sure how to bind array fields")
		return []string{}
	}

}

//FindNodeID returns the ID of a node. If it doesnt exists, returns 0
func FindNodeID(name string, relations []Relation) int {
	for _, n := range relations {
		if name == n.Node.NodeName {
			return n.NodeID
		}
	}

	return 0
}

//DeleteEmpty makes sure we do not have empty values in our slices
func DeleteEmpty(s []string) []string {
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

//EqualCnames compares two arrays of Cname type
func EqualCnames(cname1, cname2 []Cname) bool {
	var (
		intf ContainsIntf
	)
	if len(cname1) != len(cname2) {
		return false
	}
	for _, v := range cname1 {
		intf = v
		if !Contains(intf, cname2) {
			return false
		}
	}
	return true
}

func generateRandomSecret() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	s := make([]rune, 10)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}

	secret := base64.StdEncoding.EncodeToString([]byte(string(s)))
	return secret
}

/*//////////////Custom Validator Tags/////////////////////////*/

//customValidators adds our new tags in the govalidator
func customValidators() {
	govalidator.TagMap["hash"] = govalidator.Validator(func(str string) bool {
		len := "60"
		match, _ := regexp.MatchString("^[a-zA-Z0-9^/.$]{"+len+"}$", str)
		return match
	})

	govalidator.TagMap["nodes"] = govalidator.Validator(func(str string) bool {
		if len(str) > 0 {
			var allowed = regexp.MustCompile(`^[a-z][a-z0-9\-]*[a-z0-9]$`)
			part := strings.Split(str, ".")
			for _, p := range part {
				if !allowed.MatchString(p) || !govalidator.InRange(len(p), 2, 40) {
					log.Errorf("Not valid node name: %v", str)
					return false
				}
			}
			return true
		}

		return true
	})

	//govalidator.TagMap["cnames"] = govalidator.Iterator(func(str interface{},index int)  {
	govalidator.TagMap["cnames"] = govalidator.Validator(func(str string) bool {

		if len(str) > 0 {
			var allowed = regexp.MustCompile(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`)
			if !allowed.MatchString(str) || !govalidator.InRange(len(str), 2, 511) {
				log.Errorf("Not valid cname: %v ", str)
				return false
			}
			return true
		}

		return false
	})

	govalidator.TagMap["alarms"] = govalidator.Validator(func(str string) bool {

		if len(str) > 0 {
			alarm := strings.Split(str, ":")
			if !StringInSlice(alarm[0], []string{"minimum"}) {
				log.Error("No valid type of alarm")
				return false
			}
			if !govalidator.IsEmail(alarm[1]) {
				log.Errorf("No valid e-mail address %v in alarm %v ", alarm[1], str)
				return false

			}
			if !govalidator.IsInt(alarm[2]) {
				log.Errorf("No valid parameter value %v in alarm %v", alarm[2], str)
				return false

			}
			return true
		}

		return false
	})

	govalidator.TagMap["best_hosts"] = govalidator.Validator(func(str string) bool {
		param, err := strconv.Atoi(str)
		if err != nil {
			return false
		}
		return param >= -1

	})

	govalidator.TagMap["hostgroup"] = govalidator.Validator(func(str string) bool {

		if len(str) > 0 {
			var allowed = regexp.MustCompile(`^[a-z][a-z0-9\_\/]*[a-z0-9]$`)
			if !allowed.MatchString(str) || !govalidator.InRange(len(str), 2, 50) {
				log.Errorf("Not valid hostgroup: %v ", str)
				return false
			}
			return true
		}
		return false
	})

}

/*/////////////Unified way to return responses to user /////////////////////////////*/

//MessageToUser renders the reply for the user
func MessageToUser(c echo.Context, status int, message string, page string) error {
	username := c.Request().Header.Get("X-Forwarded-User")
	httphost := c.Request().Header.Get("X-Forwarded-Host")
	if message != "" {
		if 200 <= status && status < 300 {
			log.Infof("[%v]%v", username, message)
		} else {
			log.Errorf("[%v]%v", username, message)
		}
	}

	return c.Render(status, page, map[string]interface{}{
		"Auth":    true,
		"csrf":    c.Get("csrf"),
		"User":    username,
		"Message": message,
		"Host":    httphost,
		"Version": fmt.Sprintf("%s-%s", Version, Release),
	})
}
