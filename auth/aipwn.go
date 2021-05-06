package auth

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

var aipwnConn *UserAuth

func init() {
	aipwnConn = getConn("https://woger.cern.ch:8202/pwn/v1/owner/")
	if err := aipwnConn.initConnection(); err != nil {
		log.Error("Error while initiating the ai-pwn connection: https://woger.cern.ch:8202/pwn/v1/owner/" + err.Error())
	}
}

//PwnHg queries teigi for the hostgroups where user is owner/memeber/privileged
func (l *UserAuth) pwnHg(username string) []string {
	type msg struct {
		Hostgroup []string
	}
	m := msg{}
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
	return aipwnConn.pwnHg(username)
}
