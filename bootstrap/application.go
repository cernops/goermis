package bootstrap

import (
	"flag"
	"fmt"
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
		Adapter   string `yaml:"adapter"`
		Database  string `yaml:"database"`
		Username  string `yaml:"username"`
		Password  string `yaml:"password"`
		Host      string `yaml:"host"`
		Port      int    `yaml:"port"`
		IdleConns int    `yaml:"idle_conns"`
		OpenConns int    `yaml:"open_conns"`
		Sslmode   string `yaml:"sslmode"`
	}
	Soap struct {
		SoapUser     string `yaml:"soap_user"`
		SoapPassword string `yaml:"soap_password"`
		SoapKeynameI string `yaml:"soap_keyname_i"`
		SoapKeynameE string `yaml:"soap_keyname_e"`
		SoapURL      string `yaml:"soap_url"`
	}
	Certs struct {
		GoermisCert string `yaml:"goermiscert"`
		GoermisKey  string `yaml:"goermiskey"`
		CACert      string `yaml:"cacert"`
	}
	Log struct {
		LoggingFile string `yaml:"logging_file"`
	}
}

var (
	configFileFlag = flag.String("config", "/usr/local/etc/goermis.yaml", "specify configuration file path")

	//HomeFlag grabs the location of staticfiles & templates
	HomeFlag = flag.String("home", "/var/lib/ermis/", "specify statics path")
)

func init() {
	log.SetLevel(1)
	log.SetHeader("${time_rfc3339} ${level} ${short_file} ${line} ")
	file, err := os.OpenFile(GetConf().Log.LoggingFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(file)
		log.Info("File set as logger output")
	} else {
		log.Info("Failed to log to file, using default stderr")
	}
	log.Info("Init of the application")
	flag.Parse()
}

//GetConf returns the Conf file
func GetConf() (conf *Config) {
	cfg, err := NewConfig(*configFileFlag)
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}

// NewConfig returns a new decoded Config struct
func NewConfig(configPath string) (*Config, error) {
	// Create config structure
	config := &Config{}
	//Validate
	if err := ValidateConfigPath(configPath); err != nil {
		return nil, err
	}
	// Open config file
	file, err := os.Open(configPath)
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

// ValidateConfigPath just makes sure, that the path provided is a file,
// that can be read
func ValidateConfigPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a normal file", path)
	}
	return nil
}
