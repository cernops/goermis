package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/gommon/log"
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

	// Metric validation
	govalidator.TagMap["metric"] = govalidator.Validator(func(str string) bool {

		allowed := []string{"minino", "minimum", "cmsfrontier"}

		return models.StringInSlice(str, allowed)
	})

	govalidator.TagMap["nodes"] = govalidator.Validator(func(str string) bool {
		if len(str) > 0 {
			split := strings.Split(str, ",")
			var allowed = regexp.MustCompile(`^[a-z][a-z0-9\-]*[a-z0-9]$`)

			for _, s := range split {
				part := strings.Split(s, ".")
				for _, p := range part {
					if !allowed.MatchString(p) || !govalidator.InRange(len(p), 2, 40) {
						log.Error("Not valid node name: " + s)
						return false
					}
				}
			}
		}
		return true
	})

	govalidator.TagMap["cnames"] = govalidator.Validator(func(str string) bool {

		if len(str) > 0 {
			split := strings.Split(str, ",")
			var allowed = regexp.MustCompile(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`)

			for _, s := range split {
				if !allowed.MatchString(s) || !govalidator.InRange(len(s), 2, 511) {
					log.Error("Not valid cname: " + s)
					return false
				}
			}
		}
		return true
	})

	govalidator.TagMap["external"] = govalidator.Validator(func(str string) bool {
		options := []string{"yes", "no"}
		return models.StringInSlice(str, options)

	})

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
