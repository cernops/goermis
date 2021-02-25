package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql" // driver for sql connection
	"gitlab.cern.ch/lb-experts/goermis/bootstrap"

	gsql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// GormLogger struct
type GormLogger struct{}

var (
	//Conn serves to have a single connection to DB
	conn  *gorm.DB
	sqlDB *sql.DB
	cfg   = bootstrap.GetConf()
	log   = bootstrap.GetLog()
	err   error
	value int
)

// Initiating the db connection
func init() {
	//Loglevels for GORM are {INFO,WARN,ERROR,SILENT}
	if *bootstrap.DebugLevel {
		value = 4 //this is INFO
	} else {
		value = 2 //this is ERROR --> works like DEBUG in this case
	}

	newLogger := logger.New(
		log, //Here we pass the GORM logs to our default logger
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.LogLevel(value),
			Colorful:      true,
		},
	)

	connection := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", cfg.Database.Username, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Database)

	//A generic interface that allows us pinging, pooling, and managing idle connections
	sqlDB, err = sql.Open(cfg.Database.Adapter, connection)

	/*On top of the generic sql interface, we create a
	gorm interface that allows us to actually use the gorm tools
	Reference: https://gorm.io/docs/generic_interface.html */
	if conn, err = gorm.Open(gsql.New(gsql.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: newLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, //keeps table names singular
			TablePrefix:   "ermis_api_",
		},
	}); err != nil {
		sqlDB.Close()
		fmt.Println("Connection  error")
	}

	sqlDB.Ping()

	////ENABLE WHEN GO v.1.15 is supported by C8//////////////
	/*This can be used only with go version > 1.15. As of now, Feb 2021 , it cannot be
	//used because that version is not yet supported in CC8*/
	//sqlDB.SetConnMaxIdleTime(time.Duration(cfg.Database.MaxIdleTime) * time.Minute)
	/////////////////////////////////////////////////////

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(cfg.Database.IdleConns)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(cfg.Database.OpenConns)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Minute)
}

//GetConn returns a db connection
func GetConn() *gorm.DB {
	return conn
}

//Close connection
func Close() {
	sqlDB.Close()
}
