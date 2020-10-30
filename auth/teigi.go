package auth

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/http"
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

type pwnMsg struct {
	Hostgroup []string
}

//CheckCud checks a user if he is member of egroup
func (l *Group) CheckCud(username string) bool {

	if isMemberOf(username, "ermis-lbaas-admins") {
		return true
	}

	return false

}

//GetConn prepares the initial structure for starting a connection
func GetConn(url string) *UserAuth {
	var (
		cfg  = bootstrap.GetConf()
		conn = &UserAuth{
			authRogerBaseURL: url,
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
		log.Error(err)

	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	cert, err := tls.LoadX509KeyPair(l.authRogerCert, l.authRogerCertKey)
	if err != nil {
		log.Error(err)

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

//PwnHg queries teigi for the hostgroups where user is owner/memeber/privileged
func (l *UserAuth) PwnHg(username string) []string {
	var m pwnMsg
	URL := l.authRogerBaseURL + username + "/"
	log.Info("[" + username + "] Querying teigi for user's hostgroups. url = " + URL)
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Error("Error on creating request object. ", err.Error())
		return []string{}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := l.Client.Do(req)
	if err != nil {
		log.Error("["+username+"] Error on dispatching pwn request to teigi ", err.Error())
		return []string{}
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("["+username+"]Error reading Body of Request ", err.Error())
		return []string{}
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		log.Error("["+username+"]User not authorized.Status Code: ", resp.StatusCode)
		return []string{}
	}
	if err = json.Unmarshal(data, &m); err != nil {
		log.Error("["+username+"]Error on unmarshalling response from teigi ", err.Error())
		return []string{}
	}
	return m.Hostgroup

}

//GetPwn returns a list of hostgroups where the user is owner or privileged
func GetPwn(username string) (pwnedHg []string) {
	conn := GetConn("https://woger.cern.ch:8202/pwn/v1/owner/")
	if err := conn.InitConnection(); err != nil {
		log.Error("Error while contacting: https://woger.cern.ch:8202/pwn/v1/owner/" + err.Error())
		return []string{}

	}
	pwnedHg = conn.PwnHg(username)
	return


}
