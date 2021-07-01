package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/jcmturner/gokrb5/v8/client"
	"github.com/jcmturner/gokrb5/v8/config"
	"github.com/jcmturner/gokrb5/v8/spnego"
	"gitlab.cern.ch/lb-experts/goermis/bootstrap"
)

var (
	cfg = bootstrap.GetConf()
)

const (
	kRB5CONF = `
  [libdefaults]
 default_realm = CERN.CH
 ticket_lifetime = 25h
 renew_lifetime = 120h
 forwardable = true 
 proxiable = true
 default_tkt_enctypes = arcfour-hmac-md5 aes256-cts aes128-cts des3-cbc-sha1 des-cbc-md5 des-cbc-crc
 chpw_prompt = true
 rdns = true


[domain_realm]
.cern.ch = CERN.CH


[realms]
CERN.CH  = {
  default_domain = cern.ch
  kpasswd_server = cerndc.cern.ch
  admin_server = cerndc.cern.ch
  dns_lookup_kdc = false
  master_kdc = cerndc.cern.ch
  kdc = cerndc.cern.ch
}
`
)

var secretsCache = make(map[string]string)

//GetSecret queries tbag for the secret of an alias
func (l *UserAuth) get(aliasname string) string {
	type msg struct {
		Secret string
	}
	var (
		message msg
	)

	//check local cache first
	if v, found := secretsCache[aliasname]; found && len(v) != 0 {
		return secretsCache[aliasname]
	}
	//find hostname of node
	hostname, _ := os.Hostname()

	URL := l.authRogerBaseURL + hostname + "/secret/" + aliasname

	log.Info("Querying tbag for the secret of alias" + aliasname + ". URL = " + URL)
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Errorf("Error on creating request object. %v ", err)
		return ""
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := l.Client.Do(req)
	if err != nil {
		log.Errorf("Error on dispatching secret request to tbag for node  %v , error %v", aliasname, err)
		return ""
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error reading Body of Request while querying the secret of node %v , error: %v ", aliasname, err)
		return ""
	}

	if resp.StatusCode == http.StatusUnauthorized {
		log.Errorf("User not authorized.Status Code:%v ", resp.StatusCode)
		return ""
	}

	if err = json.Unmarshal(data, &message); err != nil {
		log.Errorf("Error on unmarshalling response from tbag %v", err)
		return ""
	}
	defer resp.Body.Close()
	//save locally
	secretsCache[aliasname] = message.Secret

	return message.Secret

}

//
func (l *UserAuth) modify(method, aliasname, secret string) error {

	buffer := &bytes.Buffer{}
	decodedPass, _ := base64.StdEncoding.DecodeString(cfg.Teigi.Password)

	//prepare secret
	if method == "POST" {
		secret := map[string]string{"secret": secret}
		json_secret, err := json.Marshal(secret)
		if err != nil {
			log.Error(err)
		}
		buffer.Write(json_secret)

	}

	//prepare URL
	URL := l.authRogerBaseURL + cfg.Teigi.Service + "/secret/" + aliasname

	//prepare request
	r, err := http.NewRequest(method, URL, buffer)
	if err != nil {
		log.Errorf("could not create request: %v", err)
	}

	// Load the client krb5 config
	conf, err := config.NewFromString(kRB5CONF)
	if err != nil {
		log.Errorf("could not load krb5.conf: %v", err)
	}

	// Create the client with ccache
	cl := client.NewWithPassword(
		cfg.Teigi.User,
		"CERN.CH",
		string(decodedPass),
		conf,
		client.DisablePAFXFAST(true),
	)
	if err != nil {
		log.Errorf("could not create client: %v", err)
	}

	// Log in the client
	err = cl.Login()
	if err != nil {
		log.Errorf("could not login client: %v", err)
	}

	spnegoCl := spnego.NewClient(cl, nil, "")

	// Make the request
	resp, err := spnegoCl.Do(r)
	if err != nil {
		log.Errorf("error making request: %v", err)
	}
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("error reading response body: %v", err)
	}
	return nil
}

func PostSecret(aliasname, secret string) error {
	conn := getConn(cfg.Teigi.Krbtbag, cfg.Certs.HostCert, cfg.Certs.HostKey)
	return conn.modify("POST", aliasname, secret)
}

func GetSecret(aliasname string) string {
	conn := getConn(cfg.Teigi.Ssltbag, cfg.Certs.HostCert, cfg.Certs.HostKey)
	return conn.get(aliasname)
}

func DeleteSecret(aliasname string) error {
	conn := getConn(cfg.Teigi.Krbtbag, cfg.Certs.HostCert, cfg.Certs.HostKey)
	//delete from local secret cache (basically a map)
	_, ok := secretsCache[aliasname]
	if ok {
		delete(secretsCache, aliasname)
	}
    //send a delete request to tbag
	return conn.modify("DELETE", aliasname, "")
}
