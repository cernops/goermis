package dns

import (
	"encoding/base64"

	log "github.com/sirupsen/logrus"
	"gitlab.cern.ch/lb-experts/goermis/bootstrap"
)

var errs error

//SOAPConn is a struct with the soap connection params
type SOAPConn struct {
	User     string
	Password string
	KeynameI string
	KeynameE string
	URL      string
}

func init() {

	conn := &SOAPConn{}
	conn.User = bootstrap.App.IFConfig.String("soap_user")
	conn.Password, errs = base64Decode(bootstrap.App.IFConfig.String("soap_password"))
	if errs != nil {
		log.Error("Error while decoding SOAP Password", errs.Error())
	}
	conn.KeynameI = bootstrap.App.IFConfig.String("soap_keyname_i")
	conn.KeynameE = bootstrap.App.IFConfig.String("soap_keyname_e")
	conn.URL = bootstrap.App.IFConfig.String("soap_url")
	conn.initConnection()

}
//initConnection is a function that initiates a SOAP connection
func (conn SOAPConn) initConnection() {


}
func base64Decode(encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(data), nil

}
