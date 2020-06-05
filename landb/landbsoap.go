package landbsoap

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"gitlab.cern.ch/lb-experts/goermis/bootstrap"
)

//LandbSoap defines the structure
type LandbSoap struct {
	Username  string
	Password  string
	Ca        string
	HostCert  string
	HostKey   string
	URL       string
	AuthToken string
	Client    *http.Client
}

var (
	//Soap is the externally visible instance of Soap interface
	Soap = LandbSoap{}
)

func init() {
	password := bootstrap.App.IFConfig.String("soap_password")
	decodedPass, err := base64.StdEncoding.DecodeString(password)

	Soap = LandbSoap{
		Username:  bootstrap.App.IFConfig.String("soap_user"),
		Password:  string(decodedPass),
		Ca:        "/etc/ssl/certs/ca-bundle.crt",
		HostCert:  bootstrap.App.IFConfig.String("goermiscert"),
		HostKey:   bootstrap.App.IFConfig.String("goermiskey"),
		URL:       bootstrap.App.IFConfig.String("soap_url"),
		AuthToken: "",
		Client:    &http.Client{}}

	err = Soap.InitConnection()

	if err != nil {
		log.Fatal("Error initiating SOAP interface")
		//os.Exit(1)
	}
}

//InitConnection initiates a SOAP connection
func (landbself *LandbSoap) InitConnection() error {
	caCert, err := ioutil.ReadFile(landbself.Ca)
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	cert, err := tls.LoadX509KeyPair(landbself.HostCert, landbself.HostKey)
	if err != nil {
		log.Fatal(err)
	}

	landbself.Client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      caCertPool,
				Certificates: []tls.Certificate{cert},
			},
		},
	}

	authpayload := []byte(strings.TrimSpace(fmt.Sprintf(`
<Envelope xmlns="http://schemas.xmlsoap.org/soap/envelope/">
    <Body>
        <getAuthToken xmlns="urn:NetworkService">
            <Login>%s</Login>
            <Password>%s</Password>
            <Type>CERN</Type>
        </getAuthToken>
    </Body>
</Envelope>`, landbself.Username, landbself.Password)))
	authSoapAction := "urn:getAuthToken"

	authreq, err := http.NewRequest("POST", landbself.URL, bytes.NewReader(authpayload))
	if err != nil {
		log.Fatal("Error on creating request object. ", err.Error())
		return err
	}
	authreq.Header.Set("Content-type", "text/xml")
	authreq.Header.Set("SOAPAction", authSoapAction)
	authresp, err := landbself.Client.Do(authreq)
	if err != nil {
		log.Fatal("Error on dispatching request. ", err.Error())
		return err
	}

	authreqhtmlData, err := ioutil.ReadAll(authresp.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer authresp.Body.Close()
	fmt.Printf("Status is %v\n", authresp.Status)
	//fmt.Printf(string(authreqhtmlData) + "\n")

	type AuthToken struct {
		XMLName xml.Name
		Body    struct {
			XMLName              xml.Name
			GetAuthTokenResponse struct {
				XMLName xml.Name
				Token   string `xml:"token"`
			} `xml:"getAuthTokenResponse"`
		}
	}

	authresult := new(AuthToken)
	//err = xml.NewDecoder(authresp.Body).Decode(authresult)
	err = xml.NewDecoder(bytes.NewReader(authreqhtmlData)).Decode(authresult)
	if err != nil {
		log.Fatal("Error on unmarshaling xml. ", err.Error())
		return err
	}

	landbself.AuthToken = authresult.Body.GetAuthTokenResponse.Token
	//fmt.Println("Token = " + landbself.AuthToken)

	return nil
}

