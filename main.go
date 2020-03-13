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
	autoCreateTables(&models.Alias{}, &models.Node{}, &models.Cname{}, &models.AliasesNodes{})
	autoMigrateTables()
	//Seeding
	/*db.ManagerDB().Debug().Save(&models.Alias{

		AliasName:        "seeder",
		Behaviour:        "rogue",
		BestHosts:        1,
		External:         "yes",
		Metric:           "yes",
		PollingInterval:  5,
		Statistics:       "cmsfrontier",
		Clusters:         "none",
		LastModification: time.Now(),
		Tenant:           "kkouros",
		Hostgroup:        "ailbd",
		User:             "kkouros",
		TTL:              7,
		Relations: []*models.Relation{
			{
				Node: &models.Node{

					NodeName:         "node",
					Hostgroup:        "ailbd",
					LastModification: time.Now(),
				},

				Blacklist: true,
			},

			{
				Node: &models.Node{

					NodeName:         "node2",
					Hostgroup:        "ailbd",
					LastModification: time.Now(),
				},

				Blacklist: true,
			},
		},

		Cnames: []models.Cname{
			{
				CName: "seed",
			},
		},
	})*/

	e.Logger.Fatal(e.Start(":8080"))

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

/*
// auto drop tables on dev mode
func autoDropTables() {
	if bootstrap.App.ENV == "dev" {
		gorm.DBManager().DropTableIfExists(&models.User{}, &models.User{})
	}*/
