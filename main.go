package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"time"

	"gitlab.cern.ch/lb-experts/goermis/alarms"
	"gitlab.cern.ch/lb-experts/goermis/api/ermis"
	"gitlab.cern.ch/lb-experts/goermis/bootstrap"
	"gitlab.cern.ch/lb-experts/goermis/db"
	"gitlab.cern.ch/lb-experts/goermis/router"
	"gitlab.cern.ch/lb-experts/goermis/views"
)

const (
	// Version number
	Version = "1.3.0"
	// Release number
	Release = "1"
)

var (
	log = bootstrap.GetLog()
	cfg = bootstrap.GetConf()
)

func main() {
	bootstrap.ParseFlags()
	bootstrap.SetLogLevel()
	log.Info("============Service Started=============")

	// Echo instance
	echo := router.New()
	db.InitDB()
	defer db.Close()
	//Initiate template views
	views.InitViews(echo)
	autoMigrateTables()

	//Alarms periodic check/update

	ticker := time.NewTicker(time.Duration(cfg.Timers.Alarms) * time.Minute)

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
				log.Debugf("%v minutes passed, preparing to check for active alarms", cfg.Timers.Alarms)
				alarms.PeriodicAlarmCheck()
			}
		}
	}()
	log.Debug("alarms updated")

	/* Start server
	       Error handling is done a bit differently in this situation. The reason is that
		   when server is restarted we force it to reuse the same socket. Despite being successfully
		   restarted, it throws a bind error. This interferes with other important cases where we need
		   to shut down the service */

	go func() {

		err := echo.StartTLS(":8080",
			cfg.Certs.ErmisCert,
			cfg.Certs.ErmisKey)
		//Avoiding uneccesary logs and failures when restarting
		if !strings.HasSuffix(err.Error(), "bind: address already in use") {
			log.Fatalf("Failed to start server: %v", err)

		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds. It is needed to accomplish socket recycling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	s := <-quit
	log.Infof("received quit signal %v", s)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := echo.Shutdown(ctx); err != nil {
		log.Fatal("Fatal error while shutting server down " + err.Error())

	}

}

// autoMigrateTables: migrate table columns using GORM. Will not delete/change types for security reasons
func autoMigrateTables() {
	db.GetConn().AutoMigrate(&ermis.Alias{}, &ermis.Node{}, &ermis.Cname{}, &ermis.Alarm{}, &ermis.Relation{})

}
