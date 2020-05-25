package bootstrap

import (
	"flag"
	"fmt"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	configFileFlag = flag.String("config", "config.yaml", "specify configuration file path")

	//HomeFlag grabs the location of staticfiles & templates
	HomeFlag string
)

//App prototype
var App *Application

//Config file
type Config viper.Viper

//Application struct
type Application struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	ENV       string `json:"env"`
	AppConfig Config `json:"application_config"`
	IFConfig  Config `json:"interface_config"`
}

func init() {
	flag.StringVar(&HomeFlag, "home", "/goermis", "specify statics path")
	flag.Parse()
	App = &Application{}
	App.Name = "APP_NAME"
	App.Version = "APP_VERSION"
	App.loadENV()
	App.loadAppConfig()
	App.loadIFConfig()
}

// loadAppConfig: read application config and build viper object
func (app *Application) loadAppConfig() {
	var (
		appConfig *viper.Viper
		err       error
	)
	appConfig = viper.New()
	appConfig.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	appConfig.SetEnvPrefix("APP_")
	appConfig.AutomaticEnv()
	appConfig.SetConfigName("config")
	appConfig.AddConfigPath(*configFileFlag)
	appConfig.SetConfigType("yaml")
	if err = appConfig.ReadInConfig(); err != nil {
		panic(err)
	}
	appConfig.WatchConfig()
	appConfig.OnConfigChange(func(e fsnotify.Event) {
		Log.Info("App Config file changed %s:", e.Name)
	})
	app.AppConfig = Config(*appConfig)
}

// loadDBConfig: read application config and build viper object
func (app *Application) loadIFConfig() {
	var (
		ifConfig *viper.Viper
		err      error
	)
	ifConfig = viper.New()
	ifConfig.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	ifConfig.SetEnvPrefix("IF_")
	ifConfig.AutomaticEnv()
	ifConfig.SetConfigName("config")
	ifConfig.AddConfigPath(*configFileFlag)
	ifConfig.SetConfigType("yaml")
	if err = ifConfig.ReadInConfig(); err != nil {
		panic(err)
	}
	ifConfig.WatchConfig()
	ifConfig.OnConfigChange(func(e fsnotify.Event) {
		Log.Info("App Config file changed %s:", e.Name)
	})
	app.IFConfig = Config(*ifConfig)
}

// loadENV
func (app *Application) loadENV() {
	var APPENV string
	var appConfig viper.Viper
	appConfig = viper.Viper(app.AppConfig)
	APPENV = appConfig.GetString("env")
	switch APPENV {
	case "dev":
		app.ENV = "dev"
		break
	case "staging":
		app.ENV = "staging"
		break
	case "production":
		app.ENV = "production"
		break
	default:
		app.ENV = "dev"
		break
	}
}

// String: read string value from viper.Viper
func (config *Config) String(key string) string {
	var viperConfig viper.Viper
	viperConfig = viper.Viper(*config)
	return viperConfig.GetString(fmt.Sprintf("%s.%s", App.ENV, key))
}

//Int read int value from viper.Viper
func (config *Config) Int(key string) int {
	var viperConfig viper.Viper
	viperConfig = viper.Viper(*config)
	return viperConfig.GetInt(fmt.Sprintf("%s.%s", App.ENV, key))
}

// Boolean read boolean value from viper.Viper
func (config *Config) Boolean(key string) bool {
	var viperConfig viper.Viper
	viperConfig = viper.Viper(*config)
	return viperConfig.GetBool(fmt.Sprintf("%s.%s", App.ENV, key))
}
