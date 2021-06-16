package ermis

import (
	"errors"
	"strings"

	landbsoap "gitlab.cern.ch/lb-experts/goermis/landb"
)

/////////////////////////DNS EQUIVALENT METHODS/////////////////////////
//DNS functions are responsible for creating/updating and deleting aliases and their cnames in DNS.

//A) CREATE in DNS

//createInDNS creates 1 alias entry if the view is "internal" and
//an additional entry if the view is "external"
func (alias Alias) createInDNS() error {

	//check for existing aliases in DNS with the same name
	entries := landbsoap.Conn().DNSDelegatedSearch(strings.Split(alias.AliasName, ".")[0] + "*")
	//Double-check that DNS doesn't contain such an alias
	
	if len(entries) == 0 {
		log.Infof("[%v] preparing to add %v in DNS", alias.User, alias.AliasName)
		//If view is external, we need to create two entries in DNS
		views := make(map[string]string)
		views["internal"] = cfg.Soap.SoapKeynameI
		if alias.External == "yes" {
			views["external"] = cfg.Soap.SoapKeynameE
		}

		//Create the alias first
		for view, keyname := range views {
			//If alias creation succeeds, then we proceed with the rest
			if landbsoap.Conn().DNSDelegatedAdd(alias.AliasName, view, keyname, "Created by:"+alias.User, "goermis") {
				log.Infof("[%v] %s/%s has been created", alias.User, alias.AliasName, view)
				//If alias is created successfully and there are also cnames...
				if len(alias.Cnames) != 0 {
					if alias.createCnamesDNS(view) {
						log.Infof("[%v] cnames added in DNS for alias %v/%v ", alias.User, alias.AliasName, view)
					} else {
						return errors.New("failed to create cnames")
					}

				}

			} else {
				return errors.New("Failed to create domain " + alias.AliasName + "/" + view + " in DNS")
			}
		}
		return nil

	}
	return errors.New("Alias entry with the same name exist in DNS, skipping creation")

}

//B) DELETE

func (alias Alias) deleteFromDNS() error {

	entries := landbsoap.Conn().DNSDelegatedSearch(strings.Split(alias.AliasName, ".")[0] + "*")
	if len(entries) != 0 {
		log.Infof("[%v] preparing to delete %v from DNS", alias.User, alias.AliasName)
		var views []string
		views = append(views, "internal")
		if alias.External == "yes" {
			views = append(views, "external")
		}
		for _, view := range views {
			if landbsoap.Conn().DNSDelegatedRemove(alias.AliasName, view) {
				log.Infof("[%v] %v/%v has been deleted", alias.User, alias.AliasName, view)
			} else {
				return errors.New("Failed to delete " + alias.AliasName + "/" + view + " from DNS. Rolling Back")

			}

		}

		return nil
	}
	return nil
}

//C) UPDATE

//updateDNS updates the cname or visibility changes in DNS
func (alias Alias) updateDNS(oldObject Alias) (err error) {
	if alias.External != oldObject.External {
		if err := alias.updateView(oldObject); err != nil {
			return err
		}

	}
	//2.If there is a change in cnames, update DNS
	if !EqualCnames(alias.Cnames, oldObject.Cnames) {
		if err := alias.updateCnamesInDNS(oldObject.Cnames); err != nil {
			return err
		}
	}
	return nil
}

//////Logical sub-functions of the CREATE/DELETE/UPDATE////

//createCnamesDNS adds a list of cnames in the defined alias/view.
func (alias Alias) createCnamesDNS(view string) bool {
	for _, cname := range alias.Cnames {
		log.Infof("[%v] adding in DNS the cname %v", alias.User, cname.Cname)
		if !landbsoap.Conn().DNSDelegatedAliasAdd(alias.AliasName, view, cname.Cname) {
			return false
		}
	}

	return true
}

func (alias Alias) updateView(oldObject Alias) error {
	oview := oldObject.External
	nview := alias.External
	log.Infof("[%v] visibility has changed from %v to %v", alias.User, oview, nview)
	//Changing view from internal to external
	if nview == "yes" && oview == "no" {
		//Create the external visibility
		if landbsoap.Conn().DNSDelegatedAdd(alias.AliasName, "external", cfg.Soap.SoapKeynameE, "Created by:"+alias.User, "goermis") {
			//Make a copy the cnames from the existing internal DNS entry to the external one
			if len(oldObject.Cnames) != 0 {
				if !oldObject.createCnamesDNS("external") {
					return errors.New("failed to create cnames for the external DNS entry")
				}
			}

		}

	} else if nview == "no" && oview == "yes" {
		//If fails to delete external view...
		if !landbsoap.Conn().DNSDelegatedRemove(alias.AliasName, "external") {
			//...add again what we just deleted
			landbsoap.Conn().DNSDelegatedAdd(alias.AliasName, "external", cfg.Soap.SoapKeynameE, "Created by:"+alias.User, "goermis")
			if len(oldObject.Cnames) != 0 {
				if !oldObject.createCnamesDNS("external") {
					return errors.New("failed to create cnames for the external DNS entry")
				}
			}

			return errors.New("failed to update visibility from external to internal")
		}
	}
	return nil
}
func ContainsCname(Cname, []Cname) {}

//updateCnamesInDNS updates DNS with any possible Cnames changes.
func (alias Alias) updateCnamesInDNS(oldCnames []Cname) error {
	//If we reached this point, it means that theres been a change in cnames,
	// but not in visibility. Thats why we can use either the old or new visibility value.
	//We use the new one to minimize the number of variables.
	var (
		views []string
		intf  ContainsIntf
	)

	views = append(views, "internal")
	if alias.External == "yes" {
		views = append(views, "external")
	}

	//We iterate over the views so that we update both DNS entries if the view is external
	for _, view := range views {
		//If there are new cnames...
		if len(alias.Cnames) != 0 {
			for _, cname := range oldCnames {
				intf = cname
				//...and one of the existing cnames doesn't exist in the new list
				if !Contains(intf, alias.Cnames) {
					//we delete that cname
					if !landbsoap.Conn().DNSDelegatedAliasRemove(alias.AliasName, view, cname.Cname) {
						return errors.New("Failed to delete existing cname " +
							cname.Cname + " while updating DNS")
					}
				}
			}

			for _, cname := range alias.Cnames {
				intf = cname
				//...if a cname from the new list doesn't exist
				if !Contains(intf, oldCnames) {
					//...we add that one
					if !landbsoap.Conn().DNSDelegatedAliasAdd(alias.AliasName, view, cname.Cname) {
						return errors.New("Failed to add new cname in DNS " +
							cname.Cname + " while updating alias " + alias.AliasName)
					}
				}

			}
			//If there are no new cnames, there's been a mass purge(user deleted all cnames).
			//We clean the DNS from the old cnames
		} else {
			for _, cname := range oldCnames {
				if !landbsoap.Conn().DNSDelegatedAliasRemove(alias.AliasName, view, cname.Cname) {
					return errors.New("Failed to delete cname from DNS" +
						cname.Cname + " while purging all")
				}
			}
		}
	}
	return nil

}
