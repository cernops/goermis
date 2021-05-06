package auth

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"time"

	"gitlab.cern.ch/lb-experts/goermis/bootstrap"
)

//UserAuth describes a teigi connection
type UserAuth struct {
	authRogerBaseURL string
	authRogerCert    string
	authRogerCertKey string
	authRogerCA      string
	authRogerTimeout int
	Client           *http.Client
}

//GetConn prepares the initial structure for starting a connection
func getConn(url string) *UserAuth {
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
func (l *UserAuth) initConnection() error {
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
