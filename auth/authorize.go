package auth

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
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

//CheckAuthorization checks if user is in the egroup and if he is allowed to create in the hostgroup
func CheckAuthorization(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := "kkouros"
		hostgroup := c.FormValue("hostgroup")
		teigi := true
		ldap := false

		log.Info(hostgroup)
		conn := GetConn()
		var d Group
		if err := conn.InitConnection(); err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		if hostgroup != "" && c.Request().Method == "POST" {
			teigi = conn.CheckWithForeman(username, hostgroup)
		}
		if username != "" {
			ldap = d.CheckCrud(username)
		}
		if teigi && ldap {
			return next(c)
		}
		return c.JSON(http.StatusUnauthorized, "Unauthorized")
	}
}

//CheckCrud checks a user if he is member of egroup
func (l *Group) CheckCrud(username string) bool {

	if IsMemberOf(username, "ermis-lbaas-admins") {
		return true
	}
	return false

}

//GetConn prepares the initial structure for starting a connection
func GetConn() *UserAuth {
	var conn = &UserAuth{
		authRogerBaseURL: "https://woger.cern.ch:8202/authz/v1/hostgroup/",
		authRogerCert:    "/etc/httpd/conf/ermiscert.pem",
		authRogerCertKey: "/etc/httpd/conf/ermiskey.pem",
		authRogerCA:      "/etc/httpd/conf/ca.pem",
		authRogerTimeout: 5,
	}

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
		log.Fatal("Error on creating request object. ", err.Error())
		return false
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := l.Client.Do(req)
	if err != nil {
		log.Fatal("["+username+"] Error on dispatching authorization request to teigi ", err.Error())
		return false
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("["+username+"]Error reading Body of Request ", err.Error())
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		log.Fatal("["+username+"]User not authorized.Status Code: ", resp.StatusCode)
		return false
	}
	log.Info(data)
	if err = json.Unmarshal(data, &m); err != nil {
		log.Fatal("["+username+"]Error on unmarshalling response from teigi ", err.Error())
		return false
	}
	if m.Authorized {
		fmt.Println("[" + username + "]Foreman authorized user with hostgroup " + m.Hostgroup)
		return true
	}
	return false

}
