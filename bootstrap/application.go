package bootstrap

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/labstack/gommon/log"
	"gopkg.in/yaml.v3"
)

type (
	//Config describes the yaml file
	Config struct {
		App      App
		Database Database
		Soap     Soap
		Certs    Certs
		Log      Logging
		DNS      DNS
		Timers   Timers
		Teigi    Teigi
	}
	//App struct describes application config parameters
	App struct {
		AppName    string `yaml:"app_name"`
		AppVersion string `yaml:"app_version"`
		AppEnv     string `yaml:"app_env"`
	}
	//Database struct describes database connection params
	Database struct {
		Adapter         string
		Database        string
		Username        string
		Password        string
		Host            string
		Port            int
		IdleConns       int `yaml:"idle_conns"`
		OpenConns       int `yaml:"open_conns"`
		MaxIdleTime     int `yaml:"max_idle_time"`
		ConnMaxLifetime int `yaml:"conn_max_lifetime"`
		Sslmode         string
	}
	//Soap describes the params for LanDB connection
	Soap struct {
		SoapUser     string `yaml:"soap_user"`
		SoapPassword string `yaml:"soap_password"`
		SoapKeynameI string `yaml:"soap_keyname_i"`
		SoapKeynameE string `yaml:"soap_keyname_e"`
		SoapURL      string `yaml:"soap_url"`
	}
	//Certs describes the service certificates
	Certs struct {
		GoermisCert string `yaml:"goermis_cert"`
		GoermisKey  string `yaml:"goermis_key"`
		CACert      string `yaml:"ca_cert"`
	}
	//Logging describes logging params
	Logging struct {
		LoggingFile string `yaml:"logging_file"`
		Stdout      bool
	}
	//DNS describes the config params for the DNS Manager
	DNS struct {
		Manager string
	}
	//Timers describes the parameters for configuring the different timers
	Timers struct {
		Alarms int
	}
	//The host which has access to tbag for saving the secrets
	Teigi struct {
		Host    string
		Service string
		Ssltbag string
		Krbtbag string
		Pwn     string
	}
)

var (
	configFileFlag = flag.String("config", "/usr/local/etc/goermis.yaml", "specify configuration file path")
	//HomeFlag grabs the location of staticfiles & templates
	HomeFlag = flag.String("home", "/var/lib/ermis/", "specify statics path")
	//DebugLevel enable flag
	DebugLevel = flag.Bool("debug", false, "display debug messages")
	//Log tesst
	Log = log.New("\r\n")
)

func init() {
	//Init log in the bootstrap package, since its the first that its executed
	if *DebugLevel {
		Log.SetLevel(1) //DEBUG
	} else {
		Log.SetLevel(2) //INFO
	}

	//Init log in the bootstrap package, since its the first that its executed

	Log.SetHeader("${time_rfc3339} ${level} ${short_file} ${line} ")
	file, err := os.OpenFile(GetConf().Log.LoggingFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		Log.Errorf("failed to log to file, using default stderr %v", err)
	}
	if GetConf().Log.Stdout {
		mw := io.MultiWriter(os.Stdout, file)
		Log.SetOutput(mw)
		Log.Info("file and console set as output")

	} else {
		Log.Info("file set as logger output")
		Log.SetOutput(file)

	}

}

//ParseFlags checks the command line arguments
func ParseFlags() {
	//Parse flags
	flag.Parse()

}

//GetConf returns the Conf file
func GetConf() *Config {
	cfg, err := NewConfig(*configFileFlag)
	if err != nil {
		Log.Fatal(err)

	}
	return cfg
}

// NewConfig returns a new decoded Config struct
func NewConfig(configFileFlag string) (*Config, error) {
	// Create config structure
	config := &Config{}
	//Validate its a readable file
	if err := ValidateConfigFile(configFileFlag); err != nil {
		return nil, err
	}
	// Open config file
	file, err := os.Open(configFileFlag)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

// ValidateConfigFile just makes sure, that the path provided is a file,
// that can be read
func ValidateConfigFile(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		e := fmt.Errorf("provided filepath for the config file is a directory")
		log.Error(e)
		return e
	}
	return nil
}

//GetLog returns the log instance
func GetLog() *log.Logger {
	return Log
}
