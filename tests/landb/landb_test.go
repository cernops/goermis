package landbsoap

import (
	"fmt"
	"testing"

	bootstrap "gitlab.cern.ch/lb-experts/goermis/bootstrap"
	landbsoap "gitlab.cern.ch/lb-experts/goermis/landb"
)

var aliasName = "goermis-ci-test.cern.ch"
var keyname = "ITPES-INTERNAL"
var view = "internal"

// TestBlank : place holder for tests
func TestCreateAlias(t *testing.T) {

	bootstrap.ParseFlags()

	if !landbsoap.Conn().DNSDelegatedAdd(aliasName, view, keyname, "Created by: ci_test", "goermis") {
		t.Errorf("Error creating the alias %v", aliasName)
	}
	fmt.Printf("And if we create it again, it should fail")

	if landbsoap.Conn().DNSDelegatedAdd(aliasName, view, keyname, "Created by: ci_test", "goermis") {
		t.Errorf("Re-creating the alias %v did not fail", aliasName)
	}
}

func TestDeleteAlias(t *testing.T) {
	bootstrap.ParseFlags()

	if !landbsoap.Conn().DNSDelegatedRemove(aliasName, view) {
		t.Errorf("Error creating the alias %v", aliasName)
	}
	fmt.Printf("And deleting it again should fail")
	if landbsoap.Conn().DNSDelegatedRemove(aliasName, view) {
		t.Errorf("A deleted alias can be deleted again %v", aliasName)
	}

}
