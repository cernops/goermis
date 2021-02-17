package db

import (
	"fmt"
	"time"

	"gitlab.cern.ch/lb-experts/goermis/bootstrap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// GormLogger struct
type GormLogger struct{}

var (
	db  *gorm.DB
	cfg = bootstrap.GetConf()
	log = bootstrap.GetLog()
)

// mysqlConn: setup mysql database connection using the configuration from database.yaml
func mysqlConn() {
	var (
		err   error
		value int
	)
	//Loglevels {INFO,WARN,ERROR,SILENT}
	if *bootstrap.DebugLevel {
		value = 4 //INFO
	} else {
		value = 2  //ERROR
	}
	newLogger := logger.New(
		log,
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.LogLevel(value),
			Colorful:      true,
		},
	)

	connection := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", cfg.Database.Username, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Database)
	if db, err = gorm.Open(mysql.Open(connection), &gorm.Config{
		Logger: newLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
			TablePrefix:   "ermis_api_",
		},
	}); err != nil {
		fmt.Println("Connection  error")
	}
}

//ManagerDB return GORM's database connection instance.
func ManagerDB() *gorm.DB {
	var adapter string
	adapter = cfg.Database.Adapter
	if adapter == "mysql" {
		mysqlConn()
	} else {
		log.Panicf("Undefined connection '%s' on the configuration file", adapter)
	}

	return db
}
