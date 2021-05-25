package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"gitlab.cern.ch/lb-experts/goermis/bootstrap"
)

var (
	secretsCache map[string][]string
	cfg          = bootstrap.GetConf()
)

//GetSecret queries tbag for the secret of an alias
func (l *UserAuth) get(aliasname string) []string {
	type msg struct {
		Content []string
	}
	var (
		message msg
	)
	//check local cache first
	if v, found := secretsCache[aliasname]; found && len(v) != 0 {
		return secretsCache[aliasname]
	}

	URL := l.authRogerBaseURL + cfg.Teigi.Host + "/secret/" + aliasname + "_secret"

	log.Info("Querying tbag for the secret of node" + aliasname + ". URL = " + URL)
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Errorf("Error on creating request object. %v ", err)
		return []string{}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := l.Client.Do(req)
	if err != nil {
		log.Errorf("Error on dispatching secret request to tbag for node  %v , error %v", aliasname, err)
		return []string{}
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error reading Body of Request while querying the secret of node %v , error: %v ", aliasname, err)
		return []string{}
	}

	if resp.StatusCode == http.StatusUnauthorized {
		log.Errorf("User not authorized.Status Code:%v ", resp.StatusCode)
		return []string{}
	}

	if err = json.Unmarshal(data, &message); err != nil {
		log.Errorf("Error on unmarshalling response from tbag %v", err)
		return []string{}
	}
	defer resp.Body.Close()
	//save locally
	secretsCache[aliasname] = message.Content

	return message.Content

}

//
func (l *UserAuth) post(aliasname, secret string) error {
	URL := l.authRogerBaseURL + cfg.Teigi.Host + "/secret/" + aliasname + "_secret"
	load := fmt.Sprintf("secret:%v", secret)
	jsonload, err := json.Marshal(load)
	if err != nil {
		return err
	}

	log.Infof("Creating new secret for alias %v in tbag", aliasname)
	req, err := http.NewRequest("POST", URL, bytes.NewReader(jsonload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := l.Client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("user unauthorized to POST new secret in tbag")
	}
	defer resp.Body.Close()

	return nil

}
func (l *UserAuth) delete(aliasname string) error {
	URL := l.authRogerBaseURL + cfg.Teigi.Host + "/secret/" + aliasname + "_secret"

	log.Infof("Deleting secret for alias %v in tbag", aliasname)
	req, err := http.NewRequest("DELETE", URL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := l.Client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("user unauthorized to POST new secret in tbag")
	}
	defer resp.Body.Close()

	return nil

}

func PostSecret(aliasname, secret string) error {
	tbagConn := getConn(cfg.Teigi.Tbag)
	if err := tbagConn.initConnection(); err != nil {
		return fmt.Errorf("error while initiating the tbag connection: https://woger.cern.ch:8202/tbag/v2/host/, error %v", err)
	}
	return tbagConn.post(aliasname, secret)
}

func DeleteSecret(aliasname string) error {
	tbagConn := getConn(cfg.Teigi.Tbag)
	if err := tbagConn.initConnection(); err != nil {
		return fmt.Errorf("error while initiating the tbag connection: https://woger.cern.ch:8202/tbag/v2/host/, error %v", err)
	}
	return tbagConn.delete(aliasname)
}

func GetSecret(aliasname string) []string {
	tbagConn := getConn(cfg.Teigi.Tbag)
	if err := tbagConn.initConnection(); err != nil {
		return []string{}
	}
	return tbagConn.get(aliasname)
}
