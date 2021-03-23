package main

/*This script checks for discrepancies in LanDB,
by comparing the Cnames defined internal view vs. external view
It uses the methods defined in the landb package*/

import (
	"fmt"
	"net/http"
	"os"
	"sort"

	landbsoap "gitlab.cern.ch/lb-experts/goermis/landb"
)

func main() {

	/*Use this for running the script autonomously from
	the goermis configuration. It still needs the landbsoap.go functions*/
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

	entries := ldbs.DNSDelegatedSearch("*")

	for i, v := range entries {
		for _, k := range entries[i+1:] {
			if v.Domain == k.Domain && v.View != k.View {
				if len(v.Aliases) > len(k.Aliases) {
					fmt.Printf("Extra aliases for %s in %s:%v\n", v.Domain, v.View, missing(k.Aliases, v.Aliases))

				} else if len(v.Aliases) < len(k.Aliases) {
					fmt.Printf("Extra aliases for %s in %s:%v\n", k.Domain, k.View, missing(v.Aliases, k.Aliases))

				} else {
					m := testEq(v.Aliases, k.Aliases)
					if len(m) > 0 {
						for key, value := range m {
							fmt.Printf("Alias %s has in %s:%s, while in %s:%s\n", v.Domain, v.View, key, k.View, value)
						}

					}

				}

			}
		}

	}

}

func missing(a, b []string) []string {
	ma := make(map[string]bool, len(a))
	diffs := []string{}
	for _, ka := range a {
		ma[ka] = true
	}
	for _, kb := range b {
		if !ma[kb] {
			diffs = append(diffs, kb)
		}
	}
	return diffs
}

func testEq(a, b []string) map[string]string {
	sort.Strings(a)
	sort.Strings(b)
	m := make(map[string]string)
	for i := range a {
		if a[i] != b[i] {
			m[a[i]] = b[i]
		}
	}

	return m
}
