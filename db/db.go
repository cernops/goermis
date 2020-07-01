package db

import (
	"fmt"

	"github.com/labstack/gommon/log"

	"github.com/jinzhu/gorm"
	"gitlab.cern.ch/lb-experts/goermis/bootstrap"

	_ "github.com/go-sql-driver/mysql" //need this

	_ "github.com/jinzhu/gorm/dialects/mysql" //need this too
)

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
	log.Info(connectionString)
	if db, err = gorm.Open("mysql", connectionString); err != nil {
		log.Panic("Database connection error")
	}
	if err = db.DB().Ping(); err != nil {
		log.Panic("Unreachable database")
	}

	db.SingularTable(true)
	db.LogMode(true)
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
