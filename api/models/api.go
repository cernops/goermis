package models

import (
	"net/url"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	schema "github.com/gorilla/Schema"
	"github.com/jinzhu/gorm"
	"github.com/labstack/gommon/log"
	cgorm "gitlab.cern.ch/lb-experts/goermis/db"
)

var (
	decoder = schema.NewDecoder()
)

//CreateObject creates an alias
func (a Alias) CreateObject(params url.Values) (err error) {

	decoder.IgnoreUnknownKeys(true)
	err = decoder.Decode(&a, params)
	if err != nil {
		log.Error(err)
		//panic(err)

	}
	cnames := DeleteEmpty(strings.Split(params.Get("cnames"), ","))

	return WithinTransaction(func(tx *gorm.DB) (err error) {

		// check new object
		if !cgorm.ManagerDB().NewRecord(&a) {
			return err
		}
		if err = tx.Create(&a).Error; err != nil {
			tx.Rollback() // rollback
			log.Error("Error in creation")
			return err
		}

		if len(cnames) > 0 {
			for _, cname := range cnames {
				if !cgorm.ManagerDB().NewRecord(&Cname{CName: cname}) {
					return err
				}

				if err = tx.Model(&a).Association("Cnames").Append(&Cname{CName: cname}).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
		}

		return nil
	})
}

//Prepare prepares alias before creation with def values
func (a *Alias) Prepare() {
	//Populate the struct with the default values
	a.User = "kkouros"
	a.Behaviour = "mindless"
	a.Metric = "cmsfrontier"
	a.PollingInterval = 300
	a.Statistics = "long"
	a.Clusters = "none"
	a.Tenant = "golang"
	a.LastModification = time.Now()
}

//DeleteObject deletes an alias and its Relations
func (a Alias) DeleteObject() (err error) {

	return WithinTransaction(func(tx *gorm.DB) (err error) {

		/*if tx.Where("alias_name = ?", a.AliasName).First(&a).RecordNotFound() {
			return err

		}*/
		err = tx.Where("alias_id= ?", a.ID).Delete(&Cname{}).Error
		if err != nil {
			log.Error("Cname deletion failed")
			return err
		}
		//con.Model(&Alias).Where("alias_name = ?", alias).Preload("Cnames").Delete(&Alias.Cnames)
		err = tx.Where("alias_name = ?", a.AliasName).Delete(&Alias{}).Error
		if err != nil {
			return err
		}

		return nil
	})
}

//ModifyObject modifies aliases and its associations
func (a Alias) ModifyObject(params url.Values) (err error) {
	//Prepare cnames separately
	cnames := DeleteEmpty(strings.Split(params.Get("cnames"), ","))
	spew.Dump(params)
	WithinTransaction(func(tx *gorm.DB) (err error) {

		if err = tx.Model(&a).Updates(
			map[string]interface{}{
				"external":   params.Get("external"),
				"hostgroup":  params.Get("hostgroup"),
				"best_hosts": stringToInt(params.Get("best_hosts")),
			}).Error; err != nil {
			print("Erroooorr")
			tx.Rollback()
			return err

		}

		return err
	})
	//err = a.UpdateNodes()
	err = a.UpdateCnames(cnames)
	if err != nil {
		return err
	}

	return nil
}

//UpdateCnames updates cnames
func (a Alias) UpdateCnames(newCnames []string) (err error) {
	// If there are no cnames from UI , delete them all, otherwise append them
	existingCnames := getExistingCnames(a)
	print(existingCnames)
	print(newCnames)
	if len(newCnames) > 0 {
		for _, value := range existingCnames {
			if !stringInSlice(value, newCnames) {
				a.DeleteCname(value)
			}
		}

		for _, value := range newCnames {
			if value == "" {
				continue
			}
			if !stringInSlice(value, existingCnames) {
				a.AddCname(value)
			}

		}

	} else {
		for _, value := range existingCnames {

			a.DeleteCname(value)
		}
	}
	return nil
}

/*//UpdateNodes updates cnames
func (a Alias) UpdateNodes() (err error) {
	// If there are no cnames from UI , delete them all, otherwise append them
	return err

}*/

//AddCname appends a Cname
func (a Alias) AddCname(cname string) error {
	return WithinTransaction(func(tx *gorm.DB) (err error) {
		if !cgorm.ManagerDB().NewRecord(&Cname{CName: cname}) {
			log.Error("Cname exists")
			return err
		}

		if err = tx.Model(&a).Association("Cnames").Append(&Cname{CName: cname}).Error; err != nil {
			tx.Rollback()
			return err
		}
		return nil
	})

}

//DeleteCname cname from db during modification
func (a Alias) DeleteCname(cname string) error {
	return WithinTransaction(func(tx *gorm.DB) (err error) {
		if err = tx.Where("alias_id = ? AND c_name = ?", a.ID, cname).Delete(&Cname{}).Error; err != nil {
			tx.Rollback()
			print("Erroooorr")
			return err
		}
		return err

	})

}

//GetExistingData retrieves all the data for a certain alias, for internal use
func GetExistingData(aliasName string) (a Alias, err error) {

	if con.Model(Alias{}).Preload("Cnames").Preload("Relations").Where("alias_name = ?", aliasName).First(&a).RecordNotFound() {
		return a, err

	}
	return a, nil
}

// WithinTransaction  accept DBFunc as parameter call DBFunc function within transaction begin, and commit and return error from DBFunc
func WithinTransaction(fn DBFunc) (err error) {
	tx := cgorm.ManagerDB().Begin() // start db transaction
	defer tx.Commit()
	err = fn(tx)
	// close db transaction
	log.Error(err)
	return err
}
