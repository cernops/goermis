package main

import (
	"fmt"
	"gitlab.cern.ch/lb-experts/goermis/landb/landbsoap"
	"net/http"
	"os"
)

func main() {
	ldbs := landbsoap.LandbSoap{Username: "------",
		Password:  "-------",
		Ca:        "-------",
		HostCert:  "-------",
		HostKey:   "-------",
		Url:       "https://network.cern.ch/sc/soap/soap.fcgi?v=6",
		AuthToken: "",
		Client:    &http.Client{}}

	err := ldbs.InitConnection()
	if err != nil {
		//log.Fatal(err)
		fmt.Println(err)
		os.Exit(1)
	}
	//search := "*"
	search := "testigna*"
	entries := ldbs.DnsDelegatedSearch(search)
	for _, v := range entries {
		fmt.Printf("entry value = %v\n", v)
		fmt.Printf("entry type = %T\n", v)
		for _, item := range v.Aliases {
			fmt.Printf("Cname of %s = %v\n", v.Domain, item)
		}
	}
	domain := "testigna1"
	cnames := ldbs.GimeCnamesOf(domain)
	fmt.Printf("cnames of %s value = %v\n", domain, cnames)
	fmt.Printf("cnames of %s type = %T\n", domain, cnames)

	//os.Exit(0)

	domain = "testgosoap14.cern.ch"
	view := "internal"
	if ldbs.DnsDelegatedAdd(domain, view, "ITPES-INTERNAL", "Created by: go", "My go test") {
		fmt.Println(domain + "/" + view + " has been created")
	}
	domain = "testgosoap12.cern.ch"
	view = "internal"
	alias := "testgosoapdouce"
	if ldbs.DnsDelegatedAliasAdd(domain, view, alias) {
		fmt.Println("Alias " + alias + " has been created for " + domain + "/" + view)
	}
	domain = "testgosoap3.cern.ch"
	view = "internal"
	if ldbs.DnsDelegatedRemove(domain, view) {
		fmt.Println(domain + "/" + view + " has been removed")
	}
	domain = "testgosoap11.cern.ch"
	view = "internal"
	alias = "testgosoaponce"
	if ldbs.DnsDelegatedAliasRemove(domain, view, alias) {
		fmt.Println("Alias " + alias + " has been removed for " + domain + "/" + view)
	}
}