func (landbself *LandbSoap) doSoap(payload []byte, soapAction, httpMethod string) bool {
	req, err := http.NewRequest(httpMethod, landbself.URL, bytes.NewReader(payload))
	if err != nil {
		log.Fatal("Error on creating request object. ", err.Error())
		return false
	}
	req.Header.Set("Content-type", "text/xml")
	req.Header.Set("SOAPAction", "urn:"+soapAction)
	resp, err := landbself.Client.Do(req)
	if err != nil {
		log.Fatal("Error on dispatching request. ", err.Error())
		return false
	}

	htmlData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer resp.Body.Close()
	fmt.Printf("Status is %v\n", resp.Status)
	fmt.Printf(string(htmlData) + "\n")

	decoder := xml.NewDecoder(bytes.NewReader(htmlData))
	result := false
	fault := false
	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == soapAction+"Response" {
				// Get the next token after the dnsDelegatedAddResponse start element
				innerThing0, err := decoder.Token()
				if err != nil {
					break
				}
				fmt.Printf("innerThing = %v\n", innerThing0)
				fmt.Printf("innerThing = %T\n", innerThing0)
				innerThingVal, err := decoder.Token()
				if err != nil {
					break
				}
				fmt.Printf("innerThingVal = %v\n", innerThingVal)
				fmt.Printf("innerThingVal = %T\n", innerThingVal)
				chardata, _ := innerThingVal.(xml.CharData)
				result, err = strconv.ParseBool(string(chardata))
				if err != nil {
					break
				}
			}
			if se.Name.Local == "Fault" {
				fault = true
			}
			if fault {
				if se.Name.Local == "faultcode" {
					faultcodeVal, err := decoder.Token()
					if err != nil {
						break
					}
					//fmt.Printf("faultcode = %v\n", faultcodeVal)
					chardata, _ := faultcodeVal.(xml.CharData)
					fmt.Printf("faultcode = %v\n", string(chardata))
					continue
				}
				if se.Name.Local == "faultstring" {
					faultstringVal, err := decoder.Token()
					if err != nil {
						break
					}
					//fmt.Printf("faultstring = %v\n", faultstringVal)
					chardata, _ := faultstringVal.(xml.CharData)
					fmt.Printf("faultcode = %v\n", string(chardata))
					continue
				}
				if se.Name.Local == "detail" {
					detailVal, err := decoder.Token()
					if err != nil {
						break
					}
					//fmt.Printf("detail = %v\n", detailVal)
					chardata, _ := detailVal.(xml.CharData)
					fmt.Printf("detail = %v\n", string(chardata))
					continue
				}
			}
		}
	}
	fmt.Printf("Result = %v\n", result)
	//fmt.Printf("Result = %T\n", result
	return result
}

//DNSDelegatedAdd is a function to add a DNS delegated Zone
func (landbself *LandbSoap) DNSDelegatedAdd(domain, view, keyname, description, userdescription string) bool {
	dnsDelegatedAddPayload := []byte(strings.TrimSpace(fmt.Sprintf(`
<Envelope xmlns="http://schemas.xmlsoap.org/soap/envelope/">
    <Header>
        <Auth><token>%s</token></Auth>
    </Header>
    <Body>
        <dnsDelegatedAdd xmlns="urn:NetworkService">
            <DNSDelegatedInput>
                <Domain>%s</Domain>
                <View>%s</View>
                <KeyName>%s</KeyName>
                <Description>%s</Description>
                <UserDescription>%s</UserDescription>
            </DNSDelegatedInput>
        </dnsDelegatedAdd>
    </Body>
</Envelope>`, landbself.AuthToken, domain, view, keyname, description, userdescription)))
	dnsDelegatedAddSoapAction := "dnsDelegatedAdd"
	dnsDelegatedAddHTTPMethod := "POST"

	return landbself.doSoap(dnsDelegatedAddPayload, dnsDelegatedAddSoapAction, dnsDelegatedAddHTTPMethod)

}

//DNSDelegatedAliasAdd adds aliases for a defined domain
func (landbself *LandbSoap) DNSDelegatedAliasAdd(domain, view, alias string) bool {
	dnsDelegatedAliasAddPayload := []byte(strings.TrimSpace(fmt.Sprintf(`
<Envelope xmlns="http://schemas.xmlsoap.org/soap/envelope/">
    <Header>
        <Auth><token>%s</token></Auth>
    </Header>
    <Body>
        <dnsDelegatedAliasAdd xmlns="urn:NetworkService">
            <Domain>%s</Domain>
            <View>%s</View>
            <Alias>%s</Alias>
        </dnsDelegatedAliasAdd>
    </Body>
</Envelope>`, landbself.AuthToken, domain, view, alias)))
	dnsDelegatedAliasAddSoapAction := "dnsDelegatedAliasAdd"
	dnsDelegatedAliasAddHTTPMethod := "POST"

	return landbself.doSoap(dnsDelegatedAliasAddPayload, dnsDelegatedAliasAddSoapAction, dnsDelegatedAliasAddHTTPMethod)
}

//DNSDelegatedRemove deletes a domain from LANDB
func (landbself *LandbSoap) DNSDelegatedRemove(domain, view string) bool {
	dnsDelegatedRemovePayload := []byte(strings.TrimSpace(fmt.Sprintf(`
<Envelope xmlns="http://schemas.xmlsoap.org/soap/envelope/">
    <Header>
        <Auth><token>%s</token></Auth>
    </Header>
    <Body>
        <dnsDelegatedRemove xmlns="urn:NetworkService">
            <Domain>%s</Domain>
            <View>%s</View>
        </dnsDelegatedRemove>
    </Body>
</Envelope>`, landbself.AuthToken, domain, view)))
	dnsDelegatedRemoveSoapAction := "dnsDelegatedRemove"
	dnsDelegatedRemoveHTTPMethod := "POST"

	return landbself.doSoap(dnsDelegatedRemovePayload, dnsDelegatedRemoveSoapAction, dnsDelegatedRemoveHTTPMethod)
}

