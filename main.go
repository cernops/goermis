package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/labstack/gommon/log"
	"gitlab.cern.ch/lb-experts/goermis/aiermis/api"
	"gitlab.cern.ch/lb-experts/goermis/alarms"
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
	autoCreateTables(&api.Alias{}, &api.Node{}, &api.Cname{}, &api.Alarm{}, &api.Relation{})
	autoMigrateTables()

	//Alarms periodic check/update
	log.Info("24 hours passed, preparing to execution check alarms")
	ticker := time.NewTicker(24 * time.Hour)
	/*done channel can be used to stop the ticker if needed,
	by issuing the command "done<-true". For now, it runs constantly */
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				ticker.Stop()
				return
			case <-ticker.C:
				alarms.PeriodicAlarmCheck()
			}
		}
	}()
	log.Info("Alarms updated")
	/* Start server
	       Error handling is done a bit differently in this situation. The reason is that
		   when server is restarted we force it to reuse the same socket. Despite being successfully
		   restarted, it throws a bind error. This interferes with other important cases where we need
		   to shut down the service */

	go func() {
		cfg := bootstrap.GetConf()
		err := echo.StartTLS(":8080",
			cfg.Certs.GoermisCert,
			cfg.Certs.GoermisKey)
		if !strings.HasSuffix(err.Error(), "bind: address already in use") {
			log.Fatal("Failed to start server: " + err.Error())

		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds. It is needed to accomplish socket recycling
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := echo.Shutdown(ctx); err != nil {
		log.Fatal("Fatal error while shutting server down " + err.Error())

	}

}

//GORM will create/migrate new data, but will not delete anything for security reasons
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

// autoMigrateTables: migrate table columns using GORM. Will not delete/change types for security reasons
func autoMigrateTables() {
	db.ManagerDB().AutoMigrate(&api.Alias{}, &api.Node{}, &api.Cname{}, &api.Alarm{}, &api.Relation{})

}
