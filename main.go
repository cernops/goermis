package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.cern.ch/lb-experts/goermis/api/models"
	"gitlab.cern.ch/lb-experts/goermis/bootstrap"
	"gitlab.cern.ch/lb-experts/goermis/db"
	"gitlab.cern.ch/lb-experts/goermis/router"
	"gitlab.cern.ch/lb-experts/goermis/views"
)

const (
	// Version number
	Version = "0.0.1"
	// Release number
	Release = "1"
)

var (
	log = bootstrap.Log
)

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
			log.WithFields(logrus.Fields{
				"package":  "main",
				"function": "StartTLS",
				"error":    err,
			}).Debug("Ignore if error is Port Binding")
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
		log.Fatal(err)
	}
}

func autoCreateTables(values ...interface{}) error {
	for _, value := range values {
		if !db.ManagerDB().HasTable(value) {
			err := db.ManagerDB().CreateTable(value).Error
			if err != nil {
				errClose := db.ManagerDB().Close()
				if errClose != nil {
					fmt.Printf("%s", errClose)
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
