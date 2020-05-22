package landbsoap

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type LandbSoap struct {
	Username  string
	Password  string
	Ca        string
	HostCert  string
	HostKey   string
	Url       string
	AuthToken string
	Client    *http.Client
}

func (self *LandbSoap) InitConnection() error {
	caCert, err := ioutil.ReadFile(self.Ca)
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	cert, err := tls.LoadX509KeyPair(self.HostCert, self.HostKey)
	if err != nil {
		log.Fatal(err)
	}

	self.Client = &http.Client{
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
</Envelope>`, self.Username, self.Password)))
	authSoapAction := "urn:getAuthToken"

	authreq, err := http.NewRequest("POST", self.Url, bytes.NewReader(authpayload))
	if err != nil {
		log.Fatal("Error on creating request object. ", err.Error())
		return err
	}
	authreq.Header.Set("Content-type", "text/xml")
	authreq.Header.Set("SOAPAction", authSoapAction)
	authresp, err := self.Client.Do(authreq)
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

	self.AuthToken = authresult.Body.GetAuthTokenResponse.Token
	//fmt.Println("Token = " + self.AuthToken)

	return nil
}

func (self *LandbSoap) doSoap(payload []byte, soapAction, httpMethod string) bool {
	req, err := http.NewRequest(httpMethod, self.Url, bytes.NewReader(payload))
	if err != nil {
		log.Fatal("Error on creating request object. ", err.Error())
		return false
	}
	req.Header.Set("Content-type", "text/xml")
	req.Header.Set("SOAPAction", "urn:"+soapAction)
	resp, err := self.Client.Do(req)
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

func (self *LandbSoap) DnsDelegatedAdd(domain, view, keyname, description, userdescription string) bool {
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
</Envelope>`, self.AuthToken, domain, view, keyname, description, userdescription)))
	dnsDelegatedAddSoapAction := "dnsDelegatedAdd"
	dnsDelegatedAddHttpMethod := "POST"

	return self.doSoap(dnsDelegatedAddPayload, dnsDelegatedAddSoapAction, dnsDelegatedAddHttpMethod)

}

func (self *LandbSoap) DnsDelegatedAliasAdd(domain, view, alias string) bool {
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
</Envelope>`, self.AuthToken, domain, view, alias)))
	dnsDelegatedAliasAddSoapAction := "dnsDelegatedAliasAdd"
	dnsDelegatedAliasAddHttpMethod := "POST"

	return self.doSoap(dnsDelegatedAliasAddPayload, dnsDelegatedAliasAddSoapAction, dnsDelegatedAliasAddHttpMethod)
}

func (self *LandbSoap) DnsDelegatedRemove(domain, view string) bool {
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
</Envelope>`, self.AuthToken, domain, view)))
	dnsDelegatedRemoveSoapAction := "dnsDelegatedRemove"
	dnsDelegatedRemoveHttpMethod := "POST"

	return self.doSoap(dnsDelegatedRemovePayload, dnsDelegatedRemoveSoapAction, dnsDelegatedRemoveHttpMethod)
}

func (self *LandbSoap) DnsDelegatedAliasRemove(domain, view, alias string) bool {
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
</Envelope>`, self.AuthToken, domain, view, alias)))
	dnsDelegatedAliasRemoveSoapAction := "dnsDelegatedAliasRemove"
	dnsDelegatedAliasRemoveHttpMethod := "POST"

	return self.doSoap(dnsDelegatedAliasRemovePayload, dnsDelegatedAliasRemoveSoapAction, dnsDelegatedAliasRemoveHttpMethod)
}
