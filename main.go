package main

import (
	"fmt"

	"gitlab.cern.ch/lb-experts/goermis/api/models"
	"gitlab.cern.ch/lb-experts/goermis/db"
	"gitlab.cern.ch/lb-experts/goermis/router"
	"gitlab.cern.ch/lb-experts/goermis/views"
)

func main() {
	// Echo instance
	e := router.New()
	router.InitRoutes(e)
	views.InitViews(e)

	db.Init()
	autoCreateTables(&models.Alias{}, &models.Node{}, &models.Cname{}, &models.Relation{})
	autoMigrateTables()
	//Seeding
	/*db.ManagerDB().Debug().Save(&models.Alias{

		AliasName:        "test62",
		Behaviour:        "rogue",
		BestHosts:        1,
		External:         "yes",
		Metric:           "yes",
		PollingInterval:  5,
		Statistics:       "very much",
		Clusters:         "alot",
		LastModification: time.Now(),
		Tenant:           "me",
		Hostgroup:        "yeap",
		User:             "againme",
		TTL:              7,
		Relations: []*models.Relation{
			{
				Node: &models.Node{

					NodeName:         "imback16",
					Hostgroup:        "yeap",
					LastModification: time.Now(),
				},

				Blacklist: true,
			},

			{
				Node: &models.Node{

					NodeName:         "im16",
					Hostgroup:        "yeap",
					LastModification: time.Now(),
				},

				Blacklist: true,
			},
		},

		Cnames: []models.Cname{
			{
				CName: "heyheyu",
			},
		},
	})*/

	e.Logger.Fatal(e.StartTLS(":8080", "host_cert.pem", "host_key.pem"))

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
	db.ManagerDB().AutoMigrate(&models.Alias{}, &models.Node{}, &models.Cname{}, &models.Relation{})
}

/*
// auto drop tables on dev mode
func autoDropTables() {
	if bootstrap.App.ENV == "dev" {
		gorm.DBManager().DropTableIfExists(&models.User{}, &models.User{})
	}*/
