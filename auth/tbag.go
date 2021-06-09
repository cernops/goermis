package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"regexp"
	"strings"

	"github.com/jcmturner/gokrb5/v8/client"
	"github.com/jcmturner/gokrb5/v8/config"
	"github.com/jcmturner/gokrb5/v8/credentials"
	"github.com/jcmturner/gokrb5/v8/spnego"
	"gitlab.cern.ch/lb-experts/goermis/bootstrap"
)

var (
	secretsCache map[string][]string
	cfg          = bootstrap.GetConf()
)

const (
	kRB5CONF = `
  [libdefaults]
 default_realm = CERN.CH
 ticket_lifetime = 25h
 renew_lifetime = 120h
 forwardable = true 
 proxiable = true
 default_tkt_enctypes = arcfour-hmac-md5 aes256-cts 
 aes128-cts des3-cbc-sha1 des-cbc-md5 des-cbc-crc
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

	URL := l.authRogerBaseURL + cfg.Teigi.Host + "/secret/" + aliasname

	log.Info("Querying tbag for the secret of alias" + aliasname + ". URL = " + URL)
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
func (l *UserAuth) modify(method, aliasname, secret string) error {

	buffer := &bytes.Buffer{}

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
		log.Errorf("could create request: %v", err)
	}

	// Load the client krb5 config
	conf, err := config.NewFromString(kRB5CONF)
	if err != nil {
		log.Errorf("could not load krb5.conf: %v", err)
	}

	//load ccache
	ccache, err := credentials.LoadCCache(findcache())
	if err != nil {
		log.Errorf("could not load cache: %v", err)
	}

	// Create the client with ccache
	cl, err := client.NewFromCCache(
		ccache,
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
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("error reading response body: %v", err)
	}
	log.Info(string(b))
	return nil
}

func PostSecret(aliasname, secret string) error {
	tbagConn := getConn(cfg.Teigi.Krbtbag)
	if err := tbagConn.initConnection(); err != nil {
		return fmt.Errorf("error while initiating the tbag connection: %v, error %v", tbagConn, err)
	}
	return tbagConn.modify("POST", aliasname, secret)
}

func GetSecret(aliasname string) []string {
	tbagConn := getConn(cfg.Teigi.Ssltbag)
	if err := tbagConn.initConnection(); err != nil {
		return []string{}
	}
	return tbagConn.get(aliasname)
}

func DeleteSecret(aliasname string) error {
	tbagConn := getConn(cfg.Teigi.Krbtbag)
	if err := tbagConn.initConnection(); err != nil {
		return fmt.Errorf("error while initiating the tbag connection: %v, error %v", tbagConn, err)
	}
	return tbagConn.modify("DELETE", aliasname, "")
}

//execute shell commands
func execute(command string) (string, error) {
	var stdout bytes.Buffer
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = &stdout
	err := cmd.Run()
	return stdout.String(), err
}

//find path of ccache
func findcache() string {
	out, err := execute("klist -l")
	if err != nil {
		log.Printf("error: %v\n", err)
	}
	found, _ := regexp.MatchString("FILE:\\/tmp\\/krb5cc_[\\w]+", out)

	if !found {
		log.Fatal("ccache not found")

	}
	path := strings.Split(out, "FILE:")[1]
	return strings.TrimSpace(path)

}
