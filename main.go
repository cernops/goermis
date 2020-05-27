package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"gitlab.cern.ch/lb-experts/goermis/api/models"
	"gitlab.cern.ch/lb-experts/goermis/db"
	"gitlab.cern.ch/lb-experts/goermis/logger"
	"gitlab.cern.ch/lb-experts/goermis/router"
	"gitlab.cern.ch/lb-experts/goermis/views"
)

const (
	// Version number
	Version = "0.0.1"
	// Release number
	Release = "2"
)

func main() {
	logger.Log.Info("Service Started...")
	// Echo instance
	e := router.New()
	logger.Log.Info("Initializing routes")
	router.InitRoutes(e)
	views.InitViews(e)

	logger.Log.Info("Initializing database")
	db.Init()
	autoCreateTables(&models.Alias{}, &models.Node{}, &models.Cname{}, &models.AliasesNodes{})
	autoMigrateTables()

	// Start server
	go func() {
		if err := e.StartTLS(":8080", "/etc/ssl/certs/goermiscert.pem", "/etc/ssl/certs/goermiskey.pem"); err != nil {
			logger.Log.Info("shutting down the server")
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
		logger.Log.Fatal(err)
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
	err := db.ManagerDB().AutoMigrate(&models.Alias{}, &models.Node{}, &models.Cname{}, &models.AliasesNodes{})
	if err != nil {
		logger.Log.Error("An error occured while migrating the tables")
	}

}
