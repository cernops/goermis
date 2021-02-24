package landbsoap

import (
	"fmt"
	"testing"

	bootstrap "gitlab.cern.ch/lb-experts/goermis/bootstrap"
	landbsoap "gitlab.cern.ch/lb-experts/goermis/landb"
)

var (
	aliasName       = "goermis-ci-test.cern.ch"
	search          = "goermis-ci-test*"
	keynameInternal = "ITPES-INTERNAL"
	viewInternal    = "internal"
	keynameExternal = "ITPES-EXTERNAL"
	viewExternal    = "external"
)

//TestCreateAliasInternal tests if an alias entry, with internal view, can be created correctly in LanDB
func TestCreateAliasInternal(t *testing.T) {

	bootstrap.ParseFlags()

	if !landbsoap.Conn().DNSDelegatedAdd(aliasName, viewInternal, keynameInternal, "Created by: ci_test", "goermis") {
		t.Errorf("Error creating the alias %v", aliasName)
	}
	fmt.Printf("And if we create it again, it should fail")

	if landbsoap.Conn().DNSDelegatedAdd(aliasName, viewInternal, keynameInternal, "Created by: ci_test", "goermis") {
		t.Errorf("Re-creating the alias %v did not fail", aliasName)
	}

	fmt.Printf("We check it was actually created as we expect")
	entries := landbsoap.Conn().DNSDelegatedSearch(search)

	if len(entries) == 0 {
		t.Errorf("Entry for alias %v could not be found", aliasName)

	} else if len(entries) > 1 {
		t.Errorf("We found more than 1 entry for %v", aliasName)
		for v := range entries {
			t.Errorf("Entry value: %v\n", v)
		}
	}

}

//TestDeleteAliasInternal tests if an alias entry, with internal view, can be deleted correctly from LanDB
func TestDeleteAliasInternal(t *testing.T) {
	bootstrap.ParseFlags()

	if !landbsoap.Conn().DNSDelegatedRemove(aliasName, viewInternal) {
		t.Errorf("Error deleting the alias %v", aliasName)
	}
	fmt.Printf("And deleting it again should fail")
	if landbsoap.Conn().DNSDelegatedRemove(aliasName, viewInternal) {
		t.Errorf("A deleted alias can be deleted again %v", aliasName)
	}

	fmt.Printf("We check it was actually deleted it")
	entries := landbsoap.Conn().DNSDelegatedSearch(search)
	if len(entries) != 0 {
		t.Errorf("Alias %v was not deleted", aliasName)
		for v := range entries {
			t.Errorf("Entry value we found: %v\n", v)
		}

	}

}

//TestCreateAliasExternal tests if an alias entry, with external view, can be created correctly in LanDB
func TestCreateAliasExternal(t *testing.T) {

	bootstrap.ParseFlags()

	//Create the internal view
	if !landbsoap.Conn().DNSDelegatedAdd(aliasName, viewInternal, keynameInternal, "Created by: ci_test", "goermis") {
		t.Errorf("Error creating the alias %v", aliasName)
	}
	fmt.Printf("And if we create it again, it should fail")

	if landbsoap.Conn().DNSDelegatedAdd(aliasName, viewInternal, keynameInternal, "Created by: ci_test", "goermis") {
		t.Errorf("Re-creating the alias %v did not fail", aliasName)
	}

	//Create the external view
	if !landbsoap.Conn().DNSDelegatedAdd(aliasName, viewExternal, keynameExternal, "Created by: ci_test", "goermis") {
		t.Errorf("Error creating the alias %v", aliasName)
	}
	fmt.Printf("And if we create it again, it should fail")

	if landbsoap.Conn().DNSDelegatedAdd(aliasName, viewExternal, keynameExternal, "Created by: ci_test", "goermis") {
		t.Errorf("Re-creating the alias %v did not fail", aliasName)
	}

	//Search the entries
	fmt.Printf("We check it was actually created as we expect")
	entries := landbsoap.Conn().DNSDelegatedSearch(search)

	if len(entries) == 0 {
		t.Errorf("No entries for alias %v could not be found", aliasName)

	} else if len(entries) == 1 {
		t.Errorf("We found only one entry for %v", aliasName)
		for v := range entries {
			t.Errorf("Entry value: %v\n", v)
		}
	} else if len(entries) > 2 {
		t.Errorf("We found more two entries for alias %v !", aliasName)
		for v := range entries {
			t.Errorf("Entry value: %v\n", v)
		}
	}

}

//TestDeleteAliasExternal tests if an alias entry, with external view, can be deleted correctly from LanDB
func TestDeleteAliasExternal(t *testing.T) {
	bootstrap.ParseFlags()

	//Delete the Internal view
	if !landbsoap.Conn().DNSDelegatedRemove(aliasName, viewInternal) {
		t.Errorf("Error deleting the alias %v", aliasName)
	}
	fmt.Printf("And deleting it again should fail")
	if landbsoap.Conn().DNSDelegatedRemove(aliasName, viewInternal) {
		t.Errorf("A deleted alias can be deleted again %v", aliasName)
	}

	//Delete the external view
	if !landbsoap.Conn().DNSDelegatedRemove(aliasName, viewExternal) {
		t.Errorf("Error deleting the alias %v", aliasName)
	}
	fmt.Printf("And deleting it again should fail")
	if landbsoap.Conn().DNSDelegatedRemove(aliasName, viewExternal) {
		t.Errorf("A deleted alias can be deleted again %v", aliasName)
	}

	//Search them
	fmt.Printf("We check it was actually deleted it")
	entries := landbsoap.Conn().DNSDelegatedSearch(search)
	if len(entries) != 0 {
		t.Errorf("Alias %v was not deleted", aliasName)
		for v := range entries {
			t.Errorf("Entry value we found: %v\n", v)
		}

	}

}