//DNSDelegatedAliasRemove deletes an alias for a defined domain
func (landbself *LandbSoap) DNSDelegatedAliasRemove(domain, view, alias string) bool {
	dnsDelegatedAliasRemovePayload := []byte(strings.TrimSpace(fmt.Sprintf(`
<Envelope xmlns="http://schemas.xmlsoap.org/soap/envelope/">
    <Header>
        <Auth><token>%s</token></Auth>
    </Header>
    <Body>
        <dnsDelegatedAliasRemove xmlns="urn:NetworkService">
            <Domain>%s</Domain>
            <View>%s</View>
            <Alias>%s</Alias>
        </dnsDelegatedAliasRemove>
    </Body>
</Envelope>`, landbself.AuthToken, domain, view, alias)))
	dnsDelegatedAliasRemoveSoapAction := "dnsDelegatedAliasRemove"
	dnsDelegatedAliasRemoveHTTPMethod := "POST"

	return landbself.doSoap(dnsDelegatedAliasRemovePayload, dnsDelegatedAliasRemoveSoapAction, dnsDelegatedAliasRemoveHTTPMethod)
}

//SearchResult serves as a blueprint for the query response
type SearchResult struct {
	XMLName xml.Name
	Body    struct {
		XMLName                    xml.Name
		DNSDelegatedSearchResponse struct {
			XMLName             xml.Name
			DNSDelegatedEntries []DNSDelegatedEntry `xml:"DNSDelegatedEntries>DNSDelegatedEntry"`
		} `xml:"dnsDelegatedSearchResponse"`
	}
}

//DNSDelegatedEntry is a blueprint for new entries in LANDB
type DNSDelegatedEntry struct {
	XMLName         xml.Name
	ID              int
	Domain          string
	View            string
	KeyName         string
	Description     string
	UserDescription string
	Aliases         []string `xml:"Aliases>item"`
}

//DNSDelegatedSearch queries LANDB for existing domain(s)
func (landbself *LandbSoap) DNSDelegatedSearch(search string) []DNSDelegatedEntry {
	dnsDelegatedSearchPayload := []byte(strings.TrimSpace(fmt.Sprintf(`
<Envelope xmlns="http://schemas.xmlsoap.org/soap/envelope/">
    <Header>
        <Auth><token>%s</token></Auth>
    </Header>
    <Body>
        <dnsDelegatedSearch xmlns="urn:NetworkService">
            <Search>%s</Search>
        </dnsDelegatedSearch>
    </Body>
</Envelope>`, landbself.AuthToken, search)))
	dnsDelegatedSearchSoapAction := "dnsDelegatedSearch"
	dnsDelegatedSearchHTTPMethod := "POST"

	req, err := http.NewRequest(dnsDelegatedSearchHTTPMethod, landbself.URL, bytes.NewReader(dnsDelegatedSearchPayload))
	if err != nil {
		log.Fatal("Error on creating request object. ", err.Error())
		return []DNSDelegatedEntry{}
	}
	req.Header.Set("Content-type", "text/xml")
	req.Header.Set("SOAPAction", "urn:"+dnsDelegatedSearchSoapAction)
	resp, err := landbself.Client.Do(req)
	if err != nil {
		log.Fatal("Error on dispatching request. ", err.Error())
		return []DNSDelegatedEntry{}
	}

	htmlData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return []DNSDelegatedEntry{}
	}
	defer resp.Body.Close()
	fmt.Printf("Status is %v\n", resp.Status)
	//fmt.Printf(string(htmlData) + "\n")

	result := new(SearchResult)
	err = xml.NewDecoder(bytes.NewReader(htmlData)).Decode(result)
	if err != nil {
		log.Fatal("Error on unmarshaling xml. ", err.Error())
		return []DNSDelegatedEntry{}
	}

	return result.Body.DNSDelegatedSearchResponse.DNSDelegatedEntries
}

//GiveMeCnamesOf returns an array of all aliases for a certain domain
func (landbself *LandbSoap) GiveMeCnamesOf(domain string) []string {
	entries := landbself.DNSDelegatedSearch(domain)
	if len(entries) != 0 {
		return entries[0].Aliases
	}
	return []string{}

}
