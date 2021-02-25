package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql" //test
	"gitlab.cern.ch/lb-experts/goermis/bootstrap"

	gsql "gorm.io/driver/mysql"
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
		value = 2 //ERROR
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
	//A generic interface that allows us pinging, pooling, and managing idle connections
	sqlDB, err := sql.Open(cfg.Database.Adapter, connection)

	/*On top of the generic sql interface, we create a
	gorm interface that allows us to actually use the gorm tools
	Reference: https://gorm.io/docs/generic_interface.html */
	if db, err = gorm.Open(gsql.New(gsql.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: newLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
			TablePrefix:   "ermis_api_",
		},
	}); err != nil {
		sqlDB.Close()
		fmt.Println("Connection  error")
	}

	sqlDB.Ping()
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.Database.MaxIdleTime) * time.Minute)

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(cfg.Database.IdleConns)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(cfg.Database.OpenConns)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Minute)

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
