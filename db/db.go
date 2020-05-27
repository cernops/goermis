package db

import (
	"fmt"

	"gitlab.cern.ch/lb-experts/goermis/bootstrap"

	_ "github.com/go-sql-driver/mysql" //need this
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" //need this too
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

var (
	db  *gorm.DB
	log = bootstrap.Log
)

// Init initialize database
func Init() {
	var adapter string
	adapter = bootstrap.App.IFConfig.String("adapter")
	switch adapter {
	case "mysql":
		mysqlConn()
		break
	default:
		log.WithFields(logrus.Fields{
			"package":  "db",
			"function": "Init",
		}).Panic("Undefined connection on config.yaml")

		panic("Undefined connection on config.yaml")
	}
}

// mysqlConn: setup mysql database connection using the configuration from database.yaml
func mysqlConn() {
	var (
		connectionString string
		err              error
	)

	connectionString = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", bootstrap.App.IFConfig.String("username"), bootstrap.App.IFConfig.String("password"), bootstrap.App.IFConfig.String("host"), bootstrap.App.IFConfig.String("port"), bootstrap.App.IFConfig.String("database"))

	if db, err = gorm.Open("mysql", connectionString); err != nil {
		log.WithFields(logrus.Fields{
			"package":  "db",
			"function": "mySqlConn",
			"error":    err,
		}).Panic("Database connection error")
		panic(err)
	}
	if err = db.DB().Ping(); err != nil {
		log.WithFields(logrus.Fields{
			"package":  "db",
			"function": "mySqlConn",
			"error":    err,
		}).Panic("Unreachable database")
		panic(err)
	}
	db.SingularTable(true)
	db.LogMode(true)
	db.DB().SetMaxIdleConns(bootstrap.App.IFConfig.Int("idle_conns"))
	db.DB().SetMaxOpenConns(bootstrap.App.IFConfig.Int("open_conns"))
}

//ManagerDB return GORM's database connection instance.
func ManagerDB() *gorm.DB {
	var adapter string
	adapter = bootstrap.App.IFConfig.String("adapter")
	switch adapter {
	case "mysql":
		mysqlConn()
		break
	default:
		log.WithFields(logrus.Fields{
			"package":  "db",
			"function": "ManagerDB",
		}).Panic("Undefined connection on config.yaml")

		panic("Undefined connection on config.yaml")
	}

	return db
}
