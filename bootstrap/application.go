package bootstrap

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/labstack/gommon/log"
	"gopkg.in/yaml.v3"
)

//Config describes the yaml file
type Config struct {
	App struct {
		AppName    string `yaml:"app_name"`
		AppVersion string `yaml:"app_version"`
		AppEnv     string `yaml:"app_env"`
	}
	Database struct {
		Adapter   string
		Database  string
		Username  string
		Password  string
		Host      string
		Port      int
		IdleConns int `yaml:"idle_conns"`
		OpenConns int `yaml:"open_conns"`
		Sslmode   string
	}
	Soap struct {
		SoapUser     string `yaml:"soap_user"`
		SoapPassword string `yaml:"soap_password"`
		SoapKeynameI string `yaml:"soap_keyname_i"`
		SoapKeynameE string `yaml:"soap_keyname_e"`
		SoapURL      string `yaml:"soap_url"`
	}
	Certs struct {
		GoermisCert string `yaml:"goermis_cert"`
		GoermisKey  string `yaml:"goermis_key"`
		CACert      string `yaml:"ca_cert"`
	}
	Log struct {
		LoggingFile string `yaml:"logging_file"`
		Stdout      bool
	}
}

var (
	configFileFlag = flag.String("config", "/usr/local/etc/goermis.yaml", "specify configuration file path")
	//HomeFlag grabs the location of staticfiles & templates
	HomeFlag   = flag.String("home", "/var/lib/ermis/", "specify statics path")
	debugLevel = flag.Bool("debug", false, "display debug messages")
)

//ParseFlags checks the command line arguments
func ParseFlags() {
	//Parse flags
	flag.Parse()
	//Init log in the bootstrap package, since its the first that its executed
	if *debugLevel {
		log.SetLevel(1)
	} else {
		log.SetLevel(2)
	}

	//Init log in the bootstrap package, since its the first that its executed

	log.SetHeader("${time_rfc3339} ${level} ${short_file} ${line} ")
	file, err := os.OpenFile(GetConf().Log.LoggingFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Error("Failed to log to file, using default stderr" + err.Error())
	}
	if GetConf().Log.Stdout {
		log.Info("File and console set as output")
		mw := io.MultiWriter(os.Stdout, file)
		log.SetOutput(mw)
	} else {
		log.Info("File set as logger output")
		log.SetOutput(file)

	}

}

//GetConf returns the Conf file
func GetConf() *Config {
	cfg, err := NewConfig(*configFileFlag)
	if err != nil {
		log.Fatal(err)

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
		return fmt.Errorf("'%s' is a directory, not a normal file", path)
	}
	return nil
}
