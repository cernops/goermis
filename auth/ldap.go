package auth

import (
	"gitlab.cern.ch/lb-experts/goermis/bootstrap"

	"github.com/go-ldap/ldap/v3"
)

const (
	baseSuffix            = ",OU=Users,OU=Organic Units,DC=cern,DC=ch"
	filterSuffix          = ",OU=e-groups,OU=Workgroups,DC=cern,DC=ch)"
	nestedfilterPrefix    = "(memberOf:1.2.840.113556.1.4.1941:=CN="
	excludeDisabledPrefix = "(&(!(userAccountControl:1.2.840.113556.1.4.803:=2))(|"
	excludeDisabled       = true
)

var (
	log = bootstrap.GetLog()
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
func isMemberOf(username string, group string) bool {
	conn := initconnection()
	defer conn.Close()
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
		log.Error("User" + username + " is not member of ermis-lbaas-admins")
		searchReq = query(base, nestedFilter)
		result, err = conn.Search(searchReq)
		if err != nil && result == nil {
			log.Errorf("user %v is not member in any of nested e-groups", username)
			return false
		}

	}

	if len(result.Entries) == 1 && result.Entries[0].GetAttributeValue("cn") == username {

		log.Debugf("got %v search results", result.Entries[0].GetAttributeValue("cn"))
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

//CheckCud checks a user if he is member of egroup
func CheckCud(username string) bool {
	return isMemberOf(username, "ermis-lbaas-admins")
}
