package main

/*This script can be used for testing || manipulating Landb manually*/

import (
	"fmt"
	"net/http"
	"os"

	landbsoap "gitlab.cern.ch/lb-experts/goermis/landb"
)

var ldbs landbsoap.LandbSoap

func landbcrud() {
	ldbs := landbsoap.LandbSoap{
		Username:  "--change--",
		Password:  "--change--",
		Ca:        "/etc/ssl/certs/ca-bundle.crt",
		HostCert:  "/etc/ssl/goermiscert.pem",
		HostKey:   "/etc/ssl/goermiskey.pem",
		URL:       "https://network.cern.ch/sc/soap/soap.fcgi?v=6",
		AuthToken: "",
		Client:    &http.Client{}}

	err := ldbs.InitConnection()
	if err != nil {
		//log.Fatal(err)
		fmt.Println(err)
		os.Exit(1)
	}
	//Possible methods

	//search
	search("test*")
	//get cnames
	cnames("test.cern.ch")

	//create domain
	createalias("test.cern.ch", "internal")

	//create cname
	createcnames("test.cern.ch", "internal", "cname1")

	//delete cname
	deletecname("test.cern.ch", "internal", "cname1")

	//delete alias
	deletealias("test.cern.ch", "internal")
	os.Exit(0)
}

//Retrieve cnames of domain
func cnames(domain string) {
	cnames := ldbs.GimeCnamesOf(domain)
	fmt.Printf("cnames of %s value = %v\n", domain, cnames)
	fmt.Printf("cnames of %s type = %T\n", domain, cnames)
}

//Creates domains
func createalias(domain, view string) {
	if ldbs.DNSDelegatedAdd(domain, view, "ITPES-INTERNAL", "Created by: go", "My go test") {
		fmt.Println(domain + "/" + view + " has been created")
	}
}
func createcnames(domain, view, alias string) {
	if ldbs.DNSDelegatedAliasAdd(domain, view, alias) {
		fmt.Println("Alias " + alias + " has been created for " + domain + "/" + view)
	}

}

func search(search string) {
	entries := ldbs.DNSDelegatedSearch(search)
	for _, v := range entries {
		fmt.Printf("entry value = %v\n", v)
		for _, item := range v.Aliases {
			fmt.Printf("Cname of %s = %v\n", v.Domain, item)
		}
	}

}

func deletealias(domain, view string) {

	if ldbs.DNSDelegatedRemove(domain, view) {
		fmt.Println(domain + "/" + view + " has been removed")
	}

}

func deletecname(domain, view, alias string) {
	if ldbs.DNSDelegatedAliasRemove(domain, view, alias) {
		fmt.Println("Alias " + alias + " has been removed for " + domain + "/" + view)
	}
}
