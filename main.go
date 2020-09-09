package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/labstack/gommon/log"
	"gitlab.cern.ch/lb-experts/goermis/aiermis/orm"
	"gitlab.cern.ch/lb-experts/goermis/bootstrap"
	"gitlab.cern.ch/lb-experts/goermis/db"
	"gitlab.cern.ch/lb-experts/goermis/router"
	"gitlab.cern.ch/lb-experts/goermis/views"
)

const (
	// Version number
	Version = "1.2.4"
	// Release number
	Release = "3"
)

func main() {
	bootstrap.ParseFlags()
	log.Info("Service Started...")

	// Echo instance
	echo := router.New()

	//Initiate template views
	views.InitViews(echo)

	//Create and keep up to date DB tables
	autoCreateTables(&orm.Alias{}, &orm.Node{}, &orm.Cname{}, &orm.Relation{})
	autoMigrateTables()

	// Start server
	go func() {
		cfg := bootstrap.GetConf()
		if err := echo.StartTLS(":8080",
			cfg.Certs.GoermisCert,
			cfg.Certs.GoermisKey); err != nil {
			log.Fatal("Failed to start server: " + err.Error())
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := echo.Shutdown(ctx); err != nil {
		log.Fatal("Fatal error while shutting server down")
	}

}

func autoCreateTables(values ...interface{}) error {
	for _, value := range values {
		if !db.ManagerDB().HasTable(value) {
			err := db.ManagerDB().CreateTable(value).Error
			gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
				return "ermis_api_" + defaultTableName
			}
			if err != nil {
				errClose := db.ManagerDB().Close()
				if errClose != nil {
					log.Error("Error while trying to close DB conn.")

				}
				return err

			}
		}
	}
	return nil
}

// autoMigrateTables: migrate table columns using GORM
func autoMigrateTables() {
	db.ManagerDB().AutoMigrate(&orm.Alias{}, &orm.Node{}, &orm.Cname{}, &orm.Relation{})

}
