package router

import (
	"github.com/labstack/echo/v4"
	"gitlab.cern.ch/lb-experts/goermis/api/handlers"
	"gitlab.cern.ch/lb-experts/goermis/api/middleware.go"
)

//InitRoutes initializes the routes
func InitRoutes(e *echo.Echo) {
	lbweb := e.Group("/lbweb")
	lbweb.Use(middleware.CheckAuthorization)

	lbweb.GET("/", handlers.HomeHandler)
	lbweb.GET("/create/", handlers.CreateHandler)
	lbweb.GET("/modify/", handlers.ModifyHandler)
	lbweb.GET("/display/", handlers.DisplayHandler)
	lbweb.GET("/delete/", handlers.DeleteHandler)
	lbweb.GET("/logs/", handlers.LogsHandler)
	lbweb.POST("/new_alias", handlers.NewAlias)
	lbweb.POST("/delete_alias", handlers.DeleteAlias)
	lbweb.POST("/modify_alias", handlers.ModifyAlias)
	lbweb.GET("/*/checkname/:hostname", handlers.CheckNameDNS)

	api := e.Group("/api/v1")
	api.GET("/aliases", handlers.GetAliases)
	api.GET("/aliases/:alias", handlers.GetAlias)

}
