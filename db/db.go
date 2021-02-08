package db

import (
	"fmt"
	"log"

	"gitlab.cern.ch/lb-experts/goermis/bootstrap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
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
		err error
	)

	/*newLogger := logger.New(
	log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
	logger.Config{
		SlowThreshold: time.Second,   // Slow SQL threshold
		LogLevel:      logger.Silent, // Log level
		Colorful:      false})
	*/
	connection := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", cfg.Database.Username, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Database)
	if db, err = gorm.Open(mysql.Open(connection), &gorm.Config{

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

/*
//Print - Log Formatter
func (*newLogger) Print(v ...interface{}) {
	//Print out only the issued sql command v[3] and the values v[4]
	if len(v) > 3 {
		log.Debug(fmt.Sprintf("%v Value(s):%v\n", v[3], v[4]))
	} else {
		log.Debug(fmt.Sprintf("%v", v))
	}
}*/
