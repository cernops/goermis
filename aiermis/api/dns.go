package api

/*
/////////////////////////DNS EQUIVALENT METHODS/////////////////////////
//DNS functions are responsible for creating/updating and deleting aliases and their cnames in DNS.

//A) CREATE in DNS

//createInDNS creates 1 alias entry if the view is "internal" and
//an additional entry if the view is "external"
func (r Resource) createInDNS() error {

	//check for existing aliases in DNS with the same name
	entries := landbsoap.Conn().DNSDelegatedSearch(strings.Split(r.AliasName, ".")[0] + "*")

	//Double-check that DNS doesn't contain such an alias
	if len(entries) == 0 {
		log.Info("[" + r.User + "]" + "Preparing to add " + r.AliasName + " in DNS")
		//If view is external, we need to create two entries in DNS
		views := make(map[string]string)
		views["internal"] = cfg.Soap.SoapKeynameI
		if r.External == "yes" {
			views["external"] = cfg.Soap.SoapKeynameE
		}

		//Create the alias first
		for view, keyname := range views {
			//If alias creation succeeds, then we proceed with the rest
			if landbsoap.Conn().DNSDelegatedAdd(r.AliasName, view, keyname, "Created by:"+r.User, "goermis") {
				log.Info("[" + r.User + "]" + r.AliasName + "/" + view + "has been created")
				//If alias is created successfully and there are also cnames...
				if len(r.Cname) != 0 {
					if r.createCnamesDNS(view) {
						log.Info("[" + r.User + "]" + "Cnames added in DNS for alias " + r.AliasName + "/" + view)
					} else {
						return errors.New("Failed to create cnames")
					}

				}

			} else {
				return errors.New("Failed to create domain " + r.AliasName + "/" + view + " in DNS")
			}
		}
		return nil

	}
	return errors.New("Alias entry with the same name exist in DNS, skipping creation")

}

//B) DELETE

func (r Resource) deleteFromDNS() error {

	entries := landbsoap.Conn().DNSDelegatedSearch(strings.Split(r.AliasName, ".")[0] + "*")
	if len(entries) != 0 {
		log.Info("[" + r.User + "]" + "Preparing to delete " + r.AliasName + " from DNS")
		var views []string
		views = append(views, "internal")
		if r.External == "yes" {
			views = append(views, "external")
		}
		for _, view := range views {
			if landbsoap.Conn().DNSDelegatedRemove(r.AliasName, view) {
				log.Info("[" + r.User + "]" + r.AliasName + "/" + view + " has been deleted ")
			} else {
				return errors.New("Failed to delete " + r.AliasName + "/" + view + " from DNS. Rolling Back")

			}

		}

		return nil
	}
	return errors.New("The requested alias for deletion doesn't exist in DNS.Skipping deletion there")
}

//C) UPDATE

//UpdateDNS updates the cname or visibility changes in DNS
func (r Resource) UpdateDNS(oldObject Resource) (err error) {
	//1.If view has changed delete and recreate alias and cnames
	if r.External != oldObject.External {
		if err := r.updateView(oldObject); err != nil {
			return err
		}

	}
	//2.If there is a change in cnames, update DNS
	if !common.Equal(r.Cname, oldObject.Cname) {
		if err := r.updateCnamesInDNS(oldObject.Cname); err != nil {
			return err
		}
	}
	return nil
}

//////Logical sub-functions of the CREATE/DELETE/UPDATE////

//createCnamesDNS adds a list of cnames in the defined alias/view.
func (r Resource) createCnamesDNS(view string) bool {
	cnames := common.DeleteEmpty(strings.Split(r.Cname, ","))
	for _, cname := range cnames {
		log.Info("[" + r.User + "]" + "Adding in DNS the cname " + cname)
		if !landbsoap.Conn().DNSDelegatedAliasAdd(r.AliasName, view, cname) {
			return false
		}
	}

	return true
}

func (r Resource) updateView(oldObject Resource) error {
	oview := oldObject.External
	nview := r.External
	alias := oldObject.AliasName
	log.Info("[" + r.User + "]" + "Visibility has changed from " + oview + " to " + nview)
	//Changing view from internal to external
	if nview == "yes" && oview == "no" {
		//Create the external visibility
		if landbsoap.Conn().DNSDelegatedAdd(alias, "external", cfg.Soap.SoapKeynameE, "Created by:"+r.User, "goermis") {
			//Make a copy the cnames from the existing internal DNS entry to the external one
			if oldObject.Cname != "" {
				if !oldObject.createCnamesDNS("external") {
					return errors.New("Failed to create cnames for the external DNS entry")
				}
			}

		}

	} else if nview == "no" && oview == "yes" {
		//If fails to delete external view...
		if !landbsoap.Conn().DNSDelegatedRemove(alias, "external") {
			//...add again what we just deleted
			landbsoap.Conn().DNSDelegatedAdd(alias, "external", cfg.Soap.SoapKeynameE, "Created by:"+r.User, "goermis")
			if oldObject.Cname != "" {
				if !oldObject.createCnamesDNS("external") {
					return errors.New("Failed to create cnames for the external DNS entry")
				}
			}

			return errors.New("Failed to update visibility from external to internal")
		}
	}
	return nil
}

//updateCnamesInDNS updates DNS with any possible Cnames changes.
func (r Resource) updateCnamesInDNS(oldCnames string) error {
	//Let's convert cnames from string to []string, because we need to iterate over them
	newCnames := common.DeleteEmpty(strings.Split(r.Cname, ","))
	existingCnames := common.DeleteEmpty(strings.Split(oldCnames, ","))

	//If we reached this point, it means that theres been a change in cnames,
	// but not in visibility. Thats why we can use the old or new visibility value.
	//We use the new one to minimize the number of variables.
	var views []string
	views = append(views, "internal")
	if r.External == "yes" {
		views = append(views, "external")
	}

	//We iterate over the views so that we update both DNS entries if the view is external
	for _, view := range views {
		//If there are new cnames...
		if len(newCnames) != 0 {
			for _, existingCname := range existingCnames {
				//...and one of the existing cnames doesn't exist in the new list
				if !common.StringInSlice(existingCname, newCnames) {
					//we delete that cname
					if !landbsoap.Conn().DNSDelegatedAliasRemove(r.AliasName, view, existingCname) {
						return errors.New("Failed to delete existing cname " +
							existingCname + " while updating DNS")
					}
				}
			}

			for _, newCname := range newCnames {
				//a precautionary check to avoid bad requests
				if newCname == "" {
					continue
				}
				//...if a cname from the new list doesn't exist
				if !common.StringInSlice(newCname, existingCnames) {
					//...we add that one
					if !landbsoap.Conn().DNSDelegatedAliasAdd(r.AliasName, view, newCname) {
						return errors.New("Failed to add new cname in DNS " +
							newCname + " while updating alias " + r.AliasName)
					}
				}

			}
			//If there are no new cnames, there's been a mass purge(user deleted all cnames).
			//We clean the DNS from the old cnames
		} else {
			for _, cname := range existingCnames {
				if !landbsoap.Conn().DNSDelegatedAliasRemove(r.AliasName, view, cname) {
					return errors.New("Failed to delete cname from DNS" +
						cname + " while purging all")
				}
			}
		}
	}
	return nil

}
*/
