package ermis

/*This file contains two interfaces, which are implemented
by three struct types*/

/*ContainsIntf is an interface for checking if a specific object
is part of a slice with the same element types*/
type ContainsIntf interface {
	contains(interface{}) bool
}

/*PrivilegeIntf is an interface for comparing the privileges of two Relations objects*/
type PrivilegeIntf interface {
	compare(interface{}) (bool, bool)
}

func Contains(ifc ContainsIntf, r interface{}) bool {
	return ifc.contains(r)
}
func Compare(ifc PrivilegeIntf, r interface{}) (bool, bool) {
	return ifc.compare(r)
}

//Contains returns true if a cname can be found in a list of Cname objects
func (e Cname) contains(s interface{}) bool {
	var slice []Cname
	switch s := s.(type) {
	case []Cname:
		slice = s
	}
	for _, a := range slice {
		if a.Cname == e.Cname {
			return true
		}
	}
	return false

}

//ContainsAlarm checks if an alarm object is in a slice of objects
func (a Alarm) contains(s interface{}) bool {
	var slice []Alarm
	switch s := s.(type) {
	case []Alarm:
		slice = s
	}
	for _, alarm := range slice {
		if alarm.Name == a.Name &&
			alarm.Recipient == a.Recipient &&
			alarm.Parameter == a.Parameter {
			return true
		}
	}
	return false

}

func (r Relation) compare(a interface{}) (bool, bool) {
	var (
		slice []Relation
	)
	switch a := a.(type) {
	case []Relation:
		slice = a
	}
	for _, v := range slice {
		if v.Node.NodeName == r.Node.NodeName {
			if v.Blacklist == r.Blacklist {

				return true, true
			}
			return true, false
		}

	}

	return false, false

}

