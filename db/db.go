package db

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql" //need this,please don't remove
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" //need this too
	"github.com/labstack/gommon/log"
	"gitlab.cern.ch/lb-experts/goermis/bootstrap"
)

// GormLogger struct
type GormLogger struct{}

var (
	db  *gorm.DB
	cfg = bootstrap.GetConf()
)

// mysqlConn: setup mysql database connection using the configuration from database.yaml
func mysqlConn() {
	var (
		connectionString string
		err              error
	)

	connectionString = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", cfg.Database.Username, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Database)
	if db, err = gorm.Open("mysql", connectionString); err != nil {
		//log.Panic("Database connection error")
		log.Info("Conn err")
	}
	if err = db.DB().Ping(); err != nil {
		//log.Panic("Unreachable database")
		log.Info("Unreachable db")
	}

	//Keep the table names singular(maintain conformity with existing DB)
	db.SingularTable(true)

	//Customize table names according to existing DB tables(alternative solution is to rename tables in DB)
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return "ermis_api_" + defaultTableName
	}

	//Enable logging
	db.LogMode(true)

	//Set our custom logger
	db.SetLogger(&GormLogger{})
	db.DB().SetMaxIdleConns(cfg.Database.IdleConns)
	db.DB().SetMaxOpenConns(cfg.Database.OpenConns)

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

// Print - Log Formatter
func (*GormLogger) Print(v ...interface{}) {
	//Print out only the issued sql command v[3] and the values v[4]
	log.Info(fmt.Sprintf("%v Value(s):%v\n", v[3], v[4]))

}
