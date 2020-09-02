package auth

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"gitlab.cern.ch/lb-experts/goermis/bootstrap"

	"github.com/labstack/gommon/log"
)

//Group decides user authorization
type Group struct {
	egroupCRUD string
}

//UserAuth describes a teigi connection
type UserAuth struct {
	authRogerBaseURL string
	authRogerCert    string
	authRogerCertKey string
	authRogerCA      string
	authRogerTimeout int
	Client           *http.Client
}
type message struct {
	Authorized bool
	Hostgroup  string
	Requestor  string
}

//CheckCud checks a user if he is member of egroup
func (l *Group) CheckCud(username string) bool {

	if isMemberOf(username, "ermis-lbaas-admins") {
		return true
	}
	log.Info(username + " is not member of egroup")
	return false

}

//GetConn prepares the initial structure for starting a connection
func GetConn() *UserAuth {
	var (
		cfg  = bootstrap.GetConf()
		conn = &UserAuth{
			authRogerBaseURL: "https://woger.cern.ch:8202/authz/v1/hostgroup/",
			authRogerCert:    cfg.Certs.GoermisCert,
			authRogerCertKey: cfg.Certs.GoermisKey,
			authRogerCA:      cfg.Certs.CACert,
			authRogerTimeout: 5,
		}
	)

	return conn
}

//InitConnection initiates a new connection with teigi
func (l *UserAuth) InitConnection() error {
	caCert, err := ioutil.ReadFile(l.authRogerCA)
	if err != nil {
		log.Fatal(err)

	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	cert, err := tls.LoadX509KeyPair(l.authRogerCert, l.authRogerCertKey)
	if err != nil {
		log.Fatal(err)

	}

	l.Client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      caCertPool,
				Certificates: []tls.Certificate{cert},
			},
		},
		Timeout: time.Duration(5 * time.Second),
	}
	return nil
}

//CheckWithForeman authorizes a user
func (l *UserAuth) CheckWithForeman(username string, group string) bool {
	var m message
	URL := l.authRogerBaseURL + strings.Split(group, "/")[0] + "/username/" + username + "/"
	log.Info("[" + username + "] Querying teigi for authorization. url = " + URL)
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Error("Error on creating request object. ", err.Error())
		return false
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := l.Client.Do(req)
	if err != nil {
		log.Error("["+username+"] Error on dispatching authorization request to teigi ", err.Error())
		return false
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("["+username+"]Error reading Body of Request ", err.Error())
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		log.Error("["+username+"]User not authorized.Status Code: ", resp.StatusCode)
		return false
	}
	if err = json.Unmarshal(data, &m); err != nil {
		log.Error("["+username+"]Error on unmarshalling response from teigi ", err.Error())
		return false
	}
	if m.Authorized {
		log.Info("[" + username + "]Foreman authorized user with hostgroup " + m.Hostgroup)
		return true
	}
	return false

}
