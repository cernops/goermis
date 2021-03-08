package ci

import (
	"fmt"
	"reflect"
	"testing"

	"gitlab.cern.ch/lb-experts/goermis/bootstrap"
)

func TestValidateConfigFile(t *testing.T) {
	type test struct {
		caseID    int
		input     string
		expectErr bool
	}
	testCases := []test{
		{
			caseID:    1,
			input:     ".",
			expectErr: true},

		{
			caseID:    2,
			input:     "test_config.yaml",
			expectErr: false},
	}
	for _, tc := range testCases {
		output := bootstrap.ValidateConfigFile(tc.input)
		if output != nil && !tc.expectErr {
			t.Errorf("Failed in TestValidateConfigFile\nFAILED CASE ID:%v\nINPUT:%v\nERROR:%v\n", tc.caseID, tc.input, output)
		}
	}

}

func TestNewConfig(t *testing.T) {
	testCase := bootstrap.Config{
		App: bootstrap.App{
			AppName:    "Goermis",
			AppVersion: "1.2.3",
			AppEnv:     "dev",
		},
		Database: bootstrap.Database{
			Adapter:         "mysql",
			Database:        "dummydatabase",
			Username:        "dummyusername",
			Password:        "dummypwd",
			Host:            "host.cern.ch",
			Port:            9999,
			IdleConns:       10,
			OpenConns:       100,
			MaxIdleTime:     2,
			ConnMaxLifetime: 10,
			Sslmode:         "disable",
		},
		Soap: bootstrap.Soap{
			SoapUser:     "dummyuser",
			SoapPassword: "FfdksDSSO!1",
			SoapKeynameI: "ITPES-INTERNAL",
			SoapKeynameE: "ITPES-EXTERNAL",
			SoapURL:      "https://network.cern.ch/sc/soap/soap.fcgi?v=6",
		},
		Certs: bootstrap.Certs{
			GoermisCert: "/etc/httpd/conf/ermiscert.pem",
			GoermisKey:  "/etc/httpd/conf/ermiskey.pem",
			CACert:      "/etc/httpd/conf/ca.pem",
		},
		Log: bootstrap.Logging{
			LoggingFile: "/var/log/ermis/ermis.log",
			Stdout:      true,
		},
	}
	fmt.Println("Now we will check that config file is readable")
	output, err := bootstrap.NewConfig("test_config.yaml")
	if err != nil {
		t.Errorf("Failed in TestNewConfig with ERROR:%v", err)

	}
	if !reflect.DeepEqual(*output, testCase) {
		t.Errorf("Failed to generate the config params properly.\nEXPECTED:%v\nRECEIVED:%v\n", testCase, *output)

	}

}
