package models

import (
	"net/url"
	"strconv"
	"strings"

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

func nodesToMap(p url.Values) map[string]int {

	temp := make(map[string]int)

	modes := map[string]int{
		"AllowedNodes":   0,
		"ForbiddenNodes": 1,
	}
	for k, v := range modes {
		if p.Get(k) != "" {
			for _, val := range DeleteEmpty(strings.Split(p.Get(k), ",")) {
				temp[val] = v
			}
		}
	}

	return temp
}
