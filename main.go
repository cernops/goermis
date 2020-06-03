package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/gommon/log"
	"gitlab.cern.ch/lb-experts/goermis/api/models"
	"gitlab.cern.ch/lb-experts/goermis/bootstrap"
	"gitlab.cern.ch/lb-experts/goermis/db"
	"gitlab.cern.ch/lb-experts/goermis/router"
	"gitlab.cern.ch/lb-experts/goermis/views"
)

const (
	// Version number
	Version = "0.0.2"
	// Release number
	Release = "2"
)

func init() {
	log.EnableColor()
	log.SetLevel(1)
	log.SetHeader("${time_rfc3339} ${level} ${short_file} ${line} ")
	file, err := os.OpenFile(bootstrap.App.IFConfig.String("logging_file"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(file)
		log.Info("File set as logger output")
	} else {
		log.Info("Failed to log to file, using default stderr")
	}
}

func main() {
	log.Info("Service Started...")
	// Echo instance
	e := router.New()
	router.InitRoutes(e)
	views.InitViews(e)

	db.Init()
	autoCreateTables(&models.Alias{}, &models.Node{}, &models.Cname{}, &models.AliasesNodes{})
	autoMigrateTables()

	// Start server
	go func() {
		if err := e.StartTLS(":8080", "/etc/ssl/certs/goermiscert.pem", "/etc/ssl/certs/goermiskey.pem"); err != nil {
			log.Debug("Ignore if error is Port Binding")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Fatal("Fatal error while shutting server down")
	}
}

func autoCreateTables(values ...interface{}) error {
	for _, value := range values {
		if !db.ManagerDB().HasTable(value) {
			err := db.ManagerDB().CreateTable(value).Error
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
	db.ManagerDB().AutoMigrate(&models.Alias{}, &models.Node{}, &models.Cname{}, &models.AliasesNodes{})

}
