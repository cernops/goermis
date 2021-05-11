package auth

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

var aipwnConn *UserAuth

func init() {
	aipwnConn = getConn(cfg.Teigi.Pwn)
	if err := aipwnConn.initConnection(); err != nil {
		log.Errorf("error while initiating the pwn connection: %v, error: %v", cfg.Teigi.Pwn, err)
	}
}

//PwnHg queries teigi for the hostgroups where user is owner/memeber/privileged
func (l *UserAuth) pwnHg(username string) []string {
	type msg struct {
		Hostgroup []string
	}
	m := msg{}
	URL := l.authRogerBaseURL + username + "/"
	log.Infof("[%v] querying teigi for user's hostgroups. url = %v", username, URL)
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Errorf("error on creating request object %v ", err)
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
		log.Errorf("[%v] error reading body of request %v ", username, err)
		return []string{}
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		log.Errorf("[%v] user not authorized.status code: %v", username, resp.StatusCode)
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
