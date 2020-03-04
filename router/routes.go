package router

import (
	"github.com/labstack/echo/v4"
	"gitlab.cern.ch/lb-experts/goermis/api/handlers"
)

//InitRoutes initializes the routes
func InitRoutes(e *echo.Echo) {

	e.GET("/", handlers.HomeHandler)
	e.GET("/lbweb", handlers.HomeHandler)
	e.GET("/lbweb/create/", handlers.CreateHandler)
	e.GET("/lbweb/modify/", handlers.ModifyHandler)
	e.GET("/lbweb/display/", handlers.DisplayHandler)
	e.GET("/lbweb/delete/", handlers.DeleteHandler)
	e.GET("/lbweb/logs/", handlers.LogsHandler)

	e.GET("/aliases", handlers.GetAliases)
	e.GET("/aliases/:alias", handlers.GetAlias)

	e.GET("/lbweb/*/checkname", handlers.CheckNameDNS)
	e.POST("/new_alias", handlers.NewAlias)
	e.POST("/delete_alias", handlers.DeleteAlias)
	e.POST("/modify_alias", handlers.ModifyAlias)
}
