package auth

import (
	"fmt"
	"log"

	"github.com/go-ldap/ldap"
)

const (
	baseSuffix            = ",OU=Users,OU=Organic Units,DC=cern,DC=ch"
	filterSuffix          = ",OU=e-groups,OU=Workgroups,DC=cern,DC=ch)"
	nestedfilterPrefix    = "(memberOf:1.2.840.113556.1.4.1941:=CN="
	excludeDisabledPrefix = "(&(!(userAccountControl:1.2.840.113556.1.4.803:=2))(|"
	excludeDisabled       = true
)

func initconnection() *ldap.Conn {
	l, err := ldap.DialURL("ldap://xldap.cern.ch")
	if err != nil {
		log.Fatal(err)
		return nil
	}
	return l
}

//IsMemberOf checks if a user is subscribed to a certain egroup
func IsMemberOf(username string, group string) bool {
	conn := initconnection()
	defer conn.Close()
	fmt.Println(group)
	base := "CN=" + username + baseSuffix
	filter := "(memberOf=CN=" + group + filterSuffix
	nestedFilter := nestedfilterPrefix + group + filterSuffix

	if excludeDisabled == true {
		filter = excludeDisabledPrefix + filter + "))"
		nestedFilter = excludeDisabledPrefix + nestedFilter + "))"

	}

	// Filters must start and finish with ()!
	searchReq := query(base, filter)
	result, err := conn.Search(searchReq)
	if err != nil && result == nil {
		fmt.Printf("No results in the egroup.LDAP response: %s", err)
		searchReq = query(base, nestedFilter)
		result, err = conn.Search(searchReq)
		if err != nil {
			fmt.Printf("No results in the nested egroups.LDAP response : %s", err)
			return false
		}

	}

	if len(result.Entries) == 1 && result.Entries[0].GetAttributeValue("cn") == username {

		log.Println("Got", result.Entries[0].GetAttributeValue("cn"), "search results")
		return true

	}
	return false
}
func query(base string, filter string) *ldap.SearchRequest {
	q := ldap.NewSearchRequest(
		base,
		ldap.ScopeWholeSubtree,
		0,
		1, //sizeLimit
		0,
		false,
		filter,
		[]string{"cn"}, //attrs
		[]ldap.Control{})
	return q
}
